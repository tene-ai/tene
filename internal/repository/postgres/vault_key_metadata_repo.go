package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tomo-kay/tene/internal/domain"
)

// VaultKeyMetadataRepo implements vault key metadata operations with PostgreSQL.
type VaultKeyMetadataRepo struct {
	pool *pgxpool.Pool
}

// NewVaultKeyMetadataRepo creates a PostgreSQL vault key metadata repository.
func NewVaultKeyMetadataRepo(pool *pgxpool.Pool) *VaultKeyMetadataRepo {
	return &VaultKeyMetadataRepo{pool: pool}
}

// GetKeyMetadata returns key metadata for a vault and environment.
func (r *VaultKeyMetadataRepo) GetKeyMetadata(ctx context.Context, vaultID, env string) ([]domain.VaultKeyMeta, error) {
	query := `
		SELECT key_name, version, updated_at
		FROM vault_key_metadata
		WHERE vault_id = $1 AND environment = $2
		ORDER BY key_name`

	rows, err := r.pool.Query(ctx, query, vaultID, env)
	if err != nil {
		return nil, fmt.Errorf("vault_key_metadata: get: %w", err)
	}
	defer rows.Close()

	var keys []domain.VaultKeyMeta
	for rows.Next() {
		var k domain.VaultKeyMeta
		if err := rows.Scan(&k.Name, &k.Version, &k.UpdatedAt); err != nil {
			return nil, fmt.Errorf("vault_key_metadata: scan: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// GetEnvironments returns all environments for a vault.
func (r *VaultKeyMetadataRepo) GetEnvironments(ctx context.Context, vaultID string) ([]string, error) {
	query := `
		SELECT DISTINCT environment
		FROM vault_key_metadata
		WHERE vault_id = $1
		ORDER BY environment`

	rows, err := r.pool.Query(ctx, query, vaultID)
	if err != nil {
		return nil, fmt.Errorf("vault_key_metadata: envs: %w", err)
	}
	defer rows.Close()

	var envs []string
	for rows.Next() {
		var env string
		if err := rows.Scan(&env); err != nil {
			return nil, fmt.Errorf("vault_key_metadata: envs scan: %w", err)
		}
		envs = append(envs, env)
	}
	return envs, rows.Err()
}

// UpsertKeyMetadata replaces all key metadata for a vault from a push payload.
func (r *VaultKeyMetadataRepo) UpsertKeyMetadata(ctx context.Context, vaultID string, payload domain.VaultMetadataPayload) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("vault_key_metadata: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Delete existing metadata for this vault
	if _, err := tx.Exec(ctx, `DELETE FROM vault_key_metadata WHERE vault_id = $1`, vaultID); err != nil {
		return fmt.Errorf("vault_key_metadata: delete: %w", err)
	}

	// Insert new metadata
	for env, keys := range payload.Keys {
		for _, k := range keys {
			updatedAt := k.UpdatedAt
			if updatedAt.IsZero() {
				updatedAt = time.Now()
			}
			_, err := tx.Exec(ctx,
				`INSERT INTO vault_key_metadata (vault_id, environment, key_name, version, updated_at)
				 VALUES ($1, $2, $3, $4, $5)`,
				vaultID, env, k.Name, k.Version, updatedAt,
			)
			if err != nil {
				return fmt.Errorf("vault_key_metadata: insert %s/%s: %w", env, k.Name, err)
			}
		}
	}

	return tx.Commit(ctx)
}
