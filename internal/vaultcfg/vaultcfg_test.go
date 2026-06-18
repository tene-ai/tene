package vaultcfg

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/tene-ai/tene/internal/vault"
)

func tempVault(t *testing.T) *vault.Vault {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, ".tene", "vault.db")
	v, err := vault.New(dbPath)
	if err != nil {
		t.Fatalf("vault.New: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })
	return v
}

func TestGet_Defaults_OnUnsetVault(t *testing.T) {
	v := tempVault(t)
	cases := []struct {
		key, want string
	}{
		{KeyPreviewEnabled, "true"},
		{KeyPreviewFront, "0"},
		{KeyPreviewBack, "4"},
		{KeyAuditWarnAtMB, "50"},
	}
	for _, tc := range cases {
		got, err := Get(v, tc.key)
		if err != nil {
			t.Errorf("Get(%q): %v", tc.key, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Get(%q) = %q, want %q", tc.key, got, tc.want)
		}
	}
}

func TestGet_UnknownKey_ReturnsSentinel(t *testing.T) {
	v := tempVault(t)
	_, err := Get(v, "config.nope")
	if !errors.Is(err, ErrInvalidConfigKey) {
		t.Errorf("Get(unknown): %v, want ErrInvalidConfigKey", err)
	}
}

func TestSet_ValidatesAndPersists(t *testing.T) {
	v := tempVault(t)

	// preview.enabled: accept canonical and aliases.
	for _, in := range []string{"false", "no", "off", "0", "False"} {
		if err := Set(v, KeyPreviewEnabled, in); err != nil {
			t.Errorf("Set(preview.enabled=%q): %v", in, err)
		}
	}
	got, _ := Get(v, KeyPreviewEnabled)
	if got != "false" {
		t.Errorf("after multiple sets preview.enabled = %q, want %q", got, "false")
	}

	// preview.back: valid 0-8.
	if err := Set(v, KeyPreviewBack, "5"); err != nil {
		t.Fatalf("Set(preview.back=5): %v", err)
	}
	got, _ = Get(v, KeyPreviewBack)
	if got != "5" {
		t.Errorf("preview.back = %q, want %q", got, "5")
	}

	// audit.warnAtMB: valid in range.
	if err := Set(v, KeyAuditWarnAtMB, "100"); err != nil {
		t.Fatalf("Set(audit.warnAtMB=100): %v", err)
	}
	if got := GetAuditWarnAtMB(v); got != 100 {
		t.Errorf("GetAuditWarnAtMB = %d, want 100", got)
	}
}

func TestSet_RejectsOutOfRange(t *testing.T) {
	v := tempVault(t)
	cases := []struct {
		key, value string
	}{
		{KeyPreviewFront, "-1"},
		{KeyPreviewFront, "9"},
		{KeyPreviewFront, "abc"},
		{KeyPreviewBack, "-1"},
		{KeyPreviewBack, "9"},
		{KeyAuditWarnAtMB, "0"},
		{KeyAuditWarnAtMB, "1001"},
		{KeyAuditWarnAtMB, "nope"},
		{KeyPreviewEnabled, "maybe"},
	}
	for _, tc := range cases {
		err := Set(v, tc.key, tc.value)
		if !errors.Is(err, ErrInvalidConfigValue) {
			t.Errorf("Set(%q=%q): %v, want ErrInvalidConfigValue", tc.key, tc.value, err)
		}
	}
}

func TestSet_RejectsCombinedCapViolation(t *testing.T) {
	v := tempVault(t)

	// Establish front=4 first.
	if err := Set(v, KeyPreviewFront, "4"); err != nil {
		t.Fatalf("seed front=4: %v", err)
	}

	// back=5 would sum to 9, exceeding the hard cap.
	err := Set(v, KeyPreviewBack, "5")
	if !errors.Is(err, ErrInvalidConfigValue) {
		t.Errorf("Set(preview.back=5) when front=4: %v, want ErrInvalidConfigValue", err)
	}

	// back=4 keeps the sum at the cap and is fine.
	if err := Set(v, KeyPreviewBack, "4"); err != nil {
		t.Errorf("Set(preview.back=4) when front=4: %v, want nil", err)
	}

	// Now front=5 would push sum to 9 -> rejected.
	err = Set(v, KeyPreviewFront, "5")
	if !errors.Is(err, ErrInvalidConfigValue) {
		t.Errorf("Set(preview.front=5) when back=4: %v, want ErrInvalidConfigValue", err)
	}
}

func TestSet_UnknownKey(t *testing.T) {
	v := tempVault(t)
	err := Set(v, "preview.unknown", "true")
	if !errors.Is(err, ErrInvalidConfigKey) {
		t.Errorf("Set(unknown): %v, want ErrInvalidConfigKey", err)
	}
}

func TestGetPreviewSettings_DefaultsAndRoundTrip(t *testing.T) {
	v := tempVault(t)

	ps, err := GetPreviewSettings(v)
	if err != nil {
		t.Fatalf("GetPreviewSettings: %v", err)
	}
	if !ps.Enabled || ps.Front != 0 || ps.Back != 4 {
		t.Errorf("defaults = %+v, want {Enabled:true, Front:0, Back:4}", ps)
	}

	// Round-trip: disable + adjust front/back.
	_ = Set(v, KeyPreviewEnabled, "false")
	_ = Set(v, KeyPreviewBack, "2")
	_ = Set(v, KeyPreviewFront, "2")

	ps, _ = GetPreviewSettings(v)
	if ps.Enabled || ps.Front != 2 || ps.Back != 2 {
		t.Errorf("after set = %+v, want {Enabled:false, Front:2, Back:2}", ps)
	}
}

func TestGetPreviewSettings_CorruptionFallsBackToDefaults(t *testing.T) {
	// Defense-in-depth: directly poke malformed values into vault_meta
	// (simulating disk corruption or a bad downgrade) and verify the
	// loader collapses to safe defaults rather than passing them to
	// DerivePreview.
	v := tempVault(t)
	_ = v.SetMeta(KeyPreviewFront, "999")
	_ = v.SetMeta(KeyPreviewBack, "999")

	ps, err := GetPreviewSettings(v)
	if err != nil {
		t.Fatalf("GetPreviewSettings: %v", err)
	}
	if ps.Front != DefaultPreviewFront || ps.Back != DefaultPreviewBack {
		t.Errorf("corruption fallback = %+v, want defaults %d/%d",
			ps, DefaultPreviewFront, DefaultPreviewBack)
	}
}

func TestGetAuditWarnAtMB_FallbackOnCorruption(t *testing.T) {
	v := tempVault(t)
	_ = v.SetMeta(KeyAuditWarnAtMB, "garbage")
	if got := GetAuditWarnAtMB(v); got != DefaultAuditWarnAtMB {
		t.Errorf("corruption fallback = %d, want %d", got, DefaultAuditWarnAtMB)
	}

	_ = v.SetMeta(KeyAuditWarnAtMB, "0") // out of range
	if got := GetAuditWarnAtMB(v); got != DefaultAuditWarnAtMB {
		t.Errorf("out-of-range fallback = %d, want %d", got, DefaultAuditWarnAtMB)
	}
}

func TestCurrentPreviewFront_RawRead(t *testing.T) {
	v := tempVault(t)

	got, err := CurrentPreviewFront(v)
	if err != nil {
		t.Fatalf("CurrentPreviewFront: %v", err)
	}
	if got != DefaultPreviewFront {
		t.Errorf("default = %d, want %d", got, DefaultPreviewFront)
	}

	_ = Set(v, KeyPreviewFront, "4")
	got, _ = CurrentPreviewFront(v)
	if got != 4 {
		t.Errorf("after set = %d, want 4", got)
	}
}

func TestKnownKeys_StableOrder(t *testing.T) {
	want := []string{
		KeyPreviewEnabled,
		KeyPreviewFront,
		KeyPreviewBack,
		KeyAuditWarnAtMB,
	}
	got := KnownKeys()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestIsKnown(t *testing.T) {
	if !IsKnown(KeyPreviewEnabled) {
		t.Errorf("IsKnown(KeyPreviewEnabled) = false")
	}
	if IsKnown("preview.enabled") { // missing config. prefix
		t.Errorf("IsKnown(unprefixed) = true")
	}
	if IsKnown("config.bogus") {
		t.Errorf("IsKnown(bogus) = true")
	}
}
