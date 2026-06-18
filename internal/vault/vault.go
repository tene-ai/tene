package vault

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/tene-ai/tene/pkg/domain"
	_ "modernc.org/sqlite"
)

// Vault is a SQLite-based secret store.
type Vault struct {
	db     *sql.DB
	dbPath string
}

// New creates or opens a SQLite vault at the given path and initializes the schema.
func New(dbPath string) (*Vault, error) {
	// Ensure parent directory exists with proper permissions
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("vault: failed to create directory: %w", err)
	}

	// busy_timeout is appended to the DSN so it is set on every connection
	// the database/sql pool opens, not just on the first one. Without this,
	// a second tene process arriving while the first holds a write lock
	// would get SQLITE_BUSY immediately. 5000ms is enough for a schema-v2
	// ALTER TABLE to complete on any realistic vault.db.
	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("vault: failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrency. PRAGMA journal_mode=WAL
	// requires an exclusive lock; with busy_timeout set above, concurrent
	// callers wait instead of failing.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("vault: failed to set WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("vault: failed to enable foreign keys: %w", err)
	}

	v := &Vault{db: db, dbPath: dbPath}

	if err := v.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("vault: migration failed: %w", err)
	}

	// Set file permissions to 0600 (best-effort: some FS don't support chmod)
	if err := os.Chmod(dbPath, 0600); err != nil {
		slog.Warn("vault: chmod 0600 failed", "path", dbPath, "error", err)
	}

	return v, nil
}

// Close closes the database connection.
func (v *Vault) Close() error {
	if v.db != nil {
		return v.db.Close()
	}
	return nil
}

// migrate ensures vault.db's schema is at currentSchemaVersion. For a fresh
// database it runs initSchema() to create the baseline tables, then delegates
// to runMigrations() to apply forward-step migrations (see migration.go).
//
// For an already-initialized database it skips initSchema (the CREATE TABLE
// statements are IF NOT EXISTS so the call would be safe, but skipping
// avoids an unnecessary write txn) and goes straight to runMigrations(),
// which uses readSchemaVersion + per-step transactions to bring the database
// up to date without losing data.
func (v *Vault) migrate() error {
	fresh, err := v.isFreshVault()
	if err != nil {
		return fmt.Errorf("vault: detect fresh vault: %w", err)
	}
	if fresh {
		if err := v.initSchema(); err != nil {
			return fmt.Errorf("vault: init schema: %w", err)
		}
	}
	return v.runMigrations()
}

// isFreshVault returns true when vault_meta does not exist yet, which is
// the unambiguous marker of a not-yet-initialized vault.db. We probe the
// sqlite_master table (always present, even on an empty database) so we
// never get a false positive from a transient error like ErrMetaNotFound
// on a wholly different missing key.
func (v *Vault) isFreshVault() (bool, error) {
	var name string
	err := v.db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type='table' AND name='vault_meta'`,
	).Scan(&name)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

// --- Metadata ---

// SetMeta stores vault metadata (UPSERT).
func (v *Vault) SetMeta(key, value string) error {
	query := `INSERT INTO vault_meta (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`
	_, err := v.db.Exec(query, key, value)
	if err != nil {
		return fmt.Errorf("vault: failed to set meta %q: %w", key, err)
	}
	return nil
}

// GetMeta retrieves vault metadata.
func (v *Vault) GetMeta(key string) (string, error) {
	var value string
	err := v.db.QueryRow("SELECT value FROM vault_meta WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("%w: %s", ErrMetaNotFound, key)
	}
	if err != nil {
		return "", fmt.Errorf("vault: failed to get meta %q: %w", key, err)
	}
	return value, nil
}

// --- Secret CRUD ---

// SetSecret stores a secret without touching the preview column (UPSERT
// by name+environment). Preserved for callers that do not have a preview
// (legacy or test code). New call sites should use SetSecretWithPreview so
// ciphertext and preview update atomically in the same row mutation.
func (v *Vault) SetSecret(name, encryptedValue, env string) error {
	return v.SetSecretWithPreview(name, encryptedValue, env, "")
}

// SetSecretWithPreview stores a secret AND its preview substring in a
// single atomic UPSERT.
//
// Atomicity is critical here: if we wrote the ciphertext first and the
// preview second, a process crash between the two statements would leave
// vault.db with the previous preview still pointing at the previous
// plaintext -- meaning `tene list` would show stale information. Doing
// both columns in one INSERT ... ON CONFLICT DO UPDATE bundles the writes
// into one SQLite transaction-implicit row mutation.
//
// Pass preview = "" to clear the preview (e.g. when preview.enabled=false
// is the active configuration). The column is NOT NULL DEFAULT ” (schema
// v2) so this stores the empty string, not NULL.
func (v *Vault) SetSecretWithPreview(name, encryptedValue, env, preview string) error {
	const query = `
		INSERT INTO secrets (name, encrypted_value, environment, preview, version, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1, datetime('now'), datetime('now'))
		ON CONFLICT(name, environment) DO UPDATE SET
			encrypted_value = excluded.encrypted_value,
			preview         = excluded.preview,
			version         = secrets.version + 1,
			updated_at      = datetime('now')
	`
	if _, err := v.db.Exec(query, name, encryptedValue, env, preview); err != nil {
		return fmt.Errorf("vault: failed to set secret %q: %w", name, err)
	}

	return v.AddAuditLog("secret.write", name, "")
}

// SetSecretBatchWithPreview stores multiple secrets (ciphertext + preview)
// in a single transaction. Counterpart of SetSecretBatch used by
// `tene import` so the imported set lands with all previews populated
// atomically.
func (v *Vault) SetSecretBatchWithPreview(records []SecretWrite, env string) error {
	tx, err := v.db.Begin()
	if err != nil {
		return fmt.Errorf("vault: failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const query = `
		INSERT INTO secrets (name, encrypted_value, environment, preview, version, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1, datetime('now'), datetime('now'))
		ON CONFLICT(name, environment) DO UPDATE SET
			encrypted_value = excluded.encrypted_value,
			preview         = excluded.preview,
			version         = secrets.version + 1,
			updated_at      = datetime('now')
	`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("vault: failed to prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, r := range records {
		if _, err := stmt.Exec(r.Name, r.EncryptedValue, env, r.Preview); err != nil {
			return fmt.Errorf("vault: failed to set secret %q: %w", r.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("vault: failed to commit transaction: %w", err)
	}
	return nil
}

// GetSecret retrieves a secret.
func (v *Vault) GetSecret(name, env string) (*Secret, error) {
	var s Secret
	var createdAt, updatedAt string
	err := v.db.QueryRow(
		"SELECT id, name, encrypted_value, environment, version, created_at, updated_at FROM secrets WHERE name = ? AND environment = ?",
		name, env,
	).Scan(&s.ID, &s.Name, &s.EncryptedValue, &s.Environment, &s.Version, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("%w: %s", ErrSecretNotFound, name)
	}
	if err != nil {
		return nil, fmt.Errorf("vault: failed to get secret %q: %w", name, err)
	}

	if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
		s.CreatedAt = t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", updatedAt); err == nil {
		s.UpdatedAt = t
	}
	return &s, nil
}

// ListSecrets returns all secrets for the given environment.
func (v *Vault) ListSecrets(env string) ([]Secret, error) {
	rows, err := v.db.Query(
		"SELECT id, name, encrypted_value, environment, version, created_at, updated_at FROM secrets WHERE environment = ? ORDER BY name",
		env,
	)
	if err != nil {
		return nil, fmt.Errorf("vault: failed to list secrets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var secrets []Secret
	for rows.Next() {
		var s Secret
		var createdAt, updatedAt string
		if err := rows.Scan(&s.ID, &s.Name, &s.EncryptedValue, &s.Environment, &s.Version, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("vault: failed to scan secret: %w", err)
		}
		if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
			s.CreatedAt = t
		}
		if t, err := time.Parse("2006-01-02 15:04:05", updatedAt); err == nil {
			s.UpdatedAt = t
		}
		secrets = append(secrets, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("vault: list secrets rows: %w", err)
	}
	return secrets, nil
}

// DeleteSecret deletes a secret.
func (v *Vault) DeleteSecret(name, env string) error {
	result, err := v.db.Exec("DELETE FROM secrets WHERE name = ? AND environment = ?", name, env)
	if err != nil {
		return fmt.Errorf("vault: failed to delete secret %q: %w", name, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vault: delete secret rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("%w: %s", ErrSecretNotFound, name)
	}

	return v.AddAuditLog("secret.delete", name, "")
}

// SecretExists checks if a secret exists.
func (v *Vault) SecretExists(name, env string) (bool, error) {
	var count int
	err := v.db.QueryRow("SELECT COUNT(*) FROM secrets WHERE name = ? AND environment = ?", name, env).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("vault: check secret exists: %w", err)
	}
	return count > 0, nil
}

// CountSecrets returns the number of secrets in the given environment.
func (v *Vault) CountSecrets(env string) (int, error) {
	var count int
	err := v.db.QueryRow("SELECT COUNT(*) FROM secrets WHERE environment = ?", env).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("vault: count secrets: %w", err)
	}
	return count, nil
}

// GetAllSecrets returns all secrets as a name->encryptedValue map.
func (v *Vault) GetAllSecrets(env string) (map[string]string, error) {
	rows, err := v.db.Query("SELECT name, encrypted_value FROM secrets WHERE environment = ?", env)
	if err != nil {
		return nil, fmt.Errorf("vault: failed to get all secrets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[string]string)
	for rows.Next() {
		var name, val string
		if err := rows.Scan(&name, &val); err != nil {
			return nil, fmt.Errorf("vault: scan secret: %w", err)
		}
		result[name] = val
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("vault: get all secrets rows: %w", err)
	}
	return result, nil
}

// SetSecretBatch stores multiple secrets in a single transaction.
func (v *Vault) SetSecretBatch(secrets map[string]string, env string) error {
	tx, err := v.db.Begin()
	if err != nil {
		return fmt.Errorf("vault: failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		INSERT INTO secrets (name, encrypted_value, environment, version, created_at, updated_at)
		VALUES (?, ?, ?, 1, datetime('now'), datetime('now'))
		ON CONFLICT(name, environment) DO UPDATE SET
			encrypted_value = excluded.encrypted_value,
			version = secrets.version + 1,
			updated_at = datetime('now')
	`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("vault: failed to prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for name, encVal := range secrets {
		if _, err := stmt.Exec(name, encVal, env); err != nil {
			return fmt.Errorf("vault: failed to set secret %q: %w", name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("vault: failed to commit transaction: %w", err)
	}
	return nil
}

// --- Environment Management ---

// ListEnvironments returns all environments.
func (v *Vault) ListEnvironments() ([]Environment, error) {
	rows, err := v.db.Query("SELECT name, is_active, created_at FROM environments ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("vault: failed to list environments: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var envs []Environment
	for rows.Next() {
		var e Environment
		var isActive int
		var createdAt string
		if err := rows.Scan(&e.Name, &isActive, &createdAt); err != nil {
			return nil, fmt.Errorf("vault: scan environment: %w", err)
		}
		e.IsActive = isActive == 1
		if t, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
			e.CreatedAt = t
		}
		envs = append(envs, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("vault: list environments rows: %w", err)
	}
	return envs, nil
}

// GetActiveEnvironment returns the active environment name.
func (v *Vault) GetActiveEnvironment() (string, error) {
	var name string
	err := v.db.QueryRow("SELECT name FROM environments WHERE is_active = 1").Scan(&name)
	if err == sql.ErrNoRows {
		return "default", nil
	}
	if err != nil {
		return "", fmt.Errorf("vault: failed to get active environment: %w", err)
	}
	return name, nil
}

// SetActiveEnvironment changes the active environment.
// Creates the environment if it doesn't exist.
func (v *Vault) SetActiveEnvironment(name string) error {
	tx, err := v.db.Begin()
	if err != nil {
		return fmt.Errorf("vault: set active environment begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Deactivate all
	if _, err := tx.Exec("UPDATE environments SET is_active = 0"); err != nil {
		return fmt.Errorf("vault: deactivate environments: %w", err)
	}

	// Insert or update the target environment
	query := `INSERT INTO environments (name, is_active, created_at)
		VALUES (?, 1, datetime('now'))
		ON CONFLICT(name) DO UPDATE SET is_active = 1`
	if _, err := tx.Exec(query, name); err != nil {
		return fmt.Errorf("vault: set active environment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("vault: set active environment commit: %w", err)
	}
	return nil
}

// CreateEnvironment creates a new environment.
func (v *Vault) CreateEnvironment(name string) error {
	_, err := v.db.Exec("INSERT INTO environments (name, is_active, created_at) VALUES (?, 0, datetime('now'))", name)
	if err != nil {
		// Check if it's a unique constraint violation
		return fmt.Errorf("%w: %s", ErrEnvironmentExists, name)
	}
	return nil
}

// DeleteEnvironment deletes an environment and all its secrets.
func (v *Vault) DeleteEnvironment(name string) (int, error) {
	tx, err := v.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("vault: delete environment begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Count secrets that will be deleted
	var count int
	if err := tx.QueryRow("SELECT COUNT(*) FROM secrets WHERE environment = ?", name).Scan(&count); err != nil {
		return 0, fmt.Errorf("vault: count environment secrets: %w", err)
	}

	// Delete secrets
	if _, err := tx.Exec("DELETE FROM secrets WHERE environment = ?", name); err != nil {
		return 0, fmt.Errorf("vault: delete environment secrets: %w", err)
	}

	// Delete environment
	result, err := tx.Exec("DELETE FROM environments WHERE name = ?", name)
	if err != nil {
		return 0, fmt.Errorf("vault: delete environment: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return 0, fmt.Errorf("%w: %s", ErrEnvironmentNotFound, name)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("vault: delete environment commit: %w", err)
	}
	return count, nil
}

// EnvironmentExists checks if an environment exists.
func (v *Vault) EnvironmentExists(name string) (bool, error) {
	var count int
	err := v.db.QueryRow("SELECT COUNT(*) FROM environments WHERE name = ?", name).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("vault: check environment exists: %w", err)
	}
	return count > 0, nil
}

// --- Metadata Read API (Q2 — no-decrypt path) ---

// ListSecretMetadata returns name + version + updated_at + preview for every
// secret in env, ordered by name.
//
// Security invariant (I-1, I-9): this method MUST NEVER read or expose the
// encrypted_value column. It exists so commands like `tene list` can render
// useful output without unlocking the vault, and so AI assistants can learn
// the canonical key names of a project without being able to see values.
// The preview column is plaintext per the Q2 user decision; its size is
// bounded by pkg/crypto.MaxPreviewChars (=8 rune cap), enforced at write
// time by callers of DerivePreview.
//
// Returns an empty slice (not nil) when the environment has no secrets, so
// JSON callers always get a deterministic array shape.
//
// Returns an error only on database-level failures (closed connection,
// schema corruption). A missing environment is not an error -- it's
// indistinguishable from "no secrets in that env" at this layer.
func (v *Vault) ListSecretMetadata(env string) ([]domain.VaultKeyMeta, error) {
	const q = `SELECT name, version, updated_at, preview
		FROM secrets
		WHERE environment = ?
		ORDER BY name`
	rows, err := v.db.Query(q, env)
	if err != nil {
		return nil, fmt.Errorf("vault: list secret metadata: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]domain.VaultKeyMeta, 0)
	for rows.Next() {
		var m domain.VaultKeyMeta
		var updatedAt string
		if err := rows.Scan(&m.Name, &m.Version, &updatedAt, &m.Preview); err != nil {
			return nil, fmt.Errorf("vault: scan secret metadata: %w", err)
		}
		if t, err := time.Parse("2006-01-02 15:04:05", updatedAt); err == nil {
			m.UpdatedAt = t
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("vault: list secret metadata rows: %w", err)
	}
	return out, nil
}

// UpdateSecretPreview rewrites only the preview column for a single secret.
//
// Used by `tene migrate fill-previews` to populate previews on a vault that
// was created before schema v2, and by `tene config preview.enabled=false`
// followed by `tene migrate fill-previews` to clear them. The ciphertext,
// version, and updated_at columns are intentionally untouched -- this is
// not a value change, it's a derived-cache refresh.
//
// Returns ErrSecretNotFound when no row matches name+env.
func (v *Vault) UpdateSecretPreview(name, env, preview string) error {
	result, err := v.db.Exec(
		`UPDATE secrets SET preview = ? WHERE name = ? AND environment = ?`,
		preview, name, env,
	)
	if err != nil {
		return fmt.Errorf("vault: update preview %q: %w", name, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vault: update preview rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("%w: %s", ErrSecretNotFound, name)
	}
	return nil
}

// ListSecretsForBackfill returns the rows that `tene migrate fill-previews`
// needs to operate on: for each secret in env whose preview is currently
// empty, the name + encrypted_value (so the caller can decrypt + derive).
//
// This is the only metadata-read API in the package that intentionally
// returns encrypted_value. It is named explicitly so callers cannot
// accidentally use it where ListSecretMetadata would have sufficed.
// `tene migrate fill-previews` is a PermSecretRead-tier operation (it
// requires unlock to derive previews from plaintext), so reading the
// ciphertext here is appropriate.
func (v *Vault) ListSecretsForBackfill(env string) ([]SecretBackfill, error) {
	const q = `SELECT name, encrypted_value
		FROM secrets
		WHERE environment = ? AND preview = ''
		ORDER BY name`
	rows, err := v.db.Query(q, env)
	if err != nil {
		return nil, fmt.Errorf("vault: list backfill candidates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]SecretBackfill, 0)
	for rows.Next() {
		var r SecretBackfill
		if err := rows.Scan(&r.Name, &r.EncryptedValue); err != nil {
			return nil, fmt.Errorf("vault: scan backfill row: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("vault: backfill rows: %w", err)
	}
	return out, nil
}

// --- Audit Log ---

// AddAuditLog adds an audit log entry.
func (v *Vault) AddAuditLog(action, resourceName, details string) error {
	_, err := v.db.Exec(
		"INSERT INTO audit_log (action, resource_name, details, timestamp) VALUES (?, ?, ?, datetime('now'))",
		action, resourceName, details,
	)
	if err != nil {
		return fmt.Errorf("vault: failed to add audit log: %w", err)
	}
	return nil
}

// AuditLogEntry is one row of audit_log returned by QueryAuditLog.
//
// Timestamp is decoded from SQLite's default text representation
// (`datetime('now')` produces `"YYYY-MM-DD HH:MM:SS"` in UTC). Callers
// that need RFC 3339 / ISO 8601 should format on the way out — this
// type preserves the column value as-stored so audit forensics can
// quote it back verbatim.
//
// Resource and Details are nullable in the schema; QueryAuditLog
// coerces NULL to "" so callers do not need to handle sql.NullString
// at every site. ID is preserved for stable ordering and as a stable
// reference downstream callers can cite (e.g. F8 NDJSON output).
type AuditLogEntry struct {
	ID        int64
	Action    string
	Resource  string
	Details   string
	Timestamp time.Time
}

// AuditLogFilter narrows a QueryAuditLog call to a subset of rows.
//
// All fields are optional; a zero-valued filter returns every row in
// reverse chronological order (newest first) up to MaxRows. A non-zero
// Since/Until uses inclusive bounds on the SQL side. ActionLike is
// applied with SQL LIKE — the caller is expected to add wildcards
// (`cli.metaread.%`) when prefix-matching is desired. ResourceLike is
// applied with SQL LIKE against the resource_name column; the audit
// CLI layer wraps the user-supplied substring in `%...%` so a bare
// `--resource STRIPE` matches any row whose resource_name contains
// "STRIPE" (e.g. STRIPE_KEY, STRIPE_WEBHOOK). All non-zero filter
// fields are AND-ed together.
//
// MaxRows = 0 means "unlimited" (sentinel for tail with no -n flag).
type AuditLogFilter struct {
	Since        time.Time
	Until        time.Time
	ActionLike   string
	ResourceLike string
	MaxRows      int
}

// QueryAuditLog returns audit_log rows matching the filter, newest
// first. The newest-first ordering is what `tene audit tail` / `tene
// audit show` want for human reading; reversing to chronological order
// is the caller's job if they need it.
//
// Privacy note: the action / resource / details columns may contain
// user-supplied names (e.g. STRIPE_KEY as resource_name on a
// secret.write row from set.go). They MUST NOT contain plaintext
// secret values — that invariant is enforced upstream at every
// AddAuditLog call site (I-5). QueryAuditLog itself does no
// sanitisation beyond NULL-coalescing.
func (v *Vault) QueryAuditLog(f AuditLogFilter) ([]AuditLogEntry, error) {
	// Build the WHERE clause incrementally so we only bind the
	// parameters that were actually set. text/template would be
	// overkill; string concatenation is safe here because every
	// dynamic value goes through a `?` placeholder.
	query := "SELECT id, action, COALESCE(resource_name, ''), COALESCE(details, ''), timestamp FROM audit_log"
	var conds []string
	var args []any

	if !f.Since.IsZero() {
		conds = append(conds, "timestamp >= ?")
		args = append(args, f.Since.UTC().Format("2006-01-02 15:04:05"))
	}
	if !f.Until.IsZero() {
		conds = append(conds, "timestamp <= ?")
		args = append(args, f.Until.UTC().Format("2006-01-02 15:04:05"))
	}
	if f.ActionLike != "" {
		conds = append(conds, "action LIKE ?")
		args = append(args, f.ActionLike)
	}
	if f.ResourceLike != "" {
		conds = append(conds, "resource_name LIKE ?")
		args = append(args, f.ResourceLike)
	}
	if len(conds) > 0 {
		query += " WHERE " + conds[0]
		for _, c := range conds[1:] {
			query += " AND " + c
		}
	}
	query += " ORDER BY id DESC"
	if f.MaxRows > 0 {
		query += " LIMIT ?"
		args = append(args, f.MaxRows)
	}

	rows, err := v.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("vault: query audit_log: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []AuditLogEntry
	for rows.Next() {
		var e AuditLogEntry
		var ts string
		if err := rows.Scan(&e.ID, &e.Action, &e.Resource, &e.Details, &ts); err != nil {
			return nil, fmt.Errorf("vault: scan audit_log row: %w", err)
		}
		// SQLite stores datetime('now') as "YYYY-MM-DD HH:MM:SS" UTC.
		// Parse forgivingly: a non-default format slipped in via direct
		// SQL would still surface (we keep the string in Action's
		// neighbouring context for debugging), but we don't refuse to
		// return the row — audit forensics over a tampered DB is still
		// a useful capability.
		parsed, perr := time.Parse("2006-01-02 15:04:05", ts)
		if perr != nil {
			// Try RFC 3339 as a secondary format (some legacy rows or
			// hand-inserted entries may use it).
			parsed, _ = time.Parse(time.RFC3339, ts)
		}
		e.Timestamp = parsed.UTC()
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("vault: audit_log rows: %w", err)
	}
	return out, nil
}

// GetAuditLogSize returns a rough byte estimate of the audit_log
// content (NOT the on-disk SQLite file size, which would include
// indices, free pages, and the WAL). The estimate is computed as
//
//	SUM(length(action) + length(resource_name) + length(details) + 32)
//
// where 32 is a per-row constant covering the timestamp string and
// integer id plus SQLite overhead. Accuracy: ±20%. This is precise
// enough for a 50MB threshold warning whose only requirement is "fire
// roughly when the audit_log is taking up noticeable disk".
//
// Returns 0 (not an error) when audit_log is empty. The SUM of an
// empty result set is NULL in SQLite; COALESCE smooths that to 0.
func (v *Vault) GetAuditLogSize() (int64, error) {
	var size sql.NullInt64
	err := v.db.QueryRow(
		`SELECT COALESCE(SUM(length(COALESCE(action, '')) +
			length(COALESCE(resource_name, '')) +
			length(COALESCE(details, '')) +
			32), 0)
		 FROM audit_log`,
	).Scan(&size)
	if err != nil {
		return 0, fmt.Errorf("vault: audit_log size: %w", err)
	}
	if !size.Valid {
		return 0, nil
	}
	return size.Int64, nil
}

// CountAuditLogOlderThan returns how many audit_log rows have a
// timestamp strictly older than cutoff. Used by `tene audit prune`'s
// dry-run path and by the confirmation prompt — both want a count
// without actually performing the DELETE.
//
// cutoff is interpreted as UTC; passing a non-UTC time.Time is fine
// (we call .UTC() ourselves) but the comparison is against the UTC
// timestamps SQLite stores.
func (v *Vault) CountAuditLogOlderThan(cutoff time.Time) (int64, error) {
	var n int64
	err := v.db.QueryRow(
		"SELECT COUNT(*) FROM audit_log WHERE timestamp < ?",
		cutoff.UTC().Format("2006-01-02 15:04:05"),
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("vault: audit_log count: %w", err)
	}
	return n, nil
}

// PruneAuditLog is the SINGLE source of truth for removing rows from
// audit_log. It must be the only place in the entire codebase that
// issues `DELETE FROM audit_log` — that invariant is enforced by
// quality gate G10 (master-plan.md §5 G10) via a static grep test in
// internal/audit/g10_test.go.
//
// Why centralised here:
//
//   - Forensic preservation (master-plan.md §10 I-14): audit logs must
//     never be auto-deleted. Concentrating the DELETE in one chokepoint
//     means a static check (`grep -rn "DELETE FROM audit_log"`) trips
//     immediately if a future commit adds a "log rotation" path.
//
//   - Atomicity: we use a deferred-write transaction; the DELETE is the
//     only statement, so coarse locking + commit semantics suffice for
//     readers seeing either all-old-rows-still-present or
//     all-old-rows-removed, never half. modernc.org/sqlite does not
//     expose `BEGIN IMMEDIATE` (statement returns "cannot start a
//     transaction within a transaction" when issued inside an already-
//     open db.Begin() transaction) — see migration.go for the matching
//     rationale. The atomicity guarantee is unaffected because there is
//     exactly one mutating statement inside this transaction.
//
//   - Returning the rows-affected count lets the CLI report "Deleted
//     N rows" without a follow-up SELECT.
//
// cutoff is converted to UTC and formatted with the same layout SQLite
// uses internally, so the < comparison is lexicographic AND
// chronological (SQLite's default ISO-like timestamp format is
// monotone under string comparison).
//
// The caller (audit_cmd.go) is responsible for the safety policy:
// requiring `--force` or an interactive confirm, and refusing to drop
// the master-key unlock for the PermSecretWrite tier. PruneAuditLog
// itself runs unconditionally once invoked — it trusts the caller to
// have done that gating.
func (v *Vault) PruneAuditLog(cutoff time.Time) (int64, error) {
	tx, err := v.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("vault: begin prune tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	res, err := tx.Exec(
		// G10 chokepoint — DELETE FROM audit_log appears here exactly
		// once in the entire repo. Static gate enforced by
		// internal/audit/g10_test.go.
		"DELETE FROM audit_log WHERE timestamp < ?",
		cutoff.UTC().Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return 0, fmt.Errorf("vault: delete audit_log: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("vault: rows affected: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("vault: commit prune tx: %w", err)
	}
	committed = true
	return n, nil
}
