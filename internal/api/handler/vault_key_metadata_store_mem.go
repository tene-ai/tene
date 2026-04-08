package handler

import (
	"context"
	"sort"
	"sync"

	"github.com/tomo-kay/tene/internal/domain"
)

// MemVaultKeyMetadataStore is an in-memory VaultKeyMetadataStore for development.
type MemVaultKeyMetadataStore struct {
	mu   sync.RWMutex
	data map[string]map[string][]domain.VaultKeyMeta // vaultID -> env -> keys
}

// NewMemVaultKeyMetadataStore creates an in-memory vault key metadata store.
func NewMemVaultKeyMetadataStore() *MemVaultKeyMetadataStore {
	return &MemVaultKeyMetadataStore{
		data: make(map[string]map[string][]domain.VaultKeyMeta),
	}
}

// GetKeyMetadata returns key metadata for a vault and environment.
func (s *MemVaultKeyMetadataStore) GetKeyMetadata(_ context.Context, vaultID, env string) ([]domain.VaultKeyMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	envMap, ok := s.data[vaultID]
	if !ok {
		return nil, nil
	}
	keys := envMap[env]
	return keys, nil
}

// GetEnvironments returns all environments for a vault.
func (s *MemVaultKeyMetadataStore) GetEnvironments(_ context.Context, vaultID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	envMap, ok := s.data[vaultID]
	if !ok {
		return nil, nil
	}
	var envs []string
	for env := range envMap {
		envs = append(envs, env)
	}
	sort.Strings(envs)
	return envs, nil
}

// UpsertKeyMetadata replaces all key metadata for a vault.
func (s *MemVaultKeyMetadataStore) UpsertKeyMetadata(_ context.Context, vaultID string, payload domain.VaultMetadataPayload) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	envMap := make(map[string][]domain.VaultKeyMeta)
	for env, keys := range payload.Keys {
		envMap[env] = append(envMap[env], keys...)
	}
	s.data[vaultID] = envMap
	return nil
}
