package auth

import (
	"context"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// claimsKey is the context key for JWT claims
	claimsKey contextKey = "auth_claims"
)

// ContextWithClaims adds JWT claims to the context
func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// ClaimsFromContext retrieves JWT claims from the context
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	return claims, ok
}

// UserIDFromContext retrieves the user ID from context claims
func UserIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return "", false
	}
	return claims.UserID, claims.UserID != ""
}

// SessionIDFromContext retrieves the session ID from context claims
func SessionIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return "", false
	}
	return claims.SessionID, claims.SessionID != ""
}

// IsAnonymousContext checks if the context represents an anonymous user
func IsAnonymousContext(ctx context.Context) bool {
	claims, ok := ClaimsFromContext(ctx)
	if !ok || claims == nil {
		return true
	}
	return claims.IsAnonymous
}