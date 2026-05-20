# FX1 — Keychain Per-Project Isolation + `--no-keychain` Semantics

> **Bug**: B1 (CRITICAL — security)
> **Invariant introduced**: I-11
> **Files touched**: `internal/keychain/null_store.go` (new), `internal/keychain/keychain.go`, `internal/cli/root.go`, `internal/cli/init.go`, `CHANGELOG.md`, `SECURITY.md`, `README.md`
> **PR target**: `staging`
> **Commit prefix**: `fix(keychain):` + `fix(cli):` (split commits per file area)

## 1. Problem (from QA report B1)

QA에서 다음을 sandbox2에서 재현:

```bash
$ tene init sb2 --no-keychain
  Master Key saved to OS Keychain          # ← 거짓 메시지 (실제는 ~/.tene/keyfile)

$ TENE_MASTER_PASSWORD='WRONG' tene get CANARY --unsafe-stdout --no-keychain
canary_val_sb2                              # ← wrong password로도 복호화

$ unset TENE_MASTER_PASSWORD
$ tene get CANARY --unsafe-stdout --no-keychain < /dev/null
canary_val_sb2                              # ← password 없이도 복호화
```

3개 sub-원인:
- **B1-a** `keychain.go:145` — `~/.tene/keyfile` 단일 파일이 모든 프로젝트 공유
- **B1-b** `init.go:213` — storage 종류 무관 고정 메시지 ("Master Key saved to OS Keychain")
- **B1-c** `root.go:307-309` — `--no-keychain`가 file store로 fallback (이름과 반대)

## 2. Design

### 2.1 새 타입: `keychain.NullStore`

`internal/keychain/null_store.go` (new):

- `Store(key []byte) error` — no-op, returns nil
- `Load() ([]byte, error)` — 항상 `ErrKeyNotFound` 반환
- `Delete() error` — no-op
- `Exists() bool` — `false`

의미: "어떤 영구 저장소에도 master key를 저장하지 않음". `loadOrPromptMasterKey()`에서 `Load()` 실패 후 `TENE_MASTER_PASSWORD` env var → interactive prompt 순으로 폴백되는 기존 path를 활용.

### 2.2 `--no-keychain` 의 새 의미

`internal/cli/root.go:loadApp()`:

```go
if flagNoKeychain {
    if envKeyfile := os.Getenv("TENE_KEYFILE"); envKeyfile != "" {
        // Explicit opt-in to file-based store at user-specified path.
        // Use case: long-running daemons that need to skip the OS
        // keychain but cannot pipe a password on every call.
        ks = keychain.NewFileStore(envKeyfile)
    } else {
        // Default: no persistent storage. Every call must resolve the
        // master password from TENE_MASTER_PASSWORD or interactive prompt.
        // I-11 (sprint v1014-rc1-qa-fixes).
        ks = keychain.NewNullStore()
    }
}
```

### 2.3 `init.go` storage 메시지 분기

Step 9의 `--no-keychain` 분기를 root.go와 동일한 NullStore/FileStore 선택으로 통일. Step 14의 출력은 `ks` 의 concrete type을 보고 분기:

| KeyStore concrete type | 메시지 |
|---|---|
| `*keychain.KeyringStore` | `Master Key saved to OS Keychain` |
| `*keychain.FileStore` (env override) | `Master Key saved to <path> (via TENE_KEYFILE)` |
| `*keychain.FileStore` (자동 fallback) | `Master Key saved to <path> (OS keychain unavailable)` |
| `*keychain.NullStore` | `Master Key NOT persisted (--no-keychain).\n  Provide TENE_MASTER_PASSWORD on every command.` |

자동 fallback인지 사용자 의도(env override)인지 구분은 `init.go`가 직접 알 수 있도록 `--no-keychain` 플래그 + `TENE_KEYFILE` env 둘 다 보고 결정.

### 2.4 backward compatibility — `TENE_KEYFILE` escape hatch

기존 사용자가 `--no-keychain` + `~/.tene/keyfile`에 의존하고 있다면 동일 path를 `TENE_KEYFILE=$HOME/.tene/keyfile` 로 명시. CHANGELOG에 1-liner migration 안내:

```
BREAKING: --no-keychain no longer writes master key to ~/.tene/keyfile by default.
To preserve previous behavior: export TENE_KEYFILE=$HOME/.tene/keyfile-<project-hash>
(or use a project-specific path you control).
```

`~/.tene/keyfile`가 이미 디스크에 있는 경우는 그대로 두고 (다른 프로세스가 의존할 수 있음), 새 path는 사용자 책임으로 선택하게 함.

### 2.5 새 invariant I-11 (master-plan Appendix B 참조)

> `--no-keychain` 가 keychain *및* file-fallback 양쪽 모두 우회. 매 호출 password 입력 강제 (env var or stdin or TTY prompt).

## 3. Test plan

### 3.1 Unit tests — `internal/keychain/null_store_test.go` (new)

| Case | Expected |
|---|---|
| `NewNullStore().Store(arbitrary_key)` | `nil` (no-op success) |
| `NewNullStore().Load()` | `(nil, ErrKeyNotFound)` |
| `NewNullStore().Delete()` | `nil` |
| `NewNullStore().Exists()` | `false` |
| Store does NOT touch filesystem | `os.Stat("~/.tene/keyfile")` unchanged before/after |

### 3.2 Integration test — `internal/cli/no_keychain_integration_test.go` (new)

| Case | Setup | Expected |
|---|---|---|
| I-11 (a) project isolation | init projA (--no-keychain, pwA); init projB (--no-keychain, pwB) | each vault decrypts only with its own pw |
| I-11 (b) wrong password rejected | init proj with pwA; get with pwB (--no-keychain) | exit non-zero, no plaintext |
| I-11 (c) missing password rejected | unset TENE_MASTER_PASSWORD; get (--no-keychain, no stdin) | exit non-zero with "interactive terminal required" |
| backward compat | TENE_KEYFILE=/tmp/test-keyfile; init (--no-keychain); restart shell; get | succeeds without password (file store cache) |
| init message — NullStore | tene init (--no-keychain) | stdout matches "NOT persisted" |
| init message — FileStore (env) | TENE_KEYFILE=/tmp/x tene init (--no-keychain) | stdout matches "via TENE_KEYFILE" |
| init message — KeyringStore | tene init (default) | stdout matches "saved to OS Keychain" |

### 3.3 Manual sandbox replay

```bash
rm -rf /tmp/fx1-sb /tmp/fx1-keyfile
mkdir /tmp/fx1-sb && cd /tmp/fx1-sb
TENE_MASTER_PASSWORD=passA tene init proj1 --no-keychain --claude
TENE_MASTER_PASSWORD=passB tene get DUMMY --unsafe-stdout --no-keychain 2>&1 | head -5
# expect: error (no DUMMY key set), but with WRONG password the error must NOT
#         contain a decrypted value.

TENE_MASTER_PASSWORD=passA tene set DUMMY hello --no-keychain
TENE_MASTER_PASSWORD=passB tene get DUMMY --unsafe-stdout --no-keychain 2>&1 | head -5
# expect: decryption failure, not "hello"

unset TENE_MASTER_PASSWORD
tene get DUMMY --unsafe-stdout --no-keychain < /dev/null 2>&1 | head -5
# expect: "interactive terminal required" or "TENE_MASTER_PASSWORD required"
```

## 4. Risks specific to FX1

- **시스템에 ~/.tene/keyfile이 이미 있는 사용자**: 본 변경 후에도 그 파일은 그대로 두지만, `--no-keychain` 호출에서 더 이상 사용 안 함. `tene` cleanup 명령은 별도 sprint (out-of-scope).
- **TENE_KEYFILE path가 직접 sensitive 경로일 때**: 사용자 책임. CHANGELOG에 권한 0600 권고.
- **NullStore.Store()가 no-op이라 init 시 Warning이 안 뜸**: 이건 의도된 동작 — `--no-keychain` 의미가 그것이므로 메시지를 "NOT persisted"로 분명히 한다.

## 5. Acceptance criteria

- [ ] 3.1의 unit test 4 cases 모두 PASS
- [ ] 3.2의 integration test 7 cases 모두 PASS
- [ ] 3.3의 manual sandbox replay 결과가 모두 의도된 거부 동작
- [ ] `go test -race ./...` 전체 통과
- [ ] `golangci-lint run` 0 issues
- [ ] tene-cloud `go build ./...` 회귀 0건 (G3)
- [ ] CHANGELOG.md에 BREAKING note 1줄 추가
- [ ] SECURITY.md에 keychain fallback 정책 1단락 추가
- [ ] README.md (해당 섹션이 있다면)에 `TENE_KEYFILE` 1줄
