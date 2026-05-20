package vault

import "time"

// Secret represents an encrypted secret record.
type Secret struct {
	ID             int64
	Name           string
	EncryptedValue string // base64(nonce + ciphertext)
	Environment    string
	Version        int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Environment represents a secret environment setting.
type Environment struct {
	Name      string
	IsActive  bool
	CreatedAt time.Time
}

// AuditEntry represents an audit log entry.
type AuditEntry struct {
	ID           int64
	Action       string // "secret.read", "secret.write", "secret.delete", "vault.init", "vault.passwd"
	ResourceName string
	Details      string
	Timestamp    time.Time
}

// SecretWrite is the payload for SetSecretBatchWithPreview: name, ciphertext
// (base64-encoded), and the already-derived preview substring. The caller
// (cli/import) is responsible for invoking pkg/crypto.DerivePreview to fill
// Preview from the plaintext before encryption.
type SecretWrite struct {
	Name           string
	EncryptedValue string
	Preview        string
}

// SecretBackfill is returned by ListSecretsForBackfill: the name and
// ciphertext of a secret whose preview column is empty and is therefore a
// candidate for `tene migrate fill-previews`. We deliberately do not embed
// this into Secret to keep the no-decrypt and decrypt code paths visually
// distinct at every call site.
type SecretBackfill struct {
	Name           string
	EncryptedValue string
}
