package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/radjathaher/alunalun/api/internal/middleware"
	"github.com/radjathaher/alunalun/api/internal/protoconv"
	authServicePb "github.com/radjathaher/alunalun/api/internal/protocgen/v1/auth_service/auth_servicev1connect"
	pinServicePb "github.com/radjathaher/alunalun/api/internal/protocgen/v1/service/servicev1connect"
	userServicePb "github.com/radjathaher/alunalun/api/internal/protocgen/v1/service/servicev1connect"
	"github.com/radjathaher/alunalun/api/internal/repository"
	authService "github.com/radjathaher/alunalun/api/internal/services/auth"
	pinService "github.com/radjathaher/alunalun/api/internal/services/pin"
	userService "github.com/radjathaher/alunalun/api/internal/services/user"
	"github.com/radjathaher/alunalun/api/internal/utils/auth"
	"github.com/radjathaher/alunalun/api/internal/utils/oauth"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Config holds all server configuration
type Config struct {
	// Server
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	// Database
	DB      *pgxpool.Pool
	Queries *repository.Queries

	// Auth
	TokenManager   *auth.TokenManager
	StateManager   *auth.StateManager
	SessionManager *auth.SessionManager
	AuthConfig     *auth.Config

	// OAuth Providers
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// Future dependencies
	// S3Client    *s3.Client
	// RedisClient *redis.Client
	// QueueClient *sqs.Client
}

// Server represents the API server
type Server struct {
	config     *Config
	httpServer *http.Server
	mux        *http.ServeMux

	// Services
	authService *authService.Service
	userService *userService.Service
	pinService  *pinService.Service

	// Handlers
	oauthHandler *authService.OAuthHandler
}

// New creates a new server instance
func New(config *Config) (*Server, error) {
	s := &Server{
		config: config,
		mux:    http.NewServeMux(),
	}

	// Setup services
	if err := s.setupServices(); err != nil {
		return nil, fmt.Errorf("failed to setup services: %w", err)
	}

	// Setup routes
	if err := s.setupRoutes(); err != nil {
		return nil, fmt.Errorf("failed to setup routes: %w", err)
	}

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         config.Addr,
		Handler:      h2c.NewHandler(corsMiddleware(s.mux), &http2.Server{}),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return s, nil
}

// setupServices initializes all services
func (s *Server) setupServices() error {
	// Setup provider registry
	registry := auth.NewProviderRegistry()
	if err := s.setupProviders(registry); err != nil {
		return fmt.Errorf("failed to setup providers: %w", err)
	}

	// Create user store adapter for auth service
	// This implements auth.UserStore interface that auth service needs
	userStore := protoconv.NewPostgresUserStore(s.config.Queries)

	// Create auth service
	var err error
	s.authService, err = authService.NewService(
		registry,
		s.config.TokenManager,
		s.config.SessionManager,
		userStore,
		s.config.AuthConfig,
	)
	if err != nil {
		return fmt.Errorf("failed to create auth service: %w", err)
	}

	// Create user service
	s.userService = userService.NewService(
		s.config.Queries,
		s.config.TokenManager,
	)

	// Create pin service
	s.pinService = pinService.NewService(
		s.config.DB,
		s.config.Queries,
		s.config.TokenManager,
	)

	// Create OAuth HTTP handler
	s.oauthHandler = authService.NewOAuthHandler(
		s.authService,
		s.config.StateManager,
		registry,
		s.config.TokenManager,
		s.config.SessionManager,
	)

	return nil
}

// setupProviders registers auth providers
func (s *Server) setupProviders(registry *auth.ProviderRegistry) error {
	// Register Google OAuth if configured
	if s.config.GoogleClientID != "" {
		provider, err := oauth.NewGoogleProvider(
			s.config.GoogleClientID,
			s.config.GoogleClientSecret,
			s.config.GoogleRedirectURL,
		)
		if err != nil {
			return fmt.Errorf("failed to create Google provider: %w", err)
		}
		if err := registry.Register(provider); err != nil {
			return fmt.Errorf("failed to register Google provider: %w", err)
		}
	}

	// Register anonymous provider for testing
	anonProvider, err := auth.NewAnonymousProvider(s.config.SessionManager, protoconv.NewPostgresUserStore(s.config.Queries))
	if err != nil {
		return fmt.Errorf("failed to create anonymous provider: %w", err)
	}
	if err := registry.Register(anonProvider); err != nil {
		return fmt.Errorf("failed to register anonymous provider: %w", err)
	}

	// Future: Register other providers
	// - Apple OAuth
	// - GitHub OAuth
	// - Email/Password

	return nil
}

// setupRoutes mounts all HTTP and ConnectRPC endpoints
func (s *Server) setupRoutes() error {
	// Create auth interceptor
	authInterceptor := middleware.NewAuthInterceptor(s.config.TokenManager)
	interceptors := connect.WithInterceptors(authInterceptor)

	// Mount OAuth HTTP routes
	s.oauthHandler.RegisterRoutes(s.mux)

	// Mount ConnectRPC services
	authPath, authHandler := authServicePb.NewAuthServiceHandler(s.authService, interceptors)
	s.mux.Handle(authPath, authHandler)

	userPath, userHandler := userServicePb.NewUserServiceHandler(s.userService, interceptors)
	s.mux.Handle(userPath, userHandler)

	pinPath, pinHandler := pinServicePb.NewPinServiceHandler(s.pinService, interceptors)
	s.mux.Handle(pinPath, pinHandler)

	// Health check endpoint
	s.mux.HandleFunc("/health", s.handleHealth)

	// Future: Mount additional endpoints
	// - Metrics endpoint
	// - Ready endpoint
	// - WebSocket endpoints
	// - File upload endpoints

	return nil
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.config.DB.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Database unavailable"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Connect-Protocol-Version")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}