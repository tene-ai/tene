package claudemd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate_NewFile(t *testing.T) {
	dir := t.TempDir()
	gen := NewGenerator(dir)

	created, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if !created {
		t.Error("expected created = true")
	}

	content, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	if !strings.Contains(string(content), SectionHeader) {
		t.Error("file should contain section header")
	}
	if !strings.Contains(string(content), "tene get") {
		t.Error("file should contain tene usage")
	}
}

func TestGenerate_ExistingWithoutSection(t *testing.T) {
	dir := t.TempDir()
	existing := "# My Project\n\nSome existing content.\n"
	_ = os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(existing), 0644)

	gen := NewGenerator(dir)
	created, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if !created {
		t.Error("expected created = true")
	}

	content, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if !strings.Contains(string(content), "# My Project") {
		t.Error("should preserve existing content")
	}
	if !strings.Contains(string(content), SectionHeader) {
		t.Error("should append tene section")
	}
}

func TestGenerate_ExistingWithSection(t *testing.T) {
	dir := t.TempDir()
	existing := "# My Project\n\n# Secrets Management\n\nAlready has tene.\n"
	_ = os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(existing), 0644)

	gen := NewGenerator(dir)
	created, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if created {
		t.Error("expected created = false (should skip)")
	}
}

func TestTemplate_URL(t *testing.T) {
	if !strings.Contains(SecretsMdTemplate, "github.com/agent-kay-it/tene") {
		t.Error("template should contain agent-kay-it URL")
	}
	if strings.Contains(SecretsMdTemplate, "tomo-kay") {
		t.Error("template should not contain legacy tomo-kay URL")
	}
}

func TestTemplate_LlmsTxtReferences(t *testing.T) {
	// Generated rule files must point AI agents at the canonical llms.txt
	// resources so agents that only load a project rule file (e.g. CLAUDE.md)
	// can still discover the extended context.
	if !strings.Contains(SecretsMdTemplate, "https://tene.sh/llms.txt") {
		t.Error("template should reference https://tene.sh/llms.txt")
	}
	if !strings.Contains(SecretsMdTemplate, "https://tene.sh/llms-full.txt") {
		t.Error("template should reference https://tene.sh/llms-full.txt")
	}
}

func TestHasTeneSection(t *testing.T) {
	gen := NewGenerator("")

	tests := []struct {
		content  string
		expected bool
	}{
		{"# Secrets Management", true},
		{"uses tene for secret management", true},
		{"# My Project", false},
		{"", false},
	}

	for _, tc := range tests {
		if got := gen.HasTeneSection(tc.content); got != tc.expected {
			t.Errorf("HasTeneSection(%q) = %v, want %v", tc.content, got, tc.expected)
		}
	}
}

func TestGenerateAll_CreatesAllFiles(t *testing.T) {
	dir := t.TempDir()
	gen := NewGenerator(dir)

	created, err := gen.GenerateAll()
	if err != nil {
		t.Fatalf("GenerateAll() error: %v", err)
	}

	if len(created) != len(AgentFiles) {
		t.Errorf("expected %d files created, got %d: %v", len(AgentFiles), len(created), created)
	}

	// Verify all files exist
	for _, af := range AgentFiles {
		path := filepath.Join(dir, af.Path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("file %s was not created", af.Path)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("cannot read %s: %v", af.Path, err)
			continue
		}

		// All files should contain the section header
		if !strings.Contains(string(content), SectionHeader) {
			t.Errorf("%s should contain section header", af.Path)
		}

		// All files should contain Quick Reference table
		if !strings.Contains(string(content), "Quick Reference") {
			t.Errorf("%s should contain Quick Reference", af.Path)
		}
	}
}

func TestCursorMdcFormat(t *testing.T) {
	dir := t.TempDir()
	gen := NewGenerator(dir)

	_, err := gen.GenerateAll()
	if err != nil {
		t.Fatalf("GenerateAll() error: %v", err)
	}

	mdcPath := filepath.Join(dir, ".cursor", "rules", "tene.mdc")
	content, err := os.ReadFile(mdcPath)
	if err != nil {
		t.Fatalf("cannot read .mdc file: %v", err)
	}

	s := string(content)

	// Check frontmatter
	if !strings.HasPrefix(s, "---\n") {
		t.Error(".mdc file should start with frontmatter ---")
	}
	if !strings.Contains(s, "description: Secret management with tene") {
		t.Error(".mdc file should contain description in frontmatter")
	}
	if !strings.Contains(s, "alwaysApply: true") {
		t.Error(".mdc file should contain alwaysApply in frontmatter")
	}

	// Check body content is also present
	if !strings.Contains(s, SectionHeader) {
		t.Error(".mdc file should contain section header after frontmatter")
	}
}

func TestGenerateAll_ExistingFileAppend(t *testing.T) {
	dir := t.TempDir()

	// Pre-create a GEMINI.md with existing content
	existing := "# My Gemini Rules\n\nSome content.\n"
	_ = os.WriteFile(filepath.Join(dir, "GEMINI.md"), []byte(existing), 0644)

	gen := NewGenerator(dir)
	created, err := gen.GenerateAll()
	if err != nil {
		t.Fatalf("GenerateAll() error: %v", err)
	}

	// GEMINI.md should be in the created list (was modified)
	found := false
	for _, f := range created {
		if f == "GEMINI.md" {
			found = true
		}
	}
	if !found {
		t.Error("GEMINI.md should be in created list")
	}

	content, _ := os.ReadFile(filepath.Join(dir, "GEMINI.md"))
	s := string(content)

	if !strings.Contains(s, "# My Gemini Rules") {
		t.Error("should preserve existing content")
	}
	if !strings.Contains(s, SectionHeader) {
		t.Error("should append tene section")
	}
}

func TestGenerateAll_SkipExisting(t *testing.T) {
	dir := t.TempDir()

	// Pre-create AGENTS.md with tene section already present
	existing := "# My Agent Rules\n\n# Secrets Management\n\nAlready configured.\n"
	_ = os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(existing), 0644)

	gen := NewGenerator(dir)
	created, err := gen.GenerateAll()
	if err != nil {
		t.Fatalf("GenerateAll() error: %v", err)
	}

	// AGENTS.md should NOT be in the created list (was skipped)
	for _, f := range created {
		if f == "AGENTS.md" {
			t.Error("AGENTS.md should have been skipped (already has tene section)")
		}
	}

	// Other files should still be created
	if len(created) != len(AgentFiles)-1 {
		t.Errorf("expected %d files created, got %d", len(AgentFiles)-1, len(created))
	}
}

func TestCursorDirectoryCreation(t *testing.T) {
	dir := t.TempDir()
	gen := NewGenerator(dir)

	_, err := gen.GenerateAll()
	if err != nil {
		t.Fatalf("GenerateAll() error: %v", err)
	}

	// .cursor/rules/ directory should exist
	rulesDir := filepath.Join(dir, ".cursor", "rules")
	info, err := os.Stat(rulesDir)
	if os.IsNotExist(err) {
		t.Fatal(".cursor/rules/ directory was not created")
	}
	if !info.IsDir() {
		t.Error(".cursor/rules/ should be a directory")
	}
}

func TestAgentNames(t *testing.T) {
	names := AgentNames()
	expected := []string{"claude", "cursor", "windsurf", "gemini", "codex"}

	if len(names) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(names))
	}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("name[%d] = %q, want %q", i, name, expected[i])
		}
	}
}
