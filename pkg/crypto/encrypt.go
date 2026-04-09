package crypto

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

// Encrypt encrypts plaintext using XChaCha20-Poly1305.
// key: 256-bit encryption key
// plaintext: data to encrypt
// aad: Additional Authenticated Data (e.g., secret key name)
// Returns: nonce(24 bytes) + ciphertext
func Encrypt(key, plaintext, aad []byte) ([]byte, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("%w: got %d, want %d", ErrInvalidKeyLength, len(key), chacha20poly1305.KeySize)
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: failed to create AEAD: %w", err)
	}

	nonce := make([]byte, aead.NonceSize()) // 24 bytes for XChaCha20
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: failed to generate nonce: %w", err)
	}

	// Seal appends ciphertext to nonce prefix
	ciphertext := aead.Seal(nonce, nonce, plaintext, aad)
	return ciphertext, nil
}
