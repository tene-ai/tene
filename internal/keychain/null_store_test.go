package keychain

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// Sprint v1014-rc1-qa-fixes / FX1.
//
// These tests pin invariant I-11: --no-keychain (selected via NullStore in
// cli/root.go) must persist NOTHING. Every previously-shipped escape path —
// a shared ~/.tene/keyfile, an OS keyring entry, even a transient temp file
// — is forbidden. The tests in this file are deliberately small and
// behavioural: each one asks "did anything land on disk or in the OS
// keychain?" and "does Load() refuse to return any cached key?".

func TestNullStore_StoreIsNoOp(t *testing.T) {
	ns := NewNullStore()
	if err := ns.Store([]byte("anything-32-bytes-of-key-material-here")); err != nil {
		t.Fatalf("Store should be a no-op success, got %v", err)
	}
}

func TestNullStore_LoadAlwaysReturnsErrKeyNotFound(t *testing.T) {
	ns := NewNullStore()
	// Even after a Store call, Load must refuse — that is the whole
	// point of NullStore (vs FileStore which would persist + return).
	_ = ns.Store([]byte("ignored"))
	key, err := ns.Load()
	if key != nil {
		t.Errorf("Load should return nil key, got %x", key)
	}
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("Load should return ErrKeyNotFound, got %v", err)
	}
}

func TestNullStore_DeleteIsNoOp(t *testing.T) {
	ns := NewNullStore()
	if err := ns.Delete(); err != nil {
		t.Fatalf("Delete should be a no-op success, got %v", err)
	}
}

func TestNullStore_ExistsAlwaysFalse(t *testing.T) {
	ns := NewNullStore()
	if ns.Exists() {
		t.Error("Exists must return false — NullStore never persists")
	}
	_ = ns.Store([]byte("anything"))
	if ns.Exists() {
		t.Error("Exists must return false even after Store — NullStore is a no-op")
	}
}

// TestNullStore_DoesNotTouchFilesystem verifies the headline behavioural
// promise: a NullStore round-trip must not create, modify, or stat any
// file under the user's home directory. We can't trivially observe the
// OS keychain from a unit test, but we can observe the filesystem.
//
// We compare a directory listing of HOME/.tene/ (if present) before and
// after a Store/Load/Delete cycle on a generously-sized synthetic key.
// If NullStore ever grew a sneaky disk-cache (regression risk), this test
// would catch the new path appearing.
func TestNullStore_DoesNotTouchFilesystem(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine HOME for filesystem check: %v", err)
	}
	teneDir := filepath.Join(home, ".tene")

	before := snapshotDir(t, teneDir)

	ns := NewNullStore()
	if err := ns.Store(make([]byte, 64)); err != nil {
		t.Fatalf("Store: %v", err)
	}
	if _, err := ns.Load(); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Load: expected ErrKeyNotFound, got %v", err)
	}
	if err := ns.Delete(); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	after := snapshotDir(t, teneDir)
	if before != after {
		t.Errorf("NullStore must not touch %s. before=%q after=%q", teneDir, before, after)
	}
}

// snapshotDir returns a deterministic representation of the directory
// suitable for byte-comparison. Missing directory is treated as empty —
// many CI hosts won't have ~/.tene at all.
func snapshotDir(t *testing.T, dir string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("ReadDir(%s): %v", dir, err)
	}
	out := ""
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		// name + size is enough; we are checking "did anything new
		// appear" — sizes will differ if a key was written.
		out += e.Name() + "(" + itoa(info.Size()) + ");"
	}
	return out
}

// itoa avoids pulling fmt into the snapshot helper.
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// TestNullStore_SatisfiesKeyStoreInterface is a compile-time check: if a
// future refactor renames KeyStore's methods, this test stops compiling
// before any product code regresses.
func TestNullStore_SatisfiesKeyStoreInterface(t *testing.T) {
	var ks KeyStore = NewNullStore()
	_ = ks // use the assignment so the linter doesn't complain
}
