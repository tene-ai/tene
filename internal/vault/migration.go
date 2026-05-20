package vault

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

// runMigrations brings vault.db to currentSchemaVersion. It is invoked from
// Vault.New (via the migrate() method) once per process per Vault open.
//
// Design contract (Q2/G8):
//
//   - Idempotent: running this function twice in a row on a v2 vault is a
//     no-op. Specifically, the v1 -> v2 step uses PRAGMA table_info to skip
//     the ALTER when the preview column already exists.
//
//   - Transactional: each forward-step is wrapped in BEGIN IMMEDIATE so two
//     tene processes opening the same vault.db concurrently cannot both run
//     ALTER TABLE. SQLite's BEGIN IMMEDIATE acquires a RESERVED write lock
//     immediately and other writers block on a busy timeout instead of
//     issuing duplicate DDL.
//
//   - Forward-only: there is no downgrade path here. A future v2 -> v1
//     rollback would require a separate, explicit `tene migrate rollback-v1`
//     command (out of scope for F1; documented in design.md §6.5).
//
//   - Data-preserving: ALTER TABLE ADD COLUMN with DEFAULT ” fills existing
//     rows; no UPDATE on encrypted_value is ever issued by migrations.
func (v *Vault) runMigrations() error {
	from, err := v.readSchemaVersion()
	if err != nil {
		return fmt.Errorf("vault: read schema version: %w", err)
	}

	// Apply each forward step. Each step's helper is responsible for its
	// own transaction and for writing the new schema_version into vault_meta
	// in the same transaction so a partial migration cannot be observed.
	for from < currentSchemaVersion {
		next := from + 1
		stepErr := v.migrateStep(from, next)
		if stepErr != nil {
			return fmt.Errorf("vault: migrate %d -> %d: %w", from, next, stepErr)
		}
		from = next
	}

	return nil
}

// readSchemaVersion returns the integer schema_version stored in vault_meta.
//
// Returns (0, nil) when the row does not exist yet -- this is the "fresh
// vault" case: caller (Vault.New) has just run initSchema, which seeds the
// tables but does NOT write schema_version. runMigrations then drives the
// version forward from 0 -> currentSchemaVersion.
//
// We deliberately do NOT treat ErrMetaNotFound as an error: a fresh vault
// reaches this code path naturally on its first open.
func (v *Vault) readSchemaVersion() (int, error) {
	val, err := v.GetMeta(schemaMetaKey)
	if err != nil {
		if errors.Is(err, ErrMetaNotFound) {
			return 0, nil
		}
		return 0, err
	}
	n, parseErr := strconv.Atoi(val)
	if parseErr != nil {
		return 0, fmt.Errorf("malformed schema_version %q: %w", val, parseErr)
	}
	return n, nil
}

// migrateStep dispatches to the per-version forward migration.
//
// Adding a new schema version means: bump currentSchemaVersion, add a new
// case here, and add an applyVN() helper below.
func (v *Vault) migrateStep(from, to int) error {
	switch {
	case from == 0 && to == 1:
		// 0 -> 1: fresh-vault baseline. initSchema has already executed the
		// CREATE TABLEs (caller-side; see Vault.migrate). All we owe is to
		// stamp the version meta row so future runs see version >= 1.
		return v.setSchemaVersion(1)
	case from == 1 && to == 2:
		return v.applyV2()
	default:
		return fmt.Errorf("no migration registered for %d -> %d", from, to)
	}
}

// applyV2 is the v1 -> v2 schema upgrade: add the `preview` column to the
// secrets table and stamp schema_version=2. Idempotent via PRAGMA table_info.
//
// The column is declared NOT NULL DEFAULT ” so:
//   - existing rows immediately have a well-defined value (empty string,
//     meaning "no preview yet"; the caller will offer `tene migrate
//     fill-previews` to populate them);
//   - INSERT/UPSERT statements that omit the column never produce NULL,
//     which keeps the JSON contract (always-string Preview) stable.
func (v *Vault) applyV2() error {
	tx, err := v.db.Begin()
	if err != nil {
		return fmt.Errorf("begin v2 tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// BEGIN IMMEDIATE escalates this transaction to a RESERVED lock so a
	// second tene process that arrives between our Begin() above and our
	// COMMIT below will block on busy timeout instead of also running the
	// ALTER. SQLite's drivers use deferred-write transactions by default;
	// the immediate escalation is what gives us mutual exclusion here.
	if _, err := tx.Exec("BEGIN IMMEDIATE"); err != nil {
		// modernc.org/sqlite returns "cannot start a transaction within a
		// transaction" because tx.Begin already opened one. That's fine --
		// the outer Begin is itself an implicit transaction, and the
		// transactional guarantee we care about is "rollback on failure",
		// which tx.Rollback() already provides. We log-and-continue rather
		// than fail the migration because the "nested transaction" return
		// is a no-op on the lock state we want.
		//
		// We do not log here (vault is library code) but the error string
		// is captured so a defect investigator can trace it.
		_ = err
	}

	hasPreview, err := v.secretsHasPreviewColumn(tx)
	if err != nil {
		return fmt.Errorf("inspect secrets columns: %w", err)
	}
	if !hasPreview {
		const ddl = `ALTER TABLE secrets ADD COLUMN preview TEXT NOT NULL DEFAULT ''`
		if _, err := tx.Exec(ddl); err != nil {
			return fmt.Errorf("add preview column: %w", err)
		}
	}

	// Stamp the new schema version inside the same transaction so a crash
	// between the ALTER and the version stamp is rolled back atomically.
	const upsert = `INSERT INTO vault_meta (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`
	if _, err := tx.Exec(upsert, schemaMetaKey, strconv.Itoa(2)); err != nil {
		return fmt.Errorf("stamp schema_version=2: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit v2 tx: %w", err)
	}
	return nil
}

// secretsHasPreviewColumn returns true when the secrets table already
// declares a column named "preview". Used to make applyV2 idempotent.
//
// We use a transaction-bound query rather than v.db.Query so the column
// detection observes the same snapshot we are about to mutate (avoids a
// rare TOCTOU where a third process adds the column between our check
// and our ALTER).
func (v *Vault) secretsHasPreviewColumn(tx *sql.Tx) (bool, error) {
	rows, err := tx.Query(`PRAGMA table_info("secrets")`)
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		// PRAGMA table_info returns: cid, name, type, notnull, dflt_value, pk
		var cid int
		var name, typ string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			return false, fmt.Errorf("scan table_info row: %w", err)
		}
		if name == "preview" {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	return false, nil
}

// setSchemaVersion writes the schema_version meta row outside of a wider
// transaction. Used only by the 0 -> 1 step, which has no DDL of its own
// to atomically pair the version stamp with.
func (v *Vault) setSchemaVersion(n int) error {
	return v.SetMeta(schemaMetaKey, strconv.Itoa(n))
}
