// F5 — `tene permissions` info command tests.
//
// These tests assert four properties of the verb:
//
//   1. Text mode lists every entry in internal/auth.CommandTier, with
//      no fabrications and no omissions. (The CommandTier map is the
//      single source of truth — this test enforces that.)
//   2. Text mode footer reports the correct per-tier counts.
//   3. JSON mode emits the documented shape: {ok, count, byTier,
//      commands[...]} with commands sorted by (tier-order, name).
//   4. Running the verb never opens a master-password prompt: it must
//      complete WITHOUT a vault or keychain (we run it from an empty
//      directory with no .tene/vault.db on disk).
//
// The tests use the same setupTestEnv / e.run harness as other F2-F8
// cli tests (testhelper_test.go) for consistency.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/tene-ai/tene/internal/auth"
)

// TestPermissions_Text_AllEntriesPresent asserts that every command
// name in auth.CommandTier appears in the text output and every tier
// name appears at least once. This is the regression guard against
// drift between auth.CommandTier and the permissions command's
// rendering layer — the table cannot silently lose an entry.
func TestPermissions_Text_AllEntriesPresent(t *testing.T) {
	e := setupTestEnv(t)
	stdout, stderr, err := e.run("permissions")
	if err != nil {
		t.Fatalf("permissions failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	for name := range auth.CommandTier {
		if !strings.Contains(stdout, name) {
			t.Errorf("text output is missing CommandTier entry %q\n--- stdout ---\n%s", name, stdout)
		}
	}

	// Each tier string token must show up at least once in the body.
	for _, token := range []string{"metaread", "secretwrite", "secretread"} {
		if !strings.Contains(stdout, token) {
			t.Errorf("text output missing tier token %q\n--- stdout ---\n%s", token, stdout)
		}
	}

	// Header row + PASSWORD? column must be present (visual sanity).
	for _, hdr := range []string{"COMMAND", "TIER", "PASSWORD?"} {
		if !strings.Contains(stdout, hdr) {
			t.Errorf("text output missing header column %q", hdr)
		}
	}
}

// TestPermissions_Text_Counts asserts the footer summary line states
// the correct per-tier breakdown. The expected counts are computed
// from auth.CommandTier rather than hard-coded so the test stays
// honest if the table grows (e.g. when a future feature adds a verb).
func TestPermissions_Text_Counts(t *testing.T) {
	e := setupTestEnv(t)
	stdout, stderr, err := e.run("permissions")
	if err != nil {
		t.Fatalf("permissions failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	want := map[auth.PermLevel]int{}
	for _, lvl := range auth.CommandTier {
		want[lvl]++
	}
	total := len(auth.CommandTier)

	expected := fmt.Sprintf(
		"Total: %d commands  (%d metaread / %d secretwrite / %d secretread)",
		total,
		want[auth.PermMetaRead],
		want[auth.PermSecretWrite],
		want[auth.PermSecretRead],
	)
	if !strings.Contains(stdout, expected) {
		t.Errorf("footer summary mismatch\nwant substring: %q\ngot stdout:\n%s", expected, stdout)
	}

	// Sprint-locked totals — if these change, the design doc §1.1
	// CommandTier diagram and master-plan §11 must also change. The
	// hard-coded numbers double-check that the dynamic count above
	// matches the documented 15/5/5 breakdown (down from 16/5/5 after
	// sprint v1014-rc1-qa-fixes/FX4 removed the stale `logout`
	// CommandTier entry — see CHANGELOG.md B5 fix).
	if total != 25 {
		t.Errorf("expected 25 total commands (15+5+5 post-FX4), got %d", total)
	}
	if got := want[auth.PermMetaRead]; got != 15 {
		t.Errorf("expected 15 metaread entries (post-FX4), got %d", got)
	}
	if got := want[auth.PermSecretWrite]; got != 5 {
		t.Errorf("expected 5 secretwrite entries per design.md §1.1, got %d", got)
	}
	if got := want[auth.PermSecretRead]; got != 5 {
		t.Errorf("expected 5 secretread entries per design.md §1.1, got %d", got)
	}
}

// permissionsJSONDoc mirrors the documented JSON shape so the unit
// test can decode and assert each field. Keep this struct in sync
// with writePermissionsJSON in permissions.go.
type permissionsJSONDoc struct {
	OK     bool           `json:"ok"`
	Count  int            `json:"count"`
	ByTier map[string]int `json:"byTier"`
	// Commands intentionally typed as []map[string]any to verify the
	// field set ({name, tier, requiresUnlock}) and types via reflection.
	Commands []map[string]any `json:"commands"`
}

// TestPermissions_JSON_Structure decodes the --json output and checks
// the top-level fields: ok, count, byTier counts, and the commands
// array length. This is the schema contract test that protects scripts
// and AI agents from breaking on internal refactors.
func TestPermissions_JSON_Structure(t *testing.T) {
	e := setupTestEnv(t)
	stdout, stderr, err := e.runJSON("permissions")
	if err != nil {
		t.Fatalf("permissions --json failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	var doc permissionsJSONDoc
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("unmarshal JSON: %v\nstdout:\n%s", err, stdout)
	}

	if !doc.OK {
		t.Errorf("expected ok=true, got false")
	}
	if doc.Count != len(auth.CommandTier) {
		t.Errorf("count mismatch: want %d, got %d", len(auth.CommandTier), doc.Count)
	}

	want := map[string]int{
		"metaread":    0,
		"secretwrite": 0,
		"secretread":  0,
	}
	for _, lvl := range auth.CommandTier {
		want[lvl.String()]++
	}
	for tier, n := range want {
		if doc.ByTier[tier] != n {
			t.Errorf("byTier[%q]: want %d, got %d", tier, n, doc.ByTier[tier])
		}
	}
}

// TestPermissions_JSON_CommandsArray_HasAllEntries asserts that the
// JSON commands array contains exactly one object per auth.CommandTier
// entry, with all three required fields populated and typed correctly.
func TestPermissions_JSON_CommandsArray_HasAllEntries(t *testing.T) {
	e := setupTestEnv(t)
	stdout, _, err := e.runJSON("permissions")
	if err != nil {
		t.Fatalf("permissions --json failed: %v", err)
	}

	var doc permissionsJSONDoc
	if err := json.Unmarshal([]byte(stdout), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(doc.Commands) != len(auth.CommandTier) {
		t.Fatalf("commands array length: want %d, got %d", len(auth.CommandTier), len(doc.Commands))
	}

	// Walk every emitted object and verify required-field presence.
	seen := make(map[string]bool, len(doc.Commands))
	for i, c := range doc.Commands {
		name, okName := c["name"].(string)
		tier, okTier := c["tier"].(string)
		// JSON unmarshal of a Go bool field produces bool; verify type.
		_, okUnlock := c["requiresUnlock"].(bool)

		if !okName {
			t.Errorf("commands[%d].name missing or not a string: %#v", i, c["name"])
		}
		if !okTier {
			t.Errorf("commands[%d].tier missing or not a string: %#v", i, c["tier"])
		}
		if !okUnlock {
			t.Errorf("commands[%d].requiresUnlock missing or not bool: %#v", i, c["requiresUnlock"])
		}
		if okName {
			seen[name] = true
		}
		// Cross-check requiresUnlock against the source-of-truth tier.
		expectedTier, ok := auth.CommandTier[name]
		if !ok {
			t.Errorf("commands[%d] name %q not in auth.CommandTier", i, name)
			continue
		}
		if tier != expectedTier.String() {
			t.Errorf("commands[%d] tier mismatch for %q: want %q got %q",
				i, name, expectedTier.String(), tier)
		}
		wantUnlock := expectedTier.RequiresUnlock()
		gotUnlock, _ := c["requiresUnlock"].(bool)
		if gotUnlock != wantUnlock {
			t.Errorf("commands[%d] requiresUnlock for %q: want %v got %v",
				i, name, wantUnlock, gotUnlock)
		}
	}

	// Every CommandTier entry must appear exactly once.
	for name := range auth.CommandTier {
		if !seen[name] {
			t.Errorf("CommandTier entry %q absent from JSON commands array", name)
		}
	}
}

// TestPermissions_NoPasswordPrompt asserts the verb runs to completion
// WITHOUT a vault — proving the dispatcher classified it as
// PermMetaRead and never invoked loadOrPromptMasterKey or opened
// .tene/vault.db. We make the absence of a vault explicit by running
// in a fresh temp dir that has no `.tene/` subtree at all.
//
// If the verb were ever mistakenly elevated to PermSecretWrite /
// PermSecretRead the dispatcher would route it through unlock, which
// would in turn either (a) prompt on stdin (this test would block /
// fail) or (b) error out with ErrVaultNotFound. Both failure modes
// cause this test to fail and surface the regression immediately.
func TestPermissions_NoPasswordPrompt(t *testing.T) {
	// Independent temp dir — NOT a tene project, no vault.db, no
	// keychain entry. Just a directory.
	dir := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	// Explicitly unset the env-var fallback so any sneak path that
	// tried to derive a master key would fail loudly.
	t.Setenv("TENE_MASTER_PASSWORD", "")

	e := &testEnv{Dir: dir, HomeDir: home, t: t}
	stdout, stderr, err := e.run("permissions")
	if err != nil {
		t.Fatalf("permissions ran into an unlock path it should not have\n"+
			"err: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	// Sanity: still emitted the expected table.
	if !strings.Contains(stdout, "PASSWORD?") {
		t.Errorf("text output missing PASSWORD? header in fresh-dir run\nstdout:\n%s", stdout)
	}
	// And in particular the verb itself must show up — declaring
	// itself as no-password (`no` in the rightmost column).
	if !strings.Contains(stdout, "permissions") {
		t.Errorf("text output missing 'permissions' entry in fresh-dir run\nstdout:\n%s", stdout)
	}
}

// TestPermissions_Sorted_AlphabeticalWithinTier asserts the row order:
// tier-order first (metaread → secretwrite → secretread), then
// alphabetical within each tier. The test parses the text output and
// walks the data rows.
func TestPermissions_Sorted_AlphabeticalWithinTier(t *testing.T) {
	e := setupTestEnv(t)
	stdout, _, err := e.run("permissions")
	if err != nil {
		t.Fatalf("permissions failed: %v", err)
	}

	// Skip header line + extract (name, tier) pairs in stream order.
	lines := strings.Split(stdout, "\n")
	var emittedTiers []string
	var emittedNames []string
	currentTier := ""
	for _, line := range lines {
		// Skip header, footer, and blank lines.
		if line == "" || strings.HasPrefix(line, "COMMAND") || strings.HasPrefix(line, "Total:") {
			continue
		}
		fields := strings.Fields(line)
		// A row in our format is at least 3 fields: NAME TIER PASSWORD?
		// "env list" rows have 4 (name = two tokens); fold those back.
		if len(fields) < 3 {
			continue
		}
		// The last field is yes/no. The second-to-last is the tier
		// token. Everything before is the name (potentially space-
		// joined, e.g. "env list").
		tier := fields[len(fields)-2]
		name := strings.Join(fields[:len(fields)-2], " ")

		// Only count valid tier rows (avoid mis-parsing accidental
		// header continuations).
		switch tier {
		case "metaread", "secretwrite", "secretread":
			// ok
		default:
			continue
		}

		if tier != currentTier {
			emittedTiers = append(emittedTiers, tier)
			currentTier = tier
		}
		emittedNames = append(emittedNames, fmt.Sprintf("%s|%s", tier, name))
	}

	// Tier order must match tierOrder in permissions.go.
	wantTierSeq := []string{"metaread", "secretwrite", "secretread"}
	if !sliceEqual(emittedTiers, wantTierSeq) {
		t.Errorf("tier ordering mismatch\nwant: %v\ngot:  %v\nstdout:\n%s",
			wantTierSeq, emittedTiers, stdout)
	}

	// Within each tier, names must be in ascending lexical order.
	byTier := make(map[string][]string)
	for _, entry := range emittedNames {
		parts := strings.SplitN(entry, "|", 2)
		byTier[parts[0]] = append(byTier[parts[0]], parts[1])
	}
	for tier, names := range byTier {
		sorted := append([]string(nil), names...)
		sort.Strings(sorted)
		if !sliceEqual(names, sorted) {
			t.Errorf("within-tier %s not alphabetically sorted\ngot:    %v\nwanted: %v",
				tier, names, sorted)
		}
	}
}

// sliceEqual compares two string slices for exact, in-order equality.
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestPermissions_F4AuditRow asserts that invoking `tene permissions`
// inside an initialized vault produces exactly one cli.metaread.permissions
// audit row (the F4 dispatcher emission). This is the F4 ↔ F5
// integration probe and the G7 (Audit Log Completeness) guarantee for
// the new verb.
func TestPermissions_F4AuditRow(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()

	before := snapshotAuditCount(t, e.Dir)

	stdout, stderr, err := e.run("permissions")
	if err != nil {
		t.Fatalf("permissions failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}

	actions := readAuditActions(t, e.Dir)
	if len(actions) <= before {
		t.Fatalf("expected at least 1 new audit row after permissions, got delta=%d",
			len(actions)-before)
	}

	const want = "cli.metaread.permissions"
	if got := countAction(actions[before:], want); got != 1 {
		t.Errorf("expected exactly 1 %q row, got %d\nactions: %v",
			want, got, actions[before:])
	}
}

// TestPermissions_VerbItself_DeclaredMetaRead is a static, package-
// local guard rail. F4 / F8 emit cli.<tier>.<verb> rows reading the
// tier from auth.CommandTier. If a refactor ever flipped the
// 'permissions' tier to PermSecretWrite/Read this assertion would
// fail before the verb shipped a regression — and the audit row
// prefix would silently change from cli.metaread.* to
// cli.secret*.permissions, which is the kind of subtle drift this
// guard exists to catch.
//
// Note: this duplicates intent with the panic-on-missing check in
// auth.Validate(), but Validate only asserts the entry EXISTS — it
// does not pin which tier the entry is at. This test pins the tier.
func TestPermissions_VerbItself_DeclaredMetaRead(t *testing.T) {
	tier, ok := auth.TierFor("permissions")
	if !ok {
		t.Fatalf("auth.CommandTier missing 'permissions' entry")
	}
	if tier != auth.PermMetaRead {
		t.Errorf("permissions verb must be PermMetaRead (no password); got %s", tier)
	}
	if tier.RequiresUnlock() {
		t.Errorf("permissions verb RequiresUnlock() = true; must be false")
	}
}

// Build-time guard: ensure permissions_test.go's local helpers don't
// shadow std lib types accidentally. (errors import is for the future
// negative-path test; left as a placeholder reference so removing it
// surfaces in CR rather than silently in a noise commit.)
var _ = errors.New
var _ = os.Stat
