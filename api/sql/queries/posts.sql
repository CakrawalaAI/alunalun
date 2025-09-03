-- name: CreatePost :one
INSERT INTO posts (id, user_id, type, content, parent_id, metadata, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPostByID :one
SELECT * FROM posts WHERE id = $1;

-- name: GetPostWithAuthor :one
SELECT 
    p.*,
    u.id as author_id,
    u.username as author_username,
    u.email as author_email
FROM posts p
JOIN users u ON p.user_id = u.id
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

-- name: ListCommentsByPost :many
SELECT 
    p.*,
    u.username as author_username
FROM posts p
JOIN users u ON p.user_id = u.id
WHERE p.parent_id = $1
ORDER BY p.created_at ASC;

-- name: CountCommentsByPost :one
SELECT COUNT(*) FROM posts
WHERE parent_id = $1;

-- name: UpdatePost :one
UPDATE posts
SET 
    content = $2,
    metadata = $3,
    updated_at = $4
WHERE id = $1
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1;

-- name: ListRecentPins :many
SELECT 
    p.*,
    u.username as author_username,
    COALESCE(comment_counts.count, 0) as comment_count
FROM posts p
JOIN users u ON p.user_id = u.id
LEFT JOIN (
    SELECT parent_id, COUNT(*) as count
    FROM posts
    WHERE parent_id IS NOT NULL
    GROUP BY parent_id
) comment_counts ON p.id = comment_counts.parent_id
WHERE p.type = 'pin'
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountCommentsByPinID :one
SELECT COUNT(*) FROM posts
WHERE parent_id = $1 AND type = 'comment';

-- name: GetCommentsByPinID :many
SELECT * FROM posts
WHERE parent_id = $1 AND type = 'comment'
ORDER BY created_at ASC;