# Tene CLI + 랜딩페이지 고도화 계획서

> **Version**: 1.0.0
> **Date**: 2026-04-06
> **Author**: Steve + Claude Code AI Agent
> **Status**: Draft
> **Base Documents**: tene-cli-requirements.md, tene-cli-design.md, tene-cli-implementation.plan.md, tene-mvp.plan.md

---

## 1. 현재 상태 분석

### 1.1 전체 달성률: ~78%

설계 문서 대비 현재 구현의 달성률을 항목별로 산출했다.

| 영역 | 설계 요구 | 구현 완료 | 달성률 | 비고 |
|------|:---------:|:---------:|:------:|------|
| CLI 명령어 (14개) | 14 | 14+1 | 100% | `update` 명령어 추가 구현 (설계에 없음) |
| internal/crypto | 6 함수 | 6 함수 | 100% | XChaCha20-Poly1305 정상 구현 |
| internal/recovery | 4 함수 | 4 함수 | 100% | BIP-39 니모닉 정상 |
| internal/vault | 15+ 메서드 | 15+ 메서드 | 95% | 스키마가 설계와 약간 상이 |
| internal/keychain | KeyStore 인터페이스 + 2 구현체 | 동일 | 100% | |
| internal/claudemd | Generator + Template | 동일 | 100% | |
| 에러 코드 체계 (23개) | 23개 TeneError | 0개 | **0%** | **가장 큰 Gap** |
| Exit code 2 사용 | 인증 에러 시 exit 2 | 미사용 | **0%** | 모든 에러가 exit 1 |
| .tene/vault.json | 파일 생성 | 미구현 | **0%** | |
| ~/.tene/config.json | 글로벌 설정 | 미구현 | **0%** | |
| .tene.enc 바이너리 포맷 | 구조화된 바이너리 | 단순 암호화 blob | **20%** | Magic header 등 미구현 |
| CLI 통합 테스트 | 필수 | **0개** | **0%** | |
| --no-color / NO_COLOR | 색상 제어 | 플래그만 존재, 동작 안 함 | **10%** | |
| 메모리 제로화 | 시크릿 사용 후 zeroing | 미구현 | **0%** | |
| 감사 로그 세분화 | 12종 action | 5종만 사용 | **42%** | import/export/env 미기록 |
| tene sync analytics | config.json 기록 | 미구현 | **0%** | |
| 테스트 커버리지 | 90%+ 목표 | ~60% 추정 | **67%** | |
| goreleaser + CI/CD | 완비 | 완비 | **100%** | |
| 랜딩페이지 | 완비 | 완비 | **90%** | SEO 양호, 일부 개선 필요 |

### 1.2 잘 된 점

1. **암호화 구현 정확**: XChaCha20-Poly1305 (chacha20poly1305.NewX) 사용, nacl/secretbox 아님. 설계서의 최종 결정과 일치
2. **패키��� 구조 깔끔**: 설계서�� 의존성 다이어그램과 정확히 일치 (crypto <- recovery, cli <- 모두)
3. **모든 14개 명령어 구현 완료**: init, set, get, run, list, delete, import, export, env, passwd, recover, sync, whoami, version + 보너스 update
4. **SQLite 스키마 충실**: vault_meta, secrets, environments, audit_log 4개 테이블 모두 구현
5. **Keychain 인터페이스 설계 우수**: KeyStore 인터페이스 + KeyringStore + FileStore 팩토리 패턴
6. **CLAUDE.md URL 정확**: `https://github.com/agentkay/tene` (올바른 org)
7. **goreleaser + CI/CD 완비**: release.yml, ci.yml 모두 설계서와 일치
8. **랜딩페이지 SEO**: JSON-LD, Open Graph, Twitter Card, FAQ 스키마 모두 구현

### 1.3 부족한 점

1. **에러 처리 체계 미구현**: 가장 큰 Gap. TeneError 구조체 없음, 에러 코드 없음, --json 에러 응답 미구현
2. **Exit code 2 미사용**: 인증 에러�� 모두 exit 1로 처리
3. **파일 생성 누락**: .tene/vault.json, ~/.tene/config.json
4. **보안 강화 미구현**: 메모리 제로화, core dump 방지
5. **CLI 통합 테스트 0개**: 설계서에서 필수로 요구
6. **--no-color 미동작**: 플래그는 있으나 실제 색상 제어 로직 없음
7. **.tene.enc 바이너리 포맷 미구현**: 현재는 단순 raw 암호화 blob
8. **감사 로그 불완전**: import, export, env 전환 등 로그 누��
9. **tene sync analytics 미구현**: config.json 기록 없음
10. **DB 스키마 차이**: secrets 테이블이 environment TEXT 직접 참조 (설계는 environment_id FK)

---

## 2. 고도화 항목 분류

### Critical (제품 품질에 직접 영향)

| # | 항목 | 현재 | 목표 | 난이도 |
|---|------|------|------|:------:|
| C1 | 에러 코드 체계 구현 | fmt.Errorf 직접 | TeneError 23개 | 중 |
| C2 | Exit code 2 적용 | 모든 에러 exit 1 | 인증 에러 exit 2 | 하 |
| C3 | --json 에러 응답 | 미구현 | `{"ok":false,"error":"CODE"}` | 중 |
| C4 | CLI 통합 테스트 | 0개 | 핵심 시나리오 15+ | 상 |
| C5 | 메모리 제로화 | 미구현 | masterKey, encKey zeroing | 중 |

### Important (설계 완성도, 사용성)

| # | 항목 | 현재 | 목표 | 난이도 |
|---|------|------|------|:------:|
| I1 | .tene/vault.json 생성 | 미구현 | init 시 자동 생성 | 하 |
| I2 | ~/.tene/config.json | 미구현 | 글로벌 설정 + analytics | 중 |
| I3 | .tene.enc 바이너리 포맷 | raw blob | Magic + header + payload | 중 |
| I4 | --no-color / NO_COLOR 동작 | 플래그만 | TTY 감지 + 색상 제어 | 하 |
| I5 | 감사 로그 세분화 | 5종 | 12종 action 완비 | 하 |
| I6 | tene sync analytics | 미구현 | config.json에 횟수 기록 | 하 |
| I7 | DB 스키마 정규화 | environment TEXT | environment_id FK (선택) | 중 |
| I8 | set --overwrite 시 정확한 version | 추정값 2 | DB에서 실제 version ��회 | 하 |

### Nice to Have (코드 품질, 개발 편의)

| # | 항목 | 현재 | 목표 | 난이도 |
|---|------|------|------|:------:|
| N1 | 테스트 커버리지 90%+ | ~60% | 90%+ | 상 |
| N2 | 코딩 컨벤션 통일 | 혼재 | 일관된 패턴 | 하 |
| N3 | godoc 주석 보강 | 일부 누락 | 모든 exported 함수 | 하 |
| N4 | 매직 넘버/문자열 상수화 | 일부 하드코딩 | 상수 정의 | 하 |
| N5 | init.go 유틸 함수 정리 | splitString 직접 구현 | strings.Fields 활용 | 하 |
| N6 | 랜딩페이지 접근성 | 기본 수준 | WCAG 2.1 AA | 중 |
| N7 | 랜딩페이지 Core Web Vitals | 미측정 | LCP/FID/CLS 최적화 | 중 |
| N8 | CI 테스트 커버리지 리포팅 | 없음 | Codecov/Coveralls 연동 | 하 |

---

## 3. 각 항목별 상세 계획

---

### C1: 에러 코드 체계 구현

**현재 상태**: 각 CLI 명령에서 `fmt.Errorf("메시지")` 직접 반환. 에러 코드 없음. --json 모드에서 에러 응답이 비규격.

**목표 상태**: 23개 에러 코���를 `TeneError` 구조체로 정의. CLI에서 `TeneError` 를 감지하여 적절한 exit code와 --json 응답 생성.

**구현 방법**:

1. `internal/cli/errors.go` 신규 생성:

```go
package cli

import (
    "encoding/json"
    "fmt"
    "os"
)

// TeneError is a structured error with code and exit status.
type TeneError struct {
    Code    string `json:"error"`
    Message string `json:"message"`
    Exit    int    `json:"-"`
}

func (e *TeneError) Error() string {
    return e.Message
}

// 사전 정의 에러
var (
    ErrVaultNotFound          = &TeneError{"VAULT_NOT_FOUND", "Not in a Tene project. Run \"tene init\" first.", 1}
    ErrVaultAlreadyExists     = &TeneError{"VAULT_ALREADY_EXISTS", "Vault already exists. Use existing vault.", 0}
    ErrSecretAlreadyExists    = func(key string) *TeneError {
        return &TeneError{"SECRET_ALREADY_EXISTS", fmt.Sprintf("Secret %q already exists. Use --overwrite to replace.", key), 1}
    }
    ErrEnvironmentNotFound    = func(name string) *TeneError {
        return &TeneError{"ENVIRONMENT_NOT_FOUND", fmt.Sprintf("Environment %q not found. Create it with \"tene env create %s\".", name, name), 1}
    }
    ErrEnvironmentProtected   = func(name string) *TeneError {
        return &TeneError{"ENVIRONMENT_PROTECTED", fmt.Sprintf("Cannot delete the %q environment.", name), 1}
    }
    ErrInvalidKeyName         = func(name string) *TeneError {
        return &TeneError{"INVALID_KEY_NAME", fmt.Sprintf("Invalid key name %q. Keys must match [A-Z][A-Z0-9_]*.", name), 1}
    }
    ErrInvalidEnvName         = &TeneError{"INVALID_ENV_NAME", "Invalid environment name. Must match [a-z][a-z0-9-]*.", 1}
    ErrEmptyValue             = &TeneError{"EMPTY_VALUE", "Value cannot be empty.", 1}
    ErrValueTooLarge          = &TeneError{"VALUE_TOO_LARGE", "Value exceeds maximum size (64KB).", 1}
    ErrPasswordMismatch       = &TeneError{"PASSWORD_MISMATCH", "Passwords do not match. Try again.", 2}
    ErrPasswordTooShort       = &TeneError{"PASSWORD_TOO_SHORT", "Master Password must be at least 8 characters.", 2}
    ErrInvalidPassword        = &TeneError{"INVALID_PASSWORD", "Invalid current Master Password.", 2}
    ErrInvalidRecoveryKey     = &TeneError{"INVALID_RECOVERY_KEY", "Invalid Recovery Key.", 2}
    ErrDecryptFailed          = &TeneError{"DECRYPT_FAILED", "Failed to decrypt secret. Master Password may have changed.", 2}
    ErrFileNotFound           = func(path string) *TeneError {
        return &TeneError{"FILE_NOT_FOUND", fmt.Sprintf("File %q not found.", path), 1}
    }
    ErrInteractiveRequired    = func(cmd string) *TeneError {
        return &TeneError{"INTERACTIVE_REQUIRED", fmt.Sprintf("tene %s requires an interactive terminal.", cmd), 1}
    }
    ErrCommandNotFound        = func(cmd string) *TeneError {
        return &TeneError{"COMMAND_NOT_FOUND", fmt.Sprintf("Command %q not found.", cmd), 127}
    }
    ErrReservedKeyName        = func(name string) *TeneError {
        return &TeneError{"RESERVED_KEY_NAME", fmt.Sprintf("Key name %q is reserved.", name), 1}
    }
)

// handleError는 에러를 처리하여 적절한 출력과 exit code를 결정한다.
func handleError(err error) {
    if err == nil {
        return
    }
    if te, ok := err.(*TeneError); ok {
        if flagJSON {
            resp := map[string]any{"ok": false, "error": te.Code, "message": te.Message}
            json.NewEncoder(os.Stdout).Encode(resp)
        } else {
            fmt.Fprintf(os.Stderr, "Error: %s\n", te.Message)
        }
        os.Exit(te.Exit)
    }
    // 일반 에러
    if flagJSON {
        resp := map[string]any{"ok": false, "error": "INTERNAL_ERROR", "message": err.Error()}
        json.NewEncoder(os.Stdout).Encode(resp)
    } else {
        fmt.Fprintf(os.Stderr, "Error: %s\n", err)
    }
    os.Exit(1)
}
```

2. `cmd/tene/main.go` 수정: `Execute()` 반환값을 `handleError`로 처리

3. 각 CLI 명령 파일에서 `fmt.Errorf` 를 `TeneError` 반환으로 교체

**영향 범위**:
- 신규: `internal/cli/errors.go`
- 수정: `cmd/tene/main.go`, `internal/cli/init.go`, `set.go`, `get.go`, `run.go`, `list.go`, `delete.go`, `import_cmd.go`, `export.go`, `env.go`, `passwd.go`, `recover.go`, `helpers.go`, `root.go`

**예상 소요**: 3-4시간

---

### C2: Exit code 2 적용

**현재 상태**: `cmd/tene/main.go`에서 모든 에러를 `os.Exit(1)`로 처리.

**목표 상태**: 인증/암호화 관련 에러는 exit code 2, 일반 에러는 1, 명령어 미발견은 127.

**구현 방법**: C1의 `TeneError.Exit` 필드를 활용하여 `handleError()`에서 ���기. C1과 동시 구현.

**영향 범위**: `cmd/tene/main.go`
**예상 소요**: C1에 포함 (추가 0.5시간)

---

### C3: --json 에러 응답

**현재 ���태**: 에러 발생 시 --json 플래그가 있어도 stderr에 텍스트로 출력.

**목표 상태**:
```json
{
  "ok": false,
  "error": "SECRET_NOT_FOUND",
  "message": "Secret \"STRIPE_KEY\" not found in \"default\" environment."
}
```

**구현 방법**: C1의 `handleError()` 함수에서 `flagJSON` 체크하여 JSON 출력. C1과 동시 구현.

**영향 범위**: C1과 동일
**예상 소요**: C1에 포함

---

### C4: CLI 통합 테스트

**현재 상���**: CLI 통합 테스트 0개. 단위 테스트만 crypto(12), recovery(5), vault(14), keychain(4), claudemd(4) = 약 39개.

**목표 상태**: 핵심 시나리오 15개 이상의 CLI 통합 테스트.

**구현 방법**:

`internal/cli/cli_test.go` 신규 생성. `os/exec`로 빌드된 바이너리를 호출하거나, cobra의 `Execute()`를 직접 호출하는 방식.

테스트 시나리오:

```go
// internal/cli/cli_test.go

// 1. TestInit_NewProject: tene init -> vault.db 생성 확인
// 2. TestInit_AlreadyExists: 중복 init -> 정상 종료 (exit 0)
// 3. TestInit_JSON: --json 출력 검증
// 4. TestSet_Get_Roundtrip: set -> get 라운드트립
// 5. TestSet_InvalidKeyName: 잘못된 키 이름 에러
// 6. TestSet_EmptyValue: 빈 값 에러
// 7. TestGet_NotFound: 존재하지 않는 키
// 8. TestList_Empty: 시크릿 0개
// 9. TestList_WithSecrets: 시크릿 존재 시 목록
// 10. TestDelete_Existing: 삭제 정상
// 11. TestDelete_NotFound: 존재하지 ���는 키 삭제
// 12. TestImport_DotEnv: .env 파일 가져오기
// 13. TestExport_DotEnv: .env 형식 내보내기
// 14. TestEnv_CreateSwitch: 환경 생성 및 전환
// 15. TestVersion_Output: 버전 출력 검증
// 16. TestSync_FakeDoor: sync 출력 검증
```

테스트 헬퍼:

```go
func setupTestProject(t *testing.T) (string, func()) {
    dir := t.TempDir()
    os.Setenv("TENE_MASTER_PASSWORD", "testpassword123")
    os.Setenv("TENE_KEYCHAIN_FALLBACK", "file")
    // init vault
    return dir, func() {
        os.Unsetenv("TENE_MASTER_PASSWORD")
        os.Unsetenv("TENE_KEYCHAIN_FALLBACK")
    }
}
```

**영향 범위**:
- 신규: `internal/cli/cli_test.go`
- 수정 없음 (테스트만 추가)

**예상 소요**: 6-8시간

---

### C5: 메모리 제로화

**현재 상태**: masterKey, encKey 등 민감 데이터가 GC될 때까지 메모리에 남음.

**목표 상태**: 시크릿 값 사용 후 즉시 `[]byte` 슬라이스를 0으로 덮어쓰기.

**구현 방법**:

1. `internal/crypto/zeroize.go` 신규 생성:

```go
package crypto

// Zeroize overwrites the byte slice with zeros.
func Zeroize(b []byte) {
    for i := range b {
        b[i] = 0
    }
}
```

2. 모든 masterKey, encKey, plaintext 사용처에 `defer crypto.Zeroize(key)` 추가:

```go
// internal/cli/get.go
masterKey, err := loadOrPromptMasterKey(app)
if err != nil { return err }
defer crypto.Zeroize(masterKey)

encKey, err := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)
if err != nil { return err }
defer crypto.Zeroize(encKey)
```

3. 적용 대상 파일:
   - `internal/cli/set.go` (masterKey, encKey)
   - `internal/cli/get.go` (masterKey, encKey, plaintext)
   - `internal/cli/run.go` (masterKey, encKey, 각 plaintext)
   - `internal/cli/list.go` (masterKey, encKey)
   - `internal/cli/import_cmd.go` (masterKey, encKey)
   - `internal/cli/export.go` (masterKey, encKey, decrypted values)
   - `internal/cli/passwd.go` (oldMasterKey, newMasterKey, oldEncKey, newEncKey)
   - `internal/cli/recover.go` (oldMasterKey, newMasterKey, encKeys)
   - `internal/cli/root.go` (loadOrPromptMasterKey 반환값)

**영향 범위**:
- 신규: `internal/crypto/zeroize.go`, `internal/crypto/zeroize_test.go`
- 수정: 위 9개 파��

**예상 소요**: 2-3시간

---

### I1: .tene/vault.json 생성

**현재 상태**: `tene init` 시 vault.json을 생성하지 않음. 메타데이터는 모두 SQLite vault_meta에 저장.

**목표 상태**: 설계서 3.4에 정의된 스키마로 `.tene/vault.json` 생성.

**구현 방법**:

`internal/cli/init.go`의 `runInit()` 함수에 추가:

```go
// Step 10.5: Create .tene/vault.json
vaultJSON := map[string]any{
    "projectName":       projectName,
    "createdAt":         time.Now().UTC().Format(time.RFC3339),
    "vaultVersion":      1,
    "activeEnvironment": "default",
    "agents":            []string{"claude"},
}
vaultJSONBytes, _ := json.MarshalIndent(vaultJSON, "", "  ")
os.WriteFile(filepath.Join(teneDir, "vault.json"), vaultJSONBytes, 0600)
```

환경 전환 시 (`internal/cli/env.go`) vault.json의 `activeEnvironment` 업데이트도 필요:

```go
func updateVaultJSON(dir, activeEnv string) error {
    path := filepath.Join(dir, ".tene", "vault.json")
    // read, update activeEnvironment, write back
}
```

**영향 범위**:
- 수정: `internal/cli/init.go`, `internal/cli/env.go`
- 선택: `internal/cli/helpers.go` (vault.json 읽기/쓰기 헬퍼)

**예상 소요**: 1시간

---

### I2: ~/.tene/config.json

**현재 상태**: 글로벌 설정 파일이 없음.

**목표 상태**: 설계서 3.3에 정의된 스키마로 `~/.tene/config.json` ���성 및 관리.

**구현 방법**:

1. `internal/cli/config.go` 신규 생성:

```go
package cli

type GlobalConfig struct {
    Version            int             `json:"version"`
    DefaultEnvironment string          `json:"defaultEnvironment"`
    Analytics          AnalyticsConfig `json:"analytics"`
    Preferences        Preferences     `json:"preferences"`
}

type AnalyticsConfig struct {
    SyncAttempts    int     `json:"syncAttempts"`
    LastSyncAttempt *string `json:"lastSyncAttempt"`
}

type Preferences struct {
    Color        bool `json:"color"`
    AutoKeychain bool `json:"autoKeychain"`
}

func loadGlobalConfig() (*GlobalConfig, error)
func saveGlobalConfig(cfg *GlobalConfig) error
func globalConfigPath() string  // ~/.tene/config.json
func ensureGlobalConfigDir() error
```

2. `tene sync`에서 analytics 기록:

```go
func runSync(cmd *cobra.Command, args []string) error {
    cfg, _ := loadGlobalConfig()
    cfg.Analytics.SyncAttempts++
    now := time.Now().UTC().Format(time.RFC3339)
    cfg.Analytics.LastSyncAttempt = &now
    saveGlobalConfig(cfg)
    // ... 기존 로직
}
```

**영향 범위**:
- 신규: `internal/cli/config.go`
- 수정: `internal/cli/sync_cmd.go`

**예상 소요**: 1.5시간

---

### I3: .tene.enc 바이너리 포맷

**현재 상태**: `tene export --encrypted`가 단순히 .env 텍스트를 XChaCha20으로 암호화한 raw blob을 저장.

**목표 상태**: 설계서 3.5에 정의된 구조화된 바이너리 ��맷.

```
Offset  Size   Field
0       4      Magic ("TENE" = 0x54454E45)
4       1      Format Version (0x01)
5       1      KDF Algorithm (0x01 = Argon2id)
6       4      KDF Memory (65536, LE uint32)
10      4      KDF Iterations (3, LE uint32)
14      1      KDF Parallelism (1)
15      1      Salt Length (16)
16      16     KDF Salt
32      24     Nonce
56      *      Encrypted Payload (JSON)
```

**구현 방법**:

1. `internal/cli/export.go`의 `exportEncrypted()` 수정:

```go
func exportEncrypted(app *App, env string, keys []string, decrypted map[string]string, encKey []byte) error {
    // JSON payload 생성
    payload := EncryptedPayload{
        ExportVersion: 1,
        ExportedAt:    time.Now().UTC().Format(time.RFC3339),
        ProjectName:   projectName,
        Environments:  environments,
    }
    jsonBytes, _ := json.Marshal(payload)

    // 헤더 작성
    var buf bytes.Buffer
    buf.Write([]byte("TENE"))        // Magic
    buf.WriteByte(0x01)              // Format version
    buf.WriteByte(0x01)              // KDF algorithm
    binary.Write(&buf, binary.LittleEndian, uint32(65536))  // Memory
    binary.Write(&buf, binary.LittleEndian, uint32(3))      // Iterations
    buf.WriteByte(1)                 // Parallelism
    buf.WriteByte(16)                // Salt length
    // ... salt, nonce, encrypted payload
}
```

2. `internal/cli/import_cmd.go`의 `importEncrypted()` 수정:
   - Magic header 검증
   - KDF 파라미터 읽기
   - 기존 raw blob 호환도 유지 (fallback)

**영향 범위**:
- 수정: `internal/cli/export.go`, `internal/cli/import_cmd.go`

**예상 소요**: 3-4시간

---

### I4: --no-color / NO_COLOR 동작

**현재 상태**: `flagNoColor` 변수가 존재하지만 어디에서도 참조되지 않음. 색상 출력 자체가 없음 (plain text만 출력).

**목표 상태**: TTY 감지 + NO_COLOR 환경변수 + --no-color 플래그에 따라 색상 제어. 최소한의 색상 (성공: 초록, 에러: 빨강, 경고: 노랑).

**구현 방법**:

1. `internal/cli/color.go` 신규 생성:

```go
package cli

import "os"

var colorEnabled = true

func initColor() {
    if flagNoColor || os.Getenv("NO_COLOR") != "" || !isTerminal() {
        colorEnabled = false
    }
}

func green(s string) string {
    if !colorEnabled { return s }
    return "\033[32m" + s + "\033[0m"
}

func red(s string) string {
    if !colorEnabled { return s }
    return "\033[31m" + s + "\033[0m"
}

func yellow(s string) string {
    if !colorEnabled { return s }
    return "\033[33m" + s + "\033[0m"
}

func dim(s string) string {
    if !colorEnabled { return s }
    return "\033[2m" + s + "\033[0m"
}
```

2. `root.go`의 `init()`에서 `cobra.OnInitialize(initColor)` 추가

3. 주요 출력에 색상 적용:
   - `init.go`: "Created" -> green, Recovery Key 박스 -> yellow
   - `set.go`: "saved" -> green
   - `delete.go`: "deleted" -> green
   - 에러: red

**영향 범위**:
- 신규: `internal/cli/color.go`
- 수정: `internal/cli/root.go`, `init.go`, `set.go`, `get.go`, `delete.go`, `list.go`, `run.go`

**예상 소요**: 2시간

---

### I5: 감사 로그 세분화

**현재 상태**: `vault.init`, `secret.write`, `secret.read`, `secret.delete`, `vault.passwd_changed`, `vault.recovered`, `secrets.inject` — 7종만 사용.

**목표 상태**: 설계서 정의 12종 모두 기록.

| action | 현�� | 필요 작업 |
|--------|:----:|----------|
| `vault.init` | O | - |
| `vault.passwd_changed` | O | - |
| `vault.recovered` | O | - |
| `secret.write` | O | - |
| `secret.read` | O | - |
| `secret.delete` | O | - |
| `secret.import` | X | `import_cmd.go`에 추가 |
| `secret.export` | X | `export.go`에 추가 |
| `secret.export_encrypted` | X | `export.go`에 추가 |
| `env.create` | X | `env.go`에 추가 |
| `env.delete` | X | `env.go`에 추가 |
| `env.switch` | X | `env.go`에 추가 |

**구현 방법**: 각 해당 파일의 성공 경로에 `app.Vault.AddAuditLog(action, resource, details)` 한 줄 추가.

**영향 범위**: `internal/cli/import_cmd.go`, `export.go`, `env.go`
**예상 소요**: 0.5시간

---

### I6: tene sync analytics

**현재 상태**: `tene sync` 실행 시 화면만 출력, 횟수 추적 없음.

**목표 상태**: `~/.tene/config.json`에 syncAttempts, lastSyncAttempt 기록.

**구현 방법**: I2 (config.json) 구현 후 `sync_cmd.go`에서 호출. I2에 포함.

**영향 범위**: `internal/cli/sync_cmd.go` (I2 의존)
**예상 소요**: I2에 포함

---

### I7: DB 스키마 정규화 (선택적)

**현재 상태**: `secrets.environment` 가 TEXT로 직접 환경 이름 저장. 설계서는 `secrets.environment_id` FK로 정의.

**목표 상태**: 현재 방식 유지 또는 FK 방식으로 변���.

**분석**: 현재 TEXT 방식이 더 단순하고, MVP에서 성능 차이 무의미. 환경 수가 적으므로 FK 정규화의 이점이 크지 않음. **현재 방식 유지를 권장**. 단, 설계 문서와의 차이를 주석으로 명시.

**구현 방법**: `internal/vault/schema.go`에 주석 추가:
```go
// Note: secrets.environment uses TEXT directly (not FK to environments.id)
// for simplicity. The design doc specifies environment_id FK but the current
// approach is functionally equivalent and simpler for MVP.
```

**영향 범위**: `internal/vault/schema.go` (주석만)
**예상 소요**: 5분

---

### I8: set --overwrite 시 정확한 version

**현재 상태**: `set.go`에서 `--overwrite` 시 `version: 2`를 하드코딩 (추정값).

```go
version := 1
if secretExists {
    version = 2 // approximate
}
```

**목표 상태**: DB에서 실제 version 값을 조회하여 반환.

**구현 방법**:

```go
// set.go 수정
if err := app.Vault.SetSecret(keyName, encoded, env); err != nil {
    return err
}

// 저장 후 실제 version 조회
secret, _ := app.Vault.GetSecret(keyName, env)
actualVersion := 1
if secret != nil {
    actualVersion = secret.Version
}
```

**영향 범위**: `internal/cli/set.go`
**예상 소요**: 15분

---

### N1: 테스트 커버리지 90%+

**현재 상태**: 약 60% 추정. crypto(12), recovery(5), vault(14), keychain(4), claudemd(4) = 39개 단위 테스트.

**목표 상태**: 90%+ 커버리지.

**필요한 추가 테스트**:

| 패키지 | 현재 | 추가 필요 |
|--------|------|----------|
| crypto | 12 | 대용량 데이터(1MB), nil aad, 병렬 암호화 |
| recovery | 5 | 잘못된 blob 길이, 빈 mnemonic |
| vault | 14 | SetSecretBatch 에러, EnvironmentExists, 마이그레이션 |
| keychain | 4 (FileStore만) | KeyringStore 모킹 테스트 |
| claudemd | 4 | 빈 파일, 권한 에러 |
| cli | 0 | C4에서 15+ 통합 테스트 |

**영향 범위**: 각 `*_test.go` 파일
**예상 소요**: 4-5시간 (C4 제외)

---

### N2: 코딩 컨벤션 통일

**현재 상태**: 대부분 양호하나 일부 불일치:
- `init.go`의 `splitString()`, `joinWords()` — `strings.Fields()`, `strings.Join()` 사용 가능
- `init.go` 끝의 `var _ = json.Marshal` — 불필요한 import 유지 트릭
- `run.go`의 `extractArgsAfterDash` — 복잡한 로직 단순화 가능

**구현 방법**:

1. `init.go`: `splitMnemonicWords` -> `strings.Fields(mnemonic)`, `joinWords` -> `strings.Join(words, " ")`, `var _ = json.Marshal` 삭제
2. 테스트 네이밍 확인: 현재 `TestXXX` 스타일 ���지 (Go 표준)
3. `promptConfirm`에서 비대화형 기본값 `true` -> `--force` 플래그와의 일관성 확인

**영향 범위**: `internal/cli/init.go`, `run.go`
**예상 소요**: 0.5시간

---

### N3: godoc 주석 보강

**현재 상태**: 대부분의 exported 함수에 godoc 주석 있음. 일부 누락:
- `internal/cli/` 패키지의 `runXxx` 함수들 (unexported이므로 선택)
- `App` 구���체 필드 주석 없음
- `internal/cli/helpers.go`의 helper 함수들

**구현 방법**: `App` 구조체와 exported 헬���에 주석 추가.

**영향 범위**: `internal/cli/root.go`, `helpers.go`
**예상 소요**: 0.5시간

---

### N4: 매직 넘버/문자열 상수화

**현재 상태**:
- `"tene-export"` (export.go, import_cmd.go) — AAD로 사용되는 문자열
- `64 * 1024` (set.go) — 최대 값 크기
- `8` (helpers.go) — 최소 패스워드 길이
- `256` (helpers.go) — 최대 키 이름 길이
- `3` (helpers.go) — 최대 패스워드 시도 횟수

**구현 방법**:

`internal/cli/constants.go` 신규 생성:

```go
package cli

const (
    MinPasswordLength = 8
    MaxKeyNameLength  = 256
    MaxValueSize      = 64 * 1024 // 64KB
    MaxPasswordAttempts = 3
    ExportAAD         = "tene-export"
)
```

**영향 범위**: 신규 `constants.go`, 수정 `set.go`, `helpers.go`, `export.go`, `import_cmd.go`
**예상 소요**: 0.5시간

---

### N5: init.go 유틸 함수 정리

**현재 상태**: `splitString()`, `joinWords()`, `splitMnemonicWords()` 3개 함수가 `strings` 패키지 표준 함수로 대체 가능.

**구현 방법**:

```go
// 변경 전
words := splitMnemonicWords(mnemonic)
fmt.Printf("  |   %-47s|\n", joinWords(words[:6]))

// 변경 후
words := strings.Fields(mnemonic)
fmt.Printf("  |   %-47s|\n", strings.Join(words[:6], " "))
```

`splitString`, `joinWords`, `splitMnemonicWords` 3개 함수 삭제. `var _ = json.Marshal` 도 삭제.

**영향 범위**: `internal/cli/init.go`, `passwd.go`, `recover.go`
**예상 소요**: 15분

---

### N6: 랜딩페이지 접근성 (a11y)

**현재 상태**: 기본적인 시맨틱 HTML 사용. 일부 개선 필요:
- `hero.tsx`의 SVG 아이콘에 `aria-hidden` 누락
- `copy-command.tsx`의 복사 버튼에 `aria-label` 필요
- 색상 대비 비율 미확인 (accent vs background)
- 키보드 내비게이션 테스트 미완

**구현 방법**:
1. 모든 장식용 SVG에 `aria-hidden="true"` 추가
2. 인터랙티브 요소에 `aria-label` 추가
3. Skip to content 링크 추가 (layout.tsx)
4. 색상 대비 비율 ��인 (WCAG 2.1 AA: 4.5:1)

**영향 범위**: `apps/web/src/components/hero.tsx`, `copy-command.tsx`, `nav.tsx`, `layout.tsx`
**예상 소요**: 2시간

---

### N7: 랜딩페이지 Core Web Vitals

**현재 상태**: Next.js 기본 설정. `noise-overlay.tsx` 사용으로 CLS 영향 가능성.

**구현 방법**:
1. Lighthouse로 현재 점수 측정
2. 이미지 최적화 (og-image.png next/image 사용)
3. 폰트 로딩 최적화 (이미 `next/font` 사용 중 — 양호)
4. 불필요한 JS 번들 제거

**영향 범위**: `apps/web/` 전반
**예상 소요**: 2-3시간

---

### N8: CI 테스트 커���리지 리포팅

**현재 상태**: CI에서 `go test ./... -v -count=1` 실행. 커버리지 수집 안 함.

**목표 상태**: Codecov 또는 Coveralls 연동.

**구현 방법**:

`.github/workflows/ci.yml` 수정:

```yaml
- name: Test with coverage
  run: go test ./... -v -count=1 -coverprofile=coverage.out -covermode=atomic

- name: Upload coverage
  uses: codecov/codecov-action@v4
  with:
    file: ./coverage.out
    flags: unittests
```

**영향 범위**: `.github/workflows/ci.yml`
**예상 소요**: 0.5시간

---

## 4. 우선순위 로드맵

### Phase A: 에러 체계 (Day 1 — 4-5시간)

```
C1 에러 코드 체계 ──> C2 Exit code 2 ──> C3 --json 에러 응답
  (3-4시간)            (포함)               (포함)
```

모든 CLI 명령의 에러 처리를 일괄 교체. 가장 큰 Gap이며 다른 개선의 기반.

### Phase B: 보안 강화 + 파일 생성 (Day 2 — 4-5시간)

```
C5 메모리 제로화 ──> I1 vault.json ──> I2 config.json ──> I6 sync analytics
  (2-3시간)            (1시간)            (1.5시간)          (포함)
```

### Phase C: 품질 개선 (Day 3 — 5-6시간)

```
I4 색상 제어 ──> I5 감사 로그 ──> N2 컨벤션 ──> N4 상수화 ──> N5 유틸 정리
  (2시간)         (0.5시간)       (0.5시간)     (0.5시간)     (15분)
```

### Phase D: 테스트 (Day 4-5 — 10-13시간)

```
C4 CLI 통합 테스트 ──> N1 커버리지 90% ──> N8 CI 리포팅
  (6-8시간)             (4-5시간)           (0.5시간)
```

### Phase E: 고급 기능 + 랜딩페이지 (Day 6 — 5-7시간)

```
I3 .tene.enc 포맷 ──> I8 version 정확도
  (3-4시간)             (15분)

N6 접근성 ──> N7 Core Web Vitals
  (2시간)      (2-3시간)
```

### 전체 일정

| Phase | 소요 시간 | ��적 |
|-------|:---------:|:----:|
| A: 에러 체계 | 4-5h | 4-5h |
| B: 보안 + 파�� | 4-5h | 8-10h |
| C: 품질 개선 | 5-6h | 13-16h |
| D: 테스트 | 10-13h | 23-29h |
| E: 고급 + 웹 | 5-7h | 28-36h |
| **총계** | **28-36시간** | ~4-5 working days |

---

## 5. 고도화 후 목��� 달성률

| 영역 | 현재 | Phase A 후 | Phase B 후 | Phase C 후 | Phase D 후 | Phase E 후 |
|------|:----:|:---------:|:---------:|:---------:|:---------:|:---------:|
| 에러 코드 체계 | 0% | **100%** | 100% | 100% | 100% | 100% |
| Exit code 2 | 0% | **100%** | 100% | 100% | 100% | 100% |
| --json 에러 | 0% | **100%** | 100% | 100% | 100% | 100% |
| 메모리 제로화 | 0% | 0% | **100%** | 100% | 100% | 100% |
| vault.json | 0% | 0% | **100%** | 100% | 100% | 100% |
| config.json | 0% | 0% | **100%** | 100% | 100% | 100% |
| --no-color | 10% | 10% | 10% | **100%** | 100% | 100% |
| 감사 로그 | 42% | 42% | 42% | **100%** | 100% | 100% |
| CLI 통합 테스트 | 0% | 0% | 0% | 0% | **100%** | 100% |
| 테스트 커버리지 | ~60% | ~60% | ~60% | ~60% | **90%+** | 90%+ |
| .tene.enc 포맷 | 20% | 20% | 20% | 20% | 20% | **100%** |
| **전체 달성률** | **~78%** | **~85%** | **~89%** | **~92%** | **~96%** | **~99%** |

### 최종 목표: 99% 달성률

남은 1%는:
- DB 스키마 FK 정규화 (I7) — 현재 방식 유지 결정으로 의도적 미이행
- 설계서의 nacl/secretbox 참조 vs 실제 chacha20poly1305 사용 — 구현이 더 올바름 (설계 보정 완료)

---

## 부록: 파일별 수정 요약

| 파일 경로 | 변경 유형 | 관련 항목 |
|----------|----------|---------|
| `internal/cli/errors.go` | **신규** | C1, C2, C3 |
| `internal/cli/config.go` | **신규** | I2, I6 |
| `internal/cli/color.go` | **신규** | I4 |
| `internal/cli/constants.go` | **신규** | N4 |
| `internal/cli/cli_test.go` | **신규** | C4 |
| `internal/crypto/zeroize.go` | **신규** | C5 |
| `internal/crypto/zeroize_test.go` | **신규** | C5 |
| `cmd/tene/main.go` | 수정 | C1, C2 |
| `internal/cli/root.go` | 수정 | C1, I4, N3 |
| `internal/cli/init.go` | 수정 | C1, I1, N2, N5 |
| `internal/cli/set.go` | 수정 | C1, C5, I8, N4 |
| `internal/cli/get.go` | 수정 | C1, C5, I4 |
| `internal/cli/run.go` | 수정 | C1, C5, N2 |
| `internal/cli/list.go` | 수정 | C1, C5 |
| `internal/cli/delete.go` | 수정 | C1, I4 |
| `internal/cli/import_cmd.go` | 수정 | C1, C5, I3, I5, N4 |
| `internal/cli/export.go` | 수정 | C1, C5, I3, I5, N4 |
| `internal/cli/env.go` | 수정 | C1, I1, I5 |
| `internal/cli/passwd.go` | 수정 | C1, C5, N5 |
| `internal/cli/recover.go` | 수정 | C1, C5, N5 |
| `internal/cli/sync_cmd.go` | 수정 | I2, I6 |
| `internal/cli/helpers.go` | 수정 | C1, I1, N4 |
| `internal/vault/schema.go` | 수정 | I7 (주석만) |
| `.github/workflows/ci.yml` | 수정 | N8 |
| `apps/web/src/components/hero.tsx` | 수정 | N6 |
| `apps/web/src/components/copy-command.tsx` | 수정 | N6 |
| `apps/web/src/app/layout.tsx` | 수정 | N6 |
