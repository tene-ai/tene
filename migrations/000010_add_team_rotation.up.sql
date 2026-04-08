-- 000010: Add key rotation tracking columns to teams table
-- Supports team key rotation on member removal (S-07)

ALTER TABLE teams ADD COLUMN IF NOT EXISTS rotation_version BIGINT NOT NULL DEFAULT 0;
ALTER TABLE teams ADD COLUMN IF NOT EXISTS rotation_pending BOOLEAN NOT NULL DEFAULT false;
