package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	teneerr "github.com/tene-ai/tene/pkg/errors"
	"github.com/tene-ai/tene/internal/vault"
)

var envCmd = &cobra.Command{
	Use:   "env [name|list|create|delete]",
	Short: "Manage environments",
	RunE:  runEnv,
}

var envCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Create a new environment",
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvCreate,
}

// envDeleteFlagForce is the env-delete-scoped equivalent of deleteFlagForce
// (delete.go). They are kept separate because the two verbs live in
// distinct cobra subtrees and a single shared variable would let one verb
// silently affect the next verb in the same process, which test harnesses
// have already tripped over once. Sprint v1014-rc1-qa-fixes / FX2.
var envDeleteFlagForce bool

var envDeleteCmd = &cobra.Command{
	Use:   "delete NAME",
	Short: "Delete an environment",
	Args:  cobra.ExactArgs(1),
	RunE:  runEnvDelete,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environments",
	RunE:  runEnvList,
}

func init() {
	envCmd.AddCommand(envCreateCmd)
	envCmd.AddCommand(envDeleteCmd)
	envCmd.AddCommand(envListCmd)
	envDeleteCmd.Flags().BoolVar(&envDeleteFlagForce, "force", false, "Skip confirmation prompt")
}

func runEnv(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runEnvList(cmd, args)
	}

	// Switch to named environment
	envName := args[0]

	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	// Check if environment exists
	exists, err := app.Vault.EnvironmentExists(envName)
	if err != nil {
		return err
	}
	if !exists {
		return teneerr.ErrEnvironmentNotFound(envName)
	}

	previous, _ := app.Vault.GetActiveEnvironment()

	if err := app.Vault.SetActiveEnvironment(envName); err != nil {
		return err
	}

	// Update vault.json (non-critical; SQLite is primary source)
	vaultJSONPath := filepath.Join(app.Dir, ".tene", "vault.json")
	_ = vault.UpdateVaultJSONEnv(vaultJSONPath, envName)

	// Audit log
	_ = app.Vault.AddAuditLog("env.switch", envName, fmt.Sprintf("from=%s,to=%s", previous, envName))

	if flagJSON {
		return printJSON(map[string]any{
			"ok":       true,
			"previous": previous,
			"current":  envName,
		})
	}

	if !flagQuiet {
		fmt.Printf("Switched to %q environment.\n", envName)
	}
	return nil
}

func runEnvList(cmd *cobra.Command, args []string) error {
	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	envs, err := app.Vault.ListEnvironments()
	if err != nil {
		return err
	}

	if flagJSON {
		type envItem struct {
			Name        string `json:"name"`
			SecretCount int    `json:"secretCount"`
			IsActive    bool   `json:"isActive"`
		}
		active, _ := app.Vault.GetActiveEnvironment()
		items := make([]envItem, 0, len(envs))
		for _, e := range envs {
			count, _ := app.Vault.CountSecrets(e.Name)
			items = append(items, envItem{
				Name:        e.Name,
				SecretCount: count,
				IsActive:    e.IsActive,
			})
		}
		return printJSON(map[string]any{
			"ok":           true,
			"active":       active,
			"environments": items,
		})
	}

	fmt.Println("  Environments:")
	for _, e := range envs {
		count, _ := app.Vault.CountSecrets(e.Name)
		// FX6 B10 + B11: two explicit format strings remove the double
		// space rc1 produced for non-active envs ("  dev (  0 secrets)")
		// and pluralise "secret" correctly when the count is 1.
		if e.IsActive {
			fmt.Printf("  * %s (active, %d %s)\n", e.Name, count, pluralize(count, "secret"))
		} else {
			fmt.Printf("    %s (%d %s)\n", e.Name, count, pluralize(count, "secret"))
		}
	}
	return nil
}

func runEnvCreate(cmd *cobra.Command, args []string) error {
	envName := args[0]

	if err := validateEnvName(envName); err != nil {
		return err
	}

	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	if err := app.Vault.CreateEnvironment(envName); err != nil {
		return teneerr.ErrEnvironmentAlreadyExists(envName)
	}

	// Audit log
	_ = app.Vault.AddAuditLog("env.create", envName, "")

	if flagJSON {
		return printJSON(map[string]any{
			"ok":      true,
			"name":    envName,
			"created": true,
		})
	}

	if !flagQuiet {
		fmt.Printf("Environment %q created.\n", envName)
	}
	return nil
}

func runEnvDelete(cmd *cobra.Command, args []string) error {
	envName := args[0]

	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	// Cannot delete "default"
	if envName == "default" {
		return teneerr.ErrEnvironmentProtected("default", "It is the default environment.")
	}

	// Cannot delete active environment
	active, _ := app.Vault.GetActiveEnvironment()
	if envName == active {
		return teneerr.ErrEnvironmentProtected(envName, "Switch to another first.")
	}

	// Confirm. Sprint v1014-rc1-qa-fixes / FX2 (invariant I-12):
	// - if --force is passed, skip the prompt entirely.
	// - otherwise call promptConfirm which is fail-closed on non-TTY.
	// This is the read side of B2/B9: the previous code used
	// deleteFlagForce (the unrelated `tene delete KEY` flag) and silently
	// reached promptConfirm even when the user typed `tene env delete X
	// --force` (which produced "unknown flag: --force" instead of doing
	// what the user meant).
	if !envDeleteFlagForce {
		count, _ := app.Vault.CountSecrets(envName)
		msg := fmt.Sprintf("Delete environment %q and all its secrets?", envName)
		if count > 0 {
			msg = fmt.Sprintf("Delete environment %q and all %d secrets?", envName, count)
		}
		if !promptConfirm(msg) {
			if !flagQuiet {
				fmt.Println("Cancelled.")
			}
			// Treat refusal as an explicit error so CI/CD pipelines do
			// not interpret a non-deleted env as a successful run.
			// Exit non-zero via cobra by returning an error; the user
			// already saw the "Refusing to confirm..." stderr line
			// from promptConfirm.
			return teneerr.New("CONFIRMATION_REQUIRED",
				fmt.Sprintf("env delete %q cancelled: pass --force to confirm in a non-interactive shell", envName), 1)
		}
	}

	secretsRemoved, err := app.Vault.DeleteEnvironment(envName)
	if err != nil {
		return teneerr.ErrEnvironmentNotFound(envName)
	}

	// Audit log
	_ = app.Vault.AddAuditLog("env.delete", envName, fmt.Sprintf("secretsRemoved=%d", secretsRemoved))

	if flagJSON {
		return printJSON(map[string]any{
			"ok":             true,
			"name":           envName,
			"secretsRemoved": secretsRemoved,
		})
	}

	if !flagQuiet {
		// FX6 B10: pluralise "secret(s)" so single-secret deletes read
		// naturally ("1 secret removed" not "1 secrets removed").
		fmt.Printf("Environment %q deleted (%d %s removed).\n",
			envName, secretsRemoved, pluralize(secretsRemoved, "secret"))
	}
	return nil
}
