-- Owner: feature/location-feed
-- Dependencies: posts, post_locations tables (from 004_posts_and_locations.sql)
-- Migration: 007_feed_optimizations.sql
-- Tables: feed_sessions, feed_hot_zones

-- === QUERY DEFINITIONS START ===
-- Add SQLC query definitions below this line
-- Follow naming convention: ActionEntityCondition

-- Main Feed Queries
-- name: GetLocationFeed :many
-- name: GetRadiusFeed :many
-- name: GetGeohashFeed :many

-- Feed Refresh Operations
-- name: CountNewPosts :one
-- name: GetNewPostIDs :many

-- Trending/Popular Queries
-- name: GetTrendingInArea :many
-- name: GetHotZones :many

-- Feed Session Management
-- name: CreateFeedSession :one
-- name: UpdateFeedSession :exec
-- name: GetFeedSession :one

-- === QUERY DEFINITIONS END ===