package user

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	entitiesv1 "github.com/radjathaher/alunalun/api/internal/protocgen/v1/entities"
	servicev1 "github.com/radjathaher/alunalun/api/internal/protocgen/v1/service"
	"github.com/radjathaher/alunalun/api/internal/protocgen/v1/service/servicev1connect"
	"github.com/radjathaher/alunalun/api/internal/protoconv"
	"github.com/radjathaher/alunalun/api/internal/repository"
	"github.com/radjathaher/alunalun/api/internal/utils/auth"
)

// Service implements the UserService
type Service struct {
	servicev1connect.UnimplementedUserServiceHandler
	queries      *repository.Queries
	tokenManager *auth.TokenManager
}

// NewService creates a new user service
func NewService(queries *repository.Queries, tokenManager *auth.TokenManager) *Service {
	return &Service{
		queries:      queries,
		tokenManager: tokenManager,
	}
}

// GetCurrentUser returns the current authenticated user or error if not authenticated
func (s *Service) GetCurrentUser(
	ctx context.Context,
	req *connect.Request[servicev1.GetCurrentUserRequest],
) (*connect.Response[servicev1.GetCurrentUserResponse], error) {
	// Extract user from JWT context
	claims := s.extractClaims(req.Header())
	if claims == nil || claims.UserID == "" {
		// No authentication - return error prompting to sign in
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("authentication required - please sign in"),
		)
	}

	// Get user from database
	user, err := s.queries.GetUserByID(ctx, claims.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			// User in JWT but not in database - invalid token
			return nil, connect.NewError(
				connect.CodeUnauthenticated,
				errors.New("invalid authentication - please sign in again"),
			)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get user: %w", err))
	}

	// Convert to proto
	protoUser := protoconv.UserToProto(user)

	// Return user with same token (no need to refresh if valid)
	return connect.NewResponse(&servicev1.GetCurrentUserResponse{
		User:        protoUser,
		AccessToken: "", // Client already has the token
	}), nil
}

// RegisterUser converts an anonymous user to registered or creates new registered user
// This is typically called after OAuth authentication succeeds
func (s *Service) RegisterUser(
	ctx context.Context,
	req *connect.Request[servicev1.RegisterUserRequest],
) (*connect.Response[servicev1.RegisterUserResponse], error) {
	// This would typically be called internally after OAuth flow
	// For MVP, registration happens through OAuth providers only
	
	// Validate request
	if req.Msg.Email == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("email is required"),
		)
	}
	if req.Msg.Username == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("username is required"),
		)
	}

	// Check if email already exists
	existingUser, err := s.queries.GetUserByEmail(ctx, &req.Msg.Email)
	if err == nil && existingUser != nil {
		return nil, connect.NewError(
			connect.CodeAlreadyExists,
			errors.New("email already registered"),
		)
	}

	// Check if username is taken
	existingUser, err = s.queries.GetUserByUsername(ctx, req.Msg.Username)
	if err == nil && existingUser != nil {
		return nil, connect.NewError(
			connect.CodeAlreadyExists,
			errors.New("username already taken"),
		)
	}

	// Create new user
	newUser := &entitiesv1.User{
		Email:    &req.Msg.Email,
		Username: req.Msg.Username,
		Metadata: map[string]string{
			"registered_via": "oauth",
		},
	}

	// Convert to repository params
	createParams, err := protoconv.ProtoToCreateUserParams(newUser)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to prepare user: %w", err))
	}

	// Create user in database
	createdUser, err := s.queries.CreateUser(ctx, createParams)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create user: %w", err))
	}

	// Convert to proto
	protoUser := protoconv.UserToProto(createdUser)

	// Generate JWT token for the new user
	claims := &auth.Claims{
		UserID:      createdUser.ID,
		Username:    createdUser.Username,
		Email:       *createdUser.Email,
		Provider:    "oauth",
		IsAnonymous: false,
	}

	token, err := s.tokenManager.GenerateToken(claims, 3600) // 1 hour
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate token: %w", err))
	}

	return connect.NewResponse(&servicev1.RegisterUserResponse{
		User:        protoUser,
		AccessToken: token,
	}), nil
}

// GetUser returns a user by ID (public information only)
func (s *Service) GetUser(
	ctx context.Context,
	req *connect.Request[servicev1.GetUserRequest],
) (*connect.Response[servicev1.GetUserResponse], error) {
	if req.Msg.UserId == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("user_id is required"),
		)
	}

	// Get user from database
	user, err := s.queries.GetUserByID(ctx, req.Msg.UserId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get user: %w", err))
	}

	// Convert to proto (public info only)
	protoUser := protoconv.UserToProto(user)
	
	// Remove sensitive metadata for public view
	publicUser := &entitiesv1.User{
		Id:        protoUser.Id,
		Username:  protoUser.Username,
		CreatedAt: protoUser.CreatedAt,
		// Don't expose email or full metadata publicly
	}

	return connect.NewResponse(&servicev1.GetUserResponse{
		User: publicUser,
	}), nil
}

// extractClaims extracts JWT claims from request headers
func (s *Service) extractClaims(headers connect.Headers) *auth.Claims {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return nil
	}

	// Remove "Bearer " prefix
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	// Validate token
	claims, err := s.tokenManager.ValidateToken(token)
	if err != nil {
		return nil
	}

	return claims
}

// Future extension for anonymous support (not implemented in MVP)
// When OAuth limits are hit, this could be enabled:
/*
func (s *Service) CreateAnonymousUser(
	ctx context.Context,
	fingerprint string,
) (*entitiesv1.User, string, error) {
	// Create anonymous user with NULL email
	anonUser := protoconv.CreateAnonymousUser(fingerprint)
	
	// Save to database
	createParams, _ := protoconv.ProtoToCreateUserParams(anonUser)
	createdUser, err := s.queries.CreateUser(ctx, createParams)
	if err != nil {
		return nil, "", err
	}
	
	// Generate never-expiring token
	claims := &auth.Claims{
		UserID:      createdUser.ID,
		Username:    createdUser.Username,
		IsAnonymous: true,
	}
	
	token, _ := s.tokenManager.GenerateToken(claims, 0) // No expiry
	
	return protoconv.UserToProto(createdUser), token, nil
}

func (s *Service) MigrateAnonymousToAuthenticated(
	ctx context.Context,
	anonUserID string,
	email string,
	username string,
) error {
	// Simply update the user record - posts stay linked
	updateParams := &repository.UpdateUserParams{
		ID:       anonUserID,
		Email:    &email,
		Username: username,
		Metadata: []byte(`{"migrated": true, "type": "authenticated"}`),
	}
	
	_, err := s.queries.UpdateUser(ctx, updateParams)
	return err
}
*/