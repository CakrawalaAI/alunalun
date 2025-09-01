package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"
	
	authService "github.com/radjathaher/alunalun/api/internal/services/auth"
	"github.com/radjathaher/alunalun/api/internal/utils/auth"
	"github.com/radjathaher/alunalun/api/internal/utils/oauth"
)

func main() {
	// Configuration from environment
	config := &auth.Config{
		JWT: auth.JWTConfig{
			Issuer:          getEnv("JWT_ISSUER", "alunalun"),
			Audience:        getEnv("JWT_AUDIENCE", "web"),
			AccessTokenTTL:  time.Hour,     // 1 hour
			RefreshTokenTTL: 7 * 24 * time.Hour, // 7 days
		},
		Session: auth.SessionConfig{
			TTL:             time.Hour, // 1 hour
			CleanupInterval: time.Hour,
			MaxPerUser:      10,
		},
		Providers: map[string]auth.ProviderConfig{
			"google": {
				Type:    "oauth",
				Enabled: true,
				Config: map[string]string{
					"client_id":     os.Getenv("GOOGLE_CLIENT_ID"),
					"client_secret": os.Getenv("GOOGLE_CLIENT_SECRET"),
					"redirect_url":  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/oauth/google/callback"),
				},
			},
			"email": {
				Type:    "internal",
				Enabled: true,
				Config: map[string]string{
					"min_length":     "8",
					"bcrypt_cost":    "12",
					"require_verify": "false",
				},
			},
			"anonymous": {
				Type:    "internal",
				Enabled: true,
				Config:  map[string]string{},
			},
		},
	}
	
	// Generate or load RSA keys
	var privateKey, publicKey []byte
	if os.Getenv("JWT_PRIVATE_KEY_PATH") != "" {
		// Load from files
		var err error
		privateKey, err = os.ReadFile(os.Getenv("JWT_PRIVATE_KEY_PATH"))
		if err != nil {
			log.Fatal("Failed to read private key:", err)
		}
		publicKey, err = os.ReadFile(os.Getenv("JWT_PUBLIC_KEY_PATH"))
		if err != nil {
			log.Fatal("Failed to read public key:", err)
		}
	} else {
		// Generate for development
		var err error
		privateKey, publicKey, err = auth.GenerateKeyPair()
		if err != nil {
			log.Fatal("Failed to generate key pair:", err)
		}
		log.Println("Generated RSA key pair for development")
	}
	
	// Create token manager
	tokenManager, err := auth.NewTokenManager(privateKey, publicKey, config.JWT.Issuer, config.JWT.Audience)
	if err != nil {
		log.Fatal("Failed to create token manager:", err)
	}
	
	// Create session manager with in-memory store (use Redis in production)
	sessionStore := auth.NewInMemorySessionStore()
	sessionManager := auth.NewSessionManager(sessionStore)
	
	// Create state manager for OAuth
	var stateKey []byte
	if envKey := os.Getenv("OAUTH_STATE_KEY"); envKey != "" {
		stateKey, err = base64.StdEncoding.DecodeString(envKey)
		if err != nil {
			log.Fatal("Invalid OAUTH_STATE_KEY:", err)
		}
	} else {
		// Generate for development
		stateKey, err = auth.GenerateEncryptionKey()
		if err != nil {
			log.Fatal("Failed to generate state key:", err)
		}
		log.Printf("Generated state encryption key: %s\n", base64.StdEncoding.EncodeToString(stateKey))
	}
	
	stateManager, err := auth.NewStateManager(stateKey, 10*time.Minute) // 10 minute state TTL
	if err != nil {
		log.Fatal("Failed to create state manager:", err)
	}
	
	// Create provider registry
	registry := auth.NewProviderRegistry()
	
	// Create in-memory user store (replace with database in production)
	userStore := NewInMemoryUserStore()
	
	// Register providers
	// Google OAuth
	if config.Providers["google"].Enabled {
		googleProvider, err := oauth.NewGoogleProvider(
			config.Providers["google"].Config["client_id"],
			config.Providers["google"].Config["client_secret"],
			config.Providers["google"].Config["redirect_url"],
		)
		if err != nil {
			log.Printf("Failed to create Google provider: %v", err)
		} else {
			registry.Register(googleProvider)
			log.Println("Registered Google OAuth provider")
		}
	}
	
	// Email/Password provider
	if config.Providers["email"].Enabled {
		emailProvider, err := auth.NewEmailPasswordProvider(
			userStore,
			auth.EmailProviderConfig{
				MinLength:           8,
				RequireUppercase:    true,
				RequireLowercase:    true,
				RequireNumbers:      true,
				RequireSpecialChar:  false,
				BcryptCost:          12,
				RequireVerification: false,
			},
		)
		if err != nil {
			log.Printf("Failed to create email provider: %v", err)
		} else {
			registry.Register(emailProvider)
			log.Println("Registered email/password provider")
		}
	}
	
	// Anonymous provider
	if config.Providers["anonymous"].Enabled {
		anonProvider, err := auth.NewAnonymousProvider(sessionManager, userStore)
		if err != nil {
			log.Printf("Failed to create anonymous provider: %v", err)
		} else {
			registry.Register(anonProvider)
			log.Println("Registered anonymous provider")
		}
	}
	
	// Create auth service
	service, err := authService.NewService(
		registry,
		tokenManager,
		sessionManager,
		userStore,
		config,
	)
	if err != nil {
		log.Fatal("Failed to create auth service:", err)
	}
	
	// Create router
	router := authService.NewRouter(
		service,
		stateManager,
		registry,
		tokenManager,
		sessionManager,
	)
	
	// Start server
	addr := getEnv("SERVER_ADDR", ":8080")
	log.Printf("Starting auth server on %s", addr)
	log.Printf("OAuth endpoints:")
	log.Printf("  - Initiate: http://localhost%s/auth/oauth/{provider}", addr)
	log.Printf("  - Callback: http://localhost%s/auth/oauth/{provider}/callback", addr)
	log.Printf("  - Refresh:  http://localhost%s/auth/refresh", addr)
	log.Printf("ConnectRPC endpoints:")
	log.Printf("  - http://localhost%s/api.v1.service.auth.AuthService/*", addr)
	
	if err := router.ListenAndServe(addr); err != nil {
		log.Fatal("Server failed:", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// InMemoryUserStore is a simple in-memory implementation for testing
type InMemoryUserStore struct {
	users map[string]*auth.User
	emails map[string]string // email -> userID mapping
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users:  make(map[string]*auth.User),
		emails: make(map[string]string),
	}
}

func (s *InMemoryUserStore) CreateUser(ctx context.Context, user *auth.User) error {
	if user.ID == "" {
		user.ID = generateID()
	}
	s.users[user.ID] = user
	s.emails[user.Email] = user.ID
	return nil
}

func (s *InMemoryUserStore) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	userID, exists := s.emails[email]
	if !exists {
		return nil, &auth.AuthError{Code: auth.ErrUserNotFound, Message: "user not found"}
	}
	return s.users[userID], nil
}

func (s *InMemoryUserStore) GetUserByID(ctx context.Context, userID string) (*auth.User, error) {
	user, exists := s.users[userID]
	if !exists {
		return nil, &auth.AuthError{Code: auth.ErrUserNotFound, Message: "user not found"}
	}
	return user, nil
}

func (s *InMemoryUserStore) UpdateUser(ctx context.Context, user *auth.User) error {
	s.users[user.ID] = user
	return nil
}

func (s *InMemoryUserStore) CheckUsernameAvailable(ctx context.Context, username string) (bool, error) {
	for _, user := range s.users {
		if user.Username == username {
			return false, nil
		}
	}
	return true, nil
}

func generateID() string {
	return base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("user_%d", time.Now().UnixNano())))
}