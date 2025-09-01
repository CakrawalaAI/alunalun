package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	
	"github.com/radjathaher/alunalun/api/internal/utils/auth"
	"github.com/radjathaher/alunalun/api/internal/utils/oauth"
)

// OAuthHandler handles HTTP OAuth endpoints
type OAuthHandler struct {
	service       *Service
	stateManager  *auth.StateManager
	registry      *auth.ProviderRegistry
	tokenManager  *auth.TokenManager
	sessionManager *auth.SessionManager
}

// NewOAuthHandler creates a new OAuth HTTP handler
func NewOAuthHandler(
	service *Service,
	stateManager *auth.StateManager,
	registry *auth.ProviderRegistry,
	tokenManager *auth.TokenManager,
	sessionManager *auth.SessionManager,
) *OAuthHandler {
	return &OAuthHandler{
		service:       service,
		stateManager:  stateManager,
		registry:      registry,
		tokenManager:  tokenManager,
		sessionManager: sessionManager,
	}
}

// RegisterRoutes registers OAuth HTTP routes
func (h *OAuthHandler) RegisterRoutes(mux *http.ServeMux) {
	// OAuth initiation endpoints
	mux.HandleFunc("/auth/oauth/", h.handleOAuthRoot)
	mux.HandleFunc("/auth/oauth/google", h.handleOAuthInitiate)
	mux.HandleFunc("/auth/oauth/apple", h.handleOAuthInitiate)
	
	// OAuth callback endpoints
	mux.HandleFunc("/auth/oauth/google/callback", h.handleOAuthCallback)
	mux.HandleFunc("/auth/oauth/apple/callback", h.handleOAuthCallback)
	
	// Generic endpoints
	mux.HandleFunc("/auth/refresh", h.handleRefreshToken)
	mux.HandleFunc("/auth/public-key", h.handlePublicKey)
}

// handleOAuthRoot provides OAuth endpoint information
func (h *OAuthHandler) handleOAuthRoot(w http.ResponseWriter, r *http.Request) {
	providers := h.registry.GetProviderInfo()
	oauthProviders := []string{}
	
	for _, p := range providers {
		if p.Type == "oauth" {
			oauthProviders = append(oauthProviders, p.Name)
		}
	}
	
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"providers": oauthProviders,
		"endpoints": map[string]string{
			"initiate": "/auth/oauth/{provider}",
			"callback": "/auth/oauth/{provider}/callback",
			"refresh":  "/auth/refresh",
		},
	})
}

// handleOAuthInitiate starts the OAuth flow
func (h *OAuthHandler) handleOAuthInitiate(w http.ResponseWriter, r *http.Request) {
	// Extract provider from URL path
	provider := h.extractProvider(r.URL.Path)
	if provider == "" {
		h.respondError(w, http.StatusBadRequest, "invalid provider")
		return
	}
	
	// Get redirect URI from query params
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		h.respondError(w, http.StatusBadRequest, "redirect_uri is required")
		return
	}
	
	// Optional session ID for migration
	sessionID := r.URL.Query().Get("session_id")
	
	// Get the OAuth provider
	p, err := h.registry.Get(provider)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "provider not found")
		return
	}
	
	// Check if it's an OAuth provider
	oauthProvider, ok := p.(oauth.Provider)
	if !ok {
		h.respondError(w, http.StatusBadRequest, "not an OAuth provider")
		return
	}
	
	// Generate encrypted state
	state, err := h.stateManager.GenerateState(provider, redirectURI, sessionID)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to generate state")
		return
	}
	
	// Get OAuth authorization URL
	authURL := oauthProvider.GetAuthURL(state)
	
	// Redirect to OAuth provider
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// handleOAuthCallback handles the OAuth provider callback
func (h *OAuthHandler) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Get state and code from query params
	stateParam := r.URL.Query().Get("state")
	if stateParam == "" {
		h.respondError(w, http.StatusBadRequest, "state parameter missing")
		return
	}
	
	code := r.URL.Query().Get("code")
	if code == "" {
		// Check for error from OAuth provider
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			errDesc := r.URL.Query().Get("error_description")
			h.respondError(w, http.StatusUnauthorized, fmt.Sprintf("OAuth error: %s - %s", errParam, errDesc))
			return
		}
		h.respondError(w, http.StatusBadRequest, "authorization code missing")
		return
	}
	
	// Validate and decrypt state
	state, err := h.stateManager.ValidateState(stateParam)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid or expired state")
		return
	}
	
	// Get the OAuth provider
	p, err := h.registry.Get(state.Provider)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "provider not found")
		return
	}
	
	oauthProvider, ok := p.(oauth.Provider)
	if !ok {
		h.respondError(w, http.StatusInternalServerError, "invalid provider type")
		return
	}
	
	// Exchange code for tokens
	ctx := context.Background()
	token, err := oauthProvider.ExchangeCode(ctx, code)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, fmt.Sprintf("code exchange failed: %v", err))
		return
	}
	
	// Get user info using the access token
	userInfo, err := oauthProvider.Authenticate(ctx, token.AccessToken)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get user info: %v", err))
		return
	}
	
	// Find or create user
	user, err := h.service.findOrCreateUser(ctx, userInfo)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to process user")
		return
	}
	
	// Handle session migration if session ID was provided
	sessionMigrated := false
	if state.SessionID != "" {
		session, err := h.sessionManager.Validate(ctx, state.SessionID)
		if err == nil && session.IsAnonymous {
			if err := h.sessionManager.MigrateToUser(ctx, state.SessionID, user.ID); err == nil {
				sessionMigrated = true
			}
		}
	}
	
	// Create authenticated session
	session, err := h.sessionManager.CreateAuthenticated(ctx, user.ID, time.Hour)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	
	// Generate JWT token
	claims := &auth.Claims{
		UserID:      user.ID,
		SessionID:   session.ID,
		Username:    user.Username,
		Email:       user.Email,
		Provider:    state.Provider,
		IsAnonymous: false,
	}
	
	jwtToken, err := h.tokenManager.GenerateToken(claims, time.Hour)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	
	// Detect client type and respond accordingly
	if h.isMobileClient(r) {
		// Mobile: Return JSON response
		h.respondJSON(w, http.StatusOK, map[string]interface{}{
			"token":            jwtToken,
			"user":             h.userToJSON(user),
			"session_migrated": sessionMigrated,
		})
	} else {
		// Web: Redirect with token
		// Using fragment (#) to keep token client-side only
		redirectURL := fmt.Sprintf("%s#token=%s", state.RedirectURI, jwtToken)
		if sessionMigrated {
			redirectURL += "&session_migrated=true"
		}
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// handleRefreshToken handles JWT refresh requests
func (h *OAuthHandler) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == "OPTIONS" {
		h.setCORSHeaders(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "POST" {
		h.respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	
	var req struct {
		ExpiredToken string `json:"expired_token"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	
	if req.ExpiredToken == "" {
		h.respondError(w, http.StatusBadRequest, "expired_token is required")
		return
	}
	
	// Refresh the token
	newToken, err := h.tokenManager.RefreshToken(req.ExpiredToken, time.Hour)
	if err != nil {
		h.respondError(w, http.StatusUnauthorized, fmt.Sprintf("refresh failed: %v", err))
		return
	}
	
	h.respondJSON(w, http.StatusOK, map[string]string{
		"token": newToken,
	})
}

// handlePublicKey returns the JWT public key
func (h *OAuthHandler) handlePublicKey(w http.ResponseWriter, r *http.Request) {
	publicKeyPEM, err := h.tokenManager.GetPublicKeyPEM()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to get public key")
		return
	}
	
	h.respondJSON(w, http.StatusOK, map[string]string{
		"public_key": string(publicKeyPEM),
		"algorithm":  "RS256",
	})
}

// Helper methods

// extractProvider extracts provider name from URL path
func (h *OAuthHandler) extractProvider(path string) string {
	// Path format: /auth/oauth/{provider} or /auth/oauth/{provider}/callback
	parts := strings.Split(strings.TrimPrefix(path, "/auth/oauth/"), "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

// isMobileClient detects if the request is from a mobile client
func (h *OAuthHandler) isMobileClient(r *http.Request) bool {
	// Check custom header first
	if r.Header.Get("X-Client-Type") == "mobile" {
		return true
	}
	
	// Check User-Agent for mobile patterns
	userAgent := strings.ToLower(r.Header.Get("User-Agent"))
	mobilePatterns := []string{
		"ios", "iphone", "ipad",
		"android",
		"mobile",
		"react-native",
		"flutter",
	}
	
	for _, pattern := range mobilePatterns {
		if strings.Contains(userAgent, pattern) {
			return true
		}
	}
	
	return false
}

// setCORSHeaders sets CORS headers for cross-origin requests
func (h *OAuthHandler) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Client-Type")
	w.Header().Set("Access-Control-Max-Age", "3600")
}

// respondJSON sends a JSON response
func (h *OAuthHandler) respondJSON(w http.ResponseWriter, code int, data interface{}) {
	h.setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func (h *OAuthHandler) respondError(w http.ResponseWriter, code int, message string) {
	h.respondJSON(w, code, map[string]string{
		"error": message,
	})
}

// userToJSON converts User to JSON-friendly format
func (h *OAuthHandler) userToJSON(user *auth.User) map[string]interface{} {
	return map[string]interface{}{
		"id":              user.ID,
		"email":           user.Email,
		"username":        user.Username,
		"first_name":      user.FirstName,
		"last_name":       user.LastName,
		"picture":         user.Picture,
		"email_verified":  user.EmailVerified,
		"status":          user.Status,
	}
}