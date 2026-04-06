package vault

import (
	"encoding/json"
	"os"
	"time"
)

// VaultJSON represents the .tene/vault.json file structure.
// Stores human-readable metadata separate from the SQLite vault.db.
type VaultJSON struct {
	ProjectName       string   `json:"projectName"`
	CreatedAt         string   `json:"createdAt"`
	VaultVersion      int      `json:"vaultVersion"`
	ActiveEnvironment string   `json:"activeEnvironment"`
	Agents            []string `json:"agents"`
}

// WriteVaultJSON creates the .tene/vault.json file.
func WriteVaultJSON(path, projectName, activeEnv string) error {
	vj := VaultJSON{
		ProjectName:       projectName,
		CreatedAt:         time.Now().UTC().Format(time.RFC3339),
		VaultVersion:      1,
		ActiveEnvironment: activeEnv,
		Agents:            []string{"claude", "cursor", "windsurf", "gemini", "codex"},
	}

	data, err := json.MarshalIndent(vj, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0600)
}

// ReadVaultJSON reads the .tene/vault.json file.
func ReadVaultJSON(path string) (*VaultJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var vj VaultJSON
	if err := json.Unmarshal(data, &vj); err != nil {
		return nil, err
	}
	return &vj, nil
}

// UpdateVaultJSONEnv updates the activeEnvironment in vault.json.
func UpdateVaultJSONEnv(path, env string) error {
	vj, err := ReadVaultJSON(path)
	if err != nil {
		return err
	}

	vj.ActiveEnvironment = env

	data, err := json.MarshalIndent(vj, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0600)
}
