CREATE TABLE audit_logs (
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL,
    vault_id   UUID,
    action     TEXT        NOT NULL,
    detail     TEXT,
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create initial partitions (current + next month)
CREATE TABLE audit_logs_default PARTITION OF audit_logs DEFAULT;

CREATE INDEX idx_audit_logs_user_id    ON audit_logs (user_id);
CREATE INDEX idx_audit_logs_vault_id   ON audit_logs (vault_id) WHERE vault_id IS NOT NULL;
CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at);
