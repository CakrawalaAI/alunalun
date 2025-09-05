package auth

import (
	"context"
	"net/http"
	
	"connectrpc.com/connect"
	"github.com/radjathaher/alunalun/api/internal/protocgen/v1/auth_service/auth_servicev1connect"
	"github.com/radjathaher/alunalun/api/internal/utils/auth"
	"github.com/rs/cors"
)

// Router handles both ConnectRPC and HTTP OAuth endpoints
type Router struct {
	service       *Service
	oauthHandler  *OAuthHandler
	corsHandler   *cors.Cors
}

// NewRouter creates a new combined router
func NewRouter(
	service *Service,
	stateManager *auth.StateManager,
	registry *auth.ProviderRegistry,
	tokenManager *auth.TokenManager,
	sessionManager *auth.SessionManager,
) *Router {
	// Create OAuth handler
	oauthHandler := NewOAuthHandler(
		service,
		stateManager,
		registry,
		tokenManager,
		sessionManager,
	)
	
	// Configure CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure based on environment
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
	
	return &Router{
		service:      service,
		oauthHandler: oauthHandler,
		corsHandler:  corsHandler,
	}
}

// Handler returns the main HTTP handler combining all routes
func (r *Router) Handler() http.Handler {
	mux := http.NewServeMux()
	
	// Register OAuth HTTP endpoints
	r.oauthHandler.RegisterRoutes(mux)
	
	// Register ConnectRPC endpoints
	path, handler := auth_servicev1connect.NewAuthServiceHandler(r.service)
	mux.Handle(path, handler)
	
	// Apply CORS to all routes
	return r.corsHandler.Handler(mux)
}

// ListenAndServe starts the server
func (r *Router) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, r.Handler())
}

// AuthInterceptor provides authentication middleware for ConnectRPC
func AuthInterceptor(tokenManager *auth.TokenManager) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Skip auth for certain endpoints
			procedure := req.Spec().Procedure
			skipAuth := []string{
				"CheckUsername",
				"InitAnonymous",
				"Authenticate",
				"RefreshToken",
			}
			
			for _, skip := range skipAuth {
				if procedure == skip {
					return next(ctx, req)
				}
			}
			
			// Extract token from Authorization header
			authHeader := req.Header().Get("Authorization")
			if authHeader == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}
			
			// Validate token
			token := authHeader
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
			}
			
			claims, err := tokenManager.ValidateToken(token)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			
			// Add claims to context
			ctx = auth.ContextWithClaims(ctx, claims)
			
			return next(ctx, req)
		}
	}
}