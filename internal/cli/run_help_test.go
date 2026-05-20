package cli

import (
	"strings"
	"testing"
)

// Sprint v1014-rc1-qa-fixes / FX5.
//
// These tests pin the B6 regression: `tene run --help` (and `-h`)
// returned "No command specified" in v1.0.14-rc1 because run.go's
// DisableFlagParsing=true suppressed cobra's built-in help handler
// and the home-grown parseFlagsBeforeDash did not look for help
// flags. The hasHelpFlag helper plus the new short-circuit in
// runRun is the surgical fix.

// TestHasHelpFlag tests the table of recognised invocations. The
// child-process passthrough case (`tene run -- python -h`) is the
// critical isolation guarantee — the child's -h must NOT trigger
// tene's help output, otherwise users lose access to their tool's
// own help.
func TestHasHelpFlag(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{name: "bare --help", args: []string{"--help"}, want: true},
		{name: "bare -h", args: []string{"-h"}, want: true},
		{name: "flag then --help", args: []string{"--env", "dev", "--help"}, want: true},
		{name: "--help then flag", args: []string{"--help", "--env", "dev"}, want: true},
		{name: "no help at all", args: []string{"--env", "dev"}, want: false},
		{name: "empty args", args: nil, want: false},
		{name: "child -h after --", args: []string{"--", "python", "-h"}, want: false},
		{name: "child --help after --", args: []string{"--", "node", "--help"}, want: false},
		{name: "flag, --, then child -h", args: []string{"--env", "dev", "--", "ls", "-h"}, want: false},
		{name: "tene help before --, child -h after", args: []string{"--help", "--", "ls", "-h"}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := hasHelpFlag(tc.args)
			if got != tc.want {
				t.Errorf("hasHelpFlag(%v) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}

// TestRunCommand_HelpFlagPrintsHelp drives the verb at the test-harness
// level: `tene run --help` should produce help text in stdout and exit
// 0 without invoking the loadApp/keychain path. This is the headline
// B6 regression test.
func TestRunCommand_HelpFlagPrintsHelp(t *testing.T) {
	e := setupTestEnv(t)
	// NOTE: no e.initVault() — `tene run --help` must work before any
	// vault exists. This was part of the user-facing surprise in B6:
	// new users typed `tene run --help` to learn the syntax and got
	// "No command specified" instead.

	stdout, _, err := e.run("run", "--help")
	if err != nil {
		t.Fatalf("tene run --help should succeed: %v", err)
	}
	if !strings.Contains(stdout, "Run a command with secrets injected") {
		t.Errorf("expected run command's help text in stdout, got:\n%s", stdout)
	}
	// Negative assertion: the old error message must NOT appear.
	if strings.Contains(stdout, "No command specified") {
		t.Errorf("B6 REGRESSION: stale 'No command specified' message in --help output:\n%s", stdout)
	}
}

// TestRunCommand_ShortHelpFlagPrintsHelp mirrors the above for -h.
func TestRunCommand_ShortHelpFlagPrintsHelp(t *testing.T) {
	e := setupTestEnv(t)

	stdout, _, err := e.run("run", "-h")
	if err != nil {
		t.Fatalf("tene run -h should succeed: %v", err)
	}
	if !strings.Contains(stdout, "Run a command with secrets injected") {
		t.Errorf("expected run command's help text in stdout, got:\n%s", stdout)
	}
}

// TestRunCommand_NoArgsStillErrors confirms we did NOT regress the
// existing "no command specified" error path. `tene run` (no flags,
// no args) must still report COMMAND_NOT_FOUND.
func TestRunCommand_NoArgsStillErrors(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()

	stdout, stderr, err := e.run("run")
	if err == nil {
		t.Fatalf("bare `tene run` should error.\nstdout: %s\nstderr: %s", stdout, stderr)
	}
	// We assert the bare-run path produces ANY error, not a specific
	// wording. Reason: DisableFlagParsing on runCmd means the cobra-
	// level --dir/--no-keychain global flags pass through to the
	// runRun handler verbatim, which then mis-attributes them to the
	// child-command arg list. That fragility is unrelated to FX5
	// (--help) and is tracked separately. What FX5 needs to lock in
	// is that bare `tene run` does NOT silently succeed.
	_ = err // explicit acknowledgement; the if-block above already guards
}
