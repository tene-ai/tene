package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the ~/.tene/config.json structure.
type Config struct {
	Version     int         `json:"version"`
	Analytics   Analytics   `json:"analytics"`
	Preferences Preferences `json:"preferences"`
}

// Analytics represents usage statistics.
type Analytics struct {
	SyncAttempts    int     `json:"syncAttempts"`
	LastSyncAttempt *string `json:"lastSyncAttempt"` // nullable ISO 8601
}

// Preferences represents user preferences.
type Preferences struct {
	Color        bool `json:"color"`
	AutoKeychain bool `json:"autoKeychain"`
}

// CloudConfig holds Tene Cloud connection settings.
type CloudConfig struct {
	APIURL       string `json:"apiUrl"`       // e.g. "https://api.tene.sh"
	AccessToken  string `json:"accessToken"`  // JWT (cached, may be expired)
	RefreshToken string `json:"refreshToken"` // for token renewal
	UserID       string `json:"userId"`
	Plan         string `json:"plan"` // "free" or "pro"
}

// SyncInfo holds per-project sync state metadata.
type SyncInfo struct {
	VaultID      string `json:"vaultId"`
	LocalVersion int64  `json:"localVersion"`
	RemoteVersion int64 `json:"remoteVersion"`
	LocalHash    string `json:"localHash"`
	RemoteHash   string `json:"remoteHash"`
	LastPushedAt string `json:"lastPushedAt,omitempty"`
	LastPulledAt string `json:"lastPulledAt,omitempty"`
	BaseSnapshot string `json:"baseSnapshot,omitempty"` // path to base vault.db snapshot for 3-way merge
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Analytics: Analytics{
			SyncAttempts:    0,
			LastSyncAttempt: nil,
		},
		Preferences: Preferences{
			Color:        true,
			AutoKeychain: true,
		},
	}
}

// ConfigDir returns the ~/.tene/ path.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: get home dir: %w", err)
	}
	return filepath.Join(home, ".tene"), nil
}

// ConfigPath returns the ~/.tene/config.json path.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", fmt.Errorf("config: get config path: %w", err)
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads ~/.tene/config.json.
// Returns default config if the file does not exist.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("config: read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), nil // parse failure: return default
	}
	return &cfg, nil
}

// Save writes to ~/.tene/config.json.
// Creates the directory if it does not exist.
func Save(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return fmt.Errorf("config: save get dir: %w", err)
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("config: create config dir: %w", err)
	}

	path := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("config: marshal config: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0600); err != nil {
		return fmt.Errorf("config: write config file: %w", err)
	}
	return nil
}

// EnsureConfigDir creates ~/.tene/ if it does not exist.
func EnsureConfigDir() error {
	dir, err := ConfigDir()
	if err != nil {
		return fmt.Errorf("config: ensure config dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("config: create config dir: %w", err)
	}
	return nil
}

// IncrementSyncAttempts increments syncAttempts and records a timestamp.
func IncrementSyncAttempts() error {
	cfg, err := Load()
	if err != nil {
		cfg = DefaultConfig()
	}

	cfg.Analytics.SyncAttempts++
	now := time.Now().UTC().Format(time.RFC3339)
	cfg.Analytics.LastSyncAttempt = &now

	return Save(cfg)
}
