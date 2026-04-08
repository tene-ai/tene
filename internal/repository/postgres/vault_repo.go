package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tomo-kay/tene/internal/domain"
)

// VaultRepo implements handler.VaultStore with PostgreSQL.
type VaultRepo struct {
	pool *pgxpool.Pool
}

// NewVaultRepo creates a PostgreSQL vault repository.
func NewVaultRepo(pool *pgxpool.Pool) *VaultRepo {
	return &VaultRepo{pool: pool}
}

// CreateVault inserts a new vault record.
// The vault's ID, CreatedAt, and UpdatedAt are populated from the database.
func (r *VaultRepo) CreateVault(v *domain.Vault) error {
	query := `
		INSERT INTO vaults (user_id, team_id, project_name, s3_key, vault_version, vault_hash)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, 0, $5)
		RETURNING id, created_at, updated_at`

	var teamID *string
	if v.TeamID != "" {
		teamID = &v.TeamID
	}

	return r.pool.QueryRow(context.Background(), query,
		v.UserID, teamID, v.ProjectName, v.S3Key, v.VaultHash,
	).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
}

// GetVault retrieves a vault by ID and owner. Returns domain.ErrVaultNotFound if missing.
func (r *VaultRepo) GetVault(id, userID string) (*domain.Vault, error) {
	query := `
		SELECT id, user_id, COALESCE(team_id::text, ''), project_name, s3_key,
		       vault_version, vault_hash, COALESCE(secret_count, 0),
		       COALESCE(size, 0), created_at, updated_at, last_pushed_at
		FROM vaults WHERE id = $1 AND user_id = $2`

	v := &domain.Vault{}
	err := r.pool.QueryRow(context.Background(), query, id, userID).Scan(
		&v.ID, &v.UserID, &v.TeamID, &v.ProjectName, &v.S3Key,
		&v.Version, &v.VaultHash, &v.SecretCount,
		&v.Size, &v.CreatedAt, &v.UpdatedAt, &v.LastPushedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrVaultNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("vault: get: %w", err)
	}
	return v, nil
}

// GetVaultByProject finds a vault by user and project name.
func (r *VaultRepo) GetVaultByProject(userID, projectName string) (*domain.Vault, error) {
	query := `
		SELECT id, user_id, COALESCE(team_id::text, ''), project_name, s3_key,
		       vault_version, vault_hash, COALESCE(secret_count, 0),
		       COALESCE(size, 0), created_at, updated_at, last_pushed_at
		FROM vaults WHERE user_id = $1 AND project_name = $2`

	v := &domain.Vault{}
	err := r.pool.QueryRow(context.Background(), query, userID, projectName).Scan(
		&v.ID, &v.UserID, &v.TeamID, &v.ProjectName, &v.S3Key,
		&v.Version, &v.VaultHash, &v.SecretCount,
		&v.Size, &v.CreatedAt, &v.UpdatedAt, &v.LastPushedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrVaultNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("vault: get by project: %w", err)
	}
	return v, nil
}

// ListVaults returns all vaults owned by the user, ordered by last update.
func (r *VaultRepo) ListVaults(userID string) ([]domain.Vault, error) {
	query := `
		SELECT id, user_id, COALESCE(team_id::text, ''), project_name, s3_key,
		       vault_version, vault_hash, COALESCE(secret_count, 0),
		       COALESCE(size, 0), created_at, updated_at, last_pushed_at
		FROM vaults WHERE user_id = $1
		ORDER BY updated_at DESC`

	rows, err := r.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("vault: list: %w", err)
	}
	defer rows.Close()

	var vaults []domain.Vault
	for rows.Next() {
		var v domain.Vault
		if err := rows.Scan(
			&v.ID, &v.UserID, &v.TeamID, &v.ProjectName, &v.S3Key,
			&v.Version, &v.VaultHash, &v.SecretCount,
			&v.Size, &v.CreatedAt, &v.UpdatedAt, &v.LastPushedAt,
		); err != nil {
			return nil, fmt.Errorf("vault: list scan: %w", err)
		}
		vaults = append(vaults, v)
	}
	return vaults, rows.Err()
}

// UpdateVaultVersion atomically increments version with optimistic locking.
// Returns the new version number or domain.ErrVersionConflict if the current version has changed.
func (r *VaultRepo) UpdateVaultVersion(id string, currentVersion int64, hash []byte, size int64, secretCount int) (int64, error) {
	query := `
		UPDATE vaults
		SET vault_version = vault_version + 1,
		    vault_hash = $1,
		    size = $2,
		    secret_count = CASE WHEN $3 > 0 THEN $3 ELSE secret_count END,
		    last_pushed_at = now()
		WHERE id = $4 AND vault_version = $5
		RETURNING vault_version`

	var newVersion int64
	err := r.pool.QueryRow(context.Background(), query,
		hash, size, secretCount, id, currentVersion,
	).Scan(&newVersion)
	if err == pgx.ErrNoRows {
		return 0, domain.ErrVersionConflict
	}
	if err != nil {
		return 0, fmt.Errorf("vault: update version: %w", err)
	}
	return newVersion, nil
}

// DeleteVault removes a vault by ID and owner.
func (r *VaultRepo) DeleteVault(id, userID string) error {
	ct, err := r.pool.Exec(context.Background(),
		`DELETE FROM vaults WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("vault: delete: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrVaultNotFound
	}
	return nil
}

// CreateAuditLog records an audit trail entry in the partitioned audit_logs table.
func (r *VaultRepo) CreateAuditLog(log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (user_id, vault_id, action, detail, ip_address)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5::inet)`

	_, err := r.pool.Exec(context.Background(), query,
		log.UserID, log.VaultID, log.Action, log.Detail, log.IPAddress)
	if err != nil {
		return fmt.Errorf("audit: create: %w", err)
	}
	return nil
}
