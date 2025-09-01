package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// MagicLinkProvider implements passwordless authentication via email
type MagicLinkProvider struct {
	userStore   UserStore
	tokenStore  MagicLinkTokenStore
	emailSender EmailSender
	config      MagicLinkConfig
}

// MagicLinkTokenStore handles magic link token storage
type MagicLinkTokenStore interface {
	// SaveToken stores a magic link token
	SaveToken(ctx context.Context, token *MagicLinkToken) error
	
	// GetToken retrieves and validates a token
	GetToken(ctx context.Context, token string) (*MagicLinkToken, error)
	
	// DeleteToken removes a used token
	DeleteToken(ctx context.Context, token string) error
	
	// DeleteExpiredTokens removes all expired tokens
	DeleteExpiredTokens(ctx context.Context) error
	
	// CountRecentAttempts counts recent token requests for rate limiting
	CountRecentAttempts(ctx context.Context, email string, since time.Time) (int, error)
}

// EmailSender handles sending emails
type EmailSender interface {
	SendMagicLink(ctx context.Context, email, token, linkURL string) error
}

// MagicLinkToken represents a magic link token
type MagicLinkToken struct {
	Token     string    `json:"token"`
	Email     string    `json:"email"`
	UserID    string    `json:"user_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

// MagicLinkRequest represents a magic link request
type MagicLinkRequest struct {
	Action string `json:"action"` // "send" or "verify"
	Email  string `json:"email,omitempty"`
	Token  string `json:"token,omitempty"`
}

// NewMagicLinkProvider creates a new magic link provider
func NewMagicLinkProvider(userStore UserStore, tokenStore MagicLinkTokenStore, emailSender EmailSender, config MagicLinkConfig) (*MagicLinkProvider, error) {
	if userStore == nil {
		return nil, errors.New("user store is required")
	}
	if tokenStore == nil {
		return nil, errors.New("token store is required")
	}
	if emailSender == nil {
		return nil, errors.New("email sender is required")
	}
	
	// Set defaults
	if config.TokenLength == 0 {
		config.TokenLength = 32
	}
	if config.TokenTTL == 0 {
		config.TokenTTL = 15 * time.Minute
	}
	if config.MaxAttemptsPerEmail == 0 {
		config.MaxAttemptsPerEmail = 5
	}
	if config.AttemptWindowTTL == 0 {
		config.AttemptWindowTTL = time.Hour
	}
	
	return &MagicLinkProvider{
		userStore:   userStore,
		tokenStore:  tokenStore,
		emailSender: emailSender,
		config:      config,
	}, nil
}

// Name returns the provider name
func (p *MagicLinkProvider) Name() string {
	return "magic_link"
}

// Type returns the provider type
func (p *MagicLinkProvider) Type() string {
	return "internal"
}

// Authenticate handles magic link authentication
func (p *MagicLinkProvider) Authenticate(ctx context.Context, credential string) (*UserInfo, error) {
	// Parse request
	var req MagicLinkRequest
	if err := json.Unmarshal([]byte(credential), &req); err != nil {
		return nil, &AuthError{
			Code:    ErrInvalidCredentials,
			Message: "invalid request format",
		}
	}
	
	switch req.Action {
	case "send":
		return nil, p.sendMagicLink(ctx, req.Email)
	case "verify":
		return p.verifyMagicLink(ctx, req.Token)
	default:
		return nil, &AuthError{
			Code:    ErrInvalidCredentials,
			Message: "invalid action",
		}
	}
}

// sendMagicLink sends a magic link to the specified email
func (p *MagicLinkProvider) sendMagicLink(ctx context.Context, email string) error {
	// Check rate limiting
	since := time.Now().Add(-p.config.AttemptWindowTTL)
	attempts, err := p.tokenStore.CountRecentAttempts(ctx, email, since)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	
	if attempts >= p.config.MaxAttemptsPerEmail {
		return &AuthError{
			Code:    "RATE_LIMITED",
			Message: fmt.Sprintf("too many attempts, please try again later"),
		}
	}
	
	// Check if user exists or create a new one
	user, err := p.userStore.GetUserByEmail(ctx, email)
	if err != nil {
		// Create a new user if they don't exist
		user = &User{
			Email:         email,
			Username:      email,
			EmailVerified: false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			Status:        "pending", // Pending until first login
			Metadata:      make(map[string]interface{}),
		}
		
		if err := p.userStore.CreateUser(ctx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}
	
	// Generate token
	tokenBytes := make([]byte, p.config.TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	tokenString := base64.URLEncoding.EncodeToString(tokenBytes)
	
	// Save token
	token := &MagicLinkToken{
		Token:     tokenString,
		Email:     email,
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(p.config.TokenTTL),
		Used:      false,
	}
	
	if err := p.tokenStore.SaveToken(ctx, token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}
	
	// Send email
	// TODO: Generate proper link URL based on configuration
	linkURL := fmt.Sprintf("https://example.com/auth/magic-link?token=%s", tokenString)
	if err := p.emailSender.SendMagicLink(ctx, email, tokenString, linkURL); err != nil {
		// Delete token if email fails
		_ = p.tokenStore.DeleteToken(ctx, tokenString)
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	// Return a special error to indicate success (email sent)
	return &AuthError{
		Code:    "MAGIC_LINK_SENT",
		Message: "magic link sent to email",
		Details: map[string]interface{}{
			"email": email,
		},
	}
}

// verifyMagicLink verifies a magic link token
func (p *MagicLinkProvider) verifyMagicLink(ctx context.Context, tokenString string) (*UserInfo, error) {
	// Get token
	token, err := p.tokenStore.GetToken(ctx, tokenString)
	if err != nil {
		return nil, &AuthError{
			Code:    ErrTokenInvalid,
			Message: "invalid or expired token",
		}
	}
	
	// Check if already used
	if token.Used {
		return nil, &AuthError{
			Code:    ErrTokenInvalid,
			Message: "token already used",
		}
	}
	
	// Check expiry
	if time.Now().After(token.ExpiresAt) {
		return nil, &AuthError{
			Code:    ErrTokenExpired,
			Message: "token expired",
		}
	}
	
	// Get user
	user, err := p.userStore.GetUserByEmail(ctx, token.Email)
	if err != nil {
		return nil, &AuthError{
			Code:    ErrUserNotFound,
			Message: "user not found",
		}
	}
	
	// Mark token as used
	now := time.Now()
	token.Used = true
	token.UsedAt = &now
	_ = p.tokenStore.SaveToken(ctx, token) // Update token
	
	// Delete the token after use
	_ = p.tokenStore.DeleteToken(ctx, tokenString)
	
	// Update user if needed
	if !user.EmailVerified {
		user.EmailVerified = true
		user.EmailVerifiedAt = &now
	}
	if user.Status == "pending" {
		user.Status = "active"
	}
	user.LastLoginAt = &now
	user.UpdatedAt = now
	
	if err := p.userStore.UpdateUser(ctx, user); err != nil {
		// Log error but continue
		fmt.Printf("failed to update user: %v\n", err)
	}
	
	// Return user info
	return &UserInfo{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Picture:       user.Picture,
		Provider:      "magic_link",
		EmailVerified: true,
		VerifiedAt:    now,
		Metadata:      user.Metadata,
	}, nil
}

// ValidateConfig checks if the provider is properly configured
func (p *MagicLinkProvider) ValidateConfig() error {
	if p.userStore == nil {
		return errors.New("user store is not configured")
	}
	if p.tokenStore == nil {
		return errors.New("token store is not configured")
	}
	if p.emailSender == nil {
		return errors.New("email sender is not configured")
	}
	if p.config.TokenLength < 16 {
		return errors.New("token length must be at least 16 bytes")
	}
	if p.config.TokenTTL < time.Minute {
		return errors.New("token TTL must be at least 1 minute")
	}
	return nil
}

// InMemoryMagicLinkTokenStore provides an in-memory implementation for testing
type InMemoryMagicLinkTokenStore struct {
	tokens map[string]*MagicLinkToken
}

// NewInMemoryMagicLinkTokenStore creates a new in-memory token store
func NewInMemoryMagicLinkTokenStore() *InMemoryMagicLinkTokenStore {
	return &InMemoryMagicLinkTokenStore{
		tokens: make(map[string]*MagicLinkToken),
	}
}

func (s *InMemoryMagicLinkTokenStore) SaveToken(ctx context.Context, token *MagicLinkToken) error {
	s.tokens[token.Token] = token
	return nil
}

func (s *InMemoryMagicLinkTokenStore) GetToken(ctx context.Context, token string) (*MagicLinkToken, error) {
	t, exists := s.tokens[token]
	if !exists {
		return nil, errors.New("token not found")
	}
	return t, nil
}

func (s *InMemoryMagicLinkTokenStore) DeleteToken(ctx context.Context, token string) error {
	delete(s.tokens, token)
	return nil
}

func (s *InMemoryMagicLinkTokenStore) DeleteExpiredTokens(ctx context.Context) error {
	now := time.Now()
	for token, t := range s.tokens {
		if t.ExpiresAt.Before(now) {
			delete(s.tokens, token)
		}
	}
	return nil
}

func (s *InMemoryMagicLinkTokenStore) CountRecentAttempts(ctx context.Context, email string, since time.Time) (int, error) {
	count := 0
	for _, t := range s.tokens {
		if t.Email == email && t.CreatedAt.After(since) {
			count++
		}
	}
	return count, nil
}