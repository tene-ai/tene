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
	"path/filepath"
	"strings"
	"testing"

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
