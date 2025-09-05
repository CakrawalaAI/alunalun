-- Create posts table (polymorphic: pins, comments, etc.)
CREATE TABLE posts (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    parent_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    visibility VARCHAR(20) DEFAULT 'public',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Essential indexes only
CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_parent_id ON posts(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_posts_type_created ON posts(type, created_at DESC);