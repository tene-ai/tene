package crypto

import (
	"bytes"
	"errors"
	"testing"
)

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error: %v", err)
	}
	if len(salt1) != SaltLen {
		t.Errorf("salt length = %d, want %d", len(salt1), SaltLen)
	}

	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error: %v", err)
	}
	if bytes.Equal(salt1, salt2) {
		t.Error("two salts should not be equal")
	}
}

func TestDeriveKey(t *testing.T) {
	salt, _ := GenerateSalt()
	key, err := DeriveKey("testpassword", salt)
	if err != nil {
		t.Fatalf("DeriveKey() error: %v", err)
	}
	if len(key) != ArgonKeyLen {
		t.Errorf("key length = %d, want %d", len(key), ArgonKeyLen)
	}
}

func TestDeriveKey_EmptyPassword(t *testing.T) {
	salt, _ := GenerateSalt()
	_, err := DeriveKey("", salt)
	if err == nil {
		t.Error("expected error for empty password")
	}
}

func TestDeriveKey_InvalidSaltLength(t *testing.T) {
	_, err := DeriveKey("password", []byte("short"))
	if !errors.Is(err, ErrInvalidSaltLength) {
		t.Errorf("expected ErrInvalidSaltLength, got %v", err)
	}
}

func TestDeriveSubKey(t *testing.T) {
	salt, _ := GenerateSalt()
	masterKey, _ := DeriveKey("testpassword", salt)

	subKey1, err := DeriveSubKey(masterKey, PurposeEncryption, 32)
	if err != nil {
		t.Fatalf("DeriveSubKey() error: %v", err)
	}
	if len(subKey1) != 32 {
		t.Errorf("subkey length = %d, want 32", len(subKey1))
	}

	subKey2, err := DeriveSubKey(masterKey, PurposeAuth, 32)
	if err != nil {
		t.Fatalf("DeriveSubKey() error: %v", err)
	}
	if bytes.Equal(subKey1, subKey2) {
		t.Error("sub keys for different purposes should differ")
	}
}

func TestDeriveSubKey_InvalidKeyLength(t *testing.T) {
	_, err := DeriveSubKey([]byte("short"), PurposeEncryption, 32)
	if !errors.Is(err, ErrInvalidKeyLength) {
		t.Errorf("expected ErrInvalidKeyLength, got %v", err)
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	plaintext := []byte("hello world secret")
	aad := []byte("MY_SECRET_KEY")

	ciphertext, err := Encrypt(key, plaintext, aad)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	decrypted, err := Decrypt(key, ciphertext, aad)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptDecrypt_EmptyPlaintext(t *testing.T) {
	key := make([]byte, 32)

	ciphertext, err := Encrypt(key, []byte{}, []byte("aad"))
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	decrypted, err := Decrypt(key, ciphertext, []byte("aad"))
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("expected empty plaintext, got %q", decrypted)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key2[0] = 1

	ciphertext, _ := Encrypt(key1, []byte("secret"), []byte("aad"))

	_, err := Decrypt(key2, ciphertext, []byte("aad"))
	if !errors.Is(err, ErrDecryptionFailed) {
		t.Errorf("expected ErrDecryptionFailed, got %v", err)
	}
}

func TestDecrypt_WrongAAD(t *testing.T) {
	key := make([]byte, 32)

	ciphertext, _ := Encrypt(key, []byte("secret"), []byte("aad1"))

	_, err := Decrypt(key, ciphertext, []byte("aad2"))
	if !errors.Is(err, ErrDecryptionFailed) {
		t.Errorf("expected ErrDecryptionFailed, got %v", err)
	}
}

func TestDecrypt_InvalidCiphertext(t *testing.T) {
	key := make([]byte, 32)

	_, err := Decrypt(key, []byte("short"), nil)
	if !errors.Is(err, ErrInvalidCiphertext) {
		t.Errorf("expected ErrInvalidCiphertext, got %v", err)
	}
}

func TestEncrypt_InvalidKeyLength(t *testing.T) {
	_, err := Encrypt([]byte("short"), []byte("data"), nil)
	if !errors.Is(err, ErrInvalidKeyLength) {
		t.Errorf("expected ErrInvalidKeyLength, got %v", err)
	}
}
