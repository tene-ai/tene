package crypto

import (
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

// Decrypt decrypts ciphertext using XChaCha20-Poly1305.
// key: 256-bit encryption key
// ciphertext: nonce(24 bytes) + encrypted data
// aad: Additional Authenticated Data (must match what was used during encryption)
// Returns: decrypted plaintext
func Decrypt(key, ciphertext, aad []byte) ([]byte, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("%w: got %d, want %d", ErrInvalidKeyLength, len(key), chacha20poly1305.KeySize)
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: failed to create AEAD: %w", err)
	}

	if len(ciphertext) < aead.NonceSize() {
		return nil, fmt.Errorf("%w: too short", ErrInvalidCiphertext)
	}

	nonce := ciphertext[:aead.NonceSize()]
	message := ciphertext[aead.NonceSize():]

	plaintext, err := aead.Open(nil, nonce, message, aad)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}
