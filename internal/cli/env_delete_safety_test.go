package cli

import (
	"strings"
	"testing"
)

// Sprint v1014-rc1-qa-fixes / FX2.
//
// These tests pin invariant I-12: destructive ops (`tene delete KEY`,
// `tene env delete`) refuse to proceed on a non-interactive shell unless
// --force is passed explicitly. The QA reproduction (B2) typed
// `tene env delete testdel` in a piped shell and saw "Environment
// 'testdel' deleted (1 secrets removed)." with no prompt. The tests in
// this file lock the new fail-closed behaviour so the same one-character
// regression cannot happen again.

// TestEnvDelete_RefusesWithoutForceOnNonTTY is the headline B2
// regression test. It populates an environment, attempts a delete
// without --force from the test harness (which runs without a TTY),
// and verifies the env is still present and the secret is intact.
func TestEnvDelete_RefusesWithoutForceOnNonTTY(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()

	if _, _, err := e.run("env", "create", "stagingx"); err != nil {
		t.Fatalf("env create: %v", err)
	}
	if _, _, err := e.run("set", "CANARY", "value-1", "--env", "stagingx"); err != nil {
		t.Fatalf("set: %v", err)
	}

	stdout, stderr, err := e.run("env", "delete", "stagingx")
	if err == nil {
		t.Fatalf("B2 REGRESSION: env delete with no --force succeeded on non-TTY.\nstdout: %s\nstderr: %s",
			stdout, stderr)
	}
	combined := strings.ToLower(stdout + stderr)
	if !strings.Contains(combined, "refusing") &&
		!strings.Contains(combined, "non-interactive") &&
		!strings.Contains(combined, "cancelled") {
		t.Logf("note: refusal message wording: stderr=%s stdout=%s", stderr, stdout)
	}

	// env must still exist + secret must still be there.
	stdout, _, err = e.run("env", "list", "--json")
	if err != nil {
		t.Fatalf("env list: %v", err)
	}
	if !strings.Contains(stdout, "stagingx") {
		t.Fatalf("env stagingx should still exist after refused delete, got:\n%s", stdout)
	}

	stdout, _, err = e.run("list", "--env", "stagingx", "--json")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(stdout, "CANARY") {
		t.Fatalf("CANARY should still exist in stagingx after refused delete, got:\n%s", stdout)
	}
}

// TestEnvDelete_SucceedsWithForce confirms the documented escape hatch
// works: a CI/CD pipeline that genuinely wants an unattended env delete
// passes --force and the delete proceeds.
func TestEnvDelete_SucceedsWithForce(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()

	if _, _, err := e.run("env", "create", "stagingy"); err != nil {
		t.Fatalf("env create: %v", err)
	}
	if _, _, err := e.run("set", "CANARY", "v", "--env", "stagingy"); err != nil {
		t.Fatalf("set: %v", err)
	}

	stdout, stderr, err := e.run("env", "delete", "stagingy", "--force")
	if err != nil {
		t.Fatalf("env delete --force should succeed: %v\nstdout: %s\nstderr: %s",
			err, stdout, stderr)
	}

	// confirm gone
	stdout, _, _ = e.run("env", "list", "--json")
	if strings.Contains(stdout, "stagingy") {
		t.Fatalf("env stagingy should be gone after --force delete, got:\n%s", stdout)
	}
}

// TestEnvDelete_ProtectedDefaultStillProtected verifies the existing
// guards (cannot delete "default", cannot delete the active env) are
// preserved when the --force flag is present. --force only suppresses
// the confirm prompt, not the structural protections.
func TestEnvDelete_ProtectedDefaultStillProtected(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()

	stdout, stderr, err := e.run("env", "delete", "default", "--force")
	if err == nil {
		t.Fatalf("delete of default env must always fail.\nstdout: %s\nstderr: %s", stdout, stderr)
	}
}

// TestEnvDelete_ActiveEnvStillProtected mirrors the above for the
// "cannot delete the active environment" check.
func TestEnvDelete_ActiveEnvStillProtected(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()

	// switch to a non-default env so we have one we can target
	if _, _, err := e.run("env", "create", "active1"); err != nil {
		t.Fatalf("env create: %v", err)
	}
	if _, _, err := e.run("env", "active1"); err != nil {
		t.Fatalf("env switch: %v", err)
	}

	stdout, stderr, err := e.run("env", "delete", "active1", "--force")
	if err == nil {
		t.Fatalf("delete of active env must always fail.\nstdout: %s\nstderr: %s", stdout, stderr)
	}
}

// TestDelete_RefusesWithoutForceOnNonTTY confirms the same I-12 stance
// applies to `tene delete KEY` for single-secret deletion. The shared
// promptConfirm helper makes this come for free, but the test pins the
// expected behaviour so a future refactor that splits the helpers cannot
// silently regress one verb while keeping the other safe.
func TestDelete_RefusesWithoutForceOnNonTTY(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()

	if _, _, err := e.run("set", "PROBE", "secret-value"); err != nil {
		t.Fatalf("set: %v", err)
	}

	stdout, stderr, err := e.run("delete", "PROBE")
	if err == nil {
		t.Fatalf("single-secret delete with no --force must refuse on non-TTY.\nstdout: %s\nstderr: %s",
			stdout, stderr)
	}

	// PROBE should still exist
	stdout, _, _ = e.run("list", "--json")
	if !strings.Contains(stdout, "PROBE") {
		t.Fatalf("PROBE should still exist after refused delete, got:\n%s", stdout)
	}
}

// TestDelete_SucceedsWithForce locks the symmetric escape hatch for
// single-secret deletion.
func TestDelete_SucceedsWithForce(t *testing.T) {
	e := setupTestEnv(t)
	e.initVault()

	if _, _, err := e.run("set", "PROBE2", "v"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if _, _, err := e.run("delete", "PROBE2", "--force"); err != nil {
		t.Fatalf("delete --force should succeed: %v", err)
	}
	stdout, _, _ := e.run("list", "--json")
	if strings.Contains(stdout, "PROBE2") {
		t.Fatalf("PROBE2 should be gone after --force delete, got:\n%s", stdout)
	}
}
