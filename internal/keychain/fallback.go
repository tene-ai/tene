package keychain

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

// FileStore is a file-based fallback KeyStore for environments where the OS keychain
// is unavailable (CI, Docker, headless servers).
type FileStore struct {
	path string
}

// NewFileStore creates a new file-based KeyStore.
func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (f *FileStore) Store(key []byte) error {
	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("keychain: create key dir: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(key)
	if err := os.WriteFile(f.path, []byte(encoded), 0600); err != nil {
		return fmt.Errorf("keychain: write key file: %w", err)
	}
	return nil
}

func (f *FileStore) Load() ([]byte, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("keychain: read key file: %w", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("keychain: decode key: %w", err)
	}
	return decoded, nil
}

func (f *FileStore) Delete() error {
	err := os.Remove(f.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("keychain: delete key file: %w", err)
	}
	return nil
}

func (f *FileStore) Exists() bool {
	_, err := os.Stat(f.path)
	return err == nil
}
