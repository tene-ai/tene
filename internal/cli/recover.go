package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/agent-kay-it/tene/pkg/crypto"
	teneerr "github.com/agent-kay-it/tene/pkg/errors"
	"github.com/agent-kay-it/tene/internal/recovery"
)

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover vault using Recovery Key",
	RunE:  runRecover,
}

func runRecover(cmd *cobra.Command, args []string) error {
	if !isTerminal() {
		return teneerr.ErrInteractiveRequired
	}

	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	// 1. Get recovery key (12 words)
	fmt.Fprint(os.Stderr, "Enter Recovery Key (12 words): ")
	reader := bufio.NewReader(os.Stdin)
	mnemonic, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read recovery key: %w", err)
	}
	mnemonic = strings.TrimSpace(mnemonic)

	if !recovery.ValidateMnemonic(mnemonic) {
		return teneerr.ErrInvalidRecoveryKey
	}

	// 2. Load recovery blob from vault
	blobB64, err := app.Vault.GetMeta("recovery_blob")
	if err != nil {
		return fmt.Errorf("recovery data not found in vault")
	}
	blob, err := decodeBase64(blobB64)
	if err != nil {
		return fmt.Errorf("failed to decode recovery blob: %w", err)
	}

	// 3. Recover old master key
	oldMasterKey, err := recovery.RecoverMasterKey(blob, mnemonic)
	if err != nil {
		return teneerr.ErrInvalidRecoveryKey
	}
	defer crypto.ZeroBytes(oldMasterKey)

	oldEncKey, err := crypto.DeriveSubKey(oldMasterKey, crypto.PurposeEncryption, 32)
	if err != nil {
		return err
	}
	defer crypto.ZeroBytes(oldEncKey)

	// 4. Get new master password
	newPassword, err := promptPasswordConfirm("Enter new Master Password: ")
	if err != nil {
		return err
	}

	// 5. Generate new salt + master key
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

	// 6. Re-encrypt all secrets across all environments
	envs, err := app.Vault.ListEnvironments()
	if err != nil {
		return err
	}

	totalReencrypted := 0
	for _, e := range envs {
		allSecrets, err := app.Vault.GetAllSecrets(e.Name)
		if err != nil {
			return err
		}

		for name, encVal := range allSecrets {
			ct, err := decodeBase64(encVal)
			if err != nil {
				return err
			}
			pt, err := crypto.Decrypt(oldEncKey, ct, []byte(name))
			if err != nil {
				return err
			}

			newCt, err := crypto.Encrypt(newEncKey, pt, []byte(name))
			if err != nil {
				return err
			}

			if err := app.Vault.SetSecret(name, encodeBase64(newCt), e.Name); err != nil {
				return err
			}
			totalReencrypted++
		}
	}

	// 7. Update vault meta
	if err := app.Vault.SetMeta("kdf_salt", encodeBase64(newSalt)); err != nil {
		return err
	}

	// 8. Generate new recovery key
	newMnemonic, err := recovery.GenerateMnemonic()
	if err != nil {
		return err
	}
	newBlob, err := recovery.EncryptMasterKey(newMasterKey, newMnemonic)
	if err != nil {
		return err
	}
	if err := app.Vault.SetMeta("recovery_blob", encodeBase64(newBlob)); err != nil {
		return err
	}

	// 9. Update keychain
	if err := app.Keychain.Store(newMasterKey); err != nil {
		return err
	}

	// 10. Audit log
	_ = app.Vault.AddAuditLog("vault.recovered", "", "")

	if !flagQuiet {
		fmt.Println()
		fmt.Println("  Master Password reset successfully!")
		fmt.Printf("  Re-encrypting vault...\n")
		fmt.Printf("  %d secrets re-encrypted.\n", totalReencrypted)
		fmt.Println()
		fmt.Println("  New Recovery Key (write this down and keep it safe!):")
		fmt.Println("  +--------------------------------------------------+")
		words := splitMnemonicWords(newMnemonic)
		if len(words) >= 12 {
			fmt.Printf("  |   %-47s|\n", joinWords(words[:6]))
			fmt.Printf("  |   %-47s|\n", joinWords(words[6:12]))
		}
		fmt.Println("  |                                                  |")
		fmt.Println("  |   Your previous Recovery Key is now invalid.     |")
		fmt.Println("  +--------------------------------------------------+")
	}

	return nil
}
