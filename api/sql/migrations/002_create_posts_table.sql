-- Create posts table (polymorphic: pins, comments, posts)
CREATE TABLE posts (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    type TEXT NOT NULL, -- 'pin', 'comment', 'post'
    content TEXT,
    parent_id TEXT REFERENCES posts(id),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Index for user's posts
CREATE INDEX idx_posts_user_id ON posts(user_id);

-- Index for post type filtering
CREATE INDEX idx_posts_type ON posts(type);

-- Index for comment threads (parent_id)
CREATE INDEX idx_posts_parent_id ON posts(parent_id) WHERE parent_id IS NOT NULL;

-- Index for chronological ordering
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);

-- Composite index for common query pattern (get comments for a post)
CREATE INDEX idx_posts_parent_created ON posts(parent_id, created_at) WHERE parent_id IS NOT NULL;