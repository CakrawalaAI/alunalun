package middleware

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/radjathaher/alunalun/api/internal/utils/auth"
)

// Context keys for auth values
type contextKey string

const (
	claimsKey   contextKey = "claims"
	userIDKey   contextKey = "user_id"
	usernameKey contextKey = "username"
	emailKey    contextKey = "email"
)

// AuthInterceptor handles JWT authentication for ConnectRPC
type AuthInterceptor struct {
	tokenManager *auth.TokenManager
}

// NewAuthInterceptor creates a new auth interceptor
func NewAuthInterceptor(tokenManager *auth.TokenManager) *AuthInterceptor {
	return &AuthInterceptor{
		tokenManager: tokenManager,
	}
}

// WrapUnary creates a unary interceptor for authentication
func (a *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Check if endpoint requires authentication
		if !isPublicEndpoint(req.Spec().Procedure) {
			// Extract and validate token
			token := extractBearerToken(req.Header())
			if token == "" {
				// No token on protected endpoint
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					nil, // Don't leak info about why auth failed
				)
			}

			// Validate token
			claims, err := a.tokenManager.ValidateToken(token)
			if err != nil {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					nil, // Don't leak token validation errors
				)
			}

			// Add claims to context
			ctx = context.WithValue(ctx, claimsKey, claims)
			ctx = context.WithValue(ctx, userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, usernameKey, claims.Username)
			ctx = context.WithValue(ctx, emailKey, claims.Email)
		} else {
			// Public endpoint - optionally extract token if provided
			token := extractBearerToken(req.Header())
			if token != "" {
				// Best effort - validate but don't fail if invalid
				if claims, err := a.tokenManager.ValidateToken(token); err == nil {
					ctx = context.WithValue(ctx, claimsKey, claims)
					ctx = context.WithValue(ctx, userIDKey, claims.UserID)
					ctx = context.WithValue(ctx, usernameKey, claims.Username)
					ctx = context.WithValue(ctx, emailKey, claims.Email)
				}
			}
		}

		return next(ctx, req)
	}
}

// WrapStreamingClient creates a streaming client interceptor
func (a *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		// Streaming not used in MVP, but structure is here
		return next(ctx, spec)
	}
}

// WrapStreamingHandler creates a streaming handler interceptor
func (a *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		// Streaming not used in MVP, but structure is here
		return next(ctx, conn)
	}
}

// extractBearerToken extracts token from Authorization header
func extractBearerToken(headers http.Header) string {
	auth := headers.Get("Authorization")
	if auth == "" {
		return ""
	}
	
	// Handle both "Bearer " and direct token
	if strings.HasPrefix(auth, "Bearer ") {
		return auth[7:]
	}
	return auth
}

// isPublicEndpoint checks if an endpoint allows public access
func isPublicEndpoint(procedure string) bool {
	// Public endpoints that don't require auth
	publicEndpoints := []string{
		// Auth endpoints
		"/api.v1.service.auth.AuthService/CheckUsername",
		"/api.v1.service.auth.AuthService/InitAnonymous", // For future use
		"/api.v1.service.auth.AuthService/Authenticate",
		"/api.v1.service.auth.AuthService/RefreshToken",
		
		// Public read endpoints for pins
		"/api.v1.service.PinService/ListPins",  // Public map viewing
		"/api.v1.service.PinService/GetPin",     // Public pin details
		
		// Public user info
		"/api.v1.service.UserService/GetUser",   // Public user profile
	}
	
	for _, endpoint := range publicEndpoints {
		if procedure == endpoint {
			return true
		}
	}
	return false
}

// GetClaims retrieves JWT claims from context
func GetClaims(ctx context.Context) *auth.Claims {
	claims, _ := ctx.Value(claimsKey).(*auth.Claims)
	return claims
}

// GetUserID gets user ID from context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

// GetUsername gets username from context
func GetUsername(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameKey).(string)
	return username, ok
}

// GetEmail gets email from context
func GetEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(emailKey).(string)
	return email, ok
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(ctx context.Context) bool {
	claims := GetClaims(ctx)
	return claims != nil && claims.UserID != "" && !claims.IsAnonymous
}

// IsAnonymous checks if the request is from an anonymous user (future)
func IsAnonymous(ctx context.Context) bool {
	claims := GetClaims(ctx)
	return claims != nil && claims.IsAnonymous
}