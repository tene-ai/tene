package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInit_Basic(t *testing.T) {
	env := setupTestEnv(t)

	stdout, _, err := env.run("init", "test-project", "--quiet")
	if err != nil {
		t.Fatalf("init error: %v", err)
	}

	// vault.db should exist
	vaultPath := filepath.Join(env.Dir, ".tene", "vault.db")
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Error("vault.db not created")
	}

	// .gitignore should exist
	gitignorePath := filepath.Join(env.Dir, ".tene", ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error(".gitignore not created")
	}

	// vault.json should exist
	vaultJSONPath := filepath.Join(env.Dir, ".tene", "vault.json")
	if _, err := os.Stat(vaultJSONPath); os.IsNotExist(err) {
		t.Error("vault.json not created")
	}

	// AI editor context files should exist
	agentContextFiles := []string{
		"CLAUDE.md",
		".cursor/rules/tene.mdc",
		".windsurfrules",
		"GEMINI.md",
		"AGENTS.md",
	}
	for _, f := range agentContextFiles {
		p := filepath.Join(env.Dir, f)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("agent context file %s not created", f)
		}
	}

	// Recovery Key should be in stdout (quiet mode)
	if len(stdout) == 0 {
		t.Error("expected recovery key in stdout")
	}
}

func TestInit_AlreadyInitialized(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	// Second init should not error
	_, _, err := env.run("init", "--quiet")
	if err != nil {
		t.Fatalf("second init should not error: %v", err)
	}
}

func TestInit_JSON(t *testing.T) {
	env := setupTestEnv(t)

	stdout, _, err := env.runJSON("init", "test-project")
	if err != nil {
		t.Fatalf("init error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse error: %v\nstdout: %s", err, stdout)
	}

	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if result["project"] != "test-project" {
		t.Errorf("project = %v, want test-project", result["project"])
	}
}
