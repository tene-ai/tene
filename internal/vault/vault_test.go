package vault

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func tempVault(t *testing.T) *Vault {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, ".tene", "vault.db")
	v, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })
	return v
}

func TestNew(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, ".tene", "vault.db")
	v, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer func() { _ = v.Close() }()

	// Check file permissions
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("file perm = %o, want 0600", info.Mode().Perm())
	}
}

func TestSetGetSecret(t *testing.T) {
	v := tempVault(t)

	err := v.SetSecret("API_KEY", "encrypted_value_here", "default")
	if err != nil {
		t.Fatalf("SetSecret() error: %v", err)
	}

	secret, err := v.GetSecret("API_KEY", "default")
	if err != nil {
		t.Fatalf("GetSecret() error: %v", err)
	}

	if secret.Name != "API_KEY" {
		t.Errorf("Name = %q, want %q", secret.Name, "API_KEY")
	}
	if secret.EncryptedValue != "encrypted_value_here" {
		t.Errorf("EncryptedValue = %q, want %q", secret.EncryptedValue, "encrypted_value_here")
	}
	if secret.Version != 1 {
		t.Errorf("Version = %d, want 1", secret.Version)
	}
}

func TestSetSecret_Upsert(t *testing.T) {
	v := tempVault(t)

	_ = v.SetSecret("API_KEY", "v1", "default")
	_ = v.SetSecret("API_KEY", "v2", "default")

	secret, _ := v.GetSecret("API_KEY", "default")
	if secret.Version != 2 {
		t.Errorf("Version = %d, want 2", secret.Version)
	}
	if secret.EncryptedValue != "v2" {
		t.Errorf("EncryptedValue = %q, want %q", secret.EncryptedValue, "v2")
	}
}

func TestGetSecret_NotFound(t *testing.T) {
	v := tempVault(t)

	_, err := v.GetSecret("NONEXISTENT", "default")
	if !errors.Is(err, ErrSecretNotFound) {
		t.Errorf("expected ErrSecretNotFound, got %v", err)
	}
}

func TestListSecrets(t *testing.T) {
	v := tempVault(t)

	_ = v.SetSecret("A_KEY", "val1", "default")
	_ = v.SetSecret("B_KEY", "val2", "default")
	_ = v.SetSecret("C_KEY", "val3", "prod")

	secrets, err := v.ListSecrets("default")
	if err != nil {
		t.Fatalf("ListSecrets() error: %v", err)
	}
	if len(secrets) != 2 {
		t.Errorf("len = %d, want 2", len(secrets))
	}
}

func TestDeleteSecret(t *testing.T) {
	v := tempVault(t)

	_ = v.SetSecret("API_KEY", "val", "default")
	err := v.DeleteSecret("API_KEY", "default")
	if err != nil {
		t.Fatalf("DeleteSecret() error: %v", err)
	}

	_, err = v.GetSecret("API_KEY", "default")
	if !errors.Is(err, ErrSecretNotFound) {
		t.Errorf("expected ErrSecretNotFound after delete, got %v", err)
	}
}

func TestDeleteSecret_NotFound(t *testing.T) {
	v := tempVault(t)

	err := v.DeleteSecret("NONEXISTENT", "default")
	if !errors.Is(err, ErrSecretNotFound) {
		t.Errorf("expected ErrSecretNotFound, got %v", err)
	}
}

func TestCountSecrets(t *testing.T) {
	v := tempVault(t)

	_ = v.SetSecret("A", "v", "default")
	_ = v.SetSecret("B", "v", "default")

	count, err := v.CountSecrets("default")
	if err != nil {
		t.Fatalf("CountSecrets() error: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestGetAllSecrets(t *testing.T) {
	v := tempVault(t)

	_ = v.SetSecret("A", "va", "default")
	_ = v.SetSecret("B", "vb", "default")

	all, err := v.GetAllSecrets("default")
	if err != nil {
		t.Fatalf("GetAllSecrets() error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("len = %d, want 2", len(all))
	}
	if all["A"] != "va" || all["B"] != "vb" {
		t.Errorf("unexpected values: %v", all)
	}
}

func TestSetSecretBatch(t *testing.T) {
	v := tempVault(t)

	batch := map[string]string{
		"X": "vx",
		"Y": "vy",
	}
	err := v.SetSecretBatch(batch, "default")
	if err != nil {
		t.Fatalf("SetSecretBatch() error: %v", err)
	}

	count, _ := v.CountSecrets("default")
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestEnvironmentCRUD(t *testing.T) {
	v := tempVault(t)

	// Set active creates "default"
	err := v.SetActiveEnvironment("default")
	if err != nil {
		t.Fatalf("SetActiveEnvironment() error: %v", err)
	}

	active, err := v.GetActiveEnvironment()
	if err != nil {
		t.Fatalf("GetActiveEnvironment() error: %v", err)
	}
	if active != "default" {
		t.Errorf("active = %q, want %q", active, "default")
	}

	// Create another env
	err = v.CreateEnvironment("staging")
	if err != nil {
		t.Fatalf("CreateEnvironment() error: %v", err)
	}

	// List
	envs, err := v.ListEnvironments()
	if err != nil {
		t.Fatalf("ListEnvironments() error: %v", err)
	}
	if len(envs) != 2 {
		t.Errorf("len = %d, want 2", len(envs))
	}

	// Switch
	err = v.SetActiveEnvironment("staging")
	if err != nil {
		t.Fatalf("SetActiveEnvironment() error: %v", err)
	}

	active, _ = v.GetActiveEnvironment()
	if active != "staging" {
		t.Errorf("active = %q, want %q", active, "staging")
	}

	// Delete
	_, err = v.DeleteEnvironment("staging")
	if err != nil {
		t.Fatalf("DeleteEnvironment() error: %v", err)
	}
}

func TestEnvironmentIsolation(t *testing.T) {
	v := tempVault(t)

	_ = v.SetSecret("KEY", "default_val", "default")
	_ = v.SetSecret("KEY", "prod_val", "prod")

	s1, _ := v.GetSecret("KEY", "default")
	s2, _ := v.GetSecret("KEY", "prod")

	if s1.EncryptedValue == s2.EncryptedValue {
		t.Error("secrets in different environments should have different values")
	}
}

func TestMeta(t *testing.T) {
	v := tempVault(t)

	err := v.SetMeta("test_key", "test_value")
	if err != nil {
		t.Fatalf("SetMeta() error: %v", err)
	}

	val, err := v.GetMeta("test_key")
	if err != nil {
		t.Fatalf("GetMeta() error: %v", err)
	}
	if val != "test_value" {
		t.Errorf("value = %q, want %q", val, "test_value")
	}

	// Update
	_ = v.SetMeta("test_key", "updated")
	val, _ = v.GetMeta("test_key")
	if val != "updated" {
		t.Errorf("value = %q, want %q", val, "updated")
	}
}

func TestMeta_NotFound(t *testing.T) {
	v := tempVault(t)

	_, err := v.GetMeta("nonexistent")
	if !errors.Is(err, ErrMetaNotFound) {
		t.Errorf("expected ErrMetaNotFound, got %v", err)
	}
}

func TestAuditLog(t *testing.T) {
	v := tempVault(t)

	err := v.AddAuditLog("vault.init", "", "test init")
	if err != nil {
		t.Fatalf("AddAuditLog() error: %v", err)
	}

	// Verify by setting and getting a secret (which also adds audit entries)
	_ = v.SetSecret("KEY", "val", "default")

	// Just verify no errors occurred
}
