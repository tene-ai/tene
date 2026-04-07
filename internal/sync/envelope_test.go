package sync

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSealOpen_RoundTrip(t *testing.T) {
	syncKey := make([]byte, 32)
	_, err := rand.Read(syncKey)
	require.NoError(t, err)

	plaintext := []byte("vault.db contents here -- secrets encrypted at L1")
	projectID := "my-project"
	env := "default"

	envelope, err := Seal(syncKey, plaintext, projectID, env)
	require.NoError(t, err)
	assert.True(t, len(envelope) > headerSize+24+16, "envelope should be larger than overhead")

	// Verify header magic
	assert.Equal(t, byte(0x54), envelope[0]) // 'T'
	assert.Equal(t, byte(0x45), envelope[1]) // 'E'
	assert.Equal(t, byte(0x4E), envelope[2]) // 'N'
	assert.Equal(t, byte(0x45), envelope[3]) // 'E'

	decrypted, err := Open(syncKey, envelope, projectID, env)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestOpen_WrongKey(t *testing.T) {
	syncKey := make([]byte, 32)
	_, _ = rand.Read(syncKey)
	wrongKey := make([]byte, 32)
	_, _ = rand.Read(wrongKey)

	envelope, err := Seal(syncKey, []byte("secret data"), "proj", "dev")
	require.NoError(t, err)

	_, err = Open(wrongKey, envelope, "proj", "dev")
	assert.Error(t, err, "should fail with wrong key")
}

func TestOpen_WrongAAD(t *testing.T) {
	syncKey := make([]byte, 32)
	_, _ = rand.Read(syncKey)

	envelope, err := Seal(syncKey, []byte("secret data"), "proj-a", "dev")
	require.NoError(t, err)

	_, err = Open(syncKey, envelope, "proj-b", "dev")
	assert.Error(t, err, "should fail with wrong project AAD")

	_, err = Open(syncKey, envelope, "proj-a", "staging")
	assert.Error(t, err, "should fail with wrong environment AAD")
}

func TestSeal_EmptyPlaintext(t *testing.T) {
	syncKey := make([]byte, 32)
	_, _ = rand.Read(syncKey)

	_, err := Seal(syncKey, nil, "proj", "dev")
	assert.Error(t, err, "should reject empty plaintext")

	_, err = Seal(syncKey, []byte{}, "proj", "dev")
	assert.Error(t, err, "should reject empty plaintext")
}

func TestSeal_InvalidKeyLength(t *testing.T) {
	_, err := Seal([]byte("short"), []byte("data"), "proj", "dev")
	assert.Error(t, err, "should reject short key")
}

func TestOpen_TruncatedEnvelope(t *testing.T) {
	syncKey := make([]byte, 32)
	_, _ = rand.Read(syncKey)

	_, err := Open(syncKey, []byte("short"), "proj", "dev")
	assert.Error(t, err, "should reject truncated envelope")
}

func TestOpen_InvalidMagic(t *testing.T) {
	syncKey := make([]byte, 32)
	_, _ = rand.Read(syncKey)

	envelope, err := Seal(syncKey, []byte("data"), "proj", "dev")
	require.NoError(t, err)

	// Corrupt magic bytes
	envelope[0] = 0xFF
	_, err = Open(syncKey, envelope, "proj", "dev")
	assert.Error(t, err, "should reject invalid magic")
}

func TestChecksum(t *testing.T) {
	data := []byte("hello world")
	hash := Checksum(data)
	assert.Len(t, hash, 32, "SHA-256 should be 32 bytes")

	// Same input should produce same hash
	hash2 := Checksum(data)
	assert.Equal(t, hash, hash2)

	// Different input should produce different hash
	hash3 := Checksum([]byte("different"))
	assert.NotEqual(t, hash, hash3)
}

func TestDeriveSyncKey(t *testing.T) {
	masterKey := make([]byte, 32)
	_, _ = rand.Read(masterKey)

	key1, err := DeriveSyncKey(masterKey)
	require.NoError(t, err)
	assert.Len(t, key1, 32)

	// Deterministic: same master key → same sync key
	key2, err := DeriveSyncKey(masterKey)
	require.NoError(t, err)
	assert.Equal(t, key1, key2)

	// Different master key → different sync key
	otherMaster := make([]byte, 32)
	_, _ = rand.Read(otherMaster)
	key3, err := DeriveSyncKey(otherMaster)
	require.NoError(t, err)
	assert.NotEqual(t, key1, key3)
}

func TestSealOpen_LargePayload(t *testing.T) {
	syncKey := make([]byte, 32)
	_, _ = rand.Read(syncKey)

	// 1MB payload
	largePlaintext := make([]byte, 1<<20)
	_, _ = rand.Read(largePlaintext)

	envelope, err := Seal(syncKey, largePlaintext, "big-project", "prod")
	require.NoError(t, err)

	decrypted, err := Open(syncKey, envelope, "big-project", "prod")
	require.NoError(t, err)
	assert.Equal(t, largePlaintext, decrypted)
}
