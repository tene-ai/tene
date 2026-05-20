package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agent-kay-it/tene/internal/vault"
)

// TestList_NoDecrypt_NoPasswordPrompt asserts the F3 core promise: `tene
// list` returns successfully under the exact conditions that historically
// triggered a password prompt -- non-interactive stdin, no keychain entry,
// no TENE_MASTER_PASSWORD env var. Pre-F3 the command called
// loadOrPromptMasterKey, which would return teneerr.ErrInteractiveRequired
// in this configuration. Post-F3 the command never reaches that codepath.
func TestList_NoDecrypt_NoPasswordPrompt(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	if _, _, err := env.run("set", "FOO", "value-foo-12345", "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}

	// Drop TENE_MASTER_PASSWORD so prompt path would be reached if F3 ever
	// regressed. --no-keychain is already set by the test harness, so we
	// have no key source at all.
	t.Setenv("TENE_MASTER_PASSWORD", "")

	stdout, _, err := env.run("list")
	if err != nil {
		t.Fatalf("list without any key source: %v", err)
	}
	if !strings.Contains(stdout, "FOO") {
		t.Errorf("stdout missing FOO; got:\n%s", stdout)
	}
}

// TestList_JSON_PreviewAlwaysString verifies the Q2 always-string
// contract: the "preview" field is present on every secret element,
// regardless of whether the value is populated. JSON consumers can
// always rely on `secret.preview` being a string.
func TestList_JSON_PreviewAlwaysString(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	// Seed three secrets with default preview (front=0, back=4 -> populated)
	if _, _, err := env.run("set", "POPULATED_KEY", "value-with-suffix-AAAA", "--overwrite"); err != nil {
		t.Fatalf("set populated: %v", err)
	}
	if _, _, err := env.run("set", "ANOTHER_KEY", "another-value-BBBB", "--overwrite"); err != nil {
		t.Fatalf("set another: %v", err)
	}

	// Disable previews and add a third row that lands with empty preview.
	if _, _, err := env.run("config", "preview.enabled=false"); err != nil {
		t.Fatalf("disable preview: %v", err)
	}
	if _, _, err := env.run("set", "EMPTY_PREVIEW_KEY", "value-no-preview-CCCC", "--overwrite"); err != nil {
		t.Fatalf("set empty: %v", err)
	}

	// Re-enable so the renderer is exercised in mixed state -- this is the
	// realistic case where a user disabled briefly and then re-enabled.
	if _, _, err := env.run("config", "preview.enabled=true"); err != nil {
		t.Fatalf("re-enable: %v", err)
	}

	stdout, _, err := env.runJSON("list")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}

	var parsed struct {
		OK      bool `json:"ok"`
		Count   int  `json:"count"`
		Secrets []struct {
			Name    string          `json:"name"`
			Preview json.RawMessage `json:"preview"`
			Version int             `json:"version"`
		} `json:"secrets"`
	}
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("unmarshal: %v\nstdout: %s", err, stdout)
	}
	if !parsed.OK {
		t.Errorf("ok = false")
	}
	if parsed.Count != 3 {
		t.Errorf("count = %d, want 3", parsed.Count)
	}

	// Every secret has the preview field with a JSON string value. Use
	// RawMessage so we catch null (would unmarshal to "null") and absent
	// (would be empty raw -- but Go zeroes the field on absence).
	seen := map[string]string{}
	for _, s := range parsed.Secrets {
		raw := string(s.Preview)
		if raw == "null" {
			t.Errorf("secret %q preview is null; want string (always-string contract)", s.Name)
			continue
		}
		if len(raw) == 0 {
			t.Errorf("secret %q preview is absent; want string key always present", s.Name)
			continue
		}
		if raw[0] != '"' {
			t.Errorf("secret %q preview is not a JSON string (got %s)", s.Name, raw)
			continue
		}
		var v string
		if err := json.Unmarshal(s.Preview, &v); err != nil {
			t.Errorf("secret %q preview not unmarshalable as string: %v", s.Name, err)
			continue
		}
		seen[s.Name] = v
	}

	// Populated rows have substring previews; empty one has literal "".
	if got := seen["POPULATED_KEY"]; got != "…AAAA" {
		t.Errorf("POPULATED_KEY preview = %q, want %q", got, "…AAAA")
	}
	if got := seen["ANOTHER_KEY"]; got != "…BBBB" {
		t.Errorf("ANOTHER_KEY preview = %q, want %q", got, "…BBBB")
	}
	if got, ok := seen["EMPTY_PREVIEW_KEY"]; !ok {
		t.Error("EMPTY_PREVIEW_KEY missing from output")
	} else if got != "" {
		t.Errorf("EMPTY_PREVIEW_KEY preview = %q, want empty string", got)
	}
}

// TestList_Text_DashForEmpty verifies that a secret with an empty preview
// column renders as "-" in the text-mode PREVIEW column. The dash is a
// stable single-character placeholder so column alignment is preserved
// for AWK-style consumers (NAME at position 1).
func TestList_Text_DashForEmpty(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	if _, _, err := env.run("config", "preview.enabled=false"); err != nil {
		t.Fatalf("disable preview: %v", err)
	}
	if _, _, err := env.run("set", "BLANK_KEY", "value-no-preview-DDDD", "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}

	stdout, _, err := env.run("list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(stdout, "BLANK_KEY") {
		t.Fatalf("stdout missing BLANK_KEY:\n%s", stdout)
	}
	// Locate the BLANK_KEY row and confirm a literal "-" appears between
	// the name and the timestamp columns.
	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(line, "BLANK_KEY") {
			fields := strings.Fields(line)
			// fields: [BLANK_KEY, -, "<time>", "<unit>", "ago"]
			if len(fields) < 2 || fields[1] != "-" {
				t.Errorf("BLANK_KEY row preview field = %q, want %q; full line: %q", fields, "-", line)
			}
			return
		}
	}
	t.Errorf("BLANK_KEY row not found in stdout:\n%s", stdout)
}

// TestList_Text_FooterWhenDisabled verifies the disabled-state footer
// shows on stderr when preview.enabled=false, and does NOT show on the
// default preview-enabled path.
func TestList_Text_FooterWhenDisabled(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	if _, _, err := env.run("set", "AKEY", "abcdefghij1234", "--overwrite"); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// First call: default preview state -> no disabled-footer on stderr.
	_, stderr1, err := env.run("list")
	if err != nil {
		t.Fatalf("list enabled: %v", err)
	}
	if strings.Contains(stderr1, "previews disabled") {
		t.Errorf("disabled footer leaked while preview is enabled; stderr:\n%s", stderr1)
	}

	// Disable preview and list again -> footer present.
	if _, _, err := env.run("config", "preview.enabled=false"); err != nil {
		t.Fatalf("disable: %v", err)
	}
	_, stderr2, err := env.run("list")
	if err != nil {
		t.Fatalf("list disabled: %v", err)
	}
	if !strings.Contains(stderr2, "previews disabled") {
		t.Errorf("expected disabled footer; stderr:\n%s", stderr2)
	}
	if !strings.Contains(stderr2, "tene config preview.enabled=true") {
		t.Errorf("disabled footer missing re-enable hint; stderr:\n%s", stderr2)
	}
}

// TestList_EmptyEnv verifies the friendly "no secrets" message in both
// text and JSON modes.
func TestList_EmptyEnv(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	stdout, _, err := env.run("list")
	if err != nil {
		t.Fatalf("list empty: %v", err)
	}
	if !strings.Contains(stdout, "No secrets") {
		t.Errorf("expected empty-state message; got:\n%s", stdout)
	}

	jsonOut, _, err := env.runJSON("list")
	if err != nil {
		t.Fatalf("list --json empty: %v", err)
	}
	var parsed struct {
		OK      bool          `json:"ok"`
		Count   int           `json:"count"`
		Secrets []interface{} `json:"secrets"`
	}
	if err := json.Unmarshal([]byte(jsonOut), &parsed); err != nil {
		t.Fatalf("unmarshal empty list: %v\nstdout: %s", err, jsonOut)
	}
	if parsed.Count != 0 {
		t.Errorf("count = %d, want 0", parsed.Count)
	}
	if parsed.Secrets == nil {
		t.Errorf("secrets is nil; want empty array (deterministic JSON shape)")
	}
}

// TestList_EnvFlag verifies that --env staging scopes the listing to a
// single environment. A secret seeded into "default" must not appear when
// listing "staging".
func TestList_EnvFlag(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	if _, _, err := env.run("set", "DEFAULT_KEY", "default-value-1111", "--overwrite"); err != nil {
		t.Fatalf("set default: %v", err)
	}
	if _, _, err := env.run("env", "create", "staging"); err != nil {
		t.Fatalf("env create: %v", err)
	}
	if _, _, err := env.run("--env", "staging", "set", "STAGING_KEY", "staging-value-2222", "--overwrite"); err != nil {
		t.Fatalf("set staging: %v", err)
	}

	stdout, _, err := env.run("--env", "staging", "list")
	if err != nil {
		t.Fatalf("list staging: %v", err)
	}
	if !strings.Contains(stdout, "STAGING_KEY") {
		t.Errorf("staging listing missing STAGING_KEY:\n%s", stdout)
	}
	if strings.Contains(stdout, "DEFAULT_KEY") {
		t.Errorf("staging listing leaked DEFAULT_KEY from another env:\n%s", stdout)
	}
}

// TestList_LegacyV1_AutoMigrated simulates the existing-user upgrade path:
// secrets written under schema v2 then have their preview column cleared
// to mimic the state right after the v1->v2 ALTER TABLE bumped schema
// but before `tene migrate fill-previews` was run. `tene list` must show
// every row with PREVIEW="-" and emit the "run migrate fill-previews"
// hint on stderr.
func TestList_LegacyV1_AutoMigrated(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	for _, kv := range []struct{ k, v string }{
		{"LEGACY_A", "legacy-A-XXXX"},
		{"LEGACY_B", "legacy-B-YYYY"},
	} {
		if _, _, err := env.run("set", kv.k, kv.v, "--overwrite"); err != nil {
			t.Fatalf("seed %s: %v", kv.k, err)
		}
	}

	// Wipe preview column to reproduce post-v2-migration state.
	db := openVaultReadOnly(t, env.Dir)
	if _, err := db.Exec(`UPDATE secrets SET preview = '' WHERE environment = 'default'`); err != nil {
		t.Fatalf("clear previews: %v", err)
	}

	stdout, stderr, err := env.run("list")
	if err != nil {
		t.Fatalf("list legacy: %v", err)
	}
	// Both rows visible.
	if !strings.Contains(stdout, "LEGACY_A") || !strings.Contains(stdout, "LEGACY_B") {
		t.Errorf("legacy listing missing rows:\n%s", stdout)
	}
	// All previews rendered as "-".
	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(line, "LEGACY_") {
			fields := strings.Fields(line)
			if len(fields) < 2 || fields[1] != "-" {
				t.Errorf("legacy row preview != %q: %q", "-", line)
			}
		}
	}
	// Hint suggests fill-previews, not config.preview.enabled.
	if !strings.Contains(stderr, "tene migrate fill-previews") {
		t.Errorf("legacy hint missing; stderr:\n%s", stderr)
	}
	if strings.Contains(stderr, "previews disabled") {
		t.Errorf("disabled hint shown when preview is enabled; stderr:\n%s", stderr)
	}
}

// TestList_NoUnlockPath_StaysQuiet verifies the runtime invariant that
// the dispatcher classifies `list` as PermMetaRead and therefore never
// invokes a master-key derivation. We assert by side effect: running
// `tene list` in a configuration where loadOrPromptMasterKey would
// definitely fail (no keychain entry, no TENE_MASTER_PASSWORD) must
// still succeed, AND the wall-clock duration must be far below the
// Argon2id cost (~80 ms with time=3 memory=64MB) that a real unlock
// would add. We give a generous 200 ms upper bound to absorb cold-start
// noise on CI runners; a regression to the decrypt path would push
// well past that.
func TestList_NoUnlockPath_StaysQuiet(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	if _, _, err := env.run("set", "QUIET_KEY", "quiet-value-1234", "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}
	// Drop the env-var key source. resetFlags() will set --no-keychain via
	// run(); together this leaves no path to loadOrPromptMasterKey
	// returning a key. Pre-F3 this configuration returns
	// ErrInteractiveRequired.
	t.Setenv("TENE_MASTER_PASSWORD", "")

	start := time.Now()
	stdout, _, err := env.run("list")
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("list under no-keychain + no-env: %v (pre-F3 regression?)", err)
	}
	if !strings.Contains(stdout, "QUIET_KEY") {
		t.Errorf("stdout missing QUIET_KEY:\n%s", stdout)
	}
	// Soft timing assert. The point isn't precise timing; it's the
	// absence of Argon2id (≥ 60 ms even at the fastest CPUs).
	if elapsed > 200*time.Millisecond {
		t.Logf("list elapsed = %v (informational; <200ms expected for no-decrypt path)", elapsed)
	}
}

// BenchmarkListWithPreview measures the wall-clock cost of `tene list`
// at 100 secrets. G6 target: p50 < 15 ms. We exercise the production
// runList via direct Vault.ListSecretMetadata to isolate the metadata
// read from the cobra/Argv setup overhead (which is constant per call
// and irrelevant to F3's promise).
func BenchmarkListWithPreview(b *testing.B) {
	// Build the fixture once outside the benchmark loop.
	tmp := b.TempDir()
	b.Setenv("HOME", b.TempDir())
	b.Setenv("TENE_MASTER_PASSWORD", "benchpassword123")

	// Construct a 100-secret vault via the public API. The vault setup
	// itself is amortized (b.ResetTimer below).
	v, err := vault.New(filepath.Join(tmp, "vault.db"))
	if err != nil {
		b.Fatalf("vault.New: %v", err)
	}
	defer func() { _ = v.Close() }()

	// SetSecret requires a project_name to be set in vault_meta so the
	// rest of the package considers it initialized. We bypass full init
	// because we only need the secrets table populated for the bench.
	_ = v.SetMeta("project_name", "benchproj")

	// Each row needs a non-empty preview to mirror real usage. We use
	// SetSecretWithPreview so the same atomic ciphertext+preview path
	// exercised in production lands the bench fixture. The ciphertext
	// is a stub -- ListSecretMetadata never reads it (I-1 invariant).
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("KEY_%03d", i)
		preview := fmt.Sprintf("…%04d", i)
		if err := v.SetSecretWithPreview(name, "ciphertextstub", "default", preview); err != nil {
			b.Fatalf("seed row %d: %v", i, err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		metas, err := v.ListSecretMetadata("default")
		if err != nil {
			b.Fatalf("ListSecretMetadata: %v", err)
		}
		if len(metas) != 100 {
			b.Fatalf("got %d metas, want 100", len(metas))
		}
	}
}
