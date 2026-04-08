-- Add family column to refresh_tokens for H-04 token reuse detection.
ALTER TABLE refresh_tokens ADD COLUMN IF NOT EXISTS family TEXT;
CREATE INDEX idx_refresh_tokens_family ON refresh_tokens (family) WHERE revoked_at IS NULL;
