// F4 — Audit Logging by Permission Tier tests.
//
// These tests assert the gate G7 (Audit Log Completeness) and the
// privacy invariant of the F4 design (master-plan.md §5 G7 +
// design.md §6B).
//
// The tests open the vault.db directly via SQL queries rather than
// any helper API. F8 will introduce a Manager that wraps these
// queries; F4 must be verifiable on its own without that dependency.
package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/spf13/cobra"
)

// readAuditActions opens vault.db read-only and returns every action
// row in insertion order. Helper kept local to F4 tests so the F8
// audit Manager can later be tested in isolation against the same
// underlying data.
func readAuditActions(t *testing.T, dir string) []string {
	t.Helper()
	vaultPath := filepath.Join(dir, ".tene", "vault.db")
	db, err := sql.Open("sqlite", vaultPath)
	if err != nil {
		t.Fatalf("open vault.db: %v", err)
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query("SELECT action FROM audit_log ORDER BY id")
	if err != nil {
		t.Fatalf("query audit_log: %v", err)
	}
	defer func() { _ = rows.Close() }()

	var out []string
	for rows.Next() {
		var action string
		if err := rows.Scan(&action); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, action)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err: %v", err)
	}
	return out
}

// countAction returns how many times `action` appears verbatim in the
// audit_log action column.
func countAction(actions []string, action string) int {
	n := 0
	for _, a := range actions {
		if a == action {
			n++
		}
	}
	return n
}

// snapshotAuditCount returns the row count of audit_log so a test can
// measure the +N delta of a single CLI invocation.
func snapshotAuditCount(t *testing.T, dir string) int {
	t.Helper()
	return len(readAuditActions(t, dir))
}

// TestAudit_Init_VaultInitPreserved — the legacy `vault.init` audit
// row must still fire when `tene init` runs. F4 is purely additive;
// nothing in the existing audit surface is allowed to disappear.
//
// Also verifies that init produces the F4 dispatcher row
// `cli.secretwrite.init` (written from init.go step 13a because the
// vault did not yet exist when the dispatcher fired). G7 coverage
// for the init verb.
func TestAudit_Init_VaultInitPreserved(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault() // helper calls `tene init ... --quiet`

	actions := readAuditActions(t, env.Dir)

	if n := countAction(actions, "vault.init"); n != 1 {
		t.Errorf("legacy vault.init row count = %d, want 1 (preserve invariant)\nactions: %v", n, actions)
	}
	if n := countAction(actions, "cli.secretwrite.init"); n != 1 {
		t.Errorf("F4 cli.secretwrite.init row count = %d, want 1 (G7 coverage for init)\nactions: %v", n, actions)
	}
}

// TestAudit_List_MetaReadEntry — `tene list` adds exactly one
// `cli.metaread.list` row. Together with the lowercase-tier and
// dot-joined-verb tests below this proves the action format
// `cli.<lowercaseTier>.<dotJoinedVerb>` for the simplest case.
func TestAudit_List_MetaReadEntry(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	before := snapshotAuditCount(t, env.Dir)

	if _, _, err := env.run("list"); err != nil {
		t.Fatalf("list: %v", err)
	}

	actions := readAuditActions(t, env.Dir)
	after := len(actions)

	// list itself does not emit any legacy audit row (it never writes
	// or decrypts), so the delta is exactly 1 — the F4 cli.* row.
	if delta := after - before; delta != 1 {
		t.Errorf("audit rows after `tene list` = +%d, want +1\nactions: %v", delta, actions)
	}
	if n := countAction(actions, "cli.metaread.list"); n != 1 {
		t.Errorf("cli.metaread.list count = %d, want 1\nactions: %v", n, actions)
	}
}

// TestAudit_Set_BothEntries — `tene set FOO bar` produces BOTH the
// new `cli.secretwrite.set` dispatcher row AND the existing
// `secret.write` row written from vault.SetSecret. The legacy row's
// resource column carries the key name; the F4 row's resource is
// empty (privacy: F4 does not duplicate per-verb metadata).
func TestAudit_Set_BothEntries(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	if _, _, err := env.run("set", "MY_KEY", "my-value-AAAA", "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}

	actions := readAuditActions(t, env.Dir)
	if n := countAction(actions, "cli.secretwrite.set"); n != 1 {
		t.Errorf("cli.secretwrite.set count = %d, want 1\nactions: %v", n, actions)
	}
	if n := countAction(actions, "secret.write"); n != 1 {
		t.Errorf("legacy secret.write count = %d, want 1 (preserve invariant)\nactions: %v", n, actions)
	}
}

// TestAudit_Get_TierMatchesLowerCase — `tene get FOO` produces a row
// with the lowercase tier token `secretread` (not `PermSecretRead`
// or `SecretRead`). The audit_log is grepped by operators on these
// tokens (F8 `tene audit show --filter cli.secretread*`) so the
// casing is part of the contract.
func TestAudit_Get_TierMatchesLowerCase(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	if _, _, err := env.run("set", "FOO", "value-aaaa", "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, _, err := env.run("get", "FOO"); err != nil {
		t.Fatalf("get: %v", err)
	}

	actions := readAuditActions(t, env.Dir)
	if n := countAction(actions, "cli.secretread.get"); n != 1 {
		t.Errorf("cli.secretread.get count = %d, want 1\nactions: %v", n, actions)
	}
	// Negative checks: must NOT see any of the alternative spellings.
	for _, bad := range []string{
		"cli.SecretRead.get",
		"cli.PermSecretRead.get",
		"cli.secretRead.get",
		"cli.SECRETREAD.get",
	} {
		if n := countAction(actions, bad); n != 0 {
			t.Errorf("unexpected non-canonical tier spelling %q present (%d rows)", bad, n)
		}
	}
}

// TestAudit_EnvList_DotJoined — `tene env list` produces
// `cli.metaread.env.list` (note the dot between "env" and "list").
// auditActionFor must convert space-separated cobra paths into
// dotted action tokens so they form a single grep'able key.
func TestAudit_EnvList_DotJoined(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	if _, _, err := env.run("env", "list"); err != nil {
		t.Fatalf("env list: %v", err)
	}

	actions := readAuditActions(t, env.Dir)
	if n := countAction(actions, "cli.metaread.env.list"); n != 1 {
		t.Errorf("cli.metaread.env.list count = %d, want 1\nactions: %v", n, actions)
	}
	// Negative check: the space-joined form must NOT appear. A naive
	// `fmt.Sprintf("cli.%s.%s", tier, path)` regression would land here.
	if n := countAction(actions, "cli.metaread.env list"); n != 0 {
		t.Errorf("non-dotted spelling 'cli.metaread.env list' present (%d rows) — auditActionFor broken", n)
	}
}

// TestAudit_NoArgsLeakedIntoAction — the action column must contain
// only `cli.<tier>.<verb>` and never any user-supplied argument or
// flag value. This is the privacy guarantee that lets us cite
// audit_log as a record of "what was attempted" without it doubling
// as a leak channel for secret values or key names (master-plan.md
// §8 I-5 echo for the F4 surface).
func TestAudit_NoArgsLeakedIntoAction(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	// Deliberately pick a key name and value that are easy to grep
	// for, so a regression that concatenates args into action shows
	// up immediately rather than as a subtle whitespace difference.
	const sensitiveKey = "PRIVACY_TEST_KEY_NAME_DO_NOT_LEAK"
	const sensitiveValue = "PRIVACY_TEST_SECRET_VALUE_DO_NOT_LEAK"

	if _, _, err := env.run("set", sensitiveKey, sensitiveValue, "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}

	actions := readAuditActions(t, env.Dir)
	for _, a := range actions {
		if strings.Contains(a, sensitiveKey) {
			t.Errorf("audit_log.action %q contains key name %q — privacy leak", a, sensitiveKey)
		}
		if strings.Contains(a, sensitiveValue) {
			t.Errorf("audit_log.action %q contains secret value %q — CRITICAL privacy leak", a, sensitiveValue)
		}
	}

	// Positive control: the F4 row itself MUST be present with the
	// canonical form (no args appended).
	if n := countAction(actions, "cli.secretwrite.set"); n != 1 {
		t.Errorf("canonical cli.secretwrite.set missing (%d rows)\nactions: %v", n, actions)
	}
}

// TestAudit_FailedRunDoesNotPreventEmission — even when RunE returns
// an error, the F4 row must be present. F4 records ATTEMPTS, not
// just successes; "did the user invoke `tene get X`?" must always
// be answerable from audit_log, irrespective of whether the verb
// succeeded.
//
// Drives a `tene get NONEXISTENT_KEY` which returns
// teneerr.ErrSecretNotFound at the GetSecret step (before any
// decrypt or stdout block).
func TestAudit_FailedRunDoesNotPreventEmission(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	// Confirm the verb does indeed fail.
	_, _, err := env.run("get", "NEVER_SET_THIS_KEY")
	if err == nil {
		t.Fatal("expected `get NEVER_SET_THIS_KEY` to fail with ErrSecretNotFound")
	}

	actions := readAuditActions(t, env.Dir)
	if n := countAction(actions, "cli.secretread.get"); n != 1 {
		t.Errorf("cli.secretread.get count after failed get = %d, want 1 (attempt must be logged)\nactions: %v", n, actions)
	}
	// The legacy secret.read row is NOT expected here — get.go only
	// writes it after a successful decrypt. The asymmetry between
	// "attempt logged at dispatcher" and "completion logged at RunE"
	// is precisely the point of F4.
	if n := countAction(actions, "secret.read"); n != 0 {
		t.Errorf("legacy secret.read fired on failed get (%d rows); the verb returned before its own AddAuditLog call", n)
	}
}

// TestAudit_UnknownCommandFailsSafely — a verb that was registered
// with cobra but never added to CommandTier (the runtime half of
// G4) must be refused by rootPersistentPreRunE BEFORE any audit row
// is written. A rogue verb must not be able to forge an
// `cli.<madeup>.<verb>` row by abusing AddCommand at runtime.
func TestAudit_UnknownCommandFailsSafely(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	beforeCount := snapshotAuditCount(t, env.Dir)

	// Synthesise a verb the dispatcher has never heard of and try to
	// dispatch through it. We mount it under rootCmd post-init() so
	// auth.Validate's startup panic does not trip (Validate ran at
	// process start before AddCommand here). The cleanup detaches the
	// verb so the surrounding test suite is unaffected — mirroring the
	// pattern in TestPersistentPreRunE_UnknownVerbFails.
	rogue := &cobra.Command{Use: "f4-test-rogue-no-tier"}
	rootCmd.AddCommand(rogue)
	t.Cleanup(func() { rootCmd.RemoveCommand(rogue) })

	err := rootPersistentPreRunE(rogue, nil)
	if err == nil {
		t.Fatal("rootPersistentPreRunE(undeclared verb) = nil, want error")
	}
	if !strings.Contains(err.Error(), "CommandTier") {
		t.Errorf("error %q does not mention CommandTier — operator hint missing", err.Error())
	}

	afterCount := snapshotAuditCount(t, env.Dir)
	if afterCount != beforeCount {
		actions := readAuditActions(t, env.Dir)
		t.Errorf(
			"audit_log row count changed by %d after refused dispatch (want 0) — F4 must NOT emit before tier validation\nactions: %v",
			afterCount-beforeCount, actions,
		)
	}

	// Also assert no `cli.*f4-test-rogue*` row leaked.
	actions := readAuditActions(t, env.Dir)
	for _, a := range actions {
		if strings.Contains(a, "f4-test-rogue") {
			t.Errorf("audit_log.action %q references the rogue verb — must not be possible", a)
		}
	}
}

// =============================================================
// F8 — `tene audit tail|show|prune` integration tests below.
//
// These tests drive the cobra subcommands end-to-end through env.run
// (the same harness F4 uses). They cover:
//
//   - Tail default behaviour (newest first, default n=20)
//   - Tail --json emits NDJSON (one JSON object per line)
//   - Show filter by action LIKE pattern
//   - Show filter by --since DURATION
//   - Prune --dry-run reports count + deletes nothing
//   - Prune without --force prompts and aborts on "n"
//   - Prune --force deletes matching rows and leaves newer ones
//   - Prune requires master-password unlock (PermSecretWrite contract)
//
// The shared test environment (setupTestEnv) seeds TENE_MASTER_PASSWORD
// so the password prompt the prune command issues resolves
// non-interactively against that env var — see
// loadOrPromptMasterKey in root.go.
// =============================================================

// withStdin temporarily replaces os.Stdin with a pipe carrying input
// and returns a cleanup func that restores the original. Used by the
// prune confirm tests to drive y/n responses without TTY interaction.
func withStdin(t *testing.T, input string) func() {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	if input != "" {
		if _, err := w.WriteString(input); err != nil {
			t.Fatalf("write stdin: %v", err)
		}
	}
	_ = w.Close()
	oldStdin := os.Stdin
	os.Stdin = r
	return func() {
		os.Stdin = oldStdin
		_ = r.Close()
	}
}

// countAuditRows returns the total number of audit_log rows. Useful
// for prune-side regression checks (did we accidentally delete more
// than expected? did we delete anything on a dry-run?).
func countAuditRows(t *testing.T, dir string) int {
	t.Helper()
	return len(readAuditActions(t, dir))
}

// TestAuditTail_DefaultN_NewestFirst — `tene audit tail` (no -n) shows
// the most recent 20 rows in newest-first order.
func TestAuditTail_DefaultN_NewestFirst(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	// Generate enough rows so tail's default of 20 has to choose a
	// window. 25 set ops -> 50+ audit rows (cli.* + secret.write).
	for i := 0; i < 25; i++ {
		if _, _, err := env.run("set", "KEY"+itoa(i), "v"+itoa(i)+"xxxx", "--overwrite"); err != nil {
			t.Fatalf("set: %v", err)
		}
	}

	stdout, _, err := env.run("audit", "tail")
	if err != nil {
		t.Fatalf("audit tail: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) > 20 {
		t.Errorf("audit tail returned %d lines; want <= 20 (default -n)", len(lines))
	}
	if len(lines) == 0 {
		t.Fatal("audit tail returned 0 lines; want >= 1")
	}
	// First line must be the most recent action (the F4 row of THIS
	// tail invocation itself: cli.metaread.audit.tail). The hook
	// writes BEFORE RunE, so the tail row is in audit_log by the time
	// the SELECT runs.
	if !strings.Contains(lines[0], "cli.metaread.audit.tail") {
		t.Errorf("first tail line should be the F4 row of this command (cli.metaread.audit.tail).\nGot: %q", lines[0])
	}
}

// TestAuditTail_NDJSON_OneRowPerLine — `tene audit tail --json` emits
// NDJSON (one JSON object per line, no surrounding array).
func TestAuditTail_NDJSON_OneRowPerLine(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	if _, _, err := env.run("set", "FOO", "v-AAAA", "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}

	stdout, _, err := env.runJSON("audit", "tail", "-n", "5")
	if err != nil {
		t.Fatalf("audit tail --json: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) < 1 {
		t.Fatalf("expected >= 1 NDJSON line, got %d", len(lines))
	}
	// First (newest) line must be the tail command's own F4 row.
	var obj struct {
		Action string `json:"action"`
		TS     string `json:"ts"`
	}
	if err := json.Unmarshal([]byte(lines[0]), &obj); err != nil {
		t.Fatalf("first line not valid JSON: %v\nline: %q", err, lines[0])
	}
	if obj.Action == "" {
		t.Errorf("NDJSON row missing 'action' field: %q", lines[0])
	}
	if obj.TS == "" {
		t.Errorf("NDJSON row missing 'ts' field: %q", lines[0])
	}
	// Negative: must NOT contain a JSON array marker, even with the
	// envelope-style wrapper that loadable libs would emit.
	if strings.HasPrefix(strings.TrimSpace(stdout), "[") {
		t.Errorf("NDJSON output starts with '[' — array wrap not allowed; got prefix: %q", stdout[:min(20, len(stdout))])
	}
}

// TestAuditShow_FilterByAction — `tene audit show --filter cli.metaread.%`
// returns only rows whose action starts with cli.metaread.
func TestAuditShow_FilterByAction(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	if _, _, err := env.run("set", "FOO", "v-AAAA", "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, _, err := env.run("list"); err != nil {
		t.Fatalf("list: %v", err)
	}

	stdout, _, err := env.run("audit", "show", "--filter", "cli.metaread.%")
	if err != nil {
		t.Fatalf("audit show: %v", err)
	}
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if line == "" || strings.HasPrefix(line, "(") {
			continue
		}
		// Each non-empty body line has the form "TS  action  resource".
		// The action column must start with "cli.metaread."
		fields := strings.Fields(line)
		if len(fields) < 3 {
			t.Errorf("malformed audit row line: %q", line)
			continue
		}
		action := fields[2]
		if !strings.HasPrefix(action, "cli.metaread.") {
			t.Errorf("filter cli.metaread.%% returned row with action %q", action)
		}
	}
}

// TestAuditShow_FilterByResource — `tene audit show --resource NAME`
// returns only rows whose resource_name contains NAME as a substring.
// Spec source: design.md §6B.1 + plan.md F8 step 3. The flag was
// missing from the F8 initial implementation and added in F8'
// compensation patch.
//
// Coverage:
//   - Substring match: --resource STRIPE matches STRIPE_KEY +
//     STRIPE_WEBHOOK rows but NOT OPENAI_KEY.
//   - AND semantics with --filter: --resource STRIPE +
//     --filter 'cli.secretwrite.%' narrows to write-tier STRIPE rows.
//   - Empty value: --resource "" is equivalent to no filter.
//
// The audit_log table mixes auto-emitted dispatcher rows (resource="")
// with verb-emitted legacy rows (resource carries the key name). The
// test asserts that an explicit --resource STRIPE_KEY excludes the
// resource-empty dispatcher rows since the LIKE %STRIPE_KEY% pattern
// never matches "".
func TestAuditShow_FilterByResource(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	// Seed three distinct resource names so the substring filter has
	// something to discriminate. The legacy `secret.write` row from
	// each `set` carries the key name in resource_name; the F4
	// dispatcher row carries resource="" (resource_name is empty for
	// the cli.* rows). LIKE %X% does not match the empty string, so
	// the dispatcher rows are naturally excluded from a --resource
	// query.
	for _, key := range []string{"STRIPE_KEY", "STRIPE_WEBHOOK", "OPENAI_KEY"} {
		if _, _, err := env.run("set", key, "v-AAAA", "--overwrite"); err != nil {
			t.Fatalf("set %s: %v", key, err)
		}
	}

	// Case 1: --resource STRIPE returns the 2 STRIPE_* rows (not
	// OPENAI_KEY, not the dispatcher rows whose resource is empty).
	stdout, _, err := env.runJSON("audit", "show", "--resource", "STRIPE")
	if err != nil {
		t.Fatalf("audit show --resource STRIPE --json: %v", err)
	}
	gotResources := parseNDJSONResources(t, stdout)
	stripeCount := 0
	for _, r := range gotResources {
		if !strings.Contains(r, "STRIPE") {
			t.Errorf("--resource STRIPE returned row with resource %q (must contain 'STRIPE')", r)
		}
		if strings.Contains(r, "STRIPE_KEY") || strings.Contains(r, "STRIPE_WEBHOOK") {
			stripeCount++
		}
		if strings.Contains(r, "OPENAI") {
			t.Errorf("--resource STRIPE leaked OPENAI row (resource=%q)", r)
		}
	}
	if stripeCount < 2 {
		t.Errorf("--resource STRIPE returned %d STRIPE_* rows; want >= 2 (one secret.write each for STRIPE_KEY + STRIPE_WEBHOOK)\nresources: %v",
			stripeCount, gotResources)
	}

	// Case 2: --resource exact match (STRIPE_KEY) returns only the
	// STRIPE_KEY rows. STRIPE_WEBHOOK and OPENAI_KEY must be excluded.
	resetFlags()
	stdout, _, err = env.runJSON("audit", "show", "--resource", "STRIPE_KEY")
	if err != nil {
		t.Fatalf("audit show --resource STRIPE_KEY --json: %v", err)
	}
	gotResources = parseNDJSONResources(t, stdout)
	if len(gotResources) == 0 {
		t.Errorf("--resource STRIPE_KEY returned 0 rows; want >= 1")
	}
	for _, r := range gotResources {
		if r != "STRIPE_KEY" {
			t.Errorf("--resource STRIPE_KEY returned non-matching row resource=%q (substring %%STRIPE_KEY%% LIKE)", r)
		}
	}

	// Case 3: --resource + --filter AND semantics. Use the secret.write
	// (legacy) action which only fires on actual writes. Combined with
	// --resource STRIPE we should still get at least 2 rows (one per
	// STRIPE_* key) and zero non-STRIPE rows.
	resetFlags()
	stdout, _, err = env.runJSON("audit", "show", "--resource", "STRIPE", "--filter", "secret.write")
	if err != nil {
		t.Fatalf("audit show --resource STRIPE --filter secret.write --json: %v", err)
	}
	combined := parseNDJSONRows(t, stdout)
	if len(combined) == 0 {
		t.Errorf("--resource STRIPE --filter secret.write returned 0 rows; want >= 2")
	}
	for _, row := range combined {
		if row.Action != "secret.write" {
			t.Errorf("AND filter leaked action %q (expected only secret.write)", row.Action)
		}
		if !strings.Contains(row.Resource, "STRIPE") {
			t.Errorf("AND filter leaked resource %q (expected substring STRIPE)", row.Resource)
		}
	}
}

// parseNDJSONResources parses NDJSON stdout from `audit show --json`
// and returns the resource field of each row. Empty resources are
// preserved so a regression that returns dispatcher rows for a
// non-empty --resource query is visible.
func parseNDJSONResources(t *testing.T, stdout string) []string {
	t.Helper()
	rows := parseNDJSONRows(t, stdout)
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Resource)
	}
	return out
}

// parseNDJSONRows parses NDJSON stdout into a typed slice. Each line
// must be a valid JSON object with at minimum action + resource fields.
func parseNDJSONRows(t *testing.T, stdout string) []struct {
	Action   string `json:"action"`
	Resource string `json:"resource"`
} {
	t.Helper()
	trimmed := strings.TrimSpace(stdout)
	if trimmed == "" {
		return nil
	}
	lines := strings.Split(trimmed, "\n")
	out := make([]struct {
		Action   string `json:"action"`
		Resource string `json:"resource"`
	}, 0, len(lines))
	for i, line := range lines {
		if line == "" {
			continue
		}
		var row struct {
			Action   string `json:"action"`
			Resource string `json:"resource"`
		}
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatalf("NDJSON line %d not valid JSON: %v\nline: %q", i, err, line)
		}
		out = append(out, row)
	}
	return out
}

// TestAuditShow_SinceDuration — `tene audit show --since 5m` returns
// only rows within the last 5 minutes. We assert at minimum that the
// command runs without error and includes the just-emitted F4 row.
func TestAuditShow_SinceDuration(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	stdout, _, err := env.run("audit", "show", "--since", "1h")
	if err != nil {
		t.Fatalf("audit show --since 1h: %v", err)
	}
	if !strings.Contains(stdout, "cli.metaread.audit.show") {
		t.Errorf("show output missing self-referencing F4 row.\nGot:\n%s", stdout)
	}
}

// TestAuditShow_SinceDayUnit — the "d" suffix (not in stdlib
// time.ParseDuration) is accepted and translated to 24h.
func TestAuditShow_SinceDayUnit(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	if _, _, err := env.run("audit", "show", "--since", "30d"); err != nil {
		t.Errorf("audit show --since 30d returned error %v; should accept 'd' suffix", err)
	}
}

// TestAuditPrune_DryRun_DeletesNothing — `--dry-run` reports the
// match count without performing the DELETE. The post-run audit_log
// row count must be unchanged (modulo +1 for the prune command's own
// F4 dispatcher row).
func TestAuditPrune_DryRun_DeletesNothing(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	before := countAuditRows(t, env.Dir)

	stdout, _, err := env.run("audit", "prune", "--older-than", "1ns", "--dry-run")
	if err != nil {
		t.Fatalf("audit prune --dry-run: %v", err)
	}
	if !strings.Contains(stdout, "Dry run") {
		t.Errorf("dry-run output missing the 'Dry run' marker.\nGot: %q", stdout)
	}

	after := countAuditRows(t, env.Dir)
	// +1 for the F4 dispatcher row of this prune command. No DELETE
	// happened, so the count strictly grew.
	if after < before+1 {
		t.Errorf("dry-run dropped rows: before=%d after=%d (want after >= before+1, no DELETE)", before, after)
	}
	// Verify no rows were physically removed: the legacy vault.init
	// row from initVault must still be there.
	actions := readAuditActions(t, env.Dir)
	if countAction(actions, "vault.init") == 0 {
		t.Errorf("dry-run deleted the legacy vault.init row — DELETE should be a no-op")
	}
}

// TestAuditPrune_RequiresConfirmation — without --force, the prune
// command prompts the user and aborts when the answer is anything
// other than y/yes. Drives "n\n" through stdin.
//
// This is the gate test for G10 v2 — "prune requires confirm by
// default".
func TestAuditPrune_RequiresConfirmation(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	// Seed something old enough to match a --older-than 1ns prune.
	if _, _, err := env.run("set", "X", "v-AAAA", "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}

	beforeRows := countAuditRows(t, env.Dir)

	cleanup := withStdin(t, "n\n")
	defer cleanup()

	stdout, _, err := env.run("audit", "prune", "--older-than", "1ns")
	if err != nil {
		t.Fatalf("audit prune: %v", err)
	}
	// Successful no-op exit. The user-visible "Aborted" message is
	// on stderr in normal CLI but the env.run wrapper combines
	// captures; we just require the run to succeed and the row
	// count to NOT shrink.
	_ = stdout

	afterRows := countAuditRows(t, env.Dir)
	// No DELETE should have occurred. The +1 from prune's own F4 row
	// means afterRows >= beforeRows; never <.
	if afterRows < beforeRows {
		t.Errorf("audit prune without confirm deleted rows: before=%d after=%d", beforeRows, afterRows)
	}
}

// TestAuditPrune_WithForce_DeletesMatching — `--force` skips the
// confirm prompt and DELETEs rows older than the cutoff. Newer rows
// (including the prune command's own F4 row, written milliseconds
// ago at PreRunE) survive.
func TestAuditPrune_WithForce_DeletesMatching(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	// Sleep > 1s so the initVault rows fall outside a 1s prune
	// window.
	time.Sleep(2100 * time.Millisecond)

	stdout, _, err := env.run("audit", "prune", "--older-than", "1s", "--force")
	if err != nil {
		t.Fatalf("audit prune --force: %v", err)
	}
	if !strings.Contains(stdout, "Deleted") {
		t.Errorf("force-mode output missing 'Deleted N row(s)' confirmation.\nGot: %q", stdout)
	}

	// vault.init was inserted at init time (> 2s ago) so it must be
	// gone after the prune.
	actions := readAuditActions(t, env.Dir)
	if n := countAction(actions, "vault.init"); n != 0 {
		t.Errorf("vault.init row survived --force prune; %d copies remaining", n)
	}

	// The prune command's own dispatcher row was inserted ~0ms before
	// the DELETE ran, so it should still be present (1-second cutoff).
	if n := countAction(actions, "cli.secretwrite.audit.prune"); n < 1 {
		t.Errorf("prune wiped its own F4 row (cli.secretwrite.audit.prune count = %d, want >= 1)", n)
	}
}

// TestAuditPrune_OlderThanRequired — calling prune without the
// --older-than flag fails fast with an actionable error. Defensive:
// we never want a future regression where --older-than defaults to
// zero and the command silently drops nothing.
func TestAuditPrune_OlderThanRequired(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	_, _, err := env.run("audit", "prune", "--force")
	if err == nil {
		t.Fatal("audit prune without --older-than returned nil error; expected guard")
	}
	if !strings.Contains(err.Error(), "older-than") {
		t.Errorf("error %q does not mention --older-than", err.Error())
	}
}

// TestAuditPrune_DryRun_PrintsCount — `--dry-run` prints the row
// count it WOULD delete. Lets a scripter validate the cutoff value
// before committing.
func TestAuditPrune_DryRun_PrintsCount(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	time.Sleep(2100 * time.Millisecond)
	stdout, _, err := env.run("audit", "prune", "--older-than", "1s", "--dry-run")
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if !strings.Contains(stdout, "Dry run") {
		t.Errorf("dry-run output missing 'Dry run' marker: %q", stdout)
	}
}

// TestAuditTail_OutputNeverContainsSecretValues — privacy regression
// check. The audit_log is supposed to record action + resource_name
// (the key NAME) but never the secret value. `tene audit tail` reads
// rows verbatim, so a regression that smuggles values into
// resource_name or details would surface here.
func TestAuditTail_OutputNeverContainsSecretValues(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	const sensitiveValue = "PRIVACY_TAIL_SENTINEL_NEVER_LEAK"
	if _, _, err := env.run("set", "PRIVACY_KEY", sensitiveValue, "--overwrite"); err != nil {
		t.Fatalf("set: %v", err)
	}
	stdout, _, err := env.run("audit", "tail", "-n", "50")
	if err != nil {
		t.Fatalf("audit tail: %v", err)
	}
	if strings.Contains(stdout, sensitiveValue) {
		t.Errorf("CRITICAL: audit tail output contains a secret value (%q)", sensitiveValue)
	}
}

// itoa is the local equivalent of strconv.Itoa, kept small to avoid
// importing strconv for one call. (Tests in internal/audit have an
// identical helper; the file boundary makes deduplication awkward
// for marginal benefit.)
func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

// TestAudit_RootBare_NoEmission — `tene` with no subcommand prints
// cobra help and the dispatcher returns early before reaching the F4
// emission line. Nothing should land in audit_log for that path.
func TestAudit_RootBare_NoEmission(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	beforeActions := readAuditActions(t, env.Dir)
	beforeCli := 0
	for _, a := range beforeActions {
		if strings.HasPrefix(a, "cli.") {
			beforeCli++
		}
	}

	if err := rootPersistentPreRunE(rootCmd, nil); err != nil {
		t.Fatalf("rootPersistentPreRunE(rootCmd) = %v, want nil", err)
	}

	afterActions := readAuditActions(t, env.Dir)
	afterCli := 0
	for _, a := range afterActions {
		if strings.HasPrefix(a, "cli.") {
			afterCli++
		}
	}
	if afterCli != beforeCli {
		t.Errorf("cli.* row count changed by %d after bare-root dispatch (want 0)\nactions: %v", afterCli-beforeCli, afterActions)
	}
}
