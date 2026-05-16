# Tene Open Core 설계서

> 기획서: docs/01-plan/tene-private-repo-migration.plan.md
> 제안서: docs/01-plan/tene-private-repo-migration.proposal.md
> 작성일: 2026-04-09
> 상태: Draft

---

## 1. 설계 개요

### AS-IS: Monorepo (Public)

현재 `agent-kay-it/tene` (public) 단일 레포지토리에 CLI, Cloud API, Dashboard, Landing, Terraform 인프라가 모두 포함되어 있다.

```
agent-kay-it/tene (Public)
├── cmd/tene/           CLI
├── cmd/server/         Cloud API
├── internal/           Go 패키지 15개
├── apps/web/           Landing (tene.sh)
├── apps/dashboard/     Dashboard (app.tene.sh)
├── infra/terraform/    AWS 인프라 13개 모듈
├── migrations/         PostgreSQL 마이그레이션
└── .github/workflows/  CI/CD 파이프라인
```

문제점:
- AWS Account ID, VPC CIDR, IAM Role ARN 등 인프라 정보 공개 노출
- Cloud API 비즈니스 로직 (billing, auth, sync) 전체 공개
- 유료 제품의 전체 소스코드가 무방비 상태

### TO-BE: 2-Repo Open Core

```
agent-kay-it/tene (Public, MIT)
├── cmd/tene/           CLI entrypoint
├── pkg/domain/         데이터 모델 (Shared)
├── pkg/crypto/         암호화 유틸 (Shared)
├── pkg/errors/         에러 타입 (Shared)
├── internal/cli/       Cobra commands
├── internal/vault/     SQLite vault
├── internal/keychain/  OS Keychain
├── internal/sync/      Push/Pull/Merge
├── internal/config/    CLI config
├── internal/recovery/  BIP-39 mnemonic
├── internal/claudemd/  CLAUDE.md generation
├── internal/encfile/   Encrypted file format
├── apps/web/           Landing (tene.sh)
└── .github/workflows/  CLI 릴리스 파이프라인

agent-kay-it/tene-cloud (Private)
├── cmd/server/         Cloud API entrypoint
├── internal/api/       Echo server, handlers, middleware
├── internal/auth/      OAuth + JWT
├── internal/billing/   LemonSqueezy
├── internal/repository/postgres/  PostgreSQL 구현체
├── apps/dashboard/     Dashboard (app.tene.sh)
├── infra/terraform/    AWS 인프라 13개 모듈
├── migrations/         PostgreSQL 마이그레이션
└── .github/workflows/  API 배포 + Dashboard CI
```

### 설계 원칙

1. **Zero Circular Dependency**: Public repo는 Private repo를 절대 import하지 않는다.
2. **Shared Packages in Public**: 양쪽에서 사용하는 패키지(domain, crypto, errors)는 Public repo의 `pkg/`에 배치한다.
3. **internal/ 유지**: CLI-only/Cloud-only 패키지는 각 repo의 `internal/`에 유지하여 외부 import를 차단한다.
4. **독립 빌드**: 각 repo는 독립적으로 `go build`, `go test`가 가능해야 한다.

---

## 2. Repository 구조

### 2-1. agent-kay-it/tene (Public, MIT License)

```
agent-kay-it/tene/
├── cmd/
│   └── tene/
│       └── main.go                 # CLI entrypoint (goreleaser 빌드 대상)
│
├── pkg/                            # Public API — 외부 모듈에서 import 가능
│   ├── domain/                     # 데이터 모델: User, Vault, Team, Device, AuditLog
│   │   ├── models.go
│   │   └── errors.go               # 도메인 sentinel errors
│   ├── crypto/                     # XChaCha20-Poly1305, Argon2id, HKDF, X25519
│   │   ├── argon2id.go
│   │   ├── xchacha20.go
│   │   ├── hkdf.go
│   │   └── x25519.go
│   └── errors/                     # CLI 에러 코드, TeneError 타입, recovery hints
│       └── errors.go
│
├── internal/                       # CLI-only — 외부 import 불가
│   ├── cli/                        # Cobra commands (25+)
│   │   ├── root.go
│   │   ├── init.go
│   │   ├── set.go
│   │   ├── get.go
│   │   ├── run.go
│   │   ├── list.go
│   │   ├── delete.go
│   │   ├── import.go
│   │   ├── export.go
│   │   ├── env.go
│   │   ├── passwd.go
│   │   ├── recover.go
│   │   ├── push.go
│   │   ├── pull.go
│   │   ├── sync.go
│   │   ├── login.go
│   │   ├── logout.go
│   │   ├── team.go
│   │   ├── billing.go
│   │   ├── whoami.go
│   │   ├── version.go
│   │   └── update.go
│   ├── vault/                      # SQLite vault CRUD, schema, migrations
│   ├── keychain/                   # OS Keychain (macOS/Linux) + file fallback
│   ├── sync/                       # Push/Pull engine, 3-way merge, Sync Envelope
│   ├── config/                     # CLI + global config management
│   ├── recovery/                   # BIP-39 mnemonic generation/recovery
│   ├── claudemd/                   # CLAUDE.md auto-generation
│   └── encfile/                    # Encrypted file format (header + ciphertext)
│
├── apps/
│   └── web/                        # Landing page (Next.js 15, tene.sh)
│       ├── src/
│       ├── public/
│       │   └── install.sh          # CLI installer (S3 다운로드)
│       ├── package.json
│       └── next.config.ts
│
├── .github/
│   └── workflows/
│       └── release.yml             # v* 태그 → GoReleaser → S3 + GitHub Releases
│
├── .goreleaser.yml                 # CLI 바이너리 빌드 + S3 업로드
├── go.mod
├── go.sum
├── LICENSE                         # MIT
├── README.md
└── CLAUDE.md
```

각 패키지의 역할:

| 패키지 | 경로 | 역할 |
|--------|------|------|
| **domain** | `pkg/domain/` | User, Vault, Team, Device 등 데이터 모델 + sentinel errors |
| **crypto** | `pkg/crypto/` | Argon2id KDF, XChaCha20-Poly1305 암복호화, HKDF 키 파생, X25519 ECDH |
| **errors** | `pkg/errors/` | TeneError 타입 (Code, Message, Exit), CLI 에러 코드, recovery hints |
| **cli** | `internal/cli/` | 25+ Cobra commands, global flags (--json, --env, --quiet) |
| **vault** | `internal/vault/` | SQLite CRUD, schema (secrets, vault_meta, environments, audit_log) |
| **keychain** | `internal/keychain/` | go-keyring wrapper, macOS Keychain / Linux libsecret / file fallback |
| **sync** | `internal/sync/` | Sync Envelope seal/open, push/pull engine, 3-way merge |
| **config** | `internal/config/` | ~/.tene/config.json, .tene/vault.json 관리 |
| **recovery** | `internal/recovery/` | BIP-39 mnemonic 생성, master key 복구 |
| **claudemd** | `internal/claudemd/` | CLAUDE.md 템플릿 생성, 5개 AI 에디터 지원 |
| **encfile** | `internal/encfile/` | 암호화 파일 포맷 (TENV magic + version + nonce + ciphertext) |

### 2-2. agent-kay-it/tene-cloud (Private)

```
agent-kay-it/tene-cloud/
├── cmd/
│   └── server/
│       └── main.go                 # API server entrypoint (Docker 빌드 대상)
│
├── internal/                       # Cloud-only — 외부 import 불가
│   ├── api/
│   │   ├── server.go               # Echo server 초기화, DI, route 등록
│   │   ├── handler/                # HTTP handlers
│   │   │   ├── auth.go             # OAuth callback, refresh, signout, me
│   │   │   ├── vault.go            # vault CRUD, push/pull
│   │   │   ├── team.go             # team CRUD, invite, remove, RBAC
│   │   │   ├── billing.go          # subscription, checkout, portal, webhook
│   │   │   ├── device.go           # device register, list, delete
│   │   │   ├── audit.go            # audit logs
│   │   │   └── waitlist.go         # waitlist registration
│   │   ├── middleware/             # JWT auth, rate limit, CORS, security headers, RBAC
│   │   ├── response/              # Structured JSON responses
│   │   └── storage/               # S3 client for vault blob storage
│   ├── auth/                       # OAuth (GitHub PKCE) + JWT (HS256, 15min/30day)
│   ├── billing/                    # LemonSqueezy integration + HMAC webhooks
│   └── repository/
│       └── postgres/               # PostgreSQL 구현체 (users, vaults, teams, etc.)
│
├── apps/
│   └── dashboard/                  # Dashboard (Next.js 15, app.tene.sh)
│       ├── src/
│       ├── package.json
│       └── next.config.ts
│
├── infra/
│   └── terraform/
│       ├── modules/                # 13개 모듈
│       │   ├── vpc/
│       │   ├── nat/
│       │   ├── ecs/
│       │   ├── alb/
│       │   ├── rds/
│       │   ├── s3/
│       │   ├── ecr/
│       │   ├── route53/
│       │   ├── acm/
│       │   ├── iam/
│       │   ├── secrets/
│       │   ├── cloudwatch/
│       │   └── waf/
│       └── environments/
│           ├── prod/
│           └── staging/
│
├── migrations/                     # PostgreSQL 마이그레이션 (000001-000008)
│
├── .github/
│   └── workflows/
│       ├── ci.yml                  # test + lint → Docker → ECR → ECS deploy
│       └── dashboard.yml           # TypeScript check + build 검증
│
├── Dockerfile.server               # API server Docker image
├── docker-compose.dev.yml          # Local dev (PostgreSQL + MinIO)
├── scripts/
│   ├── dev.sh
│   └── sync-secrets.sh
├── go.mod
├── go.sum
└── CLAUDE.md
```

각 패키지의 역할:

| 패키지 | 경로 | 역할 |
|--------|------|------|
| **api/server** | `internal/api/server.go` | Echo 서버 초기화, 의존성 주입, 라우트 등록 |
| **api/handler** | `internal/api/handler/` | 7개 핸들러 (auth, vault, team, billing, device, audit, waitlist) |
| **api/middleware** | `internal/api/middleware/` | JWT 인증, rate limit, CORS, security headers, RBAC |
| **api/response** | `internal/api/response/` | 구조화된 JSON 응답 (Success, Error, Paginated) |
| **api/storage** | `internal/api/storage/` | S3 client (presigned URL, vault blob upload/download) |
| **auth** | `internal/auth/` | GitHub OAuth (PKCE), JWT 발급/검증, token family tracking |
| **billing** | `internal/billing/` | LemonSqueezy API, webhook HMAC 검증, subscription 관리 |
| **repository/postgres** | `internal/repository/postgres/` | pgx/v5 기반 PostgreSQL 저장소 구현체 |

---

## 3. Go 모듈 의존성 설계

### 3-1. Public repo go.mod

```go
module github.com/agent-kay-it/tene

go 1.25.0

require (
    // CLI framework
    github.com/spf13/cobra v1.8.1

    // Crypto
    golang.org/x/crypto v0.46.0     // Argon2id, XChaCha20-Poly1305, HKDF, X25519

    // Local DB
    modernc.org/sqlite v1.34.5      // Pure Go SQLite (no CGo)

    // OS Keychain
    github.com/zalando/go-keyring v0.2.6

    // Recovery
    github.com/tyler-smith/go-bip39 v1.1.0

    // Testing
    github.com/stretchr/testify v1.11.1

    // Terminal
    golang.org/x/term v0.38.0
)
```

제거되는 의존성 (Cloud API 전용):
- `github.com/labstack/echo/v4` — Echo web framework
- `github.com/golang-jwt/jwt/v5` — JWT
- `github.com/jackc/pgx/v5` — PostgreSQL driver
- `github.com/aws/aws-sdk-go-v2/*` — AWS SDK (S3, etc.)
- `golang.org/x/oauth2` — OAuth
- `golang.org/x/time` — Rate limiting
- `github.com/google/uuid` — UUID generation

### 3-2. Private repo go.mod

```go
module github.com/agent-kay-it/tene-cloud

go 1.25.0

require (
    // Shared packages from public repo
    github.com/agent-kay-it/tene v0.x.x

    // Web framework
    github.com/labstack/echo/v4 v4.15.1

    // Auth
    github.com/golang-jwt/jwt/v5 v5.3.1
    golang.org/x/oauth2 v0.36.0

    // Database
    github.com/jackc/pgx/v5 v5.9.1

    // AWS
    github.com/aws/aws-sdk-go-v2 v1.41.5
    github.com/aws/aws-sdk-go-v2/config v1.32.14
    github.com/aws/aws-sdk-go-v2/service/s3 v1.98.0

    // Utilities
    github.com/google/uuid v1.6.0
    golang.org/x/time v0.15.0         // Rate limiting

    // Testing
    github.com/stretchr/testify v1.11.1
)
```

### 3-3. Shared 패키지 사용 패턴

#### 문제: Go `internal/` 제약

Go의 `internal/` 디렉토리는 같은 모듈 내부에서만 import 가능하다. 현재 shared 패키지(domain, crypto, errors)가 `internal/`에 있으므로, Private repo에서 직접 import할 수 없다.

```go
// tene-cloud에서 이렇게 import하면 컴파일 에러:
import "github.com/agent-kay-it/tene/internal/domain"  // ERROR: use of internal package
```

#### 해결: `internal/` -> `pkg/` 이동

Shared 패키지 3개를 `internal/`에서 `pkg/`로 이동한다. `pkg/`는 Go에서 특별한 제약이 없으므로 외부 모듈에서 자유롭게 import 가능하다.

```
# 이동 대상 (3개 패키지)
internal/domain/  → pkg/domain/
internal/crypto/  → pkg/crypto/
internal/errors/  → pkg/errors/
```

이동 후 Private repo에서의 import:

```go
// tene-cloud/internal/api/handler/vault.go
import (
    "github.com/agent-kay-it/tene/pkg/domain"
    "github.com/agent-kay-it/tene/pkg/crypto"
)

func (h *VaultHandler) Push(c echo.Context) error {
    var vault domain.Vault
    // ...
}
```

#### 대안 검토

| 방식 | 장점 | 단점 | 채택 |
|------|------|------|:----:|
| `pkg/`로 이동 | 간단, Go 관례, 외부 import 가능 | public API 노출 (의도적) | **O** |
| git submodule | 독립 버전 관리 | 운영 복잡도 높음, submodule 동기화 부담 | X |
| shared 패키지 복제 | 독립 유지, 서로 영향 없음 | DRY 위반, 버그 수정 시 양쪽 모두 수정 | X |
| 별도 모듈 (tene-core) | 완전한 독립성 | 3번째 repo 관리 부담, 1인 운영에 과도 | X |

#### GOPRIVATE 설정 불필요

Public repo(`agent-kay-it/tene`)는 Go module proxy(`proxy.golang.org`)에서 자동으로 접근 가능하다. Private repo에서 `go get github.com/agent-kay-it/tene`을 실행할 때 별도의 GOPRIVATE 설정이 필요하지 않다.

단, CI/CD 환경(GitHub Actions)에서 Private repo 자체의 `go mod download`는 정상 동작한다 (`actions/checkout`으로 소스를 이미 받은 상태에서 빌드하므로).

---

## 4. CI/CD 파이프라인 설계

### 4-1. Public repo workflows

#### release.yml: CLI 바이너리 릴리스

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write
  id-token: write  # OIDC for AWS S3 upload

env:
  AWS_REGION: ap-northeast-2
  RELEASE_BUCKET: tene-releases

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Configure AWS credentials (OIDC)
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::507221376909:role/tene-prod-github-actions
          aws-region: ${{ env.AWS_REGION }}

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v7
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          AWS_REGION: ${{ env.AWS_REGION }}

      - name: Update LATEST_VERSION
        run: |
          VERSION="${GITHUB_REF_NAME#v}"
          echo "$VERSION" > /tmp/LATEST_VERSION
          aws s3 cp /tmp/LATEST_VERSION s3://${{ env.RELEASE_BUCKET }}/LATEST_VERSION \
            --content-type "text/plain" \
            --cache-control "max-age=60"
```

#### ci.yml (선택): CLI 코드 품질 검증

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.25"
      - run: go mod download
      - run: go vet ./...
      - run: go test -race -coverprofile=coverage.out ./...

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.25"
      - name: Run golangci-lint
        run: |
          go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
          golangci-lint run
```

### 4-2. Private repo workflows

#### ci.yml: API 서버 빌드 + 배포

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, staging]
  pull_request:
    branches: [main, staging]

permissions:
  contents: read
  id-token: write  # OIDC for AWS

env:
  AWS_REGION: ap-northeast-2
  ECR_REPO: 507221376909.dkr.ecr.ap-northeast-2.amazonaws.com/tene-api

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.25"
      - run: go mod download
      - run: go vet ./...
      - run: go test -race -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out | tail -1

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.25"
      - name: Run golangci-lint
        run: |
          go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6
          golangci-lint run

  deploy-prod:
    runs-on: ubuntu-latest
    needs: [test, lint]
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    environment: production
    env:
      ECS_CLUSTER: tene-prod-cluster
      ECS_SERVICE: tene-prod-api
      TASK_FAMILY: tene-prod-api
    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS credentials (OIDC)
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::507221376909:role/tene-prod-github-actions
          aws-region: ${{ env.AWS_REGION }}

      - name: Login to ECR
        uses: aws-actions/amazon-ecr-login@v2

      - name: Build, tag, push Docker image
        run: |
          IMAGE_TAG="${{ github.sha }}"
          docker build --platform linux/amd64 -t $ECR_REPO:$IMAGE_TAG -f Dockerfile.server .
          docker push $ECR_REPO:$IMAGE_TAG

      - name: Deploy to ECS
        run: |
          IMAGE_TAG="${{ github.sha }}"
          TASK_DEF=$(aws ecs describe-task-definition --task-definition $TASK_FAMILY --query taskDefinition)
          NEW_TASK_DEF=$(echo $TASK_DEF | jq --arg IMG "$ECR_REPO:$IMAGE_TAG" \
            '.containerDefinitions[0].image = $IMG | del(.taskDefinitionArn, .revision, .status, .requiresAttributes, .compatibilities, .registeredAt, .registeredBy)')
          NEW_ARN=$(aws ecs register-task-definition --cli-input-json "$NEW_TASK_DEF" --query 'taskDefinition.taskDefinitionArn' --output text)
          aws ecs update-service --cluster $ECS_CLUSTER --service $ECS_SERVICE --task-definition $NEW_ARN --force-new-deployment
          echo "Deployed $IMAGE_TAG to ECS (prod)"

  deploy-staging:
    runs-on: ubuntu-latest
    needs: [test, lint]
    if: github.ref == 'refs/heads/staging' && github.event_name == 'push'
    environment: staging
    env:
      ECS_CLUSTER: tene-staging-cluster
      ECS_SERVICE: tene-staging-api
      TASK_FAMILY: tene-staging-api
    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS credentials (OIDC)
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::507221376909:role/tene-staging-github-actions
          aws-region: ${{ env.AWS_REGION }}

      - name: Login to ECR
        uses: aws-actions/amazon-ecr-login@v2

      - name: Build, tag, push Docker image
        run: |
          IMAGE_TAG="${{ github.sha }}"
          docker build --platform linux/amd64 -t $ECR_REPO:$IMAGE_TAG -f Dockerfile.server .
          docker push $ECR_REPO:$IMAGE_TAG

      - name: Deploy to ECS
        run: |
          IMAGE_TAG="${{ github.sha }}"
          TASK_DEF=$(aws ecs describe-task-definition --task-definition $TASK_FAMILY --query taskDefinition)
          NEW_TASK_DEF=$(echo $TASK_DEF | jq --arg IMG "$ECR_REPO:$IMAGE_TAG" \
            '.containerDefinitions[0].image = $IMG | del(.taskDefinitionArn, .revision, .status, .requiresAttributes, .compatibilities, .registeredAt, .registeredBy)')
          NEW_ARN=$(aws ecs register-task-definition --cli-input-json "$NEW_TASK_DEF" --query 'taskDefinition.taskDefinitionArn' --output text)
          aws ecs update-service --cluster $ECS_CLUSTER --service $ECS_SERVICE --task-definition $NEW_ARN --force-new-deployment
          echo "Deployed $IMAGE_TAG to ECS (staging)"
```

#### dashboard.yml: Dashboard 빌드 검증

```yaml
# .github/workflows/dashboard.yml
name: Dashboard CI

on:
  push:
    branches: [main]
    paths:
      - 'apps/dashboard/**'
  pull_request:
    branches: [main]
    paths:
      - 'apps/dashboard/**'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with:
          version: 9
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: 'pnpm'
          cache-dependency-path: apps/dashboard/pnpm-lock.yaml
      - run: cd apps/dashboard && pnpm install --frozen-lockfile
      - run: cd apps/dashboard && pnpm run build
```

---

## 5. 배포 파이프라인 (S3 릴리스)

### 5-1. S3 버킷 설계

| 항목 | 값 |
|------|-----|
| 버킷명 | `tene-releases` |
| 리전 | `ap-northeast-2` |
| 접근 정책 | Public Read (s3:GetObject) |
| 버전 관리 | Enabled |
| CORS | GET/HEAD from * |
| 암호화 | SSE-S3 (AES-256) |

버킷 구조:

```
s3://tene-releases/
├── LATEST_VERSION                              # "0.3.1" (텍스트 파일, max-age=60)
├── v0.3.1/
│   ├── tene_0.3.1_darwin_amd64.tar.gz
│   ├── tene_0.3.1_darwin_arm64.tar.gz
│   ├── tene_0.3.1_linux_amd64.tar.gz
│   ├── tene_0.3.1_linux_arm64.tar.gz
│   ├── tene_0.3.1_windows_amd64.zip
│   └── checksums.txt                           # SHA-256 체크섬
├── v0.3.0/
│   └── ...
└── ...
```

다운로드 URL 패턴:
- 버전 조회: `https://tene-releases.s3.ap-northeast-2.amazonaws.com/LATEST_VERSION`
- 바이너리: `https://tene-releases.s3.ap-northeast-2.amazonaws.com/v{version}/tene_{version}_{os}_{arch}.tar.gz`
- 체크섬: `https://tene-releases.s3.ap-northeast-2.amazonaws.com/v{version}/checksums.txt`

향후 `releases.tene.sh` 서브도메인 CNAME 설정 후 CloudFront 전환 시 URL 변경 없이 가능.

### 5-2. GoReleaser S3 설정

`.goreleaser.yml` 변경:

```yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - id: tene
    main: ./cmd/tene
    binary: tene
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.ShortCommit}}
      - -X main.date={{.Date}}

archives:
  - id: tene-archive
    builds:
      - tene
    format: tar.gz
    name_template: >-
      {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'
  algorithm: sha256

snapshot:
  version_template: "{{ .Tag }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^ci:'
      - '^chore:'
      - Merge pull request
      - Merge branch

release:
  github:
    owner: tomo-kay
    name: tene
  draft: false
  prerelease: auto
  name_template: "v{{.Version}}"

# S3 업로드 (추가)
blobs:
  - provider: s3
    bucket: tene-releases
    region: ap-northeast-2
    directory: "v{{.Version}}"
    extra_files:
      - glob: dist/checksums.txt
```

변경점:
- `blobs:` 섹션 추가: S3에 바이너리 + checksums.txt 업로드
- `release.github` 섹션 유지: public repo이므로 GitHub Releases도 병행 (dual publish)
- GoReleaser가 AWS 자격증명을 환경변수(`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` 또는 OIDC)에서 자동 감지

### 5-3. install.sh 변경

변경 전 (`apps/web/public/install.sh`):

```sh
REPO="agent-kay-it/tene"

get_latest_version() {
  if command -v curl > /dev/null 2>&1; then
    curl -sSfL "https://api.github.com/repos/${REPO}/releases/latest" |
      grep '"tag_name"' | sed 's/.*"tag_name": *"v\([^"]*\)".*/\1/'
  elif command -v wget > /dev/null 2>&1; then
    wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" |
      grep '"tag_name"' | sed 's/.*"tag_name": *"v\([^"]*\)".*/\1/'
  else
    error "curl or wget is required"
  fi
}

# ...
url="https://github.com/${REPO}/releases/download/v${version}/${filename}"
download "$url" "${tmpdir}/${filename}"
tar xzf "${tmpdir}/${filename}" -C "$tmpdir"
```

변경 후:

```sh
RELEASE_BASE="https://tene-releases.s3.ap-northeast-2.amazonaws.com"

get_latest_version() {
  if command -v curl > /dev/null 2>&1; then
    curl -sSfL "${RELEASE_BASE}/LATEST_VERSION"
  elif command -v wget > /dev/null 2>&1; then
    wget -qO- "${RELEASE_BASE}/LATEST_VERSION"
  else
    error "curl or wget is required"
  fi
}

# ...
url="${RELEASE_BASE}/v${version}/${filename}"
checksum_url="${RELEASE_BASE}/v${version}/checksums.txt"

download "$url" "${tmpdir}/${filename}"
download "$checksum_url" "${tmpdir}/checksums.txt"

# SHA-256 체크섬 검증
if command -v sha256sum > /dev/null 2>&1; then
  expected=$(grep "$filename" "${tmpdir}/checksums.txt" | awk '{print $1}')
  actual=$(sha256sum "${tmpdir}/${filename}" | awk '{print $1}')
elif command -v shasum > /dev/null 2>&1; then
  expected=$(grep "$filename" "${tmpdir}/checksums.txt" | awk '{print $1}')
  actual=$(shasum -a 256 "${tmpdir}/${filename}" | awk '{print $1}')
fi

if [ -n "$expected" ] && [ "$expected" != "$actual" ]; then
  error "Checksum verification failed (expected: ${expected}, got: ${actual})"
fi

tar xzf "${tmpdir}/${filename}" -C "$tmpdir"
```

변경 포인트 3곳:
1. **상수**: `REPO="agent-kay-it/tene"` -> `RELEASE_BASE="https://tene-releases.s3..."`
2. **버전 조회**: GitHub API -> S3 `LATEST_VERSION` 파일
3. **다운로드 + 체크섬**: GitHub Releases URL -> S3 URL + SHA-256 검증 추가

### 5-4. tene update 변경

`internal/cli/update.go` 변경:

변경 전:

```go
// 상수 (없음, 함수 내에 하드코딩)

func fetchLatestRelease() (*githubRelease, error) {
    url := "https://api.github.com/repos/agent-kay-it/tene/releases/latest"
    // ...GitHub API JSON 파싱...
}

// 다운로드 URL
downloadURL := fmt.Sprintf(
    "https://github.com/agent-kay-it/tene/releases/download/%s/%s",
    targetVersion, assetName,
)
```

변경 후:

```go
const releaseBaseURL = "https://tene-releases.s3.ap-northeast-2.amazonaws.com"

// fetchLatestRelease: S3에서 최신 버전 조회, 실패 시 GitHub API fallback
func fetchLatestRelease() (*releaseInfo, error) {
    // 1차: S3 시도
    info, err := fetchFromS3()
    if err == nil {
        return info, nil
    }

    // 2차: GitHub API fallback (구버전 호환)
    info, err = fetchFromGitHub()
    if err != nil {
        return nil, fmt.Errorf("cannot check for updates (try: curl -sSfL https://tene.sh/install.sh | sh): %w", err)
    }
    return info, nil
}

func fetchFromS3() (*releaseInfo, error) {
    url := releaseBaseURL + "/LATEST_VERSION"

    resp, err := cliHTTPClient.Get(url)
    if err != nil {
        return nil, fmt.Errorf("s3 version check failed: %w", err)
    }
    defer func() { _ = resp.Body.Close() }()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("s3 returned HTTP %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read version: %w", err)
    }

    ver := strings.TrimSpace(string(body))
    return &releaseInfo{
        Version:     ver,
        DownloadURL: fmt.Sprintf("%s/v%s/", releaseBaseURL, ver),
    }, nil
}

// 다운로드 URL
downloadURL := fmt.Sprintf("%s/v%s/%s", releaseBaseURL, version, assetName)

// 체크섬 검증 (필수)
checksumsURL := fmt.Sprintf("%s/v%s/checksums.txt", releaseBaseURL, version)
if err := verifyChecksum(downloadedPath, checksumsURL, assetName); err != nil {
    return fmt.Errorf("checksum verification failed: %w", err)
}
```

---

## 6. Terraform 인프라 변경

### 6-1. S3 릴리스 버킷 모듈

**파일: `infra/terraform/modules/s3/variables.tf`** (추가)

```hcl
variable "create_release_bucket" {
  description = "Whether to create the CLI release artifacts bucket"
  type        = bool
  default     = false
}

variable "release_bucket_name" {
  description = "Name for the CLI release artifacts bucket"
  type        = string
  default     = ""
}

variable "release_domain" {
  description = "Custom domain for releases (e.g., releases.tene.sh)"
  type        = string
  default     = ""
}
```

**파일: `infra/terraform/modules/s3/main.tf`** (추가)

```hcl
# ──────────────────────────────────────────────────
# CLI release artifacts (public read)
# ──────────────────────────────────────────────────

resource "aws_s3_bucket" "releases" {
  count  = var.create_release_bucket ? 1 : 0
  bucket = var.release_bucket_name

  tags = { Name = var.release_bucket_name }
}

resource "aws_s3_bucket_versioning" "releases" {
  count  = var.create_release_bucket ? 1 : 0
  bucket = aws_s3_bucket.releases[0].id
  versioning_configuration { status = "Enabled" }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "releases" {
  count  = var.create_release_bucket ? 1 : 0
  bucket = aws_s3_bucket.releases[0].id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "releases" {
  count                   = var.create_release_bucket ? 1 : 0
  bucket                  = aws_s3_bucket.releases[0].id
  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "releases_public_read" {
  count  = var.create_release_bucket ? 1 : 0
  bucket = aws_s3_bucket.releases[0].id

  depends_on = [aws_s3_bucket_public_access_block.releases]

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "PublicRead"
      Effect    = "Allow"
      Principal = "*"
      Action    = "s3:GetObject"
      Resource  = "${aws_s3_bucket.releases[0].arn}/*"
    }]
  })
}

resource "aws_s3_bucket_cors_configuration" "releases" {
  count  = var.create_release_bucket ? 1 : 0
  bucket = aws_s3_bucket.releases[0].id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD"]
    allowed_origins = ["*"]
    max_age_seconds = 3600
  }
}
```

**파일: `infra/terraform/modules/s3/outputs.tf`** (추가)

```hcl
output "release_bucket_name" {
  description = "CLI release artifacts bucket name"
  value       = var.create_release_bucket ? aws_s3_bucket.releases[0].id : ""
}

output "release_bucket_arn" {
  description = "CLI release artifacts bucket ARN"
  value       = var.create_release_bucket ? aws_s3_bucket.releases[0].arn : ""
}

output "release_bucket_domain" {
  description = "CLI release artifacts bucket regional domain"
  value       = var.create_release_bucket ? aws_s3_bucket.releases[0].bucket_regional_domain_name : ""
}
```

### 6-2. IAM 정책 확장

**파일: `infra/terraform/modules/iam/variables.tf`** (추가)

```hcl
variable "release_bucket_arn" {
  description = "ARN of the CLI release artifacts S3 bucket"
  type        = string
  default     = ""
}
```

**파일: `infra/terraform/modules/iam/main.tf`** (추가)

`github_actions_deploy` IAM 정책에 S3 릴리스 버킷 업로드 권한 추가:

```hcl
# 기존 policy의 Statement 배열에 추가
{
  Sid      = "S3ReleaseUpload"
  Effect   = "Allow"
  Action   = [
    "s3:PutObject",
    "s3:PutObjectAcl",
    "s3:ListBucket",
    "s3:GetBucketLocation"
  ]
  Resource = [
    var.release_bucket_arn,
    "${var.release_bucket_arn}/*"
  ]
}
```

### 6-3. 환경 설정 연결

**파일: `infra/terraform/environments/prod/main.tf`** (변경)

```hcl
module "s3" {
  source = "../../modules/s3"

  project                = local.project
  environment            = local.environment
  aws_region             = local.region

  # CLI release artifacts (추가)
  create_release_bucket  = true
  release_bucket_name    = "tene-releases"
}

module "iam" {
  source = "../../modules/iam"

  project          = local.project
  environment      = local.environment
  vault_bucket_arn = module.s3.vault_bucket_arn
  secrets_arn      = module.secrets.secrets_arn
  github_org       = "tomo-kay"
  github_repo      = "tene"

  # CLI release upload 권한 (추가)
  release_bucket_arn = module.s3.release_bucket_arn
}
```

**IAM OIDC 주의사항**: `github_repo` 값에 주의해야 한다.
- Public repo (`tene`): release.yml에서 S3 업로드
- Private repo (`tene-cloud`): ci.yml에서 ECR push + ECS deploy

두 repo 모두 동일 OIDC provider를 사용하지만, trust policy의 `sub` 조건이 다르다:
- `repo:agent-kay-it/tene:*` — public repo (CLI 릴리스)
- `repo:agent-kay-it/tene-cloud:*` — private repo (API 배포)

IAM 모듈에서 두 repo 모두에 대한 trust policy를 설정해야 한다.

```hcl
# infra/terraform/modules/iam/main.tf (trust policy 확장)
variable "github_repos" {
  description = "List of GitHub repos allowed to assume the role"
  type        = list(string)
  default     = ["tene", "tene-cloud"]
}

# trust policy condition
condition {
  test     = "StringLike"
  variable = "token.actions.githubusercontent.com:sub"
  values   = [for repo in var.github_repos : "repo:${var.github_org}/${repo}:*"]
}
```

---

## 7. Frontend 분리 설계

### 7-1. Landing (apps/web/) -> Public repo

Landing 페이지는 public repo에 유지한다. Open Core 모델에서 "View on GitHub" 링크가 실제 public repo를 가리키므로 신뢰 구축에 유리하다.

**pnpm workspace 설정 (Public repo)**:

```yaml
# pnpm-workspace.yaml
packages:
  - "apps/*"
```

**Vercel 배포 설정**:

| 항목 | 값 |
|------|-----|
| Framework | Next.js |
| Root Directory | `apps/web` |
| Build Command | `pnpm run build` |
| Output Directory | `.next` |
| Node.js Version | 22.x |
| Domain | `tene.sh` |
| Git Integration | `agent-kay-it/tene` (public) |

환경 변수:

```
NEXT_PUBLIC_API_URL=https://api.tene.sh
```

### 7-2. Dashboard (apps/dashboard/) -> Private repo

Dashboard는 private repo로 이동한다. 유료 기능(billing, team 관리)의 UI이므로 비공개가 적절하다.

**독립 pnpm 설정 (Private repo)**:

Dashboard는 monorepo workspace가 아닌 독립 프로젝트로 운영한다.

```json
// apps/dashboard/package.json
{
  "name": "@tene/dashboard",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "next dev --port 3001",
    "build": "next build",
    "start": "next start",
    "lint": "next lint"
  },
  "dependencies": {
    "@tanstack/react-query": "^5.60.0",
    "clsx": "^2.1.0",
    "cmdk": "^1.1.1",
    "lucide-react": "^0.460.0",
    "next": "^15.2.0",
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "zustand": "^5.0.0"
  }
}
```

**Vercel 배포 설정**:

| 항목 | 값 |
|------|-----|
| Framework | Next.js |
| Root Directory | `apps/dashboard` |
| Build Command | `pnpm run build` |
| Output Directory | `.next` |
| Node.js Version | 22.x |
| Domain | `app.tene.sh` |
| Git Integration | `agent-kay-it/tene-cloud` (private) |

Vercel Hobby 플랜에서 private repo 배포가 지원된다 (추가 비용 없음).

환경 변수:

```
NEXT_PUBLIC_API_URL=https://api.tene.sh
```

### 7-3. 디자인 시스템 일관성

두 프론트엔드 앱이 다른 repo에 위치하므로, 디자인 시스템 동기화 전략이 필요하다.

**현재 디자인 시스템**:

```css
/* 공통 변수 */
--background: #0a0a0a;
--foreground: #ededed;
--accent: #00ff88;
--accent-dim: #00cc6a;
--surface: #141414;
--surface-2: #1e1e1e;
--border: #2a2a2a;
--muted: #888888;
--danger: #ff4444;        /* dashboard only */
--warning: #ffaa00;       /* dashboard only */
```

**동기화 전략: 문서 기반 동기화**

현재 1인 개발 단계에서 디자인 시스템 패키지를 별도로 만드는 것은 과도하다. 대신:

1. **CLAUDE.md에 디자인 시스템 명세 유지** (양쪽 repo 모두)
2. **Tailwind CSS 색상/폰트 설정을 동일하게 유지** (수동, 변경 빈도 낮음)
3. **향후 확장**: 사용자/팀 규모 증가 시 `@tene/design-tokens` npm 패키지 추출 검토

---

## 8. Landing 페이지 변경 상세

### 8-1. GitHub 링크 대체 전략

Public repo이므로 GitHub 링크를 대부분 유지할 수 있다. 기존 기획서(전체 Private)와 달리 대폭 간소화된다.

| 현재 GitHub 링크 용도 | 변경 | 이유 |
|----------------------|------|------|
| Nav: GitHub 아이콘 링크 | **유지** | public repo를 가리키므로 정상 |
| Footer: GitHub 링크 | **유지** | 동일 |
| Footer: Issues 링크 | **유지** | public repo에서 issue 관리 |
| Security: 소스코드 공개 링크 | **유지** (pkg/crypto/ 경로로 변경) | crypto 코드가 public |
| CTA: GitHub 링크 | **유지** | 동일 |
| Hero: GitHub 링크 | **유지** | 동일 |
| Layout: JSON-LD installUrl | **S3 URL로 변경** | install.sh가 S3에서 다운로드 |
| llms.txt: GitHub URL | **유지** | public repo |

Open Core 모델의 핵심 이점: **Landing 페이지 변경이 거의 없다.**

변경이 필요한 곳 (1곳):

1. `apps/web/src/app/layout.tsx` — JSON-LD `installUrl`에서 GitHub Releases 언급 제거, `https://tene.sh/install.sh`로 통일

### 8-2. install.sh URL 변경

install.sh 자체는 여전히 `apps/web/public/install.sh`에 위치하며 `https://tene.sh/install.sh`로 접근 가능하다.

내부 로직 변경:
- 버전 조회: GitHub API -> S3 `LATEST_VERSION`
- 다운로드: GitHub Releases -> S3 버킷
- 체크섬 검증: 추가 (SHA-256)

상세 변경 내용은 섹션 5-3 참조.

---

## 9. 마이그레이션 절차 (Step-by-Step)

### Step 1: S3 릴리스 인프라 구축

**소요: 0.5일**

1. Terraform S3 모듈에 release 버킷 리소스 추가 (섹션 6-1)
2. IAM 정책에 S3 업로드 권한 추가 (섹션 6-2)
3. `terraform plan` -> `terraform apply`
4. 검증: S3 버킷 접근 확인

```bash
cd infra/terraform/environments/prod
TF_VAR_db_password="$(tene get DB_PASSWORD --env prod)" \
TF_VAR_jwt_secret="$(tene get JWT_SECRET --env prod)" \
terraform plan   # S3 버킷 + IAM 변경 확인
terraform apply

# 검증
aws s3 ls s3://tene-releases/ --profile monsa-sandbox
```

### Step 2: GoReleaser + install.sh + update.go 수정

**소요: 1일**

1. `.goreleaser.yml`에 `blobs:` S3 섹션 추가 (섹션 5-2)
2. `release.yml`에 OIDC 인증 + LATEST_VERSION 갱신 스텝 추가 (섹션 4-1)
3. `apps/web/public/install.sh` S3 전환 + 체크섬 검증 (섹션 5-3)
4. `internal/cli/update.go` S3 전환 + fallback + 체크섬 검증 (섹션 5-4)
5. 검증: GoReleaser snapshot 빌드 + install.sh 로컬 테스트

```bash
# GoReleaser dry-run
goreleaser release --snapshot --clean

# install.sh 로컬 테스트
sh apps/web/public/install.sh
```

### Step 3: Shared 패키지 `internal/` -> `pkg/` 이동

**소요: 0.5일**

1. 디렉토리 이동:

```bash
mkdir -p pkg
git mv internal/domain pkg/domain
git mv internal/crypto pkg/crypto
git mv internal/errors pkg/errors
```

2. import 경로 일괄 변경:

```bash
# 모든 Go 파일에서 import path 변경
# "github.com/agent-kay-it/tene/internal/domain"
# -> "github.com/agent-kay-it/tene/pkg/domain"

# domain (가장 많이 import됨)
find . -name "*.go" -exec sed -i '' \
  's|"github.com/agent-kay-it/tene/internal/domain"|"github.com/agent-kay-it/tene/pkg/domain"|g' {} +

# crypto
find . -name "*.go" -exec sed -i '' \
  's|"github.com/agent-kay-it/tene/internal/crypto"|"github.com/agent-kay-it/tene/pkg/crypto"|g' {} +

# errors
find . -name "*.go" -exec sed -i '' \
  's|"github.com/agent-kay-it/tene/internal/errors"|"github.com/agent-kay-it/tene/pkg/errors"|g' {} +
```

3. 빌드 검증:

```bash
go build ./...
go test ./...
```

4. 커밋: `refactor: move shared packages from internal/ to pkg/ for Open Core`

### Step 4: tene-cloud private repo 생성 + 코드 이동

**소요: 1일**

1. GitHub에서 `agent-kay-it/tene-cloud` private repo 생성

2. 이동할 코드 목록:

```
cmd/server/              -> tene-cloud/cmd/server/
internal/api/            -> tene-cloud/internal/api/
internal/auth/           -> tene-cloud/internal/auth/
internal/billing/        -> tene-cloud/internal/billing/
internal/repository/     -> tene-cloud/internal/repository/
apps/dashboard/          -> tene-cloud/apps/dashboard/
infra/terraform/         -> tene-cloud/infra/terraform/
migrations/              -> tene-cloud/migrations/
Dockerfile.server        -> tene-cloud/Dockerfile.server
docker-compose.dev.yml   -> tene-cloud/docker-compose.dev.yml
scripts/dev.sh           -> tene-cloud/scripts/dev.sh
scripts/sync-secrets.sh  -> tene-cloud/scripts/sync-secrets.sh
```

3. tene-cloud의 `go.mod` 초기화:

```bash
cd tene-cloud
go mod init github.com/agent-kay-it/tene-cloud
```

4. public repo 의존성 추가:

```bash
go get github.com/agent-kay-it/tene@latest
```

5. import 경로 변경:

```bash
# tene-cloud 내부의 모든 Go 파일에서
# "github.com/agent-kay-it/tene/internal/api"
# -> "github.com/agent-kay-it/tene-cloud/internal/api"

find . -name "*.go" -exec sed -i '' \
  's|"github.com/agent-kay-it/tene/internal/api|"github.com/agent-kay-it/tene-cloud/internal/api|g' {} +

find . -name "*.go" -exec sed -i '' \
  's|"github.com/agent-kay-it/tene/internal/auth"|"github.com/agent-kay-it/tene-cloud/internal/auth"|g' {} +

find . -name "*.go" -exec sed -i '' \
  's|"github.com/agent-kay-it/tene/internal/billing"|"github.com/agent-kay-it/tene-cloud/internal/billing"|g' {} +

find . -name "*.go" -exec sed -i '' \
  's|"github.com/agent-kay-it/tene/internal/repository"|"github.com/agent-kay-it/tene-cloud/internal/repository"|g' {} +
```

6. Shared 패키지 import는 그대로 유지:

```go
// 이것들은 public repo의 pkg/에서 import (변경 불필요)
import "github.com/agent-kay-it/tene/pkg/domain"
import "github.com/agent-kay-it/tene/pkg/crypto"
import "github.com/agent-kay-it/tene/pkg/errors"
```

7. 빌드 검증:

```bash
go build ./cmd/server
go test ./...
```

### Step 5: CI/CD 파이프라인 분리

**소요: 0.5일**

1. Public repo (`tene`):
   - `.github/workflows/ci.yml` 유지 (test + lint만, deploy 제거)
   - `.github/workflows/release.yml` S3 업로드 추가 (섹션 4-1)
   - `.github/workflows/dashboard.yml` 제거 (dashboard는 private repo로 이동)

2. Private repo (`tene-cloud`):
   - `.github/workflows/ci.yml` 생성 (test + lint + deploy-prod + deploy-staging) (섹션 4-2)
   - `.github/workflows/dashboard.yml` 생성 (섹션 4-2)

3. IAM OIDC trust policy 업데이트:
   - `agent-kay-it/tene` + `agent-kay-it/tene-cloud` 모두 허용 (섹션 6-2)

### Step 6: Vercel 재설정

**소요: 0.5일**

1. **Landing (tene.sh)**:
   - Vercel 프로젝트 설정 확인: Git Integration이 `agent-kay-it/tene` (public)을 가리키는지 확인
   - Root Directory: `apps/web`
   - 변경 필요 없음 (public repo 유지)

2. **Dashboard (app.tene.sh)**:
   - Vercel 프로젝트의 Git Integration을 `agent-kay-it/tene-cloud` (private)로 변경
   - Root Directory: `apps/dashboard`
   - GitHub App 권한에서 `tene-cloud` repo 접근 허용 확인
   - 검증: 빈 커밋 push -> Vercel 빌드 트리거 확인

```bash
# Dashboard Vercel 재연결 후 검증
cd tene-cloud
git commit --allow-empty -m "test: verify Vercel deployment"
git push origin main
# Vercel Dashboard에서 빌드 성공 확인
```

### Step 7: 현재 tene repo에서 Cloud 코드 제거

**소요: 0.5일**

public repo (`tene`)에서 private repo로 이동 완료된 코드를 제거한다.

```bash
# Cloud-only 코드 제거
rm -rf cmd/server/
rm -rf internal/api/
rm -rf internal/auth/
rm -rf internal/billing/
rm -rf internal/repository/
rm -rf apps/dashboard/
rm -rf infra/
rm -rf migrations/
rm -f Dockerfile.server
rm -f docker-compose.dev.yml
rm -rf scripts/

# Cloud-only CI/CD 제거
rm -f .github/workflows/dashboard.yml
# ci.yml에서 deploy-prod, deploy-staging job 제거

# go.mod 정리 (불필요한 Cloud 의존성 제거)
go mod tidy
```

빌드 검증:

```bash
go build ./cmd/tene
go test ./...
```

커밋: `refactor: remove cloud code (moved to tene-cloud private repo)`

### Step 8: 검증 + 릴리스

**소요: 0.5일**

전체 시스템 E2E 검증 (섹션 11 체크리스트 실행).

S3 릴리스 테스트:

```bash
# 테스트 태그 push (public repo)
git tag v0.x.x-rc1
git push origin v0.x.x-rc1

# release.yml 실행 확인
# S3에 바이너리 + LATEST_VERSION 업로드 확인
curl -I https://tene-releases.s3.ap-northeast-2.amazonaws.com/LATEST_VERSION

# install.sh 테스트
curl -sSfL https://tene.sh/install.sh | sh
```

### Step 9: README, CLAUDE.md 업데이트

**소요: 0.5일**

1. Public repo README.md:
   - CLI 기능 중심으로 재작성
   - "Open Core" 모델 설명 추가
   - Cloud/Pro 기능은 `app.tene.sh` 링크로 안내
   - MIT 라이선스 명시

2. Public repo CLAUDE.md:
   - Cloud API 관련 내용 제거
   - CLI + pkg/ 패키지 구조로 업데이트
   - Terraform/infra 관련 내용 제거

3. Private repo CLAUDE.md:
   - Cloud API + Dashboard + Infra 중심으로 작성
   - Public repo의 shared 패키지 의존 관계 명시

---

## 10. 변경 파일 전체 목록

### Public repo (agent-kay-it/tene) 변경

| # | 파일 | 변경 내용 | 구분 |
|:-:|------|----------|------|
| 1 | `pkg/domain/*.go` | `internal/domain/` -> `pkg/domain/` 이동 | 패키지 이동 |
| 2 | `pkg/crypto/*.go` | `internal/crypto/` -> `pkg/crypto/` 이동 | 패키지 이동 |
| 3 | `pkg/errors/*.go` | `internal/errors/` -> `pkg/errors/` 이동 | 패키지 이동 |
| 4 | `internal/**/*.go` | import path `internal/{domain,crypto,errors}` -> `pkg/` 변경 | import 수정 |
| 5 | `go.mod` | Cloud-only 의존성 제거 (echo, pgx, aws-sdk, oauth2 등) | 의존성 정리 |
| 6 | `.goreleaser.yml` | `blobs:` S3 섹션 추가 | S3 릴리스 |
| 7 | `.github/workflows/release.yml` | OIDC 인증 + LATEST_VERSION 갱신 스텝 추가 | CI/CD |
| 8 | `.github/workflows/ci.yml` | deploy-prod, deploy-staging job 제거 | CI/CD |
| 9 | `.github/workflows/dashboard.yml` | 파일 삭제 | CI/CD |
| 10 | `apps/web/public/install.sh` | S3 URL 전환 + 체크섬 검증 추가 | install.sh |
| 11 | `internal/cli/update.go` | S3 전환 + fallback + 체크섬 검증 | CLI update |
| 12 | `apps/web/src/app/layout.tsx` | JSON-LD installUrl 변경 | Landing |
| 13 | `README.md` | CLI 중심으로 재작성, Open Core 설명 추가 | 문서 |
| 14 | `CLAUDE.md` | Cloud 코드 제거, CLI + pkg/ 구조 반영 | 문서 |
| 15 | `LICENSE` | MIT 라이선스 파일 추가 (또는 확인) | 라이선스 |

삭제 대상:

| # | 파일/디렉토리 | 이유 |
|:-:|-------------|------|
| D1 | `cmd/server/` | Private repo로 이동 |
| D2 | `internal/api/` | Private repo로 이동 |
| D3 | `internal/auth/` | Private repo로 이동 |
| D4 | `internal/billing/` | Private repo로 이동 |
| D5 | `internal/repository/` | Private repo로 이동 |
| D6 | `apps/dashboard/` | Private repo로 이동 |
| D7 | `infra/terraform/` | Private repo로 이동 |
| D8 | `migrations/` | Private repo로 이동 |
| D9 | `Dockerfile.server` | Private repo로 이동 |
| D10 | `docker-compose.dev.yml` | Private repo로 이동 |
| D11 | `scripts/` | Private repo로 이동 |

### Private repo (agent-kay-it/tene-cloud) 신규

| # | 파일 | 내용 |
|:-:|------|------|
| 1 | `cmd/server/main.go` | API server entrypoint (기존 코드 그대로) |
| 2 | `internal/api/**` | Echo server, handlers, middleware, response, storage |
| 3 | `internal/auth/**` | OAuth + JWT (기존 코드 + import path 변경) |
| 4 | `internal/billing/**` | LemonSqueezy (기존 코드 + import path 변경) |
| 5 | `internal/repository/postgres/**` | PostgreSQL 구현체 (기존 코드 + import path 변경) |
| 6 | `apps/dashboard/**` | Dashboard Next.js 앱 (기존 코드 그대로) |
| 7 | `infra/terraform/**` | 13개 모듈 + environments (기존 코드 + S3 릴리스 버킷 추가) |
| 8 | `migrations/**` | PostgreSQL 마이그레이션 (기존 코드 그대로) |
| 9 | `Dockerfile.server` | API server Docker image (기존 코드 그대로) |
| 10 | `docker-compose.dev.yml` | Local dev (기존 코드 그대로) |
| 11 | `scripts/dev.sh` | Local dev orchestrator (기존 코드 그대로) |
| 12 | `scripts/sync-secrets.sh` | AWS Secrets Manager 동기화 (기존 코드 그대로) |
| 13 | `go.mod` | 신규 작성 (module github.com/agent-kay-it/tene-cloud + public repo 의존) |
| 14 | `go.sum` | 자동 생성 |
| 15 | `.github/workflows/ci.yml` | test + lint + deploy (섹션 4-2) |
| 16 | `.github/workflows/dashboard.yml` | Dashboard 빌드 검증 (섹션 4-2) |
| 17 | `CLAUDE.md` | Cloud API + Dashboard + Infra 중심 문서 |
| 18 | `.gitignore` | 기존 .gitignore 기반 |

---

## 11. 검증 체크리스트

### Public repo (agent-kay-it/tene)

- [ ] `go build ./cmd/tene` 성공
- [ ] `go test ./...` 성공
- [ ] `go vet ./...` 성공
- [ ] `golangci-lint run` 성공
- [ ] GoReleaser snapshot 빌드 성공: `goreleaser release --snapshot --clean`
- [ ] install.sh S3에서 다운로드 성공: `sh apps/web/public/install.sh`
- [ ] `tene update --check` S3에서 버전 조회 성공
- [ ] GitHub Actions CI (test + lint) 성공
- [ ] GitHub Actions Release (S3 업로드 + LATEST_VERSION) 성공
- [ ] Landing: `tene.sh` Vercel 배포 정상
- [ ] pkg/ 패키지가 외부에서 import 가능 확인

### Private repo (agent-kay-it/tene-cloud)

- [ ] `go build ./cmd/server` 성공
- [ ] `go test ./...` 성공
- [ ] `go vet ./...` 성공
- [ ] `golangci-lint run` 성공
- [ ] Docker build 성공: `docker build -f Dockerfile.server .`
- [ ] ECS 배포 성공 (staging)
- [ ] ECS 배포 성공 (prod)
- [ ] API 헬스체크: `curl https://api.tene.sh/health` 정상
- [ ] Dashboard: `app.tene.sh` Vercel 배포 정상
- [ ] GitHub Actions CI (test + lint + deploy) 성공
- [ ] `tene login` OAuth 정상 동작
- [ ] Shared 패키지(`pkg/domain`, `pkg/crypto`, `pkg/errors`) import 정상

### 통합 검증

- [ ] `tene init` -> `tene set` -> `tene get` -> `tene run` (CLI 로컬 기능)
- [ ] `tene login` -> `tene push` -> `tene pull` (Cloud sync)
- [ ] install.sh로 신규 설치 -> `tene update`로 업데이트
- [ ] Dashboard 로그인 -> vault 목록 조회

---

## 12. 롤백 계획

### Scenario A: Step 3 (pkg/ 이동) 이전 문제 발생

S3 인프라 + GoReleaser 변경만 된 상태이므로, 기존 GitHub Releases 경로가 여전히 동작한다.

```bash
# GoReleaser에서 blobs: 섹션 제거
# release.yml에서 OIDC + LATEST_VERSION 스텝 제거
# install.sh, update.go 원복
git revert HEAD~N  # 관련 커밋들 revert
```

소요: ~30분

### Scenario B: Step 4 (repo 분리) 이후 문제 발생

Private repo는 이미 생성되었으나, public repo에서 Cloud 코드 제거(Step 7) 전이라면:

```bash
# tene-cloud repo 삭제 (또는 무시)
# tene repo의 pkg/ 이동 커밋만 유지 (또는 revert)
# 기존 monorepo 상태로 복원
```

소요: ~1시간

### Scenario C: Step 7 (Cloud 코드 제거) 이후 문제 발생

가장 위험한 시나리오. Cloud 코드가 public repo에서 삭제된 상태.

```bash
# git reflog으로 삭제 전 커밋 확인
git reflog
git checkout <before-deletion-commit> -- cmd/server/ internal/api/ internal/auth/ ...

# 또는 tene-cloud repo에서 다시 복사
```

소요: ~2시간

**완화 전략**: Step 7 실행 전에 반드시 tene-cloud에서 모든 검증(섹션 11)을 통과한 후 진행한다.

---

## 13. 리스크 및 완화

| # | 리스크 | 영향 | 확률 | 완화 방안 |
|:-:|--------|------|:----:|----------|
| 1 | **pkg/ 이동 시 import 누락** | 빌드 실패 | 중간 | `go build ./...` + `go test ./...`로 전수 검증. IDE의 find-replace 활용. |
| 2 | **기존 CLI 사용자의 `tene update` 실패** | 구버전 CLI가 GitHub API 404 수신 | 높음 | S3 우선 + GitHub fallback 이중 시도 로직 (섹션 5-4). 전환 전 마지막 릴리스에 안내 추가. |
| 3 | **GoReleaser S3 업로드 실패** | 릴리스 바이너리 미배포 | 낮음 | GoReleaser snapshot으로 사전 검증. GitHub Releases를 dual publish로 유지. |
| 4 | **S3 버킷 public read 설정 오류** | install.sh 다운로드 403 | 중간 | Terraform plan에서 policy 확인. apply 후 즉시 curl 테스트. |
| 5 | **OIDC trust policy 두 repo 미지원** | CI/CD 인증 실패 | 중간 | IAM trust policy에 두 repo 모두 포함 (섹션 6-2). 사전 테스트. |
| 6 | **Vercel Dashboard 재연결 실패** | app.tene.sh 배포 중단 | 낮음 | 재연결 절차 문서화 (섹션 7-2). Vercel CLI 수동 배포 fallback. |
| 7 | **go.mod tidy 후 의존성 누락** | 빌드 실패 | 낮음 | `go mod tidy` 후 반드시 `go build ./...` 실행. |
| 8 | **pkg/ 패키지 API 변경 시 양쪽 영향** | Private repo 빌드 실패 | 중간 | Semantic versioning 준수. Breaking change 시 major 버전 업. |
| 9 | **OAuth callback URL 문제** | 로그인 불가 | 없음 | OAuth App은 repo visibility와 무관. 변경 불필요. |
| 10 | **git history 민감 정보 잔존** | 이전 public 기간 clone 복사본 존재 | 낮음 | Private 전환으로 추가 노출 차단. AWS IAM 정책이 방어선. |

---

## 14. 일정 (예상)

| 단계 | 작업 | 소요 | 누적 |
|------|------|:----:|:----:|
| **Day 1 AM** | Step 1: S3 릴리스 인프라 (Terraform) | 0.5일 | 0.5일 |
| **Day 1 PM** | Step 2: GoReleaser + install.sh + update.go | 0.5일 | 1일 |
| **Day 2 AM** | Step 3: Shared 패키지 internal/ -> pkg/ 이동 | 0.5일 | 1.5일 |
| **Day 2 PM** | Step 2 나머지 + Step 3 검증 | 0.5일 | 2일 |
| **Day 3** | Step 4: tene-cloud private repo 생성 + 코드 이동 | 1일 | 3일 |
| **Day 4 AM** | Step 5: CI/CD 파이프라인 분리 | 0.5일 | 3.5일 |
| **Day 4 PM** | Step 6: Vercel 재설정 | 0.5일 | 4일 |
| **Day 5 AM** | Step 7: Public repo Cloud 코드 제거 | 0.5일 | 4.5일 |
| **Day 5 PM** | Step 8: 검증 + 릴리스 | 0.5일 | 5일 |
| **Day 6** | Step 9: README, CLAUDE.md 업데이트 + 최종 검증 | 0.5일 | 5.5일 |
| **Buffer** | 예상치 못한 이슈 대응 | 1.5일 | **7일** |

**총 소요: 약 1주일**

### 의존 관계

```
Step 1 (S3 인프라)
  └── Step 2 (GoReleaser + install.sh + update.go)
        └── Step 3 (pkg/ 이동) ←── 독립적으로 병렬 가능
              └── Step 4 (tene-cloud 생성)
                    ├── Step 5 (CI/CD 분리)
                    └── Step 6 (Vercel 재설정)
                          └── Step 7 (Cloud 코드 제거) ←── Step 4, 5, 6 모두 완료 후
                                └── Step 8 (검증)
                                      └── Step 9 (문서 업데이트)
```

핵심 의존성: **Step 7 (Cloud 코드 제거)은 반드시 Step 4, 5, 6이 모두 완료되고 검증된 후 실행한다.** 이것이 롤백 난이도를 결정하는 분기점이다.
