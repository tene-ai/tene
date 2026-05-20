package vault

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	_ "modernc.org/sqlite"
)

// openV1Vault creates a fresh vault.db at the v1 schema state by running
// initSchema (which deposits v1-shape tables) and then stamping
// schema_version = 1 directly, bypassing runMigrations. This simulates the
// pre-sprint state of an existing user's vault.db before the binary that
// introduces schema v2 is run for the first time.
func openV1Vault(t *testing.T) (*Vault, string) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, ".tene", "vault.db")

	// Pre-create directory the way Vault.New would, so we can open the raw
	// sql.DB without involving the New() helper (which would auto-migrate
	// to v2 -- defeating the purpose of this fixture).
	if err := mkdirAll(filepath.Dir(dbPath)); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open raw db: %v", err)
	}
	v := &Vault{db: db, dbPath: dbPath}
	if err := v.initSchema(); err != nil {
		t.Fatalf("initSchema: %v", err)
	}
	// Manually stamp v1 -- the migration code path would have bumped to 2.
	if err := v.SetMeta(schemaMetaKey, "1"); err != nil {
		t.Fatalf("stamp v1: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return v, dbPath
}

func mkdirAll(p string) error {
	// Indirect through a single-line helper so the test file does not need
	// to import os (the production code already imports it; we just want
	// readability here). The actual call is os.MkdirAll(p, 0700).
	return mkdirAllImpl(p)
}

func TestMigrate_V1_to_V2_AddsPreviewColumn(t *testing.T) {
	v, dbPath := openV1Vault(t)

	// Sanity: v1 vault has no preview column.
	hasPreview, err := columnExists(v.db, "secrets", "preview")
	if err != nil {
		t.Fatalf("inspect v1: %v", err)
	}
	if hasPreview {
		t.Fatalf("v1 fixture already has preview column; bad fixture")
	}

	// Close and reopen via Vault.New, which is the real upgrade path users
	// hit on first command after binary bump.
	_ = v.db.Close()
	v2, err := New(dbPath)
	if err != nil {
		t.Fatalf("New after v1 fixture: %v", err)
	}
	t.Cleanup(func() { _ = v2.Close() })

	// schema_version is now "2"
	got, err := v2.GetMeta(schemaMetaKey)
	if err != nil {
		t.Fatalf("read schema_version: %v", err)
	}
	if got != strconv.Itoa(currentSchemaVersion) {
		t.Errorf("schema_version = %q, want %q", got, strconv.Itoa(currentSchemaVersion))
	}

	// secrets table now has the preview column
	hasPreview, err = columnExists(v2.db, "secrets", "preview")
	if err != nil {
		t.Fatalf("inspect v2: %v", err)
	}
	if !hasPreview {
		t.Fatalf("expected preview column after v2 migration")
	}
}

func TestMigrate_V2_Idempotent(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, ".tene", "vault.db")

	// First open: 0 -> 2.
	v1, err := New(dbPath)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	_ = v1.Close()

	// Second open: should be a no-op (version already 2). Must not error,
	// must not duplicate the ALTER (which would be a SQLite error anyway).
	v2, err := New(dbPath)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	t.Cleanup(func() { _ = v2.Close() })

	got, err := v2.GetMeta(schemaMetaKey)
	if err != nil {
		t.Fatalf("read schema_version: %v", err)
	}
	if got != "2" {
		t.Errorf("after re-open schema_version = %q, want %q", got, "2")
	}

	// Third open via repeated runMigrations on the same vault.
	if err := v2.runMigrations(); err != nil {
		t.Fatalf("manual re-run runMigrations: %v", err)
	}
	got, _ = v2.GetMeta(schemaMetaKey)
	if got != "2" {
		t.Errorf("after manual re-run schema_version = %q, want %q", got, "2")
	}
}

func TestMigrate_PreservesExistingSecrets(t *testing.T) {
	v, dbPath := openV1Vault(t)

	// Plant three secrets under the v1 schema (no preview column).
	// We have to use raw SQL here because SetSecretWithPreview's INSERT
	// references the preview column, which doesn't exist yet.
	for _, s := range []struct{ name, val, env string }{
		{"STRIPE_KEY", "ct1", "default"},
		{"OPENAI_KEY", "ct2", "default"},
		{"DB_PASS", "ct3", "prod"},
	} {
		_, err := v.db.Exec(`INSERT INTO secrets (name, encrypted_value, environment) VALUES (?, ?, ?)`,
			s.name, s.val, s.env)
		if err != nil {
			t.Fatalf("seed v1 row %q: %v", s.name, err)
		}
	}
	_ = v.db.Close()

	// Reopen via Vault.New: migrate runs.
	v2, err := New(dbPath)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	t.Cleanup(func() { _ = v2.Close() })

	// All three rows present, ciphertext intact.
	for _, expect := range []struct {
		name, val, env string
	}{
		{"STRIPE_KEY", "ct1", "default"},
		{"OPENAI_KEY", "ct2", "default"},
		{"DB_PASS", "ct3", "prod"},
	} {
		got, err := v2.GetSecret(expect.name, expect.env)
		if err != nil {
			t.Errorf("GetSecret(%q): %v", expect.name, err)
			continue
		}
		if got.EncryptedValue != expect.val {
			t.Errorf("%q: encrypted_value = %q, want %q", expect.name, got.EncryptedValue, expect.val)
		}
	}

	// Previews are all empty (we have NOT run fill-previews yet).
	metas, err := v2.ListSecretMetadata("default")
	if err != nil {
		t.Fatalf("ListSecretMetadata: %v", err)
	}
	if len(metas) != 2 {
		t.Fatalf("default env metas len = %d, want 2", len(metas))
	}
	for _, m := range metas {
		if m.Preview != "" {
			t.Errorf("post-migrate preview = %q, want empty", m.Preview)
		}
	}
}

func TestMigrate_FreshVault_StartsAtV2(t *testing.T) {
	// A vault freshly created via Vault.New (no v1 fixture in front of it)
	// must arrive at schema v2 directly.
	dir := t.TempDir()
	dbPath := filepath.Join(dir, ".tene", "vault.db")
	v, err := New(dbPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })

	got, err := v.GetMeta(schemaMetaKey)
	if err != nil {
		t.Fatalf("read schema_version: %v", err)
	}
	if got != "2" {
		t.Errorf("fresh vault schema_version = %q, want %q", got, "2")
	}

	has, err := columnExists(v.db, "secrets", "preview")
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if !has {
		t.Fatalf("fresh vault is missing preview column")
	}
}

func TestMigrate_ConcurrentOpens(t *testing.T) {
	// Two goroutines (simulating two tene processes) open the same vault.db
	// concurrently. Neither should fail; the late arriver should observe
	// the schema already at v2 and no-op.
	dir := t.TempDir()
	dbPath := filepath.Join(dir, ".tene", "vault.db")

	// Pre-create a v1 vault so the migration has actual work to do.
	v1, _ := openV1Vault(t) // ignore dbPath here; we use our own
	_ = v1.db.Close()
	// openV1Vault used its own tempdir; rebuild a v1 fixture at our dbPath.
	if err := mkdirAll(filepath.Dir(dbPath)); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("seed open: %v", err)
	}
	seed := &Vault{db: db, dbPath: dbPath}
	if err := seed.initSchema(); err != nil {
		t.Fatalf("seed initSchema: %v", err)
	}
	_ = seed.SetMeta(schemaMetaKey, "1")
	_ = db.Close()

	const N = 4
	var wg sync.WaitGroup
	errs := make([]error, N)
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			vv, err := New(dbPath)
			if err != nil {
				errs[i] = err
				return
			}
			defer func() { _ = vv.Close() }()
		}(i)
	}
	wg.Wait()
	for i, e := range errs {
		if e != nil {
			t.Errorf("concurrent open %d: %v", i, e)
		}
	}

	// Final state: schema_version = 2, preview column exists.
	v, err := New(dbPath)
	if err != nil {
		t.Fatalf("final open: %v", err)
	}
	defer func() { _ = v.Close() }()
	ver, _ := v.GetMeta(schemaMetaKey)
	if ver != "2" {
		t.Errorf("final schema_version = %q, want %q", ver, "2")
	}
}

func TestApplyV2_IdempotentColumnDetection(t *testing.T) {
	// Drive applyV2 twice in a row on a v1 fixture. The PRAGMA-based column
	// detection must skip the second ALTER without error.
	v, _ := openV1Vault(t)
	if err := v.applyV2(); err != nil {
		t.Fatalf("first applyV2: %v", err)
	}
	if err := v.applyV2(); err != nil {
		t.Fatalf("second applyV2 (should be no-op): %v", err)
	}
	has, err := columnExists(v.db, "secrets", "preview")
	if err != nil {
		t.Fatalf("inspect: %v", err)
	}
	if !has {
		t.Fatalf("preview column should exist after applyV2")
	}
}

// columnExists is a test helper that returns true when the named table has
// a column with the given name. It is intentionally separate from
// Vault.secretsHasPreviewColumn (which takes *sql.Tx) so we can call it
// outside transactions in test assertions.
func columnExists(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(`PRAGMA table_info("` + table + `")`)
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}
