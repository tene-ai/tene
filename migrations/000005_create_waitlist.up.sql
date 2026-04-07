CREATE TABLE waitlist (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email      TEXT        NOT NULL,
    plan       TEXT        DEFAULT 'pro',
    source     TEXT        DEFAULT 'cli',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_waitlist_email ON waitlist (email);
