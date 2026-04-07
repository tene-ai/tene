CREATE TABLE devices (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_name       TEXT        NOT NULL,
    x25519_public_key BYTEA       CHECK (x25519_public_key IS NULL OR octet_length(x25519_public_key) = 32),
    last_seen_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_devices_user_id ON devices (user_id);
