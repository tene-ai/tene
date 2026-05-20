package domain

import "time"

// VaultKeyMeta represents a single secret key's metadata (no value).
//
// Preview is a short, plaintext sub-string of the secret value used to give
// the user a visual cue in `tene list` without having to unlock the vault.
// It is intentionally stored plaintext in the local vault.db on user opt-in
// (sprint cli-ux-permission-model, Q2 decision, 2026-05-20). The substring
// length is hard-capped at front+back <= 8 chars (enforced by
// pkg/crypto.DerivePreview) so that no realistic API key prefix can be
// recovered from it when defaults (front=0, back=4) are in effect.
//
// JSON tag does NOT use omitempty: Preview must always emit, even when it
// is the empty string. An empty string means "no preview available" (either
// preview.enabled=false config, schema-v1 vault that has not been backfilled,
// or a value too short to derive a safe preview from). This "always-string"
// contract removes type instability for JSON consumers.
type VaultKeyMeta struct {
	Name      string    `json:"name"`
	Version   int       `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
	Preview   string    `json:"preview"`
}

// VaultMetadataPayload is the metadata sent with push (key names + environments).
type VaultMetadataPayload struct {
	Environments []string                  `json:"environments"`
	Keys         map[string][]VaultKeyMeta `json:"keys"`
	SecretCount  int                       `json:"secret_count"`
}
