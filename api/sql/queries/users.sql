-- name: CreateUser :one
INSERT INTO users (email, display_name, avatar_url)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT u.*, un.username
FROM users u
JOIN usernames un ON un.user_id = u.id
WHERE u.id = $1;

-- name: GetUserByEmail :one
SELECT u.*, un.username
FROM users u
JOIN usernames un ON un.user_id = u.id
WHERE u.email = $1;

-- name: GetUserByUsername :one
SELECT u.*, un.username
FROM users u
JOIN usernames un ON un.user_id = u.id
WHERE un.username = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET 
  display_name = COALESCE($2, display_name),
  avatar_url = COALESCE($3, avatar_url),
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;