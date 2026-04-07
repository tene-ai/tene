package domain

import "time"

// Team represents a Tene Cloud team for shared vault access.
type Team struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Slug                string    `json:"slug"`
	OwnerID             string    `json:"owner_id"`
	LemonSubscriptionID string    `json:"lemon_subscription_id,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

// TeamMember represents a user's membership in a team.
type TeamMember struct {
	TeamID            string   `json:"team_id"`
	UserID            string   `json:"user_id"`
	Role              string   `json:"role"`              // "admin" or "member"
	EnvPermissions    []string `json:"env_permissions"`   // e.g. ["dev", "staging"]
	WrappedProjectKey []byte   `json:"wrapped_project_key,omitempty"`
	JoinedAt          string   `json:"joined_at"`
}

// TeamInvite holds the data needed to invite a member to a team.
type TeamInvite struct {
	TeamID            string `json:"team_id"`
	Email             string `json:"email"`
	Role              string `json:"role"`
	WrappedProjectKey []byte `json:"wrapped_project_key"` // X25519 ECDH wrapped PK
}
