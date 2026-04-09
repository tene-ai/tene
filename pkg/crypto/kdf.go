package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id parameters
	ArgonTime    = 3         // iterations
	ArgonMemory  = 64 * 1024 // 64MB
	ArgonThreads = 1         // parallelism
	ArgonKeyLen  = 32        // 256-bit output

	// Salt/Nonce lengths
	SaltLen  = 16 // 128-bit salt
	NonceLen = 24 // 192-bit nonce (XChaCha20)

	// HKDF purpose labels
	PurposeEncryption = "tene-encryption-key"
	PurposeAuth       = "tene-auth-hash"
	PurposeDeviceKey  = "tene-device-key"
)

// GenerateSalt generates a 128-bit random salt.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("crypto: failed to generate salt: %w", err)
	}
	return salt, nil
}

// DeriveKey derives a Master Key from password using Argon2id.
func DeriveKey(password string, salt []byte) ([]byte, error) {
	if len(salt) != SaltLen {
		return nil, fmt.Errorf("%w: got %d, want %d", ErrInvalidSaltLength, len(salt), SaltLen)
	}
	if password == "" {
		return nil, errors.New("crypto: password cannot be empty")
	}

	key := argon2.IDKey(
		[]byte(password),
		salt,
		ArgonTime,
		ArgonMemory,
		ArgonThreads,
		ArgonKeyLen,
	)
	return key, nil
}
