-- name: CreateAuthProvider :one
INSERT INTO user_auth_providers (
    user_id, provider, provider_user_id, email_verified, is_primary
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetUserByAuthProvider :one
SELECT u.*, un.username FROM users u
JOIN user_auth_providers uap ON u.id = uap.user_id
JOIN usernames un ON un.user_id = u.id
WHERE uap.provider = $1 AND uap.provider_user_id = $2;

-- name: GetAuthProvidersByUserID :many
SELECT * FROM user_auth_providers 
WHERE user_id = $1 
ORDER BY is_primary DESC, created_at ASC;

-- name: UpdateAuthProviderLastUsed :exec
UPDATE user_auth_providers 
SET last_used_at = NOW() 
WHERE id = $1;

-- name: UnlinkAuthProvider :exec
DELETE FROM user_auth_providers 
WHERE user_id = $1 AND provider = $2 
AND NOT is_primary;