package crypto

// DeriveSubKey derives a sub-key from the master key using HKDF-SHA256.
// When salt is available (e.g. from vault_meta kdf_salt), use DeriveSubKeyWithSalt instead.
// Per RFC 5869 Section 3.1, providing a salt significantly strengthens key derivation.
func DeriveSubKey(masterKey []byte, purpose string, length int) ([]byte, error) {
	return DeriveSubKeyWithSalt(masterKey, nil, purpose, length)
}
