package keychain

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"
)

func TestFileStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "keyfile"))

	key := []byte("0123456789abcdef0123456789abcdef")

	err := store.Store(key)
	if err != nil {
		t.Fatalf("Store() error: %v", err)
	}

	if !store.Exists() {
		t.Error("Exists() = false, want true")
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if !bytes.Equal(key, loaded) {
		t.Errorf("loaded key does not match stored key")
	}
}

func TestFileStore_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "nonexistent"))

	_, err := store.Load()
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}

	if store.Exists() {
		t.Error("Exists() = true, want false")
	}
}

func TestFileStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "keyfile"))

	_ = store.Store([]byte("somekey"))

	err := store.Delete()
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	if store.Exists() {
		t.Error("Exists() = true after delete")
	}
}

func TestFileStore_DeleteNonExistent(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(filepath.Join(dir, "nonexistent"))

	err := store.Delete()
	if err != nil {
		t.Fatalf("Delete() should not error for non-existent file, got %v", err)
	}
}
