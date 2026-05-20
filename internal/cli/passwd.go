package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/agent-kay-it/tene/pkg/crypto"
	teneerr "github.com/agent-kay-it/tene/pkg/errors"
	"github.com/agent-kay-it/tene/internal/recovery"
)

var passwdCmd = &cobra.Command{
	Use:   "passwd",
	Short: "Change the Master Password",
	Long: `Change the Master Password.

This command is interactive-only by design: it refuses to run from a
non-TTY shell (CI/CD, log-redirected scripts, AI agent contexts) even
when TENE_MASTER_PASSWORD is set in the environment. The reason is
deliberate — master-password rotation is the highest-trust operation
the binary performs, and forcing a human-witnessed terminal session
closes a CI-driven brute-force vector where an attacker who can run
"tene passwd" repeatedly in the background could otherwise pivot a
weak old password into an arbitrary new one without an interactive
proof of presence.

If you need to rotate a password in CI, do it manually on a
maintainer's workstation and re-distribute the resulting recovery key
through your team's secret-sharing channel.`,
	RunE: runPasswd,
}

func runPasswd(cmd *cobra.Command, args []string) error {
	if !isTerminal() {
		return teneerr.ErrInteractiveRequired
	}

	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	// 1. Verify current password
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Enter current Master Password:")
	oldMasterKey, err := loadOrPromptMasterKey(app)
	if err != nil {
		return teneerr.ErrInvalidPassword
	}
	defer crypto.ZeroBytes(oldMasterKey)

	oldEncKey, err := crypto.DeriveSubKey(oldMasterKey, crypto.PurposeEncryption, 32)
	if err != nil {
		return err
	}
	defer crypto.ZeroBytes(oldEncKey)

	// 2. Get new password
	newPassword, err := promptPasswordConfirm("Enter new Master Password: ")
	if err != nil {
		return err
	}

	// 3. Generate new salt + master key
	newSalt, err := crypto.GenerateSalt()
	if err != nil {
		return err
	}
	newMasterKey, err := crypto.DeriveKey(newPassword, newSalt)
	if err != nil {
		return err
	}
	defer crypto.ZeroBytes(newMasterKey)
	newEncKey, err := crypto.DeriveSubKey(newMasterKey, crypto.PurposeEncryption, 32)
	if err != nil {
		return err
	}
	defer crypto.ZeroBytes(newEncKey)

	// 4. Re-encrypt all secrets
	env := resolveEnv(app)
	envs, err := app.Vault.ListEnvironments()
	if err != nil {
		return err
	}

	// Phase 1: Decrypt all secrets and re-encrypt in memory (atomic preparation)
	type reEncEntry struct {
		name, env, newEncVal string
	}
	var prepared []reEncEntry

	for _, e := range envs {
		allSecrets, err := app.Vault.GetAllSecrets(e.Name)
		if err != nil {
			return err
		}

		for name, encVal := range allSecrets {
			ct, err := decodeBase64(encVal)
			if err != nil {
				return fmt.Errorf("re-encryption failed (vault unchanged): decode %s: %w", name, err)
			}
			pt, err := crypto.Decrypt(oldEncKey, ct, []byte(name))
			if err != nil {
				return fmt.Errorf("re-encryption failed (vault unchanged): decrypt %s: %w", name, err)
			}

			newCt, err := crypto.Encrypt(newEncKey, pt, []byte(name))
			if err != nil {
				return fmt.Errorf("re-encryption failed (vault unchanged): encrypt %s: %w", name, err)
			}

			prepared = append(prepared, reEncEntry{name: name, env: e.Name, newEncVal: encodeBase64(newCt)})
		}
	}

	// Phase 2: Write all re-encrypted secrets (all-or-nothing intent)
	for _, entry := range prepared {
		if err := app.Vault.SetSecret(entry.name, entry.newEncVal, entry.env); err != nil {
			return fmt.Errorf("re-encryption write failed at %s/%s: %w", entry.env, entry.name, err)
		}
	}
	totalReencrypted := len(prepared)

	_ = env

	// 5. Update vault meta
	if err := app.Vault.SetMeta("kdf_salt", encodeBase64(newSalt)); err != nil {
		return err
	}

	// 6. Generate new recovery key
	mnemonic, err := recovery.GenerateMnemonic()
	if err != nil {
		return err
	}
	blob, err := recovery.EncryptMasterKey(newMasterKey, mnemonic)
	if err != nil {
		return err
	}
	if err := app.Vault.SetMeta("recovery_blob", encodeBase64(blob)); err != nil {
		return err
	}

	// 7. Update keychain
	if err := app.Keychain.Store(newMasterKey); err != nil {
		return err
	}

	// 8. Audit log
	_ = app.Vault.AddAuditLog("vault.passwd_changed", "", "")

	if flagJSON {
		return printJSON(map[string]any{
			"ok":           true,
			"reEncrypted":  totalReencrypted,
			"recoveryKey":  mnemonic,
		})
	}

	if !flagQuiet {
		fmt.Printf("\n  Re-encrypting vault...\n")
		fmt.Printf("  %d secrets re-encrypted.\n", totalReencrypted)
		fmt.Println("  Master Key updated in OS Keychain.")
		fmt.Println()
		fmt.Println("  New Recovery Key (write this down and keep it safe!):")
		fmt.Println("  +--------------------------------------------------+")
		words := splitMnemonicWords(mnemonic)
		if len(words) >= 12 {
			fmt.Printf("  |   %-47s|\n", joinWords(words[:6]))
			fmt.Printf("  |   %-47s|\n", joinWords(words[6:12]))
		}
		fmt.Println("  |                                                  |")
		fmt.Println("  |   Your previous Recovery Key is now invalid.     |")
		fmt.Println("  +--------------------------------------------------+")
		fmt.Println()
		fmt.Println("  Master Password changed successfully.")
	}

	return nil
}
