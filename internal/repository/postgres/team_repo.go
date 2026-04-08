package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tomo-kay/tene/internal/domain"
)

// TeamRepo implements handler.TeamStore with PostgreSQL.
type TeamRepo struct {
	pool *pgxpool.Pool
}

// NewTeamRepo creates a PostgreSQL team repository.
func NewTeamRepo(pool *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{pool: pool}
}

// CreateTeam inserts a new team.
// The team's ID, CreatedAt, and UpdatedAt are populated from the database.
func (r *TeamRepo) CreateTeam(t *domain.Team) error {
	query := `
		INSERT INTO teams (name, slug, owner_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(context.Background(), query,
		t.Name, t.Slug, t.OwnerID,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		// Check for unique slug violation
		if isDuplicateKeyError(err) {
			return domain.ErrProjectAlreadyExists
		}
		return fmt.Errorf("team: create: %w", err)
	}
	return nil
}

// GetTeam retrieves a team by ID.
func (r *TeamRepo) GetTeam(id string) (*domain.Team, error) {
	query := `
		SELECT id, name, slug, owner_id,
		       COALESCE(lemon_subscription_id, ''),
		       COALESCE(rotation_version, 0), COALESCE(rotation_pending, false),
		       created_at, updated_at
		FROM teams WHERE id = $1`

	t := &domain.Team{}
	err := r.pool.QueryRow(context.Background(), query, id).Scan(
		&t.ID, &t.Name, &t.Slug, &t.OwnerID,
		&t.LemonSubscriptionID,
		&t.RotationVersion, &t.RotationPending,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("team: get: %w", err)
	}
	return t, nil
}

// ListTeamsByUser returns all teams where the user is owner or member.
func (r *TeamRepo) ListTeamsByUser(userID string) ([]domain.Team, error) {
	query := `
		SELECT t.id, t.name, t.slug, t.owner_id,
		       COALESCE(t.lemon_subscription_id, ''),
		       COALESCE(t.rotation_version, 0), COALESCE(t.rotation_pending, false),
		       t.created_at, t.updated_at
		FROM teams t
		WHERE t.owner_id = $1
		UNION
		SELECT t.id, t.name, t.slug, t.owner_id,
		       COALESCE(t.lemon_subscription_id, ''),
		       COALESCE(t.rotation_version, 0), COALESCE(t.rotation_pending, false),
		       t.created_at, t.updated_at
		FROM teams t
		JOIN team_members tm ON tm.team_id = t.id
		WHERE tm.user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("team: list by user: %w", err)
	}
	defer rows.Close()

	var teams []domain.Team
	for rows.Next() {
		var t domain.Team
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Slug, &t.OwnerID,
			&t.LemonSubscriptionID,
			&t.RotationVersion, &t.RotationPending,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("team: list scan: %w", err)
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

// AddMember adds a user to a team.
func (r *TeamRepo) AddMember(m *domain.TeamMember) error {
	envPermsJSON, err := json.Marshal(m.EnvPermissions)
	if err != nil {
		envPermsJSON = []byte(`["dev"]`)
	}
	if len(m.EnvPermissions) == 0 {
		envPermsJSON = []byte(`["dev"]`)
	}

	query := `
		INSERT INTO team_members (team_id, user_id, role, env_permissions, wrapped_project_key)
		VALUES ($1, $2, $3, $4, $5)`

	_, err = r.pool.Exec(context.Background(), query,
		m.TeamID, m.UserID, m.Role, envPermsJSON, m.WrappedProjectKey)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrProjectAlreadyExists
		}
		return fmt.Errorf("team: add member: %w", err)
	}
	return nil
}

// RemoveMember removes a user from a team.
func (r *TeamRepo) RemoveMember(teamID, userID string) error {
	ct, err := r.pool.Exec(context.Background(),
		`DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`,
		teamID, userID)
	if err != nil {
		return fmt.Errorf("team: remove member: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotTeamMember
	}
	return nil
}

// UpdateMemberRole changes a member's role.
func (r *TeamRepo) UpdateMemberRole(teamID, userID, role string) error {
	ct, err := r.pool.Exec(context.Background(),
		`UPDATE team_members SET role = $1 WHERE team_id = $2 AND user_id = $3`,
		role, teamID, userID)
	if err != nil {
		return fmt.Errorf("team: update role: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotTeamMember
	}
	return nil
}

// ListMembers returns all members of a team.
func (r *TeamRepo) ListMembers(teamID string) ([]domain.TeamMember, error) {
	query := `
		SELECT team_id, user_id, role, env_permissions,
		       wrapped_project_key, joined_at
		FROM team_members WHERE team_id = $1
		ORDER BY joined_at`

	rows, err := r.pool.Query(context.Background(), query, teamID)
	if err != nil {
		return nil, fmt.Errorf("team: list members: %w", err)
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var m domain.TeamMember
		var envPermsJSON []byte
		var joinedAt interface{}
		if err := rows.Scan(
			&m.TeamID, &m.UserID, &m.Role, &envPermsJSON,
			&m.WrappedProjectKey, &joinedAt,
		); err != nil {
			return nil, fmt.Errorf("team: list members scan: %w", err)
		}
		if envPermsJSON != nil {
			_ = json.Unmarshal(envPermsJSON, &m.EnvPermissions)
		}
		if t, ok := joinedAt.(time.Time); ok {
			m.JoinedAt = t.Format(time.RFC3339)
		} else if joinedAt != nil {
			m.JoinedAt = fmt.Sprintf("%v", joinedAt)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// IsMember checks if a user is a member of a team (including owner).
func (r *TeamRepo) IsMember(teamID, userID string) bool {
	var exists bool
	_ = r.pool.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM teams WHERE id = $1 AND owner_id = $2
			UNION ALL
			SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2
		)`, teamID, userID).Scan(&exists)
	return exists
}

// IsAdmin checks if a user is an admin or owner of a team.
func (r *TeamRepo) IsAdmin(teamID, userID string) bool {
	var exists bool
	_ = r.pool.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM teams WHERE id = $1 AND owner_id = $2
			UNION ALL
			SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2 AND role = 'admin'
		)`, teamID, userID).Scan(&exists)
	return exists
}

// GetEnvPermissions returns the environment permissions for a user in a team.
func (r *TeamRepo) GetEnvPermissions(teamID, userID string) ([]string, error) {
	// Check if user is owner (full access)
	var isOwner bool
	_ = r.pool.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM teams WHERE id = $1 AND owner_id = $2)`,
		teamID, userID).Scan(&isOwner)
	if isOwner {
		return []string{"*"}, nil
	}

	var envPermsJSON []byte
	var role string
	err := r.pool.QueryRow(context.Background(),
		`SELECT role, env_permissions FROM team_members WHERE team_id = $1 AND user_id = $2`,
		teamID, userID).Scan(&role, &envPermsJSON)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrNotTeamMember
	}
	if err != nil {
		return nil, fmt.Errorf("team: get env permissions: %w", err)
	}

	var perms []string
	if envPermsJSON != nil {
		_ = json.Unmarshal(envPermsJSON, &perms)
	}
	if len(perms) > 0 {
		return perms, nil
	}

	// Default: admin gets all, member gets dev only
	if role == "admin" {
		return []string{"*"}, nil
	}
	return []string{"dev", "staging"}, nil
}

// SetRotationPending marks a team as needing key rotation and increments the rotation version.
func (r *TeamRepo) SetRotationPending(teamID string) error {
	ct, err := r.pool.Exec(context.Background(),
		`UPDATE teams SET rotation_pending = true, rotation_version = rotation_version + 1 WHERE id = $1`,
		teamID)
	if err != nil {
		return fmt.Errorf("team: set rotation pending: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// InvalidateWrappedKeys clears wrapped_project_key for all members except excludeUserID.
// This forces remaining members to re-wrap on next sync.
func (r *TeamRepo) InvalidateWrappedKeys(teamID, excludeUserID string) error {
	_, err := r.pool.Exec(context.Background(),
		`UPDATE team_members SET wrapped_project_key = NULL WHERE team_id = $1 AND user_id != $2`,
		teamID, excludeUserID)
	if err != nil {
		return fmt.Errorf("team: invalidate wrapped keys: %w", err)
	}
	return nil
}

// isDuplicateKeyError checks if a pgx error is a unique constraint violation.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// pgx wraps PostgreSQL error codes; unique_violation = 23505
	return contains(err.Error(), "23505") || contains(err.Error(), "duplicate key")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
