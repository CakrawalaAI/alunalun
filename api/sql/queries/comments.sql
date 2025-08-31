-- Owner: feature/comments-system
-- Dependencies: posts table (from 004_posts_and_locations.sql)
-- Migration: 005_comments.sql
-- Tables: post_comments

-- === QUERY DEFINITIONS START ===
-- Add SQLC query definitions below this line
-- Follow naming convention: ActionEntityCondition

-- Comment CRUD Operations
-- name: CreateComment :one
-- name: GetCommentByID :one
-- name: UpdateComment :one
-- name: SoftDeleteComment :exec

-- Comment Query Operations
-- name: GetCommentsByPost :many
-- name: GetCommentsByPostFlat :many
-- name: GetCommentReplies :many
-- name: GetUserComments :many

-- Comment Utility Operations
-- name: CountCommentsByPost :one
-- name: CountRepliesByComment :one
-- name: CheckCommentOwnership :one

-- === QUERY DEFINITIONS END ===