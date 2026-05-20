package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/agent-kay-it/tene/internal/vault"
	"github.com/agent-kay-it/tene/internal/vaultcfg"
	"github.com/agent-kay-it/tene/pkg/crypto"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [SUBCOMMAND]",
	Short: "Inspect or run vault.db migrations",
	Long: `Inspect or run vault.db migrations.

Without arguments, prints the current vault schema version and whether
all forward migrations have been applied (this is the normal state after
opening the vault with any modern 'tene' binary).

Subcommands:
  fill-previews   Derive a short preview substring for every secret whose
                  preview column is currently empty. Requires master
                  password unlock (the operation decrypts each secret to
                  derive its preview). Idempotent: secrets that already
                  have a non-empty preview are skipped.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMigrate,
}

func runMigrate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runMigrateStatus()
	}
	switch args[0] {
	case "fill-previews":
		return runMigrateFillPreviews()
	default:
		return fmt.Errorf("unknown migrate subcommand %q (valid: fill-previews)", args[0])
	}
}

func runMigrateStatus() error {
	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	version, err := app.Vault.GetMeta("schema_version")
	if err != nil {
		return fmt.Errorf("failed to read schema_version: %w", err)
	}

	if flagJSON {
		n, _ := strconv.Atoi(version)
		return printJSON(map[string]any{
			"ok":            true,
			"schemaVersion": n,
		})
	}
	fmt.Printf("vault schema_version = %s\n", version)
	return nil
}

func runMigrateFillPreviews() error {
	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	settings, err := vaultcfg.GetPreviewSettings(app.Vault)
	if err != nil {
		return fmt.Errorf("failed to read preview settings: %w", err)
	}

	// If the user has previews disabled, fill-previews would be a no-op
	// that the user almost certainly did not intend. Refuse loudly rather
	// than silently doing nothing: a CI script that runs migrate in
	// fill-previews mode then sees blank previews would be very confusing.
	if !settings.Enabled {
		return fmt.Errorf(
			"preview.enabled is currently false; fill-previews would have nothing to derive. " +
				"Run `tene config preview.enabled=true` first, or pass front+back > 0 manually")
	}
	if settings.Front == 0 && settings.Back == 0 {
		return fmt.Errorf("preview.front and preview.back are both 0; nothing to derive")
	}

	// Listing all environments and iterating them is the only way to
	// cover the entire vault: secrets are scoped by environment and
	// ListSecretsForBackfill takes an env argument.
	envs, err := app.Vault.ListEnvironments()
	if err != nil {
		return fmt.Errorf("failed to list environments: %w", err)
	}
	// Always include "default" even if no Environment row was ever inserted
	// (early `tene init` did not always seed it explicitly).
	envNames := envNamesIncludingDefault(envs)

	// Determine total work before unlocking the vault. This lets us avoid
	// the password prompt entirely when nothing needs backfilling.
	totalCandidates := 0
	candidatesByEnv := make(map[string][]vault.SecretBackfill, len(envNames))
	for _, env := range envNames {
		cands, err := app.Vault.ListSecretsForBackfill(env)
		if err != nil {
			return fmt.Errorf("failed to list backfill candidates in env %q: %w", env, err)
		}
		candidatesByEnv[env] = cands
		totalCandidates += len(cands)
	}

	if totalCandidates == 0 {
		if flagJSON {
			return printJSON(map[string]any{
				"ok":      true,
				"filled":  0,
				"skipped": 0,
				"message": "all previews are already populated",
			})
		}
		if !flagQuiet {
			fmt.Println("All previews are already populated. Nothing to do.")
		}
		return nil
	}

	// Unlock once for the whole sweep.
	masterKey, err := loadOrPromptMasterKey(app)
	if err != nil {
		return err
	}
	defer crypto.ZeroBytes(masterKey)

	encKey, err := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)
	if err != nil {
		return err
	}
	defer crypto.ZeroBytes(encKey)

	totalFilled := 0
	totalSkipped := 0
	for _, env := range envNames {
		cands := candidatesByEnv[env]
		if len(cands) == 0 {
			continue
		}
		envFilled := 0
		envSkipped := 0
		for _, c := range cands {
			ct, err := decodeBase64(c.EncryptedValue)
			if err != nil {
				envSkipped++
				if !flagQuiet && !flagJSON {
					fmt.Fprintf(os.Stderr, "warning: skip %s (base64 decode): %v\n", c.Name, err)
				}
				continue
			}
			plaintext, err := crypto.Decrypt(encKey, ct, []byte(c.Name))
			if err != nil {
				envSkipped++
				if !flagQuiet && !flagJSON {
					fmt.Fprintf(os.Stderr, "warning: skip %s (decrypt failed): %v\n", c.Name, err)
				}
				continue
			}

			preview := crypto.DerivePreview(string(plaintext), settings.Front, settings.Back)
			// Wipe plaintext immediately; we only needed it for the
			// derivation above and decryption already returned a fresh
			// allocation.
			for i := range plaintext {
				plaintext[i] = 0
			}

			if err := app.Vault.UpdateSecretPreview(c.Name, env, preview); err != nil {
				envSkipped++
				if !flagQuiet && !flagJSON {
					fmt.Fprintf(os.Stderr, "warning: skip %s (update failed): %v\n", c.Name, err)
				}
				continue
			}
			envFilled++
		}
		totalFilled += envFilled
		totalSkipped += envSkipped
		if !flagQuiet && !flagJSON {
			fmt.Printf("Filled previews for %d/%d secrets in env=%s.\n",
				envFilled, len(cands), env)
		}
	}

	if flagJSON {
		return printJSON(map[string]any{
			"ok":      true,
			"filled":  totalFilled,
			"skipped": totalSkipped,
		})
	}
	if !flagQuiet {
		fmt.Printf("Done. Filled %d preview(s); skipped %d.\n", totalFilled, totalSkipped)
	}
	return nil
}

// envNamesIncludingDefault extracts environment names from an Environment
// slice, guaranteeing "default" appears even if the table was empty (a
// quirk of older `tene init` flows that did not always seed it).
func envNamesIncludingDefault(envs []vault.Environment) []string {
	out := make([]string, 0, len(envs)+1)
	seenDefault := false
	for _, e := range envs {
		out = append(out, e.Name)
		if e.Name == "default" {
			seenDefault = true
		}
	}
	if !seenDefault {
		out = append(out, "default")
	}
	return out
}
