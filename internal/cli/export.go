package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tomo-kay/tene/internal/crypto"
	teneerr "github.com/tomo-kay/tene/internal/errors"
)

var (
	exportFlagFile      string
	exportFlagEncrypted bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export secrets to .env format or encrypted backup",
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().StringVar(&exportFlagFile, "file", "", "Output file path")
	exportCmd.Flags().BoolVar(&exportFlagEncrypted, "encrypted", false, "Create encrypted backup file")
}

func runExport(cmd *cobra.Command, args []string) error {
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

	allSecrets, err := app.Vault.GetAllSecrets(env)
	if err != nil {
		return err
	}

	if len(allSecrets) == 0 {
		return fmt.Errorf("no secrets to export in %q environment", env)
	}

	// Decrypt all secrets
	decrypted := make(map[string]string)
	var sortedKeys []string
	for name, encVal := range allSecrets {
		ct, err := decodeBase64(encVal)
		if err != nil {
			return fmt.Errorf("failed to decode %s: %w", name, err)
		}
		pt, err := crypto.Decrypt(encKey, ct, []byte(name))
		if err != nil {
			return teneerr.ErrDecryptFailed
		}
		decrypted[name] = string(pt)
		sortedKeys = append(sortedKeys, name)
	}
	sort.Strings(sortedKeys)

	// Audit log
	_ = app.Vault.AddAuditLog("secrets.export", "", fmt.Sprintf("count=%d,env=%s,encrypted=%v", len(sortedKeys), env, exportFlagEncrypted))

	if exportFlagEncrypted {
		return exportEncrypted(app, env, sortedKeys, decrypted, encKey)
	}

	return exportDotEnv(env, sortedKeys, decrypted)
}

func exportDotEnv(env string, keys []string, decrypted map[string]string) error {
	// Build .env content
	var content string
	for _, key := range keys {
		content += fmt.Sprintf("%s=%s\n", key, decrypted[key])
	}

	if flagJSON {
		return printJSON(map[string]any{
			"ok":          true,
			"environment": env,
			"count":       len(keys),
			"secrets":     decrypted,
		})
	}

	if exportFlagFile != "" {
		if err := os.WriteFile(exportFlagFile, []byte(content), 0600); err != nil {
			return fmt.Errorf("cannot write to %q: %w", exportFlagFile, err)
		}
		if !flagQuiet {
			fmt.Printf("%d secrets exported to %s\n", len(keys), exportFlagFile)
			fmt.Println("Warning: This file contains plain-text secrets. Do not commit it.")
		}
		return nil
	}

	// stdout
	fmt.Print(content)
	return nil
}

func exportEncrypted(app *App, env string, keys []string, decrypted map[string]string, encKey []byte) error {
	// Build plaintext content
	var content string
	for _, key := range keys {
		content += fmt.Sprintf("%s=%s\n", key, decrypted[key])
	}

	// Encrypt the content
	ciphertext, err := crypto.Encrypt(encKey, []byte(content), []byte("tene-export"))
	if err != nil {
		return err
	}

	// Determine output file
	outFile := exportFlagFile
	if outFile == "" {
		projectName, _ := app.Vault.GetMeta("project_name")
		if projectName == "" {
			projectName = "tene"
		}
		outFile = filepath.Base(projectName) + ".tene.enc"
	}

	if err := os.WriteFile(outFile, ciphertext, 0600); err != nil {
		return fmt.Errorf("cannot write to %q: %w", outFile, err)
	}

	if flagJSON {
		return printJSON(map[string]any{
			"ok":          true,
			"environment": env,
			"file":        outFile,
			"encrypted":   true,
			"count":       len(keys),
		})
	}

	if !flagQuiet {
		fmt.Printf("Encrypted vault exported to: %s\n\n", outFile)
		fmt.Println("This file is encrypted with your Master Password.")
		fmt.Printf("To restore: tene import --encrypted %s\n\n", outFile)
		fmt.Println("Store this file in a safe place (USB, cloud drive, etc.)")
	}
	return nil
}
