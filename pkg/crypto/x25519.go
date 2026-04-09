package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

const (
	// X25519KeySize is the size of X25519 public and private keys in bytes.
	X25519KeySize = 32
)

// X25519KeyPair holds a generated X25519 key pair.
type X25519KeyPair struct {
	PrivateKey []byte // 32 bytes, stored in OS Keychain
	PublicKey  []byte // 32 bytes, registered on server
}

// GenerateX25519KeyPair creates a new X25519 key pair for device key exchange.
func GenerateX25519KeyPair() (*X25519KeyPair, error) {
	privateKey := make([]byte, X25519KeySize)
	if _, err := io.ReadFull(rand.Reader, privateKey); err != nil {
		return nil, fmt.Errorf("crypto: generate x25519 private key: %w", err)
	}

	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("crypto: derive x25519 public key: %w", err)
	}

	return &X25519KeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

// X25519SharedSecret computes the ECDH shared secret between a private key and a peer's public key.
func X25519SharedSecret(privateKey, peerPublicKey []byte) ([]byte, error) {
	if len(privateKey) != X25519KeySize {
		return nil, fmt.Errorf("%w: private key must be %d bytes", ErrInvalidKeyLength, X25519KeySize)
	}
	if len(peerPublicKey) != X25519KeySize {
		return nil, fmt.Errorf("%w: public key must be %d bytes", ErrInvalidKeyLength, X25519KeySize)
	}

	shared, err := curve25519.X25519(privateKey, peerPublicKey)
	if err != nil {
		return nil, fmt.Errorf("crypto: x25519 ecdh: %w", err)
	}

	return shared, nil
}

// DeriveSubKeyWithSalt derives a sub-key from the master key using HKDF-SHA256 with an explicit salt.
// This is used for team key wrapping where the salt provides domain separation.
func DeriveSubKeyWithSalt(masterKey, salt []byte, purpose string, length int) ([]byte, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("%w: master key must be 32 bytes", ErrInvalidKeyLength)
	}

	hkdfReader := hkdf.New(sha256.New, masterKey, salt, []byte(purpose))
	subKey := make([]byte, length)
	if _, err := io.ReadFull(hkdfReader, subKey); err != nil {
		return nil, fmt.Errorf("crypto: HKDF expand with salt failed: %w", err)
	}
	return subKey, nil
}
