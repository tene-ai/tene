package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current project and vault status",
	RunE:  runWhoami,
}

func runWhoami(cmd *cobra.Command, args []string) error {
	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	env := resolveEnv(app)
	projectName, _ := app.Vault.GetMeta("project_name")
	createdAt, _ := app.Vault.GetMeta("created_at")
	count, _ := app.Vault.CountSecrets(env)

	keychainStatus := "active"
	keychainProvider := runtime.GOOS + " Keychain"
	if !app.Keychain.Exists() {
		keychainStatus = "inactive"
	}
	if flagNoKeychain {
		keychainStatus = "disabled"
		keychainProvider = "file fallback"
	}

	switch runtime.GOOS {
	case "darwin":
		keychainProvider = "macOS Keychain"
	case "linux":
		keychainProvider = "Linux Secret Service"
	case "windows":
		keychainProvider = "Windows Credential Vault"
	}

	if flagJSON {
		return printJSON(map[string]any{
			"ok":               true,
			"project":          projectName,
			"vault":            ".tene/vault.db",
			"environment":      env,
			"secretCount":      count,
			"keychainStatus":   keychainStatus,
			"keychainProvider": keychainProvider,
			"createdAt":        createdAt,
			"agents":           []string{"claude"},
			"vaultVersion":     1,
		})
	}

	fmt.Printf("  Project: %s\n", projectName)
	fmt.Printf("  Vault: .tene/vault.db\n")
	fmt.Printf("  Environment: %s (active)\n", env)
	fmt.Printf("  Secrets: %d\n", count)
	fmt.Printf("  Keychain: %s (%s)\n", keychainProvider, keychainStatus)
	if createdAt != "" {
		fmt.Printf("  Created: %s\n", createdAt[:10])
	}
	fmt.Printf("  Agents: claude\n")
	return nil
}
