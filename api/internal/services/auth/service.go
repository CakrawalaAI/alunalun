package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	
	"connectrpc.com/connect"
	"github.com/radjathaher/alunalun/api/internal/protocgen/v1/entities"
	servicev1 "github.com/radjathaher/alunalun/api/internal/protocgen/v1/auth_service"
	"github.com/radjathaher/alunalun/api/internal/protocgen/v1/auth_service/auth_servicev1connect"
	"github.com/radjathaher/alunalun/api/internal/utils/auth"
	"github.com/radjathaher/alunalun/api/internal/utils/oauth"
)

// Service implements the authentication service orchestrator
type Service struct {
	auth_servicev1connect.UnimplementedAuthServiceHandler
	registry       *auth.ProviderRegistry
	tokenManager   *auth.TokenManager
	sessionManager *auth.SessionManager
	userStore      auth.UserStore
	config         *auth.Config
}

// NewService creates a new auth service
func NewService(
	registry *auth.ProviderRegistry,
	tokenManager *auth.TokenManager,
	sessionManager *auth.SessionManager,
	userStore auth.UserStore,
	config *auth.Config,
) (*Service, error) {
	if registry == nil {
		return nil, errors.New("provider registry is required")
	}
	if tokenManager == nil {
		return nil, errors.New("token manager is required")
	}
	if sessionManager == nil {
		return nil, errors.New("session manager is required")
	}
	if userStore == nil {
		return nil, errors.New("user store is required")
	}
	if config == nil {
		config = auth.DefaultConfig()
	}
	
	return &Service{
		registry:       registry,
		tokenManager:   tokenManager,
		sessionManager: sessionManager,
		userStore:      userStore,
		config:         config,
	}, nil
}

// CheckUsername checks if a username is available
func (s *Service) CheckUsername(
	ctx context.Context,
	req *connect.Request[servicev1.CheckUsernameRequest],
) (*connect.Response[servicev1.CheckUsernameResponse], error) {
	if req.Msg.Username == "" {
		return connect.NewResponse(&servicev1.CheckUsernameResponse{
			Available: false,
			Message:   "Username is required",
		}), nil
	}
	
	// Check username availability
	available, err := s.userStore.CheckUsernameAvailable(ctx, req.Msg.Username)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to check username: %w", err))
	}
	
	response := &servicev1.CheckUsernameResponse{
		Available: available,
	}
	
	if !available {
		response.Message = "Username is already taken"
	}
	
	return connect.NewResponse(response), nil
}

// InitAnonymous initializes an anonymous session
func (s *Service) InitAnonymous(
	ctx context.Context,
	req *connect.Request[servicev1.InitAnonymousRequest],
) (*connect.Response[servicev1.InitAnonymousResponse], error) {
	if req.Msg.Username == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("username is required"))
	}
	
	// Use the anonymous provider
	provider, err := s.registry.Get("anonymous")
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("anonymous provider not available: %w", err))
	}
	
	// Create anonymous request
	anonReq := auth.AnonymousRequest{
		Username: req.Msg.Username,
	}
	
	credential, err := json.Marshal(anonReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to marshal request: %w", err))
	}
	
	// Authenticate anonymously
	userInfo, err := provider.Authenticate(ctx, string(credential))
	if err != nil {
		if authErr, ok := err.(*auth.AuthError); ok {
			if authErr.Code == "USERNAME_TAKEN" {
				return nil, connect.NewError(connect.CodeAlreadyExists, errors.New(authErr.Message))
			}
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create anonymous session: %w", err))
	}
	
	// Generate JWT token (no expiry for anonymous)
	claims := &auth.Claims{
		SessionID:   userInfo.ID, // Session ID is used as user ID for anonymous
		Username:    userInfo.Username,
		Provider:    "anonymous",
		IsAnonymous: true,
	}
	
	token, err := s.tokenManager.GenerateToken(claims, 0) // No expiry
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate token: %w", err))
	}
	
	return connect.NewResponse(&servicev1.InitAnonymousResponse{
		Token:     token,
		SessionId: userInfo.ID,
		Username:  req.Msg.Username,
	}), nil
}

// Authenticate handles authentication with various providers
func (s *Service) Authenticate(
	ctx context.Context,
	req *connect.Request[servicev1.AuthenticateRequest],
) (*connect.Response[servicev1.AuthenticateResponse], error) {
	if req.Msg.Provider == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("provider is required"))
	}
	if req.Msg.Credential == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("credential is required"))
	}
	
	// Get the provider
	provider, err := s.registry.Get(req.Msg.Provider)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("provider %s not found: %w", req.Msg.Provider, err))
	}
	
	// Authenticate with the provider
	userInfo, err := provider.Authenticate(ctx, req.Msg.Credential)
	if err != nil {
		// Check if it's a special case (like magic link sent)
		if authErr, ok := err.(*auth.AuthError); ok {
			switch authErr.Code {
			case "MAGIC_LINK_SENT":
				// Return success with no token (client should show "check email" message)
				return connect.NewResponse(&servicev1.AuthenticateResponse{
					Token: "", // No token yet
					User:  nil,
				}), nil
			case auth.ErrInvalidCredentials:
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New(authErr.Message))
			case auth.ErrUserNotFound:
				return nil, connect.NewError(connect.CodeNotFound, errors.New(authErr.Message))
			case auth.ErrUserDisabled:
				return nil, connect.NewError(connect.CodePermissionDenied, errors.New(authErr.Message))
			case auth.ErrEmailNotVerified:
				return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New(authErr.Message))
			}
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authentication failed: %w", err))
	}
	
	// Find or create user in our system
	user, err := s.findOrCreateUser(ctx, userInfo)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to process user: %w", err))
	}
	
	// Handle session migration if provided
	sessionMigrated := false
	if req.Msg.SessionId != nil && *req.Msg.SessionId != "" {
		// Check if this is an anonymous session
		session, err := s.sessionManager.Validate(ctx, *req.Msg.SessionId)
		if err == nil && session.IsAnonymous {
			// Migrate the session to the authenticated user
			if err := s.sessionManager.MigrateToUser(ctx, *req.Msg.SessionId, user.ID); err != nil {
				// Log error but continue
				fmt.Printf("failed to migrate session: %v\n", err)
			} else {
				sessionMigrated = true
			}
		}
	}
	
	// Create a new authenticated session
	session, err := s.sessionManager.CreateAuthenticated(ctx, user.ID, s.config.Session.TTL)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create session: %w", err))
	}
	
	// Generate JWT token
	claims := &auth.Claims{
		UserID:      user.ID,
		SessionID:   session.ID,
		Username:    user.Username,
		Email:       user.Email,
		Provider:    req.Msg.Provider,
		IsAnonymous: false,
	}
	
	token, err := s.tokenManager.GenerateToken(claims, s.config.JWT.AccessTokenTTL)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate token: %w", err))
	}
	
	// Convert to proto user
	protoUser := s.userToProto(user)
	
	return connect.NewResponse(&servicev1.AuthenticateResponse{
		Token:           token,
		User:            protoUser,
		SessionMigrated: sessionMigrated,
	}), nil
}

// RefreshToken refreshes an expired JWT token
func (s *Service) RefreshToken(
	ctx context.Context,
	req *connect.Request[servicev1.RefreshTokenRequest],
) (*connect.Response[servicev1.RefreshTokenResponse], error) {
	if req.Msg.ExpiredToken == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("expired token is required"))
	}
	
	// Refresh the token
	newToken, err := s.tokenManager.RefreshToken(req.Msg.ExpiredToken, s.config.JWT.AccessTokenTTL)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("failed to refresh token: %w", err))
	}
	
	return connect.NewResponse(&servicev1.RefreshTokenResponse{
		Token: newToken,
	}), nil
}

// findOrCreateUser finds an existing user or creates a new one based on UserInfo
func (s *Service) findOrCreateUser(ctx context.Context, info *auth.UserInfo) (*auth.User, error) {
	// Try to find user by email
	user, err := s.userStore.GetUserByEmail(ctx, info.Email)
	if err != nil {
		// User doesn't exist, create new one
		now := time.Now()
		user = &auth.User{
			Email:           info.Email,
			Username:        info.Username,
			FirstName:       info.FirstName,
			LastName:        info.LastName,
			Picture:         info.Picture,
			EmailVerified:   info.EmailVerified,
			EmailVerifiedAt: &info.VerifiedAt,
			CreatedAt:       now,
			UpdatedAt:       now,
			LastLoginAt:     &now,
			Status:          "active",
			Metadata: map[string]interface{}{
				"provider":    info.Provider,
				"provider_id": info.ProviderID,
			},
		}
		
		// If username is empty, use email
		if user.Username == "" {
			user.Username = user.Email
		}
		
		if err := s.userStore.CreateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		
		// Reload user to get generated ID
		user, err = s.userStore.GetUserByEmail(ctx, info.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to reload user: %w", err)
		}
	} else {
		// Update last login
		now := time.Now()
		user.LastLoginAt = &now
		user.UpdatedAt = now
		
		// Update user info if changed
		if info.FirstName != "" && user.FirstName != info.FirstName {
			user.FirstName = info.FirstName
		}
		if info.LastName != "" && user.LastName != info.LastName {
			user.LastName = info.LastName
		}
		if info.Picture != "" && user.Picture != info.Picture {
			user.Picture = info.Picture
		}
		
		if err := s.userStore.UpdateUser(ctx, user); err != nil {
			// Log error but continue
			fmt.Printf("failed to update user: %v\n", err)
		}
	}
	
	return user, nil
}

// userToProto converts internal User to proto User
func (s *Service) userToProto(user *auth.User) *entities.User {
	if user == nil {
		return nil
	}
	
	protoUser := &entities.User{
		Id:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarUrl: user.Picture,
		Status:    s.statusToProto(user.Status),
	}
	
	if user.EmailVerifiedAt != nil {
		protoUser.EmailVerifiedAt = user.EmailVerifiedAt.Unix()
	}
	
	if user.CreatedAt.Unix() > 0 {
		protoUser.CreatedAt = user.CreatedAt.Unix()
	}
	
	if user.UpdatedAt.Unix() > 0 {
		protoUser.UpdatedAt = user.UpdatedAt.Unix()
	}
	
	return protoUser
}

// statusToProto converts status string to proto enum
func (s *Service) statusToProto(status string) entities.UserStatus {
	switch status {
	case "active":
		return entities.UserStatus_USER_STATUS_ACTIVE
	case "disabled":
		return entities.UserStatus_USER_STATUS_DISABLED
	case "pending":
		return entities.UserStatus_USER_STATUS_PENDING
	default:
		return entities.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

// RegisterProviders registers all configured providers with the registry
func (s *Service) RegisterProviders() error {
	// Register OAuth providers
	for name, config := range s.config.Providers {
		if !config.Enabled {
			continue
		}
		
		switch config.Type {
		case "oauth":
			// Create OAuth provider based on name
			switch name {
			case "google":
				provider, err := oauth.NewGoogleProvider(
					config.Config["client_id"],
					config.Config["client_secret"],
					config.Config["redirect_url"],
				)
				if err != nil {
					return fmt.Errorf("failed to create Google provider: %w", err)
				}
				if err := s.registry.Register(provider); err != nil {
					return fmt.Errorf("failed to register Google provider: %w", err)
				}
			// Add more OAuth providers as needed
			}
			
		case "internal":
			// Register internal providers
			switch name {
			case "email":
				emailConfig := auth.EmailProviderConfig{
					MinLength:           8,
					RequireUppercase:    true,
					RequireLowercase:    true,
					RequireNumbers:      true,
					RequireSpecialChar:  false,
					BcryptCost:          12,
					RequireVerification: false, // For development
				}
				
				provider, err := auth.NewEmailPasswordProvider(s.userStore, emailConfig)
				if err != nil {
					return fmt.Errorf("failed to create email provider: %w", err)
				}
				if err := s.registry.Register(provider); err != nil {
					return fmt.Errorf("failed to register email provider: %w", err)
				}
				
			case "anonymous":
				provider, err := auth.NewAnonymousProvider(s.sessionManager, s.userStore)
				if err != nil {
					return fmt.Errorf("failed to create anonymous provider: %w", err)
				}
				if err := s.registry.Register(provider); err != nil {
					return fmt.Errorf("failed to register anonymous provider: %w", err)
				}
				
			// Add magic link provider when email sender is available
			}
		}
	}
	
	return nil
}