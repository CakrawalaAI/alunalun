-- Owner: feature/posts-core
-- Dependencies: none (base feature)
-- Migration: 004_posts_and_locations.sql
-- Tables: posts, post_locations, post_media

-- === QUERY DEFINITIONS START ===
-- Add SQLC query definitions below this line
-- Follow naming convention: ActionEntityCondition

-- Post CRUD Operations
-- name: CreatePost :one
-- name: GetPostByID :one
-- name: UpdatePost :one
-- name: DeletePost :exec

-- Post Location Operations
-- name: CreatePostLocation :one
-- name: UpdatePostLocation :one

-- Post Media Operations  
-- name: CreatePostMedia :one
-- name: GetPostMedia :many
-- name: DeletePostMedia :exec

-- Feed Queries
-- name: GetMyPosts :many
-- name: GetUserPosts :many
-- name: GetPostsByBoundingBox :many
-- name: GetPostsByRadius :many
-- name: GetPostsByGeohashPrefix :many

-- === QUERY DEFINITIONS END ===