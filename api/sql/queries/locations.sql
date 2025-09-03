-- name: CreatePostLocation :one
INSERT INTO posts_location (post_id, location, geohash, created_at)
VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), $4, $5)
RETURNING 
    post_id,
    ST_X(location) as longitude,
    ST_Y(location) as latitude,
    geohash,
    created_at;

-- name: GetPostLocation :one
SELECT 
    post_id,
    ST_X(location) as longitude,
    ST_Y(location) as latitude,
    geohash,
    created_at
FROM posts_location
WHERE post_id = $1;

-- name: ListPinsInBoundingBox :many
SELECT 
    p.*,
    u.username as author_username,
    ST_X(pl.location) as longitude,
    ST_Y(pl.location) as latitude,
    pl.geohash,
    COALESCE(comment_counts.count, 0) as comment_count
FROM posts p
JOIN users u ON p.user_id = u.id
JOIN posts_location pl ON p.id = pl.post_id
LEFT JOIN (
    SELECT parent_id, COUNT(*) as count
    FROM posts
    WHERE parent_id IS NOT NULL
    GROUP BY parent_id
) comment_counts ON p.id = comment_counts.parent_id
WHERE p.type = 'pin'
    AND ST_Within(
        pl.location,
        ST_MakeEnvelope($1, $2, $3, $4, 4326)
    )
ORDER BY p.created_at DESC
LIMIT $5 OFFSET $6;

-- name: ListNearbyPins :many
SELECT 
    p.*,
    u.username as author_username,
    ST_X(pl.location) as longitude,
    ST_Y(pl.location) as latitude,
    pl.geohash,
    ST_Distance(
        pl.location::geography,
        ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
    ) as distance_meters,
    COALESCE(comment_counts.count, 0) as comment_count
FROM posts p
JOIN users u ON p.user_id = u.id
JOIN posts_location pl ON p.id = pl.post_id
LEFT JOIN (
    SELECT parent_id, COUNT(*) as count
    FROM posts
    WHERE parent_id IS NOT NULL
    GROUP BY parent_id
) comment_counts ON p.id = comment_counts.parent_id
WHERE p.type = 'pin'
    AND ST_DWithin(
        pl.location::geography,
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
    pl.*
FROM posts p
JOIN posts_location pl ON p.id = pl.post_id
WHERE p.type = 'pin'
    AND pl.geohash LIKE $1 || '%'
ORDER BY p.created_at DESC
LIMIT $2;

-- name: GetPinWithLocation :one
SELECT 
    p.*,
    pl.*
FROM posts p
JOIN posts_location pl ON p.id = pl.post_id
WHERE p.id = $1
    AND p.type = 'pin';