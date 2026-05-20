package cli

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/agent-kay-it/tene/internal/vault"
	"github.com/agent-kay-it/tene/internal/vaultcfg"
)

// openVaultReadOnly returns a low-level *sql.DB on the test vault.db so we
// can verify what landed in storage independently of the CLI layer. We
// open the same file the CLI just wrote, not a copy, so the assertions
// reflect the actual on-disk state.
func openVaultReadOnly(t *testing.T, dir string) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(dir, ".tene", "vault.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open vault for verification: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func readSchemaVersion(t *testing.T, db *sql.DB) string {
	t.Helper()
	var v string
	if err := db.QueryRow(`SELECT value FROM vault_meta WHERE key='schema_version'`).Scan(&v); err != nil {
		t.Fatalf("read schema_version: %v", err)
	}
	return v
}

func hasColumn(t *testing.T, db *sql.DB, table, column string) bool {
	t.Helper()
	rows, err := db.Query(`PRAGMA table_info("` + table + `")`)
	if err != nil {
		t.Fatalf("PRAGMA table_info: %v", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if name == column {
			return true
		}
	}
	return false
}

func readPreviewColumn(t *testing.T, db *sql.DB, name, env string) string {
	t.Helper()
	var p string
	err := db.QueryRow(`SELECT preview FROM secrets WHERE name = ? AND environment = ?`, name, env).Scan(&p)
	if err != nil {
		t.Fatalf("read preview for %q: %v", name, err)
	}
	return p
}

func readEncryptedValueColumn(t *testing.T, db *sql.DB, name, env string) string {
	t.Helper()
	var v string
	err := db.QueryRow(`SELECT encrypted_value FROM secrets WHERE name = ? AND environment = ?`, name, env).Scan(&v)
	if err != nil {
		t.Fatalf("read encrypted_value for %q: %v", name, err)
	}
	return v
}

// TestDataflow_Case1_FreshInit_VaultIsAtSchemaV2 verifies that a fresh
// `tene init` lands at schema_version=2 with the preview column present
// from the very first command -- no implicit upgrade required.
func TestDataflow_Case1_FreshInit_VaultIsAtSchemaV2(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	db := openVaultReadOnly(t, env.Dir)

	ver := readSchemaVersion(t, db)
	if ver != "2" {
		t.Errorf("schema_version = %q, want %q", ver, "2")
	}
	if !hasColumn(t, db, "secrets", "preview") {
		t.Errorf("secrets table is missing the preview column")
	}
}

// TestDataflow_Case2_SetSecret_WritesCiphertextAndDefaultPreview verifies
// that `tene set FOO bar456789` with default preview settings stores a
// preview of "…6789" (front=0, back=4) and a ciphertext that is NOT the
// plaintext.
func TestDataflow_Case2_SetSecret_WritesCiphertextAndDefaultPreview(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	const plaintext = "bar-secret-value-9876"
	if _, _, err := env.run("set", "FOO", plaintext, "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}

	db := openVaultReadOnly(t, env.Dir)

	preview := readPreviewColumn(t, db, "FOO", "default")
	want := "…9876"
	if preview != want {
		t.Errorf("preview = %q, want %q", preview, want)
	}

	ct := readEncryptedValueColumn(t, db, "FOO", "default")
	if ct == plaintext {
		t.Errorf("encrypted_value contains plaintext")
	}
	if strings.Contains(ct, "bar-secret") {
		t.Errorf("encrypted_value leaks plaintext substring: %q", ct)
	}

	// Round-trip via tene get: confirms the encryption is decryptable.
	stdout, _, err := env.run("get", "FOO")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if strings.TrimSpace(stdout) != plaintext {
		t.Errorf("get = %q, want %q", strings.TrimSpace(stdout), plaintext)
	}
}

// TestDataflow_Case3_SetSecret_WithPreviewDisabled verifies that after
// `tene config preview.enabled=false`, subsequent `tene set` writes land
// with an empty preview column, not a populated one.
func TestDataflow_Case3_SetSecret_WithPreviewDisabled(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	if _, _, err := env.run("config", "preview.enabled=false"); err != nil {
		t.Fatalf("disable preview: %v", err)
	}

	if _, _, err := env.run("set", "BAZ", "qux-secret-1234", "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}

	db := openVaultReadOnly(t, env.Dir)
	preview := readPreviewColumn(t, db, "BAZ", "default")
	if preview != "" {
		t.Errorf("preview = %q, want empty (preview disabled)", preview)
	}
}

// TestDataflow_Case4_MigrateFillPreviews_PopulatesAfterV1Upgrade simulates
// the existing-user upgrade path: a v1 vault.db is created out-of-band,
// then opened by `tene` (which auto-migrates to v2), and `tene migrate
// fill-previews` populates the previews that the migration left empty.
func TestDataflow_Case4_MigrateFillPreviews_PopulatesAfterV1Upgrade(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	// First, write three secrets at v2 (the test harness already
	// initialized at v2 since it goes through `tene init`).
	for _, kv := range []struct{ k, v string }{
		{"A_KEY", "value-A-12345"},
		{"B_KEY", "value-B-67890"},
		{"C_KEY", "value-C-ABCDE"},
	} {
		if _, _, err := env.run("set", kv.k, kv.v, "--overwrite"); err != nil {
			t.Fatalf("seed %s: %v", kv.k, err)
		}
	}

	// Now simulate "previews never derived" by clearing them via direct SQL.
	// This is the same disk state that a v1->v2 auto-migration produces.
	db := openVaultReadOnly(t, env.Dir)
	if _, err := db.Exec(`UPDATE secrets SET preview = '' WHERE environment = 'default'`); err != nil {
		t.Fatalf("clear previews: %v", err)
	}
	_ = db.Close()

	// Sanity check: all three previews are empty.
	db2 := openVaultReadOnly(t, env.Dir)
	for _, name := range []string{"A_KEY", "B_KEY", "C_KEY"} {
		if got := readPreviewColumn(t, db2, name, "default"); got != "" {
			t.Fatalf("pre-fill preview for %s = %q, want empty", name, got)
		}
	}
	_ = db2.Close()

	// Run fill-previews.
	if _, _, err := env.run("migrate", "fill-previews"); err != nil {
		t.Fatalf("migrate fill-previews: %v", err)
	}

	// All three should now have populated previews (default front=0, back=4).
	db3 := openVaultReadOnly(t, env.Dir)
	want := map[string]string{
		"A_KEY": "…2345",
		"B_KEY": "…7890",
		"C_KEY": "…BCDE",
	}
	for name, expected := range want {
		got := readPreviewColumn(t, db3, name, "default")
		if got != expected {
			t.Errorf("%s preview = %q, want %q", name, got, expected)
		}
	}

	// Idempotence: running fill-previews again should be a no-op.
	if _, _, err := env.run("migrate", "fill-previews"); err != nil {
		t.Fatalf("second migrate fill-previews: %v", err)
	}
	db4 := openVaultReadOnly(t, env.Dir)
	for name, expected := range want {
		got := readPreviewColumn(t, db4, name, "default")
		if got != expected {
			t.Errorf("idempotent: %s preview = %q, want %q", name, got, expected)
		}
	}
}

// TestDataflow_Case5_ConfigPreviewFront_PersistsInVaultMeta verifies that
// `tene config preview.front=4 --force` writes the new value to the
// vault_meta table under the canonical "config.preview.front" key.
func TestDataflow_Case5_ConfigPreviewFront_PersistsInVaultMeta(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	if _, _, err := env.run("config", "preview.front=4", "--force"); err != nil {
		t.Fatalf("config preview.front=4: %v", err)
	}

	db := openVaultReadOnly(t, env.Dir)
	var val string
	err := db.QueryRow(`SELECT value FROM vault_meta WHERE key = ?`, vaultcfg.KeyPreviewFront).Scan(&val)
	if err != nil {
		t.Fatalf("read vault_meta[%q]: %v", vaultcfg.KeyPreviewFront, err)
	}
	if val != "4" {
		t.Errorf("stored preview.front = %q, want %q", val, "4")
	}
}

// TestDataflow_Case6_OptInPrefix_ExposesFrontChars verifies the end-to-end
// flow when the user explicitly opts into preview.front=4: a subsequent
// `tene set` writes a preview that includes the requested prefix.
func TestDataflow_Case6_OptInPrefix_ExposesFrontChars(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	if _, _, err := env.run("config", "preview.front=4", "--force"); err != nil {
		t.Fatalf("config preview.front=4: %v", err)
	}

	const plaintext = "defghijklmn5678" // 15 runes
	if _, _, err := env.run("set", "ABC", plaintext, "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}

	db := openVaultReadOnly(t, env.Dir)
	got := readPreviewColumn(t, db, "ABC", "default")
	const want = "defg…5678"
	if got != want {
		t.Errorf("preview = %q, want %q", got, want)
	}
}

// TestDataflow_Case7_ToggleEnabled_ApplyOnNextSet verifies that toggling
// preview.enabled affects only secrets written AFTER the toggle, not
// previously stored ones. This is the documented behavior in design.md
// §6A.1 ("existing entry는 그대로").
func TestDataflow_Case7_ToggleEnabled_ApplyOnNextSet(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	// Write FIRST while previews are enabled (default).
	if _, _, err := env.run("set", "ENABLED_KEY", "value-ENABLED-zzzz", "--overwrite"); err != nil {
		t.Fatalf("first set: %v", err)
	}

	// Disable previews, then write SECOND.
	if _, _, err := env.run("config", "preview.enabled=false"); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if _, _, err := env.run("set", "DISABLED_KEY", "value-DISABLED-yyyy", "--overwrite"); err != nil {
		t.Fatalf("second set: %v", err)
	}

	db := openVaultReadOnly(t, env.Dir)

	first := readPreviewColumn(t, db, "ENABLED_KEY", "default")
	if first != "…zzzz" {
		t.Errorf("ENABLED_KEY preview = %q, want %q (set BEFORE disable)", first, "…zzzz")
	}
	second := readPreviewColumn(t, db, "DISABLED_KEY", "default")
	if second != "" {
		t.Errorf("DISABLED_KEY preview = %q, want empty (set AFTER disable)", second)
	}
}

// TestDataflow_Case8_ListSecretMetadata_NoEncryptedValueLeak is the final
// regression guard for invariant I-1: even after a full lifecycle of init
// -> set -> config -> set, calling Vault.ListSecretMetadata directly must
// never return ciphertext.
func TestDataflow_Case8_ListSecretMetadata_NoEncryptedValueLeak(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	const sentinel = "PLAINTEXT-DO-NOT-LEAK-9999"
	if _, _, err := env.run("set", "SENTINEL", sentinel, "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}

	// Open the vault directly (not via CLI) and call the metadata API.
	v, err := vault.New(filepath.Join(env.Dir, ".tene", "vault.db"))
	if err != nil {
		t.Fatalf("vault.New: %v", err)
	}
	defer func() { _ = v.Close() }()

	metas, err := v.ListSecretMetadata("default")
	if err != nil {
		t.Fatalf("ListSecretMetadata: %v", err)
	}

	for _, m := range metas {
		for _, field := range []string{m.Name, m.Preview} {
			if strings.Contains(field, sentinel) {
				t.Fatalf("metadata field %q contains plaintext sentinel; I-1 invariant violated", field)
			}
		}
	}

	// And the underlying encrypted_value column is still ciphertext.
	db := openVaultReadOnly(t, env.Dir)
	ct := readEncryptedValueColumn(t, db, "SENTINEL", "default")
	if strings.Contains(ct, sentinel) {
		t.Errorf("encrypted_value leaks plaintext: %q", ct)
	}
}

// TestDataflow_ConfigJSON_ListAllKeys verifies the `tene config --json`
// (zero-arg) output shape.
func TestDataflow_ConfigJSON_ListAllKeys(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	stdout, _, err := env.runJSON("config")
	if err != nil {
		t.Fatalf("config --json: %v", err)
	}
	var parsed struct {
		OK     bool              `json:"ok"`
		Config map[string]string `json:"config"`
	}
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("unmarshal: %v\nstdout=%s", err, stdout)
	}
	if !parsed.OK {
		t.Errorf("ok = false")
	}
	// All four known keys must be present, with defaults.
	want := map[string]string{
		vaultcfg.KeyPreviewEnabled: "true",
		vaultcfg.KeyPreviewFront:   "0",
		vaultcfg.KeyPreviewBack:    "4",
		vaultcfg.KeyAuditWarnAtMB:  "50",
	}
	for k, v := range want {
		if got := parsed.Config[k]; got != v {
			t.Errorf("config[%q] = %q, want %q", k, got, v)
		}
	}
}

// TestDataflow_ConfigSet_RejectsCombinedCapViolation is the integration-level
// equivalent of the vaultcfg unit test by the same name, exercising the
// full CLI -> vaultcfg -> vault round-trip including error surfacing.
func TestDataflow_ConfigSet_RejectsCombinedCapViolation(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	if _, _, err := env.run("config", "preview.back=5"); err != nil {
		t.Fatalf("config preview.back=5: %v", err)
	}
	// front=5 + back=5 = 10, exceeds the hard cap of 8.
	_, _, err := env.run("config", "preview.front=5", "--force")
	if err == nil {
		t.Fatalf("expected error for combined cap violation, got nil")
	}
	if !strings.Contains(err.Error(), "must not exceed") {
		t.Errorf("error = %q, want substring 'must not exceed'", err.Error())
	}
}

// TestDataflow_ImportDotEnv_WritesPreviewsForEachRow exercises the batched
// import path: every imported secret should end up with both an encrypted
// value and a derived preview.
func TestDataflow_ImportDotEnv_WritesPreviewsForEachRow(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	// Drop a .env file in the project dir.
	envFile := filepath.Join(env.Dir, "input.env")
	contents := strings.Join([]string{
		"AA_KEY=aa-secret-1234",
		"BB_KEY=bb-secret-5678",
		"CC_KEY=cc-secret-90ab",
		"",
	}, "\n")
	if err := writeFile(envFile, contents); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	if _, _, err := env.run("import", envFile); err != nil {
		t.Fatalf("import: %v", err)
	}

	db := openVaultReadOnly(t, env.Dir)
	want := map[string]string{
		"AA_KEY": "…1234",
		"BB_KEY": "…5678",
		"CC_KEY": "…90ab",
	}
	for name, expected := range want {
		got := readPreviewColumn(t, db, name, "default")
		if got != expected {
			t.Errorf("%s preview = %q, want %q", name, got, expected)
		}
	}
}

// TestDataflow_PreviewFrontConfirmText_ByteIdentical guards against
// accidental drift between the constant in this package, the wording in
// plan.md F1 step 7, and the wording in prd.md §5 Failure Mode G. The
// text is security-sensitive: it is the user's last chance to refuse an
// opt-in that exposes API key prefixes.
//
// We do not assert byte-equality against the doc files here (that would
// be brittle to whitespace), but we do pin the exact wording so future
// refactors that touch the constant must consciously update this test.
func TestDataflow_PreviewFrontConfirmText_ByteIdentical(t *testing.T) {
	const want = "WARNING: setting preview.front > 0 will expose API key prefixes (sk-, ghp_, AKIA...) in vault.db.\n" +
		"This makes service identification possible if vault.db leaks. Continue? [y/N]"
	if previewFrontConfirmText != want {
		t.Fatalf("previewFrontConfirmText drift detected.\ngot:  %q\nwant: %q",
			previewFrontConfirmText, want)
	}
}

// --- helpers ---

func writeFile(path, content string) error {
	return writeFileImpl(path, content)
}
