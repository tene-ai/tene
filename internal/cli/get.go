package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tene-ai/tene/pkg/crypto"
	teneerr "github.com/tene-ai/tene/pkg/errors"
)

var flagUnsafeStdout bool

var getCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Retrieve a decrypted secret",
	Long: `Decrypt and print a secret.

⚠️  Security note: when stdout is piped or redirected (non-TTY), this
command refuses by default to prevent accidental secret leakage into AI
agent context windows, log aggregators, or shell history files.

To override, use one of:
  --unsafe-stdout              (per-invocation opt-in)
  TENE_ALLOW_STDOUT_SECRETS=1  (environment override)

Preferred alternative: use 'tene run -- <cmd>' which injects secrets as
environment variables without ever printing them to stdout.`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func init() {
	getCmd.Flags().BoolVar(&flagUnsafeStdout, "unsafe-stdout", false,
		"Explicitly allow printing plaintext to non-TTY stdout "+
			"(equivalent to TENE_ALLOW_STDOUT_SECRETS=1)")
}

func runGet(cmd *cobra.Command, args []string) error {
	keyName := args[0]

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
	secret, err := app.Vault.GetSecret(keyName, env)
	if err != nil {
		return teneerr.ErrSecretNotFound(keyName, env)
	}

	ciphertext, err := decodeBase64(secret.EncryptedValue)
	if err != nil {
		return teneerr.ErrDecryptFailed
	}

	plaintext, err := crypto.Decrypt(encKey, ciphertext, []byte(keyName))
	if err != nil {
		return teneerr.ErrDecryptFailed
	}
	defer crypto.ZeroBytes(plaintext)

	// Audit log
	_ = app.Vault.AddAuditLog("secret.read", keyName, "")

	// JSON output: warn on stderr when non-TTY + not opted in, but still
	// proceed (JSON callers know what they're doing — typically scripts
	// that parse the output explicitly).
	if flagJSON {
		if !isTerminal() && !stdoutSecretsAllowed() {
			fmt.Fprintln(os.Stderr,
				"warning: emitting plaintext secret as JSON to non-TTY stdout. "+
					"Consider `tene run --` for safer injection.")
		}
		return printJSON(map[string]any{
			"ok":          true,
			"name":        keyName,
			"value":       string(plaintext),
			"environment": env,
		})
	}

	// Plain text output: block on non-TTY unless the caller explicitly
	// opted in. This prevents accidental `tene get K | anything` from
	// surfacing plaintext into logs, AI agent tool results, or shell history.
	if !isTerminal() && !stdoutSecretsAllowed() {
		return teneerr.ErrStdoutSecretBlocked
	}

	fmt.Print(string(plaintext))
	if isTerminal() {
		fmt.Println()
	}
	return nil
}

// stdoutSecretsAllowed returns true if the caller explicitly opted in to
// plaintext secret output on non-TTY stdout, either via the --unsafe-stdout
// flag or the TENE_ALLOW_STDOUT_SECRETS=1 environment variable.
func stdoutSecretsAllowed() bool {
	if flagUnsafeStdout {
		return true
	}
	if v := os.Getenv("TENE_ALLOW_STDOUT_SECRETS"); v == "1" || v == "true" {
		return true
	}
	return false
}
