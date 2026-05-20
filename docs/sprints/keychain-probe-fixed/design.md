# Keychain Probe Service Name — Fixed-Identifier Fix

> **Sprint ID**: `keychain-probe-fixed` (follow-up to v1014-rc1-qa-fixes)
> **Working branch**: `fix/keychain-probe-service-fixed`
> **Trust Level**: L3 (single-feature follow-up; small scope)
> **Status**: design v1.0 — pending implementation
> **Triggering report**: v1.0.14 user feedback — macOS "키체인을 발견할 수 없음" dialogs piling up during QA session

---

## §0 Mission

Stop the v1.0.14 keychain probe from creating a **new macOS Keychain
service entry per project**. Replace the per-project `tene-<hash>` probe
target with a single fixed `tene-probe` service name. Master-key storage
itself stays per-project for isolation; only the *availability check*
(`NewStoreWithStatus`) is consolidated.

Two side effects, both desirable:

1. **Far fewer dialogs**. Today a developer who runs `tene` in N new
   project directories gets up to N "tene wants to access keychain"
   dialogs *just* from the probe path. After this fix they get one
   dialog total for the probe (the master-key save path still triggers
   one per project, but that is genuine ACL-relevant access).
2. **No more orphan accumulation**. The probe `Set`/`Delete` cycle
   leaves an ACL trace on the service entry even after the value is
   deleted; the next time the same service is touched, macOS asks the
   user again. With a single fixed probe service, that trace exists in
   exactly one place across the lifetime of the binary on this host.

## §1 Anti-Mission

- ❌ Change the master-key service name. `tene-<hashPath(projectDir)>`
  remains the production storage location — cross-project isolation
  is the whole point of that hash.
- ❌ Touch `pkg/`. tene-cloud must stay green; this is `internal/keychain`
  only.
- ❌ Bypass the probe. The probe exists because some hosts (CI, headless
  Linux, Docker) genuinely lack a usable OS keychain; we must still
  detect that and fall back to `FileStore`.
- ❌ Migrate existing `tene-<hash>` keychain entries. They are harmless
  leftovers; users who care can delete them via Keychain Access.app
  (see CHANGELOG note).

## §2 Root cause (from v1.0.14 user session)

`internal/keychain/keychain.go:155-168` (current v1.0.14):

```go
service := ServiceName + "-" + hashPath(projectPath)   // tene-<projectHash>
ks := NewKeyringStore(service)

testKey := "keychain-test"
if err := keyring.Set(service, testKey, "test"); err != nil {
    // file fallback
}
_ = keyring.Delete(service, testKey)
```

The probe `Set(service, "keychain-test", "test")` writes a dummy entry
under the **per-project service name** purely to confirm the OS keychain
accepts writes. The follow-up `Delete` cleans up the value but
**macOS still keeps the service entry registered in its metadata DB**,
along with an ACL grant ("tene is allowed to access this service").

Consequence on a developer machine that touches K new project dirs:

- K entries appear under `security dump-keychain | grep tene-`
- macOS prompts for ACL approval up to K times
- When K is large (during QA, K reached 95+) the keychain metadata DB
  can fall into transient inconsistency states that surface as the
  "키체인을 발견할 수 없음" dialog the user reported.

## §3 Design

### 3.1 New exported constant

`internal/keychain/keychain.go`:

```go
// ProbeServiceName is the fixed macOS Keychain service used by
// NewStoreWithStatus to test keychain availability. Sprint
// keychain-probe-fixed.
//
// One name across all projects on a host means the ACL prompt fires at
// most once per binary install, regardless of how many project
// directories the user touches. Master-key storage continues to use
// the per-project "tene-<hashPath(projectDir)>" service so vault keys
// stay isolated.
const ProbeServiceName = "tene-probe"
```

### 3.2 Modified `NewStoreWithStatus`

```go
service := ServiceName + "-" + hashPath(projectPath)
ks := NewKeyringStore(service)

// Availability probe. Sprint keychain-probe-fixed: this uses the
// fixed ProbeServiceName, NOT `service`, so we do not register a
// new keychain entry per project. The master-key store at `service`
// is only touched later when the user actually runs tene init / set.
if err := keyring.Set(ProbeServiceName, "probe", "ok"); err != nil {
    return NewFileStore(keyfilePath), FallbackInfo{
        Used:   true,
        Reason: "keychain_unavailable",
        Path:   keyfilePath,
    }
}
_ = keyring.Delete(ProbeServiceName, "probe")

return ks, FallbackInfo{Used: false}
```

### 3.3 Test harness hardening (rolled in from rc1 follow-up)

`internal/cli/testhelper_test.go` adds
`t.Setenv("TENE_KEYCHAIN_FALLBACK", "file")` in `setupTestEnv` so test
runs skip the OS keychain probe entirely. The CLI integration tests
exercise `KeyStore` contract uniformly via the file fallback;
`internal/keychain/keychain_test.go` covers the env-override branch
and the new probe-service test cover the OS path.

`internal/cli/no_keychain_integration_test.go` adds a
`testing.Short()` skip on the OS-keychain-probing test plus an
explicit unset of `TENE_KEYCHAIN_FALLBACK` so that test still genuinely
exercises `KeyringStore` selection when developers want to verify
keychain integration.

## §4 Test plan

### 4.1 New unit test in `internal/keychain/keychain_test.go`

`TestNewStoreWithStatus_ProbeUsesFixedService` — uses `keyring.MockInit()`
to swap the global provider for an in-memory map, then asserts:

1. Two successive `NewStoreWithStatus(projectA)` and
   `NewStoreWithStatus(projectB)` calls leave the mock store with
   **zero `tene-<hashA>`-keyed entries** and **zero
   `tene-<hashB>`-keyed entries**.
2. The fixed `ProbeServiceName` has no leftover value either (the
   probe's `Delete` cleaned it up).
3. The returned `KeyStore` is the project-specific `*KeyringStore`
   (`Used == false`).

That triple covers: (a) per-project hash service is never the probe
target, (b) no value leaks past the probe, (c) the user still gets the
correct production store handle.

### 4.2 Regression coverage

- All existing tests in `internal/keychain/` and `internal/cli/...`
  must remain green.
- `golangci-lint run` 0 issues.
- `tene-cloud` cross-repo build + tests (G3 gate) — `pkg/` unchanged so
  this should be a no-op but stays a hard gate.

## §5 Migration

Existing `tene-<hashPath>` macOS Keychain entries from prior versions
are harmless artifacts of the v1.0.14 (and earlier) probe path. They
do not contain any secret value (the probe `Delete` removed the
"keychain-test" value), only the service registration + ACL.

Users who want to clean them up:

```bash
# Inventory
security dump-keychain 2>/dev/null | grep -oE '"tene-[a-f0-9]+"' | sort -u

# Delete (sandbox-only entries — safe)
for s in $(security dump-keychain 2>/dev/null | grep -oE '"tene-[a-f0-9]+"' | tr -d '"' | sort -u); do
  security delete-generic-password -s "$s" 2>/dev/null
done
```

The CHANGELOG entry will include the snippet above so users can clean
up at their leisure. The probe-service fix itself does NOT need an
explicit migration — new tene-probe entry is created on first probe
after upgrade; old entries are inert.

## §6 Acceptance criteria

- [ ] `ProbeServiceName` constant exported from `internal/keychain`
- [ ] `NewStoreWithStatus` uses `ProbeServiceName` (not the per-project
  service) for the availability probe
- [ ] New unit test `TestNewStoreWithStatus_ProbeUsesFixedService`
  exercises the contract under `keyring.MockInit()`
- [ ] Test harness `setupTestEnv` sets `TENE_KEYCHAIN_FALLBACK=file`
  (P2 follow-up applied)
- [ ] `TestSelectKeyStore_NoFlag_UsesOSKeychainOrFileFallback` gains a
  `testing.Short()` skip + explicit `TENE_KEYCHAIN_FALLBACK` unset
- [ ] Full `go test -race ./...` green
- [ ] `golangci-lint run` 0 issues
- [ ] `cd ../tene-cloud && go build ./... && go test -race ./...` green
- [ ] CHANGELOG `[Unreleased]` Fixed entry naming the probe-service
  regression with the macOS cleanup snippet
- [ ] SECURITY.md "Master Key Storage Modes" section gets a one-line
  note distinguishing the per-project master-key service from the
  fixed probe service
- [ ] PR opened against `staging`

## §7 Risks

| Risk | Mitigation |
|---|---|
| MockInit alters a global; later tests in the same package run see the mock provider | `t.Cleanup` cannot truly restore the previous provider (library exposes no setter). Document the test order requirement; place the mock test last by sorted name (Go testing's default order is source code order, so put the test at the bottom of the file). |
| Cross-platform: Linux libsecret and Windows Credential Manager also use `service` parameter — must still work with the fixed `ProbeServiceName` | The go-keyring API is uniform: `Set(service, user, password)` is implemented for each platform. Linux libsecret organises secrets by `application` attribute (mapped from `service`); a fixed service is fine there too. Windows Credential Manager keys by `service + user` and similarly accepts the fixed name. |
| Users running a CI image without secret-service installed | Existing behaviour preserved: `keyring.Set` fails → `FallbackInfo{Used: true, Reason: "keychain_unavailable"}`. The fix changes only the service name, not the fallback path. |

## §8 Cross-repo impact

- **tene-cloud**: unchanged. `pkg/` not touched.
- **tene-biz**: docs/05-qa update once this lands (post-fix QA report follow-up).
- **apps/web**: unchanged.
- **homebrew tap, GHCR**: next release tag (likely `v1.0.15` if other
  changes accumulate, or `v1.0.14.1` as a patch) ships this fix.

## §9 Reference

- Sprint `v1014-rc1-qa-fixes` `FX1.md` — original keychain fix
  (introduced `NullStore` and the `selectKeyStore` precedence). This
  follow-up is strictly additive on top of that.
- `tene-biz/docs/05-qa/tene-cli-v1.0.14-postfix.qa-report.md` §5 — the
  "Outstanding items / deferred to v1.0.15" line that called for this
  cleanup.
