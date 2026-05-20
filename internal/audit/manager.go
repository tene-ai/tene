// Package audit provides high-level operations over the vault's
// audit_log table: tail, filtered query, size estimation, and pruning.
//
// The package sits between the cobra CLI commands (internal/cli) and
// the raw SQL surface (internal/vault). Layering rationale:
//
//   - internal/vault owns the SQL chokepoint for DELETE (G10): every
//     audit_log mutation flows through vault.PruneAuditLog. The audit
//     Manager wraps that with business-level concerns (duration → cutoff
//     conversion, returning a count + cutoff for confirmation prompts)
//     so internal/cli can stay free of time arithmetic.
//
//   - internal/audit takes a *vault.Vault by injection. No global
//     state, no init() side effects — tests construct a Manager around
//     a t.TempDir() vault and exercise it directly.
//
//   - The Manager API mirrors the cobra subcommand surface 1:1
//     (Tail / Show / SizeBytes / RowCount / Prune). Each method
//     returns plain Go values; rendering (text vs NDJSON) lives in
//     internal/cli/audit_cmd.go.
//
// The Q3 user decision (master-plan §11) accepted the doubled audit
// log volume from F4 in exchange for full forensic coverage. F8 adds
// the management commands and the 50 MB threshold notice but never
// auto-deletes — invariant I-14 (master-plan §10).
package audit

import (
	"fmt"
	"time"

	"github.com/agent-kay-it/tene/internal/vault"
)

// Manager exposes audit_log operations over an open vault.Vault.
//
// The Manager does not own the vault — callers pass an existing handle
// (typically App.Vault) and remain responsible for closing it. This is
// the same pattern internal/cli's commands use, so a Manager never
// "leaks" the vault handle if the calling RunE returns early.
type Manager struct {
	v *vault.Vault
}

// New constructs a Manager that delegates to v.
//
// New does not retain ownership of v; closing v is the caller's job.
// Passing a nil vault is a programmer error and will panic on first
// method call — we deliberately do not nil-check here to keep the
// happy path branchless.
func New(v *vault.Vault) *Manager {
	return &Manager{v: v}
}

// LogEntry is the user-visible form of one audit_log row. The fields
// mirror the columns the SQL schema exposes (id, action, resource,
// details, timestamp), with NULL coerced to "" at the vault layer so
// callers do not have to handle sql.NullString.
//
// Timestamp is UTC. CLI rendering converts to the user's locale only
// at the very last step (text output); the NDJSON form keeps UTC for
// stable cross-machine forensics.
type LogEntry struct {
	ID        int64     `json:"id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"-"`
}

// Filter narrows a Show() call. All fields are optional.
//
//   - Since / Until are inclusive bounds on the timestamp column. A
//     zero time.Time means "unbounded on this side".
//   - ActionMatch is a SQL LIKE pattern (use '%' for wildcards). When
//     non-empty, the caller-supplied glob style ("cli.metaread*") is
//     translated to LIKE syntax by the CLI layer; the Manager passes
//     the value through verbatim.
//   - Resource is a substring match on the resource_name column,
//     implemented via `resource_name LIKE %Resource%`. When non-empty
//     and combined with ActionMatch, both conditions are AND-ed.
//     Spec source: design.md §6B.1 + plan.md F8 step 3.
//   - Limit > 0 caps the row count; Limit == 0 means unlimited.
type Filter struct {
	Since       time.Time
	Until       time.Time
	ActionMatch string
	Resource    string
	Limit       int
}

// Tail returns the most recent n audit_log rows, newest first.
//
// When n <= 0 the call returns an empty slice (not an error) — the
// CLI layer interprets this as "no rows requested" and prints nothing.
func (m *Manager) Tail(n int) ([]LogEntry, error) {
	if n <= 0 {
		return nil, nil
	}
	rows, err := m.v.QueryAuditLog(vault.AuditLogFilter{MaxRows: n})
	if err != nil {
		return nil, fmt.Errorf("audit: tail: %w", err)
	}
	return toLogEntries(rows), nil
}

// Show returns audit_log rows matching f. The implementation delegates
// directly to vault.QueryAuditLog with one translation pass: the
// caller's ActionMatch is passed through unchanged (the CLI layer is
// expected to add wildcards where appropriate).
func (m *Manager) Show(f Filter) ([]LogEntry, error) {
	rows, err := m.v.QueryAuditLog(vault.AuditLogFilter{
		Since:        f.Since,
		Until:        f.Until,
		ActionLike:   f.ActionMatch,
		ResourceLike: f.Resource,
		MaxRows:      f.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("audit: show: %w", err)
	}
	return toLogEntries(rows), nil
}

// SizeBytes returns the rough byte estimate of audit_log content.
// Accuracy is ±20% (see vault.GetAuditLogSize for the SQL). The
// estimate is intentionally cheap — the threshold warning fires on
// every command and must not add measurable latency.
func (m *Manager) SizeBytes() (int64, error) {
	n, err := m.v.GetAuditLogSize()
	if err != nil {
		return 0, fmt.Errorf("audit: size: %w", err)
	}
	return n, nil
}

// RowCount returns how many rows audit_log currently holds.
// Helper for tests + `tene audit prune --dry-run`.
func (m *Manager) RowCount() (int64, error) {
	// COUNT(*) is the obvious implementation; piggy-back on
	// QueryAuditLog with an empty filter and len() the result rather
	// than adding a new vault method whose only caller is here.
	rows, err := m.v.QueryAuditLog(vault.AuditLogFilter{})
	if err != nil {
		return 0, fmt.Errorf("audit: row count: %w", err)
	}
	return int64(len(rows)), nil
}

// CountOlderThan returns the number of rows strictly older than now -
// olderThan. Used by `tene audit prune` to display "About to delete N
// rows" before requesting confirmation.
//
// olderThan == 0 returns the total row count (everything is older
// than "now"); negative durations are clamped to 0 to avoid pruning
// rows from the future.
func (m *Manager) CountOlderThan(olderThan time.Duration) (int64, error) {
	if olderThan < 0 {
		olderThan = 0
	}
	cutoff := time.Now().UTC().Add(-olderThan)
	n, err := m.v.CountAuditLogOlderThan(cutoff)
	if err != nil {
		return 0, fmt.Errorf("audit: count older than: %w", err)
	}
	return n, nil
}

// Prune deletes every audit_log row older than now - olderThan and
// returns the rows-affected count.
//
// This method is the only API in the entire codebase that ultimately
// reaches `DELETE FROM audit_log` (the SQL lives in
// vault.PruneAuditLog; see G10). The caller (audit_cmd.go) is
// responsible for the safety policy:
//
//   - The PermSecretWrite tier (declared in internal/auth/permissions.go
//     for "audit prune") requires master-password unlock before any
//     destructive op.
//   - Either `--force` or interactive "y" confirmation is required to
//     proceed past the row-count display.
//
// Prune itself trusts the caller has gated correctly. Negative or
// zero olderThan returns an error rather than deleting everything —
// "purge all audit rows" must be an explicit operator decision (out
// of F8 scope; would be a separate command).
func (m *Manager) Prune(olderThan time.Duration) (int64, error) {
	if olderThan <= 0 {
		return 0, fmt.Errorf("audit: prune requires positive olderThan duration (got %v)", olderThan)
	}
	cutoff := time.Now().UTC().Add(-olderThan)
	n, err := m.v.PruneAuditLog(cutoff)
	if err != nil {
		return 0, fmt.Errorf("audit: prune: %w", err)
	}
	return n, nil
}

// toLogEntries maps the vault layer's AuditLogEntry slice to the
// audit layer's LogEntry slice. Both types are deliberately separate
// even though their fields line up 1:1: the vault type belongs to
// the SQL layer (could grow a sql.NullString column in future), and
// the audit type is the stable API the CLI and tests depend on.
func toLogEntries(in []vault.AuditLogEntry) []LogEntry {
	if len(in) == 0 {
		return nil
	}
	out := make([]LogEntry, len(in))
	for i, e := range in {
		out[i] = LogEntry{
			ID:        e.ID,
			Action:    e.Action,
			Resource:  e.Resource,
			Details:   e.Details,
			Timestamp: e.Timestamp,
		}
	}
	return out
}
