package pin

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
	var locationRow *repository.CreatePostLocationRow
	if locationParams != nil {
		locationRow, err = qtx.CreatePostLocation(ctx, locationParams)
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
	var authorUserID pgtype.UUID
	err = authorUserID.Scan(claims.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid user ID: %w", err))
	}
	author, _ := s.queries.GetUserByID(ctx, authorUserID)

	// Convert to proto using row data directly 
	var pin *entitiesv1.Pin
	if locationRow != nil {
		pin = protoconv.PinFromRowToProto(
			post.ID.String(),
			post.AuthorUserID.String(),
			post.Content,
			post.CreatedAt.Time.Unix(),
			post.UpdatedAt.Time.Unix(),
			locationRow.Longitude,
			locationRow.Latitude,
			locationRow.Geohash,
			author,
			0, // No comments for new pin
		)
	} else {
		// Fallback for pins without location
		pin = protoconv.PinFromRowToProto(
			post.ID.String(),
			post.AuthorUserID.String(),
			post.Content,
			post.CreatedAt.Time.Unix(),
			post.UpdatedAt.Time.Unix(),
			nil, nil, nil,
			author,
			0,
		)
	}

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

	// Zoom-based query with center point
	zoom := req.Msg.Zoom

	// Check minimum zoom level
	if zoom < 12 {
		return connect.NewResponse(&servicev1.ListPinsResponse{
			Pins: []*entitiesv1.Pin{},
		}), nil
	}

	// Calculate geohash precision
	precision := protoconv.CalculateGeohashPrecision(zoom)
	ghash := geohash.EncodeWithPrecision(req.Msg.Latitude, req.Msg.Longitude, uint(precision))

	// Calculate limit
	limit := protoconv.CalculateLimit(zoom)

	// Query pins by geohash
	pins, err := s.queries.ListPinsByGeohash(ctx, &repository.ListPinsByGeohashParams{
		Column1: &ghash,
		Limit:   limit,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list pins: %w", err))
	}

	// Convert to proto
	protoPins := make([]*entitiesv1.Pin, len(pins))
	for i, pinRow := range pins {
		// Fetch author and comment count for each pin
		author, _ := s.queries.GetUserByID(ctx, pinRow.AuthorUserID)
		commentCount, _ := s.queries.CountCommentsByPinID(ctx)

		// Convert pin row to proto using direct row conversion
		protoPins[i] = protoconv.PinFromRowToProto(
			pinRow.ID.String(),
			pinRow.AuthorUserID.String(),
			pinRow.Content,
			pinRow.CreatedAt.Time.Unix(),
			pinRow.UpdatedAt.Time.Unix(),
			pinRow.Longitude,
			pinRow.Latitude,
			pinRow.Geohash,
			author,
			int32(commentCount),
		)
	}

	return connect.NewResponse(&servicev1.ListPinsResponse{
		Pins: protoPins,
	}), nil
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

	// Parse UUID
	var pinID pgtype.UUID
	err := pinID.Scan(req.Msg.PinId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid pin ID: %w", err))
	}

	// Get pin with location
	pinWithLocation, err := s.queries.GetPinWithLocation(ctx, pinID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("pin not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get pin: %w", err))
	}

	// Get author
	author, _ := s.queries.GetUserByID(ctx, pinWithLocation.AuthorUserID)

	// Get comments - for now, return empty array as comments are not implemented
	protoComments := []*entitiesv1.Comment{}

	// Convert pin to proto using row data directly
	pin := protoconv.PinFromRowToProto(
		pinWithLocation.ID.String(),
		pinWithLocation.AuthorUserID.String(),
		pinWithLocation.Content,
		pinWithLocation.CreatedAt.Time.Unix(),
		pinWithLocation.UpdatedAt.Time.Unix(),
		pinWithLocation.Longitude,
		pinWithLocation.Latitude,
		pinWithLocation.Geohash,
		author,
		int32(0), // No comments for now
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

	// Parse UUID
	var pinID pgtype.UUID
	err := pinID.Scan(req.Msg.PinId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid pin ID: %w", err))
	}

	// Get pin to verify ownership
	post, err := s.queries.GetPostByID(ctx, pinID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("pin not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get pin: %w", err))
	}

	// Check ownership
	if post.AuthorUserID.String() != claims.UserID {
		return nil, connect.NewError(
			connect.CodePermissionDenied,
			errors.New("you can only delete your own pins"),
		)
	}

	// Delete pin (cascades to posts_location and comments)
	if err := s.queries.DeletePost(ctx, pinID); err != nil {
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

	// Parse pin UUID
	var pinID pgtype.UUID
	err := pinID.Scan(req.Msg.PinId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid pin ID: %w", err))
	}

	// Verify pin exists
	_, err = s.queries.GetPostByID(ctx, pinID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("pin not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to verify pin: %w", err))
	}

	// For now, comments are not fully implemented, so return a placeholder response
	// TODO: Implement proper comment creation when schema supports parent_id relationships
	
	// Get author for response
	var authorID pgtype.UUID
	err = authorID.Scan(claims.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid user ID: %w", err))
	}
	author, _ := s.queries.GetUserByID(ctx, authorID)

	// Create a placeholder comment response
	placeholderComment := &entitiesv1.Comment{
		Id:        req.Msg.PinId, // Use pin ID as placeholder
		UserId:    claims.UserID,
		Content:   req.Msg.Content,
		CreatedAt: 0, // Placeholder timestamp
	}
	if author != nil && author.DisplayName != nil {
		placeholderComment.Author = protoconv.UserToProto(author)
	}

	return connect.NewResponse(&servicev1.AddCommentResponse{
		Comment: placeholderComment,
	}), nil
}

// extractClaims extracts JWT claims from request headers
func (s *Service) extractClaims(headers http.Header) *auth.Claims {
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