package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestInit_SelectiveAgent(t *testing.T) {
	env := setupTestEnv(t)

	// Init with --claude only
	_, _, err := env.run("init", "test-project", "--quiet", "--claude")
	if err != nil {
		t.Fatalf("init --claude error: %v", err)
	}

	// CLAUDE.md should exist
	claudePath := filepath.Join(env.Dir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Error("CLAUDE.md not created")
	}

	// GEMINI.md should NOT exist
	geminiPath := filepath.Join(env.Dir, "GEMINI.md")
	if _, err := os.Stat(geminiPath); !os.IsNotExist(err) {
		t.Error("GEMINI.md should not exist after --claude only init")
	}
}

func TestInit_AddAgentToExistingVault(t *testing.T) {
	env := setupTestEnv(t)

	// Init with --claude only
	_, _, err := env.run("init", "test-project", "--quiet", "--claude")
	if err != nil {
		t.Fatalf("init --claude error: %v", err)
	}

	// Now add --gemini to existing vault
	stdout, _, err := env.run("init", "--gemini")
	if err != nil {
		t.Fatalf("init --gemini on existing vault error: %v", err)
	}

	// GEMINI.md should now exist
	geminiPath := filepath.Join(env.Dir, "GEMINI.md")
	if _, err := os.Stat(geminiPath); os.IsNotExist(err) {
		t.Error("GEMINI.md not created on second init")
	}

	if !strings.Contains(stdout, "GEMINI.md") {
		t.Errorf("expected output to mention GEMINI.md, got: %s", stdout)
	}
}

func TestInit_AddAgentAlreadyExists(t *testing.T) {
	env := setupTestEnv(t)

	// Init with all agents
	_, _, err := env.run("init", "test-project", "--quiet")
	if err != nil {
		t.Fatalf("init error: %v", err)
	}

	// Try adding --gemini again — should say already exists
	stdout, _, err := env.run("init", "--gemini")
	if err != nil {
		t.Fatalf("init --gemini error: %v", err)
	}

	if !strings.Contains(stdout, "already exist") {
		t.Errorf("expected 'already exist' message, got: %s", stdout)
	}
}

func TestInit_ExistingFileWithoutTeneSection(t *testing.T) {
	env := setupTestEnv(t)

	// Create GEMINI.md with custom content BEFORE init
	geminiPath := filepath.Join(env.Dir, "GEMINI.md")
	customContent := "# My Project\n\nSome existing gemini instructions.\n"
	if err := os.WriteFile(geminiPath, []byte(customContent), 0644); err != nil {
		t.Fatalf("failed to write GEMINI.md: %v", err)
	}

	// Init with --gemini
	_, _, err := env.run("init", "test-project", "--quiet", "--gemini")
	if err != nil {
		t.Fatalf("init --gemini error: %v", err)
	}

	// GEMINI.md should contain BOTH original content and tene section
	content, err := os.ReadFile(geminiPath)
	if err != nil {
		t.Fatalf("failed to read GEMINI.md: %v", err)
	}

	if !strings.Contains(string(content), "My Project") {
		t.Error("original content was lost")
	}
	if !strings.Contains(string(content), "Secrets Management") {
		t.Error("tene section was not appended")
	}
}

func TestInit_ExistingFileWithTeneSection(t *testing.T) {
	env := setupTestEnv(t)

	// Create GEMINI.md that already has tene section
	geminiPath := filepath.Join(env.Dir, "GEMINI.md")
	existingContent := "# My Project\n\n# Secrets Management\n\nAlready has tene.\n"
	if err := os.WriteFile(geminiPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("failed to write GEMINI.md: %v", err)
	}

	// Init with --gemini
	_, _, err := env.run("init", "test-project", "--quiet", "--gemini")
	if err != nil {
		t.Fatalf("init --gemini error: %v", err)
	}

	// Content should be unchanged (not duplicated)
	content, err := os.ReadFile(geminiPath)
	if err != nil {
		t.Fatalf("failed to read GEMINI.md: %v", err)
	}

	if string(content) != existingContent {
		t.Errorf("file was modified when it shouldn't have been.\ngot: %s", string(content))
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
