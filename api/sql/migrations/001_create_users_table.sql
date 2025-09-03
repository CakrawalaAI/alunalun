-- Create users table
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE,
    username TEXT NOT NULL UNIQUE,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- Index for email lookups (NULL-safe)
CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;

-- Index for username lookups
CREATE INDEX idx_users_username ON users(username);

-- Index for created_at for sorting
CREATE INDEX idx_users_created_at ON users(created_at DESC);