package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tomo-kay/tene/pkg/crypto"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets (values masked)",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	env := resolveEnv(app)
	secrets, err := app.Vault.ListSecrets(env)
	if err != nil {
		return err
	}

	if flagJSON {
		// Need master key to decrypt for preview
		masterKey, err := loadOrPromptMasterKey(app)
		if err != nil {
			return err
		}
		encKey, _ := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)

		projectName, _ := app.Vault.GetMeta("project_name")

		type secretItem struct {
			Name      string `json:"name"`
			Preview   string `json:"preview"`
			Version   int    `json:"version"`
			UpdatedAt string `json:"updatedAt"`
		}
		items := make([]secretItem, 0, len(secrets))
		for _, s := range secrets {
			preview := "*****"
			if ct, err := decodeBase64(s.EncryptedValue); err == nil {
				if pt, err := crypto.Decrypt(encKey, ct, []byte(s.Name)); err == nil {
					preview = maskValue(string(pt))
				}
			}
			items = append(items, secretItem{
				Name:      s.Name,
				Preview:   preview,
				Version:   s.Version,
				UpdatedAt: s.UpdatedAt.Format(time.RFC3339),
			})
		}
		return printJSON(map[string]any{
			"ok":          true,
			"project":     projectName,
			"environment": env,
			"secrets":     items,
			"count":       len(items),
		})
	}

	if len(secrets) == 0 {
		fmt.Printf("No secrets in %q environment. Use \"tene set KEY VALUE\" to add one.\n", env)
		return nil
	}

	// Get project name for display
	projectName, _ := app.Vault.GetMeta("project_name")
	if projectName == "" {
		projectName = "unknown"
	}

	// Decrypt for preview
	masterKey, err := loadOrPromptMasterKey(app)
	if err != nil {
		return err
	}
	encKey, _ := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)

	fmt.Printf("  Project: %s (%s)\n\n", projectName, env)
	fmt.Printf("  %-30s %-16s %s\n", "NAME", "VALUE", "UPDATED")

	for _, s := range secrets {
		preview := "*****"
		if ct, err := decodeBase64(s.EncryptedValue); err == nil {
			if pt, err := crypto.Decrypt(encKey, ct, []byte(s.Name)); err == nil {
				preview = maskValue(string(pt))
			}
		}

		updated := formatTimeAgo(s.UpdatedAt)
		fmt.Printf("  %-30s %-16s %s\n", s.Name, preview, updated)
	}

	fmt.Printf("\n  %d secrets in %q environment\n", len(secrets), env)
	return nil
}

func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	diff := time.Since(t)
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hrs := int(diff.Hours())
		if hrs == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hrs)
	default:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
