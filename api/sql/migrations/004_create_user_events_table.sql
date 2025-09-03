-- Create user_events table for tracking and abuse prevention
CREATE TABLE user_events (
    id TEXT PRIMARY KEY,
    user_id TEXT REFERENCES users(id),
    session_id TEXT NOT NULL,
    event_type TEXT NOT NULL, -- 'create_pin', 'create_comment', 'view'
    ip_address INET NOT NULL,
    fingerprint TEXT,
    created_at TIMESTAMPTZ NOT NULL
);

-- Index for rate limiting by IP
CREATE INDEX idx_user_events_ip_created ON user_events(ip_address, created_at DESC);

-- Index for session tracking
CREATE INDEX idx_user_events_session ON user_events(session_id);

-- Index for user activity tracking
CREATE INDEX idx_user_events_user_id ON user_events(user_id) WHERE user_id IS NOT NULL;

-- Index for fingerprint tracking (abuse detection)
CREATE INDEX idx_user_events_fingerprint ON user_events(fingerprint) WHERE fingerprint IS NOT NULL;

-- Index for event type analysis
CREATE INDEX idx_user_events_type_created ON user_events(event_type, created_at DESC);