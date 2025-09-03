package pin

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmcloughlin/geohash"
	entitiesv1 "github.com/radjathaher/alunalun/api/internal/protocgen/v1/entities"
	servicev1 "github.com/radjathaher/alunalun/api/internal/protocgen/v1/service"
	"github.com/radjathaher/alunalun/api/internal/protocgen/v1/service/servicev1connect"
	"github.com/radjathaher/alunalun/api/internal/protoconv"
	"github.com/radjathaher/alunalun/api/internal/repository"
	"github.com/radjathaher/alunalun/api/internal/utils/auth"
)

// Service implements the PinService
type Service struct {
	servicev1connect.UnimplementedPinServiceHandler
	db           *pgxpool.Pool
	queries      *repository.Queries
	tokenManager *auth.TokenManager
}

// NewService creates a new pin service
func NewService(db *pgxpool.Pool, queries *repository.Queries, tokenManager *auth.TokenManager) *Service {
	return &Service{
		db:           db,
		queries:      queries,
		tokenManager: tokenManager,
	}
}

// CreatePin creates a new pin on the map (requires authentication)
func (s *Service) CreatePin(
	ctx context.Context,
	req *connect.Request[servicev1.CreatePinRequest],
) (*connect.Response[servicev1.CreatePinResponse], error) {
	// Extract user from JWT context
	claims := s.extractClaims(req.Header())
	if claims == nil || claims.UserID == "" {
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("authentication required to create pins"),
		)
	}

	// Validate request
	if req.Msg.Location == nil {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("location is required"),
		)
	}

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to start transaction: %w", err))
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	// Create post and location params
	postParams, locationParams, err := protoconv.ProtoToCreatePinParams(
		claims.UserID,
		req.Msg.Content,
		req.Msg.Location,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to prepare params: %w", err))
	}

	// Create post
	post, err := qtx.CreatePost(ctx, postParams)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create post: %w", err))
	}

	// Create location
	var location *repository.PostsLocation
	if locationParams != nil {
		location, err = qtx.CreatePostLocation(ctx, locationParams)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create location: %w", err))
		}
	}

	// Track event
	if err := s.trackUserEvent(ctx, qtx, claims.UserID, "create_pin", req); err != nil {
		// Log but don't fail
		fmt.Printf("failed to track event: %v\n", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to commit transaction: %w", err))
	}

	// Get author for response
	author, _ := s.queries.GetUserByID(ctx, claims.UserID)

	// Convert to proto
	pin := protoconv.PinToProto(post, location, author, 0)

	return connect.NewResponse(&servicev1.CreatePinResponse{
		Pin: pin,
	}), nil
}

// ListPins returns pins within a geographic area (public read)
func (s *Service) ListPins(
	ctx context.Context,
	req *connect.Request[servicev1.ListPinsRequest],
) (*connect.Response[servicev1.ListPinsResponse], error) {
	// Public read - no auth required

	// Check if using zoom-based or bounding box query
	if req.Msg.Zoom != nil && req.Msg.Zoom != nil {
		// Zoom-based query with center point
		zoom := *req.Msg.Zoom

		// Check minimum zoom level
		if zoom < 12 {
			return connect.NewResponse(&servicev1.ListPinsResponse{
				Pins: []*entitiesv1.Pin{},
				Message: strPtr("Zoom in to see pins"),
			}), nil
		}

		// Calculate geohash precision
		precision := protoconv.CalculateGeohashPrecision(zoom)
		ghash := geohash.EncodeWithPrecision(req.Msg.Latitude, req.Msg.Longitude, precision)

		// Calculate limit
		limit := protoconv.CalculateLimit(zoom)

		// Query pins by geohash
		pins, err := s.queries.ListPinsByGeohash(ctx, &repository.ListPinsByGeohashParams{
			Geohash: ghash,
			Limit:   limit,
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list pins: %w", err))
		}

		// Convert to proto
		protoPins := make([]*entitiesv1.Pin, len(pins))
		for i, pinRow := range pins {
			// Fetch author and comment count for each pin
			author, _ := s.queries.GetUserByID(ctx, pinRow.UserID)
			commentCount, _ := s.queries.CountCommentsByPinID(ctx, pinRow.ID)

			protoPins[i] = protoconv.PinToProto(
				&pinRow.Post,
				&pinRow.PostsLocation,
				author,
				int32(commentCount),
			)
		}

		return connect.NewResponse(&servicev1.ListPinsResponse{
			Pins: protoPins,
		}), nil

	} else if req.Msg.North != nil && req.Msg.South != nil && 
		req.Msg.East != nil && req.Msg.West != nil {
		// Bounding box query
		pins, err := s.queries.ListPinsInBoundingBox(ctx, &repository.ListPinsInBoundingBoxParams{
			North: *req.Msg.North,
			South: *req.Msg.South,
			East:  *req.Msg.East,
			West:  *req.Msg.West,
			Limit: valueOr(req.Msg.Limit, 50),
		})
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list pins: %w", err))
		}

		// Convert to proto
		protoPins := make([]*entitiesv1.Pin, len(pins))
		for i, pinRow := range pins {
			author, _ := s.queries.GetUserByID(ctx, pinRow.UserID)
			commentCount, _ := s.queries.CountCommentsByPinID(ctx, pinRow.ID)

			protoPins[i] = protoconv.PinToProto(
				&pinRow.Post,
				&pinRow.PostsLocation,
				author,
				int32(commentCount),
			)
		}

		return connect.NewResponse(&servicev1.ListPinsResponse{
			Pins: protoPins,
		}), nil
	}

	return nil, connect.NewError(
		connect.CodeInvalidArgument,
		errors.New("must provide either zoom+center or bounding box"),
	)
}

// GetPin returns a single pin with its comments (public read)
func (s *Service) GetPin(
	ctx context.Context,
	req *connect.Request[servicev1.GetPinRequest],
) (*connect.Response[servicev1.GetPinResponse], error) {
	// Public read - no auth required

	if req.Msg.PinId == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("pin_id is required"),
		)
	}

	// Get pin with location
	pinWithLocation, err := s.queries.GetPinWithLocation(ctx, req.Msg.PinId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("pin not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get pin: %w", err))
	}

	// Get author
	author, _ := s.queries.GetUserByID(ctx, pinWithLocation.Post.UserID)

	// Get comments
	comments, err := s.queries.GetCommentsByPinID(ctx, req.Msg.PinId)
	if err != nil && err != pgx.ErrNoRows {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get comments: %w", err))
	}

	// Convert comments to proto
	protoComments := make([]*entitiesv1.Comment, len(comments))
	for i, comment := range comments {
		commentAuthor, _ := s.queries.GetUserByID(ctx, comment.UserID)
		protoComments[i] = protoconv.CommentToProto(&comment, commentAuthor)
	}

	// Convert pin to proto
	pin := protoconv.PinToProto(
		&pinWithLocation.Post,
		&pinWithLocation.PostsLocation,
		author,
		int32(len(comments)),
	)

	return connect.NewResponse(&servicev1.GetPinResponse{
		Pin:      pin,
		Comments: protoComments,
	}), nil
}

// DeletePin deletes a pin (requires auth + ownership)
func (s *Service) DeletePin(
	ctx context.Context,
	req *connect.Request[servicev1.DeletePinRequest],
) (*connect.Response[servicev1.DeletePinResponse], error) {
	// Extract user from JWT context
	claims := s.extractClaims(req.Header())
	if claims == nil || claims.UserID == "" {
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("authentication required"),
		)
	}

	if req.Msg.PinId == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("pin_id is required"),
		)
	}

	// Get pin to verify ownership
	post, err := s.queries.GetPostByID(ctx, req.Msg.PinId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("pin not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get pin: %w", err))
	}

	// Check ownership
	if post.UserID != claims.UserID {
		return nil, connect.NewError(
			connect.CodePermissionDenied,
			errors.New("you can only delete your own pins"),
		)
	}

	// Delete pin (cascades to posts_location and comments)
	if err := s.queries.DeletePost(ctx, req.Msg.PinId); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to delete pin: %w", err))
	}

	return connect.NewResponse(&servicev1.DeletePinResponse{
		Success: true,
	}), nil
}

// AddComment adds a comment to a pin (requires authentication)
func (s *Service) AddComment(
	ctx context.Context,
	req *connect.Request[servicev1.AddCommentRequest],
) (*connect.Response[servicev1.AddCommentResponse], error) {
	// Extract user from JWT context
	claims := s.extractClaims(req.Header())
	if claims == nil || claims.UserID == "" {
		return nil, connect.NewError(
			connect.CodeUnauthenticated,
			errors.New("authentication required to add comments"),
		)
	}

	// Validate request
	if req.Msg.PinId == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("pin_id is required"),
		)
	}
	if req.Msg.Content == "" {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			errors.New("content is required"),
		)
	}

	// Verify pin exists
	_, err := s.queries.GetPostByID(ctx, req.Msg.PinId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("pin not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to verify pin: %w", err))
	}

	// Create comment params
	commentParams, err := protoconv.ProtoToCreateCommentParams(
		claims.UserID,
		req.Msg.PinId,
		req.Msg.Content,
		req.Msg.ParentId,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to prepare params: %w", err))
	}

	// Create comment
	comment, err := s.queries.CreatePost(ctx, commentParams)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create comment: %w", err))
	}

	// Get author for response
	author, _ := s.queries.GetUserByID(ctx, claims.UserID)

	// Convert to proto
	protoComment := protoconv.CommentToProto(comment, author)

	return connect.NewResponse(&servicev1.AddCommentResponse{
		Comment: protoComment,
	}), nil
}

// extractClaims extracts JWT claims from request headers
func (s *Service) extractClaims(headers connect.Headers) *auth.Claims {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return nil
	}

	// Remove "Bearer " prefix
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	// Validate token
	claims, err := s.tokenManager.ValidateToken(token)
	if err != nil {
		return nil
	}

	return claims
}

// trackUserEvent tracks user events for analytics and rate limiting
func (s *Service) trackUserEvent(
	ctx context.Context,
	qtx *repository.Queries,
	userID string,
	eventType string,
	req interface{},
) error {
	// TODO: Extract IP address and fingerprint from request context
	// This would typically come from middleware or request headers
	
	// For now, just return nil to not block operations
	return nil
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func valueOr[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}