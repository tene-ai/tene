package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/tomo-kay/tene/internal/crypto"
)

var runCmd = &cobra.Command{
	Use:                "run -- COMMAND [ARGS...]",
	Short:              "Run a command with secrets injected as environment variables",
	DisableFlagParsing: true,
	RunE:               runRun,
}

func runRun(cmd *cobra.Command, args []string) error {
	// Parse args after "--"
	cmdArgs := extractArgsAfterDash(args)
	if len(cmdArgs) == 0 {
		return fmt.Errorf("No command specified. Usage: tene run -- <command>")
	}

	app, err := loadApp()
	if err != nil {
		return err
	}
	defer app.Vault.Close()

	masterKey, err := loadOrPromptMasterKey(app)
	if err != nil {
		return err
	}

	encKey, err := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)
	if err != nil {
		return err
	}

	env := resolveEnv(app)
	allSecrets, err := app.Vault.GetAllSecrets(env)
	if err != nil {
		return err
	}

	// Decrypt all secrets and build environment
	environ := os.Environ()
	injectedCount := 0
	for name, encVal := range allSecrets {
		ct, err := decodeBase64(encVal)
		if err != nil {
			return fmt.Errorf("failed to decode secret %s: %w", name, err)
		}
		pt, err := crypto.Decrypt(encKey, ct, []byte(name))
		if err != nil {
			return fmt.Errorf("failed to decrypt %s: %w", name, err)
		}
		environ = append(environ, fmt.Sprintf("%s=%s", name, string(pt)))
		injectedCount++
	}

	// Output info
	if flagJSON {
		info := map[string]any{
			"injectedCount": injectedCount,
			"environment":   env,
			"command":       cmdArgs[0],
		}
		data, _ := json.Marshal(info)
		fmt.Fprintln(os.Stderr, string(data))
	} else if !flagQuiet {
		if injectedCount == 0 {
			fmt.Fprintf(os.Stderr, "Warning: No secrets found in %q. Running without injected secrets.\n", env)
		} else {
			fmt.Fprintf(os.Stderr, "Injecting %d secrets into environment...\n", injectedCount)
			fmt.Fprintf(os.Stderr, "Starting: %s\n", cmdArgs[0])
		}
	}

	// Audit log
	_ = app.Vault.AddAuditLog("secrets.inject", "", fmt.Sprintf("count=%d,env=%s,cmd=%s", injectedCount, env, cmdArgs[0]))

	// Execute command
	c := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	c.Env = environ
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		if execErr, ok := err.(*exec.Error); ok {
			return fmt.Errorf("Command %q not found.", execErr.Name)
		}
		return err
	}
	return nil
}

func extractArgsAfterDash(args []string) []string {
	for i, arg := range args {
		if arg == "--" {
			return args[i+1:]
		}
	}
	// If no --, treat all args as the command
	// But first filter out any flags we know about
	var result []string
	for _, arg := range args {
		if arg == "--json" || arg == "--quiet" || arg == "-q" {
			continue
		}
		if len(arg) > 0 && arg[0] == '-' && len(result) == 0 {
			// Skip unknown flags before command
			continue
		}
		result = append(result, arg)
	}
	return result
}
