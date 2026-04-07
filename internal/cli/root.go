package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	teneerr "github.com/tomo-kay/tene/internal/errors"
	"github.com/tomo-kay/tene/internal/keychain"
	"github.com/tomo-kay/tene/internal/vault"
	"golang.org/x/term"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// SetVersion sets build-time version information.
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

// App holds the dependencies needed for CLI execution.
type App struct {
	Vault    *vault.Vault
	Keychain keychain.KeyStore
	Dir      string // project directory
	Env      string // active environment
	JSON     bool   // --json flag
	Quiet    bool   // --quiet flag
}

// Global flags
var (
	flagJSON       bool
	flagQuiet      bool
	flagEnv        string
	flagDir        string
	flagNoColor    bool
	flagNoKeychain bool
)

var rootCmd = &cobra.Command{
	Use:     "tene",
	Short:   "Agentic Secret Runtime -- local-first encrypted secret management",
	Version: version,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Minimal output (errors only)")
	rootCmd.PersistentFlags().StringVarP(&flagEnv, "env", "e", "", "Environment name (default: active environment)")
	rootCmd.PersistentFlags().StringVar(&flagDir, "dir", "", "Project directory (default: current directory)")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVar(&flagNoKeychain, "no-keychain", false, "Do not use OS Keychain (CI/CD)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(passwdCmd)
	rootCmd.AddCommand(recoverCmd)
	rootCmd.AddCommand(newSyncCmd())
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)

	// Cloud commands
	rootCmd.AddCommand(newLoginCmd())
	rootCmd.AddCommand(newLogoutCmd())
	rootCmd.AddCommand(newPushCmd())
	rootCmd.AddCommand(newPullCmd())
	rootCmd.AddCommand(newBillingCmd())
	rootCmd.AddCommand(newTeamCmd())
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// --- Helper functions ---

func resolveDir() string {
	if flagDir != "" {
		return flagDir
	}
	dir, _ := os.Getwd()
	return dir
}

func resolveEnv(app *App) string {
	if flagEnv != "" {
		return flagEnv
	}
	if app.Env != "" {
		return app.Env
	}
	return "default"
}

func loadApp() (*App, error) {
	dir := resolveDir()
	vaultPath := filepath.Join(dir, ".tene", "vault.db")

	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return nil, teneerr.ErrVaultNotFound
	}

	v, err := vault.New(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open vault: %w", err)
	}

	var ks keychain.KeyStore
	if flagNoKeychain {
		home, _ := os.UserHomeDir()
		ks = keychain.NewFileStore(filepath.Join(home, ".tene", "keyfile"))
	} else {
		ks = keychain.NewStore(dir)
	}

	env, err := v.GetActiveEnvironment()
	if err != nil {
		env = "default"
	}

	return &App{
		Vault:    v,
		Keychain: ks,
		Dir:      dir,
		Env:      env,
		JSON:     flagJSON,
		Quiet:    flagQuiet,
	}, nil
}

func loadOrPromptMasterKey(app *App) ([]byte, error) {
	// Try keychain first
	key, err := app.Keychain.Load()
	if err == nil {
		return key, nil
	}

	// Try TENE_MASTER_PASSWORD env var
	if pw := os.Getenv("TENE_MASTER_PASSWORD"); pw != "" {
		return deriveMasterKeyFromPassword(app, pw)
	}

	// Interactive prompt
	if !isTerminal() {
		return nil, teneerr.ErrInteractiveRequired
	}

	fmt.Fprint(os.Stderr, "Enter Master Password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}

	return deriveMasterKeyFromPassword(app, string(password))
}

func deriveMasterKeyFromPassword(app *App, password string) ([]byte, error) {
	// Load salt from vault meta
	saltB64, err := app.Vault.GetMeta("kdf_salt")
	if err != nil {
		return nil, fmt.Errorf("failed to read KDF salt: %w", err)
	}

	salt, err := decodeBase64(saltB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode KDF salt: %w", err)
	}

	masterKey, err := deriveMasterKey(password, salt)
	if err != nil {
		return nil, err
	}

	return masterKey, nil
}

func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}


// envOrDefault returns the environment variable value or a fallback default.
// Used for CLI flag defaults that should respect tene-injected env vars.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
