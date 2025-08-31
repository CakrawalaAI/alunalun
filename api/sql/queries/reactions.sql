-- Owner: feature/reactions-system
-- Dependencies: posts table (from 004_posts_and_locations.sql)
-- Migration: 006_reactions.sql
-- Tables: post_reactions, reaction_activity

-- === QUERY DEFINITIONS START ===
-- Add SQLC query definitions below this line
-- Follow naming convention: ActionEntityCondition

-- Reaction CRUD Operations
-- name: AddReaction :one
-- name: RemoveReaction :exec
-- name: RemoveAllReactions :exec
-- name: ChangeReaction :one

-- Reaction Query Operations
-- name: GetPostReactions :many
-- name: GetUserReactionsForPost :many
-- name: GetUserReactionsForPosts :many
-- name: GetPostsWithReactionCounts :many

-- Analytics Operations
-- name: GetTopReactors :many
-- name: GetReactionActivity :many
-- name: GetUserReactionHistory :many
-- name: CountReactionsByType :one

-- === QUERY DEFINITIONS END ===