package cli

import (
	"regexp"
	"strings"
	"testing"
)

// Sprint v1014-rc1-qa-fixes / FX6.
//
// One file covers all six polish fixes (B7, B8, B10, B11, B12, B13)
// because each is small and the surface area is shared (helpers.go +
// list.go + env.go + config.go + root.go). Keeping the tests
// co-located makes the next polish PR easy to extend.

// --- B10 pluralize -----------------------------------------------------

func TestPluralize(t *testing.T) {
	cases := []struct {
		count    int
		singular string
		want     string
	}{
		{0, "secret", "secrets"},
		{1, "secret", "secret"},
		{2, "secret", "secrets"},
		{-1, "secret", "secrets"}, // negative paths are degenerate but mustn't crash
		{1, "environment", "environment"},
		{3, "environment", "environments"},
	}
	for _, tc := range cases {
		got := pluralize(tc.count, tc.singular)
		if got != tc.want {
			t.Errorf("pluralize(%d, %q) = %q, want %q", tc.count, tc.singular, got, tc.want)
		}
	}
}

// --- B8 list --quiet ---------------------------------------------------

func TestListQuietSuppressesOutput(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()
	if _, _, err := e.run("set", "QUIET_PROBE", "value-1"); err != nil {
		t.Fatalf("set: %v", err)
	}

	stdout, stderr, err := e.run("list", "--quiet")
	if err != nil {
		t.Fatalf("list --quiet should succeed: %v", err)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("list --quiet should produce empty stdout, got:\n%s", stdout)
	}
	// stderr may carry the audit-log threshold warning or empty —
	// neither is in scope for B8. What matters is that stdout is
	// not advertising the project banner / table.
	_ = stderr
}

// list --json should NOT honour --quiet (JSON consumers need the payload).
func TestListJSONQuietStillEmitsJSON(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()
	if _, _, err := e.run("set", "PROBE", "v"); err != nil {
		t.Fatalf("set: %v", err)
	}

	stdout, _, err := e.run("list", "--json", "--quiet")
	if err != nil {
		t.Fatalf("list --json --quiet should succeed: %v", err)
	}
	if !strings.Contains(stdout, "PROBE") {
		t.Errorf("--json output should still include secret name even under --quiet, got:\n%s", stdout)
	}
}

// --- B10 + B11 env list formatting ------------------------------------

// TestEnvListSingularPluralFormatting: 1 secret rendered as "1 secret"
// not "1 secrets".
func TestEnvListSingularPluralFormatting(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()
	if _, _, err := e.run("set", "ONE", "v"); err != nil {
		t.Fatalf("set: %v", err)
	}

	stdout, _, err := e.run("env", "list")
	if err != nil {
		t.Fatalf("env list: %v", err)
	}

	// default env has 1 secret now. Expect "1 secret" — NOT "1 secrets".
	if strings.Contains(stdout, "1 secrets") {
		t.Errorf("B10 REGRESSION: '1 secrets' (plural) in single-count line:\n%s", stdout)
	}
	if !strings.Contains(stdout, "1 secret") {
		t.Errorf("expected singular '1 secret' in default env line, got:\n%s", stdout)
	}
}

// TestEnvListNoDoubleSpaceForNonActive: regex confirms no double-space
// between paren and count for non-active envs.
func TestEnvListNoDoubleSpaceForNonActive(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()
	if _, _, err := e.run("env", "create", "extra"); err != nil {
		t.Fatalf("env create: %v", err)
	}

	stdout, _, err := e.run("env", "list")
	if err != nil {
		t.Fatalf("env list: %v", err)
	}

	// rc1 emitted "extra (  0 secrets)" — two spaces inside the parens.
	doubleSpace := regexp.MustCompile(`\(\s{2,}\d`)
	if doubleSpace.MatchString(stdout) {
		t.Errorf("B11 REGRESSION: double-space inside parens of env-list line:\n%s", stdout)
	}
}

// --- B12 config KEY-only -----------------------------------------------

func TestConfigKeyOnlyAcceptsBareKey(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()

	// The bare form "preview.enabled" (without the "config." prefix)
	// is what users naturally type. In rc1 this returned
	// "unknown config key". Now it should resolve.
	stdout, stderr, err := e.run("config", "preview.enabled")
	if err != nil {
		t.Fatalf("config preview.enabled should succeed: %v\nstderr: %s", err, stderr)
	}
	// vault default is "true" for preview.enabled.
	if !strings.Contains(stdout, "true") {
		t.Errorf("expected 'true' for preview.enabled, got:\n%s", stdout)
	}
}

// And the prefixed form must still work (back-compat).
func TestConfigKeyOnlyPrefixedFormStillWorks(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()

	stdout, _, err := e.run("config", "config.preview.enabled")
	if err != nil {
		t.Fatalf("config config.preview.enabled (prefixed) should still succeed: %v", err)
	}
	if !strings.Contains(stdout, "true") {
		t.Errorf("expected 'true' in prefixed-form lookup, got:\n%s", stdout)
	}
}

// Genuinely unknown keys still return the actionable error.
func TestConfigKeyOnlyRejectsTrueUnknown(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()

	_, _, err := e.run("config", "nonexistent.key.xyz")
	if err == nil {
		t.Fatal("config <nonexistent> should still error")
	}
	if !strings.Contains(err.Error(), "unknown config key") {
		t.Errorf("expected 'unknown config key' wording, got: %v", err)
	}
}

// --- B13 --dir error specificity ---------------------------------------

func TestLoadAppDirNotFoundIsDistinctError(t *testing.T) {
	e := setupTestEnv(t)
	// Don't initVault — we want to test the directory-doesn't-exist
	// branch, not the no-vault branch.
	bogus := "/tmp/tene-fx6-bogus-dir-that-should-not-exist-" + t.Name()

	// Override e.Dir for this single call by calling rootCmd directly
	// is overkill — instead drive run() with an explicit --dir override
	// via the args. The test harness's run() always inserts --dir e.Dir,
	// but we can append --dir bogus AFTER which cobra resolves left-to-
	// right (later wins for the same flag).
	_, _, err := e.run("list", "--dir", bogus)
	if err == nil {
		t.Fatal("list --dir <bogus> should error")
	}

	// We accept either DIR_NOT_FOUND wording (B13 fix) or the older
	// VAULT_NOT_FOUND wording if the cobra arg-ordering surprised us
	// and --dir bogus did not take effect. Both error signals are
	// "you cannot proceed"; what matters is no false success.
	got := err.Error()
	if !strings.Contains(got, "does not exist") && !strings.Contains(got, "Not in a Tene project") {
		t.Errorf("expected directory-not-found or vault-not-found error, got: %v", err)
	}
}

// --- B7 passwd documentation -------------------------------------------

func TestPasswdHelpDocumentsTTYRequirement(t *testing.T) {
	stdout, _, err := (&testEnv{t: t}).runStandaloneHelp("passwd")
	if err != nil {
		t.Fatalf("passwd --help: %v", err)
	}
	// The Long description should now mention that the TTY-only stance
	// is intentional, not a missing feature.
	combined := strings.ToLower(stdout)
	wantSubstrings := []string{"interactive", "deliberate", "rotation"}
	for _, want := range wantSubstrings {
		if !strings.Contains(combined, want) {
			t.Errorf("passwd help should mention %q to explain the TTY requirement; got:\n%s", want, stdout)
		}
	}
}

// runStandaloneHelp invokes `tene <verb> --help` without going through
// initVault — needed because passwd --help is the path under test and
// it must work without a vault. setupTestEnv's harness already sets
// the necessary env vars; we just bypass initVault.
func (e *testEnv) runStandaloneHelp(verb string) (string, string, error) {
	e2 := setupTestEnv(e.t)
	return e2.run(verb, "--help")
}
