CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    verification_code TEXT NULL,
    verification_code_expires_at TIMESTAMPTZ NULL,
    role TEXT NOT NULL CHECK (role IN ('user', 'admin', 'support')),
    status TEXT NOT NULL CHECK (status IN ('pending', 'active', 'locked')),
    username TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    last_login_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);