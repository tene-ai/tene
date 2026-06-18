package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	teneerr "github.com/tene-ai/tene/pkg/errors"
)

var deleteFlagForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete KEY",
	Short: "Delete a secret",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteFlagForce, "force", false, "Skip confirmation prompt")
}

func runDelete(cmd *cobra.Command, args []string) error {
	keyName := args[0]

	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	env := resolveEnv(app)

	// Check if secret exists
	exists, err := app.Vault.SecretExists(keyName, env)
	if err != nil {
		return err
	}
	if !exists {
		return teneerr.ErrSecretNotFound(keyName, env)
	}

	// Confirm. Sprint v1014-rc1-qa-fixes / FX2 (invariant I-12):
	// promptConfirm is now fail-closed on non-TTY, so we must surface
	// the refusal as a non-zero exit. Previously runDelete returned nil
	// after promptConfirm == false, which combined with the old
	// "non-TTY defaults to yes" gave CI/CD pipelines a green build even
	// when the secret was untouched — a confusing and brittle contract.
	if !deleteFlagForce {
		msg := fmt.Sprintf("Delete secret %q from %q?", keyName, env)
		if !promptConfirm(msg) {
			if !flagQuiet {
				fmt.Println("Cancelled.")
			}
			return teneerr.New("CONFIRMATION_REQUIRED",
				fmt.Sprintf("delete %q cancelled: pass --force to confirm in a non-interactive shell", keyName), 1)
		}
	}

	if err := app.Vault.DeleteSecret(keyName, env); err != nil {
		return err
	}

	if flagJSON {
		return printJSON(map[string]any{
			"ok":          true,
			"name":        keyName,
			"environment": env,
			"deleted":     true,
		})
	}

	if !flagQuiet {
		fmt.Printf("%s deleted.\n", keyName)
	}
	return nil
}
