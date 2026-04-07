CREATE TABLE vaults (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id      UUID,
    project_name TEXT        NOT NULL,
    s3_key       TEXT        NOT NULL,
    vault_version INTEGER    NOT NULL DEFAULT 1,
    vault_hash   BYTEA       NOT NULL CHECK (octet_length(vault_hash) = 32),
    secret_count INTEGER     DEFAULT 0,
    size         BIGINT      DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_pushed_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_vaults_user_project ON vaults (user_id, project_name);
CREATE        INDEX idx_vaults_user_id      ON vaults (user_id);
CREATE        INDEX idx_vaults_team_id      ON vaults (team_id) WHERE team_id IS NOT NULL;

CREATE TRIGGER trg_vaults_updated_at
    BEFORE UPDATE ON vaults
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
