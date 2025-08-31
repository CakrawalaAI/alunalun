-- name: CheckUsername :one
SELECT COUNT(*) > 0 AS taken
FROM usernames
WHERE username = $1;

-- name: ClaimUsernameForAnonymous :one
INSERT INTO usernames (username, session_id)
VALUES ($1, $2)
ON CONFLICT (username) DO NOTHING
RETURNING username, session_id;

-- name: ClaimUsernameForUser :one
INSERT INTO usernames (username, user_id)
VALUES ($1, $2)
ON CONFLICT (username) DO NOTHING
RETURNING username, user_id;

-- name: GetUsernameOwner :one
SELECT username, user_id, session_id
FROM usernames
WHERE username = $1;

-- name: MigrateUsernameToUser :exec
UPDATE usernames
SET user_id = $2, session_id = NULL
WHERE session_id = $1;