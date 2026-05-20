package keychain

// NullStore is a no-persistence KeyStore selected when --no-keychain is set
// and TENE_KEYFILE is not configured. Sprint v1014-rc1-qa-fixes / FX1.
//
// Why this exists
//
// Prior to v1.0.14 the --no-keychain flag fell back to a single shared
// FileStore at ~/.tene/keyfile (per-user, not per-project). Two consequences
// followed and were caught by QA (B1, CRITICAL):
//
//   1. Cross-project bleed. Project A's master key landed in
//      ~/.tene/keyfile; a later "tene init --no-keychain" inside project B
//      overwrote it with B's key. Any subsequent `tene get --no-keychain`
//      inside A succeeded with the wrong password because Load() returned
//      B's cached key. The vault.db itself was per-project, but the keystore
//      was global.
//
//   2. CI/CD password verification bypass. A pipeline that intentionally
//      passes a wrong TENE_MASTER_PASSWORD to test rejection still decrypted
//      because Load() returned a cached key from a previous run.
//
// NullStore restores the invariant that "--no-keychain means no persistent
// storage of the master key". Every CLI invocation must resolve the key
// from TENE_MASTER_PASSWORD env var or interactive prompt — there is no
// cache to leak from. This is invariant I-11.
//
// Users who want a file-based store with this flag must opt in explicitly
// via TENE_KEYFILE=<absolute path>; that path is then used through the
// existing FileStore implementation (fallback.go) and the user is
// responsible for the file's location and permissions.
type NullStore struct{}

// NewNullStore returns a KeyStore that never persists the master key.
//
// Use case: --no-keychain mode without an explicit TENE_KEYFILE override.
// The intent is that the master key is held only in process memory for
// the duration of a single CLI invocation.
func NewNullStore() *NullStore { return &NullStore{} }

// Store is a no-op. NullStore deliberately discards the key so that
// nothing persists between invocations.
//
// Returning nil (success) preserves the contract callers expect:
// "Store should not fail in normal operation". The non-persistence is
// communicated to the user via init.go's keystore-aware status message,
// not via an error here.
func (n *NullStore) Store(_ []byte) error { return nil }

// Load always returns ErrKeyNotFound. This forces the caller
// (loadOrPromptMasterKey in cli/root.go) to fall through to the
// TENE_MASTER_PASSWORD env var path and, failing that, to the
// interactive prompt path — exactly what --no-keychain promises.
func (n *NullStore) Load() ([]byte, error) {
	return nil, ErrKeyNotFound
}

// Delete is a no-op. There is nothing to remove because Store never
// persisted anything. Returning nil keeps the KeyStore interface
// uniform for callers that don't care which concrete store they hold.
func (n *NullStore) Delete() error { return nil }

// Exists always reports false: there is no persisted key. This is the
// truthful answer for NullStore and the signal callers use to decide
// whether to prompt or to attempt Load first.
func (n *NullStore) Exists() bool { return false }
