package crypto

import "errors"

var (
	// ErrInvalidKeyLength is returned when the key length is incorrect.
	ErrInvalidKeyLength = errors.New("crypto: invalid key length")

	// ErrInvalidSaltLength is returned when the salt length is incorrect.
	ErrInvalidSaltLength = errors.New("crypto: invalid salt length")

	// ErrDecryptionFailed is returned when decryption fails (wrong key or tampered data).
	ErrDecryptionFailed = errors.New("crypto: decryption failed")

	// ErrInvalidCiphertext is returned when the ciphertext format is invalid.
	ErrInvalidCiphertext = errors.New("crypto: invalid ciphertext format")
)
