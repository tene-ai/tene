package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tomo-kay/tene/pkg/crypto"
	teneerr "github.com/tomo-kay/tene/pkg/errors"
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
