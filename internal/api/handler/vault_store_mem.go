package handler

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tomo-kay/tene/internal/domain"
)

// MemVaultStore is an in-memory VaultStore for development/testing.
// Replace with PostgreSQL repository in production.
type MemVaultStore struct {
	mu     sync.RWMutex
	vaults map[string]*domain.Vault // id -> vault
	audits []domain.AuditLog
}

// NewMemVaultStore creates an in-memory vault store.
func NewMemVaultStore() *MemVaultStore {
	return &MemVaultStore{
		vaults: make(map[string]*domain.Vault),
	}
}

// CreateVault stores a new vault record. Returns ErrProjectAlreadyExists if duplicate.
func (s *MemVaultStore) CreateVault(v *domain.Vault) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate project name per user
	for _, existing := range s.vaults {
		if existing.UserID == v.UserID && existing.ProjectName == v.ProjectName {
			return domain.ErrProjectAlreadyExists
		}
	}

	v.ID = uuid.New().String()
	v.CreatedAt = time.Now().UTC()
	v.UpdatedAt = v.CreatedAt
	s.vaults[v.ID] = v
	return nil
}

// GetVault retrieves a vault by ID and owner. Returns ErrVaultNotFound if missing.
func (s *MemVaultStore) GetVault(id, userID string) (*domain.Vault, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.vaults[id]
	if !ok || v.UserID != userID {
		return nil, domain.ErrVaultNotFound
	}
	return v, nil
}

// GetVaultByProject finds a vault by user and project name.
func (s *MemVaultStore) GetVaultByProject(userID, projectName string) (*domain.Vault, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, v := range s.vaults {
		if v.UserID == userID && v.ProjectName == projectName {
			return v, nil
		}
	}
	return nil, domain.ErrVaultNotFound
}

// ListVaults returns all vaults owned by the user.
func (s *MemVaultStore) ListVaults(userID string) ([]domain.Vault, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.Vault
	for _, v := range s.vaults {
		if v.UserID == userID {
			result = append(result, *v)
		}
	}
	return result, nil
}

// UpdateVaultVersion atomically increments version with optimistic locking.
func (s *MemVaultStore) UpdateVaultVersion(id string, currentVersion int64, hash []byte, size int64, secretCount int) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.vaults[id]
	if !ok {
		return 0, domain.ErrVaultNotFound
	}
	if v.Version != currentVersion {
		return 0, domain.ErrVersionConflict
	}

	v.Version++
	v.VaultHash = hash
	v.Size = size
	if secretCount > 0 {
		v.SecretCount = secretCount
	}
	now := time.Now().UTC()
	v.UpdatedAt = now
	v.LastPushedAt = &now
	return v.Version, nil
}

// DeleteVault removes a vault by ID and owner.
func (s *MemVaultStore) DeleteVault(id, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.vaults[id]
	if !ok || v.UserID != userID {
		return domain.ErrVaultNotFound
	}
	delete(s.vaults, id)
	return nil
}

// CreateAuditLog records an audit trail entry.
func (s *MemVaultStore) CreateAuditLog(log *domain.AuditLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.ID = uuid.New().String()
	log.CreatedAt = time.Now().UTC()
	s.audits = append(s.audits, *log)
	return nil
}
