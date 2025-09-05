-- name: CreatePostLocation :one
INSERT INTO posts_location (post_id, coordinates, geohash, created_at)
VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), $4, $5)
RETURNING 
    post_id,
    ST_X(coordinates) as longitude,
    ST_Y(coordinates) as latitude,
    geohash,
    created_at;

-- name: GetPostLocation :one
SELECT 
    post_id,
    ST_X(coordinates) as longitude,
    ST_Y(coordinates) as latitude,
    geohash,
    created_at
FROM posts_location
WHERE post_id = $1;

-- name: ListPinsInBoundingBox :many
SELECT 
    p.*,
    u.username as author_username,
    u.display_name as author_display_name,
    u.avatar_url as author_avatar_url,
    ST_X(pl.coordinates) as longitude,
    ST_Y(pl.coordinates) as latitude,
    pl.geohash
FROM posts p
LEFT JOIN users u ON p.user_id = u.id
JOIN posts_location pl ON p.id = pl.post_id
WHERE p.type = 'pin' 
    AND p.created_at > NOW() - INTERVAL '24 hours'
    AND ST_Within(
        pl.coordinates,
        ST_MakeEnvelope($1, $2, $3, $4, 4326)
    )
ORDER BY p.created_at DESC
LIMIT $5 OFFSET $6;

-- name: ListNearbyPins :many
SELECT 
    p.*,
    u.username as author_username,
    u.display_name as author_display_name,
    u.avatar_url as author_avatar_url,
    ST_X(pl.coordinates) as longitude,
    ST_Y(pl.coordinates) as latitude,
    pl.geohash,
    ST_Distance(
        pl.coordinates::geography,
        ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
    ) as distance_meters
FROM posts p
LEFT JOIN users u ON p.user_id = u.id
JOIN posts_location pl ON p.id = pl.post_id
WHERE p.type = 'pin' 
    AND p.created_at > NOW() - INTERVAL '24 hours'
    AND ST_DWithin(
        pl.coordinates::geography,
        ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
        $3  -- radius in meters
    )
ORDER BY distance_meters ASC
LIMIT $4;

-- name: DeletePostLocation :exec
DELETE FROM posts_location WHERE post_id = $1;

-- name: ListPinsByGeohash :many
SELECT 
    p.*,
    u.username as author_username,
    u.display_name as author_display_name,
    u.avatar_url as author_avatar_url,
    ST_X(pl.coordinates) as longitude,
    ST_Y(pl.coordinates) as latitude,
    pl.geohash
FROM posts p
LEFT JOIN users u ON p.user_id = u.id
JOIN posts_location pl ON p.id = pl.post_id
WHERE p.type = 'pin' 
    AND p.created_at > NOW() - INTERVAL '24 hours'
    AND pl.geohash LIKE $1 || '%'
ORDER BY p.created_at DESC
LIMIT $2;

-- name: GetPinWithLocation :one
SELECT 
    p.*,
    u.username as author_username,
    u.display_name as author_display_name,
    u.avatar_url as author_avatar_url,
    ST_X(pl.coordinates) as longitude,
    ST_Y(pl.coordinates) as latitude,
    pl.geohash,
    pl.created_at as location_created_at
FROM posts p
LEFT JOIN users u ON p.user_id = u.id
JOIN posts_location pl ON p.id = pl.post_id
WHERE p.id = $1 AND p.type = 'pin';

-- name: CountPinsInArea :one
SELECT COUNT(*) 
FROM posts p
JOIN posts_location pl ON p.id = pl.post_id
WHERE p.type = 'pin'
    AND p.created_at > NOW() - INTERVAL '24 hours'
    AND ST_Within(
        pl.coordinates,
        ST_MakeEnvelope($1, $2, $3, $4, 4326)
    );