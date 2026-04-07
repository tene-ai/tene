-- Rollback 000008: Remove added constraints and columns

DROP TRIGGER IF EXISTS trg_devices_updated_at ON devices;
ALTER TABLE devices DROP COLUMN IF EXISTS updated_at;

DROP TRIGGER IF EXISTS trg_teams_updated_at ON teams;
ALTER TABLE teams DROP COLUMN IF EXISTS updated_at;

ALTER TABLE teams DROP CONSTRAINT IF EXISTS teams_owner_id_fkey;
ALTER TABLE teams
    ADD CONSTRAINT teams_owner_id_fkey
    FOREIGN KEY (owner_id) REFERENCES users(id);

ALTER TABLE vaults DROP CONSTRAINT IF EXISTS fk_vaults_team_id;
