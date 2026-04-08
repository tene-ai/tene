package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WaitlistRepo implements handler.WaitlistStore with PostgreSQL.
type WaitlistRepo struct {
	pool *pgxpool.Pool
}

// NewWaitlistRepo creates a PostgreSQL waitlist repository.
func NewWaitlistRepo(pool *pgxpool.Pool) *WaitlistRepo {
	return &WaitlistRepo{pool: pool}
}

// AddToWaitlist registers an email in the waitlist.
// Duplicate emails are silently ignored (ON CONFLICT DO NOTHING).
func (r *WaitlistRepo) AddToWaitlist(email, plan, source string) error {
	query := `
		INSERT INTO waitlist (email, plan, source)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING`

	_, err := r.pool.Exec(context.Background(), query, email, plan, source)
	if err != nil {
		return fmt.Errorf("waitlist: add: %w", err)
	}
	return nil
}
