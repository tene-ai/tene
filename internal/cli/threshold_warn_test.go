// Tests for F8's audit-log threshold notice (threshold_warn.go).
//
// These tests exercise maybeEmitAuditThresholdWarning directly with
// a writer + a controllable vault, so the sentinel state-machine
// (emit vs suppress vs refresh) is verifiable without going through
// the cobra dispatcher.
//
// Each test uses t.TempDir() for both HOME and the vault path so the
// sentinel under ~/.tene/.audit-warned-<hash> is isolated to that
// test and never collides with the user's real ~/.tene.
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tene-ai/tene/internal/vault"
	"github.com/tene-ai/tene/internal/vaultcfg"
)

// openTestVault constructs a vault.Vault under dir/.tene/vault.db and
// returns it (closed via t.Cleanup). Sets HOME to a separate temp dir
// so any sentinel writes land in an isolated location.
func openTestVault(t *testing.T) (*vault.Vault, string) {
	t.Helper()
	dir := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)

	vaultPath := filepath.Join(dir, ".tene", "vault.db")
	v, err := vault.New(vaultPath)
	if err != nil {
		t.Fatalf("vault.New: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })
	return v, dir
}

// fillAuditLogToBytes inserts dummy rows until SizeBytes >= target.
// Each row contributes ~80 bytes (per-row constant 32 + action/
// resource lengths) so reaching 1 MB requires ~13k rows; cap by
// row count too so a buggy condition never spins forever.
func fillAuditLogToBytes(t *testing.T, v *vault.Vault, target int64) {
	t.Helper()
	const cap = 200_000 // hard upper bound; way past any realistic test target
	for i := 0; i < cap; i++ {
		// Long resource string to amortise the per-insert SQL cost.
		// 200 bytes per row makes target=1MB reachable in ~5k inserts.
		_ = v.AddAuditLog(
			"cli.metaread.test",
			strings.Repeat("R", 200),
			"",
		)
		if i%500 == 0 {
			n, _ := v.GetAuditLogSize()
			if n >= target {
				return
			}
		}
	}
	n, _ := v.GetAuditLogSize()
	if n < target {
		t.Fatalf("failed to fill audit_log to %d bytes after %d inserts (got %d)", target, cap, n)
	}
}

func TestAuditWarn_BelowThreshold_NoEmission(t *testing.T) {
	v, projectDir := openTestVault(t)
	// Default threshold is 50 MB. A fresh vault has only the
	// vault.init audit row — well under 50 MB.
	var buf bytes.Buffer
	maybeEmitAuditThresholdWarning(&buf, v, projectDir, false)
	if buf.Len() != 0 {
		t.Errorf("expected no notice when size < threshold; got %q", buf.String())
	}
}

func TestAuditWarn_QuietFlag_Suppresses(t *testing.T) {
	v, projectDir := openTestVault(t)
	// Lower the threshold to 1 MB, then fill to 2 MB so the size
	// check would normally fire.
	if err := vaultcfg.Set(v, vaultcfg.KeyAuditWarnAtMB, "1"); err != nil {
		t.Fatalf("set warnAtMB=1: %v", err)
	}
	fillAuditLogToBytes(t, v, 2*1024*1024)

	var buf bytes.Buffer
	maybeEmitAuditThresholdWarning(&buf, v, projectDir, true /* quiet */)
	if buf.Len() != 0 {
		t.Errorf("--quiet should suppress threshold notice; got %q", buf.String())
	}

	// Sentinel must NOT be written when quiet is set, so the next
	// non-quiet run still gets the notice.
	sentinel, _ := auditSentinelPath(projectDir)
	if _, err := os.Stat(sentinel); err == nil {
		t.Errorf("quiet path wrote sentinel %s; should not", sentinel)
	}
}

func TestAuditWarn_FirstTime_EmitsAndWritesSentinel(t *testing.T) {
	v, projectDir := openTestVault(t)
	if err := vaultcfg.Set(v, vaultcfg.KeyAuditWarnAtMB, "1"); err != nil {
		t.Fatalf("set warnAtMB=1: %v", err)
	}
	fillAuditLogToBytes(t, v, 2*1024*1024)

	var buf bytes.Buffer
	maybeEmitAuditThresholdWarning(&buf, v, projectDir, false)

	got := buf.String()
	if got == "" {
		t.Fatal("expected threshold notice on first call over threshold; got empty")
	}
	if !strings.Contains(got, "audit log is ~") {
		t.Errorf("notice missing 'audit log is ~' prefix: %q", got)
	}
	if !strings.Contains(got, "tene audit prune") {
		t.Errorf("notice missing the prune-command hint: %q", got)
	}

	// Sentinel must exist after first emission.
	sentinel, err := auditSentinelPath(projectDir)
	if err != nil {
		t.Fatalf("auditSentinelPath: %v", err)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Errorf("sentinel %s missing after first emission: %v", sentinel, err)
	}
}

func TestAuditWarn_SentinelFresh_NoEmission(t *testing.T) {
	v, projectDir := openTestVault(t)
	if err := vaultcfg.Set(v, vaultcfg.KeyAuditWarnAtMB, "1"); err != nil {
		t.Fatalf("set warnAtMB=1: %v", err)
	}
	fillAuditLogToBytes(t, v, 2*1024*1024)

	// First call: emits + writes sentinel.
	var first bytes.Buffer
	maybeEmitAuditThresholdWarning(&first, v, projectDir, false)
	if first.Len() == 0 {
		t.Fatal("first call did not emit; sentinel-fresh test cannot proceed")
	}

	// Second call: sentinel is fresh (just touched), so no emission.
	var second bytes.Buffer
	maybeEmitAuditThresholdWarning(&second, v, projectDir, false)
	if second.Len() != 0 {
		t.Errorf("second call emitted despite fresh sentinel; got %q", second.String())
	}
}

func TestAuditWarn_SentinelStale25h_EmitsAgainAndRefreshes(t *testing.T) {
	v, projectDir := openTestVault(t)
	if err := vaultcfg.Set(v, vaultcfg.KeyAuditWarnAtMB, "1"); err != nil {
		t.Fatalf("set warnAtMB=1: %v", err)
	}
	fillAuditLogToBytes(t, v, 2*1024*1024)

	// Pre-create the sentinel with an mtime 25 hours ago to simulate
	// a stale marker. Subsequent emission should fire again and bump
	// the mtime to ~now.
	sentinel, err := auditSentinelPath(projectDir)
	if err != nil {
		t.Fatalf("auditSentinelPath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(sentinel), 0o700); err != nil {
		t.Fatalf("mkdir sentinel parent: %v", err)
	}
	if err := os.WriteFile(sentinel, nil, 0o600); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}
	staleTime := time.Now().Add(-25 * time.Hour)
	if err := os.Chtimes(sentinel, staleTime, staleTime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	// Sanity: sentinel exists and is stale.
	st0, _ := os.Stat(sentinel)
	if time.Since(st0.ModTime()) < auditWarnWindow {
		t.Fatalf("test setup: sentinel mtime %v is not stale enough (window %v)",
			st0.ModTime(), auditWarnWindow)
	}

	var buf bytes.Buffer
	maybeEmitAuditThresholdWarning(&buf, v, projectDir, false)
	if buf.Len() == 0 {
		t.Errorf("stale sentinel should trigger re-emission; got empty")
	}

	// Sentinel mtime must be bumped to recent.
	st1, err := os.Stat(sentinel)
	if err != nil {
		t.Fatalf("stat sentinel after emission: %v", err)
	}
	if time.Since(st1.ModTime()) >= auditWarnWindow {
		t.Errorf("sentinel mtime still stale after emission: %v", st1.ModTime())
	}
}

func TestAuditWarn_SentinelWriteFailure_DoesNotBlockCommand(t *testing.T) {
	v, projectDir := openTestVault(t)
	if err := vaultcfg.Set(v, vaultcfg.KeyAuditWarnAtMB, "1"); err != nil {
		t.Fatalf("set warnAtMB=1: %v", err)
	}
	fillAuditLogToBytes(t, v, 2*1024*1024)

	// Simulate sentinel write failure by pre-creating the sentinel
	// path as a DIRECTORY (so OpenFile O_WRONLY fails). The function
	// must still emit the notice and return normally — never panic
	// or abort the calling command.
	sentinel, err := auditSentinelPath(projectDir)
	if err != nil {
		t.Fatalf("auditSentinelPath: %v", err)
	}
	if err := os.MkdirAll(sentinel, 0o700); err != nil {
		t.Fatalf("create sentinel as dir: %v", err)
	}

	var buf bytes.Buffer
	// Should not panic; should emit the notice.
	maybeEmitAuditThresholdWarning(&buf, v, projectDir, false)
	if buf.Len() == 0 {
		t.Errorf("expected notice even when sentinel write fails; got empty")
	}
}

func TestAuditWarn_NilVault_NoOp(t *testing.T) {
	var buf bytes.Buffer
	// Must not panic. The dispatcher path explicitly guards via
	// emitAuditThresholdHook (vault.New failure -> early return), so
	// we double-cover here.
	maybeEmitAuditThresholdWarning(&buf, nil, t.TempDir(), false)
	if buf.Len() != 0 {
		t.Errorf("nil vault produced output: %q", buf.String())
	}
}

func TestResetAuditSentinel_RemovesFile(t *testing.T) {
	_, projectDir := openTestVault(t)
	sentinel, err := auditSentinelPath(projectDir)
	if err != nil {
		t.Fatalf("auditSentinelPath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(sentinel), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(sentinel, []byte(""), 0o600); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	resetAuditSentinel(projectDir)

	if _, err := os.Stat(sentinel); err == nil {
		t.Errorf("resetAuditSentinel did not remove %s", sentinel)
	}
}
