package domain

import "time"

// Vault represents a cloud-synced vault record.
type Vault struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	TeamID      string    `json:"team_id,omitempty"`
	ProjectName string    `json:"project_name"`
	S3Key       string    `json:"s3_key"`
	Version     int64     `json:"vault_version"`
	VaultHash   []byte    `json:"vault_hash"`
	SecretCount int       `json:"secret_count"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastPushedAt *time.Time `json:"last_pushed_at,omitempty"`
}

// Device represents a registered device for sync.
type Device struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	DeviceName      string    `json:"device_name"`
	X25519PublicKey []byte    `json:"x25519_public_key"`
	LastSeenAt      time.Time `json:"last_seen_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// AuditLog represents an audit trail entry.
type AuditLog struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	VaultID   string    `json:"vault_id,omitempty"`
	Action    string    `json:"action"` // "push", "pull", "create", "delete"
	Detail    string    `json:"detail,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// SyncState tracks the local sync state for a vault.
type SyncState struct {
	VaultID      string `json:"vault_id"`
	LocalVersion int64  `json:"local_version"`
	RemoteVersion int64 `json:"remote_version"`
	LocalHash    string `json:"local_hash"`
	RemoteHash   string `json:"remote_hash"`
	LastPushedAt string `json:"last_pushed_at,omitempty"`
	LastPulledAt string `json:"last_pulled_at,omitempty"`
}
