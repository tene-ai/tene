# Tene CLI 상세 설계 (Design Document)

> **Summary**: Tene Go CLI의 패키지별 상세 설계. 함수 시그니처, struct, interface, SQLite 스키마, 에러 처리, 테스트 설계를 정의
>
> **Project**: Tene
> **Version**: 1.0.0
> **Author**: CTO Lead (Steve)
> **Date**: 2026-04-06
> **Status**: Draft
> **Planning Doc**: [tene-cli-implementation.plan.md](../01-plan/features/tene-cli-implementation.plan.md)

---

## Context Anchor

| Key | Value |
|-----|-------|
| **WHY** | Go CLI 구현을 위한 코드 레벨 상세 설계. Plan에서 정의한 패키지 구조를 함수 시그니처와 인터페이스 수준으로 구체화 |
| **WHO** | Tene CLI를 구현하는 Go 개발자 (Steve + Claude Code AI Agent) |
| **RISK** | 암호화 구현 결함 시 제품 신뢰도 치명적 타격. modernc.org/sqlite 순수 Go 성능 검증 필요. OS Keychain 크로스 플랫폼 호환성 |
| **SUCCESS** | 13개 CLI 명령어 로컬 동작, 오프라인 100%, XChaCha20-Poly1305 + Argon2id 암호화, CLAUDE.md 자동 생성, 테스트 커버리지 90%+ |
| **SCOPE** | Week 1: crypto + recovery + vault + keychain. Week 2: claudemd + cli + cmd/tene + goreleaser |

---

## 1. Overview

### 1.1 설계 목표

- 패키지 간 의존성을 최소화하여 독립 테스트 및 병렬 개발 가능
- 인터페이스 기반 설계로 테스트 시 모킹 용이
- 모든 공개 함수의 입출력 타입을 명확히 정의
- 에러 타입을 패키지별로 분리하여 CLI에서 적절한 메시지 출력

### 1.2 설계 원칙

- **단일 책임**: 각 패키지는 하나의 영역만 담당
- **의존성 역전**: cli 패키지는 인터페이스에 의존, 구현체는 주입
- **실패 명시**: 모든 에러는 sentinel 에러로 정의, fmt.Errorf %w로 래핑
- **순수 Go**: CGo 불필요, 모든 의존성은 순수 Go 라이브러리

---

## 2. 패키지 다이어그램

```
cmd/tene/main.go
    |
    v
internal/cli        (cobra 명령어 정의)
    |
    +--- internal/crypto       암호화/복호화/KDF
    |        ^
    |        |
    +--- internal/recovery     BIP-39 니모닉, Master Key 복구
    |
    +--- internal/vault        SQLite CRUD
    |
    +--- internal/keychain     OS Keychain 저장/로드
    |
    +--- internal/claudemd     CLAUDE.md 생성


의존성 방향:
  cmd/tene -> cli -> crypto, recovery, vault, keychain, claudemd
  recovery -> crypto
  (vault, keychain, claudemd는 서로 독립)
```

### 2.1 패키지 간 의존성 매트릭스

| 패키지 | crypto | recovery | vault | keychain | claudemd | cli |
|--------|:------:|:--------:|:-----:|:--------:|:--------:|:---:|
| **crypto** | - | | | | | |
| **recovery** | O | - | | | | |
| **vault** | | | - | | | |
| **keychain** | | | | - | | |
| **claudemd** | | | | | - | |
| **cli** | O | O | O | O | O | - |

---

## 3. internal/crypto 상세 설계

### 3.1 상수 정의

```go
package crypto

const (
    // Argon2id 파라미터
    ArgonTime    = 3         // iterations
    ArgonMemory  = 64 * 1024 // 64MB
    ArgonThreads = 1         // parallelism
    ArgonKeyLen  = 32        // 256-bit output

    // Salt/Nonce 길이
    SaltLen  = 16 // 128-bit salt
    NonceLen = 24 // 192-bit nonce (XChaCha20)

    // HKDF 목적별 라벨
    PurposeEncryption = "tene-encryption-key"
    PurposeAuth       = "tene-auth-hash"
)
```

### 3.2 함수 시그니처

```go
// kdf.go

// GenerateSalt는 128-bit 랜덤 salt를 생성한다.
func GenerateSalt() ([]byte, error)

// DeriveKey는 Master Password에서 Argon2id로 Master Key를 유도한다.
// password: 사용자 입력 패스워드
// salt: 128-bit salt (GenerateSalt으로 생성)
// 반환: 256-bit Master Key
func DeriveKey(password string, salt []byte) ([]byte, error)
```

```go
// keymanager.go

// DeriveSubKey는 Master Key에서 HKDF-SHA256으로 용도별 서브키를 유도한다.
// masterKey: 256-bit Master Key
// purpose: 키 용도 라벨 (PurposeEncryption, PurposeAuth)
// length: 출력 키 길이 (바이트)
func DeriveSubKey(masterKey []byte, purpose string, length int) ([]byte, error)
```

```go
// encrypt.go

// Encrypt는 XChaCha20-Poly1305로 평문을 암호화한다.
// key: 256-bit Encryption Key (DeriveSubKey로 유도)
// plaintext: 암호화할 평문
// aad: Additional Authenticated Data (시크릿 키 이름)
// 반환: nonce(24바이트) + ciphertext (base64 인코딩 전 원시 바이트)
func Encrypt(key, plaintext, aad []byte) ([]byte, error)
```

```go
// decrypt.go

// Decrypt는 XChaCha20-Poly1305로 암호문을 복호화한다.
// key: 256-bit Encryption Key
// ciphertext: nonce(24바이트) + 암호문
// aad: Additional Authenticated Data (암호화 시 사용한 것과 동일해야 함)
// 반환: 복호화된 평문
func Decrypt(key, ciphertext, aad []byte) ([]byte, error)
```

### 3.3 에러 타입

```go
// errors.go

package crypto

import "errors"

var (
    // ErrInvalidKeyLength는 키 길이가 잘못된 경우 반환된다.
    ErrInvalidKeyLength = errors.New("crypto: invalid key length")

    // ErrInvalidSaltLength는 salt 길이가 잘못된 경우 반환된다.
    ErrInvalidSaltLength = errors.New("crypto: invalid salt length")

    // ErrDecryptionFailed는 복호화에 실패한 경우 반환된다 (잘못된 키 또는 변조).
    ErrDecryptionFailed = errors.New("crypto: decryption failed")

    // ErrInvalidCiphertext는 암호문 형식이 잘못된 경우 반환된다.
    ErrInvalidCiphertext = errors.New("crypto: invalid ciphertext format")
)
```

### 3.4 암호화 플로우 상세

```
Encrypt(key, plaintext, aad):
  1. key 길이 검증 (32바이트)
  2. nonce = crypto/rand.Read(24바이트)
  3. var key32 [32]byte; copy(key32[:], key)
  4. var nonce24 [24]byte; copy(nonce24[:], nonce)
  5. sealed = secretbox.Seal(nonce, plaintext, &nonce24, &key32)
     -- secretbox.Seal은 nonce를 prefix로 사용하므로 결과 = nonce + ciphertext
  6. return sealed, nil

Decrypt(key, ciphertext, aad):
  1. key 길이 검증 (32바이트)
  2. ciphertext 최소 길이 검증 (> NonceLen)
  3. nonce = ciphertext[:24]
  4. message = ciphertext[24:]
  5. var key32 [32]byte; copy(key32[:], key)
  6. var nonce24 [24]byte; copy(nonce24[:], nonce)
  7. opened, ok = secretbox.Open(nil, message, &nonce24, &key32)
  8. if !ok: return nil, ErrDecryptionFailed
  9. return opened, nil
```

**참고**: `golang.org/x/crypto/nacl/secretbox`은 XSalsa20-Poly1305를 사용한다. XChaCha20-Poly1305가 필요하면 `golang.org/x/crypto/chacha20poly1305`의 `NewX()`를 사용해야 한다. 설계 결정: **chacha20poly1305.NewX()를 사용**하여 진정한 XChaCha20-Poly1305를 구현한다.

```go
// encrypt.go (실제 구현)

import "golang.org/x/crypto/chacha20poly1305"

func Encrypt(key, plaintext, aad []byte) ([]byte, error) {
    if len(key) != chacha20poly1305.KeySize {
        return nil, fmt.Errorf("%w: got %d, want %d", ErrInvalidKeyLength, len(key), chacha20poly1305.KeySize)
    }

    aead, err := chacha20poly1305.NewX(key)
    if err != nil {
        return nil, fmt.Errorf("crypto: failed to create AEAD: %w", err)
    }

    nonce := make([]byte, aead.NonceSize()) // 24 bytes for XChaCha20
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, fmt.Errorf("crypto: failed to generate nonce: %w", err)
    }

    // Seal appends ciphertext to nonce prefix
    ciphertext := aead.Seal(nonce, nonce, plaintext, aad)
    return ciphertext, nil
}

func Decrypt(key, ciphertext, aad []byte) ([]byte, error) {
    if len(key) != chacha20poly1305.KeySize {
        return nil, fmt.Errorf("%w: got %d, want %d", ErrInvalidKeyLength, len(key), chacha20poly1305.KeySize)
    }

    aead, err := chacha20poly1305.NewX(key)
    if err != nil {
        return nil, fmt.Errorf("crypto: failed to create AEAD: %w", err)
    }

    if len(ciphertext) < aead.NonceSize() {
        return nil, fmt.Errorf("%w: too short", ErrInvalidCiphertext)
    }

    nonce := ciphertext[:aead.NonceSize()]
    message := ciphertext[aead.NonceSize():]

    plaintext, err := aead.Open(nil, nonce, message, aad)
    if err != nil {
        return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
    }

    return plaintext, nil
}
```

### 3.5 KDF 상세 구현

```go
// kdf.go

import "golang.org/x/crypto/argon2"

func GenerateSalt() ([]byte, error) {
    salt := make([]byte, SaltLen)
    if _, err := io.ReadFull(rand.Reader, salt); err != nil {
        return nil, fmt.Errorf("crypto: failed to generate salt: %w", err)
    }
    return salt, nil
}

func DeriveKey(password string, salt []byte) ([]byte, error) {
    if len(salt) != SaltLen {
        return nil, fmt.Errorf("%w: got %d, want %d", ErrInvalidSaltLength, len(salt), SaltLen)
    }
    if password == "" {
        return nil, errors.New("crypto: password cannot be empty")
    }

    key := argon2.IDKey(
        []byte(password),
        salt,
        ArgonTime,    // 3 iterations
        ArgonMemory,  // 64MB
        ArgonThreads, // 1
        ArgonKeyLen,  // 32 bytes
    )
    return key, nil
}
```

### 3.6 HKDF 서브키 유도

```go
// keymanager.go

import (
    "crypto/sha256"
    "golang.org/x/crypto/hkdf"
)

func DeriveSubKey(masterKey []byte, purpose string, length int) ([]byte, error) {
    if len(masterKey) != 32 {
        return nil, fmt.Errorf("%w: master key must be 32 bytes", ErrInvalidKeyLength)
    }

    hkdfReader := hkdf.New(sha256.New, masterKey, nil, []byte(purpose))
    subKey := make([]byte, length)
    if _, err := io.ReadFull(hkdfReader, subKey); err != nil {
        return nil, fmt.Errorf("crypto: HKDF expand failed: %w", err)
    }
    return subKey, nil
}
```

---

## 4. internal/vault 상세 설계

### 4.1 SQLite 스키마 (DDL)

```sql
-- schema version tracking
CREATE TABLE IF NOT EXISTS vault_meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- encrypted secrets
CREATE TABLE IF NOT EXISTS secrets (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT    NOT NULL,
    encrypted_value TEXT    NOT NULL,
    environment     TEXT    NOT NULL DEFAULT 'default',
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE(name, environment)
);

CREATE INDEX IF NOT EXISTS idx_secrets_env ON secrets(environment);
CREATE INDEX IF NOT EXISTS idx_secrets_name ON secrets(name);

-- environment management
CREATE TABLE IF NOT EXISTS environments (
    name       TEXT    PRIMARY KEY,
    is_active  INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- audit log
CREATE TABLE IF NOT EXISTS audit_log (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    action        TEXT NOT NULL,
    resource_name TEXT,
    details       TEXT,
    timestamp     TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_log(timestamp);
```

### 4.2 Go Struct 정의

```go
// models.go

package vault

import "time"

// Secret은 암호화된 시크릿 레코드를 나타낸다.
type Secret struct {
    ID             int64
    Name           string
    EncryptedValue string // base64(nonce + ciphertext)
    Environment    string
    Version        int
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

// Environment는 시크릿 환경 설정을 나타낸다.
type Environment struct {
    Name      string
    IsActive  bool
    CreatedAt time.Time
}

// AuditEntry는 감사 로그 항목을 나타낸다.
type AuditEntry struct {
    ID           int64
    Action       string // "secret.read", "secret.write", "secret.delete", "vault.init", "vault.passwd"
    ResourceName string
    Details      string
    Timestamp    time.Time
}
```

### 4.3 Vault 구조체 및 인터페이스

```go
// vault.go

package vault

import "database/sql"

// Vault는 SQLite 기반 시크릿 저장소이다.
type Vault struct {
    db     *sql.DB
    dbPath string
}

// New는 지정 경로에 SQLite 볼트를 생성하거나 열고, 스키마를 초기화한다.
// dbPath가 존재하지 않으면 새 DB 파일을 생성한다.
func New(dbPath string) (*Vault, error)

// Close는 데이터베이스 연결을 닫는다.
func (v *Vault) Close() error

// --- 시크릿 CRUD ---

// SetSecret은 시크릿을 저장한다 (UPSERT: name+environment 기준).
// 기존 시크릿이 있으면 version을 증가시키고 updated_at을 갱신한다.
func (v *Vault) SetSecret(name, encryptedValue, env string) error

// GetSecret은 시크릿을 조회한다.
// 존재하지 않으면 ErrSecretNotFound를 반환한다.
func (v *Vault) GetSecret(name, env string) (*Secret, error)

// ListSecrets는 지정 환경의 모든 시크릿을 반환한다.
func (v *Vault) ListSecrets(env string) ([]Secret, error)

// DeleteSecret은 시크릿을 삭제한다.
// 존재하지 않으면 ErrSecretNotFound를 반환한다.
func (v *Vault) DeleteSecret(name, env string) error

// SecretExists는 시크릿 존재 여부를 반환한다.
func (v *Vault) SecretExists(name, env string) (bool, error)

// CountSecrets는 지정 환경의 시크릿 수를 반환한다.
func (v *Vault) CountSecrets(env string) (int, error)

// --- 환경 관리 ---

// ListEnvironments는 모든 환경을 반환한다.
func (v *Vault) ListEnvironments() ([]Environment, error)

// GetActiveEnvironment는 현재 활성 환경 이름을 반환한다.
func (v *Vault) GetActiveEnvironment() (string, error)

// SetActiveEnvironment는 활성 환경을 변경한다.
// 환경이 존재하지 않으면 새로 생성한다.
func (v *Vault) SetActiveEnvironment(name string) error

// CreateEnvironment는 새 환경을 생성한다.
func (v *Vault) CreateEnvironment(name string) error

// --- 메타데이터 ---

// SetMeta는 볼트 메타데이터를 저장한다 (UPSERT).
func (v *Vault) SetMeta(key, value string) error

// GetMeta는 볼트 메타데이터를 조회한다.
// 존재하지 않으면 ErrMetaNotFound를 반환한다.
func (v *Vault) GetMeta(key string) (string, error)

// --- 감사 로그 ---

// AddAuditLog는 감사 로그 항목을 추가한다.
func (v *Vault) AddAuditLog(action, resourceName, details string) error

// --- 일괄 작업 ---

// SetSecretBatch는 여러 시크릿을 트랜잭션으로 일괄 저장한다.
func (v *Vault) SetSecretBatch(secrets map[string]string, env string) error

// GetAllSecrets는 지정 환경의 모든 시크릿을 name->encryptedValue 맵으로 반환한다.
func (v *Vault) GetAllSecrets(env string) (map[string]string, error)
```

### 4.4 에러 타입

```go
// errors.go

package vault

import "errors"

var (
    // ErrSecretNotFound는 요청한 시크릿이 존재하지 않을 때 반환된다.
    ErrSecretNotFound = errors.New("vault: secret not found")

    // ErrMetaNotFound는 요청한 메타데이터가 존재하지 않을 때 반환된다.
    ErrMetaNotFound = errors.New("vault: metadata not found")

    // ErrEnvironmentNotFound는 요청한 환경이 존재하지 않을 때 반환된다.
    ErrEnvironmentNotFound = errors.New("vault: environment not found")

    // ErrVaultNotInitialized는 볼트가 초기화되지 않았을 때 반환된다.
    ErrVaultNotInitialized = errors.New("vault: not initialized, run 'tene init' first")

    // ErrDatabaseCorrupted는 데이터베이스가 손상되었을 때 반환된다.
    ErrDatabaseCorrupted = errors.New("vault: database corrupted")
)
```

### 4.5 마이그레이션 전략

```go
// migration.go

package vault

// 현재 스키마 버전
const CurrentSchemaVersion = 1

// migrate는 스키마 버전을 확인하고 필요 시 마이그레이션을 수행한다.
func (v *Vault) migrate() error {
    version, err := v.getSchemaVersion()
    if err != nil {
        // 첫 실행: 스키마 초기화
        return v.initSchema()
    }

    // 향후 마이그레이션: version < CurrentSchemaVersion
    switch {
    case version < 2:
        // v1 -> v2 마이그레이션 (향후)
    }

    return nil
}

func (v *Vault) getSchemaVersion() (int, error) {
    val, err := v.GetMeta("schema_version")
    if err != nil {
        return 0, err
    }
    return strconv.Atoi(val)
}
```

### 4.6 UPSERT 구현

```go
// vault.go (SetSecret 내부)

func (v *Vault) SetSecret(name, encryptedValue, env string) error {
    query := `
        INSERT INTO secrets (name, encrypted_value, environment, version, created_at, updated_at)
        VALUES (?, ?, ?, 1, datetime('now'), datetime('now'))
        ON CONFLICT(name, environment) DO UPDATE SET
            encrypted_value = excluded.encrypted_value,
            version = secrets.version + 1,
            updated_at = datetime('now')
    `
    _, err := v.db.Exec(query, name, encryptedValue, env)
    if err != nil {
        return fmt.Errorf("vault: failed to set secret %q: %w", name, err)
    }

    return v.AddAuditLog("secret.write", name, "")
}
```

---

## 5. internal/keychain 상세 설계

### 5.1 인터페이스 정의

```go
// keychain.go

package keychain

// KeyStore는 Master Key를 안전하게 저장하고 로드하는 인터페이스이다.
type KeyStore interface {
    // Store는 Master Key를 저장한다.
    Store(key []byte) error

    // Load는 저장된 Master Key를 로드한다.
    // 저장된 키가 없으면 ErrKeyNotFound를 반환한다.
    Load() ([]byte, error)

    // Delete는 저장된 Master Key를 삭제한다.
    Delete() error

    // Exists는 Master Key가 저장되어 있는지 확인한다.
    Exists() bool
}
```

### 5.2 go-keyring 구현

```go
// keychain.go

import "github.com/zalando/go-keyring"

const (
    ServiceName = "tene"
    AccountName = "master-key"
)

// KeyringStore는 OS 키체인을 사용하는 KeyStore 구현이다.
type KeyringStore struct {
    service string // 키체인 서비스 이름 (프로젝트별 구분)
}

// NewKeyringStore는 OS 키체인 기반 KeyStore를 생성한다.
// service: 프로젝트별 고유 식별자 (예: "tene-/path/to/project")
func NewKeyringStore(service string) *KeyringStore

func (k *KeyringStore) Store(key []byte) error {
    encoded := base64.StdEncoding.EncodeToString(key)
    return keyring.Set(k.service, AccountName, encoded)
}

func (k *KeyringStore) Load() ([]byte, error) {
    encoded, err := keyring.Get(k.service, AccountName)
    if err != nil {
        if err == keyring.ErrNotFound {
            return nil, ErrKeyNotFound
        }
        return nil, fmt.Errorf("keychain: failed to load key: %w", err)
    }
    return base64.StdEncoding.DecodeString(encoded)
}

func (k *KeyringStore) Delete() error {
    return keyring.Delete(k.service, AccountName)
}

func (k *KeyringStore) Exists() bool {
    _, err := keyring.Get(k.service, AccountName)
    return err == nil
}
```

### 5.3 파일 폴백 구현

```go
// fallback.go

// FileStore는 파일 시스템을 사용하는 KeyStore 폴백 구현이다.
// OS 키체인을 사용할 수 없는 환경 (CI, Docker, headless server)에서 사용된다.
// 파일 퍼미션은 0600으로 제한된다.
type FileStore struct {
    path string // 키 파일 경로 (예: ~/.tene/keyfile)
}

// NewFileStore는 파일 기반 KeyStore를 생성한다.
func NewFileStore(path string) *FileStore

func (f *FileStore) Store(key []byte) error {
    dir := filepath.Dir(f.path)
    if err := os.MkdirAll(dir, 0700); err != nil {
        return err
    }
    encoded := base64.StdEncoding.EncodeToString(key)
    return os.WriteFile(f.path, []byte(encoded), 0600)
}

func (f *FileStore) Load() ([]byte, error) {
    data, err := os.ReadFile(f.path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, ErrKeyNotFound
        }
        return nil, err
    }
    return base64.StdEncoding.DecodeString(string(data))
}
```

### 5.4 팩토리 함수

```go
// keychain.go

// NewStore는 환경에 따라 적절한 KeyStore를 반환한다.
// 1. TENE_KEYCHAIN_FALLBACK=file이면 FileStore
// 2. OS 키체인 사용 가능하면 KeyringStore
// 3. OS 키체인 실패 시 FileStore 폴백
func NewStore(projectPath string) KeyStore {
    if os.Getenv("TENE_KEYCHAIN_FALLBACK") == "file" {
        return NewFileStore(filepath.Join(os.UserHomeDir(), ".tene", "keyfile"))
    }

    service := ServiceName + "-" + hashPath(projectPath)
    ks := NewKeyringStore(service)

    // 키체인 사용 가능 여부 테스트
    testKey := "keychain-test"
    if err := keyring.Set(service, testKey, "test"); err != nil {
        // 키체인 불가 -> 파일 폴백
        return NewFileStore(filepath.Join(os.UserHomeDir(), ".tene", "keyfile"))
    }
    keyring.Delete(service, testKey)

    return ks
}
```

### 5.5 에러 타입

```go
// errors.go

package keychain

import "errors"

var (
    // ErrKeyNotFound는 저장된 키가 없을 때 반환된다.
    ErrKeyNotFound = errors.New("keychain: master key not found")

    // ErrKeychainUnavailable는 OS 키체인을 사용할 수 없을 때 반환된다.
    ErrKeychainUnavailable = errors.New("keychain: OS keychain unavailable")
)
```

### 5.6 OS별 동작

| OS | Backend | 보안 수준 | 비고 |
|----|---------|----------|------|
| macOS | Keychain Services | T2/Apple Silicon HW 암호화 | go-keyring 네이티브 지원 |
| Linux | libsecret (GNOME Keyring / KWallet) | 로그인 세션 암호화 | D-Bus 필요, headless는 폴백 |
| Windows | Credential Vault (DPAPI) | DPAPI 사용자별 암호화 | go-keyring 네이티브 지원 |
| CI/Docker | 파일 폴백 | 0600 퍼미션 | `TENE_KEYCHAIN_FALLBACK=file` 자동 감지 |

---

## 6. internal/recovery 상세 설계

### 6.1 니모닉 생성 플로우

```go
// mnemonic.go

package recovery

import (
    "github.com/tyler-smith/go-bip39"
    "github.com/tomo-kay/tene/internal/crypto"
)

// GenerateMnemonic은 128-bit 엔트로피에서 12단어 BIP-39 니모닉을 생성한다.
func GenerateMnemonic() (string, error) {
    entropy, err := bip39.NewEntropy(128)
    if err != nil {
        return "", fmt.Errorf("recovery: failed to generate entropy: %w", err)
    }
    mnemonic, err := bip39.NewMnemonic(entropy)
    if err != nil {
        return "", fmt.Errorf("recovery: failed to generate mnemonic: %w", err)
    }
    return mnemonic, nil
}

// ValidateMnemonic은 BIP-39 니모닉의 유효성을 검증한다.
func ValidateMnemonic(mnemonic string) bool {
    return bip39.IsMnemonicValid(mnemonic)
}
```

### 6.2 Recovery Key로 Master Key 암호화/복구 플로우

```go
// recover.go

const (
    RecoverySalt    = "tene-recovery"
    RecoveryPurpose = "tene-recovery-key"
)

// EncryptMasterKey는 니모닉에서 Recovery Key를 유도하고,
// 이를 사용하여 Master Key를 암호화한다.
// 반환값(blob)은 vault_meta에 'recovery_blob'으로 저장한다.
func EncryptMasterKey(masterKey []byte, mnemonic string) ([]byte, error) {
    if !ValidateMnemonic(mnemonic) {
        return nil, ErrInvalidMnemonic
    }

    // 니모닉에서 Recovery Key 유도
    recoveryKey := crypto.DeriveKey(mnemonic, []byte(RecoverySalt))

    // Recovery Key에서 Encryption 서브키 유도
    encKey, err := crypto.DeriveSubKey(recoveryKey, RecoveryPurpose, 32)
    if err != nil {
        return nil, err
    }

    // Master Key 암호화
    blob, err := crypto.Encrypt(encKey, masterKey, []byte("recovery"))
    if err != nil {
        return nil, fmt.Errorf("recovery: failed to encrypt master key: %w", err)
    }
    return blob, nil
}

// RecoverMasterKey는 니모닉에서 Recovery Key를 유도하고,
// 암호화된 blob에서 Master Key를 복구한다.
func RecoverMasterKey(blob []byte, mnemonic string) ([]byte, error) {
    if !ValidateMnemonic(mnemonic) {
        return nil, ErrInvalidMnemonic
    }

    recoveryKey := crypto.DeriveKey(mnemonic, []byte(RecoverySalt))

    encKey, err := crypto.DeriveSubKey(recoveryKey, RecoveryPurpose, 32)
    if err != nil {
        return nil, err
    }

    masterKey, err := crypto.Decrypt(encKey, blob, []byte("recovery"))
    if err != nil {
        return nil, fmt.Errorf("%w: invalid recovery key", ErrRecoveryFailed)
    }
    return masterKey, nil
}
```

### 6.3 에러 타입

```go
// errors.go

package recovery

import "errors"

var (
    ErrInvalidMnemonic = errors.New("recovery: invalid mnemonic phrase")
    ErrRecoveryFailed  = errors.New("recovery: master key recovery failed")
)
```

---

## 7. internal/claudemd 상세 설계

### 7.1 CLAUDE.md 템플릿 (확정본, 영어)

```go
// template.go

package claudemd

const SectionHeader = "# Secrets Management"

const SecretsMdTemplate = `# Secrets Management

This project uses [tene](https://github.com/tomo-kay/tene) for secret management.

## Usage
- Get a secret: ` + "`tene get <KEY>`" + `
- List secrets: ` + "`tene list`" + `
- Run with secrets injected: ` + "`tene run -- <command>`" + `
- Set a secret: ` + "`tene set <KEY> <VALUE>`" + `

## Rules
- Never hardcode secret values in source code
- Access secrets via environment variables
- Do not create .env files -- use ` + "`tene run`" + ` instead
- Use ` + "`tene list`" + ` to see available secrets
`
```

### 7.2 생성/병합 로직

```go
// generator.go

package claudemd

import (
    "os"
    "path/filepath"
    "strings"
)

// Generator는 CLAUDE.md를 생성하고 관리한다.
type Generator struct {
    projectDir string
}

// NewGenerator는 Generator를 생성한다.
func NewGenerator(projectDir string) *Generator {
    return &Generator{projectDir: projectDir}
}

// Generate는 CLAUDE.md를 생성하거나 기존 파일에 tene 섹션을 추가한다.
// 반환: 생성된 경우 true, 스킵된 경우 false
func (g *Generator) Generate() (bool, error) {
    path := filepath.Join(g.projectDir, "CLAUDE.md")

    content, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            // 새 파일 생성
            return true, os.WriteFile(path, []byte(SecretsMdTemplate), 0644)
        }
        return false, err
    }

    // 이미 tene 섹션이 있으면 스킵
    if g.HasTeneSection(string(content)) {
        return false, nil
    }

    // 기존 파일 끝에 섹션 추가
    separator := "\n\n"
    if !strings.HasSuffix(string(content), "\n") {
        separator = "\n\n"
    }
    updated := string(content) + separator + SecretsMdTemplate
    return true, os.WriteFile(path, []byte(updated), 0644)
}

// HasTeneSection은 파일 내용에 tene Secrets Management 섹션이 있는지 확인한다.
func (g *Generator) HasTeneSection(content string) bool {
    return strings.Contains(content, SectionHeader) ||
        strings.Contains(content, "tene") && strings.Contains(content, "secret management")
}

// FilePath는 CLAUDE.md의 절대 경로를 반환한다.
func (g *Generator) FilePath() string {
    return filepath.Join(g.projectDir, "CLAUDE.md")
}
```

---

## 8. internal/cli 상세 설계

### 8.1 Cobra 명령어 트리

```
tene (rootCmd)
├── init       프로젝트 초기화
├── set        시크릿 저장
├── get        시크릿 조회
├── run        시크릿 주입 후 명령 실행
├── list       시크릿 목록
├── delete     시크릿 삭제
├── import     .env 파일 가져오기
├── export     시크릿 내보내기
├── env        환경 관리
├── passwd     마스터 패스워드 변경
├── recover    Recovery Key로 복구
├── sync       Cloud waitlist (Fake Door)
└── whoami     현재 프로젝트 정보
```

### 8.2 App 구조체 (의존성 주입)

```go
// root.go

package cli

import (
    "github.com/spf13/cobra"
    "github.com/tomo-kay/tene/internal/crypto"
    "github.com/tomo-kay/tene/internal/keychain"
    "github.com/tomo-kay/tene/internal/vault"
)

var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

// SetVersion은 빌드 시 주입된 버전 정보를 설정한다.
func SetVersion(v, c, d string) {
    version = v
    commit = c
    date = d
}

// App은 CLI 실행에 필요한 의존성을 보유한다.
type App struct {
    Vault    *vault.Vault
    Keychain keychain.KeyStore
    Dir      string // 프로젝트 디렉토리
    Env      string // 활성 환경
    JSON     bool   // --json 플래그
}

// 글로벌 플래그
var (
    flagJSON bool
    flagEnv  string
    flagDir  string
)

var rootCmd = &cobra.Command{
    Use:     "tene",
    Short:   "Agentic Secret Runtime -- local-first encrypted secret management",
    Version: version,
}

func init() {
    rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output in JSON format")
    rootCmd.PersistentFlags().StringVar(&flagEnv, "env", "", "Environment name (default: active environment)")
    rootCmd.PersistentFlags().StringVar(&flagDir, "dir", "", "Project directory (default: current directory)")

    // 명령어 등록
    rootCmd.AddCommand(initCmd)
    rootCmd.AddCommand(setCmd)
    rootCmd.AddCommand(getCmd)
    rootCmd.AddCommand(runCmd)
    rootCmd.AddCommand(listCmd)
    rootCmd.AddCommand(deleteCmd)
    rootCmd.AddCommand(importCmd)
    rootCmd.AddCommand(exportCmd)
    rootCmd.AddCommand(envCmd)
    rootCmd.AddCommand(passwdCmd)
    rootCmd.AddCommand(recoverCmd)
    rootCmd.AddCommand(syncCmd)
    rootCmd.AddCommand(whoamiCmd)
}

// Execute는 루트 명령어를 실행한다.
func Execute() error {
    return rootCmd.Execute()
}
```

### 8.3 각 명령어의 RunE 함수 로직

#### tene init

```go
// init.go

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize a new tene vault in the current project",
    RunE: func(cmd *cobra.Command, args []string) error {
        dir := resolveDir()

        // 1. 이미 초기화되었는지 확인
        vaultPath := filepath.Join(dir, ".tene", "vault.db")
        if fileExists(vaultPath) {
            return fmt.Errorf("vault already initialized at %s", vaultPath)
        }

        // 2. Master Password 입력 (2회 확인)
        password, err := promptPasswordConfirm("Enter master password: ")
        if err != nil {
            return err
        }

        // 3. Salt 생성 + KDF -> Master Key
        salt, err := crypto.GenerateSalt()
        if err != nil {
            return err
        }
        masterKey, err := crypto.DeriveKey(password, salt)
        if err != nil {
            return err
        }

        // 4. OS Keychain에 Master Key 저장
        ks := keychain.NewStore(dir)
        if err := ks.Store(masterKey); err != nil {
            return err
        }

        // 5. SQLite 볼트 생성
        os.MkdirAll(filepath.Join(dir, ".tene"), 0700)
        v, err := vault.New(vaultPath)
        if err != nil {
            return err
        }
        defer v.Close()

        // 6. 메타데이터 저장
        v.SetMeta("schema_version", "1")
        v.SetMeta("created_at", time.Now().UTC().Format(time.RFC3339))
        v.SetMeta("kdf_salt", base64.StdEncoding.EncodeToString(salt))

        // 7. Recovery Key 생성
        mnemonic, err := recovery.GenerateMnemonic()
        if err != nil {
            return err
        }
        blob, err := recovery.EncryptMasterKey(masterKey, mnemonic)
        if err != nil {
            return err
        }
        v.SetMeta("recovery_blob", base64.StdEncoding.EncodeToString(blob))

        // 8. 기본 환경 생성
        v.SetActiveEnvironment("default")

        // 9. .gitignore 생성
        writeGitignore(filepath.Join(dir, ".tene", ".gitignore"))

        // 10. CLAUDE.md 생성
        gen := claudemd.NewGenerator(dir)
        created, _ := gen.Generate()

        // 11. 감사 로그
        v.AddAuditLog("vault.init", "", "")

        // 12. 결과 출력
        fmt.Println("Vault created at .tene/vault.db")
        if created {
            fmt.Println("CLAUDE.md generated (Claude Code will auto-detect tene)")
        }
        fmt.Println()
        fmt.Println("Recovery Key (write this down and store safely):")
        fmt.Printf("  %s\n", mnemonic)
        fmt.Println()
        fmt.Println("WARNING: This is the ONLY way to recover your vault if you forget your password.")
        fmt.Println("         Store it in a safe place. It will NOT be shown again.")

        return nil
    },
}
```

#### tene set KEY VALUE

```go
// set.go

var setCmd = &cobra.Command{
    Use:   "set KEY VALUE",
    Short: "Store an encrypted secret",
    Args:  cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
        keyName := args[0]
        value := args[1]

        app, err := loadApp()
        if err != nil {
            return err
        }
        defer app.Vault.Close()

        // Master Key 로드 (Keychain -> 없으면 패스워드 요청)
        masterKey, err := loadOrPromptMasterKey(app)
        if err != nil {
            return err
        }

        // Encryption Key 파생
        encKey, err := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)
        if err != nil {
            return err
        }

        // 암호화
        ciphertext, err := crypto.Encrypt(encKey, []byte(value), []byte(keyName))
        if err != nil {
            return err
        }

        // 저장
        encoded := base64.StdEncoding.EncodeToString(ciphertext)
        env := resolveEnv(app)
        if err := app.Vault.SetSecret(keyName, encoded, env); err != nil {
            return err
        }

        if flagJSON {
            return printJSON(map[string]string{"status": "ok", "key": keyName})
        }
        fmt.Printf("Secret '%s' stored.\n", keyName)
        return nil
    },
}
```

#### tene get KEY

```go
// get.go

var getCmd = &cobra.Command{
    Use:   "get KEY",
    Short: "Retrieve a decrypted secret",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        keyName := args[0]

        app, err := loadApp()
        if err != nil {
            return err
        }
        defer app.Vault.Close()

        masterKey, err := loadOrPromptMasterKey(app)
        if err != nil {
            return err
        }

        encKey, err := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)
        if err != nil {
            return err
        }

        env := resolveEnv(app)
        secret, err := app.Vault.GetSecret(keyName, env)
        if err != nil {
            return err
        }

        ciphertext, err := base64.StdEncoding.DecodeString(secret.EncryptedValue)
        if err != nil {
            return err
        }

        plaintext, err := crypto.Decrypt(encKey, ciphertext, []byte(keyName))
        if err != nil {
            return err
        }

        // 감사 로그
        app.Vault.AddAuditLog("secret.read", keyName, "")

        if flagJSON {
            return printJSON(map[string]string{
                "status": "ok",
                "key":    keyName,
                "value":  string(plaintext),
            })
        }
        fmt.Print(string(plaintext))
        return nil
    },
}
```

#### tene run -- CMD

```go
// run.go

var runCmd = &cobra.Command{
    Use:                "run -- COMMAND [ARGS...]",
    Short:              "Run a command with secrets injected as environment variables",
    DisableFlagParsing: true,
    RunE: func(cmd *cobra.Command, args []string) error {
        // "--" 이후의 인자 파싱
        cmdArgs := extractArgsAfterDash(args)
        if len(cmdArgs) == 0 {
            return fmt.Errorf("usage: tene run -- <command> [args...]")
        }

        app, err := loadApp()
        if err != nil {
            return err
        }
        defer app.Vault.Close()

        masterKey, err := loadOrPromptMasterKey(app)
        if err != nil {
            return err
        }

        encKey, err := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)
        if err != nil {
            return err
        }

        env := resolveEnv(app)
        allSecrets, err := app.Vault.GetAllSecrets(env)
        if err != nil {
            return err
        }

        // 모든 시크릿 복호화 + 환경변수 구성
        environ := os.Environ()
        for name, encVal := range allSecrets {
            ct, _ := base64.StdEncoding.DecodeString(encVal)
            pt, err := crypto.Decrypt(encKey, ct, []byte(name))
            if err != nil {
                return fmt.Errorf("failed to decrypt %s: %w", name, err)
            }
            environ = append(environ, fmt.Sprintf("%s=%s", name, string(pt)))
        }

        // 명령 실행
        c := exec.Command(cmdArgs[0], cmdArgs[1:]...)
        c.Env = environ
        c.Stdin = os.Stdin
        c.Stdout = os.Stdout
        c.Stderr = os.Stderr
        return c.Run()
    },
}
```

#### tene list

```go
// list.go

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List all secrets (values masked)",
    RunE: func(cmd *cobra.Command, args []string) error {
        app, err := loadApp()
        if err != nil {
            return err
        }
        defer app.Vault.Close()

        env := resolveEnv(app)
        secrets, err := app.Vault.ListSecrets(env)
        if err != nil {
            return err
        }

        if flagJSON {
            type item struct {
                Name        string `json:"name"`
                Environment string `json:"environment"`
                Version     int    `json:"version"`
                UpdatedAt   string `json:"updated_at"`
            }
            items := make([]item, len(secrets))
            for i, s := range secrets {
                items[i] = item{s.Name, s.Environment, s.Version, s.UpdatedAt.Format(time.RFC3339)}
            }
            return printJSON(map[string]any{"status": "ok", "data": map[string]any{"secrets": items}})
        }

        if len(secrets) == 0 {
            fmt.Printf("No secrets in environment '%s'.\n", env)
            return nil
        }

        fmt.Printf("Secrets in '%s' (%d):\n\n", env, len(secrets))
        fmt.Printf("  %-30s %-10s %s\n", "NAME", "VERSION", "UPDATED")
        fmt.Printf("  %-30s %-10s %s\n", "----", "-------", "-------")
        for _, s := range secrets {
            fmt.Printf("  %-30s v%-9d %s\n", s.Name, s.Version, s.UpdatedAt.Format("2006-01-02 15:04"))
        }
        return nil
    },
}
```

#### tene sync (Fake Door)

```go
// sync_cmd.go

var syncCmd = &cobra.Command{
    Use:   "sync",
    Short: "Sync vault to cloud (coming soon)",
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Println(`
  Tene Cloud Sync -- Coming Soon!

  Cloud sync will enable:
  - Multi-device secret synchronization
  - Encrypted cloud backup (zero-knowledge)
  - Web dashboard for secret overview
  - All for just $1/month

  Join the waitlist to get early access:
  -> https://tene.dev/waitlist

  In the meantime, use 'tene export --encrypted' for local backup.`)

        // TODO: Analytics - tene sync 실행 횟수 추적
        return nil
    },
}
```

### 8.4 공통 헬퍼 함수

```go
// helpers.go

package cli

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "golang.org/x/term"
)

// resolveDir는 프로젝트 디렉토리를 결정한다.
func resolveDir() string {
    if flagDir != "" {
        return flagDir
    }
    dir, _ := os.Getwd()
    return dir
}

// resolveEnv는 활성 환경을 결정한다.
func resolveEnv(app *App) string {
    if flagEnv != "" {
        return flagEnv
    }
    if app.Env != "" {
        return app.Env
    }
    return "default"
}

// loadApp은 볼트와 키체인을 로드하여 App을 구성한다.
func loadApp() (*App, error) {
    dir := resolveDir()
    vaultPath := filepath.Join(dir, ".tene", "vault.db")

    if !fileExists(vaultPath) {
        return nil, fmt.Errorf("vault not found at %s\nRun 'tene init' first", vaultPath)
    }

    v, err := vault.New(vaultPath)
    if err != nil {
        return nil, err
    }

    activeEnv, _ := v.GetActiveEnvironment()

    return &App{
        Vault:    v,
        Keychain: keychain.NewStore(dir),
        Dir:      dir,
        Env:      activeEnv,
        JSON:     flagJSON,
    }, nil
}

// loadOrPromptMasterKey는 Keychain에서 Master Key를 로드한다.
// 없으면 패스워드를 입력받아 KDF를 수행한다.
func loadOrPromptMasterKey(app *App) ([]byte, error) {
    // 1. Keychain에서 로드 시도
    key, err := app.Keychain.Load()
    if err == nil {
        return key, nil
    }

    // 2. 패스워드 입력
    password, err := promptPassword("Enter master password: ")
    if err != nil {
        return nil, err
    }

    // 3. Salt 로드
    saltB64, err := app.Vault.GetMeta("kdf_salt")
    if err != nil {
        return nil, fmt.Errorf("vault corrupted: missing kdf_salt")
    }
    salt, err := base64.StdEncoding.DecodeString(saltB64)
    if err != nil {
        return nil, err
    }

    // 4. KDF
    masterKey, err := crypto.DeriveKey(password, salt)
    if err != nil {
        return nil, err
    }

    // 5. Keychain에 캐시
    app.Keychain.Store(masterKey)

    return masterKey, nil
}

// promptPassword는 터미널에서 패스워드를 입력받는다 (에코 숨김).
func promptPassword(prompt string) (string, error) {
    fmt.Fprint(os.Stderr, prompt)
    password, err := term.ReadPassword(int(os.Stdin.Fd()))
    fmt.Fprintln(os.Stderr) // 개행
    if err != nil {
        return "", fmt.Errorf("failed to read password: %w", err)
    }
    return string(password), nil
}

// promptPasswordConfirm은 패스워드를 2회 입력받아 일치 확인한다.
func promptPasswordConfirm(prompt string) (string, error) {
    pw1, err := promptPassword(prompt)
    if err != nil {
        return "", err
    }
    pw2, err := promptPassword("Confirm password: ")
    if err != nil {
        return "", err
    }
    if pw1 != pw2 {
        return "", fmt.Errorf("passwords do not match")
    }
    if len(pw1) < 8 {
        return "", fmt.Errorf("password must be at least 8 characters")
    }
    return pw1, nil
}

// printJSON은 데이터를 JSON 형식으로 stdout에 출력한다.
func printJSON(data any) error {
    enc := json.NewEncoder(os.Stdout)
    enc.SetIndent("", "  ")
    return enc.Encode(data)
}

// fileExists는 파일 존재 여부를 반환한다.
func fileExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil
}
```

### 8.5 플래그 정의 요약

| 명령어 | 플래그 | 타입 | 설명 |
|--------|--------|------|------|
| 글로벌 | `--json` | bool | JSON 출력 |
| 글로벌 | `--env` | string | 환경 지정 |
| 글로벌 | `--dir` | string | 프로젝트 디렉토리 |
| `export` | `--encrypted` | bool | 암호화된 볼트 백업 |
| `import` | `--encrypted` | bool | 암호화된 백업에서 복원 |
| `delete` | `--force` | bool | 확인 없이 삭제 |
| `env` | `--create` | bool | 새 환경 생성 |

### 8.6 대화형/비대화형 분기

```go
// helpers.go

// isInteractive는 stdin이 터미널인지 확인한다.
func isInteractive() bool {
    return term.IsTerminal(int(os.Stdin.Fd()))
}
```

| 상황 | 대화형 (터미널) | 비대화형 (파이프/AI Agent) |
|------|:-------------:|:-------------------:|
| 패스워드 입력 | term.ReadPassword | TENE_PASSWORD 환경변수 |
| 삭제 확인 | Y/N 프롬프트 | --force 필수 |
| 출력 형식 | 사람 읽기 좋은 텍스트 | --json 권장 |

---

## 9. 에러 처리 설계

### 9.1 에러 계층

```
CLI 에러 출력 (human-readable 또는 JSON)
    |
    +--- cli/helpers.go: handleError(err)
         |
         +--- errors.Is() 로 sentinel 에러 매칭
         |
         +--- 패키지별 sentinel 에러:
              crypto.ErrDecryptionFailed
              vault.ErrSecretNotFound
              vault.ErrVaultNotInitialized
              keychain.ErrKeyNotFound
              recovery.ErrInvalidMnemonic
```

### 9.2 종료 코드 매핑

```go
// helpers.go

func handleError(err error) {
    code := 1 // 기본 에러

    switch {
    case errors.Is(err, crypto.ErrDecryptionFailed):
        code = 2
        fmt.Fprintln(os.Stderr, "Error: Wrong master password or corrupted data")
    case errors.Is(err, vault.ErrSecretNotFound):
        code = 3
        fmt.Fprintln(os.Stderr, "Error:", err.Error())
    case errors.Is(err, vault.ErrVaultNotInitialized):
        code = 4
        fmt.Fprintln(os.Stderr, "Error: Vault not initialized. Run 'tene init' first.")
    default:
        fmt.Fprintln(os.Stderr, "Error:", err.Error())
    }

    os.Exit(code)
}
```

| 종료 코드 | 의미 | 대응 에러 |
|:--------:|------|----------|
| 0 | 성공 | - |
| 1 | 일반 에러 | 기타 |
| 2 | 인증 실패 | `crypto.ErrDecryptionFailed` |
| 3 | 시크릿 미발견 | `vault.ErrSecretNotFound` |
| 4 | 볼트 미초기화 | `vault.ErrVaultNotInitialized` |

### 9.3 --json 에러 출력

```go
func handleErrorJSON(err error) {
    code := 1
    msg := err.Error()

    switch {
    case errors.Is(err, crypto.ErrDecryptionFailed):
        code = 2
        msg = "wrong master password or corrupted data"
    case errors.Is(err, vault.ErrSecretNotFound):
        code = 3
    case errors.Is(err, vault.ErrVaultNotInitialized):
        code = 4
        msg = "vault not initialized, run 'tene init' first"
    }

    json.NewEncoder(os.Stderr).Encode(map[string]any{
        "status":  "error",
        "code":    code,
        "message": msg,
    })
    os.Exit(code)
}
```

---

## 10. 테스트 설계

### 10.1 테스트 헬퍼/Fixture

```go
// internal/crypto/crypto_test.go

func TestEncrypt_Decrypt_RoundTrip(t *testing.T) {
    key := make([]byte, 32)
    rand.Read(key)

    tests := []struct {
        name      string
        plaintext string
        aad       string
    }{
        {"simple", "hello", "KEY1"},
        {"empty", "", "KEY2"},
        {"long", strings.Repeat("x", 10000), "KEY3"},
        {"unicode", "secret-value", "KEY4"},
        {"special chars", "p@$$w0rd!#%", "SPECIAL_KEY"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ct, err := Encrypt(key, []byte(tt.plaintext), []byte(tt.aad))
            require.NoError(t, err)
            assert.Greater(t, len(ct), NonceLen)

            pt, err := Decrypt(key, ct, []byte(tt.aad))
            require.NoError(t, err)
            assert.Equal(t, tt.plaintext, string(pt))
        })
    }
}

func TestDecrypt_WrongKey(t *testing.T) {
    key1 := make([]byte, 32)
    key2 := make([]byte, 32)
    rand.Read(key1)
    rand.Read(key2)

    ct, err := Encrypt(key1, []byte("secret"), []byte("KEY"))
    require.NoError(t, err)

    _, err = Decrypt(key2, ct, []byte("KEY"))
    assert.ErrorIs(t, err, ErrDecryptionFailed)
}

func TestDecrypt_WrongAAD(t *testing.T) {
    key := make([]byte, 32)
    rand.Read(key)

    ct, err := Encrypt(key, []byte("secret"), []byte("KEY1"))
    require.NoError(t, err)

    _, err = Decrypt(key, ct, []byte("KEY2"))
    assert.ErrorIs(t, err, ErrDecryptionFailed)
}
```

### 10.2 Vault 테스트 헬퍼

```go
// internal/vault/vault_test.go

func setupTestVault(t *testing.T) *Vault {
    t.Helper()
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "vault.db")
    v, err := New(dbPath)
    require.NoError(t, err)
    t.Cleanup(func() { v.Close() })
    return v
}

func TestVault_CRUD(t *testing.T) {
    v := setupTestVault(t)

    // Set
    err := v.SetSecret("API_KEY", "encrypted-value", "default")
    require.NoError(t, err)

    // Get
    s, err := v.GetSecret("API_KEY", "default")
    require.NoError(t, err)
    assert.Equal(t, "API_KEY", s.Name)
    assert.Equal(t, "encrypted-value", s.EncryptedValue)
    assert.Equal(t, 1, s.Version)

    // Update (UPSERT)
    err = v.SetSecret("API_KEY", "new-encrypted", "default")
    require.NoError(t, err)
    s, _ = v.GetSecret("API_KEY", "default")
    assert.Equal(t, "new-encrypted", s.EncryptedValue)
    assert.Equal(t, 2, s.Version) // version 증가

    // List
    secrets, err := v.ListSecrets("default")
    require.NoError(t, err)
    assert.Len(t, secrets, 1)

    // Delete
    err = v.DeleteSecret("API_KEY", "default")
    require.NoError(t, err)
    _, err = v.GetSecret("API_KEY", "default")
    assert.ErrorIs(t, err, ErrSecretNotFound)
}

func TestVault_EnvironmentIsolation(t *testing.T) {
    v := setupTestVault(t)

    v.SetSecret("KEY", "dev-value", "dev")
    v.SetSecret("KEY", "prod-value", "prod")

    devSecret, _ := v.GetSecret("KEY", "dev")
    prodSecret, _ := v.GetSecret("KEY", "prod")

    assert.Equal(t, "dev-value", devSecret.EncryptedValue)
    assert.Equal(t, "prod-value", prodSecret.EncryptedValue)
}
```

### 10.3 Keychain 모킹 전략

```go
// 테스트에서 KeyStore 인터페이스를 모킹

type MockKeyStore struct {
    stored []byte
    err    error
}

func (m *MockKeyStore) Store(key []byte) error {
    if m.err != nil {
        return m.err
    }
    m.stored = make([]byte, len(key))
    copy(m.stored, key)
    return nil
}

func (m *MockKeyStore) Load() ([]byte, error) {
    if m.stored == nil {
        return nil, keychain.ErrKeyNotFound
    }
    return m.stored, nil
}

func (m *MockKeyStore) Delete() error {
    m.stored = nil
    return nil
}

func (m *MockKeyStore) Exists() bool {
    return m.stored != nil
}
```

### 10.4 CLI 통합 테스트 시나리오

```go
// internal/cli/cli_test.go

func TestCLI_FullFlow(t *testing.T) {
    // 임시 디렉토리에서 전체 플로우 테스트
    dir := t.TempDir()

    // 환경변수로 패스워드 주입 (비대화형)
    os.Setenv("TENE_PASSWORD", "test-password-123")
    os.Setenv("TENE_KEYCHAIN_FALLBACK", "file")
    defer os.Unsetenv("TENE_PASSWORD")
    defer os.Unsetenv("TENE_KEYCHAIN_FALLBACK")

    // 1. init
    runCLI(t, dir, "init")
    assert.FileExists(t, filepath.Join(dir, ".tene", "vault.db"))
    assert.FileExists(t, filepath.Join(dir, "CLAUDE.md"))

    // 2. set
    runCLI(t, dir, "set", "API_KEY", "sk_test_123")

    // 3. get
    output := runCLI(t, dir, "get", "API_KEY")
    assert.Equal(t, "sk_test_123", output)

    // 4. get --json
    jsonOutput := runCLI(t, dir, "get", "--json", "API_KEY")
    assert.Contains(t, jsonOutput, `"value":"sk_test_123"`)

    // 5. list
    listOutput := runCLI(t, dir, "list")
    assert.Contains(t, listOutput, "API_KEY")

    // 6. delete
    runCLI(t, dir, "delete", "--force", "API_KEY")
    _, err := runCLIErr(t, dir, "get", "API_KEY")
    assert.Error(t, err)
}
```

### 10.5 통합 테스트 실행 방법

```bash
# 모든 테스트 (단위 + 통합)
go test -race ./...

# crypto 패키지만 (보안 중점)
go test -race -count=5 ./internal/crypto/...

# CLI 통합 테스트 (환경변수로 비대화형)
TENE_PASSWORD=test TENE_KEYCHAIN_FALLBACK=file go test -v ./internal/cli/...
```

---

## 11. 구현 가이드

### 11.1 구현 순서 체크리스트

- [ ] Go 프로젝트 초기화 (go.mod, go.sum, .gitignore, Makefile)
- [ ] internal/crypto: kdf.go, encrypt.go, decrypt.go, keymanager.go, errors.go, crypto_test.go
- [ ] internal/recovery: mnemonic.go, recover.go, errors.go, mnemonic_test.go
- [ ] internal/vault: models.go, schema.go, vault.go, migration.go, errors.go, vault_test.go
- [ ] internal/keychain: keychain.go, fallback.go, errors.go, keychain_test.go
- [ ] internal/claudemd: template.go, generator.go, generator_test.go
- [ ] internal/cli: root.go, helpers.go
- [ ] internal/cli: init.go, set.go, get.go (핵심 3개)
- [ ] internal/cli: run.go, list.go, delete.go
- [ ] internal/cli: import_cmd.go, export.go, env.go
- [ ] internal/cli: passwd.go, recover.go, sync_cmd.go, whoami.go
- [ ] internal/cli: cli_test.go (통합 테스트)
- [ ] cmd/tene/main.go
- [ ] .goreleaser.yml
- [ ] .github/workflows/ci.yml
- [ ] .github/workflows/release.yml
- [ ] .golangci.yml
- [ ] install.sh

### 11.2 핵심 파일 목록

| 우선순위 | 패키지 | 파일 수 | 예상 LOC |
|:--------:|--------|:------:|:--------:|
| 1 | internal/crypto | 6 | ~300 |
| 2 | internal/recovery | 4 | ~150 |
| 3 | internal/vault | 6 | ~400 |
| 4 | internal/keychain | 4 | ~200 |
| 5 | internal/claudemd | 3 | ~100 |
| 6 | internal/cli | 16 | ~800 |
| 7 | cmd/tene | 1 | ~20 |
| 8 | 빌드/CI | 5 | ~200 |
| **합계** | | **45** | **~2,170** |

### 11.3 Session Guide

| Module | 범위 | 예상 시간 | 의존성 |
|--------|------|:--------:|--------|
| module-1 | go.mod + internal/crypto | 2-3h | 없음 |
| module-2 | internal/recovery | 1-2h | module-1 |
| module-3 | internal/vault | 2-3h | 없음 (module-1과 병렬 가능) |
| module-4 | internal/keychain | 1-2h | 없음 |
| module-5 | internal/claudemd | 1h | 없음 |
| module-6 | internal/cli (핵심: init, set, get, run, list, delete) | 3-4h | module-1~5 |
| module-7 | internal/cli (나머지: import, export, env, passwd, recover, sync, whoami) | 2-3h | module-6 |
| module-8 | cmd/tene + goreleaser + CI/CD | 1-2h | module-7 |

**권장 세션 플랜**:
- Session 1: module-1 + module-3 (핵심 인프라, 병렬)
- Session 2: module-2 + module-4 + module-5 (보조 패키지)
- Session 3: module-6 (핵심 CLI 명령어)
- Session 4: module-7 + module-8 (나머지 CLI + 배포)
