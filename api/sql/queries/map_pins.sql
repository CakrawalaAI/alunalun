-- Owner: feature/map-pins
-- Dependencies: posts, post_locations tables (from 004_posts_and_locations.sql)
-- Migration: 008_map_optimizations.sql
-- Tables: map_pin_clusters, user_map_preferences

-- === QUERY DEFINITIONS START ===
-- Add SQLC query definitions below this line
-- Follow naming convention: ActionEntityCondition

-- Map Pin Queries
-- name: GetMapPins :many
-- name: GetPinDetails :one
-- name: GetClusterPins :many

-- Clustering Operations
-- name: PrecomputeClusters :exec
-- name: GetCachedClusters :many
-- name: InvalidateClusters :exec

-- User Preferences
-- name: GetUserMapPreferences :one
-- name: UpdateUserMapPreferences :one

-- === QUERY DEFINITIONS END ===