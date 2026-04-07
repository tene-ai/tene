package crypto

import "fmt"

// MemberKey holds a team member's public key for key rotation re-wrapping.
type MemberKey struct {
	UserID    string
	PublicKey []byte // X25519 public key (32 bytes)
}

// RotationResult holds the output of a key rotation operation.
type RotationResult struct {
	NewProjectKey []byte            // new PK (256-bit)
	WrappedKeys   map[string][]byte // userID -> wrapped new PK
}

// RotateProjectKey generates a new project key and re-wraps it for all remaining members.
//
// Key rotation protocol (per design §2.3):
//  1. Generate new PK' (CSPRNG)
//  2. For each remaining member: ECDH(ownerPrivate, memberPublic) → wrap PK'
//  3. Caller must re-encrypt vault secrets with PK' and sync
func RotateProjectKey(ownerPrivateKey []byte, projectID string, members []MemberKey) (*RotationResult, error) {
	// 1. Generate new project key
	newPK, err := GenerateProjectKey()
	if err != nil {
		return nil, fmt.Errorf("crypto: rotate: generate new PK: %w", err)
	}

	// 2. Wrap new PK for each remaining member
	wrappedKeys := make(map[string][]byte, len(members))
	for _, m := range members {
		wrapped, wrapErr := WrapProjectKey(ownerPrivateKey, m.PublicKey, projectID, m.UserID, newPK)
		if wrapErr != nil {
			return nil, fmt.Errorf("crypto: rotate: wrap for %s: %w", m.UserID, wrapErr)
		}
		wrappedKeys[m.UserID] = wrapped
	}

	return &RotationResult{
		NewProjectKey: newPK,
		WrappedKeys:   wrappedKeys,
	}, nil
}
