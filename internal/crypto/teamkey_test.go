package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapUnwrapProjectKey_RoundTrip(t *testing.T) {
	// Owner and member generate key pairs
	owner, err := GenerateX25519KeyPair()
	require.NoError(t, err)
	member, err := GenerateX25519KeyPair()
	require.NoError(t, err)

	// Generate a project key
	projectKey, err := GenerateProjectKey()
	require.NoError(t, err)
	assert.Len(t, projectKey, ProjectKeySize)

	projectID := "team-123-project"
	memberUserID := "user-alice"

	// Owner wraps project key for member
	wrapped, err := WrapProjectKey(owner.PrivateKey, member.PublicKey,
		projectID, memberUserID, projectKey)
	require.NoError(t, err)
	assert.True(t, len(wrapped) > ProjectKeySize, "wrapped should be larger than plaintext")

	// Member unwraps using their private key + owner's public key
	unwrapped, err := UnwrapProjectKey(member.PrivateKey, owner.PublicKey,
		projectID, memberUserID, wrapped)
	require.NoError(t, err)
	assert.Equal(t, projectKey, unwrapped)
}

func TestWrapUnwrapProjectKey_WrongRecipient(t *testing.T) {
	owner, _ := GenerateX25519KeyPair()
	member, _ := GenerateX25519KeyPair()
	attacker, _ := GenerateX25519KeyPair()
	projectKey, _ := GenerateProjectKey()

	wrapped, err := WrapProjectKey(owner.PrivateKey, member.PublicKey,
		"project-1", "alice", projectKey)
	require.NoError(t, err)

	// Attacker tries to unwrap with their own key → should fail
	_, err = UnwrapProjectKey(attacker.PrivateKey, owner.PublicKey,
		"project-1", "alice", wrapped)
	assert.Error(t, err, "wrong private key should fail")
}

func TestWrapUnwrapProjectKey_WrongAAD(t *testing.T) {
	owner, _ := GenerateX25519KeyPair()
	member, _ := GenerateX25519KeyPair()
	projectKey, _ := GenerateProjectKey()

	wrapped, err := WrapProjectKey(owner.PrivateKey, member.PublicKey,
		"project-1", "alice", projectKey)
	require.NoError(t, err)

	// Wrong project ID → should fail (different HKDF salt)
	_, err = UnwrapProjectKey(member.PrivateKey, owner.PublicKey,
		"project-WRONG", "alice", wrapped)
	assert.Error(t, err, "wrong project ID should fail")

	// Wrong user ID → should fail (different AAD)
	_, err = UnwrapProjectKey(member.PrivateKey, owner.PublicKey,
		"project-1", "bob", wrapped)
	assert.Error(t, err, "wrong user ID should fail")
}

func TestWrapProjectKey_InvalidKeyLength(t *testing.T) {
	owner, _ := GenerateX25519KeyPair()
	member, _ := GenerateX25519KeyPair()

	_, err := WrapProjectKey(owner.PrivateKey, member.PublicKey,
		"proj", "user", []byte("too-short"))
	assert.Error(t, err, "short project key should fail")
}

func TestGenerateProjectKey(t *testing.T) {
	k1, err := GenerateProjectKey()
	require.NoError(t, err)
	assert.Len(t, k1, ProjectKeySize)

	k2, err := GenerateProjectKey()
	require.NoError(t, err)
	assert.NotEqual(t, k1, k2, "two project keys should differ")
}
