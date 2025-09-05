-- Enable PostGIS extension for spatial data
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create posts_location table for spatial data
CREATE TABLE posts_location (
    post_id UUID PRIMARY KEY REFERENCES posts(id) ON DELETE CASCADE,
    coordinates GEOMETRY(POINT, 4326) NOT NULL,
    geohash VARCHAR(12) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Spatial index for coordinate-based queries
CREATE INDEX idx_posts_location_coordinates ON posts_location USING GIST(coordinates);

-- Index for geohash prefix matching (handles both exact and pattern queries)
CREATE INDEX idx_posts_location_geohash ON posts_location(geohash text_pattern_ops);