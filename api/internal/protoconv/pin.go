package protoconv

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	entitiesv1 "github.com/radjathaher/alunalun/api/internal/protocgen/v1/entities"
	"github.com/radjathaher/alunalun/api/internal/repository"
	"github.com/mmcloughlin/geohash"
	"github.com/segmentio/ksuid"
)

// PinToProto converts repository Post+PostsLocation to protobuf Pin entity
func PinToProto(post *repository.Post, location *repository.PostsLocation, author *repository.User, commentCount int32) *entitiesv1.Pin {
	if post == nil {
		return nil
	}

	pin := &entitiesv1.Pin{
		Id:           post.ID,
		UserId:       post.UserID,
		CreatedAt:    post.CreatedAt.Time.Unix(),
		UpdatedAt:    post.UpdatedAt.Time.Unix(),
		CommentCount: commentCount,
	}

	// Set content if not null
	if post.Content != nil {
		pin.Content = *post.Content
	}

	// Add location if available
	if location != nil {
		pin.Location = LocationToProto(location)
	}

	// Add author if available
	if author != nil {
		pin.Author = UserToProto(author)
	}

	return pin
}

// LocationToProto converts repository PostsLocation to protobuf Location
func LocationToProto(loc *repository.PostsLocation) *entitiesv1.Location {
	if loc == nil {
		return nil
	}

	// Parse geometry point
	// Assuming location is stored as PostGIS POINT
	// This will need proper PostGIS decoding based on your driver setup
	lat, lng, alt := parseGeometry(loc.Location)

	protoLoc := &entitiesv1.Location{
		Latitude:  lat,
		Longitude: lng,
	}

	// Set altitude if available (not 0)
	if alt != 0 {
		protoLoc.Altitude = &alt
	}

	// Set geohash if available
	if loc.Geohash != nil {
		protoLoc.Geohash = loc.Geohash
	}

	return protoLoc
}

// ProtoToCreatePinParams converts protobuf CreatePinRequest to repository params
func ProtoToCreatePinParams(userID string, content string, location *entitiesv1.Location) (*repository.CreatePostParams, *repository.CreatePostLocationParams, error) {
	// Generate ID
	pinID := "pin-" + ksuid.New().String()
	now := time.Now()

	// Create post params
	postParams := &repository.CreatePostParams{
		ID:       pinID,
		UserID:   userID,
		Type:     "pin",
		Content:  &content,
		ParentID: nil, // Pins have no parent
		Metadata: []byte("{}"),
		CreatedAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
	}

	// Create location params if location provided
	var locationParams *repository.CreatePostLocationParams
	if location != nil {
		// Calculate geohash with 8 character precision (~38m accuracy)
		ghash := geohash.EncodeWithPrecision(location.Latitude, location.Longitude, 8)

		locationParams = &repository.CreatePostLocationParams{
			PostID: pinID,
			// Location will need to be formatted as PostGIS POINT
			// Format: POINT(longitude latitude)
			Location: fmt.Sprintf("POINT(%f %f)", location.Longitude, location.Latitude),
			Geohash:  &ghash,
			CreatedAt: pgtype.Timestamptz{
				Time:  now,
				Valid: true,
			},
		}
	}

	return postParams, locationParams, nil
}

// ProtoToCreateCommentParams converts comment request to repository params
func ProtoToCreateCommentParams(userID, pinID, content string, parentID *string) (*repository.CreatePostParams, error) {
	commentID := "comment-" + ksuid.New().String()
	now := time.Now()

	// For comments, the parent_id is the pin or another comment
	actualParentID := &pinID
	if parentID != nil && *parentID != "" {
		actualParentID = parentID
	}

	metadata := map[string]interface{}{
		"pin_id": pinID, // Track which pin this comment belongs to
	}
	metadataJSON, _ := json.Marshal(metadata)

	return &repository.CreatePostParams{
		ID:       commentID,
		UserID:   userID,
		Type:     "comment",
		Content:  &content,
		ParentID: actualParentID,
		Metadata: metadataJSON,
		CreatedAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
	}, nil
}

// CommentToProto converts repository Post (type=comment) to protobuf Comment
func CommentToProto(post *repository.Post, author *repository.User) *entitiesv1.Comment {
	if post == nil || post.Type != "comment" {
		return nil
	}

	comment := &entitiesv1.Comment{
		Id:        post.ID,
		UserId:    post.UserID,
		ParentId:  post.ParentID,
		CreatedAt: post.CreatedAt.Time.Unix(),
	}

	// Set content if not null
	if post.Content != nil {
		comment.Content = *post.Content
	}

	// Add author if available
	if author != nil {
		comment.Author = UserToProto(author)
	}

	return comment
}

// parseGeometry extracts lat/lng/alt from PostGIS geometry
// This is a placeholder - actual implementation depends on your PostGIS driver
func parseGeometry(geometry interface{}) (lat, lng, alt float64) {
	// TODO: Implement based on your PostGIS driver
	// For pgx with PostGIS, you might use:
	// - github.com/twpayne/go-geom for geometry parsing
	// - Or parse the WKB/WKT directly
	
	// Placeholder implementation
	// In production, parse the actual geometry object
	return 0, 0, 0
}

// CalculateGeohashPrecision returns appropriate geohash precision for zoom level
func CalculateGeohashPrecision(zoom int32) int {
	switch {
	case zoom < 6:
		return 1 // ~5000km
	case zoom < 9:
		return 2 // ~1250km
	case zoom < 12:
		return 3 // ~156km
	case zoom < 15:
		return 4 // ~39km
	case zoom < 17:
		return 5 // ~4.9km
	case zoom < 19:
		return 6 // ~1.2km
	case zoom < 21:
		return 7 // ~152m
	default:
		return 8 // ~38m
	}
}

// CalculateLimit returns appropriate result limit for zoom level
func CalculateLimit(zoom int32) int32 {
	switch {
	case zoom < 12:
		return 0 // Don't load
	case zoom < 15:
		return 50
	case zoom < 17:
		return 100
	case zoom < 19:
		return 200
	default:
		return 500
	}
}