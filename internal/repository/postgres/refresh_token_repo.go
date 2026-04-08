package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RefreshTokenEntry represents a stored refresh token.
type RefreshTokenEntry struct {
	ID        string
	UserID    string
	TokenHash []byte
	Family    string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// RefreshTokenStore defines refresh token operations for database persistence.
type RefreshTokenStore interface {
	Store(ctx context.Context, userID string, tokenHash []byte, family string, expiresAt time.Time) error
	Find(ctx context.Context, tokenHash []byte) (*RefreshTokenEntry, error)
	Delete(ctx context.Context, tokenHash []byte) error
	RevokeFamily(ctx context.Context, family string) error
}

// RefreshTokenRepo implements RefreshTokenStore with PostgreSQL.
type RefreshTokenRepo struct {
	pool *pgxpool.Pool
}

// NewRefreshTokenRepo creates a PostgreSQL refresh token repository.
func NewRefreshTokenRepo(pool *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{pool: pool}
}

// Store saves a new refresh token hash with its family and expiry.
func (r *RefreshTokenRepo) Store(ctx context.Context, userID string, tokenHash []byte, family string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, family, expires_at)
		VALUES ($1, $2, $3, $4)`

	_, err := r.pool.Exec(ctx, query, userID, tokenHash, family, expiresAt)
	if err != nil {
		return fmt.Errorf("refresh_token: store: %w", err)
	}
	return nil
}

// Find retrieves a non-revoked refresh token entry by its hash.
// Returns nil, nil if the token is not found.
func (r *RefreshTokenRepo) Find(ctx context.Context, tokenHash []byte) (*RefreshTokenEntry, error) {
	query := `
		SELECT rt.id, rt.user_id, rt.token_hash, rt.family,
		       rt.expires_at, rt.revoked_at, rt.created_at
		FROM refresh_tokens rt
		WHERE rt.token_hash = $1 AND rt.revoked_at IS NULL`

	e := &RefreshTokenEntry{}
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&e.ID, &e.UserID, &e.TokenHash, &e.Family,
		&e.ExpiresAt, &e.RevokedAt, &e.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("refresh_token: find: %w", err)
	}
	return e, nil
}

// Delete revokes a specific refresh token by setting revoked_at.
func (r *RefreshTokenRepo) Delete(ctx context.Context, tokenHash []byte) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1`, tokenHash)
	if err != nil {
		return fmt.Errorf("refresh_token: delete: %w", err)
	}
	return nil
}

// RevokeFamily revokes all refresh tokens in a family (H-04: reuse detection).
func (r *RefreshTokenRepo) RevokeFamily(ctx context.Context, family string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE family = $1 AND revoked_at IS NULL`, family)
	if err != nil {
		return fmt.Errorf("refresh_token: revoke family: %w", err)
	}
	return nil
}
