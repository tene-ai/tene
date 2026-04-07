package vault

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

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

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("vault: failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrency
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

// migrate checks the schema version and runs migrations if needed.
func (v *Vault) migrate() error {
	// Try to get schema version - if it fails, initialize schema
	version, err := v.getSchemaVersion()
	if err != nil {
		// First run: initialize schema
		if err := v.initSchema(); err != nil {
			return fmt.Errorf("vault: init schema: %w", err)
		}
		return v.SetMeta("schema_version", strconv.Itoa(currentSchemaVersion))
	}

	// Future migrations would go here
	_ = version
	return nil
}

func (v *Vault) getSchemaVersion() (int, error) {
	val, err := v.GetMeta("schema_version")
	if err != nil {
		return 0, fmt.Errorf("vault: get schema version: %w", err)
	}
	version, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("vault: parse schema version: %w", err)
	}
	return version, nil
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

// SetSecret stores a secret (UPSERT: by name+environment).
func (v *Vault) SetSecret(name, encryptedValue, env string) error {
	query := `
		INSERT INTO secrets (name, encrypted_value, environment, version, created_at, updated_at)
		VALUES (?, ?, ?, 1, datetime('now'), datetime('now'))
		ON CONFLICT(name, environment) DO UPDATE SET
			encrypted_value = excluded.encrypted_value,
			version = secrets.version + 1,
			updated_at = datetime('now')
	`
	_, err := v.db.Exec(query, name, encryptedValue, env)
	if err != nil {
		return fmt.Errorf("vault: failed to set secret %q: %w", name, err)
	}

	return v.AddAuditLog("secret.write", name, "")
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
