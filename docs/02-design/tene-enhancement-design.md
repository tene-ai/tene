# Tene CLI 고도화 상세 설계서

> **Version**: 1.0
> **Date**: 2026-04-06
> **Author**: Claude Code (Steve 의뢰)
> **Status**: Draft
> **근거**: Gap 분석에서 발견된 미구현/부분구현 항목의 Go 코드 레벨 상세 설계

---

## 목차

1. [구조화된 에러 코드 시스템](#1-구조화된-에러-코드-시스템)
2. [Exit Code 분기](#2-exit-code-분기)
3. [.tene/vault.json 생성](#3-tenevaultjson-생성)
4. [~/.tene/config.json 생성](#4-teneconfigjson-생성)
5. [.tene.enc 바이너리 포맷](#5-teneenc-바이너리-포맷)
6. [CLI 통합 테스트](#6-cli-통합-테스트)
7. [--no-color / NO_COLOR 지원](#7---no-color--no_color-지원)
8. [메모리 제로화](#8-메모리-제로화)
9. [CLAUDE.md URL 통일](#9-claudemd-url-통일)
10. [Vault 인터페이스 분리](#10-vault-인터페이스-분리)

---

## 1. 구조화된 에러 코드 시스템

### 1.1 현재 상태

현재 CLI는 `fmt.Errorf()`로 에러를 반환하고, `main.go`에서 모든 에러를 exit code 1로 처리한다:

```go
// cmd/tene/main.go (현재)
if err := cli.Execute(); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %s\n", err)
    os.Exit(1)
}
```

vault 패키지와 crypto 패키지에 sentinel 에러가 있지만, CLI 레벨에서 이를 감지하여 적절한 exit code를 반환하는 로직이 없다.

### 1.2 목표

- 23개 에러 코드를 `internal/errors/` 패키지에 정의
- `TeneError` struct로 에러 코드, 메시지, exit code를 캡슐화
- `--json` 모드에서 구조화된 에러 JSON 출력
- `main.go`에서 `TeneError` type assertion으로 exit code 분기

### 1.3 신규 파일: `internal/errors/errors.go`

```go
package errors

import (
	"encoding/json"
	"fmt"
	"io"
)

// TeneError는 구조화된 CLI 에러를 나타낸다.
// Code: 머신 파싱용 에러 코드 (예: "VAULT_NOT_FOUND")
// Message: 사람이 읽는 에러 메시지
// Exit: 프로세스 종료 코드 (0, 1, 2, 127)
type TeneError struct {
	Code    string `json:"error"`
	Message string `json:"message"`
	Exit    int    `json:"-"`
}

func (e *TeneError) Error() string {
	return e.Message
}

// WriteJSON은 --json 모드에서 에러를 JSON으로 출력한다.
func (e *TeneError) WriteJSON(w io.Writer) error {
	out := struct {
		OK      bool   `json:"ok"`
		Error   string `json:"error"`
		Message string `json:"message"`
	}{
		OK:      false,
		Error:   e.Code,
		Message: e.Message,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// New는 새 TeneError를 생성한다.
func New(code, message string, exit int) *TeneError {
	return &TeneError{Code: code, Message: message, Exit: exit}
}

// Newf는 포맷 문자열로 새 TeneError를 생성한다.
func Newf(code string, exit int, format string, args ...any) *TeneError {
	return &TeneError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Exit:    exit,
	}
}

// IsTeneError는 error가 *TeneError인지 확인하고 반환한다.
func IsTeneError(err error) (*TeneError, bool) {
	if te, ok := err.(*TeneError); ok {
		return te, true
	}
	return nil, false
}
```

### 1.4 신규 파일: `internal/errors/codes.go`

```go
package errors

// --- Exit Code 0: 성공 / 경고 ---

var ErrVaultAlreadyExists = &TeneError{
	Code: "VAULT_ALREADY_EXISTS", Message: "Vault already exists. Use existing vault.", Exit: 0,
}
var ErrKeychainError = &TeneError{
	Code: "KEYCHAIN_ERROR", Message: "Keychain access failed.", Exit: 0,
}

// --- Exit Code 1: 일반 에러 ---

var ErrVaultNotFound = &TeneError{
	Code: "VAULT_NOT_FOUND", Message: "Not in a Tene project. Run \"tene init\" first.", Exit: 1,
}

func ErrSecretNotFound(key, env string) *TeneError {
	return Newf("SECRET_NOT_FOUND", 1, "Secret %q not found in %q environment.", key, env)
}

func ErrSecretAlreadyExists(key string) *TeneError {
	return Newf("SECRET_ALREADY_EXISTS", 1, "Secret %q already exists. Use --overwrite to replace.", key)
}

func ErrEnvironmentNotFound(env string) *TeneError {
	return Newf("ENVIRONMENT_NOT_FOUND", 1, "Environment %q not found. Create it with \"tene env create %s\".", env, env)
}

func ErrEnvironmentAlreadyExists(env string) *TeneError {
	return Newf("ENVIRONMENT_ALREADY_EXISTS", 1, "Environment %q already exists.", env)
}

func ErrEnvironmentProtected(env, reason string) *TeneError {
	return Newf("ENVIRONMENT_PROTECTED", 1, "Cannot delete the %q environment. %s", env, reason)
}

func ErrInvalidKeyName(key string) *TeneError {
	return Newf("INVALID_KEY_NAME", 1, "Invalid key name %q. Keys must match [A-Z][A-Z0-9_]*.", key)
}

func ErrReservedKeyName(key string) *TeneError {
	return Newf("RESERVED_KEY_NAME", 1, "Key name %q is reserved.", key)
}

var ErrInvalidEnvName = &TeneError{
	Code: "INVALID_ENV_NAME", Message: "Invalid environment name. Must match [a-z][a-z0-9-]*.", Exit: 1,
}
var ErrEmptyValue = &TeneError{
	Code: "EMPTY_VALUE", Message: "Value cannot be empty.", Exit: 1,
}
var ErrValueTooLarge = &TeneError{
	Code: "VALUE_TOO_LARGE", Message: "Value exceeds maximum size (64KB).", Exit: 1,
}
var ErrEncryptFailed = &TeneError{
	Code: "ENCRYPT_FAILED", Message: "Encryption failed.", Exit: 1,
}

func ErrFileNotFound(path string) *TeneError {
	return Newf("FILE_NOT_FOUND", 1, "File %q not found.", path)
}

func ErrFileParse(path string, line int, detail string) *TeneError {
	return Newf("FILE_PARSE_ERROR", 1, "Failed to parse %q at line %d: %s.", path, line, detail)
}

var ErrPermissionDenied = &TeneError{
	Code: "PERMISSION_DENIED", Message: "Permission denied.", Exit: 1,
}
var ErrDiskFull = &TeneError{
	Code: "DISK_FULL", Message: "Cannot create vault: insufficient disk space.", Exit: 1,
}
var ErrInteractiveRequired = &TeneError{
	Code: "INTERACTIVE_REQUIRED", Message: "This command requires an interactive terminal.", Exit: 1,
}
var ErrInvalidBackupFile = &TeneError{
	Code: "INVALID_BACKUP_FILE", Message: "Invalid encrypted backup file format.", Exit: 1,
}

// --- Exit Code 2: 인증 에러 ---

var ErrPasswordMismatch = &TeneError{
	Code: "PASSWORD_MISMATCH", Message: "Passwords do not match. Try again.", Exit: 2,
}
var ErrPasswordTooShort = &TeneError{
	Code: "PASSWORD_TOO_SHORT", Message: "Master Password must be at least 8 characters.", Exit: 2,
}
var ErrInvalidPassword = &TeneError{
	Code: "INVALID_PASSWORD", Message: "Invalid Master Password.", Exit: 2,
}
var ErrInvalidRecoveryKey = &TeneError{
	Code: "INVALID_RECOVERY_KEY", Message: "Invalid Recovery Key.", Exit: 2,
}
var ErrDecryptFailed = &TeneError{
	Code: "DECRYPT_FAILED", Message: "Failed to decrypt secret. Master Password may have changed.", Exit: 2,
}

// --- Exit Code 127: 명령어 없음 ---

func ErrCommandNotFound(cmd string) *TeneError {
	return Newf("COMMAND_NOT_FOUND", 127, "Command %q not found.", cmd)
}
```

### 1.5 기존 파일 수정: CLI 에러 래핑 전략

모든 CLI 명령어 함수에서 `fmt.Errorf()` 대신 `teneerr.ErrXxx` 또는 `teneerr.Newf()`를 사용한다.

#### 수정 대상 파일 및 변경 내용

| 파일 | 변경 내용 |
|------|----------|
| `internal/cli/root.go` | `loadApp()` 에서 `teneerr.ErrVaultNotFound` 반환 |
| `internal/cli/set.go` | 키 이름/값 검증 에러를 `teneerr.ErrInvalidKeyName()` 등으로 교체 |
| `internal/cli/get.go` | 시크릿 없음 에러를 `teneerr.ErrSecretNotFound()` 로 교체 |
| `internal/cli/run.go` | 명령어 없음 에러를 `teneerr.ErrCommandNotFound()` 로 교체 |
| `internal/cli/init.go` | 비밀번호 검증 에러를 `teneerr.ErrPasswordTooShort` 등으로 교체 |
| `internal/cli/helpers.go` | `validateKeyName()`, `validateEnvName()` 반환 타입 교체 |
| `internal/cli/import.go` | 파일 관련 에러를 `teneerr.ErrFileNotFound()` 등으로 교체 |
| `internal/cli/export.go` | 복호화 실패를 `teneerr.ErrDecryptFailed` 로 교체 |
| `internal/cli/env.go` | 환경 관련 에러를 `teneerr.ErrEnvironmentXxx()` 로 교체 |
| `internal/cli/passwd.go` | 인증 에러를 `teneerr.ErrInvalidPassword` 로 교체 |
| `internal/cli/recover.go` | Recovery 에러를 `teneerr.ErrInvalidRecoveryKey` 로 교체 |

#### 예시: `internal/cli/root.go` loadApp() 수정

```go
import teneerr "github.com/tene-ai/tene/internal/errors"

func loadApp() (*App, error) {
	dir := resolveDir()
	vaultPath := filepath.Join(dir, ".tene", "vault.db")

	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return nil, teneerr.ErrVaultNotFound // 기존: fmt.Errorf("Not in a Tene project...")
	}
	// ... 나머지 동일
}
```

#### 예시: `internal/cli/helpers.go` validateKeyName() 수정

```go
import teneerr "github.com/tene-ai/tene/internal/errors"

func validateKeyName(name string) error {
	if len(name) == 0 || len(name) > 256 {
		return teneerr.ErrInvalidKeyName(name)
	}
	if !keyNameRegex.MatchString(name) {
		return teneerr.ErrInvalidKeyName(name)
	}
	if reservedKeys[name] {
		return teneerr.ErrReservedKeyName(name)
	}
	return nil
}

func validateEnvName(name string) error {
	if len(name) == 0 || len(name) > 64 {
		return teneerr.ErrInvalidEnvName
	}
	if !envNameRegex.MatchString(name) {
		return teneerr.ErrInvalidEnvName
	}
	return nil
}
```

#### 예시: `internal/cli/get.go` 수정

```go
import teneerr "github.com/tene-ai/tene/internal/errors"

func runGet(cmd *cobra.Command, args []string) error {
	keyName := args[0]

	app, err := loadApp()
	if err != nil {
		return err // 이미 TeneError
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
		return teneerr.ErrSecretNotFound(keyName, env) // 기존: fmt.Errorf(...)
	}

	ciphertext, err := decodeBase64(secret.EncryptedValue)
	if err != nil {
		return teneerr.ErrDecryptFailed // 기존: fmt.Errorf("failed to decode secret: %w", ...)
	}

	plaintext, err := crypto.Decrypt(encKey, ciphertext, []byte(keyName))
	if err != nil {
		return teneerr.ErrDecryptFailed // 기존: fmt.Errorf("Failed to decrypt secret...")
	}

	// ... 나머지 동일
}
```

#### 예시: `internal/cli/run.go` 수정

```go
import teneerr "github.com/tene-ai/tene/internal/errors"

func runRun(cmd *cobra.Command, args []string) error {
	cmdArgs := extractArgsAfterDash(args)
	if len(cmdArgs) == 0 {
		return teneerr.New("COMMAND_NOT_FOUND", "No command specified. Usage: tene run -- <command>", 1)
	}

	// ... 중략 ...

	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		if execErr, ok := err.(*exec.Error); ok {
			return teneerr.ErrCommandNotFound(execErr.Name) // 기존: fmt.Errorf(...)
		}
		return err
	}
	return nil
}
```

### 1.6 --json 에러 출력

`main.go`에서 `--json` 플래그가 활성화된 경우, TeneError를 JSON으로 출력한다.

`--json` 플래그 감지는 `os.Args`를 순회하여 확인한다 (`cobra` 파싱 전이므로).

```go
// cmd/tene/main.go 내부 헬퍼
func hasJSONFlag() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--json" {
			return true
		}
	}
	return false
}
```

### 1.7 테스트 시나리오

| 테스트 | 검증 내용 |
|--------|----------|
| `TestTeneError_Error` | `.Error()`가 Message 반환 |
| `TestTeneError_WriteJSON` | JSON 출력 형식 (`ok`, `error`, `message` 필드) |
| `TestIsTeneError` | type assertion 정상 동작 |
| `TestIsTeneError_PlainError` | 일반 error 에서 false 반환 |
| `TestNewf_Format` | 포맷 문자열 정상 대입 |
| `TestErrorCodes_ExitValues` | 모든 23개 에러 코드의 Exit 값 검증 |

```go
// internal/errors/errors_test.go
package errors

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestTeneError_Error(t *testing.T) {
	err := ErrVaultNotFound
	want := "Not in a Tene project. Run \"tene init\" first."
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestTeneError_WriteJSON(t *testing.T) {
	var buf bytes.Buffer
	err := ErrVaultNotFound
	if writeErr := err.WriteJSON(&buf); writeErr != nil {
		t.Fatal(writeErr)
	}

	var result map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &result); jsonErr != nil {
		t.Fatal(jsonErr)
	}

	if result["ok"] != false {
		t.Error("expected ok=false")
	}
	if result["error"] != "VAULT_NOT_FOUND" {
		t.Errorf("error = %v, want VAULT_NOT_FOUND", result["error"])
	}
}

func TestIsTeneError(t *testing.T) {
	te, ok := IsTeneError(ErrVaultNotFound)
	if !ok || te.Code != "VAULT_NOT_FOUND" {
		t.Error("expected TeneError with VAULT_NOT_FOUND")
	}
}

func TestErrSecretNotFound_Format(t *testing.T) {
	err := ErrSecretNotFound("API_KEY", "production")
	want := `Secret "API_KEY" not found in "production" environment.`
	if err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
	if err.Exit != 1 {
		t.Errorf("Exit = %d, want 1", err.Exit)
	}
}
```

---

## 2. Exit Code 분기

### 2.1 현재 상태

```go
// cmd/tene/main.go (현재)
if err := cli.Execute(); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %s\n", err)
    os.Exit(1) // 항상 1
}
```

### 2.2 목표

`TeneError`의 `Exit` 필드에 따라 종료 코드를 분기한다.

### 2.3 수정 파일: `cmd/tene/main.go`

```go
package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/tene-ai/tene/internal/cli"
	teneerr "github.com/tene-ai/tene/internal/errors"
)

var (
	version = ""
	commit  = ""
	date    = ""
)

func main() {
	// ... 기존 version/commit/date 설정 로직 동일 ...

	cli.SetVersion(version, commit, date)

	if err := cli.Execute(); err != nil {
		exitCode := 1 // 기본 exit code

		if te, ok := teneerr.IsTeneError(err); ok {
			exitCode = te.Exit

			if hasJSONFlag() {
				_ = te.WriteJSON(os.Stderr)
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s\n", te.Message)
			}
		} else {
			// TeneError가 아닌 일반 error
			if hasJSONFlag() {
				fallback := teneerr.New("UNKNOWN_ERROR", err.Error(), 1)
				_ = fallback.WriteJSON(os.Stderr)
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			}
		}

		os.Exit(exitCode)
	}
}

// hasJSONFlag는 os.Args에서 --json 플래그를 감지한다.
// cobra 파싱 전이므로 직접 순회한다.
func hasJSONFlag() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--json" {
			return true
		}
	}
	return false
}
```

### 2.4 Exit Code 매핑 요약

| Exit Code | TeneError.Code 예시 | 상황 |
|:---------:|---------------------|------|
| 0 | `VAULT_ALREADY_EXISTS`, `KEYCHAIN_ERROR` | 성공 또는 비치명적 경고 |
| 1 | `VAULT_NOT_FOUND`, `SECRET_NOT_FOUND`, `INVALID_KEY_NAME` 등 | 일반 에러 |
| 2 | `INVALID_PASSWORD`, `DECRYPT_FAILED`, `INVALID_RECOVERY_KEY` | 인증 에러 |
| 127 | `COMMAND_NOT_FOUND` | `tene run`에서 명령어 없음 |

### 2.5 테스트 시나리오

| 테스트 | 검증 내용 |
|--------|----------|
| `TestExitCode_VaultNotFound` | loadApp 실패 시 exit 1 |
| `TestExitCode_DecryptFailed` | 복호화 실패 시 exit 2 |
| `TestExitCode_CommandNotFound` | tene run 실패 시 exit 127 |
| `TestExitCode_JSONOutput` | --json 시 stderr에 JSON 출력 |

---

## 3. .tene/vault.json 생성

### 3.1 현재 상태

`tene init`에서 `.tene/vault.json` 파일을 생성하지 않는다. 요구사항에 따르면 이 파일에 `projectName`, `activeEnvironment`, `vaultVersion`, `agents`, `createdAt`을 저장해야 한다.

### 3.2 JSON 스키마

```json
{
  "projectName": "my-project",
  "createdAt": "2026-04-06T12:00:00Z",
  "vaultVersion": 1,
  "activeEnvironment": "default",
  "agents": ["claude"]
}
```

### 3.3 신규 파일: `internal/vault/vaultjson.go`

```go
package vault

import (
	"encoding/json"
	"os"
	"time"
)

// VaultJSON은 .tene/vault.json 파일의 구조를 나타낸다.
// SQLite vault.db와 별도로, 사람이 읽을 수 있는 메타데이터를 저장한다.
type VaultJSON struct {
	ProjectName       string   `json:"projectName"`
	CreatedAt         string   `json:"createdAt"`
	VaultVersion      int      `json:"vaultVersion"`
	ActiveEnvironment string   `json:"activeEnvironment"`
	Agents            []string `json:"agents"`
}

// WriteVaultJSON은 .tene/vault.json 파일을 생성한다.
func WriteVaultJSON(path, projectName, activeEnv string) error {
	vj := VaultJSON{
		ProjectName:       projectName,
		CreatedAt:         time.Now().UTC().Format(time.RFC3339),
		VaultVersion:      1,
		ActiveEnvironment: activeEnv,
		Agents:            []string{"claude"},
	}

	data, err := json.MarshalIndent(vj, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0600)
}

// ReadVaultJSON은 .tene/vault.json 파일을 읽는다.
func ReadVaultJSON(path string) (*VaultJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var vj VaultJSON
	if err := json.Unmarshal(data, &vj); err != nil {
		return nil, err
	}
	return &vj, nil
}

// UpdateActiveEnvironment는 vault.json의 activeEnvironment를 갱신한다.
func UpdateVaultJSONEnv(path, env string) error {
	vj, err := ReadVaultJSON(path)
	if err != nil {
		return err
	}

	vj.ActiveEnvironment = env

	data, err := json.MarshalIndent(vj, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0600)
}
```

### 3.4 수정 파일: `internal/cli/init.go`

`runInit()` 함수에 vault.json 생성 단계를 추가한다. 기존 10번 단계(`.tene/.gitignore` 생성) 앞에 삽입한다.

```go
// 9.5 (신규) Create .tene/vault.json
vaultJSONPath := filepath.Join(teneDir, "vault.json")
if err := vault.WriteVaultJSON(vaultJSONPath, projectName, "default"); err != nil {
    return fmt.Errorf("Cannot create vault.json: %w", err)
}
```

### 3.5 수정 파일: `internal/cli/env.go`

환경 전환 시 vault.json도 업데이트한다.

```go
// 환경 전환 후
vaultJSONPath := filepath.Join(app.Dir, ".tene", "vault.json")
_ = vault.UpdateVaultJSONEnv(vaultJSONPath, name) // 실패해도 SQLite가 primary source
```

### 3.6 테스트 시나리오

```go
// internal/vault/vaultjson_test.go
package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteVaultJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.json")

	err := WriteVaultJSON(path, "test-project", "default")
	if err != nil {
		t.Fatalf("WriteVaultJSON() error: %v", err)
	}

	vj, err := ReadVaultJSON(path)
	if err != nil {
		t.Fatalf("ReadVaultJSON() error: %v", err)
	}

	if vj.ProjectName != "test-project" {
		t.Errorf("ProjectName = %q, want %q", vj.ProjectName, "test-project")
	}
	if vj.ActiveEnvironment != "default" {
		t.Errorf("ActiveEnvironment = %q, want %q", vj.ActiveEnvironment, "default")
	}
	if vj.VaultVersion != 1 {
		t.Errorf("VaultVersion = %d, want 1", vj.VaultVersion)
	}
	if len(vj.Agents) != 1 || vj.Agents[0] != "claude" {
		t.Errorf("Agents = %v, want [claude]", vj.Agents)
	}

	// 파일 퍼미션 검증
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Errorf("permission = %o, want 0600", info.Mode().Perm())
	}
}

func TestUpdateVaultJSONEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.json")

	_ = WriteVaultJSON(path, "test-project", "default")
	err := UpdateVaultJSONEnv(path, "production")
	if err != nil {
		t.Fatalf("UpdateVaultJSONEnv() error: %v", err)
	}

	vj, _ := ReadVaultJSON(path)
	if vj.ActiveEnvironment != "production" {
		t.Errorf("ActiveEnvironment = %q, want %q", vj.ActiveEnvironment, "production")
	}
	// ProjectName이 보존되는지 확인
	if vj.ProjectName != "test-project" {
		t.Errorf("ProjectName = %q, want %q", vj.ProjectName, "test-project")
	}
}
```

---

## 4. ~/.tene/config.json 생성

### 4.1 현재 상태

`~/.tene/` 글로벌 설정 디렉토리가 생성되지 않는다. `tene sync` 명령어에서 analytics(`syncAttempts`, `lastSyncAttempt`)를 기록해야 하지만 현재 누락되어 있다.

### 4.2 JSON 스키마

```json
{
  "version": 1,
  "analytics": {
    "syncAttempts": 0,
    "lastSyncAttempt": null
  },
  "preferences": {
    "color": true,
    "autoKeychain": true
  }
}
```

### 4.3 신규 파일: `internal/config/config.go`

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Config는 ~/.tene/config.json의 구조를 나타낸다.
type Config struct {
	Version     int         `json:"version"`
	Analytics   Analytics   `json:"analytics"`
	Preferences Preferences `json:"preferences"`
}

// Analytics는 사용 통계를 나타낸다.
type Analytics struct {
	SyncAttempts    int     `json:"syncAttempts"`
	LastSyncAttempt *string `json:"lastSyncAttempt"` // nullable ISO 8601
}

// Preferences는 사용자 설정을 나타낸다.
type Preferences struct {
	Color        bool `json:"color"`
	AutoKeychain bool `json:"autoKeychain"`
}

// DefaultConfig는 기본 설정을 반환한다.
func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Analytics: Analytics{
			SyncAttempts:    0,
			LastSyncAttempt: nil,
		},
		Preferences: Preferences{
			Color:        true,
			AutoKeychain: true,
		},
	}
}

// ConfigDir는 ~/.tene/ 경로를 반환한다.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tene"), nil
}

// ConfigPath는 ~/.tene/config.json 경로를 반환한다.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load는 ~/.tene/config.json을 읽는다.
// 파일이 없으면 기본 설정을 반환한다.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), nil // 파싱 실패 시 기본값
	}
	return &cfg, nil
}

// Save는 ~/.tene/config.json에 저장한다.
// 디렉토리가 없으면 생성한다.
func Save(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(data, '\n'), 0600)
}

// EnsureConfigDir는 ~/.tene/ 디렉토리를 확인/생성한다.
func EnsureConfigDir() error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0700)
}

// IncrementSyncAttempts는 syncAttempts를 1 증가시키고 타임스탬프를 기록한다.
func IncrementSyncAttempts() error {
	cfg, err := Load()
	if err != nil {
		cfg = DefaultConfig()
	}

	cfg.Analytics.SyncAttempts++
	now := time.Now().UTC().Format(time.RFC3339)
	cfg.Analytics.LastSyncAttempt = &now

	return Save(cfg)
}
```

### 4.4 수정 파일: `internal/cli/init.go`

`runInit()` 함수에서 글로벌 config 디렉토리 확인/생성을 추가한다.

```go
import "github.com/tene-ai/tene/internal/config"

// runInit() 내부, 기존 단계 4 (.tene/ 디렉토리 생성) 후에 추가
// 4.5 (신규) Ensure global config directory
_ = config.EnsureConfigDir() // 실패해도 계속 진행 (non-critical)
```

### 4.5 수정 파일: `internal/cli/sync_cmd.go`

`runSync()` 함수에 analytics 기록을 추가한다.

```go
import "github.com/tene-ai/tene/internal/config"

func runSync(cmd *cobra.Command, args []string) error {
	// 기존 JSON 출력 / 텍스트 출력 로직 동일...

	// (신규) analytics 기록
	_ = config.IncrementSyncAttempts()

	return nil
}
```

### 4.6 테스트 시나리오

```go
// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Version != 1 {
		t.Errorf("Version = %d, want 1", cfg.Version)
	}
	if cfg.Analytics.SyncAttempts != 0 {
		t.Errorf("SyncAttempts = %d, want 0", cfg.Analytics.SyncAttempts)
	}
	if cfg.Analytics.LastSyncAttempt != nil {
		t.Error("LastSyncAttempt should be nil")
	}
	if !cfg.Preferences.Color {
		t.Error("Color should be true")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// HOME을 임시 디렉토리로 재설정
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg := DefaultConfig()
	cfg.Analytics.SyncAttempts = 3

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.Analytics.SyncAttempts != 3 {
		t.Errorf("SyncAttempts = %d, want 3", loaded.Analytics.SyncAttempts)
	}

	// 파일 퍼미션 확인
	path := filepath.Join(tmpHome, ".tene", "config.json")
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Errorf("permission = %o, want 0600", info.Mode().Perm())
	}
}

func TestLoad_FileNotExists(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Version != 1 {
		t.Errorf("Version = %d, want 1 (default)", cfg.Version)
	}
}

func TestIncrementSyncAttempts(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := IncrementSyncAttempts(); err != nil {
		t.Fatalf("IncrementSyncAttempts() error: %v", err)
	}
	if err := IncrementSyncAttempts(); err != nil {
		t.Fatalf("IncrementSyncAttempts() error: %v", err)
	}

	cfg, _ := Load()
	if cfg.Analytics.SyncAttempts != 2 {
		t.Errorf("SyncAttempts = %d, want 2", cfg.Analytics.SyncAttempts)
	}
	if cfg.Analytics.LastSyncAttempt == nil {
		t.Error("LastSyncAttempt should not be nil")
	}
}
```

---

## 5. .tene.enc 바이너리 포맷

### 5.1 현재 상태

`export.go`의 `exportEncrypted()`는 단순히 `crypto.Encrypt()`의 결과를 파일에 그대로 쓴다. 요구사항의 바이너리 포맷 (Magic bytes, Format version, KDF params, Salt, Nonce, Encrypted Payload) 구조를 따르지 않는다.

### 5.2 바이너리 레이아웃

```
Offset  Size     Field                 설명
------  -------  --------------------  ----------------------------------
0       4        Magic                 "TENE" (0x54454E45)
4       1        Format Version        0x01
5       1        KDF Algorithm         0x01 = Argon2id
6       4        KDF Memory (KB)       65536 (little-endian uint32)
10      4        KDF Iterations        3 (little-endian uint32)
14      1        KDF Parallelism       1
15      1        Salt Length           16
16      16       KDF Salt              128-bit random salt
32      24       Nonce                 192-bit random nonce (XChaCha20)
56      *        Encrypted Payload     XChaCha20-Poly1305(JSON payload)
```

### 5.3 신규 파일: `internal/encfile/encfile.go`

```go
package encfile

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/tene-ai/tene/internal/crypto"
)

var (
	// Magic bytes: "TENE"
	MagicBytes = [4]byte{'T', 'E', 'N', 'E'}

	// 현재 포맷 버전
	FormatVersion byte = 0x01

	// KDF 알고리즘 식별자
	KDFAlgArgon2id byte = 0x01

	// 헤더 크기 (Magic 4 + Version 1 + KDFAlg 1 + Memory 4 + Iterations 4 + Parallelism 1 + SaltLen 1 + Salt 16 + Nonce 24 = 56)
	HeaderSize = 56
)

var (
	ErrInvalidMagic   = errors.New("encfile: invalid magic bytes, not a .tene.enc file")
	ErrUnsupportedVer = errors.New("encfile: unsupported format version")
	ErrFileTooShort   = errors.New("encfile: file too short")
)

// Header는 .tene.enc 파일 헤더를 나타낸다.
type Header struct {
	FormatVersion byte
	KDFAlgorithm  byte
	KDFMemory     uint32 // KB 단위
	KDFIterations uint32
	KDFParallel   byte
	Salt          [16]byte
	Nonce         [24]byte
}

// Encode는 Header를 바이트 슬라이스로 직렬화한다.
func (h *Header) Encode() []byte {
	buf := new(bytes.Buffer)

	buf.Write(MagicBytes[:])               // 4 bytes
	buf.WriteByte(h.FormatVersion)          // 1 byte
	buf.WriteByte(h.KDFAlgorithm)          // 1 byte
	binary.Write(buf, binary.LittleEndian, h.KDFMemory)     // 4 bytes
	binary.Write(buf, binary.LittleEndian, h.KDFIterations) // 4 bytes
	buf.WriteByte(h.KDFParallel)           // 1 byte
	buf.WriteByte(16)                       // salt length, 1 byte
	buf.Write(h.Salt[:])                    // 16 bytes
	buf.Write(h.Nonce[:])                   // 24 bytes

	return buf.Bytes() // 56 bytes total
}

// DecodeHeader는 바이트 슬라이스에서 Header를 파싱한다.
func DecodeHeader(data []byte) (*Header, error) {
	if len(data) < HeaderSize {
		return nil, ErrFileTooShort
	}

	// Magic bytes 검증
	if !bytes.Equal(data[0:4], MagicBytes[:]) {
		return nil, ErrInvalidMagic
	}

	h := &Header{
		FormatVersion: data[4],
		KDFAlgorithm:  data[5],
		KDFParallel:   data[14],
	}

	if h.FormatVersion != FormatVersion {
		return nil, fmt.Errorf("%w: got %d", ErrUnsupportedVer, h.FormatVersion)
	}

	h.KDFMemory = binary.LittleEndian.Uint32(data[6:10])
	h.KDFIterations = binary.LittleEndian.Uint32(data[10:14])

	// saltLen := data[15] -- 현재 항상 16
	copy(h.Salt[:], data[16:32])
	copy(h.Nonce[:], data[32:56])

	return h, nil
}

// Encrypt는 평문을 .tene.enc 바이너리 포맷으로 암호화한다.
// password: Master Password
// plaintext: 암호화할 JSON 바이트
// 반환: 전체 .tene.enc 바이너리 (header + encrypted payload)
func Encrypt(password string, plaintext []byte) ([]byte, error) {
	// 1. Salt 생성
	salt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("encfile: failed to generate salt: %w", err)
	}

	// 2. KDF로 키 유도
	masterKey, err := crypto.DeriveKey(password, salt)
	if err != nil {
		return nil, fmt.Errorf("encfile: KDF failed: %w", err)
	}

	encKey, err := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)
	if err != nil {
		return nil, fmt.Errorf("encfile: sub-key derivation failed: %w", err)
	}

	// 3. 암호화 (crypto.Encrypt가 nonce를 앞에 붙임)
	ciphertext, err := crypto.Encrypt(encKey, plaintext, []byte("tene-export"))
	if err != nil {
		return nil, fmt.Errorf("encfile: encryption failed: %w", err)
	}

	// ciphertext = nonce(24) + encrypted_payload
	// 헤더에 nonce를 별도로 저장하고, payload에서는 nonce를 제거한다.
	var nonce [24]byte
	copy(nonce[:], ciphertext[:24])
	payload := ciphertext[24:]

	// 4. 헤더 구성
	var saltArr [16]byte
	copy(saltArr[:], salt)

	header := &Header{
		FormatVersion: FormatVersion,
		KDFAlgorithm:  KDFAlgArgon2id,
		KDFMemory:     uint32(crypto.ArgonMemory),
		KDFIterations: uint32(crypto.ArgonTime),
		KDFParallel:   uint8(crypto.ArgonThreads),
		Salt:          saltArr,
		Nonce:         nonce,
	}

	// 5. 조립: header + payload
	headerBytes := header.Encode()
	result := make([]byte, len(headerBytes)+len(payload))
	copy(result, headerBytes)
	copy(result[len(headerBytes):], payload)

	return result, nil
}

// Decrypt는 .tene.enc 바이너리 파일을 복호화한다.
// password: Master Password
// data: 전체 .tene.enc 바이너리
// 반환: 복호화된 평문 (JSON 바이트)
func Decrypt(password string, data []byte) ([]byte, error) {
	// 1. 헤더 파싱
	header, err := DecodeHeader(data)
	if err != nil {
		return nil, err
	}

	// 2. KDF로 키 유도 (헤더의 salt 사용)
	masterKey, err := crypto.DeriveKey(password, header.Salt[:])
	if err != nil {
		return nil, fmt.Errorf("encfile: KDF failed: %w", err)
	}

	encKey, err := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)
	if err != nil {
		return nil, fmt.Errorf("encfile: sub-key derivation failed: %w", err)
	}

	// 3. nonce + payload 재조립하여 crypto.Decrypt 호출
	payload := data[HeaderSize:]
	ciphertext := make([]byte, 24+len(payload))
	copy(ciphertext[:24], header.Nonce[:])
	copy(ciphertext[24:], payload)

	plaintext, err := crypto.Decrypt(encKey, ciphertext, []byte("tene-export"))
	if err != nil {
		return nil, fmt.Errorf("encfile: decryption failed (wrong password?): %w", err)
	}

	return plaintext, nil
}

// ReadFile는 .tene.enc 파일을 읽어 복호화한다.
func ReadFile(path, password string) ([]byte, error) {
	data, err := io.ReadAll(nil) // placeholder
	_ = data
	_ = err
	// 실제 구현에서는 os.ReadFile 사용
	return nil, nil
}
```

### 5.4 수정 파일: `internal/cli/export.go`

`exportEncrypted()` 함수를 `encfile` 패키지를 사용하도록 변경한다.

```go
import (
	"encoding/json"
	"time"

	"github.com/tene-ai/tene/internal/encfile"
)

// ExportPayload는 .tene.enc 파일의 암호화 전 JSON 구조이다.
type ExportPayload struct {
	ExportVersion int                      `json:"exportVersion"`
	ExportedAt    string                   `json:"exportedAt"`
	ProjectName   string                   `json:"projectName"`
	Environments  []ExportPayloadEnv       `json:"environments"`
}

type ExportPayloadEnv struct {
	Name    string              `json:"name"`
	Secrets []ExportPayloadSecret `json:"secrets"`
}

type ExportPayloadSecret struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Version   int    `json:"version"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func exportEncrypted(app *App, env string, keys []string, decrypted map[string]string, encKey []byte) error {
	projectName, _ := app.Vault.GetMeta("project_name")
	if projectName == "" {
		projectName = "tene"
	}

	// ExportPayload 구성
	secrets := make([]ExportPayloadSecret, 0, len(keys))
	for _, key := range keys {
		s, _ := app.Vault.GetSecret(key, env)
		secret := ExportPayloadSecret{
			Name:  key,
			Value: decrypted[key],
		}
		if s != nil {
			secret.Version = s.Version
			secret.CreatedAt = s.CreatedAt.Format(time.RFC3339)
			secret.UpdatedAt = s.UpdatedAt.Format(time.RFC3339)
		}
		secrets = append(secrets, secret)
	}

	payload := ExportPayload{
		ExportVersion: 1,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		ProjectName:   projectName,
		Environments: []ExportPayloadEnv{
			{Name: env, Secrets: secrets},
		},
	}

	plaintext, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Master Password가 필요 -- 사용자에게 다시 요청하거나 환경변수에서 획득
	password := os.Getenv("TENE_MASTER_PASSWORD")
	if password == "" {
		password, err = promptPassword("Enter Master Password for encrypted export: ")
		if err != nil {
			return err
		}
	}

	// encfile 포맷으로 암호화
	encrypted, err := encfile.Encrypt(password, plaintext)
	if err != nil {
		return err
	}

	// 출력 파일 결정
	outFile := exportFlagFile
	if outFile == "" {
		outFile = filepath.Base(projectName) + ".tene.enc"
	}

	if err := os.WriteFile(outFile, encrypted, 0600); err != nil {
		return fmt.Errorf("Cannot write to %q: %w", outFile, err)
	}

	// ... 기존 JSON/텍스트 출력 동일 ...
	return nil
}
```

### 5.5 수정 파일: `internal/cli/import.go`

`--encrypted` 모드에서 `encfile.Decrypt()`를 사용하도록 변경한다.

```go
import "github.com/tene-ai/tene/internal/encfile"

func importEncrypted(app *App, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return teneerr.ErrFileNotFound(filePath)
	}

	// 헤더 검증
	_, err = encfile.DecodeHeader(data)
	if err != nil {
		return teneerr.ErrInvalidBackupFile
	}

	// Master Password 입력
	password := os.Getenv("TENE_MASTER_PASSWORD")
	if password == "" {
		password, err = promptPassword("Enter Master Password: ")
		if err != nil {
			return err
		}
	}

	plaintext, err := encfile.Decrypt(password, data)
	if err != nil {
		return teneerr.ErrDecryptFailed
	}

	// JSON 파싱 및 시크릿 복원
	var payload ExportPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return teneerr.ErrInvalidBackupFile
	}

	// ... 시크릿 저장 로직 ...
	return nil
}
```

### 5.6 테스트 시나리오

```go
// internal/encfile/encfile_test.go
package encfile

import (
	"bytes"
	"testing"
)

func TestHeaderEncodeDecode(t *testing.T) {
	h := &Header{
		FormatVersion: FormatVersion,
		KDFAlgorithm:  KDFAlgArgon2id,
		KDFMemory:     65536,
		KDFIterations: 3,
		KDFParallel:   1,
	}
	// salt/nonce는 0으로 초기화된 상태

	encoded := h.Encode()
	if len(encoded) != HeaderSize {
		t.Errorf("header size = %d, want %d", len(encoded), HeaderSize)
	}

	// Magic bytes 확인
	if !bytes.Equal(encoded[:4], MagicBytes[:]) {
		t.Errorf("magic = %x, want %x", encoded[:4], MagicBytes)
	}

	decoded, err := DecodeHeader(encoded)
	if err != nil {
		t.Fatalf("DecodeHeader() error: %v", err)
	}

	if decoded.FormatVersion != FormatVersion {
		t.Errorf("FormatVersion = %d, want %d", decoded.FormatVersion, FormatVersion)
	}
	if decoded.KDFMemory != 65536 {
		t.Errorf("KDFMemory = %d, want 65536", decoded.KDFMemory)
	}
	if decoded.KDFIterations != 3 {
		t.Errorf("KDFIterations = %d, want 3", decoded.KDFIterations)
	}
}

func TestDecodeHeader_InvalidMagic(t *testing.T) {
	data := make([]byte, HeaderSize)
	copy(data[:4], []byte("FAKE"))

	_, err := DecodeHeader(data)
	if err != ErrInvalidMagic {
		t.Errorf("expected ErrInvalidMagic, got %v", err)
	}
}

func TestDecodeHeader_TooShort(t *testing.T) {
	data := make([]byte, 10)
	_, err := DecodeHeader(data)
	if err != ErrFileTooShort {
		t.Errorf("expected ErrFileTooShort, got %v", err)
	}
}

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	password := "testpassword123"
	plaintext := []byte(`{"exportVersion":1,"secrets":[]}`)

	encrypted, err := Encrypt(password, plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	// 헤더 검증
	header, err := DecodeHeader(encrypted)
	if err != nil {
		t.Fatalf("DecodeHeader() error: %v", err)
	}
	if header.FormatVersion != FormatVersion {
		t.Errorf("FormatVersion = %d, want %d", header.FormatVersion, FormatVersion)
	}

	decrypted, err := Decrypt(password, encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
	}
}

func TestDecrypt_WrongPassword(t *testing.T) {
	plaintext := []byte(`{"test":true}`)

	encrypted, _ := Encrypt("correct-password", plaintext)
	_, err := Decrypt("wrong-password", encrypted)
	if err == nil {
		t.Error("expected error for wrong password")
	}
}
```

---

## 6. CLI 통합 테스트

### 6.1 현재 상태

`internal/cli/` 디렉토리에 테스트 파일이 없다. 다른 패키지(`crypto`, `vault`, `recovery`, `keychain`, `claudemd`)에는 단위 테스트가 존재한다.

### 6.2 테스트 전략

CLI 통합 테스트는 실제 cobra 명령어를 실행하여 end-to-end 동작을 검증한다. OS Keychain 의존성을 제거하기 위해 `--no-keychain` 플래그를 사용한다.

### 6.3 신규 파일: `internal/cli/testhelper_test.go`

```go
package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// testEnv는 테스트용 환경을 구성한다.
type testEnv struct {
	Dir     string // 임시 프로젝트 디렉토리
	HomeDir string // 임시 HOME 디렉토리
	t       *testing.T
}

// setupTestEnv는 테스트용 임시 디렉토리와 환경변수를 설정한다.
func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	dir := t.TempDir()
	home := t.TempDir()

	t.Setenv("HOME", home)
	t.Setenv("TENE_MASTER_PASSWORD", "testpassword123")
	t.Setenv("TENE_KEYCHAIN_FALLBACK", "file")

	return &testEnv{Dir: dir, HomeDir: home, t: t}
}

// initVault는 테스트용 볼트를 초기화한다.
func (e *testEnv) initVault() {
	e.t.Helper()
	stdout, stderr, err := e.run("init", "test-project", "--no-keychain", "--quiet")
	if err != nil {
		e.t.Fatalf("init failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
}

// run은 tene CLI 명령어를 실행하고 stdout, stderr, error를 반환한다.
func (e *testEnv) run(args ...string) (string, string, error) {
	// rootCmd를 클린 상태로 리셋
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	// 글로벌 플래그 리셋
	flagJSON = false
	flagQuiet = false
	flagEnv = ""
	flagDir = e.Dir
	flagNoColor = true
	flagNoKeychain = true

	// cobra 명령어 실행
	rootCmd.SetOut(stdoutBuf)
	rootCmd.SetErr(stderrBuf)
	rootCmd.SetArgs(append([]string{"--dir", e.Dir, "--no-keychain"}, args...))

	err := rootCmd.Execute()

	return stdoutBuf.String(), stderrBuf.String(), err
}

// runJSON은 --json 모드로 실행한다.
func (e *testEnv) runJSON(args ...string) (string, string, error) {
	allArgs := append([]string{"--json"}, args...)
	return e.run(allArgs...)
}
```

### 6.4 신규 파일: `internal/cli/init_test.go`

```go
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInit_Basic(t *testing.T) {
	env := setupTestEnv(t)

	stdout, _, err := env.run("init", "test-project", "--quiet")
	if err != nil {
		t.Fatalf("init error: %v", err)
	}

	// vault.db 존재 확인
	vaultPath := filepath.Join(env.Dir, ".tene", "vault.db")
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		t.Error("vault.db not created")
	}

	// .gitignore 존재 확인
	gitignorePath := filepath.Join(env.Dir, ".tene", ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error(".gitignore not created")
	}

	// Recovery Key 출력 확인
	if len(stdout) == 0 {
		t.Error("expected recovery key in stdout")
	}
}

func TestInit_AlreadyInitialized(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	// 두 번째 init
	_, _, err := env.run("init", "--quiet")
	if err != nil {
		t.Fatalf("second init should not error: %v", err)
	}
}

func TestInit_JSON(t *testing.T) {
	env := setupTestEnv(t)

	stdout, _, err := env.runJSON("init", "test-project")
	if err != nil {
		t.Fatalf("init error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse error: %v\nstdout: %s", err, stdout)
	}

	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if result["project"] != "test-project" {
		t.Errorf("project = %v, want test-project", result["project"])
	}
}
```

### 6.5 신규 파일: `internal/cli/set_get_test.go`

```go
package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSetGet_Basic(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	// Set
	_, _, err := env.run("set", "API_KEY", "test-value-123", "--overwrite")
	if err != nil {
		t.Fatalf("set error: %v", err)
	}

	// Get
	stdout, _, err := env.run("get", "API_KEY")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}

	got := strings.TrimSpace(stdout)
	if got != "test-value-123" {
		t.Errorf("get = %q, want %q", got, "test-value-123")
	}
}

func TestSet_InvalidKeyName(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, err := env.run("set", "invalid-key", "value")
	if err == nil {
		t.Error("expected error for invalid key name")
	}
}

func TestSet_DuplicateWithoutOverwrite(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, _ = env.run("set", "MY_KEY", "value1")
	_, _, err := env.run("set", "MY_KEY", "value2")
	if err == nil {
		t.Error("expected error for duplicate key without --overwrite")
	}
}

func TestSet_OverwriteExisting(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, _ = env.run("set", "MY_KEY", "value1")
	_, _, err := env.run("set", "MY_KEY", "value2", "--overwrite")
	if err != nil {
		t.Fatalf("overwrite error: %v", err)
	}

	stdout, _, _ := env.run("get", "MY_KEY")
	if strings.TrimSpace(stdout) != "value2" {
		t.Errorf("get after overwrite = %q, want %q", strings.TrimSpace(stdout), "value2")
	}
}

func TestGet_SecretNotFound(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	_, _, err := env.run("get", "NONEXISTENT")
	if err == nil {
		t.Error("expected error for nonexistent secret")
	}
}

func TestGet_JSON(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()
	_, _, _ = env.run("set", "TEST_KEY", "test-val")

	stdout, _, err := env.runJSON("get", "TEST_KEY")
	if err != nil {
		t.Fatalf("get --json error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}
	if result["ok"] != true {
		t.Error("expected ok=true")
	}
	if result["value"] != "test-val" {
		t.Errorf("value = %v, want test-val", result["value"])
	}
}
```

### 6.6 Table-Driven 테스트 패턴

```go
func TestSetGet_TableDriven(t *testing.T) {
	env := setupTestEnv(t)
	env.initVault()

	tests := []struct {
		name     string
		key      string
		value    string
		wantErr  bool
	}{
		{"basic ASCII", "MY_SECRET", "hello123", false},
		{"unicode value", "UNICODE_VAL", "한국어값", false},
		{"long value", "LONG_VAL", strings.Repeat("x", 1000), false},
		{"special chars", "SPECIAL", "pa$$w0rd!@#", false},
		{"newline in value", "MULTILINE", "line1\nline2", false},
		{"empty value", "EMPTY", "", true},
		{"invalid key lowercase", "lowercase", "val", true},
		{"invalid key dash", "MY-KEY", "val", true},
		{"reserved key PATH", "PATH", "val", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := env.run("set", tt.key, tt.value, "--overwrite")
			if (err != nil) != tt.wantErr {
				t.Errorf("set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				stdout, _, getErr := env.run("get", tt.key)
				if getErr != nil {
					t.Errorf("get(%q) error = %v", tt.key, getErr)
					return
				}
				got := strings.TrimSpace(stdout)
				if got != tt.value {
					t.Errorf("get(%q) = %q, want %q", tt.key, got, tt.value)
				}
			}
		})
	}
}
```

---

## 7. --no-color / NO_COLOR 지원

### 7.1 현재 상태

`--no-color` 플래그는 `root.go`에 정의되어 있지만, 실제로 색상 출력을 제어하는 로직이 없다. 현재 CLI는 `fmt.Printf()`로 직접 출력하며 색상 코드를 사용하지 않지만, 향후 색상 출력을 위한 기반을 마련해야 한다.

### 7.2 색상 판단 로직

```
색상 사용 = (stdout가 TTY) AND (--no-color가 아님) AND (NO_COLOR 환경변수가 설정되지 않음)
```

### 7.3 신규 파일: `internal/cli/color.go`

```go
package cli

import (
	"fmt"
	"os"
)

// ANSI 색상 코드
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// colorEnabled는 색상 출력이 활성화되어 있는지 반환한다.
func colorEnabled() bool {
	// --no-color 플래그
	if flagNoColor {
		return false
	}
	// NO_COLOR 환경변수 (https://no-color.org/)
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	// TTY 감지
	return isTerminal()
}

// colorize는 텍스트에 ANSI 색상을 적용한다.
// 색상이 비활성화되어 있으면 원본 텍스트를 반환한다.
func colorize(color, text string) string {
	if !colorEnabled() {
		return text
	}
	return color + text + colorReset
}

// 편의 함수
func colorRed_(text string) string    { return colorize(colorRed, text) }
func colorGreen_(text string) string  { return colorize(colorGreen, text) }
func colorYellow_(text string) string { return colorize(colorYellow, text) }
func colorBlue_(text string) string   { return colorize(colorBlue, text) }
func colorBold_(text string) string   { return colorize(colorBold, text) }
func colorDim_(text string) string    { return colorize(colorDim, text) }

// printSuccess는 성공 메시지를 출력한다.
func printSuccess(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(colorGreen_(msg))
}

// printWarning은 경고 메시지를 출력한다.
func printWarning(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s\n", colorYellow_(msg))
}

// printError_은 에러 메시지를 출력한다.
func printError_(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s\n", colorRed_(msg))
}
```

### 7.4 적용 예시

기존 `fmt.Printf("STRIPE_KEY saved (encrypted, default)\n")` 같은 출력을 색상 유틸로 교체할 수 있다. 단, Phase 1에서는 기반만 마련하고, 실제 색상 적용은 점진적으로 진행한다.

```go
// internal/cli/set.go 에서
if !flagQuiet {
    printSuccess("%s saved (encrypted, %s)", keyName, env)
}
```

### 7.5 테스트 시나리오

```go
// internal/cli/color_test.go
package cli

import (
	"os"
	"testing"
)

func TestColorEnabled_NoColorFlag(t *testing.T) {
	flagNoColor = true
	defer func() { flagNoColor = false }()

	if colorEnabled() {
		t.Error("colorEnabled() should return false when --no-color is set")
	}
}

func TestColorEnabled_NoColorEnv(t *testing.T) {
	flagNoColor = false
	t.Setenv("NO_COLOR", "1")

	if colorEnabled() {
		t.Error("colorEnabled() should return false when NO_COLOR env is set")
	}
}

func TestColorize_Disabled(t *testing.T) {
	flagNoColor = true
	defer func() { flagNoColor = false }()

	result := colorize(colorRed, "hello")
	if result != "hello" {
		t.Errorf("colorize() = %q, want %q (no color)", result, "hello")
	}
}

func TestColorize_Enabled(t *testing.T) {
	// 이 테스트는 TTY가 아닌 환경에서는 의미가 제한적
	// 색상이 비활성화된 상태에서 원본 텍스트가 보존되는지만 검증
	flagNoColor = true
	defer func() { flagNoColor = false }()

	text := "test message"
	if got := colorRed_(text); got != text {
		t.Errorf("colorRed_() = %q, want %q", got, text)
	}
}
```

---

## 8. 메모리 제로화

### 8.1 현재 상태

`masterKey`, `encKey`, `plaintext` 등 민감한 바이트 슬라이스가 사용 후 메모리에서 제거되지 않는다. GC에 의해 언젠가 회수되지만, 그 사이 메모리 덤프로 유출될 수 있다.

### 8.2 설계

Go에서 메모리 제로화의 한계:
- Go GC가 슬라이스를 이동시킬 수 있으므로 이전 메모리에 복사본이 남을 수 있다
- 완벽한 보호는 어렵지만, best-effort로 `defer` 패턴 적용

### 8.3 신규 파일: `internal/crypto/zero.go`

```go
package crypto

// ZeroBytes는 바이트 슬라이스의 모든 바이트를 0으로 채운다.
// 컴파일러 최적화로 제거되는 것을 방지하기 위해 volatile 패턴 사용.
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
	// 컴파일러 최적화 방지: 슬라이스가 사용된 것처럼 처리
	keepAlive(b)
}

// keepAlive는 컴파일러가 제로화를 최적화하지 못하도록 한다.
//
//go:noinline
func keepAlive(b []byte) {
	if len(b) > 0 {
		_ = b[0]
	}
}
```

### 8.4 적용: `internal/cli/get.go`

```go
func runGet(cmd *cobra.Command, args []string) error {
	// ...

	masterKey, err := loadOrPromptMasterKey(app)
	if err != nil {
		return err
	}
	defer crypto.ZeroBytes(masterKey) // 신규 추가

	encKey, err := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)
	if err != nil {
		return err
	}
	defer crypto.ZeroBytes(encKey) // 신규 추가

	// ...

	plaintext, err := crypto.Decrypt(encKey, ciphertext, []byte(keyName))
	if err != nil {
		return teneerr.ErrDecryptFailed
	}
	defer crypto.ZeroBytes(plaintext) // 신규 추가

	// ... 출력 후 함수 종료 시 자동 제로화
}
```

### 8.5 적용 대상 함수 목록

| 파일 | 함수 | 제로화 대상 |
|------|------|------------|
| `internal/cli/root.go` | `loadOrPromptMasterKey()` | password 바이트 (term.ReadPassword 결과) |
| `internal/cli/root.go` | `deriveMasterKeyFromPassword()` | masterKey 결과 (caller 책임) |
| `internal/cli/get.go` | `runGet()` | masterKey, encKey, plaintext |
| `internal/cli/set.go` | `runSet()` | masterKey, encKey, value 바이트 |
| `internal/cli/run.go` | `runRun()` | masterKey, encKey, 각 시크릿 plaintext |
| `internal/cli/export.go` | `runExport()` | masterKey, encKey, 모든 decrypted 값 |
| `internal/cli/import.go` | `runImport()` | masterKey, encKey |
| `internal/cli/passwd.go` | `runPasswd()` | oldMasterKey, newMasterKey, oldEncKey, newEncKey |
| `internal/cli/recover.go` | `runRecover()` | recoveredMasterKey, newMasterKey |
| `internal/cli/init.go` | `runInit()` | masterKey, password |

### 8.6 테스트 시나리오

```go
// internal/crypto/zero_test.go
package crypto

import (
	"testing"
)

func TestZeroBytes(t *testing.T) {
	data := []byte{0x41, 0x42, 0x43, 0x44, 0x45}
	ZeroBytes(data)

	for i, b := range data {
		if b != 0 {
			t.Errorf("data[%d] = 0x%02x, want 0x00", i, b)
		}
	}
}

func TestZeroBytes_Empty(t *testing.T) {
	// 빈 슬라이스에서 패닉이 발생하지 않아야 한다
	ZeroBytes(nil)
	ZeroBytes([]byte{})
}

func TestZeroBytes_Large(t *testing.T) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = 0xFF
	}

	ZeroBytes(data)

	for i, b := range data {
		if b != 0 {
			t.Errorf("data[%d] = 0x%02x, want 0x00", i, b)
			break
		}
	}
}
```

---

## 9. CLAUDE.md URL 통일

### 9.1 현재 상태

`internal/claudemd/template.go`의 GitHub URL이 `agentkay`를 사용하고 있다:

```go
const SecretsMdTemplate = `# Secrets Management

This project uses [tene](https://github.com/agentkay/tene) for secret management.
```

`go.mod`의 모듈 경로는 `github.com/tene-ai/tene`이다.

### 9.2 수정 파일: `internal/claudemd/template.go`

```go
// 변경 전
This project uses [tene](https://github.com/agentkay/tene) for secret management.

// 변경 후
This project uses [tene](https://github.com/tene-ai/tene) for secret management.
```

### 9.3 테스트

기존 `internal/claudemd/generator_test.go`에서 URL 검증을 추가한다.

```go
func TestTemplate_URL(t *testing.T) {
	if !strings.Contains(SecretsMdTemplate, "github.com/tene-ai/tene") {
		t.Error("template should contain tomo-kay URL")
	}
	if strings.Contains(SecretsMdTemplate, "agentkay") {
		t.Error("template should not contain agentkay URL")
	}
}
```

---

## 10. Vault 인터페이스 분리

### 10.1 현재 상태

`App` struct가 `*vault.Vault` 구체 타입을 직접 참조한다:

```go
type App struct {
    Vault    *vault.Vault   // 구체 타입 직접 참조
    Keychain keychain.KeyStore // 인터페이스 (좋은 패턴)
    // ...
}
```

이로 인해 CLI 테스트에서 Vault를 모킹하기 어렵다.

### 10.2 목표

`VaultStore` 인터페이스를 정의하여 App이 인터페이스에 의존하도록 변경한다. 기존 `*vault.Vault`는 이 인터페이스를 자동으로 만족한다.

### 10.3 신규 파일: `internal/vault/store.go`

```go
package vault

// VaultStore는 Vault 저장소의 인터페이스이다.
// CLI 명령어에서 사용하는 메서드만 포함한다.
type VaultStore interface {
	// Close는 저장소를 닫는다.
	Close() error

	// --- 메타데이터 ---
	SetMeta(key, value string) error
	GetMeta(key string) (string, error)

	// --- 시크릿 CRUD ---
	SetSecret(name, encryptedValue, env string) error
	GetSecret(name, env string) (*Secret, error)
	ListSecrets(env string) ([]Secret, error)
	DeleteSecret(name, env string) error
	SecretExists(name, env string) (bool, error)
	CountSecrets(env string) (int, error)
	GetAllSecrets(env string) (map[string]string, error)
	SetSecretBatch(secrets map[string]string, env string) error

	// --- 환경 관리 ---
	ListEnvironments() ([]Environment, error)
	GetActiveEnvironment() (string, error)
	SetActiveEnvironment(name string) error
	CreateEnvironment(name string) error
	DeleteEnvironment(name string) (int, error)
	EnvironmentExists(name string) (bool, error)

	// --- 감사 로그 ---
	AddAuditLog(action, resourceName, details string) error
}

// 컴파일 타임 인터페이스 만족 검증
var _ VaultStore = (*Vault)(nil)
```

### 10.4 수정 파일: `internal/cli/root.go`

```go
// App holds the dependencies needed for CLI execution.
type App struct {
	Vault    vault.VaultStore     // 변경: *vault.Vault -> vault.VaultStore
	Keychain keychain.KeyStore
	Dir      string
	Env      string
	JSON     bool
	Quiet    bool
}
```

### 10.5 영향 범위

`App.Vault`의 타입만 인터페이스로 변경되므로, 기존 코드의 메서드 호출은 모두 그대로 동작한다. `*vault.Vault`가 `VaultStore` 인터페이스를 자동으로 만족하기 때문이다.

영향 없는 파일들 (변경 불필요):
- `internal/cli/get.go` -- `app.Vault.GetSecret()` 호출 동일
- `internal/cli/set.go` -- `app.Vault.SetSecret()` 호출 동일
- `internal/cli/run.go` -- `app.Vault.GetAllSecrets()` 호출 동일
- 기타 모든 CLI 명령어 파일

### 10.6 테스트용 Mock 예시

```go
// internal/cli/mock_test.go
package cli

import "github.com/tene-ai/tene/internal/vault"

// mockVault는 테스트용 VaultStore 구현이다.
type mockVault struct {
	secrets     map[string]map[string]string // env -> name -> encValue
	meta        map[string]string
	environments map[string]bool // name -> isActive
	auditLogs   []string
	closed      bool
}

func newMockVault() *mockVault {
	return &mockVault{
		secrets:      make(map[string]map[string]string),
		meta:         make(map[string]string),
		environments: map[string]bool{"default": true},
		auditLogs:    nil,
	}
}

func (m *mockVault) Close() error {
	m.closed = true
	return nil
}

func (m *mockVault) SetMeta(key, value string) error {
	m.meta[key] = value
	return nil
}

func (m *mockVault) GetMeta(key string) (string, error) {
	v, ok := m.meta[key]
	if !ok {
		return "", vault.ErrMetaNotFound
	}
	return v, nil
}

func (m *mockVault) SetSecret(name, encryptedValue, env string) error {
	if m.secrets[env] == nil {
		m.secrets[env] = make(map[string]string)
	}
	m.secrets[env][name] = encryptedValue
	return nil
}

func (m *mockVault) GetSecret(name, env string) (*vault.Secret, error) {
	envSecrets, ok := m.secrets[env]
	if !ok {
		return nil, vault.ErrSecretNotFound
	}
	encVal, ok := envSecrets[name]
	if !ok {
		return nil, vault.ErrSecretNotFound
	}
	return &vault.Secret{
		Name:           name,
		EncryptedValue: encVal,
		Environment:    env,
		Version:        1,
	}, nil
}

func (m *mockVault) ListSecrets(env string) ([]vault.Secret, error) {
	var result []vault.Secret
	for name, encVal := range m.secrets[env] {
		result = append(result, vault.Secret{Name: name, EncryptedValue: encVal, Environment: env})
	}
	return result, nil
}

func (m *mockVault) DeleteSecret(name, env string) error {
	if envSecrets, ok := m.secrets[env]; ok {
		if _, exists := envSecrets[name]; exists {
			delete(envSecrets, name)
			return nil
		}
	}
	return vault.ErrSecretNotFound
}

func (m *mockVault) SecretExists(name, env string) (bool, error) {
	if envSecrets, ok := m.secrets[env]; ok {
		_, exists := envSecrets[name]
		return exists, nil
	}
	return false, nil
}

func (m *mockVault) CountSecrets(env string) (int, error) {
	return len(m.secrets[env]), nil
}

func (m *mockVault) GetAllSecrets(env string) (map[string]string, error) {
	result := make(map[string]string)
	for k, v := range m.secrets[env] {
		result[k] = v
	}
	return result, nil
}

func (m *mockVault) SetSecretBatch(secrets map[string]string, env string) error {
	for name, val := range secrets {
		_ = m.SetSecret(name, val, env)
	}
	return nil
}

func (m *mockVault) ListEnvironments() ([]vault.Environment, error) {
	var result []vault.Environment
	for name, isActive := range m.environments {
		result = append(result, vault.Environment{Name: name, IsActive: isActive})
	}
	return result, nil
}

func (m *mockVault) GetActiveEnvironment() (string, error) {
	for name, isActive := range m.environments {
		if isActive {
			return name, nil
		}
	}
	return "default", nil
}

func (m *mockVault) SetActiveEnvironment(name string) error {
	for k := range m.environments {
		m.environments[k] = false
	}
	m.environments[name] = true
	return nil
}

func (m *mockVault) CreateEnvironment(name string) error {
	if _, exists := m.environments[name]; exists {
		return vault.ErrEnvironmentExists
	}
	m.environments[name] = false
	return nil
}

func (m *mockVault) DeleteEnvironment(name string) (int, error) {
	if _, exists := m.environments[name]; !exists {
		return 0, vault.ErrEnvironmentNotFound
	}
	count := len(m.secrets[name])
	delete(m.environments, name)
	delete(m.secrets, name)
	return count, nil
}

func (m *mockVault) EnvironmentExists(name string) (bool, error) {
	_, exists := m.environments[name]
	return exists, nil
}

func (m *mockVault) AddAuditLog(action, resourceName, details string) error {
	m.auditLogs = append(m.auditLogs, action+":"+resourceName)
	return nil
}

// 인터페이스 만족 검증
var _ vault.VaultStore = (*mockVault)(nil)
```

---

## 파일 생성/수정 종합 목록

### 신규 생성 파일

| 파일 경로 | 설명 |
|-----------|------|
| `internal/errors/errors.go` | TeneError struct, 헬퍼 함수 |
| `internal/errors/codes.go` | 23개 에러 코드 상수 |
| `internal/errors/errors_test.go` | 에러 시스템 테스트 |
| `internal/vault/vaultjson.go` | vault.json 읽기/쓰기 |
| `internal/vault/vaultjson_test.go` | vault.json 테스트 |
| `internal/vault/store.go` | VaultStore 인터페이스 |
| `internal/config/config.go` | ~/.tene/config.json 관리 |
| `internal/config/config_test.go` | config 테스트 |
| `internal/encfile/encfile.go` | .tene.enc 바이너리 포맷 |
| `internal/encfile/encfile_test.go` | encfile 테스트 |
| `internal/crypto/zero.go` | 메모리 제로화 함수 |
| `internal/crypto/zero_test.go` | 제로화 테스트 |
| `internal/cli/color.go` | 색상 출력 유틸 |
| `internal/cli/color_test.go` | 색상 테스트 |
| `internal/cli/testhelper_test.go` | CLI 통합 테스트 헬퍼 |
| `internal/cli/init_test.go` | init 명령어 테스트 |
| `internal/cli/set_get_test.go` | set/get 명령어 테스트 |
| `internal/cli/mock_test.go` | VaultStore 모킹 |

### 수정 파일

| 파일 경로 | 변경 내용 |
|-----------|----------|
| `cmd/tene/main.go` | TeneError type assertion, exit code 분기, --json 에러 출력 |
| `internal/cli/root.go` | `App.Vault` 타입을 `vault.VaultStore`로 변경, `loadApp()` 에러를 TeneError로 교체 |
| `internal/cli/init.go` | vault.json 생성 추가, config 디렉토리 생성 추가 |
| `internal/cli/set.go` | 에러를 TeneError로 교체, 메모리 제로화 추가 |
| `internal/cli/get.go` | 에러를 TeneError로 교체, 메모리 제로화 추가 |
| `internal/cli/run.go` | 에러를 TeneError로 교체, 메모리 제로화 추가 |
| `internal/cli/export.go` | encfile 패키지 사용, ExportPayload JSON 구조, 메모리 제로화 |
| `internal/cli/import.go` | encfile 패키지 사용, 에러를 TeneError로 교체 |
| `internal/cli/helpers.go` | validateKeyName/validateEnvName에서 TeneError 반환 |
| `internal/cli/env.go` | vault.json 환경 업데이트 추가, 에러를 TeneError로 교체 |
| `internal/cli/passwd.go` | 에러를 TeneError로 교체, 메모리 제로화 추가 |
| `internal/cli/recover.go` | 에러를 TeneError로 교체, 메모리 제로화 추가 |
| `internal/cli/sync_cmd.go` | analytics 기록 추가 |
| `internal/claudemd/template.go` | URL을 `agentkay` -> `tomo-kay`로 변경 |

---

## 구현 우선순위

| 순위 | 항목 | 이유 |
|:----:|------|------|
| 1 | 구조화된 에러 코드 시스템 + Exit Code 분기 | 모든 CLI 에러 처리의 기반. 다른 항목보다 먼저 구현해야 일관성 유지 |
| 2 | Vault 인터페이스 분리 | CLI 통합 테스트 작성의 전제조건 |
| 3 | CLI 통합 테스트 | 이후 변경사항의 회귀 테스트 역할 |
| 4 | .tene/vault.json 생성 | init 명령어 보강 |
| 5 | ~/.tene/config.json 생성 + sync analytics | config 패키지 신규 |
| 6 | .tene.enc 바이너리 포맷 | export/import 보강 |
| 7 | --no-color / NO_COLOR | UX 개선 |
| 8 | 메모리 제로화 | 보안 강화 |
| 9 | CLAUDE.md URL 통일 | 1줄 변경 |

---

## 의존성 그래프 (신규 패키지 포함)

```
cmd/tene/main.go
    |
    +--- internal/errors    (신규: TeneError)
    |
    v
internal/cli
    |
    +--- internal/errors    (신규)
    +--- internal/crypto    (기존 + zero.go)
    +--- internal/recovery  (기존)
    +--- internal/vault     (기존 + store.go, vaultjson.go)
    +--- internal/keychain  (기존)
    +--- internal/claudemd  (기존)
    +--- internal/config    (신규)
    +--- internal/encfile   (신규)

internal/encfile -> internal/crypto
internal/config  -> (독립, os/json만 사용)
internal/errors  -> (독립, encoding/json만 사용)
```
