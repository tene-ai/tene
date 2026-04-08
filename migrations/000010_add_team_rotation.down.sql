-- 000010: Remove key rotation columns from teams table

ALTER TABLE teams DROP COLUMN IF EXISTS rotation_pending;
ALTER TABLE teams DROP COLUMN IF EXISTS rotation_version;
