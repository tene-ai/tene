package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tomo-kay/tene/internal/crypto"
)

var getCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Retrieve a decrypted secret",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	keyName := args[0]

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
	secret, err := app.Vault.GetSecret(keyName, env)
	if err != nil {
		return fmt.Errorf("Secret %q not found in %q environment.", keyName, env)
	}

	ciphertext, err := decodeBase64(secret.EncryptedValue)
	if err != nil {
		return fmt.Errorf("failed to decode secret: %w", err)
	}

	plaintext, err := crypto.Decrypt(encKey, ciphertext, []byte(keyName))
	if err != nil {
		return fmt.Errorf("Failed to decrypt secret. Master Password may have changed.")
	}

	// Audit log
	_ = app.Vault.AddAuditLog("secret.read", keyName, "")

	if flagJSON {
		return printJSON(map[string]any{
			"ok":          true,
			"name":        keyName,
			"value":       string(plaintext),
			"environment": env,
		})
	}

	fmt.Print(string(plaintext))
	if isTerminal() {
		fmt.Println()
	}
	return nil
}
