// F4 — Audit Logging by Permission Tier.
//
// Every CLI invocation that flows through cobra's dispatcher must leave a
// single row in audit_log with the action format
//
//	cli.<tier>.<verb>
//
// where <tier> is the lowercase PermLevel name (metaread / secretwrite /
// secretread) and <verb> is the cobra command path with spaces replaced
// by dots (e.g. "env list" -> "env.list", "audit prune" -> "audit.prune").
//
// Examples (gate G7 — master-plan.md §5):
//
//	tene list                  -> cli.metaread.list
//	tene env list              -> cli.metaread.env.list
//	tene set FOO bar           -> cli.secretwrite.set
//	tene audit prune --force   -> cli.secretwrite.audit.prune
//	tene get FOO               -> cli.secretread.get
//
// Design notes:
//
//   - The row is written BEFORE RunE so that failed runs still leave an
//     audit trail of the attempt. The deliberate trade-off: a command
//     that errors out before doing any work still appears in the log,
//     which is correct for security audit (we want to see attempts, not
//     just successes).
//
//   - Args are NEVER recorded in the action column. `tene set MY_KEY
//     SECRET_VALUE` writes only `cli.secretwrite.set` — the key name
//     goes in the resource column, the value is never recorded
//     anywhere by F4. Privacy invariant (master-plan.md §8 I-5 echo).
//
//   - The existing audit rows (vault.init, secret.read, secret.write,
//     env.switch, env.create, env.delete, secrets.export,
//     secrets.import, secrets.inject, secret.delete, vault.passwd_changed,
//     vault.recovered) keep firing exactly as before. F4 is ADDITIVE:
//     one CLI invocation -> 1 cli.* row + N existing rows. Total audit
//     log volume roughly doubles (Q3 user decision, master-plan.md §7).
//
//   - Failure to write the audit row must NEVER block the CLI's primary
//     work. emitCliAuditRow swallows every error path (vault not yet
//     created, DB locked, etc.). The user-facing behaviour of every
//     command is independent of whether the audit write succeeded.
//
//   - `tene init` is special-cased: the vault.db does not exist when
//     rootPersistentPreRunE fires, so the row would silently no-op.
//     init.go therefore calls emitCliAuditRow at the end of its RunE
//     (after the vault is created) so G7 still holds for init.
package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tene-ai/tene/internal/auth"
	"github.com/tene-ai/tene/internal/vault"
)

// auditActionFor formats the action column value for an F4 row.
//
// tier is the resolved PermLevel for the verb; cmdPath is the cobra
// command path minus the "tene " prefix (e.g. "list", "env list",
// "audit prune") — exactly the same key used to look up CommandTier.
//
// Spaces are replaced with dots so the action column reads as a single
// dotted token suitable for `audit_log.action LIKE 'cli.metaread.%'`
// style queries downstream (F8 `tene audit show --filter cli.*`).
func auditActionFor(tier auth.PermLevel, cmdPath string) string {
	verb := strings.ReplaceAll(cmdPath, " ", ".")
	return "cli." + tier.String() + "." + verb
}

// emitCliAuditRow writes one F4 row to audit_log.
//
// The function opens a fresh vault.Vault handle, writes the row, and
// closes it. This is intentionally NOT shared with the caller's existing
// vault handle (when there is one) so that the dispatcher does not need
// to depend on loadApp's full keychain probe — the dispatcher fires
// before loadApp has had a chance to run inside RunE.
//
// Privacy contract:
//
//   - action contains only the verb name and tier; no args, no flags,
//     no values.
//   - resource is the empty string. F4 v0 does not record per-invocation
//     resource identifiers — the existing per-verb audit rows
//     (secret.read, secret.write, etc.) already carry that information
//     and remain in place.
//   - details is the empty string. F8 may extend with timing/result data
//     later (e.g. {"duration_ms": ...}) but never with secret values.
//
// Error policy: every failure path is swallowed. An audit miss is
// strictly worse than a missing row; a blocked command would be a
// regression. The function returns no error by design.
func emitCliAuditRow(dir, action string) {
	if action == "" {
		return
	}
	vaultPath := filepath.Join(dir, ".tene", "vault.db")

	// Skip silently when vault.db does not yet exist. The dominant
	// case is `tene init`: vault.New would otherwise CREATE the file
	// (it ensures parent dir + bootstraps schema), racing with init's
	// own creation flow and tripping init's "vault already exists"
	// branch, which would silently swallow the legacy vault.init
	// audit row. init.go calls AddAuditLog("cli.secretwrite.init",
	// ...) directly from its RunE once the vault is fully created
	// (step 13a), so G7 coverage for the init verb is preserved.
	if _, err := os.Stat(vaultPath); err != nil {
		return
	}

	v, err := vault.New(vaultPath)
	if err != nil {
		// Vault exists but is unreadable (permission denied, schema
		// corruption, ...). Audit miss is preferable to blocking the
		// command — see error policy above.
		return
	}
	defer func() { _ = v.Close() }()

	_ = v.AddAuditLog(action, "", "")
}
