package keychain

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"
)

// Sprint keychain-probe-fixed.
//
// These tests use go-keyring's MockInit() to swap the OS keyring
// provider for an in-memory map, then assert the contract that
// `NewStoreWithStatus` does NOT touch the per-project service name
// during its availability probe. Before this fix, the probe wrote
// (and then deleted) a value under `tene-<projectHash>` for every
// project directory the user touched, causing macOS to accumulate one
// ACL-registered entry per project.
//
// Note on test ordering: keyring.MockInit() mutates the global
// `provider` variable in the go-keyring package. There is no public
// API to restore the previous provider. These tests are placed in
// their own file so source-code ordering keeps them at the end of
// the package's test sweep; the TestNewStoreWithStatus_EnvOverride
// test in keychain_test.go does not touch the keyring provider
// (env var diverts before any keyring call) so MockInit cannot
// influence it. TestFileStore_* tests also never reach the keyring
// provider. The only risk is that future tests that DO call the
// real OS keyring be added without noticing the mock is sticky —
// the mock-touching tests below should remain the last in the file
// to minimise that risk.

// TestNewStoreWithStatus_ProbeUsesFixedService is the headline
// regression test for the v1.0.14 → v1.0.15 keychain-probe-fixed
// follow-up. It verifies that an availability probe for project A
// followed by an availability probe for project B leaves:
//
//   - zero leftover values in tene-<hashA> service
//   - zero leftover values in tene-<hashB> service
//   - zero leftover value under ProbeServiceName/probe (the probe
//     `Delete` cleaned it up)
//   - FallbackInfo.Used == false (mock provider's Set succeeds, so
//     the OS keychain branch is selected)
//
// Together that proves the per-project service is never written by
// the probe, which is the whole point of the fix.
func TestNewStoreWithStatus_ProbeUsesFixedService(t *testing.T) {
	keyring.MockInit()

	// Belt and suspenders: explicitly clear the env override so the
	// keychain-probe branch is the one actually exercised.
	t.Setenv("TENE_KEYCHAIN_FALLBACK", "")
	t.Setenv("HOME", t.TempDir())

	const projectA = "/Users/example/proj-a"
	const projectB = "/Users/example/proj-b"

	ksA, infoA := NewStoreWithStatus(projectA)
	if infoA.Used {
		t.Fatalf("with mocked keyring, FallbackInfo.Used should be false; got Used=%v Reason=%q",
			infoA.Used, infoA.Reason)
	}
	if _, ok := ksA.(*KeyringStore); !ok {
		t.Fatalf("expected *KeyringStore from happy probe, got %T", ksA)
	}

	ksB, infoB := NewStoreWithStatus(projectB)
	if infoB.Used {
		t.Fatalf("project B probe should also succeed under mock; Reason=%q", infoB.Reason)
	}
	if _, ok := ksB.(*KeyringStore); !ok {
		t.Fatalf("expected *KeyringStore for project B, got %T", ksB)
	}

	// Now the contract: the probe must NOT have left anything on the
	// per-project services. (The probe's Delete cleans up the
	// transient ProbeServiceName value too, so all three slots
	// should read as ErrNotFound.)
	serviceA := ServiceName + "-" + hashPath(projectA)
	serviceB := ServiceName + "-" + hashPath(projectB)

	for _, c := range []struct {
		label   string
		service string
		account string
	}{
		{label: "per-project service A (must NOT be probed)", service: serviceA, account: "keychain-test"},
		{label: "per-project service A (alt probe account)", service: serviceA, account: probeAccount},
		{label: "per-project service B (must NOT be probed)", service: serviceB, account: "keychain-test"},
		{label: "per-project service B (alt probe account)", service: serviceB, account: probeAccount},
		{label: "fixed probe service (must be cleaned up after Delete)", service: ProbeServiceName, account: probeAccount},
	} {
		_, err := keyring.Get(c.service, c.account)
		if !errors.Is(err, keyring.ErrNotFound) {
			t.Errorf("REGRESSION [%s]: expected ErrNotFound for service=%q account=%q, got err=%v",
				c.label, c.service, c.account, err)
		}
	}
}

// TestNewStoreWithStatus_ProbeFailureFallsBackToFile mirrors the
// CI/Docker/headless path. The mock provider is initialised with an
// error so `keyring.Set` always fails, and we expect FallbackInfo to
// report "keychain_unavailable" + the returned KeyStore to be the
// FileStore.
//
// This used to be implicit (covered only by integration on hosts
// without libsecret); pinning it in unit form means the fallback
// behaviour cannot silently regress.
func TestNewStoreWithStatus_ProbeFailureFallsBackToFile(t *testing.T) {
	keyring.MockInitWithError(errors.New("mock probe failure"))
	t.Setenv("TENE_KEYCHAIN_FALLBACK", "")
	t.Setenv("HOME", t.TempDir())

	ks, info := NewStoreWithStatus("/Users/example/proj")
	if !info.Used {
		t.Fatal("probe failure must cause FallbackInfo.Used = true")
	}
	if info.Reason != "keychain_unavailable" {
		t.Errorf("FallbackInfo.Reason = %q, want %q", info.Reason, "keychain_unavailable")
	}
	if _, ok := ks.(*FileStore); !ok {
		t.Errorf("expected *FileStore after probe failure, got %T", ks)
	}
}
