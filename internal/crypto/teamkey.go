package crypto

import (
	"crypto/rand"
	"fmt"
	"io"
)

const (
	// ProjectKeySize is the size of a team project key in bytes.
	ProjectKeySize = 32
	// teamKeyWrapPurpose is the HKDF purpose string for key wrapping.
	teamKeyWrapPurpose = "tene-key-wrap"
)

// GenerateProjectKey creates a new random 256-bit project key for team vaults.
func GenerateProjectKey() ([]byte, error) {
	key := make([]byte, ProjectKeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("crypto: generate project key: %w", err)
	}
	return key, nil
}

// WrapProjectKey encrypts a project key for a specific team member using X25519 ECDH.
//
// Protocol (per design §2.3):
//  1. ECDH: shared_secret = X25519(senderPrivate, recipientPublic)
//  2. wrap_key = HKDF(shared_secret, "tene-key-wrap", salt=projectID)
//  3. wrapped_pk = XChaCha20-Poly1305(wrap_key, projectKey, AAD=recipientUserID)
func WrapProjectKey(senderPrivate, recipientPublic []byte,
	projectID, recipientUserID string, projectKey []byte) ([]byte, error) {

	if len(projectKey) != ProjectKeySize {
		return nil, fmt.Errorf("%w: project key must be %d bytes", ErrInvalidKeyLength, ProjectKeySize)
	}

	// 1. ECDH shared secret
	sharedSecret, err := X25519SharedSecret(senderPrivate, recipientPublic)
	if err != nil {
		return nil, fmt.Errorf("crypto: wrap project key: %w", err)
	}
	defer ZeroBytes(sharedSecret)

	// 2. Derive wrap key with project ID as salt
	wrapKey, err := DeriveSubKeyWithSalt(sharedSecret, []byte(projectID), teamKeyWrapPurpose, 32)
	if err != nil {
		return nil, fmt.Errorf("crypto: wrap project key: %w", err)
	}
	defer ZeroBytes(wrapKey)

	// 3. Encrypt project key with AAD = recipient user ID
	wrapped, err := Encrypt(wrapKey, projectKey, []byte(recipientUserID))
	if err != nil {
		return nil, fmt.Errorf("crypto: wrap project key: %w", err)
	}

	return wrapped, nil
}

// UnwrapProjectKey decrypts a wrapped project key using the recipient's private key.
//
// Protocol (reverse of WrapProjectKey):
//  1. ECDH: shared_secret = X25519(recipientPrivate, senderPublic)
//  2. wrap_key = HKDF(shared_secret, "tene-key-wrap", salt=projectID)
//  3. projectKey = XChaCha20-Poly1305.Open(wrap_key, wrapped, AAD=recipientUserID)
func UnwrapProjectKey(recipientPrivate, senderPublic []byte,
	projectID, recipientUserID string, wrappedKey []byte) ([]byte, error) {

	// 1. ECDH shared secret (same result due to ECDH commutativity)
	sharedSecret, err := X25519SharedSecret(recipientPrivate, senderPublic)
	if err != nil {
		return nil, fmt.Errorf("crypto: unwrap project key: %w", err)
	}
	defer ZeroBytes(sharedSecret)

	// 2. Derive same wrap key
	wrapKey, err := DeriveSubKeyWithSalt(sharedSecret, []byte(projectID), teamKeyWrapPurpose, 32)
	if err != nil {
		return nil, fmt.Errorf("crypto: unwrap project key: %w", err)
	}
	defer ZeroBytes(wrapKey)

	// 3. Decrypt
	projectKey, err := Decrypt(wrapKey, wrappedKey, []byte(recipientUserID))
	if err != nil {
		return nil, fmt.Errorf("crypto: unwrap project key: %w", err)
	}

	return projectKey, nil
}
