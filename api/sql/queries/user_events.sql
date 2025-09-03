-- name: CreateUserEvent :one
INSERT INTO user_events (id, user_id, session_id, event_type, ip_address, fingerprint, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: CountRecentEventsByIP :one
SELECT COUNT(*) FROM user_events
WHERE ip_address = $1
    AND created_at > $2
    AND event_type = $3;

-- name: CountRecentEventsByFingerprint :one
SELECT COUNT(*) FROM user_events
WHERE fingerprint = $1
    AND created_at > $2
    AND event_type = $3;

-- name: GetRecentEventsBySession :many
SELECT * FROM user_events
WHERE session_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: GetRecentEventsByUser :many
SELECT * FROM user_events
WHERE user_id = $1
    AND created_at > $2
ORDER BY created_at DESC;

-- name: CheckRateLimit :one
SELECT 
    COUNT(*) as total_events,
    COUNT(DISTINCT session_id) as unique_sessions,
    COUNT(DISTINCT COALESCE(fingerprint, session_id)) as unique_devices
FROM user_events
WHERE ip_address = $1
    AND created_at > $2
    AND event_type = ANY($3::text[]);

-- name: GetSuspiciousActivity :many
SELECT 
    ip_address,
    COUNT(*) as event_count,
    COUNT(DISTINCT session_id) as session_count,
    COUNT(DISTINCT user_id) as user_count,
    array_agg(DISTINCT event_type) as event_types
FROM user_events
WHERE created_at > $1
GROUP BY ip_address
HAVING COUNT(*) > $2
ORDER BY event_count DESC
LIMIT $3;