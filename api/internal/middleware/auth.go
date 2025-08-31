package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
)

// Context keys for auth values
type contextKey string

const (
	userIDKey     contextKey = "user_id"
	sessionIDKey  contextKey = "session_id"
	authTypeKey   contextKey = "auth_type"
	usernameKey   contextKey = "username"
)

// AuthType represents the type of authentication
type AuthType string

const (
	AuthTypeAnonymous     AuthType = "anonymous"
	AuthTypeAuthenticated AuthType = "authenticated"
	AuthTypeNone          AuthType = "none"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret string
}

// NewAuthInterceptor creates a new auth interceptor with token type routing
func NewAuthInterceptor(config AuthConfig) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Extract token from Authorization header
			token := extractBearerToken(req.Header())
			
			if token == "" {
				// No token - check if endpoint allows anonymous access
				if isProtectedEndpoint(req.Spec().Procedure) {
					return nil, connect.NewError(connect.CodeUnauthenticated, nil)
				}
				ctx = context.WithValue(ctx, authTypeKey, AuthTypeNone)
				return next(ctx, req)
			}
			
			// Parse token to determine type
			claims, err := parseTokenClaims(token, config.JWTSecret)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			
			// Route based on token type
			tokenType, _ := claims["type"].(string)
			switch tokenType {
			case "anonymous":
				// Anonymous tokens never expire
				sessionID, _ := claims["session_id"].(string)
				username, _ := claims["username"].(string)
				
				ctx = context.WithValue(ctx, authTypeKey, AuthTypeAnonymous)
				ctx = context.WithValue(ctx, sessionIDKey, sessionID)
				ctx = context.WithValue(ctx, usernameKey, username)
				
			case "authenticated", "":
				// Authenticated tokens check expiration
				if exp, ok := claims["exp"].(float64); ok {
					if time.Now().Unix() > int64(exp) {
						// Token expired but might be eligible for refresh
						if refreshUntil, ok := claims["refresh_until"].(float64); ok {
							if time.Now().Unix() <= int64(refreshUntil) {
								// Token can be refreshed - let RefreshToken endpoint handle it
								if req.Spec().Procedure == "/api.v1.service.AuthService/RefreshToken" {
									// Allow through for refresh
									ctx = context.WithValue(ctx, authTypeKey, AuthTypeAuthenticated)
									ctx = context.WithValue(ctx, "expired_token", token)
									return next(ctx, req)
								}
							}
						}
						return nil, connect.NewError(connect.CodeUnauthenticated, nil)
					}
				}
				
				userID, _ := claims["sub"].(string)
				username, _ := claims["username"].(string)
				
				ctx = context.WithValue(ctx, authTypeKey, AuthTypeAuthenticated)
				ctx = context.WithValue(ctx, userIDKey, userID)
				ctx = context.WithValue(ctx, usernameKey, username)
				
			default:
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}
			
			return next(ctx, req)
		}
	}
}

// extractBearerToken extracts token from Authorization header
func extractBearerToken(headers http.Header) string {
	auth := headers.Get("Authorization")
	if auth == "" {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}

// parseTokenClaims parses JWT without full validation (for type checking)
func parseTokenClaims(tokenString string, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	
	if err != nil && err != jwt.ErrTokenExpired {
		// Allow expired tokens through for type checking
		return nil, err
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenMalformed
	}
	
	return claims, nil
}

// isProtectedEndpoint checks if an endpoint requires authentication
func isProtectedEndpoint(procedure string) bool {
	// Public endpoints that don't require auth
	publicEndpoints := []string{
		"/api.v1.service.AuthService/CheckUsername",
		"/api.v1.service.AuthService/InitAnonymous",
		"/api.v1.service.AuthService/Authenticate",
		"/api.v1.service.AuthService/RefreshToken",
	}
	
	for _, endpoint := range publicEndpoints {
		if procedure == endpoint {
			return false
		}
	}
	return true // All other endpoints are protected
}

// GetUserID gets user ID from context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

// GetSessionID gets session ID from context
func GetSessionID(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDKey).(string)
	return sessionID, ok
}

// GetAuthType gets auth type from context
func GetAuthType(ctx context.Context) AuthType {
	authType, ok := ctx.Value(authTypeKey).(AuthType)
	if !ok {
		return AuthTypeNone
	}
	return authType
}

// GetUsername gets username from context
func GetUsername(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameKey).(string)
	return username, ok
}

// IsAuthenticated checks if the request is from an authenticated user
func IsAuthenticated(ctx context.Context) bool {
	return GetAuthType(ctx) == AuthTypeAuthenticated
}

// IsAnonymous checks if the request is from an anonymous session
func IsAnonymous(ctx context.Context) bool {
	return GetAuthType(ctx) == AuthTypeAnonymous
}