CREATE TABLE vault_key_metadata (
    vault_id    UUID NOT NULL REFERENCES vaults(id) ON DELETE CASCADE,
    environment TEXT NOT NULL,
    key_name    TEXT NOT NULL,
    version     INTEGER NOT NULL DEFAULT 1,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (vault_id, environment, key_name)
);

CREATE INDEX idx_vault_key_metadata_vault ON vault_key_metadata(vault_id);
