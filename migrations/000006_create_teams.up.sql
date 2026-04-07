CREATE TABLE teams (
    id                    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name                  TEXT        NOT NULL,
    slug                  TEXT        NOT NULL,
    owner_id              UUID        NOT NULL REFERENCES users(id),
    lemon_subscription_id TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_teams_slug     ON teams (slug);
CREATE        INDEX idx_teams_owner_id ON teams (owner_id);
