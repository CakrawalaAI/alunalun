package map

// This file is owned by: feature/map-pins
// Dependencies: posts service (must be merged first)
// Database tables: map_pin_clusters, user_map_preferences
// Proto files: proto/v1/service/map_pins.proto

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

// Core map operations will go here:
// - GetMapPins (with clustering)
// - GetPinDetails
// - GetClusterPins
// - GetMapPreferences
// - UpdateMapPreferences
// - PrecomputeClusters (background job)

// === IMPLEMENTATION END ===