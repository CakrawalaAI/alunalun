package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// AnonymousProvider implements anonymous authentication
type AnonymousProvider struct {
	sessionManager *SessionManager
	userStore      UserStore
}

// AnonymousRequest represents an anonymous authentication request
type AnonymousRequest struct {
	Username string `json:"username"`
}

// NewAnonymousProvider creates a new anonymous provider
func NewAnonymousProvider(sessionManager *SessionManager, userStore UserStore) (*AnonymousProvider, error) {
	if sessionManager == nil {
		return nil, errors.New("session manager is required")
	}
	
	return &AnonymousProvider{
		sessionManager: sessionManager,
		userStore:      userStore,
	}, nil
}

// Name returns the provider name
func (p *AnonymousProvider) Name() string {
	return "anonymous"
}

// Type returns the provider type
func (p *AnonymousProvider) Type() string {
	return "internal"
}

// Authenticate creates an anonymous session
func (p *AnonymousProvider) Authenticate(ctx context.Context, credential string) (*UserInfo, error) {
	// Parse request
	var req AnonymousRequest
	if err := json.Unmarshal([]byte(credential), &req); err != nil {
		return nil, &AuthError{
			Code:    ErrInvalidCredentials,
			Message: "invalid request format",
		}
	}
	
	if req.Username == "" {
		return nil, &AuthError{
			Code:    ErrInvalidCredentials,
			Message: "username is required",
		}
	}
	
	// Check username availability if userStore is available
	if p.userStore != nil {
		available, err := p.userStore.CheckUsernameAvailable(ctx, req.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to check username: %w", err)
		}
		if !available {
			return nil, &AuthError{
				Code:    "USERNAME_TAKEN",
				Message: "username is not available",
			}
		}
	}
	
	// Create anonymous session
	session, err := p.sessionManager.CreateAnonymous(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to create anonymous session: %w", err)
	}
	
	// Return session info as UserInfo
	return &UserInfo{
		ID:       session.ID,      // Use session ID as user ID for anonymous
		Username: session.Username,
		Provider: "anonymous",
		Metadata: map[string]interface{}{
			"session_id":   session.ID,
			"is_anonymous": true,
			"created_at":   session.CreatedAt,
		},
	}, nil
}

// ValidateConfig checks if the provider is properly configured
func (p *AnonymousProvider) ValidateConfig() error {
	if p.sessionManager == nil {
		return errors.New("session manager is not configured")
	}
	return nil
}

// GetSession retrieves an anonymous session by ID
func (p *AnonymousProvider) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	return p.sessionManager.Validate(ctx, sessionID)
}

// MigrateToUser converts an anonymous session to an authenticated user session
func (p *AnonymousProvider) MigrateToUser(ctx context.Context, sessionID, userID string) error {
	return p.sessionManager.MigrateToUser(ctx, sessionID, userID)
}

// RevokeSession deletes an anonymous session
func (p *AnonymousProvider) RevokeSession(ctx context.Context, sessionID string) error {
	return p.sessionManager.Revoke(ctx, sessionID)
}

// CleanupExpiredSessions removes expired anonymous sessions
func (p *AnonymousProvider) CleanupExpiredSessions(ctx context.Context) error {
	return p.sessionManager.CleanupExpired(ctx)
}

// AnonymousUserInfo provides additional anonymous-specific information
type AnonymousUserInfo struct {
	*UserInfo
	SessionID   string    `json:"session_id"`
	IsAnonymous bool      `json:"is_anonymous"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// ToAnonymousUserInfo converts UserInfo to AnonymousUserInfo
func ToAnonymousUserInfo(info *UserInfo) *AnonymousUserInfo {
	sessionID, _ := info.Metadata["session_id"].(string)
	isAnonymous, _ := info.Metadata["is_anonymous"].(bool)
	createdAt, _ := info.Metadata["created_at"].(time.Time)
	
	return &AnonymousUserInfo{
		UserInfo:    info,
		SessionID:   sessionID,
		IsAnonymous: isAnonymous,
		CreatedAt:   createdAt,
	}
}