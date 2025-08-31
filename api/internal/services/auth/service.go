package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	
	servicev1 "github.com/ckrwl/alunalun/api/gen/api/v1/service"
	"github.com/ckrwl/alunalun/api/gen/api/v1/service/servicev1connect"
	"github.com/ckrwl/alunalun/api/internal/repository"
)

type Service struct {
	servicev1connect.UnimplementedAuthServiceHandler
	db        *repository.Queries
	jwtSecret string
}

func NewService(db *repository.Queries, jwtSecret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// CheckUsername checks if a username is available
func (s *Service) CheckUsername(
	ctx context.Context,
	req *connect.Request[servicev1.CheckUsernameRequest],
) (*connect.Response[servicev1.CheckUsernameResponse], error) {
	taken, err := s.db.CheckUsername(ctx, req.Msg.Username)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	
	resp := &servicev1.CheckUsernameResponse{
		Available: !taken,
	}
	if taken {
		resp.Message = "Username is already taken"
	}
	
	return connect.NewResponse(resp), nil
}

// InitAnonymous creates a new anonymous session
func (s *Service) InitAnonymous(
	ctx context.Context,
	req *connect.Request[servicev1.InitAnonymousRequest],
) (*connect.Response[servicev1.InitAnonymousResponse], error) {
	// Generate session ID
	sessionID := uuid.New().String()
	
	// Try to claim username
	result, err := s.db.ClaimUsernameForAnonymous(ctx, repository.ClaimUsernameForAnonymousParams{
		Username:  req.Msg.Username,
		SessionID: uuid.MustParse(sessionID),
	})
	if err != nil || result.Username == "" {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("username not available"))
	}
	
	// Generate non-expiring JWT for anonymous user
	token, err := s.generateAnonymousToken(sessionID, req.Msg.Username)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	
	return connect.NewResponse(&servicev1.InitAnonymousResponse{
		Token:     token,
		SessionId: sessionID,
		Username:  req.Msg.Username,
	}), nil
}

// Authenticate handles provider-based authentication
func (s *Service) Authenticate(
	ctx context.Context,
	req *connect.Request[servicev1.AuthenticateRequest],
) (*connect.Response[servicev1.AuthenticateResponse], error) {
	switch req.Msg.Provider {
	case "google":
		return s.authenticateGoogle(ctx, req)
	default:
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("provider not supported"))
	}
}

// RefreshToken refreshes an expired JWT
func (s *Service) RefreshToken(
	ctx context.Context,
	req *connect.Request[servicev1.RefreshTokenRequest],
) (*connect.Response[servicev1.RefreshTokenResponse], error) {
	// Parse expired token
	token, err := jwt.Parse(req.Msg.ExpiredToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.jwtSecret), nil
	})
	
	// We allow expired tokens through for refresh
	if err != nil && err != jwt.ErrTokenExpired {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid token claims"))
	}
	
	// Check refresh_until claim
	if refreshUntil, ok := claims["refresh_until"].(float64); ok {
		if time.Now().Unix() > int64(refreshUntil) {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("token refresh period expired"))
		}
	}
	
	// Generate new token with same claims but new expiry
	userID, _ := claims["sub"].(string)
	username, _ := claims["username"].(string)
	
	newToken, err := s.generateUserToken(userID, username)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	
	return connect.NewResponse(&servicev1.RefreshTokenResponse{
		Token: newToken,
	}), nil
}

// generateAnonymousToken creates a non-expiring JWT for anonymous users
func (s *Service) generateAnonymousToken(sessionID, username string) (string, error) {
	claims := jwt.MapClaims{
		"type":       "anonymous",
		"session_id": sessionID,
		"username":   username,
		"iat":        time.Now().Unix(),
		// No exp claim - token never expires
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// generateUserToken creates a JWT for authenticated users
func (s *Service) generateUserToken(userID, username string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"type":          "authenticated",
		"sub":           userID,
		"username":      username,
		"iat":           now.Unix(),
		"exp":           now.Add(1 * time.Hour).Unix(),        // 1 hour expiry
		"refresh_until": now.Add(30 * 24 * time.Hour).Unix(), // 30 days refresh window
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// authenticateGoogle handles Google OAuth authentication
func (s *Service) authenticateGoogle(
	ctx context.Context,
	req *connect.Request[servicev1.AuthenticateRequest],
) (*connect.Response[servicev1.AuthenticateResponse], error) {
	// TODO: Implement Google OAuth token validation
	// 1. Validate Google ID token
	// 2. Extract email and profile info
	// 3. Create or get user
	// 4. Handle session migration if session_id provided
	// 5. Generate JWT
	
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Google OAuth not yet implemented"))
}

// generateSecureToken creates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}