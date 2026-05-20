package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/agent-kay-it/tene/internal/auth"
	"github.com/agent-kay-it/tene/internal/keychain"
	"github.com/agent-kay-it/tene/internal/vault"
	teneerr "github.com/agent-kay-it/tene/pkg/errors"
	"github.com/spf13/cobra"
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
	rootCmd.Version = v
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
	Use:   "tene",
	Short: "Agentic Secret Runtime -- local-first encrypted secret management",
	Long: `tene is a local-first encrypted secret manager CLI. It encrypts your
secrets with XChaCha20-Poly1305 and injects them at runtime so AI agents
(Claude Code, Cursor, Windsurf, Gemini, Codex, Copilot) never see plaintext.

Typical workflow:

  tene init                                 # create encrypted vault
  tene set STRIPE_KEY sk_test_xxx           # store a secret
  tene run -- npm start                     # run with secrets injected

Resources:
  AI index:    https://tene.sh/llms.txt
  CLI ref:     https://tene.sh/cli
  Docs:        https://github.com/agent-kay-it/tene#readme
  Issues:      https://github.com/agent-kay-it/tene/issues
  Discussions: https://github.com/agent-kay-it/tene/discussions
`,
	Version:       version,
	SilenceErrors: true,
	SilenceUsage:  true,
	// F2 dispatcher: every CLI invocation flows through this hook so the
	// permission tier of the verb can be inspected centrally. Today the
	// hook is a guard rail (fails closed for any verb without a tier
	// declared in internal/auth.CommandTier); F4 will extend it to emit
	// the cli.<tier>.<verb> audit row, and F8 to fire the audit-log
	// size threshold warning.
	PersistentPreRunE: rootPersistentPreRunE,
}

// rootPersistentPreRunE runs before any subcommand RunE. It enforces the
// declarative permission tier policy: every verb must have an entry in
// internal/auth.CommandTier or the dispatch is refused, and it emits
// the F4 audit row recording the tier + verb of every invocation.
//
// Why fail closed here even though init() already calls auth.Validate()?
// init()'s panic catches the static "developer forgot to register a new
// command" case at binary startup, but cobra also allows runtime
// AddCommand() (e.g. plugin patterns, or tests that synthesize commands
// at runtime). The PreRunE check guarantees the invariant on every
// individual invocation regardless of how the verb got into the tree.
//
// F4 emission policy:
//
//   - The row is written here, BEFORE RunE, so that a command which
//     errors out partway through still leaves an audit trail of the
//     attempt. Operators investigating "did this verb run?" get the
//     truthful answer.
//
//   - For `tene init` the vault.db does not yet exist at this point —
//     emitCliAuditRow silently no-ops in that case and init.go writes
//     the row itself at the end of its RunE (see audit.go header).
//
//   - The existing per-verb audit rows (secret.read, vault.init,
//     env.switch, ...) are NOT touched: F4 is purely additive. One
//     invocation -> 1 cli.* row PLUS the legacy rows the verb already
//     emits inside its RunE.
//
//   - Audit write failures never block the command (audit.go error
//     policy).
//
// F2 deliberately keeps actual unlock as each subcommand's
// responsibility for now — that preserves the existing test suite's
// ability to drive RunE functions directly (cli_test.go,
// set_get_test.go, etc.). F8 will layer the audit-log size threshold
// warning at the END of this hook.
func rootPersistentPreRunE(cmd *cobra.Command, args []string) error {
	path := commandTierPath(cmd)
	if path == "" {
		// The bare root invocation (`tene` with no args) prints help; no
		// tier check is meaningful here. Cobra dispatches to help text
		// without ever entering a subcommand's RunE.
		return nil
	}

	tier, ok := auth.TierFor(path)
	if !ok {
		// Mirror the wording auth.Validate() uses so an operator sees the
		// same diagnostic at startup-panic time and at runtime.
		//
		// Privacy/safety: we fail closed BEFORE writing any audit row.
		// A rogue verb that was never declared in CommandTier must not
		// be able to forge an audit entry pretending to be a known tier.
		return fmt.Errorf(
			"internal: command %q has no PermLevel entry in internal/auth.CommandTier — refusing to dispatch",
			path,
		)
	}

	// F4 emission. Args are intentionally NOT passed through (privacy):
	// auditActionFor takes only the tier + verb, never the user-supplied
	// positional args or flag values.
	//
	// `tene init` is special: when this branch fires for init the vault
	// does not yet exist, so emitCliAuditRow silently no-ops and init's
	// RunE writes the F4 row itself once vault.db has been created.
	// See audit.go header + init.go step 13a.
	dir := resolveDir()
	emitCliAuditRow(dir, auditActionFor(tier, path))

	// F8 — audit log size threshold notice. Opens the vault read-only,
	// checks the audit_log byte estimate against the configured
	// threshold (default 50 MB, master-plan §11), and emits a one-line
	// stderr notice at most once per 24h per project. Failure is
	// silently swallowed — the dispatcher must never block the
	// command's primary work.
	//
	// We use a fresh vault handle (not loadApp()) for the same reason
	// F4's emitCliAuditRow does: the dispatcher runs BEFORE RunE,
	// loadApp would unnecessarily probe the keychain, and we only
	// need read-only SQL access. emitAuditThresholdHook opens the
	// vault, runs the check, and closes — completely independent of
	// whatever the verb's RunE does next.
	emitAuditThresholdHook(dir, flagQuiet)
	return nil
}

// emitAuditThresholdHook is the F8 entry point invoked from
// rootPersistentPreRunE. It exists as a thin wrapper around
// maybeEmitAuditThresholdWarning so the dispatcher can stay free of
// vault open/close concerns (mirrors emitCliAuditRow's shape).
//
// Why open a fresh vault here rather than thread App through PreRunE:
// the F2 hook explicitly avoids loadApp() so the dispatcher does not
// trigger keychain probes for metaread commands. Opening the vault
// directly (no keychain) keeps that contract while letting F8 read
// audit_log size + the audit.warnAtMB config row.
func emitAuditThresholdHook(projectDir string, quiet bool) {
	vaultPath := filepath.Join(projectDir, ".tene", "vault.db")
	if _, err := os.Stat(vaultPath); err != nil {
		// `tene init` path: vault.db not yet present — nothing to
		// size-check. Same skip logic emitCliAuditRow uses.
		return
	}
	v, err := vault.New(vaultPath)
	if err != nil {
		return
	}
	defer func() { _ = v.Close() }()
	maybeEmitAuditThresholdWarning(os.Stderr, v, projectDir, quiet)
}

// commandTierPath returns the cobra command path normalised to the
// internal/auth.CommandTier key format: the full CommandPath() minus
// the leading root name. For "tene env list" this returns "env list".
//
// Returns the empty string if cmd is the root itself.
func commandTierPath(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	full := cmd.CommandPath()
	root := cmd.Root().Name()
	if full == root {
		return ""
	}
	// CommandPath() format: "<root> <child> <grandchild>". Strip the
	// "<root> " prefix; cobra guarantees a single space separator.
	prefix := root + " "
	return strings.TrimPrefix(full, prefix)
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
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(permissionsCmd) // F5 — info command for the tier table.

	// Cloud commands — removed from CLI while being redesigned.
	// Code preserved in: login.go, logout.go, push.go, pull.go,
	// sync_cmd.go, billing.go, team.go
	// Re-enable by uncommenting:
	// rootCmd.AddCommand(newLoginCmd())
	// rootCmd.AddCommand(newLogoutCmd())
	// rootCmd.AddCommand(newPushCmd())
	// rootCmd.AddCommand(newPullCmd())
	// rootCmd.AddCommand(newSyncCmd())
	// rootCmd.AddCommand(newBillingCmd())
	// rootCmd.AddCommand(newTeamCmd())

	// F2 quality gate G4: every registered cobra command must have a
	// declared PermLevel in internal/auth.CommandTier. Validate() walks
	// the entire command tree and Refuses to start the binary if any
	// command is missing — the panic message names the missing paths so
	// the developer can add them to permissions.go before the next build.
	//
	// This is the deliberate "static" half of G4 enforcement; the
	// "runtime" half lives in rootPersistentPreRunE above.
	if err := auth.Validate(rootCmd); err != nil {
		panic(fmt.Sprintf("F2 G4 violation: %v", err))
	}
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

	ks := selectKeyStore(dir, flagNoKeychain, flagQuiet, os.Stderr)

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

// RootCmd returns the root cobra command. Useful for docs/manpage generation.
func RootCmd() *cobra.Command {
	return rootCmd
}

// selectKeyStore returns the appropriate KeyStore for an --no-keychain or
// default invocation. Sprint v1014-rc1-qa-fixes / FX1 (invariant I-11).
//
// Selection precedence:
//
//  1. flagNoKeychain == true AND TENE_KEYFILE is set:
//     -> FileStore at the user-specified path. This is the explicit
//     opt-in for the previous v1.0.13 behaviour (sharing a file-backed
//     store across processes) at a path the user controls.
//
//  2. flagNoKeychain == true AND TENE_KEYFILE is unset:
//     -> NullStore. No persistence at all; every invocation must resolve
//     the master password from TENE_MASTER_PASSWORD or the interactive
//     prompt. Replaces the old "single shared ~/.tene/keyfile" path
//     which caused the B1 cross-project bleed bug.
//
//  3. flagNoKeychain == false:
//     -> OS keychain via NewStoreWithStatus, which may itself drop down to
//     a FileStore when the OS keychain is unavailable (CI/Docker/headless);
//     in that case emitFallbackNoticeIfNeeded prints a single stderr line
//     so the user knows. This is the F6 path from PR #116 and is
//     intentionally unchanged.
//
// stderrW lets tests inject an io.Writer; production callers pass os.Stderr.
func selectKeyStore(dir string, noKeychain, quiet bool, stderrW io.Writer) keychain.KeyStore {
	if noKeychain {
		if envKeyfile := os.Getenv("TENE_KEYFILE"); envKeyfile != "" {
			// Explicit user opt-in to a file-backed store at the path
			// they chose. Documented in CHANGELOG as the v1.0.13 migration
			// path. Permissions and location are entirely the user's
			// responsibility — we do not validate the path beyond using
			// FileStore's existing 0600 file mode on write.
			return keychain.NewFileStore(envKeyfile)
		}
		// Default --no-keychain behaviour (v1.0.14+): no persistence.
		// I-11.
		return keychain.NewNullStore()
	}

	var info keychain.FallbackInfo
	ks, info := keychain.NewStoreWithStatus(dir)
	// F6 notice (PR #116). No-op on the happy path (info.Used == false).
	emitFallbackNoticeIfNeeded(stderrW, info, dir, quiet)
	return ks
}

// describeKeyStoreStatus returns the human-readable status line(s) that
// init.go's Step 14 prints to tell the user where (or whether) the master
// key was persisted. Sprint v1014-rc1-qa-fixes / FX1.
//
// rc1 always said "Master Key saved to OS Keychain" regardless of what
// actually happened, which is how B1 went undetected for a full release
// candidate. By branching on the concrete KeyStore type here we make the
// message truthful for every path selectKeyStore picks.
//
// The return value is a slice because the NullStore branch needs two lines
// (status + guidance) to make the consequence clear; every other branch
// returns a single line.
func describeKeyStoreStatus(ks keychain.KeyStore) []string {
	switch s := ks.(type) {
	case *keychain.KeyringStore:
		return []string{"Master Key saved to OS Keychain"}
	case *keychain.FileStore:
		// Reflect the source of truth (env var vs auto-fallback). The
		// FileStore type itself does not carry that distinction, so we
		// peek at the env var the same way selectKeyStore did.
		if envKeyfile := os.Getenv("TENE_KEYFILE"); envKeyfile != "" {
			return []string{fmt.Sprintf("Master Key saved to %s (via TENE_KEYFILE)", envKeyfile)}
		}
		// Auto-fallback to the standard fallback path (~/.tene/keyfile).
		// emitFallbackNoticeIfNeeded already explained "why" on stderr; we
		// just confirm the "where" here.
		home, _ := os.UserHomeDir()
		return []string{fmt.Sprintf("Master Key saved to %s (OS keychain unavailable)", filepath.Join(home, ".tene", "keyfile"))}
	case *keychain.NullStore:
		return []string{
			"Master Key NOT persisted (--no-keychain).",
			"You will be prompted for the password on each command, or set TENE_MASTER_PASSWORD.",
		}
	default:
		// Future-proofing: if a new KeyStore type ships without a status
		// entry, fall back to a generic message instead of crashing.
		// G7's reverse-drift check in auth.Validate is the static safety
		// net; this is only here in case a test injects a custom store.
		_ = s
		return []string{"Master Key storage configured."}
	}
}
