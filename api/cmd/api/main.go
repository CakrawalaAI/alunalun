package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radjathaher/alunalun/api/internal/repository"
	"github.com/radjathaher/alunalun/api/internal/server"
	"github.com/radjathaher/alunalun/api/internal/utils/auth"
)

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup database
	db, err := setupDatabase(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create repository
	queries := repository.New(db)

	// Setup auth components
	tokenManager, stateManager, sessionManager, err := setupAuth(
		config.JWTPrivateKey,
		config.JWTPublicKey,
		config.JWTIssuer,
		config.JWTAudience,
		config.OAuthStateKey,
	)
	if err != nil {
		log.Fatalf("Failed to setup auth: %v", err)
	}

	// Create server config
	serverConfig := &server.Config{
		// Server
		Addr:         config.ServerAddr,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,

		// Database
		DB:      db,
		Queries: queries,

		// Auth
		TokenManager:   tokenManager,
		StateManager:   stateManager,
		SessionManager: sessionManager,
		AuthConfig:     config.AuthConfig,

		// OAuth Providers
		GoogleClientID:     config.GoogleClientID,
		GoogleClientSecret: config.GoogleClientSecret,
		GoogleRedirectURL:  config.GoogleRedirectURL,

		// Future: Add more dependencies here
		// S3Client:    s3Client,
		// RedisClient: redisClient,
		// QueueClient: queueClient,
	}

	// Create and start server
	srv, err := server.New(serverConfig)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", config.ServerAddr)
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// AppConfig holds all application configuration
type AppConfig struct {
	// Server
	ServerAddr string

	// Database
	DatabaseURL string

	// JWT
	JWTPrivateKey []byte
	JWTPublicKey  []byte
	JWTIssuer     string
	JWTAudience   string

	// OAuth
	OAuthStateKey      []byte
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// Auth Config
	AuthConfig *auth.Config
}

// loadConfig loads configuration from environment or defaults
func loadConfig() (*AppConfig, error) {
	config := &AppConfig{
		ServerAddr:  getEnv("SERVER_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://localhost/alunalun?sslmode=disable"),
		JWTIssuer:   getEnv("JWT_ISSUER", "alunalun"),
		JWTAudience: getEnv("JWT_AUDIENCE", "web"),

		// OAuth
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/oauth/google/callback"),
	}

	// Load or generate JWT keys
	if privateKeyPath := os.Getenv("JWT_PRIVATE_KEY_PATH"); privateKeyPath != "" {
		privateKey, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}
		config.JWTPrivateKey = privateKey

		publicKey, err := os.ReadFile(os.Getenv("JWT_PUBLIC_KEY_PATH"))
		if err != nil {
			return nil, fmt.Errorf("failed to read public key: %w", err)
		}
		config.JWTPublicKey = publicKey
	} else {
		// Generate keys for development
		privateKey, publicKey, err := auth.GenerateKeyPair()
		if err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}
		config.JWTPrivateKey = privateKey
		config.JWTPublicKey = publicKey
		log.Println("Generated JWT key pair for development")
	}

	// Load or generate OAuth state key
	if envKey := os.Getenv("OAUTH_STATE_KEY"); envKey != "" {
		stateKey, err := base64.StdEncoding.DecodeString(envKey)
		if err != nil {
			return nil, fmt.Errorf("invalid OAUTH_STATE_KEY: %w", err)
		}
		config.OAuthStateKey = stateKey
	} else {
		// Generate for development
		stateKey, err := auth.GenerateEncryptionKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate state key: %w", err)
		}
		config.OAuthStateKey = stateKey
		log.Println("Generated OAuth state key for development")
	}

	// Build auth config
	config.AuthConfig = &auth.Config{
		JWT: auth.JWTConfig{
			Issuer:          config.JWTIssuer,
			Audience:        config.JWTAudience,
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
		Session: auth.SessionConfig{
			TTL:             time.Hour,
			CleanupInterval: time.Hour,
			MaxPerUser:      10,
		},
		Providers: map[string]auth.ProviderConfig{
			"google": {
				Type:    "oauth",
				Enabled: config.GoogleClientID != "",
				Config: map[string]string{
					"client_id":     config.GoogleClientID,
					"client_secret": config.GoogleClientSecret,
					"redirect_url":  config.GoogleRedirectURL,
				},
			},
			// Future: Add more providers
		},
	}

	return config, nil
}

// setupDatabase creates database connection pool
func setupDatabase(databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Connection pool settings
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to database")
	return pool, nil
}

// setupAuth creates auth components
func setupAuth(privateKey, publicKey []byte, issuer, audience string, stateKey []byte) (*auth.TokenManager, *auth.StateManager, *auth.SessionManager, error) {
	// Create token manager
	tokenManager, err := auth.NewTokenManager(privateKey, publicKey, issuer, audience)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create token manager: %w", err)
	}

	// Create state manager
	stateManager, err := auth.NewStateManager(stateKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	// Create session manager (in-memory for now, Redis in production)
	sessionStore := auth.NewInMemorySessionStore()
	sessionManager := auth.NewSessionManager(sessionStore)

	return tokenManager, stateManager, sessionManager, nil
}

// getEnv gets environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}