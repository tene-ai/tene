package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateX25519KeyPair(t *testing.T) {
	kp, err := GenerateX25519KeyPair()
	require.NoError(t, err)
	assert.Len(t, kp.PrivateKey, X25519KeySize)
	assert.Len(t, kp.PublicKey, X25519KeySize)

	// Two key pairs should be different
	kp2, err := GenerateX25519KeyPair()
	require.NoError(t, err)
	assert.NotEqual(t, kp.PrivateKey, kp2.PrivateKey)
	assert.NotEqual(t, kp.PublicKey, kp2.PublicKey)
}

func TestX25519SharedSecret_Commutativity(t *testing.T) {
	alice, err := GenerateX25519KeyPair()
	require.NoError(t, err)
	bob, err := GenerateX25519KeyPair()
	require.NoError(t, err)

	// Alice computes shared secret with Bob's public key
	sharedA, err := X25519SharedSecret(alice.PrivateKey, bob.PublicKey)
	require.NoError(t, err)

	// Bob computes shared secret with Alice's public key
	sharedB, err := X25519SharedSecret(bob.PrivateKey, alice.PublicKey)
	require.NoError(t, err)

	// ECDH commutativity: both should be identical
	assert.Equal(t, sharedA, sharedB)
}

func TestX25519SharedSecret_InvalidKeyLength(t *testing.T) {
	_, err := X25519SharedSecret([]byte("short"), make([]byte, 32))
	assert.Error(t, err)

	_, err = X25519SharedSecret(make([]byte, 32), []byte("short"))
	assert.Error(t, err)
}

func TestDeriveSubKeyWithSalt(t *testing.T) {
	key := make([]byte, 32)
	key[0] = 0x42

	// Same inputs → same output (deterministic)
	k1, err := DeriveSubKeyWithSalt(key, []byte("salt1"), "purpose", 32)
	require.NoError(t, err)
	k2, err := DeriveSubKeyWithSalt(key, []byte("salt1"), "purpose", 32)
	require.NoError(t, err)
	assert.Equal(t, k1, k2)

	// Different salt → different output
	k3, err := DeriveSubKeyWithSalt(key, []byte("salt2"), "purpose", 32)
	require.NoError(t, err)
	assert.NotEqual(t, k1, k3)

	// Different purpose → different output
	k4, err := DeriveSubKeyWithSalt(key, []byte("salt1"), "other-purpose", 32)
	require.NoError(t, err)
	assert.NotEqual(t, k1, k4)
}
