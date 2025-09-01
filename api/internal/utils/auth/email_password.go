package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode"
	
	"golang.org/x/crypto/bcrypt"
)

// EmailPasswordProvider implements email/password authentication
type EmailPasswordProvider struct {
	userStore  UserStore
	config     EmailProviderConfig
	emailRegex *regexp.Regexp
}

// UserStore defines the interface for user storage operations
type UserStore interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *User) error
	
	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	
	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, userID string) (*User, error)
	
	// UpdateUser updates user information
	UpdateUser(ctx context.Context, user *User) error
	
	// CheckUsernameAvailable checks if a username is available
	CheckUsernameAvailable(ctx context.Context, username string) (bool, error)
}

// User represents a user in the system
type User struct {
	ID               string                 `json:"id"`
	Email            string                 `json:"email"`
	Username         string                 `json:"username"`
	PasswordHash     string                 `json:"-"`
	FirstName        string                 `json:"first_name,omitempty"`
	LastName         string                 `json:"last_name,omitempty"`
	Picture          string                 `json:"picture,omitempty"`
	EmailVerified    bool                   `json:"email_verified"`
	EmailVerifiedAt  *time.Time             `json:"email_verified_at,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	LastLoginAt      *time.Time             `json:"last_login_at,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Status           string                 `json:"status"` // active, disabled, pending
}

// EmailPasswordCredentials represents email/password login credentials
type EmailPasswordCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// NewEmailPasswordProvider creates a new email/password provider
func NewEmailPasswordProvider(userStore UserStore, config EmailProviderConfig) (*EmailPasswordProvider, error) {
	if userStore == nil {
		return nil, errors.New("user store is required")
	}
	
	// Set defaults if not provided
	if config.MinLength == 0 {
		config.MinLength = 8
	}
	if config.BcryptCost == 0 {
		config.BcryptCost = 12
	}
	
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	
	return &EmailPasswordProvider{
		userStore:  userStore,
		config:     config,
		emailRegex: emailRegex,
	}, nil
}

// Name returns the provider name
func (p *EmailPasswordProvider) Name() string {
	return "email"
}

// Type returns the provider type
func (p *EmailPasswordProvider) Type() string {
	return "internal"
}

// Authenticate validates email/password credentials
func (p *EmailPasswordProvider) Authenticate(ctx context.Context, credential string) (*UserInfo, error) {
	// Parse credentials
	var creds EmailPasswordCredentials
	if err := json.Unmarshal([]byte(credential), &creds); err != nil {
		return nil, &AuthError{
			Code:    ErrInvalidCredentials,
			Message: "invalid credential format",
		}
	}
	
	// Validate email
	if !p.emailRegex.MatchString(creds.Email) {
		return nil, &AuthError{
			Code:    ErrInvalidCredentials,
			Message: "invalid email format",
		}
	}
	
	// Get user by email
	user, err := p.userStore.GetUserByEmail(ctx, creds.Email)
	if err != nil {
		// Don't reveal whether the email exists
		return nil, &AuthError{
			Code:    ErrInvalidCredentials,
			Message: "invalid email or password",
		}
	}
	
	// Check user status
	if user.Status != "active" {
		return nil, &AuthError{
			Code:    ErrUserDisabled,
			Message: "user account is disabled",
		}
	}
	
	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		return nil, &AuthError{
			Code:    ErrInvalidCredentials,
			Message: "invalid email or password",
		}
	}
	
	// Check email verification if required
	if p.config.RequireVerification && !user.EmailVerified {
		return nil, &AuthError{
			Code:    ErrEmailNotVerified,
			Message: "email not verified",
		}
	}
	
	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	_ = p.userStore.UpdateUser(ctx, user) // Ignore error, not critical
	
	// Convert to UserInfo
	return &UserInfo{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Picture:       user.Picture,
		Provider:      "email",
		EmailVerified: user.EmailVerified,
		VerifiedAt:    func() time.Time {
			if user.EmailVerifiedAt != nil {
				return *user.EmailVerifiedAt
			}
			return time.Time{}
		}(),
		Metadata: user.Metadata,
	}, nil
}

// ValidateConfig checks if the provider is properly configured
func (p *EmailPasswordProvider) ValidateConfig() error {
	if p.userStore == nil {
		return errors.New("user store is not configured")
	}
	if p.config.BcryptCost < 10 || p.config.BcryptCost > 31 {
		return errors.New("bcrypt cost must be between 10 and 31")
	}
	if p.config.MinLength < 6 {
		return errors.New("minimum password length must be at least 6")
	}
	return nil
}

// RegisterUser creates a new user with email/password
func (p *EmailPasswordProvider) RegisterUser(ctx context.Context, email, password, username string) (*User, error) {
	// Validate email
	if !p.emailRegex.MatchString(email) {
		return nil, errors.New("invalid email format")
	}
	
	// Validate password
	if err := p.validatePassword(password); err != nil {
		return nil, err
	}
	
	// Check if email already exists
	existingUser, _ := p.userStore.GetUserByEmail(ctx, email)
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}
	
	// Check username availability if provided
	if username != "" {
		available, err := p.userStore.CheckUsernameAvailable(ctx, username)
		if err != nil {
			return nil, fmt.Errorf("failed to check username: %w", err)
		}
		if !available {
			return nil, errors.New("username not available")
		}
	} else {
		// Use email as username if not provided
		username = email
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), p.config.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Create user
	now := time.Now()
	user := &User{
		Email:         email,
		Username:      username,
		PasswordHash:  string(hashedPassword),
		EmailVerified: !p.config.RequireVerification, // If verification not required, mark as verified
		CreatedAt:     now,
		UpdatedAt:     now,
		Status:        "active",
		Metadata:      make(map[string]interface{}),
	}
	
	if !p.config.RequireVerification {
		user.EmailVerifiedAt = &now
	}
	
	// Save user
	if err := p.userStore.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return user, nil
}

// UpdatePassword updates a user's password
func (p *EmailPasswordProvider) UpdatePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// Get user
	user, err := p.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	
	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}
	
	// Validate new password
	if err := p.validatePassword(newPassword); err != nil {
		return err
	}
	
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), p.config.BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Update user
	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()
	
	if err := p.userStore.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	return nil
}

// validatePassword checks if a password meets requirements
func (p *EmailPasswordProvider) validatePassword(password string) error {
	if len(password) < p.config.MinLength {
		return fmt.Errorf("password must be at least %d characters long", p.config.MinLength)
	}
	
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	if p.config.RequireUppercase && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	
	if p.config.RequireLowercase && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	
	if p.config.RequireNumbers && !hasNumber {
		return errors.New("password must contain at least one number")
	}
	
	if p.config.RequireSpecialChar && !hasSpecial {
		return errors.New("password must contain at least one special character")
	}
	
	return nil
}