package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteVaultJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.json")

	err := WriteVaultJSON(path, "test-project", "default")
	if err != nil {
		t.Fatalf("WriteVaultJSON() error: %v", err)
	}

	vj, err := ReadVaultJSON(path)
	if err != nil {
		t.Fatalf("ReadVaultJSON() error: %v", err)
	}

	if vj.ProjectName != "test-project" {
		t.Errorf("ProjectName = %q, want %q", vj.ProjectName, "test-project")
	}
	if vj.ActiveEnvironment != "default" {
		t.Errorf("ActiveEnvironment = %q, want %q", vj.ActiveEnvironment, "default")
	}
	if vj.VaultVersion != 1 {
		t.Errorf("VaultVersion = %d, want 1", vj.VaultVersion)
	}
	expectedAgents := []string{"claude", "cursor", "windsurf", "gemini", "codex"}
	if len(vj.Agents) != len(expectedAgents) {
		t.Errorf("Agents = %v, want %v", vj.Agents, expectedAgents)
	} else {
		for i, a := range expectedAgents {
			if vj.Agents[i] != a {
				t.Errorf("Agents[%d] = %q, want %q", i, vj.Agents[i], a)
			}
		}
	}
	if vj.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}

	// Verify file permissions
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Errorf("permission = %o, want 0600", info.Mode().Perm())
	}
}

func TestUpdateVaultJSONEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.json")

	_ = WriteVaultJSON(path, "test-project", "default")
	err := UpdateVaultJSONEnv(path, "production")
	if err != nil {
		t.Fatalf("UpdateVaultJSONEnv() error: %v", err)
	}

	vj, _ := ReadVaultJSON(path)
	if vj.ActiveEnvironment != "production" {
		t.Errorf("ActiveEnvironment = %q, want %q", vj.ActiveEnvironment, "production")
	}
	// Verify ProjectName is preserved
	if vj.ProjectName != "test-project" {
		t.Errorf("ProjectName = %q, want %q", vj.ProjectName, "test-project")
	}
}

func TestReadVaultJSON_NotExists(t *testing.T) {
	_, err := ReadVaultJSON("/nonexistent/path/vault.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
