# tene CLI 완성도 감사 — 2026-05-11

> **방법**: 10명의 전문 리뷰어 에이전트가 병렬로 실행되었으며, 각 에이전트는
> `/Users/popup-kay/Documents/GitHub/agentkay/tene-biz/tene` 저장소에 대해 **한 가지 관점**
> 만 검토하도록 범위가 한정되었다. 각 에이전트는 표준화된 보고서를 반환했다 (점수
> 0-10, 핵심 강점, P0-P3 개선 사항 + file:line 근거, quick wins, big bets). 본 문서는
> 10개 보고서를 단일 우선순위 액션 레지스터로 통합한 결과다.
>
> **코드베이스 스냅샷**: 7,504 non-test LOC, 2,763 test LOC (37% 비율), Go 1.25,
> 주요 의존성: cobra, testify, go-bip39, zalando/go-keyring, golang.org/x/crypto,
> modernc.org/sqlite. 아키텍처: `cmd/tene/` (엔트리) →
> `internal/{cli, vault, sync, keychain, recovery, claudemd, encfile, config}` →
> `pkg/{crypto, domain, errors}`.

---

## 1. 스코어카드

| # | 관점 | 점수 | 한 줄 평가 |
|---|---|:---:|---|
| A1 | Security & Cryptography | **7.0** | AEAD 배선은 교과서적이지만 `tene passwd`가 old password 검증을 건너뜀 (P0) |
| A2 | Clean Architecture & SOLID | **5.0** | `internal/cli`가 god-package; `pkg/crypto`가 delivery 레이어로 누설 |
| A3 | Platform-Native Auth (Touch ID) | **3.0** | 생체인증 전무; OS keychain을 raw key cache로만 사용, trust anchor 아님 |
| A4 | Go Idioms & Convention | **6.5** | `pkg/errors`가 stdlib을 그림자처리; 린터 세트 너무 좁음; `t.Parallel()` 0개 |
| A5 | CLI UX & DX | **7.5** | `tene run --` 디자인 best-in-class; exit code가 문서와 불일치 |
| A6 | Test Strategy | **5.5** | crypto/sync 단위 테스트 양호; fuzz 0개; CLI 명령 ~50% 미커버 |
| A7 | Performance & Scalability | **6.5** | cold start 43ms 수용 가능; 벤치마크 0개; selective-decrypt 부재 |
| A8 | Cross-Platform Parity | **5.5** | CI Ubuntu only; `~/.tene` 하드코딩; `tene update` Windows 404 버그 |
| A9 | Distribution & Supply Chain | **5.0** | signing/SBOM/SLSA provenance 전무; Homebrew tap 비활성 처리됨 |
| A10 | AI Integration & MCP | **8.0** | 카테고리 최고 — 5-editor rule emission, non-TTY refusal; MCP server 아직 없음 |

**평균: 5.95 / 10**

### 종합 해석

- **암호 기반은 견고함** (A1=7, A7=6.5, A10=8). 제품을 비밀 보관용으로 신뢰할 수 있는 수준.
- **구조적 부채가 지배적 위험** (A2=5, A4=6.5, A6=5.5). 안전하게 진화시키기 어려움.
- **릴리스 파이프라인의 신뢰 격차는 보안 도구로서 이례적임** (A9=5). 서명·증명 없음.
- **AI 에이전트 포지셔닝이 가장 강한 차별점** (A10=8) 이지만 누락 기능(MCP server) 이 빠르게 격차를 좁힘.
- **생체인증 스토리가 전혀 없음** (A3=3) — 현재 가용한 가장 높은 레버리지의 UX 투자.

---

## 2. 관점 교차 P0 크리티컬 패스 (반드시 빠르게 고칠 것)

여러 에이전트에서 P0로 surface 되었거나 단독으로도 릴리스 차단급으로 심각한 6개.

### P0-1. `tene passwd`가 old password를 검증하지 않음 — A1
- **근거**: `internal/cli/passwd.go:30-34`가 `loadOrPromptMasterKey()`를 호출하는데, 이 함수는 `root.go:175-178`에서 typed password를 prompt/비교 없이 keychain 사본을 조용히 반환. 프롬프트 레이블은 출력되지만 그 뒤에 `term.ReadPassword` 호출이 없음. 상수 `PurposeAuth = "tene-auth-hash"`는 `pkg/crypto/kdf.go:25`에 이미 존재하지만 사용되지 않음.
- **위험**: 로그인된 세션에 셸 접근 권한이 있는 사람은 현재 password를 모르고도 master password를 회전 가능.
- **수정**: init 시 `HKDF(masterKey, "tene-auth-hash")`를 `vault_meta`에 `auth_hash`로 저장. `passwd`/`recover`에서 항상 prompt → 후보 키 도출 → `subtle.ConstantTimeCompare` 비교.
- **노력**: S (2-3시간).

### P0-2. `encfile` header KDF 파라미터가 장식용에 불과 — A1
- **근거**: `internal/encfile/encfile.go:163`이 `crypto.DeriveKey(password, header.Salt[:])` 를 호출하는데, 하드코딩된 `crypto.ArgonTime/Memory/Threads`를 사용. 모든 `.tene.enc` 파일에 기록된 `header.KDFMemory/Iterations/Parallel` 필드는 읽히지만 Argon2에 전달되지 않음.
- **위험**: 향후 `ArgonMemory`를 128MiB로 올리는 순간, 기존 암호화된 export 들이 모두 복호화 불가능.
- **수정**: `DeriveKey`가 명시적 `(time, memory, threads)` 파라미터를 받도록 변경. `Decrypt`는 header 값 전달, `Encrypt`는 사용한 default 기록. 하위 호환 보장.
- **노력**: S (3-4시간).

### P0-3. `pkg/errors`가 stdlib `errors` 패키지를 그림자처리 — A4
- **근거**: `pkg/errors/errors.go:1`이 `package errors`로 선언. 모든 호출자가 alias 강제 (`teneerr "github.com/agent-kay-it/tene/pkg/errors"`).
- **위험**: 교과서적인 stutter 안티패턴; 모든 contributor에게 영구 마찰; `%w` 래핑 후 `*TeneError` 타입 어서션은 `errors.Is/As`와 함께 작동하지 않음.
- **수정**: `pkg/teneerr` (또는 `pkg/errs`) 로 rename; 모든 alias 제거. 저장소 전체 단일 rename + import 업데이트.
- **노력**: S (반나절).

### P0-4. macOS Touch ID / 생체인증 통합 부재 — A3 (사용자 명시 우선순위)
- **근거**: Master Key가 `go-keyring`의 legacy `SecKeychain*` API를 통해 keychain에 위치. Default ACL: 사용자로 실행되는 모든 프로세스가 프롬프트 없이 읽기 가능. `kSecAttrAccessControl=BiometryCurrentSet` 없음, Secure Enclave 래핑 없음.
- **접근법** (Big Bet): `internal/biometric/` 패키지에 `Provider` 인터페이스 도입; macOS adapter는 SE-resident P-256 key로 Master Key를 래핑 (Touch ID 필요). Windows adapter는 CNG TPM-backed key 사용. Linux adapter는 TPM2 (선호) 또는 fprintd (인증 게이트만).
- **핵심 조합**: `tene agent` 데몬 — 세션 TTL 캐싱이 없으면 모든 `tene run --` 가 Touch ID를 트리거하여 작업 흐름 파괴.
- **노력**: L (macOS 경로만 4-6주; Windows + Linux 추가 시 +6-8주).

### P0-5. `tene audit` 읽기 명령 부재 — A10
- **근거**: `internal/vault/vault.go:424`가 9가지 이벤트 타입을 SQLite `audit_log`에 기록하지만, 읽기용 CLI surface 가 없음. 운영자는 `sqlite3 .tene/vault.db`를 수동으로 실행해야 함.
- **위험**: "AI agent forensics" 스토리는 tene 포지셔닝의 핵심이지만 현재로선 무대 장식 수준.
- **수정**: `tene audit [--since 24h] [--actor ai|human] [--json]` 출시. 테이블 이미 존재; 1일짜리 기능.
- **노력**: S.

### P0-6. CI가 `ubuntu-latest`에서만 실행됨 — A8
- **근거**: `.github/workflows/ci.yml:11,21`에 matrix 없음. Windows + macOS 회귀가 릴리스까지 도달. 이미 `tene update`의 `.tar.gz`/`.zip` Windows 404 버그가 발생.
- **수정**: `go test ./...`에 `matrix.os: [ubuntu-latest, macos-latest, windows-latest]` 추가, 스모크 테스트 포함 (`init && set && run --`).
- **노력**: S.

---

## 3. 관점별 섹션

각 섹션은 해당 에이전트의 보고서를 그대로 (가벼운 포맷팅만 적용) 옮긴 것이다. file:line 근거 그대로 보존.

### A1 — Security & Cryptography (7.0 / 10)

**한 줄**: 기본 primitive (XChaCha20-Poly1305, Argon2id, BIP39, X25519) 는 정확히 배선됨. 다만 경계 이슈 (passwd verifier 부재, encfile header KDF params 무시, audit-log 누설) 가 7/10 에서 상승을 막음.

#### 강점
1. **AEAD 배선이 교과서적** — `pkg/crypto/encrypt.go:21-32`이 `chacha20poly1305.NewX`를 사용하며 호출마다 CSPRNG로 24바이트 fresh nonce 생성; key name을 AAD로 바인딩 (`set.go:114`, `get.go:71`).
2. **표면 간 AAD-domain separation** — `recovery/recover.go:41`은 `AAD="recovery"`, `encfile/encfile.go:117`은 `AAD="tene-export"`, `teamkey.go:53`은 `AAD=recipientUserID` 사용. 한 표면의 blob이 다른 표면에서 복호화 불가.
3. **컴파일러 최적화 우회 zero-byte 처리로 메모리 위생 방어** — `pkg/crypto/zero.go:5-20`이 `//go:noinline keepAlive`로 dead-store elimination 방지. 26+ `ZeroBytes` defer 사이트. Argon2id params (t=3, m=64MiB, p=1) 이 OWASP 2024 초과.

#### 개선점
- **P0** `tene passwd`가 old password 검증 없이 재암호화 (`passwd.go:30-34`) → `vault_meta`에 `auth_hash` 추가. S (2-3h).
- **P0** encfile header KDF params 장식용 (`encfile.go:163`) → `DeriveKey`가 명시적 params 받도록 변경. S (3-4h).
- **P1** `TENE_MASTER_PASSWORD` 환경변수가 child process 에서 무한 lifetime (`run.go:104-105`) → child exec 전 `os.Unsetenv`. S (1h).
- **P1** `subtle.ConstantTimeCompare` 부재 — 테스트 파일 외 사용 0건. 향후 auth_hash 체크에 필수. S.
- **P2** Audit log + 에러 메시지가 secret name 누설 (`vault.go:425-433`, `run.go:75`) → audit에서 name hash-truncate, `--verbose` 외에는 generic error code. M.
- **P2** HKDF salt가 16+ 호출 사이트에서 nil (`pkg/crypto/keymanager.go:6-8`); `vault_meta.kdf_salt` 존재하지만 HKDF용으로 미사용 → 연결. M.
- **P2** File-fallback keyfile이 디스크에서 base64-only (`keychain/fallback.go:21-30`). Linux headless / CI / Docker 환경에서 32바이트 master key가 사실상 평문 (mode 0600 만 적용) → machine-bound secret으로 암호화하거나 명시적 `--no-keychain` 없이는 fallback 거부. M.
- **P3** Recovery salt가 고정 상수 문자열 (`recovery/recover.go:10-19`) → vault별 `recovery_salt`. S.

#### Quick wins
- `tene passwd`가 old password를 실제로 검증하도록 수정 (P0).
- child spawn 전 `os.Unsetenv("TENE_MASTER_PASSWORD")` (`run.go:104`).
- `encfile.Decrypt`가 `header.KDFMemory/Iterations/Parallel`을 존중하도록 변경.

#### Big bets
- **Vault format v2** — auth_hash + HKDF salt + machine-bound keyfile encryption 을 `tene migrate`로 묶기. 9/10 감사 점수 획득; 향후 Argon2 파라미터 인상 가능.
- **사이드채널 + key-lifetime 감사**: `valgrind --tool=memcheck`, `go test -race -gcflags=-m`, master-key 버퍼에 `golang.org/x/sys/unix.Mlock`.

---

### A2 — Clean Architecture & SOLID (5.0 / 10)

**한 줄**: `pkg/domain`은 비교적 순수, `keychain.KeyStore`는 깨끗한 port. 그 외는 모두 `internal/cli` god-package에서 인프라 직접 결합.

#### 강점
1. **`pkg/domain`이 대체로 순수** — `vault.go`, `team.go`, `user.go`, `errors.go`, `vault_key_metadata.go` 모두 무의존 데이터 struct.
2. **`keychain.KeyStore`가 실제 port** (`internal/keychain/keychain.go:20-33`) — `KeyringStore` + `FileStore` adapter + 팩토리.
3. **구조화된 error type** (`pkg/errors/errors.go`) — `TeneError{Code, Message, Exit}`이 Cobra와 분리; exit code 변환이 `cmd/tene/main.go:60-79` composition root에 위치. 교과서적 패턴.

#### 개선점
- **P0** use-case / application 레이어 부재. `init.go:50-236`는 187줄 `RunE`. `team.go:105-165`는 HTTP / ECDH / master key derivation / 파일 쓰기 / JSON 포맷팅을 한 함수에 섞음. → `internal/usecase/` 도입, CLI 파일은 ~30줄로 축소: parse → use case 호출 → render. L (2-3주).
- **P0** `internal/cli`가 `pkg/crypto`를 직접 import (`set.go:10`, `get.go:8`, `run.go:10`, `import_cmd.go:10`, `team.go:15`, `init.go:13`) → `SecretCipher` 인터페이스 정의, composition root에서 adapter 주입. M.
- **P1** `internal/vault.Vault`가 concrete struct (interface 아님), mock 불가. 30+ 호출 사이트가 `app.Vault.X` 사용 — 테스트마다 실제 SQLite 파일 작성. → `SecretStore`, `EnvironmentStore`, `MetaStore`, `AuditLogger` 인터페이스 분리 (ISP). M.
- **P1** `internal/sync.Engine`이 여러 책임을 흡수 + 경계 침범 (`engine.go:81-119`이 자체 `vault.New` 오픈). → `sync.Transport` + `sync.StateStore`로 분리; `HTTPClient` 주입. M.
- **P1** 패키지 레벨 mutable 플래그 변수 (`root.go:41-48`). `run.go:125-141`이 RunE 안에서 이들을 write. 테스트마다 글로벌 reset 필요. → `App` struct 또는 `cobra.Command.SetContext` 로 이동. S-M.
- **P2** Cloud HTTP 코드가 `team.go`, `login.go`, `push.go`, `pull.go` 에 중복. → `internal/cloud/client.go`에 typed method 단일화. M.
- **P2** Domain 오염: `domain.Vault.S3Key`, `User.LemonCustomerID`, `Team.LemonSubscriptionID`가 domain을 AWS/Lemon에 결합. → rename + `domain.BillingProfile` 로 이동. S.
- **P2** Audit logging이 fire-and-forget `_ = app.Vault.AddAuditLog(...)` 로 여러 handler에 산재. → use-case 레이어의 `AuditLogger` 인터페이스. S-M.

#### Quick wins
- `sync.Engine` constructor에 `*http.Client` 주입 (1h).
- `vault.Store` 인터페이스 추가 (인터페이스 자체는 반나절; 리라이트는 점진적).
- `domain.Vault.S3Key` → `StorageKey` rename (1 commit).
- 사용되지 않는 `internal/cli/cloud_disabled.go::wrapCloudCmd` 삭제.

#### Big bets
- **Hexagonal 재구조화**: `internal/{usecase, port, adapter}/` 와 `cmd/tene/main.go` 가 composition root. 테스트 고통의 ~70% 제거; cloud 재활성화 (`root.go:99-109`) 가 1-PR job 으로 변함.
- **단일 `tene.App` constructor** 가 모든 port 소유; cobra command 는 얇은 presenter.
- **sync + cloud HTTP 를 `internal/adapter/cloud/` 하위로 이동** — local-first core에 transitive HTTP 의존성 제거 → `tene-lite` 빌드 태그 가능.

---

### A3 — Platform-Native Auth (Touch ID, Windows Hello, Linux 지문) (3.0 / 10)

**한 줄**: tene 는 OS keychain 을 raw key cache 로 쓸 뿐, 생체인증 trust anchor 로 쓰지 않음. 기능상 1Password ~2014년 수준.

#### 강점
1. **크로스 플랫폼 keychain 추상화 이미 존재** — `internal/keychain/keychain.go`가 `KeyStore`를 정의; `NewStore()`가 keyring 프로빙 실패 시 `FileStore`로 fallback. 생체인증 구현 추가는 인터페이스 삽입 작업이지 수술이 아님.
2. **Recovery key 경로가 keychain 으로부터 독립적** — `internal/recovery/recover.go`가 BIP39 mnemonic 에서 Argon2id로 recovery key 도출 후 master key 를 `recovery_blob` 으로 재암호화. 생체인증을 crypto 변경 없이 세 번째 unlock 경로로 추가 가능.
3. **프로젝트 단위 service name** — `hashPath(projectPath)` (`keychain.go:75`) 가 keychain 항목을 프로젝트별로 키잉; Touch ID ACL 과 올바르게 조합.

#### 개선점

- **P0 — macOS Touch ID via Secure Enclave 래핑 키 (centerpiece)**

  **아키텍처**:
  ```
  Password (init)
    → Argon2id KDF (기존 kdf.go)
    → MasterKey (32B)
    → SE-resident P-256 key 로 래핑
      (kSecAttrTokenIDSecureEnclave +
       kSecAttrAccessControl=BiometryCurrentSet|UserPresence)
    → 래핑된 blob 을 keychain 에 "tene-mk-wrapped-{projectHash}" 로 저장
    → Load() 시: SE 가 LocalAuthentication framework 에 요청 → Touch ID 프롬프트
       → unwrap → MasterKey 반환
  ```

  `github.com/zalando/go-keyring`은 이 작업 불가 (legacy `SecKeychain*` API). 실용적 바인딩 두 가지:
  - **`github.com/keybase/go-keychain`** (성숙, `AccessControl` flag 지원). **권장**.
  - **CGO 직접** Security.framework + LocalAuthentication.framework (~150 LOC). 최대 통제권.

  `go-touchid` 회피 (2018년 마지막 커밋; `LAPolicy` evaluation 만, 키 wrap 안 함).

- **P0 — `internal/biometric/` 패키지 추가, `internal/keychain/` 에 볼트온하지 말 것**

  Keychain (cache) 와 biometric (presence gate) 은 의미가 다름. 제안 레이아웃:
  ```
  internal/biometric/
    biometric.go        // Provider 인터페이스, capability detection
    darwin_touchid.go   // +build darwin (cgo, Security + LocalAuthentication)
    windows_hello.go    // +build windows (cgo / golang.org/x/sys/windows for CNG)
    linux_fprintd.go    // +build linux (dbus to net.reactivated.Fprint)
    noop.go             // +build !darwin,!windows,!linux
  ```

  `Provider` 인터페이스:
  ```go
  type Provider interface {
      Available() bool                                // hw + enrolled?
      Enroll(masterKey []byte, reason string) error  // wraps + stores
      Unlock(reason string) ([]byte, error)          // prompts + returns
      Disable() error
      Status() Status                                // enabled, enrolledAt, lastUsed, hwKind
  }
  ```

  `loadOrPromptMasterKey` (`root.go:173`) 가 chain 으로 변모: env var → biometric → keychain → password → recovery.

- **P1 — Windows Hello via CNG TPM-backed key 래핑** (`NCryptCreatePersistedKey` 에 `NCRYPT_USE_VIRTUAL_ISOLATION_FLAG` + `NCRYPT_REQUIRE_HARDWARE_FLAG`). SE 와 동등한 assurance.

- **P1 — Linux: 무엇이 가능한지 정직하게**
  1. TPM2 via `github.com/google/go-tpm-tools` — PCR 에 bind 된 sealed object 로 master key wrap. 최고 보안, 지문 UX 없음. P2.
  2. `fprintd` over D-Bus — 지문 presence 만, 키 wrap 안 됨.
  3. `polkit` re-auth — password 프롬프트; 저렴한 fallback.
  4. libsecret without biometric — 현재 Linux 베이스라인.

- **P1 — `tene biometric` 서브명령 surface**
  ```
  tene biometric status              # capability + enrolled + last-used
  tene biometric enable              # password 프롬프트, key wrap, enroll
  tene biometric disable             # unwrap, keychain-only fallback
  tene biometric test                # 부작용 없이 프롬프트 강제
  ```
  CI/SSH 용 hidden `--biometric=skip` 글로벌 플래그. `loadOrPromptMasterKey` 에서 `SSH_CONNECTION` / `!isTerminal()` 감지하여 auto-skip.

- **P2 — SE/TPM 통한 세션 TTL, in-process cache 가 아님.** 현재 master key 가 keychain 에 영구 거주. 더 안전한 모델: 생체인증으로 임시 버퍼에 unwrap; `--session 5m` 가 `tene agent` 데몬 (ssh-agent 처럼) 유지.

- **P2 — `kSecAttrAccessControl=BiometryCurrentSet` 사용, `BiometryAny` 가 아님.** 사용자가 새 지문을 enroll 하면 SE 키 무효화. Signal, Bitwarden 과 일치. 문서화: 재 enroll → password 로 한 번 unlock → biometric 재활성화.

- **P3 — 2026 생태계: passkeys / FIDO2 / WebAuthn-native** `tene login` 클라우드 sign-in 용. Vault unwrap 과 cloud session 은 별개 관심사 — 혼동 금지.

#### Quick wins
- `Available()` 프로빙 + `tene biometric status` 를 wrap 로직 도입 전에 먼저 출시 (활성화 가능 사용자 telemetry 확보).
- macOS Master Key entry 만 `keybase/go-keychain` 로 전환 (cloud token 용 zalando 유지).
- `tene biometric enable` 출력에 recovery key 불변성 문서화: "Biometric does NOT replace your recovery key."

#### Big bets
- **`tene agent` 데몬** — SE/TPM-bound 세션, TTL 만료 시 생체인증 재인증. tene 를 "1Password CLI of vibe-coding" 로 변모 — 작업 세션당 Touch ID 한 번 후 모든 `tene run --` 가 즉시 실행. **현재 가용한 가장 높은 레버리지의 UX 투자.**
- **크로스 플랫폼 attestation 내보내기**: `tene biometric attest` 가 SE/TPM attestation 출력; tene-cloud 가 `push` 시 기록. `tene pull` 을 새 디바이스에서 할 때 recovery key OR attested 두 번째 디바이스의 join 승인 필요 — 하드웨어 증명 기반 passwordless team onboarding. Doppler/Infisical 과 차별화.
- **위협 모델 delta, 명시적**: 생체인증은 coercion resistance (silent keychain 읽기 불가) + post-theft protection 추가. 다음을 대체하지 않음: recovery key, master password (cross-device portability), vault crypto.

---

### A4 — Go Idioms & Convention (6.5 / 10)

**한 줄**: 견고한 구조적 선택이 얇은 lint config, 패키지 레벨 state, 일관성 없는 custom-error 사용 때문에 손상됨.

#### 강점
1. **Sync engine 의 error wrapping 규율** — `internal/sync/engine.go` 가 `fmt.Errorf("sync push: read vault: %w", err)` 같이 안정된 prefix 로 일관성 있게 wrap (line 126, 133, 142, 158, 178, 273-288).
2. **Leaf 패키지의 sentinel error 위생** — `pkg/crypto/errors.go:5-17`, `internal/encfile/encfile.go:28-30`, `internal/recovery/errors.go:6-7`. `errors.Is` 친화.
3. **방어적 close 관용구가 균일** — `defer func() { _ = v.Close() }()` 가 트리 전체에 51회 반복. errcheck 만족 의도적.

#### 개선점
- **P0** Linter 커버리지 너무 좁음. `.golangci.yml:8-14`가 6개 linter 만 활성화. 누락: **gofumpt, revive, gocritic, gosec, errorlint, bodyclose, noctx, wrapcheck, gocyclo, prealloc**. crypto + HTTP 다루는 Go 1.25 CLI 에 이건 단일 최대 갭.
- **P0** `pkg/errors` 가 stdlib `errors` 그림자처리. `pkg/errors/errors.go:1` → 모든 호출자가 alias 강제 (`teneerr "github.com/agent-kay-it/tene/pkg/errors"`). `pkg/teneerr` 또는 `pkg/errs` 로 rename.
- **P1** `TeneError` 가 `errors.As` 가 아닌 타입 어서션 사용. `pkg/errors/errors.go:54-59`의 `IsTeneError` 가 raw `err.(*TeneError)` 사용 — `%w` 래핑 후 silent 실패.
- **P1** `internal/cli` 의 패키지 레벨 mutable state (`root.go:16-20, 41-48`, `get.go:12`, `set.go:14-17`, `init.go:34-40`) + 명령어 등록하는 9개 분리된 `init()` 함수. 테스트 표면이 사실상 reset 불가. **전체에 `t.Parallel()` 0개.** `newRootCmd()` 팩토리 + struct 에 바인딩된 flag 로 이동.
- **P1** `context.Context` 가 vault/crypto 레이어로 전파되지 않음. `vault.Vault.SetMeta/GetSecret/ListSecrets/DeleteSecret` 모두 `ctx` first arg 없이 SQLite I/O. CLI 트리 전체에 `context.Background()` 가 단 1회 등장.
- **P2** `pkg/errors/codes.go`의 error 구성 비일관 — `var Err... = &TeneError{...}` 글로벌과 constructor func 가 섞임.
- **P2** `internal/cli/*.go` 에 `%w` 없는 `fmt.Errorf("...")` 22개. `errors.Is` 깨뜨림.
- **P2** 패키지 doc comment 누락. `internal/sync/envelope.go:1` 와 `pkg/domain/errors.go:1` 만 존재.
- **P3** `.golangci.yml:16`의 `disable: [unused]` 가 죽은 cloud 명령용 밴드에이드. `//go:build cloud` 태그로 이동.

#### Quick wins
- `pkg/errors` → `pkg/teneerr` rename; 모든 alias 제거.
- `.golangci.yml` 에 `errorlint`, `bodyclose`, `gosec`, `gofumpt`, `revive` 추가.
- `// Package ...` doc comment 추가 (각 1줄).
- `IsTeneError` 구현을 `errors.As` 로 교체; `func (*TeneError) Unwrap() error` 추가.
- bare `fmt.Errorf` 22개를 `%w` 래핑으로 전환.
- `engine.go:308`: `fmt.Sprintf("%d", state.Version)` 를 `strconv.FormatInt` 로 교체.

#### Big bets
- **`internal/cli` 를 `App` + 팩토리 func 중심으로 재형성**, 9개 `init()` 와 7개 `var (...)` flag 블록 제거. `t.Parallel()`, 테이블 테스트, 실제 통합 테스트 하니스 unlock.
- **`context.Context` 를 end-to-end 로 thread** (`Vault`, sync, keychain) — SIGINT 취소, `*sql.DB.ExecContext`, HTTP 데드라인에 필수.

---

### A5 — CLI UX & DX (7.5 / 10)

**한 줄**: 강력한 헤드라인 기능 (`tene run --`) 과 현대적 안전 기본값을 갖춘 세련된 CLI. discoverability 와 일부 명령어 형태 비일관성 때문에 best-in-class 에 못 미침.

#### 강점
1. **`tene run --` 디자인이 안전 스토리를 정확히 구현** — `run.go:14-30` 깔끔한 `--` passthrough, 자식 종료코드 전파 (`run.go:111-112`), child stdout 순수성을 위한 stderr 전용 diagnostics, stderr 의 `--json` info (`run.go:83-90`).
2. **AI-leak-aware 기본값** — `tene get` 이 non-TTY stdout 거부 (`get.go:100-102`), 명확한 remediation 경로. dual flag + env override (`TENE_ALLOW_STDOUT_SECRETS`) 가 진정 새로움.
3. **구조화된 error 봉투 + exit code taxonomy** — `TeneError{Code, Message, Exit}` (`pkg/errors/errors.go:13-17`); error 함수에 next-action 힌트 포함. exit code 테이블 `docs/cli-reference.md:19-33` 에 공표.

#### 개선점
- **P0** Exit code 가 문서와 코드 사이에서 drift. `docs/cli-reference.md` 가 `3=VAULT_NOT_FOUND`, `4=AUTH_REQUIRED`, `5=SECRET_NOT_FOUND`, `6=DECRYPT_FAILED`, `7=INTERACTIVE_REQUIRED` 광고. 구현은 모두 `Exit: 1` 할당. 스크립트 깨짐.
- **P0** `tene init` next-step pointer 가 절반만 맞음. `init.go:221` 가 `"Next: tene set KEY VALUE"` 출력하지만 `tene run --` (헤드라인 기능) 또는 `tene list` 언급 없음. 첫 사용자가 full happy path 를 놓침.
- **P1** `tene run` flag parsing 이 fragile. `run.go:29` 가 `DisableFlagParsing: true`; `parseFlagsBeforeDash` 가 `--env`/`-e`/`--json`/`--quiet` 만 처리. 다른 플래그 (`--dir`, `--no-color`, `--no-keychain`) 가 `--` 앞에 놓이면 silent no-op. `run.go:143-163`의 fallback 이 `--` 없을 때 모든 args 를 명령으로 관대히 취급 — `op run` 와 `doppler run` 은 `--` 강제.
- **P1** `tene env` 서브명령이 verb-noun 컨벤션과 충돌. `env.go:12-16` 이 `tene env <name>` (switch), `tene env list`, `tene env create NAME` 모두 수락. 등록 순서로만 구분. 대신 `tene env use <name>` / `tene env ls` 사용.
- **P1** Help text 품질 불균일. `init`/`set`/`get`/`run` 에 `Example:` 블록; `list.go:11-15`, `env.go:12-16`, `delete`, `whoami` 는 한 줄 `Short` 만. `tene list --help` 와 `tene env --help` 가 가장 빈번하면서 가장 부실.
- **P2** `--env` vs `TENE_*` env var 비일관. 셸 세션별 default environment 를 설정하는 `TENE_ENV` 없음.
- **P2** `tene export` foot-gun 이 문서화되었지만 게이트되지 않음. `cli-reference.md:138-140` 가 경고; 명령에는 `tene get` 의 `--unsafe-stdout` 미러 보호 없음.
- **P3** Cloud 서브명령이 비활성화 되어 있지만 참조됨. `root.go:99-109` 가 login/push/pull/sync/team/billing 주석처리; 남은 파일들이 서로 참조 (`team.go:112` "Run 'tene login' first"). 재활성화 전 audit.

#### Quick wins
- `pkg/errors/codes.go` exit code 를 `docs/cli-reference.md:19-33` 와 정렬 (30분).
- `init.go:221` next-step footer 에 `tene run --`, `tene list` 추가 (5분).
- `list/env/delete/whoami/passwd/recover` 에 `Example:` 블록 추가 (30분).
- `tene run` 의 persistent flag 한계 `Long:` 에 문서화 (5분).
- `resolveEnv` (`root.go:127-135`) 에 `TENE_ENV` env var 지원 추가 (10분).
- `tene get` 의 non-TTY 가드를 `tene export` 에도 적용.

#### Big bets
- **`tene env <name>` → `tene env use <name>` rename**, deprecation alias 동반. `tene env ls` 페어. `kubectl config use-context`, `gh repo set-default` 매칭.
- **`tene status` 채택** — `gh auth status` 모델: vault 경로, 활성 env, secret 수, keychain 상태, recovery key 등록 (boolean), update 가용성. "내 설정 건강한가?" 표준 entry point.
- **참조 문서를 코드에서 생성** Cobra `doc.GenMarkdownTree` 사용. `RootCmd()` 가 진실의 원천.

---

### A6 — Test Strategy (5.5 / 10)

**한 줄**: crypto + vault + sync primitive 의 단위 테스트는 견고; 카테고리 통째로 빠짐 (e2e, fuzz, race-on-shared-state, KAT) + CLI 명령 surface 의 ~50% 미커버.

#### 강점
1. **Crypto 패키지 커버리지가 사려깊음** — `pkg/crypto/crypto_test.go` 가 wrong-key, wrong-AAD, invalid-ciphertext, invalid-key-length, empty-plaintext 운동. 부정 경로 체계적.
2. **Three-way merge 가 잘 운동됨** — `internal/sync/merge_test.go` 가 10개 충돌 시나리오 열거; `envelope_test.go` 가 seal/open 라운드트립, truncation, invalid magic, 큰 페이로드 커버.
3. **CLI 통합 하니스 존재** — `internal/cli/testhelper_test.go` 가 `setupTestEnv` + `run()` + `resetFlags()` 제공. CI 가 `go test -race -coverprofile=coverage.out` 실행.

#### 개선점
- **P0** Fuzz 테스트 0개. `pkg/crypto.Decrypt` (잘못된 ciphertext), `internal/sync/envelope.go::Open`, `internal/encfile/encfile.go::DecodeHeader`, `internal/cli/set.go` 의 secret name 검증기 — 모두 prime fuzz target.
- **P0** `internal/sync/engine.go` (472 LOC) 에 직접 테스트 없음. `merge.go`/`envelope.go`/`conflict.go` 만 테스트됨. `httptest.Server` fake 0건.
- **P0** Recovery key 경로가 얕음. `internal/cli/recover.go` (170 LOC) 테스트 0개. 미테스트: 손상된 vault + 유효 recovery, 잘못된 key + 유효 vault, `passwd` 회전 후 recovery. `passwd.go` (167 LOC) 도 동일.
- **P1** XChaCha20-Poly1305 KAT 없음. 모든 crypto 테스트가 round-trip property-style. RFC 8439 / libsodium 벡터 3-5개를 hex-pinned in/out 으로 고정.
- **P1** CLI 명령 ~50% 미테스트. 미테스트 2,300+ LOC: `run.go`, `passwd.go`, `recover.go`, `team.go` (464!), `push.go`, `pull.go`, `login.go`, `update.go`, `sync_cmd.go`, `logout.go`, `completion.go`, `helpers.go`.
- **P1** `t.Parallel()` 전체에 0개. CLI 테스트가 `resetFlags()` 가 패키지 글로벌을 mutate 하기 때문에 병렬화 불가. Race detector 가 우연히만 통과.
- **P2** `testdata/` golden file 없음 (`find . -type d -name testdata` 결과 없음). 인라인 문자열 assert 의 취약성.
- **P2** CLI binary spawn / 진짜 e2e 없음. 기존 하니스가 `rootCmd.Execute()` 를 in-process 호출 — 시그널 핸들링, exit code, 실제 stdin TTY, 실제 keychain (macOS Security 호출이 fake 로), `tene run -- child` subprocess 상속 모두 미테스트.

#### Quick wins
- 하나의 타겟 (`pkg/crypto.FuzzDecrypt`) 에 대해 CI 에 `go test -fuzz=Fuzz -fuzztime=30s` 추가.
- `vault_test.go` 에 `TestVaultCorrupted` (파일 중간 바이트 플립).
- CLI 글로벌 안 건드리는 80+ 테스트에 `t.Parallel()` — wall-clock 3-4× 빨라짐.
- `pkg/crypto/crypto_test.go` 에 RFC 8439 KAT 1개 고정.

#### Big bets
- **빌드된 binary 로 e2e 하니스** — `tests/e2e/` 에서 `go build -o bin/tene && bin/tene init/set/run/export/import/passwd/recover` 를 tmp HOME 에서 실행. `//go:build e2e` 태그로 단위 테스트 속도 유지.
- **CI 의 커버리지 게이트** — `pkg/crypto` 70%, `internal/sync` 60%, `internal/cli` 50% 고정. 떨어뜨리는 PR 거부.
- **Merge 의 property-based 테스트** — `gopter` 생성기로 (base, local, remote) 트리플 생성, 가환성 + 멱등성 invariant assert.

---

### A7 — Performance & Scalability (6.5 / 10)

**한 줄**: 100-secret/단일 사용자 vault 에 견고; 보이지 않는 비용 누적 (keychain 프로빙, sqlite migrate, untuned PRAGMA, `tene run` 의 full `GetAllSecrets`). 벤치마크 커버리지 0개.

**측정값**: `tene version` = 9ms, `tene env list` = 43ms (macOS Darwin 24.6). 현재 파라미터 Argon2id = 98ms; p=4 = 28ms (3× 빠름, 보안 동일).

#### 강점
1. **WAL journal + 다중 프로세스 안전** (`vault.go:35`). 동일 vault 에 대한 다중 `tene run` 안전.
2. **Batch write 가 단일 트랜잭션 + prepared statement 사용** (`vault.go:262-292`, `SetSecretBatch`). 100키 `tene import` 가 한 번의 fsync.
3. **Hot path 가 keychain hit 시 Argon2 우회** (`root.go:173-178`). KDF 는 첫 unlock 시에만 지불.

#### 개선점
| # | 이슈 | P |
|---|---|:---:|
| 1 | `keychain.NewStore` 가 호출마다 `Set` + `Delete` 프로빙 (`keychain.go:91-97`). macOS 에서 = 명령마다 `securityd` IPC 왕복. CLI 호출마다 5-15ms 낭비. | P0 |
| 2 | 벤치마크 커버리지 전혀 없음. KDF 튜닝, vault 읽기 스케일링, run injection 모두 무방비. | P0 |
| 3 | SQLite PRAGMA 미튜닝: `synchronous`, `busy_timeout`, `cache_size`, `mmap_size` 모두 미설정. 기본 `synchronous=FULL` = write commit 당 두 번의 fsync. | P1 |
| 4 | `tene run` 이 모든 secret 을 순차 복호화 (`run.go:69-80`). 1k secret 에서: 단일 고루틴에서 1000× XChaCha20-Poly1305. 엔터프라이즈 vault 캡 발생. | P1 |
| 5 | `sql.DB` pool 설정 없음 (`SetMaxOpenConns`/`Idle`). modernc.org/sqlite 가 pure-Go, 단일 CGo-free shim 으로 직렬화. | P2 |
| 6 | modernc.org/sqlite 가 write-heavy 시 mattn-sqlite3 보다 ~3-5× 느림. 10k-secret 엔터프라이즈 덤프 `tene import` 가 멀티초 영역으로 캡. 단일 binary 배포 단순성용 CGo-free — 문서화되지 않은 트레이드오프. | P2 |
| 7 | Read hot path 의 prepared statement 재사용 없음. `GetAllSecrets`, `GetSecret`, `SecretExists` 가 호출마다 SQL 재파싱. | P3 |
| 8 | `Vault.New` 가 매 open 마다 `migrate()` 실행 (`vault.go:48`). 호출당 쿼리. | P3 |

#### Quick wins (~30-50% startup 누적 절감)
- `NewStore` 의 keychain 프로빙 제거 — 실패는 실제 `Load()` 호출로 미룸.
- `vault.New` 에 `PRAGMA synchronous=NORMAL` + `PRAGMA busy_timeout=5000`.
- `schema_version` meta 가 존재할 때 migrate 건너뛰기.
- `BenchmarkUnlock`, `BenchmarkListSecrets/N=1000`, `BenchmarkRunInject/N=1000` 추가.
- `ArgonThreads = runtime.NumCPU()` (cap 4) — 측정 98ms → 28ms.

#### Big bets
- **`tene run` 의 selective decryption** — `--inject=KEY1,KEY2 -- ...` 또는 프로젝트별 `tene.toml` allowlist. 10k-secret 엔터프라이즈 + least-privilege injection 스토리에 필수.
- **`mlock`/`memguard` 로 메모리 보호된 secret 페이지.** ZeroBytes 가 best-effort; Go GC 가 heap 페이지 간 복사 가능. `MasterKey` + 복호화된 plaintext slice 에 `memguard`. binary ~50KB 추가; swap 저항 + core-dump 저항.
- **modernc.org/sqlite 를 `crawshaw.io/sqlite` 또는 `zombiezen.com/go/sqlite` 로 교체** — 둘 다 CGo-free + 빠름, 명시적 per-connection 소유권.

---

### A8 — Cross-Platform Parity (5.5 / 10)

**한 줄**: 합리적인 Go 선택 (`filepath.Join`, `go-keyring`, `golang.org/x/term`), 그러나 CI 가 macOS 나 Windows 에서 실행되지 않으며 vault 경로가 모든 OS 에서 POSIX-shape. parity 가 이론적.

#### 강점
1. **경로 처리 모두 `filepath.Join`** (`init.go:60`, `engine.go:214`, `keychain.go:84`).
2. **자동 fallback keychain 추상화** (`NewStore` 가 프로빙; `~/.tene/keyfile` 0600/0700 으로 fallback). `TENE_KEYCHAIN_FALLBACK=file` 로 override.
3. **크로스 OS 브라우저 dispatch** (`login.go:146-152`): GOOS 별 `open` / `xdg-open` / `rundll32 url.dll,FileProtocolHandler`.

#### 개선점
- **P0** CI 가 `ubuntu-latest` 에서만 실행 (`.github/workflows/ci.yml:11,21`). `matrix.os: [ubuntu-latest, macos-latest, windows-latest]` 추가.
- **P0** Vault 경로가 모든 OS 에서 `~/.tene` 로 하드코딩 (`config.go:68-72`, `init.go:95`, `login.go:288`, `keychain.go:84,95`). `os.UserConfigDir()` 로 전환 (Windows %AppData%, Linux $XDG_CONFIG_HOME, macOS ~/Library/Application Support). legacy 경로 read fallback 동반.
- **P0** `.env` import 가 CRLF 라인을 silent 손상. `bufio.Scanner` 가 `\n` 만 스트립; `TrimSpace` 가 대부분 경로에서 구하지만 `trimQuotes` 가 일부 edge case 에서 `\r` 정리 전에 실행. `scanner.Text()` 직후 명시적 `strings.TrimRight(line, "\r\n")` 추가 (`import_cmd.go:74-92`).
- **P1** 파일 권한이 POSIX 전용. `os.MkdirAll(..., 0700)` 와 `os.WriteFile(..., 0600)` (`init.go:96`, `fallback.go:23-29`, `vaultjson.go:35`) **Windows 에서 silent 무시** — vault + keyfile 이 기본 상속 ACL 받음. `golang.org/x/sys/windows` 사용해 restrictive DACL.
- **P1** `tene run --` 의 셸 의미 차이. `exec.Command` 가 셸 없음. `tene run -- 'npm start && npm test'` 실패. 명시적 문서화 또는 Windows 감지하여 `cmd /c` dispatch (위험).
- **P1** password prompt 중 Ctrl-C 의 시그널 정리 없음. `term.ReadPassword` 가 터미널 상태 복원할 `signal.Notify` 없음. `defer term.Restore` + 시그널 핸들러로 wrap.
- **P2** `tene update` 의 PATH/permission 로직 Windows 취약. `tene_X_windows_amd64.tar.gz` 다운로드하지만 `.goreleaser.yml` 은 Windows 용 `.zip` 생성 → **Windows 에서 in-place self-update 가 404**.
- **P2** Shell completion 이 Linux 전용. `.goreleaser.yml` 이 bash/zsh/fish + man page 내보내지만 PowerShell 없음.
- **P3** 대소문자 무관 파일시스템 / WSL 하이브리드에서 경로 해시 충돌 위험.

#### Quick wins
- CI matrix 에 Windows + macOS 추가 (`.github/workflows/ci.yml`).
- `internal/cli/update.go:125` 의 `.tar.gz`/`.zip` 불일치 수정.
- `internal/cli/import_cmd.go:76` 에 `strings.TrimRight(line, "\r")` + CRLF fixture.
- `.goreleaser.yml` `before:` hook 에 PowerShell completion.
- 모든 `ReadPassword` 주변에 `signal.Notify` + `term.Restore`.

#### Big bets
- **`~/.tene` → `os.UserConfigDir()/tene` 마이그레이션** 일회성 silent 마이그레이션. tene 를 플랫폼 컨벤션에 맞춤, Windows 로밍 프로파일 해결.
- **Windows 네이티브 install 경로**: signed `.msi` 또는 scoop manifest 를 `.zip` 과 함께 출시. SmartScreen / Defender false positive 회피.
- **WSL 감지 + cred 라우팅**: `keychain.NewStore` 에 `runtime.GOOS == "linux" && os.Getenv("WSL_DISTRO_NAME") != ""` 분기.

---

### A9 — Distribution & Supply Chain (5.0 / 10)

**한 줄**: 견고한 GoReleaser 기반; 체인이 "checksum + GitHub auth" 에서 멈춤. signing/SBOM/SLSA provenance/notarization 없음, **활성 Homebrew tap 없음**. 가치 제안이 "비밀을 맡겨라" 인 도구에서 이 갭은 지배적.

#### 강점
1. **깨끗한 GoReleaser v2 설정** (`.goreleaser.yml:15-65`) — CGO=0, ldflags `version/commit/date`, OS 별 archive override, completion + manpage, darwin/linux/windows × amd64/arm64 매트릭스.
2. **Auto-tag staging→main 의 RC 채널** (`.github/workflows/auto-tag.yml:37-94`) — staging 이 `-rc*` 컷, main 이 RC 승격 또는 patch bump. "Verified" 배지를 위한 API 태그 생성. OIDC AWS role (정적 키 없음). 문서화된 post-mortem 의 `s3api head-object` 통한 스마트 `LATEST_VERSION`.
3. **Checksum 검증이 있는 self-update** (`internal/cli/update.go:137-141`) + 매칭 install.sh (`apps/web/public/install.sh:84-95`) — 둘 다 `checksums.txt` 에서 SHA-256 검증. install.sh 가 133 줄, 감사 가능, non-root fallback.

#### 개선점
| # | 이슈 | P |
|---|---|:---:|
| 1 | **코드 서명 전무.** macOS unsigned → Gatekeeper 벽 + 격리. Windows unsigned → SmartScreen 경고. artifact 나 GHCR 에 Sigstore/cosign 없음. | **P0** |
| 2 | **Homebrew tap 비활성** (`.goreleaser.yml:89-140` 주석처리; 누락된 `tomo-kay/homebrew-tene` repo + PAT 인용 TODO). README 가 curl-pipe + `go install` 만 광고. v1.0.5-1.0.7 실패한 3개 릴리스가 이것 때문. | **P0** |
| 3 | **SBOM 내보내기 없음.** GoReleaser 의 `sboms:` (syft → CycloneDX/SPDX) 지원이 사소함. #1 결합 시 다운스트림 소비자 (기업, GovCloud, distro) 가 수용 불가. | **P1** |
| 4 | **SLSA provenance 없음.** Workflow 가 이미 `id-token: write` 보유 (AWS OIDC). `slsa-framework/slsa-github-generator` 내보내기는 ~10줄. | **P1** |
| 5 | **재현 가능 빌드 위생 없음.** `mod_timestamp: '{{ .CommitTimestamp }}'` 없음. 동일 커밋에서 재실행이 다른 해시 생성. | **P1** |
| 6 | **CI action 들이 SHA pin 안 됨** (`actions/checkout@v4`, `actions/setup-go@v5`, `goreleaser-action@v7`, `docker/*@v3`). SHA pin + Dependabot 이 SLSA L3 베이스라인. | **P2** |
| 7 | **`packages/` scaffolding 있지만 npm wrapper 없음** (`packages/{cli,crypto,types}/` 가 존재하지만 비어있음). 출시 또는 `rm -r`. | **P2** |
| 8 | **install.sh 에 PGP/cosign 단계 없음** — 동일 S3 origin 의 SHA-256 만. S3 가 손상되면 체크섬도 손상. | **P2** |
| 9 | **third-party deps 가 있는데 `NOTICE` 파일 없음.** | **P3** |
| 10 | **Apple universal binary 없음.** GoReleaser 의 `universal_binaries:` 가 darwin/amd64 + arm64 융합. 작은 승리. | **P3** |

#### Quick wins
- 각 `builds:` 아래 `mod_timestamp: '{{ .CommitTimestamp }}'` (~1줄).
- artifact 당 syft CycloneDX + SPDX 내보내는 `sboms:` 블록 (~6줄).
- `checksums.txt` 에 연결된 SLSA 생성기 워크플로 (`slsa-framework/slsa-github-generator@v2.0.0`) (~30줄).
- 모든 `actions/*` 를 commit SHA pin + `github-actions` 용 Dependabot 활성화.
- Homebrew tap 재활성화 (`gh repo create tomo-kay/homebrew-tene --public`, `HOMEBREW_TAP_GITHUB_TOKEN` 설정, `.goreleaser.yml:109-140` 주석 해제). 블록 이미 작성되고 테스트됨.
- 비어있는 `packages/{cli,crypto,types}/` 비우거나 삭제.

#### Big bets
- **Sigstore-everywhere**: 모든 tarball, GHCR manifest, SBOM 에 cosign keyless 서명. ~20줄 YAML. 릴리스당 Rekor 투명성 로그 entry. **비밀 관리 CLI 의 단일 최대 신뢰도 향상.**
- **macOS notarization + Windows codesign**: Apple Developer ID ($99/년) + Azure Trusted Signing 또는 DigiCert ($200-400/년). 모든 HN/PH 설치에서 first-impression 전환을 비용 발생시키는 Gatekeeper/SmartScreen 마찰 영구 제거.
- **재현 빌드 검증 job**: 동일 커밋에서 parallel matrix 가 재빌드, `checksums.txt` 바이트 동등성 assert. "유지자를 신뢰" 를 "수학을 신뢰" 로 전환.

---

### A10 — AI Integration & MCP (8.0 / 10) — 최고 점수

**한 줄**: 2026 AI agent 생태계에 가장 잘 포지셔닝된 비밀 관리 도구. 두 구조적 구멍 (MCP server 없음, `tene audit` reader 없음) 이 9+ 도달 막음.

#### 강점
1. **`tene get` 이 적극적으로 non-TTY stdout 거부** (`internal/cli/get.go:100-102`, `ErrStdoutSecretBlocked`). CLI 문서 ("AI 가 절대 못 읽게") 가 런타임 동작과 일치하는 드문 케이스. `--unsafe-stdout` + `TENE_ALLOW_STDOUT_SECRETS=1` 이중 override 가 잘 설계됨.
2. **5-editor rule emission 이 카테고리 최고** (`internal/claudemd/template.go:89-95`): CLAUDE.md, `.cursor/rules/tene.mdc` (적절한 frontmatter), `.windsurfrules`, GEMINI.md, AGENTS.md 를 단일 템플릿에서 `GenerateAll()` / `GenerateSelected()` 통해. `generator.go:88` 의 `HasTeneSection` 으로 중복 없이 append. 멱등.
3. **`skills/tene-cli/SKILL.md` 가 예외적으로 잘 엔지니어링됨** (382 줄). frontmatter `description` 이 Claude 의 skill router 에 충분히 정밀; 안전 규칙이 명시적 "NEVER" 프레임으로 재번호; `--env` 가 `--` 앞 placement caveat 가 200줄에 문서화; installer 통합용 openclaw metadata.

#### 개선점
- **P0** `tene audit` 읽기 명령 없음. `vault.go:424` 가 9개 이벤트 타입 기록하지만 읽기 CLI 없음. 운영자가 수동으로 `sqlite3 .tene/vault.db`. **이게 없으면 audit 스토리는 theater.** `tene audit [--since 24h] [--actor ai|human] [--json]` 출시. 1일짜리 기능.
- **P0** MCP server 없음. 경쟁사 (Infisical Agent Sentinel, HashiCorp Vault MCP) 가 수렴 중. tene 의 "MCP 필요 없음" 은 오늘은 defensible 하지만 취약. `tene serve --mcp` 를 3개 도구로 출시: `list_secrets()` (이름만), `run_with_secrets(env, cmd)` (exit + stdout 반환), `inject_environment(env)` (*마스크된* env list 반환). **결정적으로: `get_secret` 노출 금지** — 안전 계약 유지.
- **P1** JSON output 에 `schemaVersion` 필드 없음. 깨지는 변경이 LLM tool wrapper 를 silent 하게 망가뜨림. 모든 `--json` 봉투에 `"schemaVersion": 1` 추가.
- **P1** Secret 값에 대한 prompt-injection 방어 없음. "Ignore previous instructions..." 가 든 값이 child process 로 주입됨. control 문자 + `\n` 포함 값 거부하는 `tene set --strict`; `tene list` 가 의심스러운 entry 플래그.
- **P1** 사용자 편집 시 CLAUDE.md drift 처리 안 됨. `HasTeneSection` 이 header 문자열만 체크. 사용자가 규칙 8 삭제 시 `tene init` 재실행이 못 고침. `SecretsMdTemplate` 와 현재 규칙 diff 하는 `tene doctor --fix-rules` 추가. 또한 `<!-- tene:v1 -->` content-hash sentinel.
- **P2** bkit 일급 hook 없음. PDCA checkpoint 가 Check phase 중 "이 AI 세션이 prod secret 건드렸는지" 확인용으로 `tene audit --since 1h --json` 호출 가능. `bkit:tene-audit` skill 또는 bkit 호환 JSON 내보내는 `tene pdca-check` 서브명령 — 독특한 차별화.
- **P2** Per-secret access ACL 부재. `tene run --` 실행 가능한 어떤 agent 도 모든 secret 받음. `tene run --only STRIPE_KEY,DB_URL` + frontmatter 선언 가능한 명령별 manifest.
- **P3** GEMINI.md 와 AGENTS.md 가 동일 콘텐츠 내보냄. agent 별 specialization (AGENTS.md imperative voice, GEMINI.md Jules-specific) 이 native 한 느낌.

#### Quick wins
- 모든 `--json` 봉투에 `"schemaVersion": 1` (1줄 × 8개 명령).
- `tene audit` 읽기 명령 출시 (테이블 존재).
- `SKILL.md:49` bullet: "사용자가 채팅에 secret 붙여넣으면 거부하고 말하기: `tene set KEY --stdin 을 직접 실행하세요; 저는 값을 보지 못합니다.`"
- 오래된 규칙 감지용 `template.go:5` 에 `<!-- tene-rules-version: 1 -->` 주석.
- `README.md:230` 에 `ErrStdoutSecretBlocked` non-zero exit 문서화.

#### Big bets
- **`tene serve --mcp` (Q3 2026).** 세 도구 surface; `get_secret` 명시적 생략. **마케팅 각도: "MCP server with no `get` — because your agent shouldn't have one either."** tene 를 첫 안전 by-construction MCP 비밀 관리 도구로 포지셔닝 (Infisical 의 `read_secret` 는 tene 가 닫은 누설 경로를 재도입).
- **`tene policy` + `tene audit` (Q4 2026)**: 선언적 명령별 secret manifest + actor 별 (`TENE_ACTOR_ID=claude-code` 통한 human/agent 신원) audit trail. 컴플라이언스급 "어떤 AI 세션이 어떤 secret 건드렸는지" forensics — 어떤 local-first 경쟁사도 제공 안 함. MIT core 깨지 않고 미래 유료 tier 정당화.

---

## 4. 우선순위 액션 레지스터

### P0 — 빠른 수정 필요 (릴리스 차단 또는 사용자 신뢰 핵심)

| # | 영역 | 항목 | 노력 | 출처 |
|---|---|---|:---:|---|
| 1 | Security | `tene passwd` 가 `auth_hash` 로 old password 검증 | S | A1 |
| 2 | Security | `encfile.Decrypt` 가 header KDF params 존중 | S | A1 |
| 3 | Go convention | `pkg/errors` → `pkg/teneerr` rename; alias 제거 | S | A4 |
| 4 | CLI UX | Exit code 를 `docs/cli-reference.md:19-33` 와 정렬 | S | A5 |
| 5 | AI integration | `tene audit` 읽기 명령 출시 | S | A10 |
| 6 | Cross-platform | CI matrix 에 macOS + Windows 추가 | S | A8 |
| 7 | Cross-platform | `tene update` Windows `.tar.gz`/`.zip` 불일치 수정 | S | A8 |
| 8 | Distribution | Homebrew tap 재활성화 (이미 작성됨) | S | A9 |
| 9 | Performance | `keychain.NewStore` 프로빙 제거 (호출당 5-15ms 절약) | S | A7 |
| 10 | Distribution | 코드 서명 — Apple notarization + Windows signing | M | A9 |
| 11 | Biometric | Secure Enclave 통한 macOS Touch ID (사용자 명시 우선순위) | L | A3 |
| 12 | AI integration | `tene serve --mcp` server | L | A10 |
| 13 | Architecture | `internal/usecase/` 레이어 도입 | L | A2 |
| 14 | Test | `pkg/crypto.Decrypt`, `sync/envelope.Open`, `encfile.DecodeHeader` 의 Fuzz | M | A6 |
| 15 | Test | `sync.Engine` (472 LOC, 테스트 없음) | M | A6 |

### P1 — 높은 우선순위 (가시적 품질 + 구조)

| 영역 | 항목 | 노력 |
|---|---|:---:|
| Security | child exec 전 `os.Unsetenv("TENE_MASTER_PASSWORD")` | S |
| Security | `subtle.ConstantTimeCompare` 채택 | S |
| Architecture | `pkg/crypto` 인터페이스 (`SecretCipher`) composition root 주입 | M |
| Architecture | `vault.Vault` 에서 `SecretStore`, `EnvironmentStore`, `MetaStore`, `AuditLogger` 인터페이스 추출 | M |
| Architecture | `internal/cli` 의 패키지 레벨 mutable state 제거 | S-M |
| Go convention | gofumpt/revive/gocritic/gosec/errorlint/bodyclose 를 lint 에 추가 | S |
| Go convention | Vault/sync/keychain 통해 `context.Context` thread | M |
| Go convention | `internal/cli/*.go` 의 bare `fmt.Errorf` 22개를 `%w` wrap | S |
| CLI UX | `tene env use <name>` rename; bare `tene env <name>` deprecate | S |
| CLI UX | `list/env/delete/whoami/passwd/recover` 에 `Example:` 블록 | S |
| Performance | `PRAGMA synchronous=NORMAL` + `busy_timeout=5000` | S |
| Performance | 3-5개 벤치마크 추가 (`BenchmarkUnlock` 등) | S |
| Performance | `ArgonThreads = runtime.NumCPU()` cap 4 — 측정 3× 빨라짐 | S |
| Cross-platform | `~/.tene` → `os.UserConfigDir()/tene` 마이그레이션 | M |
| Cross-platform | `import_cmd.go` 의 CRLF 처리 | S |
| Cross-platform | `golang.org/x/sys/windows` 통한 Windows ACL | M |
| Cross-platform | `ReadPassword` 주변 시그널 안전 terminal 복원 | S |
| Test | `pkg/crypto/crypto_test.go` 에 RFC 8439 KAT 고정 | S |
| Test | 80+ non-CLI 테스트에 `t.Parallel()` 추가 | S |
| Test | `recover.go`, `passwd.go`, `team.go` (464 LOC) 커버 | M |
| Distribution | `sboms:` 블록 통한 SBOM 내보내기 | S |
| Distribution | `slsa-github-generator` 통한 SLSA provenance | S |
| Distribution | 재현 가능 빌드 (`mod_timestamp`) | S |
| Distribution | Action 들을 commit SHA pin + Dependabot | S |
| Biometric | `tene biometric status/enable/disable/test` skeleton + `Available()` 프로빙 | M |
| Biometric | CNG 통한 Windows Hello | M-L |
| Biometric | Linux TPM2 sealed object | M-L |
| AI integration | 모든 `--json` 봉투에 `schemaVersion: 1` | S |
| AI integration | `tene set --strict` 가 control 문자 + `\n` 거부 | S |
| AI integration | CLAUDE.md drift 용 `tene doctor --fix-rules` | M |

### P2 — 중간 (다듬기 + 확장)

(요약 — 자세한 사항은 관점별 섹션 참조)

- File-fallback keyfile 을 machine-bound secret 으로 암호화
- 16+ 호출 사이트의 HKDF salt thread
- Cloud HTTP client 추출 (`internal/cloud/client.go`)
- Domain 오염 rename (`S3Key` → `StorageKey`)
- Audit logging cross-cutting concern 중앙화
- `TENE_ENV` env var 지원
- `tene export` non-TTY 가드
- SQLite pool 설정 (`SetMaxOpenConns`)
- modernc.org/sqlite 트레이드오프 문서화
- GoReleaser 의 PowerShell completion
- WSL keychain 라우팅
- install.sh cosign 검증 단계
- bkit 일급 hook
- Per-secret ACL / `--only` allowlist

### P3 — 다듬기

- Vault 별 `recovery_salt`
- 패키지 doc comment
- `disable: [unused]` 밴드에이드 → build 태그
- Apple universal binary
- `NOTICE` 파일 생성
- Agent 별 specialization (GEMINI vs AGENTS template)
- 사용되지 않는 `internal/cli/cloud_disabled.go::wrapCloudCmd` 이동

---

## 5. Quick wins 컴파일 (각 ≤ 1일, 1-2 sprint 으로 묶기 권장)

총 ~30-40 항목, ~5-8 dev-day, 평균 점수를 5.95 → ~7.0 으로 올림:

### Sprint A — security + correctness (먼저 권장)
- [ ] `tene passwd` 가 old password 검증 수정 (A1)
- [ ] child exec 전 `os.Unsetenv("TENE_MASTER_PASSWORD")` (A1)
- [ ] `encfile.Decrypt` 가 header KDF params 존중 (A1)
- [ ] `pkg/errors` → `pkg/teneerr` rename (A4)
- [ ] Exit code 를 문서와 정렬 (A5)
- [ ] `tene audit` 읽기 명령 출시 (A10)
- [ ] `tene set --strict` 검증 (A10)
- [ ] import 의 CRLF 처리 (A8)
- [ ] Windows `tene update` `.zip` 수정 (A8)

### Sprint B — distribution + supply chain
- [ ] Homebrew tap 재활성화 (블록 이미 작성됨)
- [ ] `.goreleaser.yml` 에 SBOM 블록 추가
- [ ] SLSA provenance workflow 추가
- [ ] 재현 가능 빌드용 `mod_timestamp` 추가
- [ ] 모든 GH action 을 SHA pin + Dependabot
- [ ] CI matrix: macOS + Windows
- [ ] `packages/{cli,crypto,types}/` 비우거나 삭제

### Sprint C — performance + DX 다듬기
- [ ] `keychain.NewStore` 프로빙 제거 (호출당 5-15ms 절약)
- [ ] `PRAGMA synchronous=NORMAL` + `busy_timeout`
- [ ] `ArgonThreads = runtime.NumCPU()` cap 4
- [ ] 3개 벤치마크 추가 (unlock, list, run-inject)
- [ ] non-CLI 테스트에 `t.Parallel()`
- [ ] Fuzz 테스트 하나 (`pkg/crypto.FuzzDecrypt`)
- [ ] RFC 8439 KAT 하나
- [ ] `init.go:221` next-step footer 확장
- [ ] `list/env/delete/whoami` 의 `Example:` 블록
- [ ] JSON 봉투에 `schemaVersion: 1`
- [ ] `.golangci.yml` 에 `errorlint`, `bodyclose`, `gosec`, `gofumpt`, `revive` 추가
- [ ] 패키지 doc comment

---

## 6. Big bets (다주 이니셔티브)

권장 시퀀싱 (대략 분기별 보기; 개발자 1명 가정):

### Q3 2026
1. **Vault format v2 + auth_hash + HKDF salt + machine-bound keyfile encryption** (A1) — 단일 마이그레이션; 9/10 감사 점수; 향후 Argon2 인상 unlock. **4-6주.**
2. **SE 래핑 키 통한 macOS Touch ID** + `internal/biometric/` 패키지 + `tene biometric` 서브명령 (A3) — 사용자 명시 centerpiece. **6-8주.**
3. **`tene serve --mcp` MCP server** (A10) — Infisical/HashiCorp MCP 수렴에 방어; "no `get_secret`" 마케팅 각도. **3-4주.**

### Q4 2026
4. **Sigstore-everywhere + Apple notarization + Windows codesign** (A9) — 최대 신뢰도 향상; Gatekeeper/SmartScreen 마찰 영구 제거. **4주.**
5. **`internal/cli` 의 Hexagonal 재구조화** → `internal/{usecase, port, adapter}/` (A2) — `t.Parallel()` unblock, cloud 재활성화를 1-PR job 으로. **3-4주.**
6. **TTL 의 생체인증 재인증이 있는 `tene agent` 데몬** (A3) — 작업 세션당 Touch ID 한 번 후 모든 `tene run --` 즉시. **2-3주 (#2 이후).**

### 2027 H1
7. **`tene policy` + actor 별 audit trail** (A10) — 컴플라이언스급 forensics; 유료 tier 정당화.
8. **빌드된 binary 통한 e2e 하니스** + 커버리지 게이트 (A6).
9. **modernc.org/sqlite 를 `crawshaw.io/sqlite` 또는 `zombiezen.com/go/sqlite` 로 마이그레이션** (A7) — CGo-free + 빠름.
10. **master key + plaintext slice 용 `memguard`** (A7) — swap + core-dump 저항.

---

## 7. 추천 다음 sprint (구체적 시작점)

위의 Sprint A 8개 항목 번들 + 가장 높은 레버리지 P0 2개 (`tene audit` + `Available()` 프로빙이 있는 Touch ID skeleton). 단일 주의 작업으로 측정 가능한 점수 delta:

- Security 점수: 7.0 → 8.0
- Go convention: 6.5 → 7.5
- CLI UX: 7.5 → 8.0
- AI integration: 8.0 → 8.5
- Cross-platform: 5.5 → 6.5

**Sprint A 후 예상 감사 점수: 6.7 / 10** (오늘 5.95 에서).

Sprint B + C 도 동일 릴리스에 들어가면: **7.2 / 10**.

Touch ID + MCP server + 서명 + Vault v2 의 big-bet 완료: **8.5 / 10** 2026 중반 지평선.

---

## 8. 부록 A — 크로스 플랫폼 생체인증 + 패스워드 fallback 심층 설계

> **이 부록의 동기**: A3 본문에서 macOS Touch ID 를 P0, Windows Hello 를 P1, Linux
> TPM2/fprintd 를 P1 로 다뤘다. 그러나 다음 4가지 질문이 충분히 명시되지 않았다:
>
> 1. 사용자 기기/OS 가 생체인증을 **지원하지 않거나 일시적으로 사용 불가일 때**
>    (CI, SSH, Docker, headless) — 어떻게 자동 감지하고 fallback 하는가?
> 2. 한 사용자가 **여러 기기**를 쓸 때 vault 가 어떻게 portable 한가? (Touch ID
>    노트북 + Windows 데스크탑 + Linux 서버 + CI 동시 운영)
> 3. **lifecycle 이벤트** — 지문 재등록, Secure Enclave 변경, TPM 클리어, 기기
>    분실/교체, OS 재설치 — 각각 어떻게 처리하는가?
> 4. **각 모드의 위협 모델 보장**이 어떻게 달라지는가? (master password only vs
>    +Touch ID vs +SE-wrapped vs +session agent)
>
> **본 부록의 결론**: 사용자가 명시적으로 우려한 두 가지 — "Windows 도 가능하면
> 함께", "생체인증 없는 기기는 지금처럼 master password" — 는 단순 fallback 이상의
> 시스템 설계를 요구한다. 핵심은 **(a) 생체인증을 _불투명한 capability_ 로 추상화
> 하여 호출자가 "있나/없나" 만 알고, (b) master password 가 _모든 모드에서_ 항상
> 유효한 univeral fallback 으로 남으며, (c) `tene biometric status` 가 사용자에게
> 현재 모드와 다음 권장 액션을 한 화면에 보여주는 것**이다.

---

### 8.1 사용자 시나리오 매트릭스 (9 가지)

```
| # | 시나리오                              | 1차 unlock                | Fallback             | 비고                                                       |
|---|--------------------------------------|---------------------------|----------------------|------------------------------------------------------------|
| 1 | macOS Touch ID 노트북                 | SE-wrapped + Touch ID     | master password      | 이상적                                                     |
| 2 | macOS (Touch ID 없음, T2 칩 있음)     | Apple Watch unlock OR pw  | master password      | LocalAuthentication 의 `deviceOwnerAuthentication` 정책   |
| 3 | macOS Intel 구형 (T2 칩 없음)         | (생체인증 미지원)         | master password      | 현재 동작과 동일                                           |
| 4 | Windows 11 + Hello 카메라/지문        | CNG TPM-bound key + 얼굴  | master password      | Windows Hello Enhanced Sign-in                             |
| 5 | Windows 11 PIN-only (생체 등록 안 함) | CNG TPM-bound + Hello PIN | master password      | PIN 도 TPM 보호 비밀; 정책상 허용/거부 토글 가능           |
| 6 | Windows TPM 없거나 비활성             | (생체인증 미지원)         | master password      | 현재 동작과 동일; `status` 에서 "TPM 없음" 명시            |
| 7 | Linux + TPM2 + fprintd                | TPM2 sealed object + 지문 | master password      | 최선의 Linux 시나리오                                      |
| 8 | Linux 데스크탑 (TPM 없음, libsecret)  | libsecret (지문 없음)     | master password      | 현재 동일; status 가 "Linux 에서는 master password 권장"   |
| 9 | CI / SSH / Docker / WSL headless      | (생체인증 강제 skip)      | TENE_MASTER_PASSWORD | non-interactive 감지 → 자동 skip; biometric 프롬프트 안 함 |
```

**핵심 invariant**: 9 시나리오 모두에서 `master password` 는 항상 유효한 unlock
경로다. 생체인증은 *대체*가 아니라 *증강*이다.

---

### 8.2 Capability detection — 어떻게 자동으로 감지하나

`internal/biometric/` 패키지의 `Provider.Available()` 이 빠르게 (≤50ms) 다음을 확인:

#### macOS (`darwin_touchid.go`)
```go
func (p *darwinProvider) Available() Capability {
    // 1. LAContext 인스턴스 생성 → CanEvaluatePolicy
    //    policy = LAPolicy.deviceOwnerAuthenticationWithBiometrics
    //    실패 시 deviceOwnerAuthentication (Touch ID + watch + password) 시도
    // 2. error code 분석:
    //    - LAErrorBiometryNotAvailable    → 하드웨어 없음
    //    - LAErrorBiometryNotEnrolled     → 하드웨어 있지만 enroll 안 됨
    //    - LAErrorPasscodeNotSet          → device passcode 없음 (드물게)
    //    - nil                            → 사용 가능
    // 3. SecAccessControlCreateWithFlags 로 SE 사용 가능 여부 확인
    //    실패하면 fallback 모드 (LA only, no SE wrap)
    return Capability{
        Kind:         BiometryTouchID,            // 또는 BiometryFaceID (Mac 미지원)
        HardwareOk:   ...,
        EnrolledOk:   ...,
        SecureEnclaveOk: ...,
    }
}
```

#### Windows (`windows_hello.go`)
```go
func (p *windowsProvider) Available() Capability {
    // 1. KeyCredentialManager.IsSupportedAsync() 호출 (WinRT)
    //    → false 면 Windows Hello 자체 미지원
    // 2. UserConsentVerifier.CheckAvailabilityAsync() →
    //    - Available                    → 사용 가능
    //    - DeviceNotPresent             → 카메라/지문 센서 없음
    //    - NotConfiguredForUser         → enroll 안 됨
    //    - DisabledByPolicy             → GPO 차단 (기업)
    //    - DeviceBusy                   → 일시적 사용 불가 → retry
    // 3. NCryptOpenStorageProvider(MS_PLATFORM_CRYPTO_PROVIDER)
    //    → TPM 사용 가능 여부 (없으면 SE 등급 보장 불가)
    return Capability{
        Kind:        BiometryWindowsHello,
        HardwareOk:  ...,
        EnrolledOk:  ...,
        TPMBacked:   ...,
    }
}
```

#### Linux (`linux_fprintd.go` 또는 `linux_tpm.go`)
```go
func (p *linuxProvider) Available() Capability {
    // 1. /sys/class/tpm/tpm0 존재 → TPM 2.0 가능성
    //    + go-tpm-tools 로 startup 시도
    // 2. D-Bus 통한 net.reactivated.Fprint.Manager.GetDevices()
    //    → 지문 센서 존재 + enroll 여부
    // 3. 둘 다 없으면 SecretService (libsecret) 만 사용 가능 — biometric=false
    return Capability{
        Kind:       BiometryLinuxFprintd, // 또는 BiometryNone
        HardwareOk: ...,
        TPMBacked:  ...,
    }
}
```

#### 모든 OS 공통 — non-interactive 감지
```go
func IsNonInteractive() bool {
    // 1. stdin 이 TTY 가 아니면 (파이프, redirect)
    if !term.IsTerminal(int(os.Stdin.Fd())) { return true }
    // 2. SSH 세션
    if os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_TTY") != "" { return true }
    // 3. CI 환경 (GitHub Actions, GitLab, CircleCI, Jenkins)
    for _, e := range []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI",
                                "CIRCLECI", "JENKINS_URL", "BUILDKITE"} {
        if os.Getenv(e) != "" { return true }
    }
    // 4. Docker (cgroup v2 감지) — 명시적 환경변수 우선
    if os.Getenv("TENE_BIOMETRIC") == "skip" { return true }
    return false
}
```

Non-interactive 인 경우 `Provider.Available()` 결과를 무시하고 즉시 master
password 경로 (선호: `TENE_MASTER_PASSWORD` env) 로 진행. **CI 파이프라인에서
Touch ID 프롬프트가 영영 hang 되는 사고 방지.**

---

### 8.3 통합된 unlock chain (의사코드)

`loadOrPromptMasterKey()` 가 다음 chain 으로 변모. **첫 성공에서 return.**

```text
loadOrPromptMasterKey(ctx, projectPath):

  ┌─ Priority 1: 환경변수 ───────────────────────────────────┐
  │ if TENE_MASTER_PASSWORD set:                              │
  │   key = Argon2id(env_password, kdf_salt)                  │
  │   verify with auth_hash (P0-1 의 sibling)                 │
  │   return key                                              │
  │ (CI/automation 의 표준 경로 — 항상 1순위)                 │
  └───────────────────────────────────────────────────────────┘

  ┌─ Priority 2: 생체인증 (활성화된 경우, interactive 한정) ──┐
  │ if biometric_enabled in vault_meta:                       │
  │   if IsNonInteractive():                                  │
  │     skip                                                  │
  │   else:                                                   │
  │     cap = biometricProvider.Available()                   │
  │     if cap.HardwareOk and cap.EnrolledOk:                 │
  │       try:                                                │
  │         key = biometricProvider.Unlock("Unlock tene vault")│
  │         return key                              [SUCCESS]│
  │       except UserCancelled:                              │
  │         fall through to next priority                     │
  │       except HardwareFailure:                             │
  │         warn user once; remember failure 이 세션;          │
  │         fall through                                      │
  │     elif cap.HardwareOk and not cap.EnrolledOk:           │
  │       warn: "Touch ID 가 더 이상 enroll 되지 않았습니다.   │
  │              `tene biometric reset` 후 재활성화 필요."     │
  │       fall through                                        │
  │     else:                                                 │
  │       fall through silently                               │
  └───────────────────────────────────────────────────────────┘

  ┌─ Priority 3: keychain 캐시 (기존 동작) ───────────────────┐
  │ if keychain.Load(projectPath) succeeds:                   │
  │   return cached_key                                       │
  │ (Note: 생체인증 활성화 후에도 keychain 캐시는 유지 가능 — │
  │  단, vault_meta.biometric_required=true 면 이 경로 차단)  │
  └───────────────────────────────────────────────────────────┘

  ┌─ Priority 4: master password prompt ──────────────────────┐
  │ if !IsNonInteractive():                                   │
  │   password = term.ReadPassword("Master Password: ")       │
  │   key = Argon2id(password, kdf_salt)                      │
  │   verify with auth_hash                                   │
  │   optionally cache to keychain (사용자 선호)              │
  │   return key                                              │
  │ else:                                                     │
  │   return ErrInteractiveRequired (exit 7)                  │
  └───────────────────────────────────────────────────────────┘

  ┌─ Priority 5: recovery key (사용자가 `--recover` 사용 시) ┐
  │ recovery_phrase = readBIP39FromTTY()                      │
  │ key = deriveRecoveryKey(recovery_phrase, vault.recovery_salt)
  │ unwrap recovery_blob → master key                         │
  │ return key                                                │
  └───────────────────────────────────────────────────────────┘
```

**디자인 노트**:
- Priority 2 의 어떤 분기에서도 master password 가 _자동으로 빠지지 않는다_.
  생체인증이 시도되었다가 사용자가 취소하면 즉시 master password 프롬프트.
- Priority 1 (env var) 이 Priority 2 (생체인증) 보다 앞이라 자동화가 항상 1순위.
- `vault_meta.biometric_required=true` (사용자가 강제 enable) 면 Priority 3 (keychain) 가
  비활성되어 항상 Touch ID/Hello 프롬프트가 뜨도록 강제 가능 — 고보안 모드.

---

### 8.4 Provisioning lifecycle — state machine

```
                       ┌─────────────────┐
                       │  PasswordOnly   │  (현재 모든 사용자의 상태)
                       │  (생체 미설치)   │
                       └────────┬────────┘
                                │ tene biometric enable
                                │ (capability 통과 시)
                                ▼
                       ┌─────────────────┐
                       │   Biometric     │ ←──┐
                       │   Enrolled      │    │ tene biometric reset
                       └────┬────────────┘    │ (지문 재등록, 하드웨어 교체)
                            │                 │
              ┌─────────────┼─────────────┐   │
              │             │             │   │
              ▼             ▼             ▼   │
       ┌──────────┐  ┌──────────┐  ┌─────────┐
       │Hardware  │  │User      │  │ SE Key  │
       │Changed   │  │Cancelled │  │Wiped    │
       └────┬─────┘  └────┬─────┘  └────┬────┘
            │             │              │
            │             │ (한 번)      │
            │             ▼              │
            │     ┌─────────────┐        │
            │     │ Fallback to │        │
            │     │ Password    │        │
            │     │ (이 세션만) │        │
            │     └──────┬──────┘        │
            │            │ next run      │
            │            ▼               │
            │     ┌─────────────┐        │
            │     │ Re-try      │        │
            │     │ Biometric   │        │
            │     └─────────────┘        │
            │                            │
            ▼                            ▼
     ┌─────────────────────────────────────────┐
     │  BiometricBroken                        │
     │  → password 자동 fallback (모든 호출)   │
     │  → 첫 실패 시 1회 경고: "tene biometric  │
     │     reset 으로 재설정하거나 disable 가능" │
     └─────────────────────────────────────────┘
```

**Edge case 처리**:

| 사건 | 감지 방법 | 동작 |
|------|----------|------|
| 사용자가 새 지문 enroll | SE 키 무효화 (`BiometryCurrentSet` 정책) | 다음 unlock 시 자동 실패 → master password fallback → `tene biometric reset` 안내 |
| Apple/Windows 가 SE/TPM 키 wipe | unwrap 실패 (errSecAuthFailed / NTE_BAD_KEY) | 동일 |
| 사용자가 OS 업그레이드 | LocalAuthentication / WinRT API 동작 변경 가능성 | `Available()` 가 매 호출마다 재평가 — 한 번 실패하면 자동 password |
| 사용자가 TPM 클리어 | NCryptOpenKey → NTE_BAD_KEY_USAGE | 동일 |
| 사용자가 기기 분실 | (감지 불가, 새 기기에서 `tene recover` 사용) | recovery key 로 복구 → 새 기기에서 `tene biometric enable` 재실행 |
| 사용자가 password 변경 | `tene passwd` 가 명시적으로 SE/TPM 래핑 키 갱신 트리거 | wrap key 재생성 + 기존 무효화 |

---

### 8.5 vault_meta 스키마 추가

기존 `vault_meta` 테이블에 다음 키 추가 (vault format v2 의 일부 — A1 의 Vault v2 big bet 와 합칠 것):

```sql
-- 생체인증 활성화 여부 (boolean, 기본 false)
INSERT INTO vault_meta (key, value) VALUES ('biometric_enabled', '0');

-- 어떤 생체인증 종류 (BiometryTouchID | BiometryFaceID | BiometryWindowsHello
--                    | BiometryLinuxFprintd | BiometryLinuxTPM2 | None)
INSERT INTO vault_meta (key, value) VALUES ('biometric_kind', '');

-- 활성화 timestamp (감사용)
INSERT INTO vault_meta (key, value) VALUES ('biometric_enrolled_at', '');

-- 마지막 성공 unlock timestamp
INSERT INTO vault_meta (key, value) VALUES ('biometric_last_used', '');

-- 디바이스 식별자 (멀티 디바이스 추적용, 해시된 machine id)
INSERT INTO vault_meta (key, value) VALUES ('biometric_device_id', '');

-- 강제 모드 (true 면 keychain 캐시 비활성, 매번 생체인증 강제)
INSERT INTO vault_meta (key, value) VALUES ('biometric_required', '0');
```

또한 **별도 테이블** `biometric_wrap`:

```sql
CREATE TABLE biometric_wrap (
  device_id TEXT PRIMARY KEY,    -- hash(machine_id || username)
  kind TEXT NOT NULL,            -- BiometryKind
  wrapped_blob BLOB NOT NULL,    -- SE/TPM/fprintd 으로 래핑된 master key
  algorithm TEXT NOT NULL,       -- "se-p256-ecies", "cng-rsa-2048-oaep", "tpm2-sealed"
  enrolled_at INTEGER NOT NULL,
  last_used_at INTEGER
);
```

**왜 별도 테이블?** — 사용자가 여러 기기에서 같은 vault 를 쓸 때 각 기기마다
한 row. 한 기기에서 disable 해도 다른 기기의 wrap 은 유지. `tene push` 가 이
테이블도 동기화 (단, wrap 자체는 device-bound 이라 다른 기기에서 사용 불가 — 새
기기는 한 번 master password 로 unlock 후 `tene biometric enable` 재실행 필요).

---

### 8.6 `tene biometric` 서브명령 완전 명세

```
$ tene biometric --help

Manage biometric authentication for the master vault.

Usage:
  tene biometric [command]

Available Commands:
  status      Show current biometric mode + capability detection
  enable      Enroll biometric for this device (requires master password)
  disable     Remove biometric wrap for this device (falls back to password)
  test        Trigger biometric prompt without side effects
  reset       Re-enroll after hardware/fingerprint change
  list        List devices with biometric enrolled for this vault

Flags:
  --device ID    Operate on a specific device (default: current machine)
  --json         Output structured result
```

#### `tene biometric status`
```
$ tene biometric status

 Current device:        kay-mbp16 (darwin/arm64)
 Hardware:              ✓ Touch ID sensor detected
 Enrolled fingerprints: ✓ 2 fingerprints
 Secure Enclave:        ✓ Available
 vault biometric:       ✓ Enabled (since 2026-05-10)
 Last unlock:           2026-05-12 09:31 KST (Touch ID)
 Fallback ready:        ✓ master password works
 Recovery ready:        ✓ BIP39 phrase available (saved 2026-04-22)

 Other devices on this vault:
   - kay-thinkpad (windows/amd64) — Windows Hello, last used 2026-05-08
   - kay-desktop (linux/amd64)    — not enrolled (master password only)

 Mode: optional (env: TENE_BIOMETRIC=optional)
```

`--json`:
```json
{
  "schemaVersion": 1,
  "device": {
    "id": "kay-mbp16-abc123",
    "os": "darwin",
    "arch": "arm64",
    "hardware": {"touchid": true, "enrolled": 2, "secureEnclave": true}
  },
  "vault": {
    "biometricEnabled": true,
    "biometricKind": "BiometryTouchID",
    "enrolledAt": "2026-05-10T00:00:00Z",
    "lastUsedAt": "2026-05-12T09:31:00+09:00",
    "biometricRequired": false
  },
  "fallback": {
    "masterPassword": true,
    "recoveryKey": true,
    "recoveryEnrolledAt": "2026-04-22T00:00:00Z"
  },
  "mode": "optional",
  "otherDevices": [
    {"id":"kay-thinkpad-...","kind":"BiometryWindowsHello","lastUsed":"..."},
    {"id":"kay-desktop-...","kind":"None","lastUsed":null}
  ]
}
```

#### `tene biometric enable`
```
$ tene biometric enable

 Detecting hardware ...
 ✓ Touch ID with Secure Enclave detected
 ✓ Enrolled fingerprints: 2

 To enable, please enter your current master password.
 This is needed once — afterwards Touch ID will be sufficient.

 Master Password: ●●●●●●●●

 Generating Secure Enclave key ...
 ✓ Key created (algorithm: se-p256-ecies, accessControl: BiometryCurrentSet)

 Wrapping master key ...
 ✓ Stored

 Test:
   (Touch ID prompt: "Unlock tene vault")
   ✓ Successfully unwrapped

 ✓ Biometric enabled for this device.

 Important — biometric does NOT replace these:
   • Master password — still required when biometric fails or on new devices
   • Recovery key   — still required if you lose this device
                      (Run `tene recover` on a new device to restore)

 To disable later:  tene biometric disable
 To re-enroll:      tene biometric reset
```

#### `tene biometric disable`
```
$ tene biometric disable

 Disabling biometric for kay-mbp16 ...
 ✓ Removed Secure Enclave key
 ✓ Removed wrapped blob from vault

 Next `tene` invocation will prompt for master password
 (or use keychain cache if previously stored).

 Other devices (kay-thinkpad) are NOT affected.
```

#### `tene biometric test`
```
$ tene biometric test

 (Touch ID prompt: "tene biometric test — no side effects")
 ✓ Authentication succeeded.
 ✓ Test unwrap matched stored vault key.
```

#### `tene biometric reset`
```
$ tene biometric reset

 You changed your fingerprints or replaced hardware.
 To re-enroll, please enter your current master password.

 Master Password: ●●●●●●●●

 ✓ Old Secure Enclave key invalidated
 ✓ New key generated and master key re-wrapped
 ✓ Test succeeded
```

---

### 8.7 환경변수 + 설정 우선순위

3-level 정책 + 강제 우선순위:

| 변수/플래그 | 의미 | 우선순위 |
|---|---|:---:|
| `TENE_BIOMETRIC=off` | 항상 skip — vault 가 enabled 여도 무시 | **최고** |
| `--no-biometric` flag | 단일 호출에서만 skip | 1 |
| `TENE_BIOMETRIC=required` | 생체인증 실패 시 password fallback 안 함 (고보안 모드) | 1 |
| `vault_meta.biometric_required=1` | 동일 (vault 단위 영구) | 2 |
| `TENE_BIOMETRIC=optional` | 시도하지만 실패하면 password fallback (기본) | 3 |
| `TENE_BIOMETRIC` 미설정 + `vault_meta.biometric_enabled=1` | optional 과 동일 | 3 |
| 모두 미설정 | password only (현재 동작) | 4 |

**자동 strict 모드**: `IsNonInteractive()` true 면 `TENE_BIOMETRIC` 설정과 무관하게
생체인증 시도 안 함 → `TENE_MASTER_PASSWORD` env 또는 exit 7 (interactive required).

---

### 8.8 위협 모델 매트릭스 (모드별 보안 보장)

| 위협 | password only (현재) | + Touch ID/Hello | + SE/TPM wrap | + session agent |
|------|:---:|:---:|:---:|:---:|
| 디스크 분실/도난 (장기 보관) | 🟡 KDF cost 의존 | 🟢 SE 키 없으면 unwrap 불가 | 🟢 동일 | 🟢 동일 |
| 로그인된 세션, 다른 프로세스 keychain 읽기 | 🔴 누설 가능 | 🟡 프롬프트 보임 (사용자 감지 가능) | 🟢 SE 키 사용 시 프롬프트 강제 | 🟢 동일 |
| Coercion (총구 협박) | 🔴 password 강요 | 🟢 손가락 인식 거부 가능 | 🟢 동일 | 🟢 동일 |
| Shoulder surfing | 🔴 password 노출 위험 | 🟢 password 입력 자체 없음 | 🟢 동일 | 🟢 동일 |
| Memory dump (단계 후 RAM) | 🟡 master key 가 short-lived | 🟡 동일 + Touch ID 직후 짧은 노출 | 🟡 동일 | 🟡 동일 (agent 메모리 보호 추가 가능) |
| Process injection (malware 같은 user 권한) | 🔴 keychain 직접 read | 🟡 SE access 가 프롬프트 강제 | 🟢 SE 키 추출 불가 | 🟢 동일 |
| Cold boot attack | 🟡 키가 RAM 에 있을 때만 | 🟡 동일 | 🟡 동일 | 🔴 agent daemon 메모리 keep alive |
| 기기 분실 후 disk forensics | 🟢 vault 자체는 암호화 | 🟢 동일 | 🟢 동일 (SE/TPM 키 추출 사실상 불가) | 🟢 동일 |
| Cross-device replay (vault 카피) | 🟢 password 필요 | 🟢 새 기기에서 password 필요 | 🟢 SE/TPM wrap 은 device-bound (replay 무효) | 🟢 동일 |
| Recovery 손실 | 🔴 vault 영구 손실 | 🔴 동일 | 🔴 동일 (생체인증은 recovery 못 함) | 🔴 동일 |

🟢 강한 보장 / 🟡 부분 보장 / 🔴 보장 없음

**핵심 메시지**: 생체인증 + SE/TPM wrap 은 *coercion*, *shoulder surfing*,
*process injection*, *cross-device replay* 4가지를 추가로 막아준다. 다른 위협들
(메모리 덤프, recovery 손실) 은 변하지 않는다. **이건 master password 의 _대체_ 가
아니라 _증강_ 이라는 의미** — 사용자 우려와 정확히 일치하는 디자인.

---

### 8.9 Migration / rollback path

**기존 사용자가 활성화하는 흐름**:

```bash
# 현재 v1 vault 사용 중인 사용자
$ tene biometric status
 ✗ Vault format v1 — biometric requires v2
 → Run: tene migrate  (one-time, takes ~3s, no data loss)

$ tene migrate
 ✓ Vault v1 → v2
 ✓ Added auth_hash (P0-1 fix)
 ✓ Added biometric_wrap table
 ✓ Backed up to .tene/vault.v1.bak

$ tene biometric enable
 (위 9.6 의 흐름)
```

**Rollback** (사용자가 disable 또는 vault 손상):
```bash
$ tene biometric disable
# → master password 만으로 동작 (현재 동작과 100% 동일)

$ tene biometric reset
# → 손상된 wrap 키만 재생성, vault 본체 안 건드림

$ tene migrate --downgrade-to v1   # (탈출구, 비추천)
# → biometric_wrap 테이블 drop, auth_hash 제거 → v1 호환
```

**기기 분실 시나리오** (가장 우려되는 경로):
```bash
# 새 기기에서:
$ tene init  # 가 기존 vault 감지하지 못함 (다른 기기)
$ tene recover  # BIP39 phrase 입력
 ✓ Master key 복원
 ✓ vault.db 가져오기 (cloud sync 활성화된 경우) OR
   사용자가 백업 vault.db 를 .tene/ 에 복사
$ tene biometric enable  # 이 새 기기에서도 생체인증 활성화
```

**중요**: 분실된 기기의 `biometric_wrap` row 는 cloud 에서도 invalidate 가능 —
`tene biometric list --revoke kay-mbp16-old` (sync 활성화 시).

---

### 8.10 실패 모드 + 사용자 메시지 매트릭스

각 실패에 대해 어떤 메시지가 어디로 가는지 명시 (구체적 사용자 경험):

| 실패 | exit code | stderr 메시지 | 다음 액션 안내 |
|------|:---:|---|---|
| Touch ID 취소 (1회) | 0 (password 로 진행) | `Touch ID cancelled — falling back to master password.` | (자동 진행) |
| Touch ID 5회 실패 | 0 | `Touch ID exceeded retry limit — using master password.` | `Run: tene biometric test` (한 번 권장) |
| 지문 재enroll 감지 | 0 (password 로 진행) | `Biometric key invalidated (likely new fingerprint enrolled). Re-enroll with: tene biometric reset` | password 입력 후 자동 reset suggest |
| TPM 클리어 (Windows) | 0 | `Windows Hello TPM key was wiped. Run: tene biometric reset` | 동일 |
| 하드웨어 사라짐 (외부 카메라 분리) | 0 | (warn 1회만) `Biometric hardware unavailable — using master password.` | (자동 진행) |
| 생체인증 enabled 인데 SSH 세션 | 0 | (info, 1회만) `Non-interactive session detected — biometric skipped.` | `TENE_MASTER_PASSWORD env` 또는 master password prompt |
| `TENE_BIOMETRIC=required` 인데 capability 없음 | 7 | `Biometric is required (TENE_BIOMETRIC=required) but unavailable on this device.` | `Unset TENE_BIOMETRIC or run on a supported device` |
| password 도 틀림 + recovery 도 없음 | 4 | `All unlock methods failed.` | `Run: tene recover` (있으면) / 없으면 vault 영구 손실 안내 |

---

### 8.11 OS 별 API 참고 (구현 노트)

#### macOS — LocalAuthentication.framework + Security.framework
```objc
// LAContext 생성, policy 평가
LAContext *ctx = [[LAContext alloc] init];
NSError *err = nil;
BOOL ok = [ctx canEvaluatePolicy:LAPolicyDeviceOwnerAuthenticationWithBiometrics
                           error:&err];
// LAErrorBiometryNotAvailable / NotEnrolled / PasscodeNotSet → fallback

// SecAccessControl 생성 (BiometryCurrentSet)
SecAccessControlRef ac = SecAccessControlCreateWithFlags(
    NULL,
    kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
    kSecAccessControlBiometryCurrentSet | kSecAccessControlPrivateKeyUsage,
    &err);
// CFErrorRef 처리: kSecMissingEntitlement → app signing 누락

// SE-resident 키 생성 (kSecAttrTokenIDSecureEnclave)
NSDictionary *params = @{
    (id)kSecAttrKeyType:       (id)kSecAttrKeyTypeECSECPrimeRandom,
    (id)kSecAttrKeySizeInBits: @256,
    (id)kSecAttrTokenID:       (id)kSecAttrTokenIDSecureEnclave,
    (id)kSecPrivateKeyAttrs:   @{
        (id)kSecAttrIsPermanent:    @YES,
        (id)kSecAttrApplicationTag: appTagBytes,
        (id)kSecAttrAccessControl:  (__bridge id)ac,
    },
};
```

Go 바인딩 후보 (이미 본문 A3 에서 언급): **`github.com/keybase/go-keychain`**.

#### Windows — WinRT KeyCredentialManager + CNG (NCrypt)
```cpp
// WinRT 비동기 (cgo + winmd 가 필요하거나, Go binding 우회로 small C shim)
KeyCredentialAvailability avail = co_await KeyCredentialManager::IsSupportedAsync();
// → KeyCredentialAvailability::Available / DeviceNotPresent / NotConfiguredForUser
//   / DisabledByPolicy / SecurityDeviceFailure / DeviceBusy

// 새 키 등록 (생성 시점에 Hello 프롬프트)
KeyCredentialRetrievalResult result = co_await
    KeyCredentialManager::RequestCreateAsync(
        L"tene-vault-v1",
        KeyCredentialCreationOption::ReplaceExisting);

// 키로 sign — 매번 Hello 프롬프트
KeyCredentialOperationResult signResult = co_await
    result.Credential().RequestSignAsync(challengeBytes);
```

또는 NCrypt 직접 (TPM 보장 강화):
```c
NCRYPT_PROV_HANDLE hProv;
NCryptOpenStorageProvider(&hProv, MS_PLATFORM_CRYPTO_PROVIDER, 0);
// flags: NCRYPT_USE_VIRTUAL_ISOLATION_FLAG (VBS-isolated 키)
//      + NCRYPT_REQUIRE_HARDWARE_FLAG (TPM 없으면 실패)
//      + NCRYPT_UI_PROTECT_KEY_FLAG (사용 시 Hello 프롬프트)
NCryptCreatePersistedKey(hProv, &hKey, NCRYPT_RSA_ALGORITHM,
                          L"tene-mk-wrap", 0,
                          NCRYPT_REQUIRE_HARDWARE_FLAG |
                          NCRYPT_UI_PROTECT_KEY_FLAG);
```

Go 통합: `github.com/microsoft/go-winio` 가 CNG primitive 일부 노출. WinRT 는
별도 cgo shim (~200 LOC) 또는 `github.com/saltosystems/winrt-go` 평가.

#### Linux — go-tpm-tools + D-Bus fprintd
```go
import "github.com/google/go-tpm-tools/client"

rwc, err := tpm2.OpenTPM("/dev/tpmrm0")  // resource manager
srk, err := client.StorageRootKeyTemplate(rwc, tpm2.HandleOwner)

// Seal master key to PCR state
sealed, err := srk.Seal(masterKey, client.SealOptions{
    Current: client.PCRSelection{
        Hash: tpm2.AlgSHA256,
        PCRs: []int{0, 7},  // BIOS + Secure Boot state
    },
})

// Unseal — 같은 PCR state 일 때만 성공
unsealed, err := srk.Unseal(sealed)
```

fprintd D-Bus:
```go
// auth gate only — wrap 안 됨
conn, _ := dbus.SystemBus()
fprintd := conn.Object("net.reactivated.Fprint",
                        "/net/reactivated/Fprint/Manager")
var devicePath dbus.ObjectPath
fprintd.Call("net.reactivated.Fprint.Manager.GetDefaultDevice",
              0).Store(&devicePath)
// → device.Claim, device.VerifyStart, signal "VerifyStatus"
```

---

### 8.12 본문 A3 와의 관계 — 무엇이 보강되었나

본 부록이 보강한 것:

1. **9 시나리오 매트릭스** — A3 본문이 OS 별 옵션을 나열했지만, *조합* (TPM 있음/없음, enroll 있음/없음, interactive/non-interactive) 의 명확한 의사결정 표는 없었음.
2. **Capability detection 의사코드** — A3 는 "어떤 API 를 쓴다" 만 말함. 본 부록은 "정확히 어떤 분기에서 어떤 결과를 반환한다" 까지 명시.
3. **통합 unlock chain 의사코드** — A3 본문에 한 줄 ("env → biometric → keychain → password → recovery") 만 있었음. 본 부록은 각 priority 의 실패/cancel/skip 분기를 명시.
4. **Lifecycle state machine** — 지문 재enroll, TPM 클리어, 기기 분실 → A3 미언급. 부록에 처리 매트릭스.
5. **vault_meta + biometric_wrap 스키마** — A3 는 인터페이스만 말함; 실제 데이터 모델 명시.
6. **`tene biometric` 5개 서브명령 완전 명세** — A3 는 4개 서브명령 이름만 나열; 부록은 출력 형식, JSON 스키마, 사용자 메시지 포함.
7. **환경변수 + 설정 우선순위 표** — `TENE_BIOMETRIC=off|optional|required` 의미와 충돌 해결 규칙.
8. **위협 모델 매트릭스 (4 모드 × 10 위협)** — 각 모드에서 어떤 보장이 추가되고 무엇이 변하지 않는지 정량.
9. **Migration / rollback 시나리오** — A1 의 Vault v2 와의 연동, 기기 분실 시 흐름.
10. **실패 모드 + 사용자 메시지 매트릭스** — 8개 실패 케이스에 대한 정확한 stderr 메시지 + exit code + 안내.

A3 본문의 점수 (3/10) 는 _현재 상태_ 의 평가로 그대로 유지. 본 부록은 **목표 상태
도달 시 점수 8/10** 의 구체적 청사진. 우선순위는 그대로 P0 (macOS) / P1
(Windows) / P1 (Linux TPM2).

---

### 8.13 사용자의 두 핵심 요구에 대한 답

> **Q1**: "Windows 도 가능하면 함께"
>
> **A**: P1 으로 계획됨. 단, 구현 순서를 명시: (1) 매월 macOS Touch ID 가 작동
> 가능한 MVP 출시 → (2) 동시에 `internal/biometric/` 인터페이스 가 Windows 추가
> 시 0 변경; `windows_hello.go` 만 추가 → (3) Windows Hello 는 2~3개월 후 출시.
> macOS 와 Windows 의 핵심 unlock chain (9.3) 은 OS 무관, 같은 코드.

> **Q2**: "생체인증 없는 기기는 지금처럼 master password"
>
> **A**: 9.1 의 시나리오 #3, #6, #8 (각 OS 의 "미지원" 케이스) 이 정확히 이 경로.
> `Provider.Available()` 가 false 면 unlock chain Priority 2 가 silently skip,
> Priority 4 (password prompt) 가 발동. **사용자는 차이를 느끼지 못함 — 현재
> 동작과 100% 동일**. 9.8 위협 모델 매트릭스의 첫 컬럼이 이 모드의 보장 = 현재
> 동작과 같음.

> **추가 확인된 invariant**: 9 시나리오 모두에서 `master password` 가 항상 유효
> (9.1 표 "Fallback" 컬럼). 9.4 state machine 어떤 path 에서도 password 가
> 빠지는 분기 없음. 9.10 실패 모드 매트릭스 모든 행에서 password 가 다음 액션.

이 invariant 가 깨지는 단 하나의 케이스는 사용자가 명시적으로 `TENE_BIOMETRIC=required`
설정한 경우 — 고보안 모드를 자발적으로 켠 사용자만. 기본 모드 (`optional`) 에서는
password 가 영원히 살아있음.

---

## 9. 방법론 푸트노트

본 감사는 10명의 전문 리뷰어 에이전트가 병렬로 실행하여 생성됨 (2026-05-11).
부록 A (9절) 는 사용자의 후속 우려 — "Windows 도 함께", "생체인증 없는 기기는
지금처럼 master password" — 에 대응하여 2026-05-12 에 보강되었음.
각 에이전트에게 다음이 제공됨:

- 동일한 코드베이스 root (`/Users/popup-kay/Documents/GitHub/agentkay/tene-biz/tene`)
- 하나의 범위 지정된 관점 (security, architecture, biometric auth, Go idioms, CLI UX,
  tests, performance, cross-platform, distribution, AI integration)
- 다음을 요청하는 표준화된 프롬프트: 0-10 점수, top-3 강점, file:line 근거가 있는
  P0-P3 개선점, quick wins (≤1d), big bets (다주)

보고서들이 실질적 발견을 paraphrase 없이 본 단일 문서로 통합됨. 5.95 평균 +
위의 우선순위화는 관점별 점수와 P-label 에서 기계적으로 도출되었으며, 독자는
시퀀싱에 대해 여전히 자체 판단을 행사해야 함, 특히 비즈니스 우선순위 (cloud
재활성화 타이밍, 유료 tier 출시, 채용) 가 기술 우선순위와 교차하는 지점에서.

각 에이전트의 전체 출력은 본 문서의 관점별 섹션에 보존되어 있으며 누락된 내용은
없음.
