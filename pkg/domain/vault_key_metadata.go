package domain

import "time"

// VaultKeyMeta represents a single secret key's metadata (no value).
type VaultKeyMeta struct {
	Name      string    `json:"name"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VaultMetadataPayload is the metadata sent with push (key names + environments).
type VaultMetadataPayload struct {
	Environments []string                  `json:"environments"`
	Keys         map[string][]VaultKeyMeta `json:"keys"`
	SecretCount  int                       `json:"secret_count"`
}
