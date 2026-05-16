package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/agent-kay-it/tene/pkg/crypto"
	"github.com/agent-kay-it/tene/internal/sync"
)

func newPullCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull encrypted vault from Tene Cloud",
		Long: `Download and decrypt the remote vault using Sync Envelope (L2).
Requires Pro plan and active login (tene login).

Creates a backup of the current vault.db before overwriting.
Use --force to skip backup and overwrite directly.`,
		Example: `  # Pull vault from cloud
  tene pull

  # Pull specific environment
  tene pull --env prod`,
		RunE: runPull,
	}
	cmd.Flags().Bool("force", false, "Force pull, overwriting local vault without backup")
	cmd.Flags().String("api-url", envOrDefault("API_URL", "https://api.tene.sh"), "Tene Cloud API base URL")
	return cmd
}

func runPull(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	apiURL, _ := cmd.Flags().GetString("api-url")

	// Check auth
	token, err := loadAuthToken()
	if err != nil || token == "" {
		return fmt.Errorf("not logged in. Run 'tene login' first")
	}

	// Pre-check plan from JWT (client-side, fail-fast)
	if err := checkProPlan(token); err != nil {
		return err
	}

	// Load local vault
	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	// Get master key for Sync Envelope decryption
	masterKey, err := loadOrPromptMasterKey(app)
	if err != nil {
		return fmt.Errorf("pull: %w", err)
	}
	defer crypto.ZeroBytes(masterKey)

	env := resolveEnv(app)
	projectName := filepath.Base(app.Dir)

	// Load vault ID from sync state
	vaultID, err := loadVaultID(app.Dir)
	if err != nil {
		return fmt.Errorf("pull: no vault ID found. Run 'tene push' first to initialize cloud sync")
	}

	if !flagQuiet {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Pulling vault '%s' (env: %s)...\n", projectName, env)
	}

	engine := sync.NewEngine()
	result, err := engine.Pull(cmd.Context(), sync.PullOptions{
		APIBaseURL:  apiURL,
		AccessToken: token,
		VaultID:     vaultID,
		ProjectName: projectName,
		Environment: env,
		VaultDBPath: filepath.Join(app.Dir, ".tene", "vault.db"),
		MasterKey:   masterKey,
		Force:       force,
	})
	if err != nil {
		return fmt.Errorf("pull: %w", err)
	}

	if flagJSON {
		return printJSON(result)
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ Pulled v%d\n", result.Version)
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "    Hash: %s\n", result.Hash[:16]+"...")
	return nil
}

func loadVaultID(projectDir string) (string, error) {
	statePath := filepath.Join(projectDir, ".tene", "sync_state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		return "", err
	}
	var state map[string]any
	if err := json.Unmarshal(data, &state); err != nil {
		return "", err
	}
	id, ok := state["vault_id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("vault_id not found in sync state")
	}
	return id, nil
}
