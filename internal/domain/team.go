package domain

import "time"

// Team represents a Tene Cloud team for shared vault access.
type Team struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Slug                string    `json:"slug"`
	OwnerID             string    `json:"owner_id"`
	LemonSubscriptionID string    `json:"lemon_subscription_id,omitempty"`
	RotationVersion     int64     `json:"rotation_version"`
	RotationPending     bool      `json:"rotation_pending"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
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

// TeamMemberView is an enriched member view with user profile info for display.
type TeamMemberView struct {
	TeamID         string   `json:"team_id"`
	UserID         string   `json:"user_id"`
	Role           string   `json:"role"`
	EnvPermissions []string `json:"env_permissions"`
	JoinedAt       string   `json:"joined_at"`
	Name           string   `json:"name"`
	AvatarURL      string   `json:"avatar_url"`
}

// TeamInvite holds the data needed to invite a member to a team.
type TeamInvite struct {
	TeamID            string `json:"team_id"`
	Email             string `json:"email"`
	Role              string `json:"role"`
	WrappedProjectKey []byte `json:"wrapped_project_key"` // X25519 ECDH wrapped PK
}
