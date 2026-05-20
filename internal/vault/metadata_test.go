package vault

import (
	"errors"
	"testing"
)

func TestListSecretMetadata_EmptyEnv_ReturnsEmptySlice(t *testing.T) {
	v := tempVault(t)
	got, err := v.ListSecretMetadata("default")
	if err != nil {
		t.Fatalf("ListSecretMetadata: %v", err)
	}
	if got == nil {
		t.Fatalf("ListSecretMetadata returned nil slice; want empty slice for deterministic JSON")
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestListSecretMetadata_PopulatedRows(t *testing.T) {
	v := tempVault(t)
	if err := v.SetSecretWithPreview("STRIPE_KEY", "ct1", "default", "…Bc1D"); err != nil {
		t.Fatalf("SetSecretWithPreview: %v", err)
	}
	if err := v.SetSecretWithPreview("OPENAI_KEY", "ct2", "default", "…aBcD"); err != nil {
		t.Fatalf("SetSecretWithPreview: %v", err)
	}
	if err := v.SetSecretWithPreview("DB_PASS", "ct3", "prod", "…dEF1"); err != nil {
		t.Fatalf("SetSecretWithPreview: %v", err)
	}

	got, err := v.ListSecretMetadata("default")
	if err != nil {
		t.Fatalf("ListSecretMetadata: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("default env len = %d, want 2", len(got))
	}

	// Order is by name ascending.
	if got[0].Name != "OPENAI_KEY" {
		t.Errorf("got[0].Name = %q, want %q", got[0].Name, "OPENAI_KEY")
	}
	if got[1].Name != "STRIPE_KEY" {
		t.Errorf("got[1].Name = %q, want %q", got[1].Name, "STRIPE_KEY")
	}

	// Preview field is populated and matches what we stored.
	if got[0].Preview != "…aBcD" {
		t.Errorf("OPENAI_KEY preview = %q, want %q", got[0].Preview, "…aBcD")
	}
	if got[1].Preview != "…Bc1D" {
		t.Errorf("STRIPE_KEY preview = %q, want %q", got[1].Preview, "…Bc1D")
	}

	// Version is 1 for fresh insertions, updated_at is non-zero.
	for _, m := range got {
		if m.Version != 1 {
			t.Errorf("%s.Version = %d, want 1", m.Name, m.Version)
		}
		if m.UpdatedAt.IsZero() {
			t.Errorf("%s.UpdatedAt is zero; expected datetime('now') value", m.Name)
		}
	}

	// Env isolation: "prod" env contains DB_PASS only.
	prod, err := v.ListSecretMetadata("prod")
	if err != nil {
		t.Fatalf("ListSecretMetadata(prod): %v", err)
	}
	if len(prod) != 1 || prod[0].Name != "DB_PASS" {
		t.Errorf("prod env = %+v, want [DB_PASS]", prod)
	}
}

func TestListSecretMetadata_NeverTouchesEncryptedValue(t *testing.T) {
	// This is the explicit guard for invariant I-1: even after we shovel
	// dangerous-looking sentinel strings into encrypted_value, the
	// metadata API output must never contain them. We assert by ensuring
	// the returned VaultKeyMeta has no field that could carry the value
	// (the struct does not expose one) and by separately verifying via
	// raw SQL that the ciphertext is still in place (unaltered).
	v := tempVault(t)
	const sentinel = "AAAA-PLAINTEXT-LEAK-SENTINEL-BBBB"
	if err := v.SetSecretWithPreview("KEY", sentinel, "default", "…leak"); err != nil {
		t.Fatalf("SetSecretWithPreview: %v", err)
	}

	metas, err := v.ListSecretMetadata("default")
	if err != nil {
		t.Fatalf("ListSecretMetadata: %v", err)
	}
	if len(metas) != 1 {
		t.Fatalf("len = %d, want 1", len(metas))
	}
	// Walk all string fields and assert the sentinel is not present.
	m := metas[0]
	for _, field := range []string{m.Name, m.Preview} {
		if containsSubstring(field, sentinel) {
			t.Fatalf("metadata field contains ciphertext sentinel: %q", field)
		}
	}

	// And the row in storage still has the original encrypted_value.
	var stored string
	row := v.db.QueryRow(`SELECT encrypted_value FROM secrets WHERE name = ?`, "KEY")
	if err := row.Scan(&stored); err != nil {
		t.Fatalf("verify encrypted_value: %v", err)
	}
	if stored != sentinel {
		t.Errorf("encrypted_value mutated: got %q, want %q", stored, sentinel)
	}
}

func TestUpdateSecretPreview(t *testing.T) {
	v := tempVault(t)
	if err := v.SetSecretWithPreview("KEY", "ct", "default", ""); err != nil {
		t.Fatalf("SetSecretWithPreview: %v", err)
	}

	// Update to a derived preview.
	if err := v.UpdateSecretPreview("KEY", "default", "…aBcD"); err != nil {
		t.Fatalf("UpdateSecretPreview: %v", err)
	}
	metas, _ := v.ListSecretMetadata("default")
	if metas[0].Preview != "…aBcD" {
		t.Errorf("after update preview = %q, want %q", metas[0].Preview, "…aBcD")
	}

	// Update to empty (e.g. preview.enabled=false then re-derive).
	if err := v.UpdateSecretPreview("KEY", "default", ""); err != nil {
		t.Fatalf("UpdateSecretPreview empty: %v", err)
	}
	metas, _ = v.ListSecretMetadata("default")
	if metas[0].Preview != "" {
		t.Errorf("after clear preview = %q, want empty", metas[0].Preview)
	}

	// Nonexistent secret -> ErrSecretNotFound.
	err := v.UpdateSecretPreview("NOPE", "default", "x")
	if !errors.Is(err, ErrSecretNotFound) {
		t.Errorf("UpdateSecretPreview(missing): %v, want ErrSecretNotFound", err)
	}
}

func TestListSecretsForBackfill_OnlyReturnsEmptyPreviews(t *testing.T) {
	v := tempVault(t)
	// Three secrets: two with empty preview (backfill candidates), one with
	// a populated preview (should be skipped).
	if err := v.SetSecretWithPreview("A", "cta", "default", ""); err != nil {
		t.Fatalf("seed A: %v", err)
	}
	if err := v.SetSecretWithPreview("B", "ctb", "default", "…BcD"); err != nil {
		t.Fatalf("seed B: %v", err)
	}
	if err := v.SetSecretWithPreview("C", "ctc", "default", ""); err != nil {
		t.Fatalf("seed C: %v", err)
	}

	got, err := v.ListSecretsForBackfill("default")
	if err != nil {
		t.Fatalf("ListSecretsForBackfill: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Name != "A" || got[1].Name != "C" {
		t.Errorf("backfill candidates = %+v, want [A, C]", got)
	}
	// Each row carries the encrypted_value so the CLI layer can decrypt.
	if got[0].EncryptedValue != "cta" || got[1].EncryptedValue != "ctc" {
		t.Errorf("encrypted values not propagated: %+v", got)
	}
}

func TestSetSecretWithPreview_AtomicallyUpdatesBothColumns(t *testing.T) {
	// Verify the round-trip: stored ciphertext and stored preview agree
	// with what we passed in, in a single call.
	v := tempVault(t)
	if err := v.SetSecretWithPreview("KEY", "ciphertext-1", "default", "…ab12"); err != nil {
		t.Fatalf("SetSecretWithPreview: %v", err)
	}

	s, err := v.GetSecret("KEY", "default")
	if err != nil {
		t.Fatalf("GetSecret: %v", err)
	}
	if s.EncryptedValue != "ciphertext-1" {
		t.Errorf("EncryptedValue = %q", s.EncryptedValue)
	}

	metas, _ := v.ListSecretMetadata("default")
	if len(metas) != 1 || metas[0].Preview != "…ab12" {
		t.Errorf("metas = %+v", metas)
	}

	// Upsert with new ciphertext + new preview: both update atomically.
	if err := v.SetSecretWithPreview("KEY", "ciphertext-2", "default", "…cd34"); err != nil {
		t.Fatalf("re-SetSecretWithPreview: %v", err)
	}
	s, _ = v.GetSecret("KEY", "default")
	if s.EncryptedValue != "ciphertext-2" || s.Version != 2 {
		t.Errorf("post-update ciphertext/version = %q/%d", s.EncryptedValue, s.Version)
	}
	metas, _ = v.ListSecretMetadata("default")
	if metas[0].Preview != "…cd34" {
		t.Errorf("post-update preview = %q", metas[0].Preview)
	}
}

func TestSetSecret_LegacySignatureStoresEmptyPreview(t *testing.T) {
	// Backward-compatibility check: tests that still use the 3-arg
	// SetSecret signature get an empty preview, not nil/NULL.
	v := tempVault(t)
	if err := v.SetSecret("KEY", "ct", "default"); err != nil {
		t.Fatalf("SetSecret: %v", err)
	}
	metas, _ := v.ListSecretMetadata("default")
	if len(metas) != 1 {
		t.Fatalf("len = %d", len(metas))
	}
	if metas[0].Preview != "" {
		t.Errorf("legacy SetSecret should yield empty preview, got %q", metas[0].Preview)
	}
}

func TestSetSecretBatchWithPreview(t *testing.T) {
	v := tempVault(t)
	records := []SecretWrite{
		{Name: "A", EncryptedValue: "ct-a", Preview: "…a4"},
		{Name: "B", EncryptedValue: "ct-b", Preview: "…b4"},
	}
	if err := v.SetSecretBatchWithPreview(records, "default"); err != nil {
		t.Fatalf("SetSecretBatchWithPreview: %v", err)
	}
	metas, _ := v.ListSecretMetadata("default")
	if len(metas) != 2 {
		t.Fatalf("len = %d, want 2", len(metas))
	}
	want := map[string]string{"A": "…a4", "B": "…b4"}
	for _, m := range metas {
		if m.Preview != want[m.Name] {
			t.Errorf("%s preview = %q, want %q", m.Name, m.Preview, want[m.Name])
		}
	}
}

func containsSubstring(haystack, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
