-- +goose Up
-- +goose StatementBegin

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Single table for ALL usernames (enforces uniqueness across tiers)
CREATE TABLE usernames (
    username VARCHAR(100) PRIMARY KEY,
    user_id UUID,           -- NULL for anonymous
    session_id UUID,        -- NULL for authenticated
    claimed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Either user_id OR session_id must be set, not both
    CONSTRAINT one_owner CHECK (
        (user_id IS NOT NULL AND session_id IS NULL) OR
        (user_id IS NULL AND session_id IS NOT NULL)
    ),
    
    -- Username format validation
    CONSTRAINT username_format CHECK (
        username ~ '^[a-zA-Z0-9_]{3,100}$'
    )
);

-- Users table for authenticated accounts
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Email format validation
    CONSTRAINT email_format CHECK (
        email ~ '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    )
);

-- Indexes for performance
CREATE INDEX idx_usernames_user_id ON usernames(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_usernames_session_id ON usernames(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Updated at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to users table
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS usernames;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "pgcrypto";
-- +goose StatementEnd