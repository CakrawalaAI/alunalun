package oauth

import (
	"context"
	"errors"
	
	"github.com/radjathaher/alunalun/api/internal/utils/auth"
	"golang.org/x/oauth2"
)

// Provider extends the base auth.Provider interface with OAuth-specific methods
type Provider interface {
	auth.Provider
	
	// GetAuthURL generates the OAuth authorization URL
	GetAuthURL(state string) string
	
	// ExchangeCode exchanges an authorization code for tokens
	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
	
	// VerifyIDToken verifies an ID token (for mobile/SPA clients)
	VerifyIDToken(ctx context.Context, idToken string) (*auth.UserInfo, error)
	
	// GetOAuth2Config returns the underlying OAuth2 configuration
	GetOAuth2Config() *oauth2.Config
}

// BaseProvider provides common OAuth functionality
type BaseProvider struct {
	Name   string
	Config *oauth2.Config
}

// GetName returns the provider name
func (p *BaseProvider) GetName() string {
	return p.Name
}

// Type returns "oauth" for all OAuth providers
func (p *BaseProvider) Type() string {
	return "oauth"
}

// GetAuthURL generates the OAuth authorization URL
func (p *BaseProvider) GetAuthURL(state string) string {
	return p.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges an authorization code for tokens
func (p *BaseProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	if code == "" {
		return nil, errors.New("authorization code is required")
	}
	return p.Config.Exchange(ctx, code)
}

// GetOAuth2Config returns the OAuth2 configuration
func (p *BaseProvider) GetOAuth2Config() *oauth2.Config {
	return p.Config
}

// ValidateConfig checks if the OAuth provider is properly configured
func (p *BaseProvider) ValidateConfig() error {
	if p.Config == nil {
		return errors.New("OAuth2 config is nil")
	}
	if p.Config.ClientID == "" {
		return errors.New("client ID is required")
	}
	if p.Config.ClientSecret == "" {
		return errors.New("client secret is required")
	}
	if p.Config.RedirectURL == "" {
		return errors.New("redirect URL is required")
	}
	if len(p.Config.Scopes) == 0 {
		return errors.New("at least one scope is required")
	}
	return nil
}

// ProviderType represents the type of OAuth provider
type ProviderType string

const (
	ProviderTypeGoogle ProviderType = "google"
	ProviderTypeApple  ProviderType = "apple"
	ProviderTypeGitHub ProviderType = "github"
)

// NewProvider creates a new OAuth provider based on the type
func NewProvider(providerType ProviderType, config map[string]string) (Provider, error) {
	switch providerType {
	case ProviderTypeGoogle:
		return NewGoogleProvider(
			config["client_id"],
			config["client_secret"],
			config["redirect_url"],
		)
	// Add more providers as needed
	// case ProviderTypeApple:
	//     return NewAppleProvider(config)
	// case ProviderTypeGitHub:
	//     return NewGitHubProvider(config)
	default:
		return nil, errors.New("unsupported OAuth provider type")
	}
}