package encfile

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/agent-kay-it/tene/pkg/crypto"
)

var (
	// MagicBytes: "TENE"
	MagicBytes = [4]byte{'T', 'E', 'N', 'E'}

	// FormatVersion is the current format version.
	FormatVersion byte = 0x01

	// KDFAlgArgon2id is the KDF algorithm identifier for Argon2id.
	KDFAlgArgon2id byte = 0x01

	// HeaderSize is the total header size in bytes.
	// Magic(4) + Version(1) + KDFAlg(1) + Memory(4) + Iterations(4) + Parallelism(1) + SaltLen(1) + Salt(16) + Nonce(24) = 56
	HeaderSize = 56
)

var (
	ErrInvalidMagic   = errors.New("encfile: invalid magic bytes, not a .tene.enc file")
	ErrUnsupportedVer = errors.New("encfile: unsupported format version")
	ErrFileTooShort   = errors.New("encfile: file too short")
)

// Header represents a .tene.enc file header.
type Header struct {
	FormatVersion byte
	KDFAlgorithm  byte
	KDFMemory     uint32 // KB
	KDFIterations uint32
	KDFParallel   byte
	Salt          [16]byte
	Nonce         [24]byte
}

// Encode serializes the Header to bytes.
func (h *Header) Encode() []byte {
	buf := new(bytes.Buffer)

	buf.Write(MagicBytes[:])                                    // 4 bytes
	buf.WriteByte(h.FormatVersion)                              // 1 byte
	buf.WriteByte(h.KDFAlgorithm)                               // 1 byte
	_ = binary.Write(buf, binary.LittleEndian, h.KDFMemory)    // 4 bytes
	_ = binary.Write(buf, binary.LittleEndian, h.KDFIterations) // 4 bytes
	buf.WriteByte(h.KDFParallel)                                // 1 byte
	buf.WriteByte(16)                                           // salt length, 1 byte
	buf.Write(h.Salt[:])                                        // 16 bytes
	buf.Write(h.Nonce[:])                                       // 24 bytes

	return buf.Bytes() // 56 bytes total
}

// DecodeHeader parses a Header from bytes.
func DecodeHeader(data []byte) (*Header, error) {
	if len(data) < HeaderSize {
		return nil, ErrFileTooShort
	}

	// Verify magic bytes
	if !bytes.Equal(data[0:4], MagicBytes[:]) {
		return nil, ErrInvalidMagic
	}

	h := &Header{
		FormatVersion: data[4],
		KDFAlgorithm:  data[5],
		KDFParallel:   data[14],
	}

	if h.FormatVersion != FormatVersion {
		return nil, fmt.Errorf("%w: got %d", ErrUnsupportedVer, h.FormatVersion)
	}

	h.KDFMemory = binary.LittleEndian.Uint32(data[6:10])
	h.KDFIterations = binary.LittleEndian.Uint32(data[10:14])

	// saltLen := data[15] -- currently always 16
	copy(h.Salt[:], data[16:32])
	copy(h.Nonce[:], data[32:56])

	return h, nil
}

// Encrypt encrypts plaintext into the .tene.enc binary format.
// password: Master Password
// plaintext: data to encrypt (typically JSON)
// Returns: full .tene.enc binary (header + encrypted payload)
func Encrypt(password string, plaintext []byte) ([]byte, error) {
	// 1. Generate salt
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("encfile: failed to generate salt: %w", err)
	}

	// 2. Derive key via KDF
	masterKey, err := crypto.DeriveKey(password, salt)
	if err != nil {
		return nil, fmt.Errorf("encfile: KDF failed: %w", err)
	}
	defer crypto.ZeroBytes(masterKey)

	encKey, err := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)
	if err != nil {
		return nil, fmt.Errorf("encfile: sub-key derivation failed: %w", err)
	}
	defer crypto.ZeroBytes(encKey)

	// 3. Encrypt (crypto.Encrypt prepends nonce)
	ciphertext, err := crypto.Encrypt(encKey, plaintext, []byte("tene-export"))
	if err != nil {
		return nil, fmt.Errorf("encfile: encryption failed: %w", err)
	}

	// ciphertext = nonce(24) + encrypted_payload
	// Store nonce separately in header, remove from payload
	var nonce [24]byte
	copy(nonce[:], ciphertext[:24])
	payload := ciphertext[24:]

	// 4. Build header
	var saltArr [16]byte
	copy(saltArr[:], salt)

	header := &Header{
		FormatVersion: FormatVersion,
		KDFAlgorithm:  KDFAlgArgon2id,
		KDFMemory:     uint32(crypto.ArgonMemory),
		KDFIterations: uint32(crypto.ArgonTime),
		KDFParallel:   uint8(crypto.ArgonThreads),
		Salt:          saltArr,
		Nonce:         nonce,
	}

	// 5. Assemble: header + payload
	headerBytes := header.Encode()
	result := make([]byte, len(headerBytes)+len(payload))
	copy(result, headerBytes)
	copy(result[len(headerBytes):], payload)

	return result, nil
}

// Decrypt decrypts a .tene.enc binary file.
// password: Master Password
// data: full .tene.enc binary
// Returns: decrypted plaintext
func Decrypt(password string, data []byte) ([]byte, error) {
	// 1. Parse header
	header, err := DecodeHeader(data)
	if err != nil {
		return nil, err
	}

	// 2. Derive key via KDF (using header's salt)
	masterKey, err := crypto.DeriveKey(password, header.Salt[:])
	if err != nil {
		return nil, fmt.Errorf("encfile: KDF failed: %w", err)
	}
	defer crypto.ZeroBytes(masterKey)

	encKey, err := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)
	if err != nil {
		return nil, fmt.Errorf("encfile: sub-key derivation failed: %w", err)
	}
	defer crypto.ZeroBytes(encKey)

	// 3. Reassemble nonce + payload for crypto.Decrypt
	payload := data[HeaderSize:]
	ciphertext := make([]byte, 24+len(payload))
	copy(ciphertext[:24], header.Nonce[:])
	copy(ciphertext[24:], payload)

	plaintext, err := crypto.Decrypt(encKey, ciphertext, []byte("tene-export"))
	if err != nil {
		return nil, fmt.Errorf("encfile: decryption failed (wrong password?): %w", err)
	}

	return plaintext, nil
}
