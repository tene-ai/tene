// Tests for internal/audit.Manager.
//
// The tests construct a vault.Vault on a t.TempDir() SQLite file and
// drive Manager methods directly. No cobra involved — those are
// exercised in internal/cli/audit_test.go. The split keeps unit
// failures localized: a regression in Tail() shows up here, a
// regression in `tene audit tail`'s flag wiring shows up in the cli
// suite.
package audit

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/tene-ai/tene/internal/vault"
)

// newTestManager opens a fresh vault under t.TempDir() and returns a
// Manager around it. Closing happens via t.Cleanup so a fatal in
// the middle of a test still releases the SQLite file lock.
func newTestManager(t *testing.T) (*Manager, *vault.Vault) {
	t.Helper()
	dir := t.TempDir()
	v, err := vault.New(filepath.Join(dir, "vault.db"))
	if err != nil {
		t.Fatalf("vault.New: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })
	return New(v), v
}

// seedRows inserts n synthetic audit_log rows with a known cadence so
// time-window tests can target specific subsets. The actions follow
// the F4 format `cli.<tier>.<verb>` so ActionMatch tests can be
// meaningful; resource carries a per-row sentinel for privacy regress
// detection.
func seedRows(t *testing.T, v *vault.Vault, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		var action string
		switch i % 3 {
		case 0:
			action = "cli.metaread.list"
		case 1:
			action = "cli.secretwrite.set"
		default:
			action = "cli.secretread.get"
		}
		if err := v.AddAuditLog(action, "RESOURCE_"+itoa(i), ""); err != nil {
			t.Fatalf("AddAuditLog: %v", err)
		}
		// SQLite datetime('now') has 1-second resolution; sleep a
		// hair so consecutive rows do not collide on the timestamp
		// when the test environment is fast enough to fit multiple
		// inserts in the same wallclock second. Without this, the
		// MaxRows + ORDER BY id DESC ordering still works (id is
		// monotonic), but Since/Until filter tests need distinct
		// timestamps to be deterministic.
		time.Sleep(1100 * time.Millisecond)
	}
}

// itoa is a tiny shim so we do not import strconv just for one call.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 4)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}

func TestManager_Tail_DefaultOrder_NewestFirst(t *testing.T) {
	mgr, v := newTestManager(t)
	// Use raw AddAuditLog so the test does not pay for a 3-second
	// sleep when the only thing we care about is ordering.
	for i := 0; i < 5; i++ {
		if err := v.AddAuditLog("cli.metaread.list", "R"+itoa(i), ""); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	rows, err := mgr.Tail(3)
	if err != nil {
		t.Fatalf("Tail: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("Tail(3) returned %d rows, want 3", len(rows))
	}
	// Newest first: id DESC. Last inserted resource is "R4".
	if rows[0].Resource != "R4" {
		t.Errorf("Tail[0].Resource = %q, want %q (newest first contract)", rows[0].Resource, "R4")
	}
	if rows[2].Resource != "R2" {
		t.Errorf("Tail[2].Resource = %q, want %q (newest-first window of 3)", rows[2].Resource, "R2")
	}
}

func TestManager_Tail_ZeroOrNegative_ReturnsEmpty(t *testing.T) {
	mgr, v := newTestManager(t)
	_ = v.AddAuditLog("cli.metaread.list", "x", "")

	for _, n := range []int{0, -1, -100} {
		rows, err := mgr.Tail(n)
		if err != nil {
			t.Errorf("Tail(%d) returned error %v; expected (nil, nil)", n, err)
		}
		if len(rows) != 0 {
			t.Errorf("Tail(%d) returned %d rows; want 0", n, len(rows))
		}
	}
}

func TestManager_Show_FilterByActionMatch(t *testing.T) {
	mgr, v := newTestManager(t)
	// Insert rows interleaving two distinct actions.
	_ = v.AddAuditLog("cli.metaread.list", "a", "")
	_ = v.AddAuditLog("cli.secretwrite.set", "b", "")
	_ = v.AddAuditLog("cli.metaread.list", "c", "")
	_ = v.AddAuditLog("cli.secretread.get", "d", "")

	rows, err := mgr.Show(Filter{ActionMatch: "cli.metaread.%"})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("Show(cli.metaread.%%) returned %d rows, want 2 (rows: %+v)", len(rows), rows)
	}
	for _, r := range rows {
		if r.Action != "cli.metaread.list" {
			t.Errorf("row action %q does not match filter cli.metaread.%%", r.Action)
		}
	}
}

func TestManager_Show_FilterBySinceUntil(t *testing.T) {
	mgr, v := newTestManager(t)
	// Need distinct timestamps. seedRows uses 1.1s spacing so 3 rows
	// span ~3 seconds. That gives Since a clean dividing point.
	seedRows(t, v, 3)
	// Pick "Since = now - 1.5s" — should match only the last row (the
	// most recent insertion). seedRows ends with a sleep before
	// returning, so "now" is ~1.1s after the last insertion.
	since := time.Now().UTC().Add(-2 * time.Second)
	rows, err := mgr.Show(Filter{Since: since})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if len(rows) < 1 {
		t.Fatalf("Show(Since=now-2s) returned 0 rows; want >= 1 (most recent insert should match)")
	}
	// Every returned row must satisfy Timestamp >= Since (rounded to
	// SQLite's second granularity, so allow a 1-second tolerance).
	for _, r := range rows {
		if r.Timestamp.Before(since.Add(-time.Second)) {
			t.Errorf("row Timestamp %v predates Since %v", r.Timestamp, since)
		}
	}
}

func TestManager_SizeBytes_EmptyVault(t *testing.T) {
	mgr, _ := newTestManager(t)
	n, err := mgr.SizeBytes()
	if err != nil {
		t.Fatalf("SizeBytes: %v", err)
	}
	// Vault.New writes its own audit_log entries during bootstrap
	// (vault.init plus any migration markers). The exact count depends
	// on internal init steps but should be a small positive number.
	// "empty" here means "no F8-style growth"; we just assert the
	// estimate stays under 10 KB, which would be impossible for
	// thousands of rows.
	if n < 0 {
		t.Errorf("SizeBytes returned negative %d", n)
	}
	if n > 10_000 {
		t.Errorf("SizeBytes on empty vault = %d bytes; want < 10000 (vault.init only)", n)
	}
}

func TestManager_SizeBytes_GrowsWithRows(t *testing.T) {
	mgr, v := newTestManager(t)
	before, err := mgr.SizeBytes()
	if err != nil {
		t.Fatalf("SizeBytes before: %v", err)
	}
	// Insert 100 rows; each carries an action + resource. The size
	// estimate should grow by at least 100 * (avg action_len +
	// resource_len + 32) ≈ 100 * 50 = 5000 bytes.
	for i := 0; i < 100; i++ {
		_ = v.AddAuditLog("cli.metaread.list", "SOME_RESOURCE_"+itoa(i), "")
	}
	after, err := mgr.SizeBytes()
	if err != nil {
		t.Fatalf("SizeBytes after: %v", err)
	}
	if diff := after - before; diff < 3000 {
		t.Errorf("SizeBytes grew by %d after 100 inserts; want >= 3000", diff)
	}
}

func TestManager_RowCount(t *testing.T) {
	mgr, v := newTestManager(t)
	before, err := mgr.RowCount()
	if err != nil {
		t.Fatalf("RowCount before: %v", err)
	}
	for i := 0; i < 7; i++ {
		_ = v.AddAuditLog("cli.metaread.list", "r", "")
	}
	after, err := mgr.RowCount()
	if err != nil {
		t.Fatalf("RowCount after: %v", err)
	}
	if diff := after - before; diff != 7 {
		t.Errorf("RowCount delta = %d, want 7", diff)
	}
}

func TestManager_Prune_DeletesOnlyOlder(t *testing.T) {
	mgr, v := newTestManager(t)

	// SQLite's datetime('now') has 1-second resolution. To put two
	// rows strictly older than the cutoff and one row strictly newer,
	// we use generous 3-second gaps between inserts and a 2-second
	// prune window. Timeline (T=0 is first insert):
	//   T=0     -> "old1" rowtime = T0
	//   T=3.1s  -> "old2" rowtime = T0 + 3 (SQLite rounds down)
	//   T=6.2s  -> "recent" rowtime = T0 + 6
	//   Prune cutoff = "recent" insert time minus 2s = T0 + 4
	//   -> old1 (T0) and old2 (T0+3) both strictly < T0+4. Recent stays.
	_ = v.AddAuditLog("cli.metaread.list", "old1", "")
	time.Sleep(3100 * time.Millisecond)
	_ = v.AddAuditLog("cli.metaread.list", "old2", "")
	time.Sleep(3100 * time.Millisecond)
	_ = v.AddAuditLog("cli.metaread.list", "recent", "")

	deleted, err := mgr.Prune(2 * time.Second)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if deleted < 2 {
		t.Errorf("Prune deleted %d rows; want >= 2 (old1, old2 should match)", deleted)
	}

	// The "recent" row must still exist.
	rows, err := mgr.Show(Filter{ActionMatch: "cli.metaread.list"})
	if err != nil {
		t.Fatalf("Show after prune: %v", err)
	}
	found := false
	for _, r := range rows {
		if r.Resource == "recent" {
			found = true
		}
		if r.Resource == "old1" || r.Resource == "old2" {
			t.Errorf("Prune left %q behind; should have been deleted", r.Resource)
		}
	}
	if !found {
		t.Errorf("Prune wiped out the 'recent' row that should have survived")
	}
}

func TestManager_Prune_Idempotent(t *testing.T) {
	mgr, v := newTestManager(t)
	_ = v.AddAuditLog("cli.metaread.list", "a", "")
	// SQLite's datetime('now') has 1-second resolution. Sleep > 1s so
	// the seeded row's timestamp is strictly less than (now - 1s).
	time.Sleep(2100 * time.Millisecond)

	first, err := mgr.Prune(1 * time.Second)
	if err != nil {
		t.Fatalf("first prune: %v", err)
	}
	second, err := mgr.Prune(1 * time.Second)
	if err != nil {
		t.Fatalf("second prune: %v", err)
	}
	if second != 0 {
		t.Errorf("second Prune deleted %d rows; want 0 (idempotent)", second)
	}
	if first == 0 {
		t.Errorf("first Prune deleted 0 rows; expected at least 1 (the seeded 'a' row)")
	}
}

func TestManager_Prune_NonPositive_ReturnsError(t *testing.T) {
	mgr, _ := newTestManager(t)
	for _, d := range []time.Duration{0, -1 * time.Second, -1 * time.Hour} {
		n, err := mgr.Prune(d)
		if err == nil {
			t.Errorf("Prune(%v) returned nil error; want guard against non-positive", d)
		}
		if n != 0 {
			t.Errorf("Prune(%v) returned deleted=%d; want 0 on error path", d, n)
		}
	}
}

func TestManager_CountOlderThan(t *testing.T) {
	mgr, v := newTestManager(t)
	_ = v.AddAuditLog("cli.metaread.list", "old", "")
	// Need 2-second gap so SQLite's second-resolution timestamp on the
	// "old" row is strictly less than (now - 1s).
	time.Sleep(2100 * time.Millisecond)
	_ = v.AddAuditLog("cli.metaread.list", "new", "")

	// Count rows older than 1s — should include the "old" row.
	n, err := mgr.CountOlderThan(1 * time.Second)
	if err != nil {
		t.Fatalf("CountOlderThan: %v", err)
	}
	if n < 1 {
		t.Errorf("CountOlderThan(1s) = %d; want >= 1", n)
	}

	// Count rows older than 10s — should be 0 (nothing is that old in
	// a fresh test vault).
	n, err = mgr.CountOlderThan(10 * time.Second)
	if err != nil {
		t.Fatalf("CountOlderThan (10s): %v", err)
	}
	if n != 0 {
		t.Errorf("CountOlderThan(10s) = %d; want 0", n)
	}
}

func TestManager_Show_LimitCaps(t *testing.T) {
	mgr, v := newTestManager(t)
	for i := 0; i < 10; i++ {
		_ = v.AddAuditLog("cli.metaread.list", "r"+itoa(i), "")
	}
	rows, err := mgr.Show(Filter{Limit: 4})
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if len(rows) != 4 {
		t.Errorf("Show(Limit=4) returned %d rows; want 4", len(rows))
	}
}
