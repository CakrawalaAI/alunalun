package auth

import (
	"context"
	"time"
)

// Provider defines the interface for all authentication providers
type Provider interface {
	// Name returns the provider identifier (e.g., "google", "email", "magic_link")
	Name() string
	
	// Type returns the provider type (e.g., "oauth", "internal")
	Type() string
	
	// Authenticate handles the authentication flow for this provider
	// The credential format varies by provider:
	// - OAuth: authorization code or ID token
	// - Email/Password: JSON with email and password
	// - Magic Link: magic token
	Authenticate(ctx context.Context, credential string) (*UserInfo, error)
	
	// ValidateConfig checks if the provider is properly configured
	ValidateConfig() error
}

// UserInfo represents the authenticated user information
type UserInfo struct {
	// Core user identifiers
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Username      string    `json:"username"`
	
	// Profile information
	FirstName     string    `json:"first_name,omitempty"`
	LastName      string    `json:"last_name,omitempty"`
	FullName      string    `json:"full_name,omitempty"`
	Picture       string    `json:"picture,omitempty"`
	
	// Auth metadata
	Provider      string    `json:"provider"`
	ProviderID    string    `json:"provider_id,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	VerifiedAt    time.Time `json:"verified_at,omitempty"`
	
	// Additional provider-specific data
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Session represents a user session (anonymous or authenticated)
type Session struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id,omitempty"`
	Username    string    `json:"username,omitempty"`
	IsAnonymous bool      `json:"is_anonymous"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// AuthError represents authentication-specific errors
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *AuthError) Error() string {
	return e.Message
}

// Common auth error codes
const (
	ErrInvalidCredentials = "INVALID_CREDENTIALS"
	ErrUserNotFound       = "USER_NOT_FOUND"
	ErrUserDisabled       = "USER_DISABLED"
	ErrProviderError      = "PROVIDER_ERROR"
	ErrTokenExpired       = "TOKEN_EXPIRED"
	ErrTokenInvalid       = "TOKEN_INVALID"
	ErrSessionNotFound    = "SESSION_NOT_FOUND"
	ErrEmailNotVerified   = "EMAIL_NOT_VERIFIED"
)