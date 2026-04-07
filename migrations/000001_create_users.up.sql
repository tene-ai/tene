CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email             TEXT NOT NULL,
    name              TEXT,
    auth_provider     TEXT NOT NULL CHECK (auth_provider IN ('github', 'google')),
    github_id         BIGINT,
    google_id         TEXT,
    avatar_url        TEXT,
    plan              TEXT NOT NULL DEFAULT 'free' CHECK (plan IN ('free', 'pro')),
    lemon_customer_id TEXT,
    x25519_public_key BYTEA CHECK (x25519_public_key IS NULL OR octet_length(x25519_public_key) = 32),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_users_email       ON users (email);
CREATE UNIQUE INDEX idx_users_github_id   ON users (github_id)   WHERE github_id IS NOT NULL;
CREATE UNIQUE INDEX idx_users_google_id   ON users (google_id)   WHERE google_id IS NOT NULL;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Refresh tokens for JWT rotation
CREATE TABLE refresh_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA       NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_refresh_tokens_hash    ON refresh_tokens (token_hash);
CREATE        INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE        INDEX idx_refresh_tokens_expires  ON refresh_tokens (expires_at) WHERE revoked_at IS NULL;
