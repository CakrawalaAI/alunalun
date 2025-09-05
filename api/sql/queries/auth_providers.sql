-- name: CreateUserAuthProvider :one
INSERT INTO user_auth_providers (id, user_id, provider, provider_user_id, provider_metadata, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserAuthProviderByID :one
SELECT * FROM user_auth_providers WHERE id = $1;

-- name: GetUserAuthProviderByProviderID :one
SELECT * FROM user_auth_providers 
WHERE provider = $1 AND provider_user_id = $2;

-- name: ListUserAuthProviders :many
SELECT * FROM user_auth_providers
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteUserAuthProvider :exec
DELETE FROM user_auth_providers WHERE id = $1;

-- name: CountUserAuthProviders :one
SELECT COUNT(*) FROM user_auth_providers WHERE user_id = $1;