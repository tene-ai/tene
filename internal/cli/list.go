package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/agent-kay-it/tene/internal/vaultcfg"
	"github.com/agent-kay-it/tene/pkg/domain"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List secret names with short previews (no password)",
	Long: `List every secret in the active environment without unlocking the vault.

Output shows three columns: NAME, PREVIEW, UPDATED.

PREVIEW is the short plaintext substring stored alongside the ciphertext
(see ` + "`tene config preview.*`" + `). Default is the last 4 chars of the value
("…aBc1"); the value itself is never decrypted by this command. When
previews are disabled, the column shows "-" and a footer hints how to
re-enable.

This is the metadata-tier read path: no Argon2id, no decrypt, no keychain
probe.`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	env := resolveEnv(app)

	// F3 path: read the preview column directly. ListSecretMetadata never
	// touches encrypted_value (invariant I-1). No unlock, no Argon2id.
	metas, err := app.Vault.ListSecretMetadata(env)
	if err != nil {
		return err
	}

	// Resolve preview config so we can render the appropriate footer hint.
	// The config does NOT change what is shown for already-stored previews
	// (those are persisted at write time); it only governs the hint text.
	settings, cfgErr := vaultcfg.GetPreviewSettings(app.Vault)
	previewEnabled := vaultcfg.DefaultPreviewEnabled
	if cfgErr == nil {
		previewEnabled = settings.Enabled
	}
	// cfgErr is intentionally swallowed -- a vault that cannot read its own
	// meta should still be able to list secret names. The fallback is the
	// security-conscious default (preview enabled = true; the footer hint
	// won't fire spuriously).

	if flagJSON {
		return renderListJSON(app, env, metas)
	}
	return renderListText(app, env, metas, previewEnabled)
}

// renderListJSON emits the always-string preview contract (Q2): every
// element has a "preview" key, value is the stored string or "" (never
// null, never absent).
func renderListJSON(app *App, env string, metas []domain.VaultKeyMeta) error {
	projectName, _ := app.Vault.GetMeta("project_name")

	type secretItem struct {
		Name      string `json:"name"`
		Preview   string `json:"preview"`
		Version   int    `json:"version"`
		UpdatedAt string `json:"updatedAt"`
	}
	items := make([]secretItem, 0, len(metas))
	for _, m := range metas {
		items = append(items, secretItem{
			Name:      m.Name,
			Preview:   m.Preview,
			Version:   m.Version,
			UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
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

// renderListText prints the 3-column table to stdout and, when relevant,
// a single footer hint to stderr explaining the empty preview state.
//
// Sprint v1014-rc1-qa-fixes / FX6 (B8): when --quiet is set, the entire
// text-mode output is suppressed. Scripts that need a parseable result
// should use --json instead; --quiet means "errors only" per the global
// flag's documentation and the convention `tene set --quiet` already
// follows.
func renderListText(app *App, env string, metas []domain.VaultKeyMeta, previewEnabled bool) error {
	if flagQuiet {
		return nil
	}

	if len(metas) == 0 {
		fmt.Printf("No secrets in %q environment. Use \"tene set KEY VALUE\" to add one.\n", env)
		return nil
	}

	projectName, _ := app.Vault.GetMeta("project_name")
	if projectName == "" {
		projectName = "unknown"
	}

	fmt.Printf("  Project: %s (%s)\n\n", projectName, env)
	fmt.Printf("  %-30s %-16s %s\n", "NAME", "PREVIEW", "UPDATED")

	emptyPreviewCount := 0
	for _, m := range metas {
		preview := m.Preview
		if preview == "" {
			preview = "-"
			emptyPreviewCount++
		}
		fmt.Printf("  %-30s %-16s %s\n", m.Name, preview, formatTimeAgo(m.UpdatedAt))
	}

	// FX6 B10: pluralise "secret" / "secrets" properly when the count
	// drops to one. Prior wording was always "%d secrets".
	fmt.Printf("\n  %d %s in %q environment\n", len(metas), pluralize(len(metas), "secret"), env)

	// Footer hints. Stderr so scripts that parse stdout (NAME column at
	// position 1 — design §F3 regression note) are unaffected.
	switch {
	case !previewEnabled:
		fmt.Fprintln(os.Stderr, "  Note: previews disabled. Run `tene config preview.enabled=true` to re-enable.")
	case emptyPreviewCount > 0 && emptyPreviewCount == len(metas):
		// All previews empty AND preview is enabled: this is the legacy
		// v1-vault-after-auto-migrate state. Suggest the one-time
		// backfill command rather than the config toggle.
		fmt.Fprintln(os.Stderr, "  Note: previews are empty. Run `tene migrate fill-previews` (one-time, requires password).")
	}

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
