package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/agent-kay-it/tene/pkg/crypto"
	teneerr "github.com/agent-kay-it/tene/pkg/errors"
)

var runCmd = &cobra.Command{
	Use:   "run -- COMMAND [ARGS...]",
	Short: "Run a command with secrets injected as environment variables",
	Example: `  # Run with injected secrets
  tene run -- npm start

  # Run with local environment secrets
  tene run --env local -- go run ./main.go

  # Run with specific environment
  tene run --env prod -- ./my-app`,
	RunE: runRun,
	// Note: We manually parse --env before "--" since DisableFlagParsing
	// would prevent cobra from parsing it, and not disabling it would
	// cause cobra to try parsing flags after "--".
	DisableFlagParsing: true,
}

func runRun(cmd *cobra.Command, args []string) error {
	// Manually parse --env flag before "--" since DisableFlagParsing is true
	parseFlagsBeforeDash(args)

	// Parse args after "--"
	cmdArgs := extractArgsAfterDash(args)
	if len(cmdArgs) == 0 {
		return teneerr.New("COMMAND_NOT_FOUND", "No command specified. Usage: tene run -- <command>", 1)
	}

	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

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
			return teneerr.ErrCommandNotFound(execErr.Name)
		}
		return err
	}
	return nil
}

// parseFlagsBeforeDash manually parses known flags (--env, --json, --quiet)
// from args before the "--" separator. This is needed because run command uses
// DisableFlagParsing to avoid consuming flags after "--" that belong to the child command.
func parseFlagsBeforeDash(args []string) {
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			break
		}
		if (args[i] == "--env" || args[i] == "-e") && i+1 < len(args) {
			flagEnv = args[i+1]
			i++ // skip value
		}
		if args[i] == "--json" {
			flagJSON = true
		}
		if args[i] == "--quiet" || args[i] == "-q" {
			flagQuiet = true
		}
	}
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
