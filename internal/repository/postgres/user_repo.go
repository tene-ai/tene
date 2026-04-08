package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tomo-kay/tene/internal/domain"
)

// UserRepo implements user operations with PostgreSQL.
// Satisfies billing.UserStore interface.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo creates a PostgreSQL user repository.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// UpsertUser inserts or updates a user (ON CONFLICT github_id).
// The user's ID, CreatedAt, and UpdatedAt are populated from the database.
func (r *UserRepo) UpsertUser(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (email, name, auth_provider, github_id, google_id, avatar_url, plan)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (github_id) WHERE github_id IS NOT NULL
		DO UPDATE SET
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			avatar_url = EXCLUDED.avatar_url,
			updated_at = now()
		RETURNING id, plan, created_at, updated_at`

	var googleID *string
	if u.GoogleID != "" {
		googleID = &u.GoogleID
	}
	var githubID *int64
	if u.GitHubID != 0 {
		githubID = &u.GitHubID
	}

	return r.pool.QueryRow(ctx, query,
		u.Email, u.Name, u.AuthProvider, githubID, googleID,
		u.AvatarURL, u.Plan,
	).Scan(&u.ID, &u.Plan, &u.CreatedAt, &u.UpdatedAt)
}

// GetUserByID retrieves a user by UUID.
func (r *UserRepo) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, name, auth_provider, github_id, google_id,
		       avatar_url, plan, lemon_customer_id, x25519_public_key,
		       created_at, updated_at
		FROM users WHERE id = $1`

	u := &domain.User{}
	var githubID *int64
	var googleID, lemonCustomerID, avatarURL *string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Name, &u.AuthProvider, &githubID, &googleID,
		&avatarURL, &u.Plan, &lemonCustomerID, &u.X25519PublicKey,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user: get by id: %w", err)
	}
	if githubID != nil {
		u.GitHubID = *githubID
	}
	if googleID != nil {
		u.GoogleID = *googleID
	}
	if lemonCustomerID != nil {
		u.LemonCustomerID = *lemonCustomerID
	}
	if avatarURL != nil {
		u.AvatarURL = *avatarURL
	}
	return u, nil
}

// GetUserByGitHubID retrieves a user by GitHub ID.
func (r *UserRepo) GetUserByGitHubID(ctx context.Context, githubID int64) (*domain.User, error) {
	query := `
		SELECT id, email, name, auth_provider, github_id, google_id,
		       avatar_url, plan, lemon_customer_id, x25519_public_key,
		       created_at, updated_at
		FROM users WHERE github_id = $1`

	u := &domain.User{}
	var gID *int64
	var googleID, lemonCustomerID, avatarURL *string
	err := r.pool.QueryRow(ctx, query, githubID).Scan(
		&u.ID, &u.Email, &u.Name, &u.AuthProvider, &gID, &googleID,
		&avatarURL, &u.Plan, &lemonCustomerID, &u.X25519PublicKey,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user: get by github id: %w", err)
	}
	if gID != nil {
		u.GitHubID = *gID
	}
	if googleID != nil {
		u.GoogleID = *googleID
	}
	if lemonCustomerID != nil {
		u.LemonCustomerID = *lemonCustomerID
	}
	if avatarURL != nil {
		u.AvatarURL = *avatarURL
	}
	return u, nil
}

// GetUserByEmail retrieves a user by email address.
func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, name, auth_provider, github_id, google_id,
		       avatar_url, plan, lemon_customer_id, x25519_public_key,
		       created_at, updated_at
		FROM users WHERE email = $1`

	u := &domain.User{}
	var githubID *int64
	var googleID, lemonCustomerID, avatarURL *string
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.Email, &u.Name, &u.AuthProvider, &githubID, &googleID,
		&avatarURL, &u.Plan, &lemonCustomerID, &u.X25519PublicKey,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user: get by email: %w", err)
	}
	if githubID != nil {
		u.GitHubID = *githubID
	}
	if googleID != nil {
		u.GoogleID = *googleID
	}
	if lemonCustomerID != nil {
		u.LemonCustomerID = *lemonCustomerID
	}
	if avatarURL != nil {
		u.AvatarURL = *avatarURL
	}
	return u, nil
}

// UpdatePlan updates user plan and LemonSqueezy customer ID by email.
// Implements billing.UserStore.
func (r *UserRepo) UpdatePlan(ctx context.Context, email string, plan string, lemonCustomerID string) error {
	query := `UPDATE users SET plan = $1, lemon_customer_id = $2 WHERE email = $3`
	ct, err := r.pool.Exec(ctx, query, plan, lemonCustomerID, email)
	if err != nil {
		return fmt.Errorf("user: update plan: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetLemonCustomerID returns the LemonSqueezy customer ID for a user.
// Implements billing.UserStore.
func (r *UserRepo) GetLemonCustomerID(ctx context.Context, userID string) (string, error) {
	var customerID *string
	err := r.pool.QueryRow(ctx,
		`SELECT lemon_customer_id FROM users WHERE id = $1`, userID,
	).Scan(&customerID)
	if err == pgx.ErrNoRows {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("user: get lemon customer id: %w", err)
	}
	if customerID == nil {
		return "", nil
	}
	return *customerID, nil
}

// UpdatePublicKey stores the user's X25519 public key for team key exchange.
func (r *UserRepo) UpdatePublicKey(ctx context.Context, userID string, publicKey []byte) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE users SET x25519_public_key = $1 WHERE id = $2`, publicKey, userID)
	if err != nil {
		return fmt.Errorf("user: update public key: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetPublicKey retrieves the X25519 public key for a user.
func (r *UserRepo) GetPublicKey(ctx context.Context, userID string) ([]byte, error) {
	var key []byte
	err := r.pool.QueryRow(ctx,
		`SELECT x25519_public_key FROM users WHERE id = $1`, userID,
	).Scan(&key)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user: get public key: %w", err)
	}
	return key, nil
}
