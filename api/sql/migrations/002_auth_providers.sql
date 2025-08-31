-- +goose Up
-- +goose StatementBegin

-- OAuth and auth provider tracking
CREATE TABLE user_auth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,           -- 'google', 'magic_link', etc
    provider_user_id VARCHAR(255) NOT NULL,  -- google_id, etc
    email_verified BOOLEAN DEFAULT FALSE,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);

-- Event tracking for abuse detection
CREATE TABLE user_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,                   -- NULL for anonymous
    session_id UUID,                 -- For anonymous tracking
    event_type VARCHAR(50) NOT NULL, -- 'signup', 'login', 'post', 'comment', etc
    ip_address INET,
    x_forwarded_for TEXT,            -- Proxy detection
    x_real_ip INET,                  -- CDN/proxy real IP
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB                   -- Event-specific data
);

-- Indexes
CREATE INDEX idx_user_auth_providers_user ON user_auth_providers(user_id);
CREATE INDEX idx_user_auth_providers_provider ON user_auth_providers(provider, provider_user_id);
CREATE INDEX idx_user_events_user ON user_events(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_user_events_session ON user_events(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX idx_user_events_created ON user_events(created_at);
CREATE INDEX idx_user_events_type ON user_events(event_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_events;
DROP TABLE IF EXISTS user_auth_providers;
-- +goose StatementEnd