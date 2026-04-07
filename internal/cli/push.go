package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/tomo-kay/tene/internal/crypto"
	"github.com/tomo-kay/tene/internal/sync"
)

func newPushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push encrypted vault to Tene Cloud",
		Long: `Encrypt the local vault with Sync Envelope (L2) and upload to cloud.
Requires Pro plan and active login (tene login).

Uses optimistic locking: if the remote version is newer, you must pull first
or use --force to overwrite.`,
		Example: `  # Push vault to cloud
  tene push

  # Push specific environment
  tene push --env prod`,
		RunE: runPush,
	}
	cmd.Flags().Bool("force", false, "Force push, overwriting remote version")
	cmd.Flags().String("api-url", envOrDefault("API_URL", "https://api.tene.sh"), "Tene Cloud API base URL")
	return cmd
}

func runPush(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	apiURL, _ := cmd.Flags().GetString("api-url")

	// Check auth
	token, err := loadAuthToken()
	if err != nil || token == "" {
		return fmt.Errorf("not logged in. Run 'tene login' first")
	}

	// Load local vault
	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	// Get master key for Sync Envelope encryption
	masterKey, err := loadOrPromptMasterKey(app)
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	defer crypto.ZeroBytes(masterKey)

	env := resolveEnv(app)
	projectName := filepath.Base(app.Dir)

	// Get or create vault ID
	vaultID, err := getOrCreateVaultID(app.Dir, apiURL, token, projectName)
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}

	if !flagQuiet {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Pushing vault '%s' (env: %s)...\n", projectName, env)
	}

	engine := sync.NewEngine()
	result, err := engine.Push(cmd.Context(), sync.PushOptions{
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
		return fmt.Errorf("push: %w", err)
	}

	if flagJSON {
		return printJSON(result)
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ Pushed v%d (%d bytes)\n", result.Version, result.Size)
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "    Hash: %s\n", result.Hash[:16]+"...")
	return nil
}

// getOrCreateVaultID retrieves or creates a vault record on the server.
func getOrCreateVaultID(projectDir, apiURL, token, projectName string) (string, error) {
	// Try to load from local sync state first
	statePath := filepath.Join(projectDir, ".tene", "sync_state.json")
	if data, err := os.ReadFile(statePath); err == nil {
		var state map[string]any
		if json.Unmarshal(data, &state) == nil {
			if id, ok := state["vault_id"].(string); ok && id != "" {
				return id, nil
			}
		}
	}

	// Create via API
	return createVaultViaAPI(apiURL, token, projectName)
}

func createVaultViaAPI(apiURL, token, projectName string) (string, error) {
	bodyBytes, _ := json.Marshal(map[string]string{"project_name": projectName})
	req, err := http.NewRequest(http.MethodPost, apiURL+"/api/v1/vaults", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("create vault: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var apiResp struct {
		OK      bool   `json:"ok"`
		Error   string `json:"error"`
		Message string `json:"message"`
		Data    struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("create vault: decode: %w", err)
	}
	if !apiResp.OK {
		return "", fmt.Errorf("create vault: %s", apiErrMsg(apiResp.Error, apiResp.Message, resp.StatusCode))
	}

	return apiResp.Data.ID, nil
}
