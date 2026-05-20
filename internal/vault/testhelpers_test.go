package vault

import "os"

// mkdirAllImpl is a tiny re-export of os.MkdirAll(p, 0700) so the
// per-feature test files (migration_test.go, metadata_test.go) can call
// it through a single named helper. Centralizing here keeps the
// permission mode aligned with what vault.New does in production.
func mkdirAllImpl(p string) error {
	return os.MkdirAll(p, 0700)
}
