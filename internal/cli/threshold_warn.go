// F8 — audit log size threshold notice.
//
// The state machine (design.md §6B.4):
//
//	start → CheckSize:
//	          size < threshold → quiet (no work)
//	          size ≥ threshold → CheckSentinel:
//	                               sentinel mtime < 24h → quiet
//	                               sentinel missing or stale → Warn
//	                                  → stderr line + touch sentinel
//
// Why one notice per 24h rather than always: master-plan §1 RISK H
// + prd.md §5 Failure Mode H call out that emitting the warning on
// every command would turn into spam and trigger "learned ignore"
// behaviour from the user. A 24h refresh window is the design's
// compromise — visible enough to drive eventual action, infrequent
// enough not to bury the rest of stderr.
//
// Why a sentinel file rather than an in-process counter: a tene CLI
// invocation is short-lived (no daemon), so in-memory state would
// not survive the next run. A sentinel under ~/.tene/.audit-warned-<dir-hash>
// is durable across invocations and isolated per project (matches
// the F6 keychain-fallback notice convention so a user grokking
// either feature only learns one mental model).
//
// Why per-project hashing: a user with three vaults on the same
// machine should get three independent notices, not have one
// project's warning suppress the others.
//
// Failure modes:
//
//   - Cannot resolve HOME → skip notice silently. The CLI's primary
//     work must never be blocked by an inability to write a notice.
//
//   - Cannot stat / mkdir / create sentinel → emit notice anyway and
//     skip the touch. The next run will re-emit (slightly annoying
//     but no information loss).
//
//   - Size query errors → skip notice silently. We do not want a
//     vault-read regression to bubble into the dispatcher path.
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/tene-ai/tene/internal/audit"
	"github.com/tene-ai/tene/internal/keychain"
	"github.com/tene-ai/tene/internal/vault"
	"github.com/tene-ai/tene/internal/vaultcfg"
)

// auditSentinelPrefix is the filename stem of the per-project notice
// marker. The directory hash (12 hex chars of sha256(projectDir)) is
// appended to make per-project isolation. The "warned" vocabulary
// mirrors F6's fallback notice for grep ergonomics.
const auditSentinelPrefix = ".audit-warned-"

// auditWarnWindow is how long a notice silences itself for the SAME
// project after being emitted. 24 hours matches the design.md §6B.4
// "warn-once-per-day" contract.
const auditWarnWindow = 24 * time.Hour

// auditSentinelPath returns the absolute path of the threshold-notice
// sentinel for projectDir. Returns an error only when HOME cannot be
// resolved (catastrophic enough that the caller should skip the notice
// rather than report — we cannot write the sentinel anyway).
func auditSentinelPath(projectDir string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	hash := keychain.HashPath(projectDir)
	return filepath.Join(home, sentinelDirName, auditSentinelPrefix+hash), nil
}

// maybeEmitAuditThresholdWarning is the public entry point invoked
// from root.go's PersistentPreRunE. It executes the state machine
// described in the package doc and returns no error — every failure
// is silently swallowed so the CLI's primary work proceeds.
//
// The function is a no-op when:
//
//   - quiet is true (--quiet flag set),
//   - v is nil (PreRunE fired before the vault was opened, e.g. for
//     `tene init` running against a not-yet-existing vault.db),
//   - the size is under threshold,
//   - the sentinel was touched within the warn window.
//
// Otherwise it writes a single line to stderr and touches the sentinel.
func maybeEmitAuditThresholdWarning(stderr io.Writer, v *vault.Vault, projectDir string, quiet bool) {
	if quiet || v == nil {
		return
	}

	thresholdMB := vaultcfg.GetAuditWarnAtMB(v)
	if thresholdMB <= 0 {
		// Defensive: vaultcfg already clamps to [1, 1000], but if a
		// future bug returns 0, treat it as "disabled" rather than
		// fire on every command.
		return
	}
	thresholdBytes := int64(thresholdMB) * 1024 * 1024

	mgr := audit.New(v)
	size, err := mgr.SizeBytes()
	if err != nil {
		// Vault read regression — do not bubble into the dispatcher.
		return
	}
	if size < thresholdBytes {
		return
	}

	sentinel, err := auditSentinelPath(projectDir)
	if err != nil {
		// Cannot resolve HOME → cannot write sentinel → would emit
		// on every command. Emit anyway (one-shot for this process)
		// but skip the touch.
		writeThresholdNotice(stderr, size, thresholdMB)
		return
	}

	if isSentinelFresh(sentinel) {
		return
	}

	// Make sure the parent dir exists before attempting the touch.
	// If MkdirAll fails (FS readonly, permission denied), emit the
	// notice without the touch — best-effort.
	if err := os.MkdirAll(filepath.Dir(sentinel), 0o700); err != nil {
		writeThresholdNotice(stderr, size, thresholdMB)
		return
	}
	if err := touchSentinel(sentinel); err != nil {
		// Cannot persist the sentinel. Still emit the notice. The
		// next run will likely emit again, which is annoying but not
		// a security issue — the user is already over threshold.
		writeThresholdNotice(stderr, size, thresholdMB)
		return
	}

	writeThresholdNotice(stderr, size, thresholdMB)
}

// isSentinelFresh returns true when the sentinel exists as a regular
// file AND was modified less than auditWarnWindow ago. A missing
// sentinel, a stat error, a non-regular entry (e.g. a directory left
// at the path by a previous bug or hostile test fixture), or an mtime
// older than the window all return false so the caller re-emits and
// re-attempts the touch.
//
// Why explicit Mode().IsRegular(): if the sentinel path collides with
// a directory the touchSentinel call below will fail. Without this
// regular-file guard the threshold notice would silently swallow
// itself for the entire 24h window every time a directory squatted on
// the sentinel path — the exact failure mode F8 must defend against
// (TestAuditWarn_SentinelWriteFailure_DoesNotBlockCommand).
func isSentinelFresh(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !st.Mode().IsRegular() {
		return false
	}
	return time.Since(st.ModTime()) < auditWarnWindow
}

// touchSentinel creates the sentinel (truncate to empty) and bumps
// its mtime to now. Idempotent: if the file already exists it is
// re-opened and its mtime is updated via Chtimes.
func touchSentinel(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	_ = f.Close()
	now := time.Now()
	if err := os.Chtimes(path, now, now); err != nil {
		// Some filesystems return EPERM for Chtimes on files just
		// touched; the OpenFile already created or truncated the
		// file so a freshly-created sentinel without an updated
		// mtime is still detectable as "newer than 24h". Swallow.
		return nil
	}
	return nil
}

// resetAuditSentinel removes the sentinel so the next threshold
// check fires again. Called by `tene audit prune` after a successful
// DELETE so that the user gets a fresh notice if the audit_log is
// still over threshold after pruning.
//
// Errors are silently ignored — the sentinel is an idempotency hint,
// not a security boundary.
func resetAuditSentinel(projectDir string) {
	sentinel, err := auditSentinelPath(projectDir)
	if err != nil {
		return
	}
	_ = os.Remove(sentinel)
}

// writeThresholdNotice produces the exact stderr text the user sees.
// Kept separate so tests can reach the same byte sequence without
// reinventing the format string.
//
// Format choices:
//
//   - "note:" lowercase prefix matches F6's fallback notice and Go
//     stdlib conventions for informational stderr lines.
//   - Size rendered in MB (integer) so the number is grokkable; the
//     SizeBytes estimate is ±20% so a fractional precision would
//     be misleading.
//   - The follow-up line names the exact prune command so the user
//     can copy-paste it without a docs round-trip.
func writeThresholdNotice(stderr io.Writer, sizeBytes int64, thresholdMB int) {
	sizeMB := sizeBytes / (1024 * 1024)
	_, _ = fmt.Fprintf(stderr,
		"note: audit log is ~%dMB (threshold %dMB). Run 'tene audit prune --older-than 30d' to clean up.\n",
		sizeMB, thresholdMB,
	)
	_, _ = fmt.Fprintln(stderr,
		"      This message shows once per project per 24h. Pass --quiet to suppress.")
}
