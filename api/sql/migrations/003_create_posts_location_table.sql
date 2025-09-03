-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create posts_location table
CREATE TABLE posts_location (
    post_id TEXT PRIMARY KEY REFERENCES posts(id) ON DELETE CASCADE,
    location GEOMETRY(Point, 4326) NOT NULL,
    geohash TEXT,
    created_at TIMESTAMPTZ NOT NULL
);

-- Spatial index for location queries
CREATE INDEX idx_posts_location_gist ON posts_location USING GIST (location);

-- Index for geohash proximity queries
CREATE INDEX idx_posts_location_geohash ON posts_location(geohash) WHERE geohash IS NOT NULL;