package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/radjathaher/alunalun/api/internal/config"
	"github.com/radjathaher/alunalun/api/internal/middleware"
	"github.com/radjathaher/alunalun/api/internal/protocgen/v1/service/servicev1connect"
	authService "github.com/radjathaher/alunalun/api/internal/services/auth"
	commentsService "github.com/radjathaher/alunalun/api/internal/services/comments"
	feedService "github.com/radjathaher/alunalun/api/internal/services/feed"
	mapService "github.com/radjathaher/alunalun/api/internal/services/map"
	mediaService "github.com/radjathaher/alunalun/api/internal/services/media"
	postsService "github.com/radjathaher/alunalun/api/internal/services/posts"
	reactionsService "github.com/radjathaher/alunalun/api/internal/services/reactions"
	"github.com/radjathaher/alunalun/api/internal/utils/db"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

type Server struct {
	config     *config.Config
	db         *db.DB
	logger     *zap.SugaredLogger
	httpServer *http.Server
}

func New(cfg *config.Config, database *db.DB, logger *zap.SugaredLogger) *Server {
	return &Server{
		config: cfg,
		db:     database,
		logger: logger,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register health check
	mux.HandleFunc("/health", s.healthHandler)

	// Initialize services
	authSvc := authService.NewService(s.db, s.config.Auth, s.logger)
	postsSvc := postsService.NewService(s.db, s.logger)
	commentsSvc := commentsService.NewService(s.db, s.logger)
	reactionsSvc := reactionsService.NewService(s.db, s.logger)
	feedSvc := feedService.NewService(s.db, s.logger)
	mapSvc := mapService.NewService(s.db, s.config.Services.MapboxToken, s.logger)
	mediaSvc := mediaService.NewService(s.db, s.config.Services.MediaPath, s.logger)

	// Create interceptors
	interceptors := connect.WithInterceptors(
		middleware.NewAuthInterceptor(authSvc),
		middleware.NewLoggingInterceptor(s.logger),
	)

	// Register Connect RPC services
	authPath, authHandler := servicev1connect.NewAuthServiceHandler(authSvc, interceptors)
	mux.Handle(authPath, authHandler)

	// Register other services when their handlers are ready
	// userPath, userHandler := servicev1connect.NewUserServiceHandler(userSvc, interceptors)
	// mux.Handle(userPath, userHandler)

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   s.config.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	handler := c.Handler(mux)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", s.config.Server.Host, s.config.Server.Port),
		Handler:      handler,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
	}

	s.logger.Infof("Starting server on %s", s.httpServer.Addr)

	// Start server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	// Close database connection
	s.db.Close()

	s.logger.Info("Server shutdown complete")
	return nil
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check database health
	if err := s.db.Health(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("unhealthy"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("healthy"))
}