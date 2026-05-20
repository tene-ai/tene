package vault

// schemaSQL is the schema v1 baseline used by initSchema() for a brand-new
// vault.db. The v2 ALTER (preview column on secrets) is applied separately
// by migrate(): when a vault is created fresh we run initSchema then bump
// the version, but the v2 ADD COLUMN is also run unconditionally so the
// preview column is always present regardless of whether the user came from
// init (v1) or from an existing v1 vault (upgrade path).
const schemaSQL = `
CREATE TABLE IF NOT EXISTS vault_meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS secrets (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT    NOT NULL,
    encrypted_value TEXT    NOT NULL,
    environment     TEXT    NOT NULL DEFAULT 'default',
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE(name, environment)
);

CREATE INDEX IF NOT EXISTS idx_secrets_env ON secrets(environment);
CREATE INDEX IF NOT EXISTS idx_secrets_name ON secrets(name);

CREATE TABLE IF NOT EXISTS environments (
    name       TEXT    PRIMARY KEY,
    is_active  INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS audit_log (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    action        TEXT NOT NULL,
    resource_name TEXT,
    details       TEXT,
    timestamp     TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_log(timestamp);
`

// currentSchemaVersion is the schema_version vault_meta value that a
// fully-migrated vault.db must report. Bumped from 1 -> 2 by the
// cli-ux-permission-model sprint (Q2 decision) to introduce the
// `secrets.preview` column.
const currentSchemaVersion = 2

// schemaMetaKey is the vault_meta key under which the schema version is
// stored. Centralized as a constant so migration code and tests cannot
// drift on the spelling.
const schemaMetaKey = "schema_version"

func (v *Vault) initSchema() error {
	_, err := v.db.Exec(schemaSQL)
	return err
}
