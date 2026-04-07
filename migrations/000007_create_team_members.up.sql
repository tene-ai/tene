CREATE TABLE team_members (
    team_id             UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role                TEXT NOT NULL CHECK (role IN ('admin', 'member')),
    env_permissions     JSONB DEFAULT '["dev"]',
    wrapped_project_key BYTEA,
    joined_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (team_id, user_id)
);

CREATE INDEX idx_team_members_user_id ON team_members (user_id);
