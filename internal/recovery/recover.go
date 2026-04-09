package recovery

import (
	"fmt"

	"github.com/tomo-kay/tene/pkg/crypto"
)

const (
	RecoverySalt    = "tene-recovery-salt" // padded to 16 bytes
	RecoveryPurpose = "tene-recovery-key"
)

// recoverySaltBytes returns a 16-byte salt for recovery key derivation.
func recoverySaltBytes() []byte {
	salt := make([]byte, crypto.SaltLen)
	copy(salt, []byte(RecoverySalt))
	return salt
}

// EncryptMasterKey encrypts the master key using a recovery key derived from the mnemonic.
// The returned blob should be stored in vault_meta as 'recovery_blob'.
func EncryptMasterKey(masterKey []byte, mnemonic string) ([]byte, error) {
	if !ValidateMnemonic(mnemonic) {
		return nil, ErrInvalidMnemonic
	}

	// Derive recovery key from mnemonic
	recoveryKey, err := crypto.DeriveKey(mnemonic, recoverySaltBytes())
	if err != nil {
		return nil, fmt.Errorf("recovery: failed to derive recovery key: %w", err)
	}

	// Derive encryption sub-key from recovery key
	encKey, err := crypto.DeriveSubKey(recoveryKey, RecoveryPurpose, 32)
	if err != nil {
		return nil, fmt.Errorf("recovery: derive sub-key: %w", err)
	}

	// Encrypt master key
	blob, err := crypto.Encrypt(encKey, masterKey, []byte("recovery"))
	if err != nil {
		return nil, fmt.Errorf("recovery: failed to encrypt master key: %w", err)
	}
	return blob, nil
}

// RecoverMasterKey recovers the master key from the encrypted blob using the mnemonic.
func RecoverMasterKey(blob []byte, mnemonic string) ([]byte, error) {
	if !ValidateMnemonic(mnemonic) {
		return nil, ErrInvalidMnemonic
	}

	recoveryKey, err := crypto.DeriveKey(mnemonic, recoverySaltBytes())
	if err != nil {
		return nil, fmt.Errorf("recovery: failed to derive recovery key: %w", err)
	}

	encKey, err := crypto.DeriveSubKey(recoveryKey, RecoveryPurpose, 32)
	if err != nil {
		return nil, fmt.Errorf("recovery: derive sub-key: %w", err)
	}

	masterKey, err := crypto.Decrypt(encKey, blob, []byte("recovery"))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid recovery key", ErrRecoveryFailed)
	}
	return masterKey, nil
}
