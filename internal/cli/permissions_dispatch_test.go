// Tests for the F2 dispatcher integration with the real rootCmd tree.
//
// internal/auth/permissions_test.go covers the table semantics with
// synthetic cobra trees; the tests here exercise the production tree so
// G4 (Permission Tier Coverage) regresses immediately if a new
// AddCommand call lands without a corresponding CommandTier entry.
package cli

import (
	"strings"
	"testing"

	"github.com/agent-kay-it/tene/internal/auth"
	"github.com/spf13/cobra"
)

// TestAllRegisteredCommandsHaveTier — the canonical G4 integration test.
// Walks the production rootCmd subtree (skipping cobra's auto-generated
// help / completion variants the same way auth.Validate does) and
// asserts every leaf has a tier declaration. Forgetting to update
// internal/auth.CommandTier when adding a new subcommand causes the
// test to name the missing path so the diff is obvious.
func TestAllRegisteredCommandsHaveTier(t *testing.T) {
	if err := auth.Validate(rootCmd); err != nil {
		t.Fatalf("rootCmd has commands missing tier declarations: %v", err)
	}
}

// TestCommandTierPath_TopLevel — `tene list` flattens to the bare verb
// "list" so the lookup key matches the CommandTier map.
func TestCommandTierPath_TopLevel(t *testing.T) {
	listCmd, _, err := rootCmd.Find([]string{"list"})
	if err != nil {
		t.Fatalf("rootCmd.Find(list) = %v", err)
	}
	got := commandTierPath(listCmd)
	if got != "list" {
		t.Errorf("commandTierPath(listCmd) = %q, want \"list\"", got)
	}
}

// TestCommandTierPath_Subcommand — `tene env list` must produce
// "env list" (space-joined) so the lookup key matches CommandTier.
func TestCommandTierPath_Subcommand(t *testing.T) {
	envListCmd, _, err := rootCmd.Find([]string{"env", "list"})
	if err != nil {
		t.Fatalf("rootCmd.Find(env list) = %v", err)
	}
	got := commandTierPath(envListCmd)
	if got != "env list" {
		t.Errorf("commandTierPath(env list) = %q, want \"env list\"", got)
	}
}

// TestCommandTierPath_RootBare — the bare root invocation returns the
// empty string so the PreRunE knows to skip the tier check entirely
// (cobra prints help and never enters a RunE in this case).
func TestCommandTierPath_RootBare(t *testing.T) {
	if got := commandTierPath(rootCmd); got != "" {
		t.Errorf("commandTierPath(rootCmd) = %q, want empty string", got)
	}
}

// TestPersistentPreRunE_KnownVerbPasses — invoking the dispatcher hook
// directly with a known verb returns nil (no error), even if the verb
// itself has not been executed. This isolates the hook's policy logic
// from the verb's RunE side effects.
func TestPersistentPreRunE_KnownVerbPasses(t *testing.T) {
	listCmd, _, err := rootCmd.Find([]string{"list"})
	if err != nil {
		t.Fatalf("rootCmd.Find(list) = %v", err)
	}
	if err := rootPersistentPreRunE(listCmd, nil); err != nil {
		t.Errorf("rootPersistentPreRunE(list) = %v, want nil", err)
	}
}

// TestPersistentPreRunE_UnknownVerbFails — the runtime half of G4: even
// if a developer somehow registers a command without updating
// CommandTier (e.g. via AddCommand from a different init function that
// ran after auth.Validate), the next invocation of that verb is
// refused with an actionable error message.
func TestPersistentPreRunE_UnknownVerbFails(t *testing.T) {
	// Synthesize a verb that exists in cobra but NOT in CommandTier.
	// We don't AddCommand it to rootCmd (which would break the Validate
	// invariant); instead we manually set its parent so CommandPath()
	// resolves correctly relative to the real root.
	rogue := &cobra.Command{Use: "f2-test-rogue-verb"}
	rootCmd.AddCommand(rogue)
	t.Cleanup(func() { rootCmd.RemoveCommand(rogue) })

	err := rootPersistentPreRunE(rogue, nil)
	if err == nil {
		t.Fatal("rootPersistentPreRunE(undeclared verb) = nil, want error")
	}
	if !strings.Contains(err.Error(), "f2-test-rogue-verb") {
		t.Errorf("error %q does not name the offending verb", err.Error())
	}
	if !strings.Contains(err.Error(), "CommandTier") {
		t.Errorf("error %q does not mention CommandTier — the operator hint is missing", err.Error())
	}
}

// TestPersistentPreRunE_RootBarePasses — calling the hook with the bare
// rootCmd (e.g. `tene` with no args) must NOT fail; cobra dispatches to
// help text in that case.
func TestPersistentPreRunE_RootBarePasses(t *testing.T) {
	if err := rootPersistentPreRunE(rootCmd, nil); err != nil {
		t.Errorf("rootPersistentPreRunE(rootCmd) = %v, want nil", err)
	}
}

// TestPersistentPreRunE_HelpSubcommandSkipped — sprint v1014-rc1-qa-fixes
// FX4 (B4 regression test). Cobra's auto-generated "help" command must
// pass through rootPersistentPreRunE without triggering the
// "no PermLevel entry" error. The synthetic help command is fabricated
// the same way auth.TestValidate_IgnoresHelpCommand does it: call
// InitDefaultHelpCmd, find the command, invoke the hook directly.
func TestPersistentPreRunE_HelpSubcommandSkipped(t *testing.T) {
	rootCmd.InitDefaultHelpCmd()
	helpCmd, _, err := rootCmd.Find([]string{"help"})
	if err != nil {
		t.Fatalf("rootCmd.Find(help) = %v", err)
	}
	if helpCmd.Name() != "help" {
		t.Fatalf("expected to resolve cobra's synthetic help command, got %q", helpCmd.Name())
	}
	if err := rootPersistentPreRunE(helpCmd, nil); err != nil {
		t.Errorf("rootPersistentPreRunE(help) = %v, want nil — B4 regression", err)
	}
}

// TestPersistentPreRunE_CompleteSubcommandSkipped — same shape as the
// help test but for cobra's __complete shell-completion helper. We can
// fabricate it directly because cobra exposes the name; we attach it
// to rootCmd so CommandPath() resolves and clean up afterwards.
func TestPersistentPreRunE_CompleteSubcommandSkipped(t *testing.T) {
	complete := &cobra.Command{Use: "__complete", Run: func(*cobra.Command, []string) {}}
	rootCmd.AddCommand(complete)
	t.Cleanup(func() { rootCmd.RemoveCommand(complete) })

	if err := rootPersistentPreRunE(complete, nil); err != nil {
		t.Errorf("rootPersistentPreRunE(__complete) = %v, want nil", err)
	}
}

// TestProductionTree_LogoutIsAbsent — pins B5: the permissions table
// the binary publishes must not include `logout` until the cloud
// feature is re-registered. A future PR that re-adds the cloud verb
// must update both CommandTier and rootCmd, and this test will
// continue to pass at that point because Find(logout) will resolve.
//
// For today's build, Find returns a non-nil error because the command
// is not registered.
func TestProductionTree_LogoutIsAbsent(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"logout"})
	// cobra.Find returns the closest ancestor (rootCmd) and a "no such
	// command" error when the leaf is missing. Either condition is
	// acceptable; what we are guarding against is the rc1 state where
	// Find found a real command for "logout" without any matching tier.
	if err == nil && cmd != rootCmd && cmd.Name() == "logout" {
		t.Fatal("logout is registered but FX4 expected it to remain unregistered until cloud is re-enabled")
	}
}
