-- 000008: Add missing FK constraints, cascade deletes, and updated_at columns
-- Addresses: audit_logs FK, vaults.team_id FK, teams cascade + updated_at, devices updated_at

-- 1. Add FK constraint to vaults.team_id (references teams)
ALTER TABLE vaults
    ADD CONSTRAINT fk_vaults_team_id
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL;

-- 2. Fix teams.owner_id to cascade on user deletion
ALTER TABLE teams DROP CONSTRAINT IF EXISTS teams_owner_id_fkey;
ALTER TABLE teams
    ADD CONSTRAINT teams_owner_id_fkey
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;

-- 3. Add updated_at to teams table
ALTER TABLE teams ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE TRIGGER trg_teams_updated_at
    BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- 4. Add updated_at to devices table
ALTER TABLE devices ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE TRIGGER trg_devices_updated_at
    BEFORE UPDATE ON devices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Note: audit_logs FK constraints are intentionally omitted because
-- partitioned tables in PostgreSQL do not support foreign key references
-- to non-partitioned tables. Referential integrity for audit_logs is
-- enforced at the application layer instead.
