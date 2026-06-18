package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/tene-ai/tene/internal/keychain"
)

// TestFallback_KeychainAvailable_NoNotice — when info.Used == false (the
// happy macOS/Linux native keychain path), emitFallbackNoticeIfNeeded
// must produce zero output and write no sentinel.
func TestFallback_KeychainAvailable_NoNotice(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectDir := t.TempDir()

	var stderr bytes.Buffer
	emitFallbackNoticeIfNeeded(&stderr, keychain.FallbackInfo{Used: false}, projectDir, false)

	if got := stderr.String(); got != "" {
		t.Errorf("expected silent stderr on happy path, got %q", got)
	}

	// Sentinel must not exist.
	sentinel, err := fallbackSentinelPath(projectDir)
	if err != nil {
		t.Fatalf("fallbackSentinelPath: %v", err)
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Errorf("sentinel should not exist on happy path, stat err = %v", err)
	}
}

// TestFallback_FirstCall_PrintsNoticeAndWritesSentinel — fallback active +
// sentinel absent + not quiet => print to stderr AND create the sentinel.
func TestFallback_FirstCall_PrintsNoticeAndWritesSentinel(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectDir := t.TempDir()

	info := keychain.FallbackInfo{
		Used:   true,
		Reason: "keychain_unavailable",
		Path:   filepath.Join(home, ".tene", "keyfile"),
	}

	var stderr bytes.Buffer
	emitFallbackNoticeIfNeeded(&stderr, info, projectDir, false)

	out := stderr.String()
	if !strings.Contains(out, "file-based keystore") {
		t.Errorf("expected stderr to mention file-based keystore, got %q", out)
	}
	if !strings.Contains(out, info.Path) {
		t.Errorf("expected stderr to mention key file path %q, got %q", info.Path, out)
	}
	if !strings.Contains(out, "--quiet to suppress") {
		t.Errorf("expected stderr to mention --quiet hint, got %q", out)
	}

	// Sentinel was created.
	sentinel, err := fallbackSentinelPath(projectDir)
	if err != nil {
		t.Fatalf("fallbackSentinelPath: %v", err)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Errorf("sentinel should exist after notice emit, stat err = %v", err)
	}
}

// TestFallback_SecondCall_NoNotice — sentinel already present, fallback
// still active, not quiet => silent on second run (the one-time guarantee).
func TestFallback_SecondCall_NoNotice(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectDir := t.TempDir()

	info := keychain.FallbackInfo{
		Used:   true,
		Reason: "keychain_unavailable",
		Path:   filepath.Join(home, ".tene", "keyfile"),
	}

	// First call -> creates sentinel.
	var firstStderr bytes.Buffer
	emitFallbackNoticeIfNeeded(&firstStderr, info, projectDir, false)
	if firstStderr.Len() == 0 {
		t.Fatal("preconditions: expected first call to print notice")
	}

	// Second call -> silent.
	var secondStderr bytes.Buffer
	emitFallbackNoticeIfNeeded(&secondStderr, info, projectDir, false)
	if got := secondStderr.String(); got != "" {
		t.Errorf("expected silent stderr on second call, got %q", got)
	}
}

// TestFallback_QuietFlag_NoNoticeNoSentinel — --quiet must suppress the
// notice AND must NOT write the sentinel. Rationale: a subsequent
// non-quiet run still deserves to see the notice once.
func TestFallback_QuietFlag_NoNoticeNoSentinel(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectDir := t.TempDir()

	info := keychain.FallbackInfo{
		Used:   true,
		Reason: "keychain_unavailable",
		Path:   filepath.Join(home, ".tene", "keyfile"),
	}

	// Quiet run -> silent + no sentinel.
	var stderr bytes.Buffer
	emitFallbackNoticeIfNeeded(&stderr, info, projectDir, true)
	if got := stderr.String(); got != "" {
		t.Errorf("expected silent stderr under --quiet, got %q", got)
	}

	sentinel, err := fallbackSentinelPath(projectDir)
	if err != nil {
		t.Fatalf("fallbackSentinelPath: %v", err)
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Errorf("sentinel must not be written under --quiet, stat err = %v", err)
	}

	// And the very next non-quiet run still emits.
	var followup bytes.Buffer
	emitFallbackNoticeIfNeeded(&followup, info, projectDir, false)
	if !strings.Contains(followup.String(), "file-based keystore") {
		t.Errorf("non-quiet follow-up must still print, got %q", followup.String())
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Errorf("sentinel should exist after non-quiet follow-up, stat err = %v", err)
	}
}

// TestFallback_PerProjectIsolation — two different project directories
// share the same HOME but have different sentinels. Notifying for one
// project must not silence the other.
func TestFallback_PerProjectIsolation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	projectA := t.TempDir()
	projectB := t.TempDir()

	info := keychain.FallbackInfo{
		Used:   true,
		Reason: "keychain_unavailable",
		Path:   filepath.Join(home, ".tene", "keyfile"),
	}

	var stderrA bytes.Buffer
	emitFallbackNoticeIfNeeded(&stderrA, info, projectA, false)
	if !strings.Contains(stderrA.String(), "file-based keystore") {
		t.Errorf("project A first call should print, got %q", stderrA.String())
	}

	// Project B still gets its own notice.
	var stderrB bytes.Buffer
	emitFallbackNoticeIfNeeded(&stderrB, info, projectB, false)
	if !strings.Contains(stderrB.String(), "file-based keystore") {
		t.Errorf("project B first call should still print despite project A sentinel, got %q", stderrB.String())
	}

	// Sentinel paths differ -> per-project isolation.
	sA, _ := fallbackSentinelPath(projectA)
	sB, _ := fallbackSentinelPath(projectB)
	if sA == sB {
		t.Errorf("sentinel paths should differ between projects, both = %q", sA)
	}
}

// TestFallback_ConcurrentEmitOnceOnly — multiple goroutines racing on the
// same project should produce exactly one notice and one sentinel
// (O_CREATE|O_EXCL guarantee).
func TestFallback_ConcurrentEmitOnceOnly(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectDir := t.TempDir()

	info := keychain.FallbackInfo{
		Used:   true,
		Reason: "keychain_unavailable",
		Path:   filepath.Join(home, ".tene", "keyfile"),
	}

	const racers = 16
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		buffers = make([]*bytes.Buffer, racers)
	)
	for i := 0; i < racers; i++ {
		buffers[i] = &bytes.Buffer{}
	}
	for i := 0; i < racers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Each goroutine writes to its own buffer to avoid I/O races
			// in the test (we are testing sentinel race, not io.Writer
			// concurrency).
			mu.Lock()
			buf := buffers[i]
			mu.Unlock()
			emitFallbackNoticeIfNeeded(buf, info, projectDir, false)
		}(i)
	}
	wg.Wait()

	printers := 0
	for _, b := range buffers {
		if strings.Contains(b.String(), "file-based keystore") {
			printers++
		}
	}
	if printers != 1 {
		t.Errorf("exactly one racer should win the sentinel race, %d printed", printers)
	}

	// And exactly one sentinel exists.
	sentinel, _ := fallbackSentinelPath(projectDir)
	if _, err := os.Stat(sentinel); err != nil {
		t.Errorf("sentinel should exist after race, stat err = %v", err)
	}
}

// TestFallback_StderrSeparation — notice must go to the io.Writer the
// caller hands in, never to stdout. This is the contract that protects
// JSON consumers parsing os.Stdout from mid-stream noise.
func TestFallback_StderrSeparation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectDir := t.TempDir()

	info := keychain.FallbackInfo{
		Used:   true,
		Reason: "keychain_unavailable",
		Path:   filepath.Join(home, ".tene", "keyfile"),
	}

	// Swap os.Stdout for a pipe so we can detect any accidental write.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	var stderr bytes.Buffer
	emitFallbackNoticeIfNeeded(&stderr, info, projectDir, false)

	// Close write end and drain the pipe.
	_ = w.Close()
	var stdoutBuf bytes.Buffer
	_, _ = stdoutBuf.ReadFrom(r)

	if got := stdoutBuf.String(); got != "" {
		t.Errorf("notice must not write to os.Stdout, got %q", got)
	}
	if !strings.Contains(stderr.String(), "file-based keystore") {
		t.Errorf("notice must reach the provided io.Writer, got %q", stderr.String())
	}
}

// TestFallbackSentinelPath_StableAcrossCalls — the sentinel path for a
// given project must be deterministic so two CLI invocations agree on
// "have we warned this user yet?".
func TestFallbackSentinelPath_StableAcrossCalls(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	projectDir := t.TempDir()

	a, err := fallbackSentinelPath(projectDir)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	b, err := fallbackSentinelPath(projectDir)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if a != b {
		t.Errorf("sentinel path must be stable, got %q vs %q", a, b)
	}
	if !strings.HasPrefix(filepath.Base(a), sentinelPrefix) {
		t.Errorf("sentinel filename should start with %q, got %q", sentinelPrefix, filepath.Base(a))
	}
}
