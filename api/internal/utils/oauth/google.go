package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	
	"github.com/radjathaher/alunalun/api/internal/utils/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleProvider implements OAuth authentication for Google
type GoogleProvider struct {
	BaseProvider
	httpClient *http.Client
}

// GoogleUserInfo represents the user info returned by Google
type GoogleUserInfo struct {
	Sub           string `json:"sub"`           // Unique Google ID
	Email         string `json:"email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
	Locale        string `json:"locale"`
	HD            string `json:"hd"` // Hosted domain (for G Suite)
}

// NewGoogleProvider creates a new Google OAuth provider
func NewGoogleProvider(clientID, clientSecret, redirectURL string) (*GoogleProvider, error) {
	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil, errors.New("clientID, clientSecret, and redirectURL are required")
	}
	
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	
	return &GoogleProvider{
		BaseProvider: BaseProvider{
			Name:   "google",
			Config: config,
		},
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Name returns the provider name
func (p *GoogleProvider) Name() string {
	return "google"
}

// Authenticate handles Google OAuth authentication
func (p *GoogleProvider) Authenticate(ctx context.Context, credential string) (*auth.UserInfo, error) {
	// Credential can be either an authorization code or an ID token
	// Try to determine which one it is
	
	// If it looks like a JWT (has three parts separated by dots), treat as ID token
	if strings.Count(credential, ".") == 2 {
		return p.VerifyIDToken(ctx, credential)
	}
	
	// Otherwise, treat as authorization code
	token, err := p.ExchangeCode(ctx, credential)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	
	// Get user info using the access token
	return p.getUserInfo(ctx, token.AccessToken)
}

// VerifyIDToken verifies a Google ID token (for mobile/SPA clients)
func (p *GoogleProvider) VerifyIDToken(ctx context.Context, idToken string) (*auth.UserInfo, error) {
	// For production, you should verify the ID token properly
	// This is a simplified version that decodes the token and fetches user info
	// In production, use Google's token verification endpoint or library
	
	// For now, we'll use the tokeninfo endpoint (not recommended for production)
	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token verification failed: %s", string(body))
	}
	
	var tokenInfo struct {
		Aud           string `json:"aud"`
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to decode token info: %w", err)
	}
	
	// Verify the audience matches our client ID
	if tokenInfo.Aud != p.Config.ClientID {
		return nil, errors.New("invalid token audience")
	}
	
	return &auth.UserInfo{
		ID:            tokenInfo.Sub,
		Email:         tokenInfo.Email,
		Username:      tokenInfo.Email, // Use email as username for OAuth users
		FirstName:     tokenInfo.GivenName,
		LastName:      tokenInfo.FamilyName,
		FullName:      tokenInfo.Name,
		Picture:       tokenInfo.Picture,
		Provider:      "google",
		ProviderID:    tokenInfo.Sub,
		EmailVerified: tokenInfo.EmailVerified == "true",
		VerifiedAt:    time.Now(),
		Metadata: map[string]interface{}{
			"google_sub": tokenInfo.Sub,
		},
	}, nil
}

// getUserInfo fetches user information from Google using an access token
func (p *GoogleProvider) getUserInfo(ctx context.Context, accessToken string) (*auth.UserInfo, error) {
	url := "https://www.googleapis.com/oauth2/v3/userinfo"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+accessToken)
	
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}
	
	var googleUser GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}
	
	// Convert Google user info to our standard format
	return &auth.UserInfo{
		ID:            googleUser.Sub,
		Email:         googleUser.Email,
		Username:      googleUser.Email, // Use email as username for OAuth users
		FirstName:     googleUser.GivenName,
		LastName:      googleUser.FamilyName,
		FullName:      googleUser.Name,
		Picture:       googleUser.Picture,
		Provider:      "google",
		ProviderID:    googleUser.Sub,
		EmailVerified: googleUser.EmailVerified,
		VerifiedAt:    time.Now(),
		Metadata: map[string]interface{}{
			"google_sub": googleUser.Sub,
			"locale":     googleUser.Locale,
			"hd":         googleUser.HD,
		},
	}, nil
}

// GetAuthURLWithOptions generates an auth URL with additional options
func (p *GoogleProvider) GetAuthURLWithOptions(state string, opts ...oauth2.AuthCodeOption) string {
	// Default options
	options := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline, // Request refresh token
	}
	
	// Add any additional options
	options = append(options, opts...)
	
	return p.Config.AuthCodeURL(state, options...)
}

// RefreshToken refreshes an OAuth token
func (p *GoogleProvider) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	if refreshToken == "" {
		return nil, errors.New("refresh token is required")
	}
	
	// Create a token source with the refresh token
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}
	
	tokenSource := p.Config.TokenSource(ctx, token)
	
	// This will automatically refresh the token
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	
	return newToken, nil
}