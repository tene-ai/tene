package claudemd

import (
	"os"
	"path/filepath"
	"strings"
)

// Generator creates and manages AI editor context files.
type Generator struct {
	projectDir string
}

// NewGenerator creates a new Generator.
func NewGenerator(projectDir string) *Generator {
	return &Generator{projectDir: projectDir}
}

// Generate creates CLAUDE.md or appends the tene section to an existing file.
// Returns true if the file was created/modified, false if skipped.
func (g *Generator) Generate() (bool, error) {
	return g.generateFile(AgentFiles[0]) // claude
}

// GenerateAll creates context files for all supported AI editors.
// Returns a list of file paths that were created or modified.
func (g *Generator) GenerateAll() ([]string, error) {
	var created []string
	for _, af := range AgentFiles {
		ok, err := g.generateFile(af)
		if err != nil {
			return created, err
		}
		if ok {
			created = append(created, af.Path)
		}
	}
	return created, nil
}

// generateFile creates or appends the tene section to a single agent file.
func (g *Generator) generateFile(af AgentFile) (bool, error) {
	path := filepath.Join(g.projectDir, af.Path)

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, err
	}

	// Build content for this file format
	fileContent := g.buildContent(af.Format)

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new file
			return true, os.WriteFile(path, []byte(fileContent), 0644)
		}
		return false, err
	}

	// Skip if tene section already exists
	if g.HasTeneSection(string(content)) {
		return false, nil
	}

	// Append section to existing file
	separator := "\n\n"
	if !strings.HasSuffix(string(content), "\n") {
		separator = "\n\n"
	}
	updated := string(content) + separator + fileContent
	return true, os.WriteFile(path, []byte(updated), 0644)
}

// buildContent returns the full file content for the given format.
func (g *Generator) buildContent(format string) string {
	if format == "mdc" {
		return CursorMdcFrontmatter + SecretsMdTemplate
	}
	return SecretsMdTemplate
}

// HasTeneSection checks if the content already contains the tene Secrets Management section.
func (g *Generator) HasTeneSection(content string) bool {
	return strings.Contains(content, SectionHeader) ||
		(strings.Contains(content, "tene") && strings.Contains(content, "secret management"))
}

// FilePath returns the absolute path to CLAUDE.md.
func (g *Generator) FilePath() string {
	return filepath.Join(g.projectDir, "CLAUDE.md")
}

// AgentNames returns the list of all supported agent names.
func AgentNames() []string {
	names := make([]string, len(AgentFiles))
	for i, af := range AgentFiles {
		names[i] = af.Name
	}
	return names
}
