-- Create user_auth_providers table for OAuth providers
CREATE TABLE user_auth_providers (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    provider_metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    UNIQUE(provider, provider_user_id)
);

-- Index for getting user's auth providers
CREATE INDEX idx_auth_providers_user_id ON user_auth_providers(user_id);