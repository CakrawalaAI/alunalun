package auth

import "time"

// Config holds the authentication system configuration
type Config struct {
	JWT       JWTConfig                  `json:"jwt"`
	Session   SessionConfig              `json:"session"`
	Providers map[string]ProviderConfig  `json:"providers"`
}

// JWTConfig holds JWT-specific configuration
type JWTConfig struct {
	// RSA key paths or PEM content
	PrivateKey string `json:"private_key" env:"JWT_PRIVATE_KEY"`
	PublicKey  string `json:"public_key" env:"JWT_PUBLIC_KEY"`
	
	// Token settings
	Issuer   string        `json:"issuer" env:"JWT_ISSUER" default:"alunalun"`
	Audience string        `json:"audience" env:"JWT_AUDIENCE" default:"alunalun-web"`
	
	// Expiry durations
	AccessTokenTTL  time.Duration `json:"access_token_ttl" env:"JWT_ACCESS_TTL" default:"1h"`
	RefreshTokenTTL time.Duration `json:"refresh_token_ttl" env:"JWT_REFRESH_TTL" default:"168h"` // 7 days
}

// SessionConfig holds session-specific configuration
type SessionConfig struct {
	// Session TTL for authenticated users
	TTL time.Duration `json:"ttl" env:"SESSION_TTL" default:"1h"`
	
	// Cleanup interval for expired sessions
	CleanupInterval time.Duration `json:"cleanup_interval" env:"SESSION_CLEANUP_INTERVAL" default:"1h"`
	
	// Maximum sessions per user (0 = unlimited)
	MaxPerUser int `json:"max_per_user" env:"SESSION_MAX_PER_USER" default:"10"`
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
	// Provider type (oauth_google, oauth_apple, email_password, magic_link, anonymous)
	Type    string            `json:"type"`
	
	// Whether the provider is enabled
	Enabled bool              `json:"enabled"`
	
	// Provider-specific configuration
	Config  map[string]string `json:"config"`
}

// OAuthProviderConfig represents OAuth-specific configuration
type OAuthProviderConfig struct {
	ClientID     string   `json:"client_id" env:"OAUTH_CLIENT_ID"`
	ClientSecret string   `json:"client_secret" env:"OAUTH_CLIENT_SECRET"`
	RedirectURL  string   `json:"redirect_url" env:"OAUTH_REDIRECT_URL"`
	Scopes       []string `json:"scopes"`
	
	// Optional: custom auth/token URLs
	AuthURL  string `json:"auth_url,omitempty"`
	TokenURL string `json:"token_url,omitempty"`
}

// EmailProviderConfig represents email/password provider configuration
type EmailProviderConfig struct {
	// Password requirements
	MinLength          int  `json:"min_length" default:"8"`
	RequireUppercase   bool `json:"require_uppercase" default:"true"`
	RequireLowercase   bool `json:"require_lowercase" default:"true"`
	RequireNumbers     bool `json:"require_numbers" default:"true"`
	RequireSpecialChar bool `json:"require_special_char" default:"false"`
	
	// Bcrypt cost factor (10-31)
	BcryptCost int `json:"bcrypt_cost" default:"12"`
	
	// Email verification
	RequireVerification bool          `json:"require_verification" default:"true"`
	VerificationTTL     time.Duration `json:"verification_ttl" default:"24h"`
}

// MagicLinkConfig represents magic link provider configuration
type MagicLinkConfig struct {
	// Token settings
	TokenLength int           `json:"token_length" default:"32"`
	TokenTTL    time.Duration `json:"token_ttl" default:"15m"`
	
	// Rate limiting
	MaxAttemptsPerEmail int           `json:"max_attempts_per_email" default:"5"`
	AttemptWindowTTL    time.Duration `json:"attempt_window_ttl" default:"1h"`
	
	// Email settings
	FromEmail string `json:"from_email" env:"MAGIC_LINK_FROM_EMAIL"`
	FromName  string `json:"from_name" env:"MAGIC_LINK_FROM_NAME" default:"Alunalun"`
	Subject   string `json:"subject" default:"Your login link"`
	Template  string `json:"template"` // Email template path or inline template
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		JWT: JWTConfig{
			Issuer:          "alunalun",
			Audience:        "alunalun-web",
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
		Session: SessionConfig{
			TTL:             time.Hour,
			CleanupInterval: time.Hour,
			MaxPerUser:      10,
		},
		Providers: map[string]ProviderConfig{
			"google": {
				Type:    "oauth",
				Enabled: false,
				Config:  map[string]string{},
			},
			"email": {
				Type:    "internal",
				Enabled: true,
				Config: map[string]string{
					"min_length":      "8",
					"bcrypt_cost":     "12",
					"require_verify":  "true",
				},
			},
			"magic_link": {
				Type:    "internal",
				Enabled: false,
				Config: map[string]string{
					"token_ttl": "15m",
				},
			},
			"anonymous": {
				Type:    "internal",
				Enabled: true,
				Config:  map[string]string{},
			},
		},
	}
}