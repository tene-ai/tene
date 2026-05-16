package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	teneerr "github.com/agent-kay-it/tene/pkg/errors"
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

	// Confirm
	if !deleteFlagForce {
		msg := fmt.Sprintf("Delete secret %q from %q?", keyName, env)
		if !promptConfirm(msg) {
			if !flagQuiet {
				fmt.Println("Cancelled.")
			}
			return nil
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
