package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	teneerr "github.com/tomo-kay/tene/pkg/errors"
	"github.com/tomo-kay/tene/internal/vault"
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
		marker := " "
		active := ""
		if e.IsActive {
			marker = "*"
			active = " (active,"
		} else {
			active = " ("
		}
		fmt.Printf("  %s %s%s %d secrets)\n", marker, e.Name, active, count)
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

	// Confirm
	if !deleteFlagForce {
		count, _ := app.Vault.CountSecrets(envName)
		msg := fmt.Sprintf("Delete environment %q and all its secrets?", envName)
		if count > 0 {
			msg = fmt.Sprintf("Delete environment %q and all %d secrets?", envName, count)
		}
		if !promptConfirm(msg) {
			if !flagQuiet {
				fmt.Println("Cancelled.")
			}
			return nil
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
		fmt.Printf("Environment %q deleted (%d secrets removed).\n", envName, secretsRemoved)
	}
	return nil
}
