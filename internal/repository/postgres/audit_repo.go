package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tomo-kay/tene/internal/domain"
)

// AuditRepo implements handler.AuditStore with PostgreSQL.
type AuditRepo struct {
	pool *pgxpool.Pool
}

// NewAuditRepo creates a PostgreSQL audit log repository.
func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{pool: pool}
}

// ListAuditLogs returns audit events for a user with optional filtering and pagination.
func (r *AuditRepo) ListAuditLogs(userID string, filter domain.AuditFilter) ([]domain.AuditLog, error) {
	// Build query with optional action filter
	query := `
		SELECT id, user_id, COALESCE(vault_id::text, ''), action,
		       COALESCE(detail, ''), COALESCE(host(ip_address), ''),
		       created_at
		FROM audit_logs
		WHERE user_id = $1`

	args := []any{userID}
	argIdx := 2

	if filter.Action != "" {
		query += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, filter.Action)
		argIdx++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("audit: list: %w", err)
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.VaultID, &l.Action,
			&l.Detail, &l.IPAddress, &l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("audit: list scan: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
