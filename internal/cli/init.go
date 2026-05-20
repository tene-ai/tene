package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agent-kay-it/tene/internal/claudemd"
	"github.com/agent-kay-it/tene/internal/config"
	"github.com/agent-kay-it/tene/internal/keychain"
	"github.com/agent-kay-it/tene/internal/recovery"
	"github.com/agent-kay-it/tene/internal/vault"
	"github.com/agent-kay-it/tene/pkg/crypto"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new tene vault in the current project",
	Long: `Initialize a new tene vault and generate AI editor context files.

By default, context files are generated for all supported AI editors:
  CLAUDE.md, .cursor/rules/tene.mdc, .windsurfrules, GEMINI.md, AGENTS.md

Use flags to generate only specific editor files:
  tene init --claude --gemini     # Only CLAUDE.md and GEMINI.md
  tene init --cursor              # Only .cursor/rules/tene.mdc`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

var (
	initFlagClaude   bool
	initFlagCursor   bool
	initFlagWindsurf bool
	initFlagGemini   bool
	initFlagCodex    bool
)

func init() {
	initCmd.Flags().BoolVar(&initFlagClaude, "claude", false, "Generate CLAUDE.md only")
	initCmd.Flags().BoolVar(&initFlagCursor, "cursor", false, "Generate .cursor/rules/tene.mdc only")
	initCmd.Flags().BoolVar(&initFlagWindsurf, "windsurf", false, "Generate .windsurfrules only")
	initCmd.Flags().BoolVar(&initFlagGemini, "gemini", false, "Generate GEMINI.md only")
	initCmd.Flags().BoolVar(&initFlagCodex, "codex", false, "Generate AGENTS.md only")
}

func runInit(cmd *cobra.Command, args []string) error {
	dir := resolveDir()

	// Determine project name
	projectName := filepath.Base(dir)
	if len(args) > 0 {
		projectName = args[0]
	}

	// 1. Check if already initialized
	vaultPath := filepath.Join(dir, ".tene", "vault.db")
	if fileExists(vaultPath) {
		selected := selectedAgents()
		if len(selected) > 0 {
			// Vault exists but agent flags specified — generate agent files only
			return addAgentFiles(dir, selected)
		}
		if flagJSON {
			return printJSON(map[string]any{
				"ok":      true,
				"message": "already_initialized",
			})
		}
		fmt.Println("Vault already exists. Use existing vault.")
		return nil
	}

	// 2. Master Password input (2x confirm)
	password, err := promptPasswordConfirm("Set your Master Password (used to encrypt all secrets):\nMaster Password: ")
	if err != nil {
		return err
	}

	// 3. Generate salt + KDF -> Master Key
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return err
	}
	masterKey, err := crypto.DeriveKey(password, salt)
	if err != nil {
		return err
	}
	defer crypto.ZeroBytes(masterKey)

	// 4. Create .tene/ directory
	teneDir := filepath.Join(dir, ".tene")
	if err := os.MkdirAll(teneDir, 0700); err != nil {
		return fmt.Errorf("cannot create .tene/ directory: %w", err)
	}

	// 5. Create SQLite vault
	v, err := vault.New(vaultPath)
	if err != nil {
		return err
	}
	defer func() { _ = v.Close() }()

	// 6. Store metadata.
	//
	// schema_version is intentionally NOT written here: vault.New has
	// already stamped it via runMigrations() to currentSchemaVersion.
	// Overwriting it would clobber the v2 stamp (sprint
	// cli-ux-permission-model, F1) and trip the migration system on the
	// next open.
	if err := v.SetMeta("vault_version", "1"); err != nil {
		return err
	}
	if err := v.SetMeta("created_at", time.Now().UTC().Format(time.RFC3339)); err != nil {
		return err
	}
	if err := v.SetMeta("kdf_salt", encodeBase64(salt)); err != nil {
		return err
	}
	if err := v.SetMeta("kdf_params", `{"time":3,"memory":65536,"threads":1,"keyLen":32}`); err != nil {
		return err
	}
	if err := v.SetMeta("project_name", projectName); err != nil {
		return err
	}

	// 7. Generate Recovery Key
	mnemonic, err := recovery.GenerateMnemonic()
	if err != nil {
		return err
	}
	blob, err := recovery.EncryptMasterKey(masterKey, mnemonic)
	if err != nil {
		return err
	}
	if err := v.SetMeta("recovery_blob", encodeBase64(blob)); err != nil {
		return err
	}

	// 8. Create default environment
	if err := v.SetActiveEnvironment("default"); err != nil {
		return err
	}

	// 9. Store master key in keychain
	var ks keychain.KeyStore
	if flagNoKeychain {
		home, _ := os.UserHomeDir()
		ks = keychain.NewFileStore(filepath.Join(home, ".tene", "keyfile"))
	} else {
		ks = keychain.NewStore(dir)
	}
	if err := ks.Store(masterKey); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not store key in keychain: %v\n", err)
	}

	// 9.5. Create .tene/vault.json
	vaultJSONPath := filepath.Join(teneDir, "vault.json")
	if err := vault.WriteVaultJSON(vaultJSONPath, projectName, "default"); err != nil {
		return fmt.Errorf("cannot create vault.json: %w", err)
	}

	// 9.6. Ensure global config directory
	_ = config.EnsureConfigDir()

	// 10. Create .tene/.gitignore
	_ = writeGitignore(filepath.Join(teneDir, ".gitignore"))

	// 11. Add .tene/ to root .gitignore
	_ = addToRootGitignore(dir)

	// 12. Generate AI editor context files (CLAUDE.md, .cursor/rules/tene.mdc, etc.)
	gen := claudemd.NewGenerator(dir)
	selected := selectedAgents()
	var agentFiles []string
	if len(selected) > 0 {
		agentFiles, _ = gen.GenerateSelected(selected)
	} else {
		agentFiles, _ = gen.GenerateAll()
	}

	// 13. Audit log
	_ = v.AddAuditLog("vault.init", "", "project="+projectName)

	// 14. Output
	if flagJSON {
		return printJSON(map[string]any{
			"ok":          true,
			"project":     projectName,
			"vault":       ".tene/vault.db",
			"agentFiles":  agentFiles,
			"recoveryKey": mnemonic,
			"environment": "default",
		})
	}

	if !flagQuiet {
		fmt.Println()
		fmt.Printf("  Created .tene/vault.db (local encrypted vault)\n")
		fmt.Printf("  Added .tene/ to .gitignore\n")
		fmt.Printf("  Master Key saved to OS Keychain\n")
		if len(agentFiles) > 0 {
			fmt.Printf("  Generated %s\n", strings.Join(agentFiles, ", "))
		}
		fmt.Println()
		fmt.Println("  Recovery Key (write this down and keep it safe!):")
		fmt.Println("  +--------------------------------------------------+")
		words := splitMnemonicWords(mnemonic)
		if len(words) >= 12 {
			fmt.Printf("  |   %-47s|\n", joinWords(words[:6]))
			fmt.Printf("  |   %-47s|\n", joinWords(words[6:12]))
		}
		fmt.Println("  |                                                  |")
		fmt.Println("  |   If you forget your Master Password,            |")
		fmt.Println("  |   this is the ONLY way to recover.               |")
		fmt.Println("  +--------------------------------------------------+")
		fmt.Println()
		fmt.Printf("  Project %q initialized.\n", projectName)
		fmt.Println("  Default environment \"default\" created.")
		fmt.Println()
		fmt.Println("  Next: tene set KEY VALUE to add your first secret.")
		fmt.Println()
		fmt.Println("  Tip: No server needed. Your secrets stay on this device.")
		fmt.Println("       AI agents will automatically use tene.")
		fmt.Println()
		fmt.Println("  Docs:  https://tene.sh")
	} else {
		fmt.Println("Created .tene/vault.db")
		if len(agentFiles) > 0 {
			fmt.Printf("Generated %s\n", strings.Join(agentFiles, ", "))
		}
		fmt.Printf("Recovery Key: %s\n", mnemonic)
	}

	return nil
}

func splitMnemonicWords(mnemonic string) []string {
	var words []string
	for _, w := range splitString(mnemonic) {
		if w != "" {
			words = append(words, w)
		}
	}
	return words
}

func splitString(s string) []string {
	result := []string{}
	current := ""
	for _, c := range s {
		if c == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func joinWords(words []string) string {
	result := ""
	for i, w := range words {
		if i > 0 {
			result += " "
		}
		result += w
	}
	return result
}

// addAgentFiles generates only the specified agent context files when vault already exists.
func addAgentFiles(dir string, names []string) error {
	gen := claudemd.NewGenerator(dir)
	created, err := gen.GenerateSelected(names)
	if err != nil {
		return fmt.Errorf("failed to generate agent files: %w", err)
	}

	if flagJSON {
		return printJSON(map[string]any{
			"ok":         true,
			"agentFiles": created,
		})
	}

	if len(created) == 0 {
		fmt.Println("  Agent files already exist. Nothing to do.")
	} else {
		fmt.Printf("  Generated %s\n", strings.Join(created, ", "))
	}
	return nil
}

// selectedAgents returns agent names from init flags. Empty means all.
func selectedAgents() []string {
	var names []string
	if initFlagClaude {
		names = append(names, "claude")
	}
	if initFlagCursor {
		names = append(names, "cursor")
	}
	if initFlagWindsurf {
		names = append(names, "windsurf")
	}
	if initFlagGemini {
		names = append(names, "gemini")
	}
	if initFlagCodex {
		names = append(names, "codex")
	}
	return names
}
