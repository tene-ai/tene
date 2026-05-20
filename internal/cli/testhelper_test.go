package cli

import (
	"bytes"
	"os"
	"testing"
)

// testEnv holds the test environment configuration.
type testEnv struct {
	Dir     string // temp project directory
	HomeDir string // temp HOME directory
	t       *testing.T
}

// setupTestEnv creates a temporary directory and environment variables for testing.
func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	dir := t.TempDir()
	home := t.TempDir()

	t.Setenv("HOME", home)
	t.Setenv("TENE_MASTER_PASSWORD", "testpassword123")

	return &testEnv{Dir: dir, HomeDir: home, t: t}
}

// initVault initializes a test vault.
func (e *testEnv) initVault() {
	e.t.Helper()
	stdout, stderr, err := e.run("init", "test-project", "--no-keychain", "--quiet")
	if err != nil {
		e.t.Fatalf("init failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
}

// resetFlags resets all subcommand-specific flags to defaults.
func resetFlags() {
	// Persistent (global) flags
	flagJSON = false
	flagQuiet = false
	flagEnv = ""
	flagDir = ""
	flagNoColor = false
	flagNoKeychain = false

	// get command flags — default to opting in to plaintext stdout for tests.
	// In production, `tene get` refuses non-TTY stdout by default (U-1), but
	// the test harness swaps os.Stdout for a pipe so every run is non-TTY.
	// The U-1 guard itself is exercised in get_guard_test.go where this is
	// explicitly overridden to `false`.
	flagUnsafeStdout = true

	// set command flags
	setFlagStdin = false
	setFlagOverwrite = false

	// delete command flags
	deleteFlagForce = false

	// export command flags
	exportFlagFile = ""
	exportFlagEncrypted = false

	// import command flags
	importFlagOverwrite = false
	importFlagEncrypted = false

	// update command flag
	updateFlagCheck = false

	// init command flags
	initFlagClaude = false
	initFlagCursor = false
	initFlagWindsurf = false
	initFlagGemini = false
	initFlagCodex = false

	// F8 audit subcommand flags
	auditTailN = 20
	auditShowSince = ""
	auditShowFilter = ""
	auditShowResource = ""
	auditShowLimit = 200
	auditPruneOlderThan = ""
	auditPruneForce = false
	auditPruneDryRun = false
}

// run executes a tene CLI command and returns stdout, stderr, error.
// It captures os.Stdout and os.Stderr since CLI functions use fmt.Print*.
func (e *testEnv) run(args ...string) (string, string, error) {
	// Reset all flags before each run
	resetFlags()

	// Capture real stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	// Set cobra args
	rootCmd.SetArgs(append([]string{"--dir", e.Dir, "--no-keychain"}, args...))

	err := rootCmd.Execute()

	// Restore stdout/stderr and read captured output
	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var stdoutBuf, stderrBuf bytes.Buffer
	_, _ = stdoutBuf.ReadFrom(rOut)
	_, _ = stderrBuf.ReadFrom(rErr)

	return stdoutBuf.String(), stderrBuf.String(), err
}

// runJSON executes in --json mode.
func (e *testEnv) runJSON(args ...string) (string, string, error) {
	allArgs := append([]string{"--json"}, args...)
	return e.run(allArgs...)
}
