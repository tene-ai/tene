# Tene CLI 요구사항 명세서

> Version: 1.0 (2026-04-06)
> 대상: Go CLI MVP (Phase 1)
> 목적: 개발자가 이 문서만 보고 바로 코딩을 시작할 수 있는 상세 명세

---

## 목차

1. [기능 요구사항 (Functional Requirements)](#1-기능-요구사항-functional-requirements)
2. [비기능 요구사항 (Non-Functional Requirements)](#2-비기능-요구사항-non-functional-requirements)
3. [데이터 모델 상세](#3-데이터-모델-상세)
4. [암호화 명세](#4-암호화-명세)
5. [OS Keychain 연동 명세](#5-os-keychain-연동-명세)
6. [CI/CD 환경 대응](#6-cicd-환경-대응)
7. [에러 코드 체계](#7-에러-코드-체계)
8. [User Stories > Acceptance Criteria](#8-user-stories--acceptance-criteria)

---

## 1. 기능 요구사항 (Functional Requirements)

### 1.0 글로벌 플래그

모든 명령어에 공통으로 적용되는 플래그.

| 플래그 | 단축 | 설명 | 기본값 |
|--------|------|------|--------|
| `--version` | `-v` | 버전 출력 후 종료 | - |
| `--help` | `-h` | 도움말 출력 후 종료 | - |
| `--json` | - | JSON 형식 출력 (AI 에이전트 파싱용) | `false` |
| `--quiet` | `-q` | 최소 출력 (에러만) | `false` |
| `--env <name>` | `-e` | 대상 환경 지정 | 현재 활성 환경 |
| `--no-color` | - | 색상 출력 비활성화 | `false` (TTY 감지) |
| `--no-keychain` | - | OS Keychain 사용 안 함 (CI/CD용) | `false` |

**TTY 감지 규칙:**
- stdout이 TTY일 때: 색상 출력, 대화형 프롬프트 가능
- stdout이 파이프/리디렉션일 때: 색상 없음, 순수 데이터만 출력
- `--no-color`가 있으면: 항상 색상 없음
- `NO_COLOR` 환경변수가 설정되어 있으면: 색상 없음

---

### 1.1 `tene init`

**목적:** 프로젝트 디렉토리에 Tene 볼트를 초기화하고, CLAUDE.md를 자동 생성한다.

#### 시그니처

```
tene init [project-name]
```

| 인자/플래그 | 필수 | 설명 |
|-------------|:----:|------|
| `project-name` | 아니오 | 프로젝트 이름. 생략 시 현재 디렉토리 이름 사용 |

#### 대화형 (Interactive) 동작 (TTY)

```
$ tene init

  Welcome to Tene! Let's set up your local secret vault.

  Project name (my-project):

  Set your Master Password (used to encrypt all secrets):
  Master Password: ********
  Confirm: ********

  Generating encryption keys...

  Recovery Key (write this down and keep it safe!):
  +--------------------------------------------------+
  |   apple banana cherry dolphin eagle frost          |
  |   grape harbor island jungle kite lemon            |
  |                                                    |
  |   If you forget your Master Password,              |
  |   this is the ONLY way to recover.                 |
  +--------------------------------------------------+

  Created .tene/vault.db (local encrypted vault)
  Added .tene/ to .gitignore
  Master Key saved to OS Keychain
  Generated CLAUDE.md (Claude Code will auto-detect tene)

  Project "my-project" initialized.
  Default environment "default" created.

  Next: tene set KEY VALUE to add your first secret.

  Tip: No server needed. Your secrets stay on this device.
       Claude Code will automatically use tene.
```

#### 비대화형 (Non-Interactive) 동작

`TENE_MASTER_PASSWORD` 환경변수가 설정된 경우 프롬프트 없이 진행.

```bash
TENE_MASTER_PASSWORD=mysecret tene init my-project --quiet
```

stdout (--quiet 없을 때):
```
Created .tene/vault.db
Generated CLAUDE.md
Recovery Key: apple banana cherry dolphin eagle frost grape harbor island jungle kite lemon
```

#### 수행 단계

1. `.tene/` 디렉토리 존재 여부 확인
   - 이미 존재: `Vault already exists. Use existing vault.` 출력하고 종료 (종료 코드 0)
2. Master Password 입력받기 (최소 8자)
   - 확인 입력과 불일치 시 재입력 요청 (최대 3회)
3. 128-bit random salt 생성
4. Argon2id KDF로 Master Key 유도
5. HKDF로 Encryption Key 파생
6. Recovery Key 생성 (BIP-39 12단어)
7. Recovery Key로 Master Key 암호화하여 recovery_blob 생성
8. `.tene/` 디렉토리 생성 (퍼미션 0700)
9. `.tene/vault.db` 생성 (SQLite, 퍼미션 0600)
   - 테이블 생성 (vault_meta, environments, secrets, audit_log)
   - vault_meta에 메타데이터 저장 (vault_version, kdf_salt, kdf_params, recovery_blob, created_at)
   - "default" 환경 생성
10. `.tene/vault.json` 생성 (퍼미션 0600)
11. `.tene/.gitignore` 생성 (내용: `*`)
12. 프로젝트 루트 `.gitignore`에 `.tene/` 추가 (없으면 생성, 이미 있으면 스킵)
13. Master Key를 OS Keychain에 저장
14. CLAUDE.md 생성/업데이트
15. 감사 로그 기록: `vault.init`

#### CLAUDE.md 생성 로직

| 상황 | 동작 |
|------|------|
| CLAUDE.md가 없음 | 새로 생성 |
| CLAUDE.md가 있고, `# Secrets Management` 섹션 없음 | 파일 끝에 빈 줄 + 섹션 추가 |
| CLAUDE.md가 있고, `# Secrets Management` 섹션 있음 | 스킵 (중복 방지) |

**생성되는 CLAUDE.md 내용 (영어, 확정본):**

```markdown
# Secrets Management

This project uses [tene](https://github.com/agentkay/tene) for secret management.

## Usage
- Get a secret: `tene get <KEY>`
- List secrets: `tene list`
- Run with secrets injected: `tene run -- <command>`
- Set a secret: `tene set <KEY> <VALUE>`

## Rules
- Never hardcode secret values in source code
- Access secrets via environment variables
- Do not create .env files -- use `tene run` instead
- Use `tene list` to see available secrets
```

#### --json 출력 스키마

```json
{
  "ok": true,
  "project": "my-project",
  "vault": ".tene/vault.db",
  "claudeMd": "CLAUDE.md",
  "recoveryKey": "apple banana cherry dolphin eagle frost grape harbor island jungle kite lemon",
  "environment": "default"
}
```

#### 에러 시나리오

| 에러 | 메시지 | 종료 코드 | --json |
|------|--------|:---------:|--------|
| 이미 초기화됨 | `Vault already exists. Use existing vault.` | 0 | `{"ok": true, "message": "already_initialized"}` |
| 패스워드 불일치 | `Passwords do not match. Try again.` | 2 | `{"ok": false, "error": "PASSWORD_MISMATCH"}` |
| 패스워드 길이 부족 | `Master Password must be at least 8 characters.` | 2 | `{"ok": false, "error": "PASSWORD_TOO_SHORT"}` |
| 디스크 공간 부족 | `Cannot create vault: insufficient disk space.` | 1 | `{"ok": false, "error": "DISK_FULL"}` |
| 퍼미션 거부 | `Cannot create .tene/ directory: permission denied.` | 1 | `{"ok": false, "error": "PERMISSION_DENIED"}` |

---

### 1.2 `tene set`

**목적:** 시크릿을 암호화하여 로컬 볼트에 저장한다.

#### 시그니처

```
tene set <KEY> [VALUE]
```

| 인자/플래그 | 필수 | 설명 |
|-------------|:----:|------|
| `KEY` | 예 | 시크릿 키 이름. 영대문자, 숫자, 밑줄만 허용 (`^[A-Z][A-Z0-9_]*$`) |
| `VALUE` | 아니오 | 시크릿 값. 생략 시 대화형 프롬프트 또는 stdin에서 읽기 |
| `--env <name>` | 아니오 | 대상 환경 (기본: 현재 활성 환경) |
| `--stdin` | 아니오 | stdin에서 값 읽기 (shell history 방지) |
| `--overwrite` | 아니오 | 기존 시크릿 덮어쓰기 (기본: 이미 존재하면 에러) |

#### 키 이름 유효성 검사

- 정규식: `^[A-Z][A-Z0-9_]*$`
- 최소 1자, 최대 256자
- 예약어 금지: `PATH`, `HOME`, `USER`, `SHELL`, `TENE_MASTER_PASSWORD`

#### 값 입력 방식

```bash
# 1. 인라인 (shell history에 남음 -- 주의)
tene set STRIPE_KEY sk_test_xxxxx

# 2. 대화형 (TTY일 때, VALUE 생략 시)
tene set STRIPE_KEY
? Value: ********  (마스킹됨)

# 3. stdin 파이프 (권장)
echo "sk_test_xxxxx" | tene set STRIPE_KEY --stdin

# 4. 파일에서
cat secret.txt | tene set STRIPE_KEY --stdin
```

#### 정상 동작

stdout:
```
STRIPE_KEY saved (encrypted, default)
```

`--quiet` 모드: 출력 없음 (종료 코드 0)

#### --json 출력 스키마

```json
{
  "ok": true,
  "name": "STRIPE_KEY",
  "environment": "default",
  "version": 1,
  "created": true
}
```

업데이트(--overwrite) 시:
```json
{
  "ok": true,
  "name": "STRIPE_KEY",
  "environment": "default",
  "version": 2,
  "created": false
}
```

#### 에러 시나리오

| 에러 | 메시지 | 종료 코드 |
|------|--------|:---------:|
| 볼트 미초기화 | `Not in a Tene project. Run "tene init" first.` | 1 |
| 잘못된 키 이름 | `Invalid key name "key". Keys must match [A-Z][A-Z0-9_]*.` | 1 |
| 이미 존재 | `Secret "KEY" already exists. Use --overwrite to replace.` | 1 |
| 빈 값 | `Value cannot be empty.` | 1 |
| 값 너무 김 | `Value exceeds maximum size (64KB).` | 1 |
| 환경 없음 | `Environment "staging" not found. Create it with "tene env create staging".` | 1 |
| Keychain 실패 | Master Password 프롬프트로 폴백 | 0 (정상 진행) |

---

### 1.3 `tene get`

**목적:** 시크릿을 복호화하여 stdout에 출력한다. Claude Code가 Bash에서 `$(tene get KEY)`로 호출.

#### 시그니처

```
tene get <KEY>
```

| 인자/플래그 | 필수 | 설명 |
|-------------|:----:|------|
| `KEY` | 예 | 시크릿 키 이름 |
| `--env <name>` | 아니오 | 대상 환경 |
| `--json` | 아니오 | JSON 형식 출력 |

#### 정상 동작

stdout (순수 값만, 개행 포함):
```
sk_test_xxxxx
```

**중요:** stdout에는 순수 값만 출력한다. 레이블, 장식 없음. 개행(`\n`)은 값 뒤에 1개.

#### --json 출력 스키마

```json
{
  "ok": true,
  "name": "STRIPE_KEY",
  "value": "sk_test_xxxxx",
  "environment": "default"
}
```

#### 에러 시나리오

| 에러 | 메시지 (stderr) | 종료 코드 |
|------|-----------------|:---------:|
| 볼트 미초기화 | `Not in a Tene project. Run "tene init" first.` | 1 |
| 시크릿 없음 | `Secret "KEY" not found in "default" environment.` | 1 |
| 복호화 실패 | `Failed to decrypt secret. Master Password may have changed.` | 2 |
| Keychain 실패 | Master Password 프롬프트로 폴백 (비대화형이면 에러) | - |

#### 감사 로그

`secret.read` 액션으로 기록. resource_name = KEY.

---

### 1.4 `tene run`

**목적:** 현재 환경의 모든 시크릿을 환경변수로 주입한 자식 프로세스를 실행한다.

#### 시그니처

```
tene run [--env <name>] -- <command> [args...]
```

| 인자/플래그 | 필수 | 설명 |
|-------------|:----:|------|
| `--env <name>` | 아니오 | 대상 환경 |
| `-- <command>` | 예 | 실행할 명령어 + 인자 |

**`--` 구분자는 필수.** cobra의 `--` 이후를 raw args로 처리.

#### 정상 동작

```
$ tene run -- claude
  Injecting 5 secrets into environment...
  Starting: claude
```

`--quiet` 모드: 메시지 없이 바로 실행.

#### 구현 상세

```go
cmd := exec.Command(command, args...)
cmd.Env = append(os.Environ(), secrets...)  // 기존 환경변수 + 시크릿
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
err := cmd.Run()
os.Exit(cmd.ProcessState.ExitCode())
```

**핵심 규칙:**
- 부모 프로세스의 모든 환경변수를 상속
- 시크릿 환경변수를 추가 (동일 이름이면 시크릿 값이 우선)
- 자식 프로세스의 종료 코드를 그대로 반환
- stdin, stdout, stderr를 모두 패스스루
- 시크릿은 디스크에 평문으로 저장되지 않음 (메모리만)

#### --json 출력

`tene run`에는 `--json`이 적용되지 않음. 자식 프로세스의 출력을 그대로 전달해야 하므로.

다만 `--json` 전달 시 시작 정보만 stderr에 JSON으로 출력:
```json
{"injectedCount": 5, "environment": "default", "command": "claude"}
```

#### 에러 시나리오

| 에러 | 메시지 | 종료 코드 |
|------|--------|:---------:|
| 볼트 미초기화 | `Not in a Tene project. Run "tene init" first.` | 1 |
| 명령어 없음 | `No command specified. Usage: tene run -- <command>` | 1 |
| 명령어 못 찾음 | `Command "xyz" not found.` | 127 |
| 시크릿 0개 | 경고 출력 후 정상 실행: `Warning: No secrets found in "default". Running without injected secrets.` | 자식 종료 코드 |
| 복호화 실패 | `Failed to decrypt secrets.` | 2 |

---

### 1.5 `tene list`

**목적:** 현재 환경의 시크릿 목록을 표시한다. 값은 마스킹.

#### 시그니처

```
tene list [--env <name>]
```

#### 정상 동작

```
$ tene list
  Project: my-project (default)

  NAME              VALUE           UPDATED
  STRIPE_KEY        sk_te*****      2 minutes ago
  DATABASE_URL      postg*****      5 minutes ago
  API_SECRET        eyJhb*****      1 hour ago

  3 secrets in "default" environment
```

**마스킹 규칙:** 값의 처음 5자 + `*****`. 값이 5자 미만이면 전부 `*****`.

#### --json 출력 스키마

```json
{
  "ok": true,
  "project": "my-project",
  "environment": "default",
  "secrets": [
    {
      "name": "STRIPE_KEY",
      "preview": "sk_te*****",
      "version": 1,
      "updatedAt": "2026-04-06T12:00:00Z"
    }
  ],
  "count": 3
}
```

#### 에러 시나리오

| 에러 | 메시지 | 종료 코드 |
|------|--------|:---------:|
| 볼트 미초기화 | `Not in a Tene project. Run "tene init" first.` | 1 |
| 환경 없음 | `Environment "staging" not found.` | 1 |
| 시크릿 0개 | `No secrets in "default" environment. Use "tene set KEY VALUE" to add one.` | 0 |

---

### 1.6 `tene delete`

**목적:** 시크릿을 삭제한다.

#### 시그니처

```
tene delete <KEY> [--env <name>] [--force]
```

| 인자/플래그 | 필수 | 설명 |
|-------------|:----:|------|
| `KEY` | 예 | 삭제할 시크릿 키 이름 |
| `--env <name>` | 아니오 | 대상 환경 |
| `--force` | 아니오 | 확인 프롬프트 스킵 |

#### 대화형 동작

```
$ tene delete STRIPE_KEY
  Delete secret "STRIPE_KEY" from "default"? (y/N) y
  STRIPE_KEY deleted.
```

#### 비대화형 동작

`--force` 또는 비TTY 환경에서는 프롬프트 없이 즉시 삭제.

#### --json 출력 스키마

```json
{
  "ok": true,
  "name": "STRIPE_KEY",
  "environment": "default",
  "deleted": true
}
```

#### 에러 시나리오

| 에러 | 메시지 | 종료 코드 |
|------|--------|:---------:|
| 볼트 미초기화 | `Not in a Tene project. Run "tene init" first.` | 1 |
| 시크릿 없음 | `Secret "KEY" not found in "default" environment.` | 1 |
| 삭제 취소 | `Cancelled.` | 0 |

---

### 1.7 `tene import`

**목적:** .env 파일 또는 암호화된 백업 파일에서 시크릿을 일괄 가져온다.

#### 시그니처

```
tene import <FILE> [--env <name>] [--overwrite] [--encrypted]
```

| 인자/플래그 | 필수 | 설명 |
|-------------|:----:|------|
| `FILE` | 예 | 가져올 파일 경로 (.env 또는 .tene.enc) |
| `--env <name>` | 아니오 | 대상 환경 |
| `--overwrite` | 아니오 | 기존 시크릿 덮어쓰기 |
| `--encrypted` | 아니오 | 암호화된 백업 파일로부터 복원 |

#### .env 파일 파싱 규칙

```
# 지원하는 형식
KEY=VALUE
KEY="VALUE WITH SPACES"
KEY='VALUE WITH SPACES'
export KEY=VALUE

# 무시
# 주석 줄
빈 줄
```

#### 정상 동작 (.env 가져오기)

```
$ tene import .env
  Found 5 secrets in .env:
    STRIPE_KEY, DATABASE_URL, API_SECRET, SENDGRID_KEY, JWT_SECRET

  Import 5 secrets to "my-project" (default)? (y/N) y
  5 secrets imported (encrypted).

  Tip: You can now delete .env and use tene run instead.
```

#### 정상 동작 (암호화 백업 복원)

```
$ tene import --encrypted my-project.tene.enc
  Enter Master Password: ********
  5 secrets restored to "my-project" vault.
```

#### --json 출력 스키마

```json
{
  "ok": true,
  "file": ".env",
  "environment": "default",
  "imported": 5,
  "skipped": 0,
  "overwritten": 0,
  "secrets": ["STRIPE_KEY", "DATABASE_URL", "API_SECRET", "SENDGRID_KEY", "JWT_SECRET"]
}
```

#### 에러 시나리오

| 에러 | 메시지 | 종료 코드 |
|------|--------|:---------:|
| 볼트 미초기화 | `Not in a Tene project. Run "tene init" first.` | 1 |
| 파일 없음 | `File ".env" not found.` | 1 |
| 파싱 실패 | `Failed to parse ".env" at line 3: invalid format.` | 1 |
| 기존 키 충돌 | `Secret "KEY" already exists. Use --overwrite to replace.` | 1 |
| 암호화 파일 복호화 실패 | `Failed to decrypt backup file. Wrong Master Password?` | 2 |
| 잘못된 파일 포맷 | `Invalid encrypted backup file format.` | 1 |

---

### 1.8 `tene export`

**목적:** 시크릿을 .env 형식 또는 암호화된 백업 파일로 내보낸다.

#### 시그니처

```
tene export [--env <name>] [--file <path>] [--encrypted]
```

| 인자/플래그 | 필수 | 설명 |
|-------------|:----:|------|
| `--env <name>` | 아니오 | 대상 환경 |
| `--file <path>` | 아니오 | 출력 파일 경로. 생략 시 stdout |
| `--encrypted` | 아니오 | 암호화된 백업 파일 생성 |

#### 정상 동작 (.env 형식, stdout)

```
$ tene export
STRIPE_KEY=sk_test_xxxxx
DATABASE_URL=postgresql://user:pass@host/db
API_SECRET=eyJhbGciOiJIUzI1NiJ9
```

**중요:** stdout 출력 시 레이블, 장식 없이 순수 .env 형식만. 리디렉션 가능:

```bash
tene export > .env.local
```

#### 정상 동작 (.env 형식, 파일 지정)

```
$ tene export --file .env.local
  5 secrets exported to .env.local
  Warning: This file contains plain-text secrets. Do not commit it.
```

#### 정상 동작 (암호화 백업)

```
$ tene export --encrypted
  Encrypted vault exported to: ./my-project.tene.enc

  This file is encrypted with your Master Password.
  To restore: tene import --encrypted my-project.tene.enc

  Store this file in a safe place (USB, cloud drive, etc.)
```

`--encrypted`이면서 `--file` 미지정 시: `{project-name}.tene.enc` 파일명 자동 생성.
`--encrypted`이면서 `--file` 지정 시: 해당 경로에 저장.

**주의:** `--encrypted`를 stdout으로 파이프하는 것은 바이너리 데이터이므로 금지. 반드시 파일로 출력.

#### --json 출력 스키마 (.env 형식)

```json
{
  "ok": true,
  "environment": "default",
  "count": 5,
  "secrets": {
    "STRIPE_KEY": "sk_test_xxxxx",
    "DATABASE_URL": "postgresql://user:pass@host/db"
  }
}
```

--json + --encrypted:
```json
{
  "ok": true,
  "environment": "default",
  "file": "./my-project.tene.enc",
  "encrypted": true,
  "count": 5
}
```

#### 에러 시나리오

| 에러 | 메시지 | 종료 코드 |
|------|--------|:---------:|
| 볼트 미초기화 | `Not in a Tene project. Run "tene init" first.` | 1 |
| 시크릿 0개 | `No secrets to export in "default" environment.` | 1 |
| 파일 쓰기 실패 | `Cannot write to "path": permission denied.` | 1 |
| 복호화 실패 | `Failed to decrypt secrets.` | 2 |

---

### 1.9 `tene env`

**목적:** 환경(dev, staging, prod 등)을 관리한다.

#### 시그니처

```
tene env                        # 현재 환경 표시 + 목록
tene env <name>                 # 환경 전환
tene env list                   # 환경 목록
tene env create <name>          # 새 환경 생성
tene env delete <name>          # 환경 삭제
```

#### 서브커맨드 상세

**`tene env` (인자 없음) / `tene env list`**

```
$ tene env
  Environments:
  * default (active, 3 secrets)
    dev (5 secrets)
    prod (8 secrets)
```

**`tene env <name>` (환경 전환)**

```
$ tene env prod
  Switched to "prod" environment.
```

활성 환경은 `.tene/vault.json`의 `activeEnvironment` 필드에 저장.

**`tene env create <name>`**

```
$ tene env create staging
  Environment "staging" created.
```

환경 이름 규칙: `^[a-z][a-z0-9-]*$`, 최소 1자, 최대 64자.

**`tene env delete <name>`**

```
$ tene env delete staging
  Delete environment "staging" and all its secrets? (y/N) y
  Environment "staging" deleted (2 secrets removed).
```

- `default` 환경은 삭제 불가
- 현재 활성 환경은 삭제 불가 (다른 환경으로 전환 먼저)

#### --json 출력 스키마

`tene env` / `tene env list`:
```json
{
  "ok": true,
  "active": "default",
  "environments": [
    {"name": "default", "secretCount": 3, "isActive": true},
    {"name": "dev", "secretCount": 5, "isActive": false},
    {"name": "prod", "secretCount": 8, "isActive": false}
  ]
}
```

`tene env <name>`:
```json
{
  "ok": true,
  "previous": "default",
  "current": "prod"
}
```

`tene env create <name>`:
```json
{
  "ok": true,
  "name": "staging",
  "created": true
}
```

#### 에러 시나리오

| 에러 | 메시지 | 종료 코드 |
|------|--------|:---------:|
| 볼트 미초기화 | `Not in a Tene project. Run "tene init" first.` | 1 |
| 환경 없음 | `Environment "staging" not found. Create it with "tene env create staging".` | 1 |
| 이미 존재 | `Environment "dev" already exists.` | 1 |
| default 삭제 시도 | `Cannot delete the "default" environment.` | 1 |
| 활성 환경 삭제 시도 | `Cannot delete the active environment. Switch to another first.` | 1 |
| 잘못된 이름 | `Invalid environment name. Must match [a-z][a-z0-9-]*.` | 1 |

---

### 1.10 `tene passwd`

**목적:** Master Password를 변경하고, 볼트를 재암호화하며, 새 Recovery Key를 발급한다.

#### 시그니처

```
tene passwd
```

대화형 전용 명령어. 비대화형 환경에서는 사용 불가.

#### 정상 동작

```
$ tene passwd
  Enter current Master Password: ********
  Enter new Master Password: ********
  Confirm new Master Password: ********

  Re-encrypting vault...
  5 secrets re-encrypted.
  Master Key updated in OS Keychain.

  New Recovery Key (write this down and keep it safe!):
  +--------------------------------------------------+
  |   mango night ocean purple quiet river             |
  |   sunset tiger umbrella violet winter xray         |
  |                                                    |
  |   Your previous Recovery Key is now invalid.       |
  +--------------------------------------------------+

  Master Password changed successfully.
```

#### 수행 단계

1. 현재 Master Password 확인 (OS Keychain의 Master Key로 검증 또는 직접 KDF)
2. 새 Master Password 입력 + 확인 (최소 8자)
3. 새 salt 생성
4. 새 Master Key 유도 (Argon2id)
5. 새 Encryption Key 파생
6. 모든 시크릿을 이전 키로 복호화 -> 새 키로 재암호화
7. vault_meta 업데이트 (kdf_salt, kdf_params)
8. 새 Recovery Key 생성 -> 새 recovery_blob 저장
9. 이전 recovery_blob 삭제
10. OS Keychain 업데이트
11. 감사 로그 기록: `vault.passwd_changed`

#### --json 출력 스키마

```json
{
  "ok": true,
  "reEncrypted": 5,
  "recoveryKey": "mango night ocean purple quiet river sunset tiger umbrella violet winter xray"
}
```

#### 에러 시나리오

| 에러 | 메시지 | 종료 코드 |
|------|--------|:---------:|
| 볼트 미초기화 | `Not in a Tene project. Run "tene init" first.` | 1 |
| 현재 비밀번호 오류 | `Invalid current Master Password.` | 2 |
| 새 비밀번호 불일치 | `New passwords do not match.` | 2 |
| 재암호화 실패 | `Re-encryption failed. Vault is unchanged.` | 1 |
| 비대화형 환경 | `tene passwd requires an interactive terminal.` | 1 |

---

### 1.11 `tene recover`

**목적:** Recovery Key(12단어)로 Master Password를 재설정한다.

#### 시그니처

```
tene recover
```

대화형 전용 명령어.

#### 정상 동작

```
$ tene recover
  Enter Recovery Key (12 words): apple banana cherry dolphin eagle frost grape harbor island jungle kite lemon
  Enter new Master Password: ********
  Confirm new Master Password: ********

  Master Password reset successfully!
  Re-encrypting vault...
  5 secrets re-encrypted.

  New Recovery Key (write this down and keep it safe!):
  +--------------------------------------------------+
  |   mango night ocean purple quiet river             |
  |   sunset tiger umbrella violet winter xray         |
  |                                                    |
  |   Your previous Recovery Key is now invalid.       |
  +--------------------------------------------------+
```

#### 수행 단계

1. Recovery Key (12단어) 입력받기
2. Recovery Key -> Argon2id -> Recovery Encryption Key 유도
3. vault_meta에서 recovery_blob 읽기
4. recovery_blob 복호화 -> 이전 Master Key 복원
5. 새 Master Password 입력 + 확인
6. 새 salt + 새 Master Key 유도
7. 모든 시크릿을 이전 키로 복호화 -> 새 키로 재암호화
8. 새 Recovery Key 생성 + 새 recovery_blob 저장
9. OS Keychain 업데이트
10. 감사 로그 기록: `vault.recovered`

#### 에러 시나리오

| 에러 | 메시지 | 종료 코드 |
|------|--------|:---------:|
| 볼트 미초기화 | `Not in a Tene project. Run "tene init" first.` | 1 |
| Recovery Key 오류 | `Invalid Recovery Key.` | 2 |
| 복호화 실패 | `Recovery failed. The Recovery Key may be incorrect.` | 2 |
| 비대화형 환경 | `tene recover requires an interactive terminal.` | 1 |

---

### 1.12 `tene whoami`

**목적:** 현재 볼트 상태를 표시한다.

#### 시그니처

```
tene whoami
```

#### 정상 동작

```
$ tene whoami
  Project: my-project
  Vault: .tene/vault.db
  Environment: default (active)
  Secrets: 5
  Keychain: macOS Keychain (active)
  Created: 2026-04-06
  Agents: claude
```

#### --json 출력 스키마

```json
{
  "ok": true,
  "project": "my-project",
  "vault": ".tene/vault.db",
  "environment": "default",
  "secretCount": 5,
  "keychainStatus": "active",
  "keychainProvider": "macOS Keychain",
  "createdAt": "2026-04-06T12:00:00Z",
  "agents": ["claude"],
  "vaultVersion": 1
}
```

#### 에러 시나리오

| 에러 | 메시지 | 종료 코드 |
|------|--------|:---------:|
| 볼트 미초기화 | `Not in a Tene project. Run "tene init" first.` | 1 |

---

### 1.13 `tene sync` (Fake Door)

**목적:** Cloud 동기화 수요를 검증하는 Fake Door. 실제 동기화는 미구현.

#### 시그니처

```
tene sync
```

#### 정상 동작

```
$ tene sync

  Tene Cloud Sync -- Coming Soon!

  Cloud sync will enable:
  - Multi-device secret synchronization
  - Encrypted cloud backup (zero-knowledge)
  - Web dashboard for secret overview
  - All for just $1/month

  Join the waitlist to get early access:
  -> https://tene.sh/waitlist

  In the meantime, use `tene export --encrypted` for local backup.

  [Open waitlist page? (Y/n)]
```

#### 구현 상세

1. 화면 출력
2. 대화형이면 `Open waitlist page?` 프롬프트
   - `Y`: `open https://tene.sh/waitlist` (macOS) 또는 `xdg-open` (Linux)
   - `N`: 종료
3. `~/.tene/config.json`에 analytics 기록:
   - `syncAttempts` += 1
   - `lastSyncAttempt` = 현재 시각 (ISO 8601)

#### --json 출력 스키마

```json
{
  "ok": true,
  "message": "cloud_sync_not_available",
  "waitlistUrl": "https://tene.sh/waitlist",
  "tip": "Use tene export --encrypted for local backup."
}
```

**종료 코드:** 항상 0.

---

### 1.14 `tene version`

**목적:** CLI 버전 정보를 출력한다.

#### 시그니처

```
tene version
tene --version
tene -v
```

#### 정상 동작

```
$ tene version
tene v0.1.0 (darwin/arm64)
```

형식: `tene v{version} ({os}/{arch})`

#### --json 출력 스키마

```json
{
  "version": "0.1.0",
  "os": "darwin",
  "arch": "arm64",
  "commit": "abc1234",
  "buildDate": "2026-04-06T12:00:00Z"
}
```

**종료 코드:** 항상 0.

---

## 2. 비기능 요구사항 (Non-Functional Requirements)

### 2.1 성능

| 항목 | 목표 | 측정 방법 | 비고 |
|------|------|----------|------|
| CLI cold start | < 20ms (P95) | `time tene version` | Go 바이너리 자연 성능 |
| `tene get` 응답 | < 100ms (P95) | `time tene get KEY` | Keychain 조회 + SQLite + 복호화 포함 |
| `tene set` 응답 | < 100ms (P95) | `time tene set KEY VALUE` | 암호화 + SQLite 쓰기 포함 |
| `tene list` 응답 | < 100ms (P95) | 1,000개 시크릿 기준 | SQLite 인덱스 활용 |
| `tene run` 오버헤드 | < 50ms | 시크릿 주입까지의 지연 | 자식 프로세스 시작 시간 제외 |
| Argon2id KDF | < 500ms | Master Password -> Master Key | 64MB 메모리, 3 iterations |
| CLAUDE.md 생성 | < 10ms | `tene init` 내부 | 파일 I/O만 |
| SQLite 쿼리 | < 5ms | 단일 시크릿 조회 | 인덱스 활용 |

### 2.2 보안

| 항목 | 요구사항 | 구현 |
|------|----------|------|
| 메모리 시크릿 제거 | 시크릿 값 사용 후 메모리에서 즉시 zero-fill | `defer memguard.WipeBytes(secret)` 패턴 또는 수동 zeroing |
| 파일 퍼미션 | `.tene/` = 0700, `vault.db` = 0600, `config.json` = 0600 | `os.MkdirAll` / `os.OpenFile` with mode |
| 환경변수 누출 방지 | `tene run` 이후 부모 프로세스에 시크릿 환경변수 남지 않음 | 자식 프로세스에만 설정 |
| 에러 메시지 | 시크릿 값이 에러 메시지에 절대 포함되지 않음 | 키 이름만 로그 |
| Core dump 방지 | `RLIMIT_CORE = 0` 설정 (가능 시) | `syscall.Setrlimit` |
| Shell history | `--stdin` 플래그로 shell history 노출 방지 | 문서에서 권장 |
| .gitignore | `.tene/` 자동 추가 | `tene init`에서 보장 |
| Master Password 정책 | 최소 8자 | `tene init`, `tene passwd`, `tene recover` |

### 2.3 호환성

| OS | 아키텍처 | 지원 | 비고 |
|----|---------|:----:|------|
| macOS 12+ (Monterey) | arm64 (Apple Silicon) | O | 주요 타겟 |
| macOS 12+ | amd64 (Intel) | O | goreleaser |
| Ubuntu 20.04+ | amd64 | O | goreleaser |
| Ubuntu 20.04+ | arm64 | O | goreleaser |
| Debian 11+ | amd64/arm64 | O | 호환 |
| Windows 10+ | WSL (amd64/arm64) | O | curl 설치 |
| Alpine Linux | amd64 | O | Docker/CI 환경 |

### 2.4 바이너리 크기

| 항목 | 목표 | 비고 |
|------|------|------|
| 비압축 바이너리 | < 20MB | modernc.org/sqlite 포함 |
| ldflags 최적화 | `-s -w` 적용 | goreleaser 기본 |
| CGO_ENABLED | 0 (필수) | 순수 Go 빌드 |

### 2.5 오프라인

- 모든 CLI 명령어는 네트워크 연결 없이 100% 동작
- DNS 조회, HTTP 요청 등 네트워크 I/O 코드가 MVP에 없어야 함
- 예외: `tene sync`의 브라우저 열기 (사용자 동의 후)

---

## 3. 데이터 모델 상세

### 3.1 SQLite 스키마 (.tene/vault.db)

```sql
-- 볼트 메타데이터 (키-값 구조)
CREATE TABLE vault_meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- 저장되는 키:
--   vault_version   = "1"
--   created_at      = "2026-04-06T12:00:00Z" (ISO 8601)
--   kdf_salt        = base64 인코딩된 128-bit salt
--   kdf_params      = JSON {"algorithm":"argon2id","memory":65536,"iterations":3,"parallelism":1,"keyLen":32}
--   recovery_blob   = base64 인코딩된 [nonce(24) + encrypted_master_key]

-- 환경 관리
CREATE TABLE environments (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL UNIQUE,
    is_active   INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- 암호화된 시크릿
CREATE TABLE secrets (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    environment_id   INTEGER NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    name             TEXT    NOT NULL,
    encrypted_value  TEXT    NOT NULL,  -- base64(nonce[24] + ciphertext)
    version          INTEGER NOT NULL DEFAULT 1,
    created_at       TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at       TEXT    NOT NULL DEFAULT (datetime('now')),
    UNIQUE(environment_id, name)
);

-- 로컬 감사 로그
CREATE TABLE audit_log (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    action         TEXT NOT NULL,
    resource_name  TEXT,
    environment    TEXT,
    source         TEXT NOT NULL DEFAULT 'cli',
    timestamp      TEXT NOT NULL DEFAULT (datetime('now'))
);

-- 인덱스
CREATE INDEX idx_secrets_env_name ON secrets(environment_id, name);
CREATE INDEX idx_audit_timestamp ON audit_log(timestamp DESC);
```

**감사 로그 action 값:**

| action | 설명 |
|--------|------|
| `vault.init` | 볼트 초기화 |
| `vault.passwd_changed` | Master Password 변경 |
| `vault.recovered` | Recovery Key로 복구 |
| `secret.write` | 시크릿 생성/수정 |
| `secret.read` | 시크릿 조회 |
| `secret.delete` | 시크릿 삭제 |
| `secret.import` | 시크릿 일괄 가져오기 |
| `secret.export` | 시크릿 내보내기 |
| `secret.export_encrypted` | 암호화 백업 내보내기 |
| `env.create` | 환경 생성 |
| `env.delete` | 환경 삭제 |
| `env.switch` | 환경 전환 |

### 3.2 파일 시스템 구조

```
~/.tene/                                 # 글로벌 CLI 설정 (퍼미션 0700)
  config.json                            # 글로벌 설정 (퍼미션 0600)

project/.tene/                           # 프로젝트 볼트 (퍼미션 0700)
  vault.db                               # SQLite 볼트 (퍼미션 0600)
  vault.json                             # 볼트 메타데이터 (퍼미션 0600)
  .gitignore                             # 내용: * (모든 파일 무시)

project/CLAUDE.md                        # Claude Code 컨텍스트 (tene init이 생성)
```

### 3.3 `~/.tene/config.json` 스키마

```json
{
  "version": 1,
  "defaultEnvironment": "default",
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

### 3.4 `.tene/vault.json` 스키마

```json
{
  "projectName": "my-project",
  "createdAt": "2026-04-06T12:00:00Z",
  "vaultVersion": 1,
  "activeEnvironment": "default",
  "agents": ["claude"]
}
```

### 3.5 `.tene.enc` 파일 포맷 (암호화 백업)

바이너리 파일 구조:

```
Offset  Size     Field                 설명
------  -------  --------------------  ----------------------------------
0       4        Magic                 "TENE" (0x54454E45)
4       1        Format Version        0x01
5       1        KDF Algorithm         0x01 = Argon2id
6       4        KDF Memory (KB)       65536 (64MB) - little-endian uint32
10      4        KDF Iterations        3 - little-endian uint32
14      1        KDF Parallelism       1
15      1        Salt Length           16
16      16       KDF Salt              128-bit random salt
32      24       Nonce                 192-bit random nonce (XChaCha20)
56      *        Encrypted Payload     XChaCha20-Poly1305 암호화된 볼트 데이터
                                       (JSON 직렬화된 시크릿 배열)
```

암호화된 Payload의 평문 형식 (JSON):

```json
{
  "exportVersion": 1,
  "exportedAt": "2026-04-06T12:00:00Z",
  "projectName": "my-project",
  "environments": [
    {
      "name": "default",
      "secrets": [
        {
          "name": "STRIPE_KEY",
          "value": "sk_test_xxxxx",
          "version": 1,
          "createdAt": "2026-04-06T12:00:00Z",
          "updatedAt": "2026-04-06T12:00:00Z"
        }
      ]
    }
  ]
}
```

---

## 4. 암호화 명세

### 4.1 Argon2id KDF 파라미터 (확정값)

| 파라미터 | 값 | 근거 |
|----------|-----|------|
| Algorithm | Argon2id | OWASP 권장, 사이드채널 공격 방어 |
| Memory | 64MB (65536 KB) | OWASP 최소 권장 |
| Iterations | 3 | OWASP 권장 |
| Parallelism | 1 | 싱글 코어 CLI |
| Key Length | 32 bytes (256-bit) | XChaCha20 키 크기 |
| Salt Length | 16 bytes (128-bit) | 충분한 엔트로피 |

**Go 구현:**

```go
import "golang.org/x/crypto/argon2"

func DeriveKey(password string, salt []byte) []byte {
    return argon2.IDKey(
        []byte(password),
        salt,
        3,        // iterations
        64*1024,  // 64MB memory
        1,        // parallelism
        32,       // 256-bit output
    )
}
```

### 4.2 XChaCha20-Poly1305 사용법

| 파라미터 | 값 | 설명 |
|----------|-----|------|
| Algorithm | XChaCha20-Poly1305 | AEAD 암호화 |
| Key Size | 32 bytes (256-bit) | Encryption Key |
| Nonce Size | 24 bytes (192-bit) | Random nonce, 재사용 위험 극소 |
| AAD | 시크릿 키 이름 (UTF-8) | 키 이름 변조 방지 |
| Tag Size | 16 bytes (128-bit) | Poly1305 인증 태그 (자동) |

**Go 구현:**

```go
import (
    "crypto/rand"
    "io"
    "golang.org/x/crypto/nacl/secretbox"
)

func Encrypt(plaintext []byte, key [32]byte, aad string) ([]byte, error) {
    var nonce [24]byte
    if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
        return nil, err
    }

    // secretbox.Seal은 nonce를 out 앞에 붙임
    encrypted := secretbox.Seal(nonce[:], plaintext, &nonce, &key)
    return encrypted, nil  // [nonce(24) + ciphertext + tag(16)]
}

func Decrypt(blob []byte, key [32]byte, aad string) ([]byte, error) {
    if len(blob) < 24 + secretbox.Overhead {
        return nil, ErrInvalidCiphertext
    }

    var nonce [24]byte
    copy(nonce[:], blob[:24])

    plaintext, ok := secretbox.Open(nil, blob[24:], &nonce, &key)
    if !ok {
        return nil, ErrDecryptFailed
    }
    return plaintext, nil
}
```

**참고:** `nacl/secretbox`는 XSalsa20-Poly1305를 사용한다. XChaCha20-Poly1305를 사용하려면 `golang.org/x/crypto/chacha20poly1305` 패키지의 `NewX()`를 사용해야 한다.

```go
import "golang.org/x/crypto/chacha20poly1305"

func Encrypt(plaintext, key, aad []byte) ([]byte, error) {
    aead, err := chacha20poly1305.NewX(key)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, aead.NonceSize()) // 24 bytes
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    // nonce를 앞에 붙여서 반환
    ciphertext := aead.Seal(nonce, nonce, plaintext, aad)
    return ciphertext, nil  // [nonce(24) + ciphertext + tag(16)]
}

func Decrypt(blob, key, aad []byte) ([]byte, error) {
    aead, err := chacha20poly1305.NewX(key)
    if err != nil {
        return nil, err
    }

    nonceSize := aead.NonceSize()
    if len(blob) < nonceSize + aead.Overhead() {
        return nil, ErrInvalidCiphertext
    }

    nonce, ciphertext := blob[:nonceSize], blob[nonceSize:]
    plaintext, err := aead.Open(nil, nonce, ciphertext, aad)
    if err != nil {
        return nil, ErrDecryptFailed
    }
    return plaintext, nil
}
```

### 4.3 키 파생 플로우

```
Master Password (사용자 입력, 최소 8자)
    |
    v
[Argon2id KDF]
  - salt: vault_meta.kdf_salt (128-bit random, 볼트 생성 시 고정)
  - memory: 64MB
  - iterations: 3
  - parallelism: 1
  - output: 32 bytes
    |
    v
Master Key (256-bit)
    |
    +--- [HKDF-SHA256 Expand]
    |    info: "tene-encryption-key"
    |    output: 32 bytes
    |    --> Encryption Key (시크릿 암호화/복호화용)
    |
    +--- [HKDF-SHA256 Expand]
         info: "tene-auth-hash"
         output: 32 bytes
         --> Auth Hash (Phase 2: Cloud 인증용, MVP에서는 미사용)
```

**Go 구현:**

```go
import (
    "crypto/sha256"
    "golang.org/x/crypto/hkdf"
)

func DeriveEncryptionKey(masterKey []byte) ([]byte, error) {
    r := hkdf.Expand(sha256.New, masterKey, []byte("tene-encryption-key"))
    key := make([]byte, 32)
    if _, err := io.ReadFull(r, key); err != nil {
        return nil, err
    }
    return key, nil
}
```

### 4.4 Recovery Key 플로우

#### 생성 (tene init 시)

```
1. 128-bit entropy 생성 (crypto/rand)
2. BIP-39 니모닉 생성 (12단어)
   예: "apple banana cherry dolphin eagle frost grape harbor island jungle kite lemon"

3. 니모닉에서 Recovery Encryption Key 유도:
   recoveryKey = Argon2id(
     password: 니모닉 문자열,
     salt: "tene-recovery" (고정 salt),
     memory: 64MB, iterations: 3, parallelism: 1,
     output: 32 bytes
   )

4. Recovery Encryption Key로 Master Key 암호화:
   recovery_blob = XChaCha20-Poly1305.Seal(
     key: recoveryKey,
     plaintext: masterKey,
     aad: "tene-recovery-blob"
   )

5. recovery_blob를 vault_meta에 base64로 저장
```

#### 검증/복구 (tene recover 시)

```
1. 사용자가 12단어 니모닉 입력
2. BIP-39 유효성 검증 (체크섬 확인)
3. 니모닉에서 Recovery Encryption Key 유도 (생성과 동일 과정)
4. vault_meta에서 recovery_blob 읽기
5. Recovery Encryption Key로 recovery_blob 복호화 -> Master Key 복원
6. 복원된 Master Key로 시크릿 접근 가능 확인
7. 새 Master Password -> 새 Master Key -> 볼트 재암호화
8. 새 Recovery Key 생성 -> 새 recovery_blob 저장
```

### 4.5 salt 생성 및 저장

| salt 용도 | 크기 | 생성 시점 | 저장 위치 | 변경 시점 |
|-----------|------|----------|----------|----------|
| KDF Salt | 16 bytes (128-bit) | `tene init` | `vault_meta.kdf_salt` (base64) | `tene passwd`, `tene recover` |
| Nonce (시크릿별) | 24 bytes (192-bit) | `tene set` (매번 새로) | `secrets.encrypted_value` 앞 24바이트 | 매 암호화 시 |
| Recovery Nonce | 24 bytes (192-bit) | `tene init` | `vault_meta.recovery_blob` 앞 24바이트 | `tene passwd`, `tene recover` |

모든 salt/nonce는 `crypto/rand.Reader`에서 생성.

---

## 5. OS Keychain 연동 명세

### 5.1 go-keyring 사용

**라이브러리:** `github.com/zalando/go-keyring`

| 필드 | 값 |
|------|-----|
| Service | `"tene"` |
| Account | 프로젝트 디렉토리의 절대 경로 해시 (SHA-256 hex, 앞 16자) |
| Password | Master Key (32 bytes, base64 인코딩) |

**Go 구현:**

```go
import "github.com/zalando/go-keyring"

const keychainService = "tene"

func keychainAccount(projectDir string) string {
    h := sha256.Sum256([]byte(projectDir))
    return hex.EncodeToString(h[:8])  // 앞 16 hex 문자
}

func SaveMasterKey(projectDir string, masterKey []byte) error {
    account := keychainAccount(projectDir)
    encoded := base64.StdEncoding.EncodeToString(masterKey)
    return keyring.Set(keychainService, account, encoded)
}

func LoadMasterKey(projectDir string) ([]byte, error) {
    account := keychainAccount(projectDir)
    encoded, err := keyring.Get(keychainService, account)
    if err != nil {
        return nil, err
    }
    return base64.StdEncoding.DecodeString(encoded)
}

func DeleteMasterKey(projectDir string) error {
    account := keychainAccount(projectDir)
    return keyring.Delete(keychainService, account)
}
```

### 5.2 OS별 백엔드

| OS | Keychain 백엔드 | 라이브러리 내부 동작 |
|----|-----------------|---------------------|
| macOS | Keychain Services | Security.framework (SecItemAdd/Copy/Delete) |
| Linux | Secret Service (D-Bus) | GNOME Keyring / KDE Wallet |
| Windows | Windows Credential Manager | wincred API |

### 5.3 폴백 메커니즘

Keychain 사용이 불가능한 경우 (headless 서버, Docker 컨테이너 등):

**1단계: `TENE_MASTER_PASSWORD` 환경변수 확인**

```go
if pw := os.Getenv("TENE_MASTER_PASSWORD"); pw != "" {
    masterKey := DeriveKey(pw, salt)
    return masterKey, nil
}
```

**2단계: `--no-keychain` 플래그인 경우 또는 Keychain 에러**

```
Keychain 호출 실패 시:
  1. TENE_MASTER_PASSWORD 환경변수 확인 -> 있으면 사용
  2. TTY이면 Master Password 직접 입력 프롬프트
  3. 비TTY이면 에러: "Cannot access OS Keychain. Set TENE_MASTER_PASSWORD or use --no-keychain."
```

**폴백 우선순위:**

```
1. OS Keychain (기본)
     |-- 성공 --> Master Key 반환
     |-- 실패 --|
                v
2. TENE_MASTER_PASSWORD 환경변수
     |-- 존재 --> Argon2id -> Master Key 반환
     |-- 없음 --|
                v
3. 대화형 프롬프트 (TTY일 때만)
     |-- 입력 --> Argon2id -> Master Key 반환
     |-- 비TTY --|
                  v
4. 에러 반환
```

---

## 6. CI/CD 환경 대응

### 6.1 비대화형 모드

CI/CD 환경에서는 대화형 프롬프트 없이 동작해야 한다.

**필수 환경변수:**

| 환경변수 | 설명 |
|----------|------|
| `TENE_MASTER_PASSWORD` | Master Password (Argon2id로 Master Key 유도) |

**CI/CD 사용 예시:**

```yaml
# GitHub Actions
env:
  TENE_MASTER_PASSWORD: ${{ secrets.TENE_MASTER_PASSWORD }}

steps:
  - name: Inject secrets and run tests
    run: tene run -- npm test
```

```bash
# Docker
docker run -e TENE_MASTER_PASSWORD=mysecret \
  -v $(pwd)/.tene:/app/.tene \
  myapp:latest tene run -- node server.js
```

### 6.2 `--no-keychain` 플래그

OS Keychain을 완전히 비활성화. Keychain 접근 시도 자체를 하지 않음.

```bash
TENE_MASTER_PASSWORD=mysecret tene get STRIPE_KEY --no-keychain
```

### 6.3 Docker 환경 대응

Docker 컨테이너 내에서는 OS Keychain이 없으므로:

1. `TENE_MASTER_PASSWORD` 환경변수 사용 (권장)
2. 자동으로 Keychain 폴백 동작

```dockerfile
FROM alpine:latest
COPY tene /usr/local/bin/tene
COPY .tene/ /app/.tene/
WORKDIR /app
ENV TENE_MASTER_PASSWORD=${TENE_MASTER_PASSWORD}
CMD ["tene", "run", "--", "node", "server.js"]
```

### 6.4 비대화형 감지

```go
func IsInteractive() bool {
    return isatty.IsTerminal(os.Stdin.Fd()) ||
           isatty.IsCygwinTerminal(os.Stdin.Fd())
}
```

비대화형 환경에서 대화형 전용 명령어(`tene passwd`, `tene recover`) 호출 시:
- 종료 코드 1
- stderr: `tene passwd requires an interactive terminal.`

---

## 7. 에러 코드 체계

### 7.1 종료 코드 (Exit Code)

| 코드 | 의미 | 설명 |
|:----:|------|------|
| 0 | 성공 | 정상 완료 |
| 1 | 일반 에러 | 볼트 미초기화, 시크릿 없음, 파일 없음, 권한 에러 등 |
| 2 | 인증 에러 | Master Password 오류, Recovery Key 오류, 복호화 실패 |
| 127 | 명령어 없음 | `tene run -- xyz`에서 xyz가 PATH에 없음 |

`tene run`의 경우: 자식 프로세스의 종료 코드를 그대로 반환.

### 7.2 `--json` 에러 응답 형식

```json
{
  "ok": false,
  "error": "ERROR_CODE",
  "message": "사람이 읽을 수 있는 에러 메시지"
}
```

### 7.3 에러 코드 목록

| 에러 코드 | 종료 코드 | 설명 |
|-----------|:---------:|------|
| `VAULT_NOT_FOUND` | 1 | `.tene/vault.db`가 없음. `tene init` 필요 |
| `VAULT_ALREADY_EXISTS` | 0 | 이미 초기화됨 (경고, 에러는 아님) |
| `SECRET_NOT_FOUND` | 1 | 지정한 키의 시크릿이 없음 |
| `SECRET_ALREADY_EXISTS` | 1 | 시크릿이 이미 존재. `--overwrite` 필요 |
| `ENVIRONMENT_NOT_FOUND` | 1 | 지정한 환경이 없음 |
| `ENVIRONMENT_ALREADY_EXISTS` | 1 | 환경이 이미 존재 |
| `ENVIRONMENT_PROTECTED` | 1 | default 환경 삭제 또는 활성 환경 삭제 시도 |
| `INVALID_KEY_NAME` | 1 | 키 이름이 `^[A-Z][A-Z0-9_]*$`에 맞지 않음 |
| `INVALID_ENV_NAME` | 1 | 환경 이름이 `^[a-z][a-z0-9-]*$`에 맞지 않음 |
| `EMPTY_VALUE` | 1 | 시크릿 값이 비어 있음 |
| `VALUE_TOO_LARGE` | 1 | 시크릿 값이 64KB 초과 |
| `PASSWORD_MISMATCH` | 2 | 패스워드 확인 불일치 |
| `PASSWORD_TOO_SHORT` | 2 | Master Password가 8자 미만 |
| `INVALID_PASSWORD` | 2 | Master Password가 올바르지 않음 |
| `INVALID_RECOVERY_KEY` | 2 | Recovery Key가 올바르지 않음 |
| `DECRYPT_FAILED` | 2 | 복호화 실패 (키 변경 등) |
| `ENCRYPT_FAILED` | 1 | 암호화 실패 (내부 오류) |
| `FILE_NOT_FOUND` | 1 | 지정한 파일이 없음 |
| `FILE_PARSE_ERROR` | 1 | 파일 파싱 실패 (.env 형식 오류) |
| `PERMISSION_DENIED` | 1 | 파일/디렉토리 권한 부족 |
| `DISK_FULL` | 1 | 디스크 공간 부족 |
| `KEYCHAIN_ERROR` | 0 | Keychain 접근 실패 (폴백 진행) |
| `COMMAND_NOT_FOUND` | 127 | `tene run`에서 명령어를 찾을 수 없음 |
| `INTERACTIVE_REQUIRED` | 1 | 대화형 전용 명령어를 비TTY에서 실행 |
| `INVALID_BACKUP_FILE` | 1 | 암호화 백업 파일의 형식이 올바르지 않음 |
| `RESERVED_KEY_NAME` | 1 | 예약된 키 이름 사용 시도 |

### 7.4 Go 커스텀 에러 타입

```go
package errors

import "fmt"

type TeneError struct {
    Code    string // 에러 코드 (예: "VAULT_NOT_FOUND")
    Message string // 사용자 표시 메시지
    Exit    int    // 종료 코드 (0, 1, 2)
}

func (e *TeneError) Error() string {
    return e.Message
}

// 사전 정의된 에러
var (
    ErrVaultNotFound       = &TeneError{"VAULT_NOT_FOUND", "Not in a Tene project. Run \"tene init\" first.", 1}
    ErrSecretNotFound      = func(key, env string) *TeneError {
        return &TeneError{"SECRET_NOT_FOUND", fmt.Sprintf("Secret %q not found in %q environment.", key, env), 1}
    }
    ErrInvalidPassword     = &TeneError{"INVALID_PASSWORD", "Invalid Master Password. Try again or use Recovery Key.", 2}
    ErrInvalidRecoveryKey  = &TeneError{"INVALID_RECOVERY_KEY", "Invalid Recovery Key.", 2}
    ErrDecryptFailed       = &TeneError{"DECRYPT_FAILED", "Failed to decrypt secret. Master Password may have changed.", 2}
    // ... 기타
)
```

---

## 8. User Stories > Acceptance Criteria

### US-01: tene init (프로젝트 초기화 + CLAUDE.md)

> 바이브코더로서, `tene init`으로 Master Password를 설정하고 로컬 볼트를 생성하며 CLAUDE.md를 자동 생성할 수 있다.

**AC-01-01: 기본 초기화**
```
Given: 프로젝트 디렉토리에 .tene/가 없다
When: tene init 실행 후 Master Password "mypassword123" 입력 + 확인
Then:
  - .tene/ 디렉토리가 생성된다 (퍼미션 0700)
  - .tene/vault.db 파일이 생성된다 (퍼미션 0600)
  - .tene/vault.json 파일이 생성된다
  - .tene/.gitignore 파일이 생성된다 (내용: *)
  - .gitignore에 .tene/ 항목이 추가된다
  - CLAUDE.md가 생성되고 "# Secrets Management" 섹션을 포함한다
  - Recovery Key 12단어가 출력된다
  - OS Keychain에 Master Key가 저장된다
  - "default" 환경이 생성된다
  - 종료 코드 0
```

**AC-01-02: 기존 CLAUDE.md 병합**
```
Given: 프로젝트 디렉토리에 "# My Project" 내용의 CLAUDE.md가 있다
When: tene init 실행
Then:
  - CLAUDE.md의 기존 내용이 보존된다
  - 파일 끝에 "# Secrets Management" 섹션이 추가된다
```

**AC-01-03: 중복 초기화**
```
Given: .tene/vault.db가 이미 존재한다
When: tene init 실행
Then:
  - "Vault already exists." 메시지 출력
  - 기존 볼트 유지 (덮어쓰지 않음)
  - 종료 코드 0
```

**AC-01-04: 패스워드 검증**
```
Given: tene init 실행 중
When: Master Password가 7자 (8자 미만)
Then:
  - "Master Password must be at least 8 characters." 에러
  - 종료 코드 2
```

**AC-01-05: 비대화형 초기화**
```
Given: TENE_MASTER_PASSWORD=mypassword123 환경변수 설정됨
When: tene init --quiet 실행
Then:
  - 프롬프트 없이 볼트 생성
  - Recovery Key가 stdout에 출력
  - 종료 코드 0
```

---

### US-02: tene set (시크릿 저장)

> 바이브코더로서, `tene set KEY VALUE`로 시크릿을 로컬에 암호화 저장할 수 있다.

**AC-02-01: 기본 저장**
```
Given: 볼트가 초기화되어 있다
When: tene set STRIPE_KEY sk_test_xxxxx
Then:
  - SQLite에 encrypted_value로 저장 (평문 아님)
  - "STRIPE_KEY saved (encrypted, default)" 출력
  - 종료 코드 0
```

**AC-02-02: stdin 입력**
```
Given: 볼트가 초기화되어 있다
When: echo "sk_test_xxxxx" | tene set STRIPE_KEY --stdin
Then:
  - STRIPE_KEY가 저장된다
  - shell history에 값이 남지 않는다
```

**AC-02-03: 대화형 입력 (값 생략)**
```
Given: 볼트가 초기화되어 있고 TTY 환경이다
When: tene set STRIPE_KEY (VALUE 생략)
Then:
  - "? Value: " 프롬프트 표시 (마스킹)
  - 입력값이 저장된다
```

**AC-02-04: 중복 키 에러**
```
Given: STRIPE_KEY가 이미 존재한다
When: tene set STRIPE_KEY new_value
Then:
  - "Secret \"STRIPE_KEY\" already exists. Use --overwrite to replace." 에러
  - 종료 코드 1
```

**AC-02-05: --overwrite로 업데이트**
```
Given: STRIPE_KEY가 이미 존재한다 (version 1)
When: tene set STRIPE_KEY new_value --overwrite
Then:
  - STRIPE_KEY가 업데이트된다 (version 2)
  - 종료 코드 0
```

**AC-02-06: 잘못된 키 이름**
```
Given: 볼트가 초기화되어 있다
When: tene set invalid-key value
Then:
  - "Invalid key name" 에러
  - 종료 코드 1
```

---

### US-03: tene get (시크릿 조회)

> 바이브코더로서, `tene get KEY`로 시크릿을 조회할 수 있다.

**AC-03-01: 기본 조회**
```
Given: STRIPE_KEY = "sk_test_xxxxx"가 저장되어 있다
When: tene get STRIPE_KEY
Then:
  - stdout에 "sk_test_xxxxx\n" 출력 (순수 값만, 레이블 없음)
  - 종료 코드 0
```

**AC-03-02: JSON 조회**
```
Given: STRIPE_KEY = "sk_test_xxxxx"가 저장되어 있다
When: tene get STRIPE_KEY --json
Then:
  - stdout에 {"ok":true,"name":"STRIPE_KEY","value":"sk_test_xxxxx","environment":"default"} 출력
  - 종료 코드 0
```

**AC-03-03: 없는 시크릿 조회**
```
Given: NONEXISTENT_KEY가 존재하지 않는다
When: tene get NONEXISTENT_KEY
Then:
  - stderr에 "Secret \"NONEXISTENT_KEY\" not found" 에러
  - 종료 코드 1
```

**AC-03-04: 변수 캡처**
```
Given: STRIPE_KEY가 저장되어 있다
When: STRIPE_KEY=$(tene get STRIPE_KEY)
Then:
  - $STRIPE_KEY 변수에 순수 값이 들어간다 (개행 없음)
```

---

### US-04: tene run (시크릿 주입 실행)

> 바이브코더로서, `tene run -- claude`로 시크릿이 주입된 환경에서 명령을 실행할 수 있다.

**AC-04-01: 기본 주입 실행**
```
Given: 3개 시크릿이 "default" 환경에 있다
When: tene run -- env | grep STRIPE
Then:
  - 자식 프로세스의 환경변수에 STRIPE_KEY가 존재한다
  - 부모 프로세스에는 STRIPE_KEY 환경변수가 없다
```

**AC-04-02: 종료 코드 전달**
```
Given: 시크릿이 존재한다
When: tene run -- bash -c "exit 42"
Then:
  - tene 종료 코드 = 42
```

**AC-04-03: stdin/stdout/stderr 패스스루**
```
Given: 시크릿이 존재한다
When: echo "input" | tene run -- cat
Then:
  - stdout에 "input" 출력 (패스스루)
```

**AC-04-04: 환경 지정**
```
Given: "prod" 환경에 8개 시크릿이 있다
When: tene run --env prod -- node server.js
Then:
  - "prod" 환경의 8개 시크릿이 주입된다
```

---

### US-05: tene list (시크릿 목록)

> 바이브코더로서, `tene list`로 시크릿 목록을 볼 수 있다.

**AC-05-01: 기본 목록**
```
Given: STRIPE_KEY, DATABASE_URL, API_SECRET이 있다
When: tene list
Then:
  - 3개 시크릿 이름과 마스킹된 값이 표시된다
  - 값의 처음 5자만 보이고 나머지는 *****
```

**AC-05-02: 빈 환경**
```
Given: "default" 환경에 시크릿이 없다
When: tene list
Then:
  - "No secrets in \"default\" environment." 메시지
  - 종료 코드 0
```

---

### US-06: tene delete (시크릿 삭제)

> 바이브코더로서, `tene delete KEY`로 시크릿을 삭제할 수 있다.

**AC-06-01: 확인 후 삭제**
```
Given: STRIPE_KEY가 존재하고 TTY 환경이다
When: tene delete STRIPE_KEY -> "y" 입력
Then:
  - "STRIPE_KEY deleted." 출력
  - tene get STRIPE_KEY -> SECRET_NOT_FOUND 에러
```

**AC-06-02: --force 삭제**
```
Given: STRIPE_KEY가 존재한다
When: tene delete STRIPE_KEY --force
Then:
  - 프롬프트 없이 즉시 삭제
  - 종료 코드 0
```

---

### US-07: tene import (시크릿 가져오기)

> 바이브코더로서, `tene import .env`로 기존 .env 파일을 마이그레이션할 수 있다.

**AC-07-01: .env 파일 가져오기**
```
Given: .env 파일에 STRIPE_KEY=sk_test, DB_URL=postgres://... 가 있다
When: tene import .env -> "y" 입력
Then:
  - 2개 시크릿이 암호화 저장된다
  - "2 secrets imported (encrypted)." 출력
```

**AC-07-02: 암호화 백업 복원**
```
Given: my-project.tene.enc 암호화 백업 파일이 있다
When: tene import --encrypted my-project.tene.enc -> Master Password 입력
Then:
  - 백업의 시크릿이 복원된다
```

---

### US-08: tene export (시크릿 내보내기)

> 바이브코더로서, `tene export`로 .env 형식으로 내보낼 수 있다.

**AC-08-01: stdout 내보내기**
```
Given: STRIPE_KEY, DATABASE_URL이 있다
When: tene export
Then:
  - stdout에 .env 형식으로 출력: "STRIPE_KEY=sk_test_xxxxx\nDATABASE_URL=postgres://..."
```

**AC-08-02: 파일 내보내기**
```
Given: 시크릿이 있다
When: tene export --file .env.local
Then:
  - .env.local 파일 생성
  - 경고 메시지 출력: "Warning: This file contains plain-text secrets."
```

---

### US-09: tene export --encrypted (암호화 백업)

> 바이브코더로서, `tene export --encrypted`로 암호화된 백업을 생성할 수 있다.

**AC-09-01: 암호화 백업 생성**
```
Given: "default" 환경에 5개 시크릿이 있다
When: tene export --encrypted
Then:
  - my-project.tene.enc 파일 생성
  - 파일이 TENE 매직 바이트로 시작
  - 파일 내용은 Master Password 없이 복호화 불가
```

**AC-09-02: 라운드트립 검증**
```
Given: tene export --encrypted로 백업을 생성했다
When: 새 프로젝트에서 tene init 후 tene import --encrypted backup.tene.enc
Then:
  - 모든 시크릿이 동일하게 복원된다
```

---

### US-10: tene env (환경 전환)

> 바이브코더로서, `tene env dev/prod`로 환경을 전환할 수 있다.

**AC-10-01: 환경 생성 및 전환**
```
Given: "default" 환경만 존재한다
When: tene env create dev -> tene env dev
Then:
  - "dev" 환경이 생성된다
  - 현재 활성 환경이 "dev"로 전환된다
```

**AC-10-02: 환경별 시크릿 분리**
```
Given: "dev" 환경에 STRIPE_KEY=sk_test, "prod" 환경에 STRIPE_KEY=sk_live
When: tene env dev && tene get STRIPE_KEY
Then:
  - "sk_test" 출력 (dev 환경의 값)
```

---

### US-11: 오프라인 동작

> 바이브코더로서, 오프라인에서도 모든 CLI 명령이 동작한다.

**AC-11-01: 네트워크 차단 동작**
```
Given: 네트워크가 완전히 차단되어 있다
When: tene set/get/run/list/delete/import/export/env 실행
Then:
  - 모든 명령어가 정상 동작한다
  - 네트워크 관련 에러가 발생하지 않는다
```

---

### US-12: brew 설치

> 바이브코더로서, `brew install tomo-kay/tap/tene` 한 줄로 설치할 수 있다.

**AC-12-01: Homebrew 설치**
```
Given: macOS 환경이다
When: brew install tomo-kay/tap/tene
Then:
  - tene 바이너리가 설치된다
  - tene version 출력 성공
  - 10초 이내 완료
```

---

### US-13: tene sync (Fake Door)

> 바이브코더로서, `tene sync`를 실행하면 Cloud waitlist 안내를 볼 수 있다.

**AC-13-01: Fake Door 표시**
```
Given: 볼트가 초기화되어 있다
When: tene sync
Then:
  - Cloud Sync 안내 메시지 출력
  - waitlist URL (https://tene.sh/waitlist) 표시
  - ~/.tene/config.json의 syncAttempts += 1
  - 종료 코드 0
```

---

### US-14: tene passwd (비밀번호 변경)

> 바이브코더로서, `tene passwd`로 Master Password를 변경하고 새 Recovery Key를 발급받을 수 있다.

**AC-14-01: 패스워드 변경 + 재암호화**
```
Given: 5개 시크릿이 있다
When: tene passwd -> 현재 비밀번호 입력 -> 새 비밀번호 입력+확인
Then:
  - 5개 시크릿이 새 키로 재암호화된다
  - 새 Recovery Key 12단어가 출력된다
  - 이전 Recovery Key로는 복구 불가
  - 새 비밀번호로 tene get 동작 확인
```

---

### US-15: tene recover (비밀번호 복구)

> 바이브코더로서, Recovery Key로 Master Password를 재설정할 수 있다.

**AC-15-01: Recovery Key 복구**
```
Given: Recovery Key 12단어를 알고 있다
When: tene recover -> 12단어 입력 -> 새 비밀번호 설정
Then:
  - Master Password가 재설정된다
  - 모든 시크릿이 새 키로 재암호화된다
  - 새 Recovery Key가 발급된다
  - 이전 Recovery Key는 무효화된다
```

---

### US-16: tene whoami (상태 확인)

> 바이브코더로서, `tene whoami`로 현재 볼트 상태를 확인할 수 있다.

**AC-16-01: 상태 표시**
```
Given: 볼트가 초기화되어 있고 5개 시크릿이 있다
When: tene whoami
Then:
  - 프로젝트 이름, 볼트 경로, 활성 환경, 시크릿 수 표시
  - Keychain 상태 표시
  - 종료 코드 0
```

---

*Generated by PM Lead Agent | 2026-04-06*
*Tene CLI Go MVP Phase 1 Requirements Specification*
