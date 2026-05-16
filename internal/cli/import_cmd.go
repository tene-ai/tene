package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/agent-kay-it/tene/pkg/crypto"
	teneerr "github.com/agent-kay-it/tene/pkg/errors"
)

var (
	importFlagOverwrite  bool
	importFlagEncrypted  bool
)

var importCmd = &cobra.Command{
	Use:   "import FILE",
	Short: "Import secrets from a .env file or encrypted backup",
	Args:  cobra.ExactArgs(1),
	RunE:  runImport,
}

func init() {
	importCmd.Flags().BoolVar(&importFlagOverwrite, "overwrite", false, "Overwrite existing secrets")
	importCmd.Flags().BoolVar(&importFlagEncrypted, "encrypted", false, "Import from encrypted backup file")
}

func runImport(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	app, err := loadApp()
	if err != nil {
		return err
	}
	defer func() { _ = app.Vault.Close() }()

	env := resolveEnv(app)

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

	if importFlagEncrypted {
		return importEncrypted(app, filePath, env, masterKey, encKey)
	}

	return importDotEnv(app, filePath, env, encKey)
}

func importDotEnv(app *App, filePath, env string, encKey []byte) error {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return teneerr.ErrFileNotFound(filePath)
		}
		return err
	}
	defer func() { _ = file.Close() }()

	secrets := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Remove "export " prefix
		line = strings.TrimPrefix(line, "export ")

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return teneerr.ErrFileParse(filePath, lineNum, "invalid format")
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes
		value = trimQuotes(value)

		secrets[key] = value
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if len(secrets) == 0 {
		return fmt.Errorf("no secrets found in %q", filePath)
	}

	// Check for existing secrets
	var names []string
	imported, skipped, overwritten := 0, 0, 0

	for key, value := range secrets {
		names = append(names, key)
		exists, _ := app.Vault.SecretExists(key, env)
		if exists && !importFlagOverwrite {
			skipped++
			continue
		}
		if exists {
			overwritten++
		}

		// Encrypt and store
		ct, err := crypto.Encrypt(encKey, []byte(value), []byte(key))
		if err != nil {
			return fmt.Errorf("failed to encrypt %s: %w", key, err)
		}
		if err := app.Vault.SetSecret(key, encodeBase64(ct), env); err != nil {
			return err
		}
		imported++
	}

	// Audit log
	_ = app.Vault.AddAuditLog("secrets.import", filePath, fmt.Sprintf("count=%d,env=%s", imported, env))

	if flagJSON {
		return printJSON(map[string]any{
			"ok":          true,
			"file":        filePath,
			"environment": env,
			"imported":    imported,
			"skipped":     skipped,
			"overwritten": overwritten,
			"secrets":     names,
		})
	}

	if !flagQuiet {
		fmt.Printf("%d secrets imported (encrypted).\n", imported)
		if skipped > 0 {
			fmt.Printf("%d skipped (already exist, use --overwrite).\n", skipped)
		}
	}

	return nil
}

func importEncrypted(app *App, filePath, env string, masterKey, encKey []byte) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return teneerr.ErrFileNotFound(filePath)
		}
		return err
	}

	// The encrypted backup is: encKey encrypted blob of KEY=VALUE pairs
	plaintext, err := crypto.Decrypt(encKey, data, []byte("tene-export"))
	if err != nil {
		return teneerr.ErrDecryptFailed
	}

	// Parse as .env format
	lines := strings.Split(string(plaintext), "\n")
	count := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		ct, err := crypto.Encrypt(encKey, []byte(value), []byte(key))
		if err != nil {
			return err
		}
		if err := app.Vault.SetSecret(key, encodeBase64(ct), env); err != nil {
			return err
		}
		count++
	}

	// Audit log
	_ = app.Vault.AddAuditLog("secrets.import", filePath, fmt.Sprintf("count=%d,env=%s,encrypted=true", count, env))

	if !flagQuiet {
		fmt.Printf("%d secrets restored to vault.\n", count)
	}
	return nil
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
