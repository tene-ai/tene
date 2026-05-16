// Package sync provides vault synchronization with Sync Envelope encryption.
package sync

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/agent-kay-it/tene/pkg/crypto"
)

const (
	// SyncKeyPurpose is the HKDF purpose string for deriving the Sync Envelope key.
	SyncKeyPurpose = "tene-sync-envelope"

	// envelopeMagic is the 4-byte magic number identifying a Sync Envelope.
	envelopeMagic = 0x54454E45 // "TENE"

	// envelopeVersion is the current envelope format version.
	envelopeVersion uint16 = 1

	// headerSize is the fixed envelope header size: magic(4) + version(2) + reserved(2) = 8 bytes.
	headerSize = 8
)

// DeriveSyncKey derives the L2 Sync Envelope encryption key from a master key.
func DeriveSyncKey(masterKey []byte) ([]byte, error) {
	return crypto.DeriveSubKey(masterKey, SyncKeyPurpose, 32)
}

// Seal encrypts a vault blob (vault.db bytes) into a Sync Envelope.
// The AAD binds the envelope to a specific project and environment.
//
// Envelope binary format:
//
//	┌────────┬──────────┬─────────────────────────┬──────────────────┐
//	│ Header │  Nonce   │       Ciphertext        │      Tag         │
//	│ 8 bytes│ 24 bytes │     variable length     │    16 bytes      │
//	└────────┴──────────┴─────────────────────────┴──────────────────┘
//
// Header: magic(4 bytes) + version(2 bytes) + reserved(2 bytes)
// AAD: projectID + ":" + environment
func Seal(syncKey, plaintext []byte, projectID, environment string) ([]byte, error) {
	if len(syncKey) != 32 {
		return nil, fmt.Errorf("sync: invalid sync key length: got %d, want 32", len(syncKey))
	}
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("sync: plaintext is empty")
	}

	aad := []byte(projectID + ":" + environment)

	// Encrypt with XChaCha20-Poly1305 (reuses existing crypto.Encrypt which prepends nonce)
	ciphertext, err := crypto.Encrypt(syncKey, plaintext, aad)
	if err != nil {
		return nil, fmt.Errorf("sync: envelope seal: %w", err)
	}

	// Build envelope: header + ciphertext (which already contains nonce)
	header := make([]byte, headerSize)
	binary.BigEndian.PutUint32(header[0:4], envelopeMagic)
	binary.BigEndian.PutUint16(header[4:6], envelopeVersion)
	// header[6:8] reserved, zero

	envelope := make([]byte, 0, headerSize+len(ciphertext))
	envelope = append(envelope, header...)
	envelope = append(envelope, ciphertext...)

	return envelope, nil
}

// Open decrypts a Sync Envelope back to the original vault blob.
func Open(syncKey, envelope []byte, projectID, environment string) ([]byte, error) {
	if len(syncKey) != 32 {
		return nil, fmt.Errorf("sync: invalid sync key length: got %d, want 32", len(syncKey))
	}
	if len(envelope) < headerSize+24+16 { // header + min nonce + min tag
		return nil, fmt.Errorf("sync: envelope too short")
	}

	// Validate header
	magic := binary.BigEndian.Uint32(envelope[0:4])
	if magic != envelopeMagic {
		return nil, fmt.Errorf("sync: invalid envelope magic: %x", magic)
	}
	ver := binary.BigEndian.Uint16(envelope[4:6])
	if ver != envelopeVersion {
		return nil, fmt.Errorf("sync: unsupported envelope version: %d", ver)
	}

	aad := []byte(projectID + ":" + environment)
	ciphertext := envelope[headerSize:]

	plaintext, err := crypto.Decrypt(syncKey, ciphertext, aad)
	if err != nil {
		return nil, fmt.Errorf("sync: envelope open: %w", err)
	}

	return plaintext, nil
}

// Checksum returns the SHA-256 hash of data.
func Checksum(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}
