-- name: CreatePost :one
INSERT INTO posts (id, user_id, type, parent_id, content, visibility, metadata, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPostByID :one
SELECT * FROM posts WHERE id = $1;

-- name: GetPostWithAuthor :one
SELECT 
    p.*,
    u.id as author_id,
    u.username as author_username,
    u.display_name as author_display_name,
    u.avatar_url as author_avatar_url
FROM posts p
LEFT JOIN users u ON p.user_id = u.id
WHERE p.id = $1;

-- name: ListPostsByUser :many
SELECT * FROM posts 
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPostsByType :many
SELECT * FROM posts
WHERE type = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListCommentsByParent :many
SELECT 
    p.*,
    u.username as author_username,
    u.display_name as author_display_name,
    u.avatar_url as author_avatar_url
FROM posts p
LEFT JOIN users u ON p.user_id = u.id
WHERE p.parent_id = $1 AND p.type = 'comment'
ORDER BY p.created_at ASC
LIMIT $2 OFFSET $3;

-- name: UpdatePost :one
UPDATE posts
SET 
    content = $2,
    visibility = $3,
    metadata = $4
WHERE id = $1
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1;

-- name: ListRecentPins :many
SELECT 
    p.*,
    u.username as author_username,
    u.display_name as author_display_name,
    u.avatar_url as author_avatar_url
FROM posts p
LEFT JOIN users u ON p.user_id = u.id
WHERE p.type = 'pin' 
    AND p.created_at > NOW() - INTERVAL '24 hours'
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountPostsByUser :one
SELECT COUNT(*) FROM posts WHERE user_id = $1;

-- name: CountCommentsByParent :one
SELECT COUNT(*) FROM posts 
WHERE parent_id = $1 AND type = 'comment';