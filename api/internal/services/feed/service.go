package feed

// This file is owned by: feature/location-feed
// Dependencies: posts service (must be merged first)
// Database tables: feed_sessions, feed_hot_zones
// Proto files: proto/v1/service/feed.proto

import (
	"context"
	"database/sql"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// === IMPLEMENTATION START ===
// Add service methods below this line
// Do not modify the struct or constructor above

// Core feed operations will go here:
// - GetLocationFeed (by bounds/radius/geohash)
// - RefreshFeed (check for new posts)
// - GetTrendingFeed
// - GetNearbyFeed
// - StartFeedSession
// - UpdateFeedSession

// === IMPLEMENTATION END ===