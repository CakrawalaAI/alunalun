-- name: CreateUserEvent :one
INSERT INTO user_events (
    user_id, session_id, event_type, ip_address, 
    x_forwarded_for, x_real_ip, user_agent, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetRecentUserEvents :many
SELECT * FROM user_events 
WHERE user_id = $1 
AND created_at > $2 
ORDER BY created_at DESC 
LIMIT $3;

-- name: GetRecentSessionEvents :many
SELECT * FROM user_events 
WHERE session_id = $1 
AND created_at > $2 
ORDER BY created_at DESC 
LIMIT $3;

-- name: CountEventsByType :one
SELECT COUNT(*) as count
FROM user_events
WHERE (user_id = $1 OR session_id = $2)
AND event_type = $3
AND created_at > $4;

-- name: CountRecentEvents :one
SELECT COUNT(*) as count
FROM user_events
WHERE (user_id = $1 OR session_id = $2)
AND created_at > $3;