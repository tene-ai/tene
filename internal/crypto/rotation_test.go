package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRotateProjectKey(t *testing.T) {
	owner, err := GenerateX25519KeyPair()
	require.NoError(t, err)

	alice, err := GenerateX25519KeyPair()
	require.NoError(t, err)
	bob, err := GenerateX25519KeyPair()
	require.NoError(t, err)

	members := []MemberKey{
		{UserID: "alice", PublicKey: alice.PublicKey},
		{UserID: "bob", PublicKey: bob.PublicKey},
	}

	result, err := RotateProjectKey(owner.PrivateKey, "team-project-1", members)
	require.NoError(t, err)

	assert.Len(t, result.NewProjectKey, ProjectKeySize)
	assert.Len(t, result.WrappedKeys, 2)

	// Alice can unwrap
	pkAlice, err := UnwrapProjectKey(alice.PrivateKey, owner.PublicKey,
		"team-project-1", "alice", result.WrappedKeys["alice"])
	require.NoError(t, err)
	assert.Equal(t, result.NewProjectKey, pkAlice)

	// Bob can unwrap
	pkBob, err := UnwrapProjectKey(bob.PrivateKey, owner.PublicKey,
		"team-project-1", "bob", result.WrappedKeys["bob"])
	require.NoError(t, err)
	assert.Equal(t, result.NewProjectKey, pkBob)
}

func TestRotateProjectKey_ExcludesRemoved(t *testing.T) {
	owner, _ := GenerateX25519KeyPair()
	alice, _ := GenerateX25519KeyPair()
	charlie, _ := GenerateX25519KeyPair()

	// Bob removed — only alice and charlie remain
	remaining := []MemberKey{
		{UserID: "alice", PublicKey: alice.PublicKey},
		{UserID: "charlie", PublicKey: charlie.PublicKey},
	}

	result, err := RotateProjectKey(owner.PrivateKey, "proj-x", remaining)
	require.NoError(t, err)

	// Only 2 wrapped keys (bob excluded)
	assert.Len(t, result.WrappedKeys, 2)
	_, hasBob := result.WrappedKeys["bob"]
	assert.False(t, hasBob, "removed member should not have wrapped key")
}
