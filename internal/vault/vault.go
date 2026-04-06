package vault

import (
	"database/sql"
	"fmt"
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
		db.Close()
		return nil, fmt.Errorf("vault: failed to set WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("vault: failed to enable foreign keys: %w", err)
	}

	v := &Vault{db: db, dbPath: dbPath}

	if err := v.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("vault: migration failed: %w", err)
	}

	// Set file permissions to 0600
	if err := os.Chmod(dbPath, 0600); err != nil {
		// Non-fatal on some systems
		_ = err
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
			return err
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
		return 0, err
	}
	return strconv.Atoi(val)
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

	s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
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
	defer rows.Close()

	var secrets []Secret
	for rows.Next() {
		var s Secret
		var createdAt, updatedAt string
		if err := rows.Scan(&s.ID, &s.Name, &s.EncryptedValue, &s.Environment, &s.Version, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("vault: failed to scan secret: %w", err)
		}
		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		secrets = append(secrets, s)
	}
	return secrets, rows.Err()
}

// DeleteSecret deletes a secret.
func (v *Vault) DeleteSecret(name, env string) error {
	result, err := v.db.Exec("DELETE FROM secrets WHERE name = ? AND environment = ?", name, env)
	if err != nil {
		return fmt.Errorf("vault: failed to delete secret %q: %w", name, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
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
	return count > 0, err
}

// CountSecrets returns the number of secrets in the given environment.
func (v *Vault) CountSecrets(env string) (int, error) {
	var count int
	err := v.db.QueryRow("SELECT COUNT(*) FROM secrets WHERE environment = ?", env).Scan(&count)
	return count, err
}

// GetAllSecrets returns all secrets as a name->encryptedValue map.
func (v *Vault) GetAllSecrets(env string) (map[string]string, error) {
	rows, err := v.db.Query("SELECT name, encrypted_value FROM secrets WHERE environment = ?", env)
	if err != nil {
		return nil, fmt.Errorf("vault: failed to get all secrets: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var name, val string
		if err := rows.Scan(&name, &val); err != nil {
			return nil, err
		}
		result[name] = val
	}
	return result, rows.Err()
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
	defer stmt.Close()

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
	defer rows.Close()

	var envs []Environment
	for rows.Next() {
		var e Environment
		var isActive int
		var createdAt string
		if err := rows.Scan(&e.Name, &isActive, &createdAt); err != nil {
			return nil, err
		}
		e.IsActive = isActive == 1
		e.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		envs = append(envs, e)
	}
	return envs, rows.Err()
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
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Deactivate all
	if _, err := tx.Exec("UPDATE environments SET is_active = 0"); err != nil {
		return err
	}

	// Insert or update the target environment
	query := `INSERT INTO environments (name, is_active, created_at)
		VALUES (?, 1, datetime('now'))
		ON CONFLICT(name) DO UPDATE SET is_active = 1`
	if _, err := tx.Exec(query, name); err != nil {
		return err
	}

	return tx.Commit()
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
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	// Count secrets that will be deleted
	var count int
	if err := tx.QueryRow("SELECT COUNT(*) FROM secrets WHERE environment = ?", name).Scan(&count); err != nil {
		return 0, err
	}

	// Delete secrets
	if _, err := tx.Exec("DELETE FROM secrets WHERE environment = ?", name); err != nil {
		return 0, err
	}

	// Delete environment
	result, err := tx.Exec("DELETE FROM environments WHERE name = ?", name)
	if err != nil {
		return 0, err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return 0, fmt.Errorf("%w: %s", ErrEnvironmentNotFound, name)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
}

// EnvironmentExists checks if an environment exists.
func (v *Vault) EnvironmentExists(name string) (bool, error) {
	var count int
	err := v.db.QueryRow("SELECT COUNT(*) FROM environments WHERE name = ?", name).Scan(&count)
	return count > 0, err
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
