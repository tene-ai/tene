# Tene Private Repository 전환 + S3 릴리스 기획서

> 작성일: 2026-04-09
> 상태: Draft
> 예상 소요: 3일

---

## 1. 개요

### 배경

Tene는 유료 SaaS(Agentic Secret Runtime)이지만, 현재 GitHub public repo(`agent-kay-it/tene`)에 전체 소스코드와 AWS 인프라 설정이 공개되어 있다.

노출 항목:
- Go CLI + Cloud API 전체 소스코드
- Terraform 13개 모듈 (VPC CIDR, IAM Role ARN, AWS Account ID `507221376909`)
- CI/CD 파이프라인 (ECR repo URL, ECS cluster/service명)
- 암호화 구현체 (XChaCha20-Poly1305, Argon2id 파라미터)

### 목적

1. **소스코드 보호**: 유료 제품의 핵심 로직 비공개 전환
2. **인프라 보안 강화**: AWS 리소스 식별 정보 노출 차단
3. **CLI 배포 경로 독립**: GitHub Releases 의존 제거, S3 자체 배포로 전환

### 범위

| 항목 | 포함 여부 |
|------|:---------:|
| GitHub repo private 전환 | O |
| CLI 바이너리 S3 배포 | O |
| GoReleaser S3 업로드 설정 | O |
| install.sh S3 URL 전환 | O |
| CLI update 명령어 S3 전환 | O |
| 랜딩페이지 GitHub 링크 수정 | O |
| Vercel 연동 확인 | O |
| In-memory state Redis 마이그레이션 | X (별도 작업) |

---

## 2. 현재 아키텍처 (AS-IS)

### 배포 파이프라인

```
[개발자] → git push → [GitHub Public Repo]
                            │
              ┌─────────────┼──────────────────┐
              │             │                   │
         [v* 태그]    [main push]          [PR 생성]
              │             │                   │
      GoReleaser      CI (test+lint)      Vercel Preview
              │             │
    GitHub Releases    Docker build
    (바이너리 5개)      → ECR push
              │             → ECS deploy
              │
    install.sh (GitHub API로 버전 조회)
    tene update (GitHub API로 최신 릴리스 조회)
    tene update (GitHub Releases에서 바이너리 다운로드)
```

### 서비스별 현재 상태

| 서비스 | 기술 | 배포 | GitHub 의존도 |
|--------|------|------|:------------:|
| CLI 바이너리 | GoReleaser | GitHub Releases | **높음** |
| install.sh | Shell script | Vercel static (`tene.sh/install.sh`) | **높음** (API + 다운로드) |
| CLI update | Go (`internal/cli/update.go`) | 런타임 GitHub API 호출 | **높음** |
| API Server | ECS Fargate | GitHub Actions → ECR → ECS | 낮음 (OIDC) |
| Dashboard | Next.js | Vercel (GitHub App) | 낮음 |
| Landing | Next.js | Vercel (GitHub App) | 낮음 |
| Terraform | 13 모듈 | 로컬 실행 | 없음 |

---

## 3. 목표 아키텍처 (TO-BE)

### Private Repo + S3 릴리스

```
[개발자] → git push → [GitHub Private Repo]
                            │
              ┌─────────────┼──────────────────┐
              │             │                   │
         [v* 태그]    [main push]          [PR 생성]
              │             │                   │
      GoReleaser      CI (test+lint)      Vercel Preview
              │             │                  (GitHub App 권한)
         S3 Upload     Docker build
              │          → ECR push
              │          → ECS deploy
     ┌────────┴────────┐
     │                 │
  tene-releases     LATEST_VERSION
  (public read)     (버전 파일)
     │
  install.sh (S3에서 버전 조회 + 다운로드)
  tene update (S3에서 최신 릴리스 조회 + 다운로드)
```

### S3 릴리스 버킷 구조

```
s3://tene-releases/
├── LATEST_VERSION                          # "0.3.1" (텍스트 파일)
├── v0.3.1/
│   ├── tene_0.3.1_darwin_amd64.tar.gz
│   ├── tene_0.3.1_darwin_arm64.tar.gz
│   ├── tene_0.3.1_linux_amd64.tar.gz
│   ├── tene_0.3.1_linux_arm64.tar.gz
│   ├── tene_0.3.1_windows_amd64.zip
│   └── checksums.txt
├── v0.3.0/
│   └── ...
```

### 변경 포인트 요약

| 구분 | AS-IS | TO-BE |
|------|-------|-------|
| 바이너리 호스팅 | GitHub Releases | S3 `tene-releases` (public read) |
| 버전 조회 | GitHub API (`/repos/.../releases/latest`) | S3 `LATEST_VERSION` 파일 |
| 다운로드 URL | `github.com/.../releases/download/...` | `releases.tene.sh/v{ver}/...` (CloudFront 또는 S3 직접) |
| 체크섬 검증 | 없음 | SHA-256 (`checksums.txt`) |
| repo 가시성 | Public | Private |
| GoReleaser 출력 | GitHub Release + assets | S3 upload + GitHub Release (private) |

---

## 4. 영향도 분석

### 분류표

| 분류 | 구성 요소 |
|------|-----------|
| **영향 없음** | CLI 로컬 명령어 (init, set, get, run, list, delete, export, import, passwd, recover, env), Cloud API 전체 (30+ 엔드포인트), Crypto/Sync 엔진, PostgreSQL/SQLite, GitHub OAuth callback URL, LemonSqueezy 연동, Terraform 모듈 (기존) |
| **설정 변경** | Vercel GitHub App 권한 재확인, CI/CD OIDC (private repo에서도 동작 확인), GitHub Actions secrets |
| **코드 변경** | GoReleaser 설정, release.yml 워크플로우, install.sh, CLI update 명령어, 랜딩페이지 GitHub 링크 (9곳), llms.txt, Terraform S3 모듈 확장 |

### 서비스별 영향도 상세

#### GoReleaser (`.goreleaser.yml`)

- **현재**: `release.github` 섹션으로 GitHub Releases에 업로드
- **변경**: `s3:` 블롭 섹션 추가, `release.disable: true`로 GitHub Release 비활성화 (또는 유지하되 private)
- **영향**: 빌드 프로세스 자체는 동일, 업로드 대상만 변경

#### CI/CD (`.github/workflows/release.yml`)

- **현재**: `GITHUB_TOKEN` 권한으로 GitHub Releases 업로드
- **변경**: AWS OIDC 인증 추가, S3 업로드 권한, LATEST_VERSION 파일 갱신 스텝 추가
- **영향**: private repo에서 `contents: write` 권한은 유효, OIDC도 repo 조건 충족

#### install.sh (`apps/web/public/install.sh`)

- **현재**: GitHub API로 최신 버전 조회, GitHub Releases URL로 다운로드
- **변경**: S3 LATEST_VERSION 파일로 버전 조회, S3 URL로 다운로드, 체크섬 검증 추가
- **영향**: 사용자 인스톨 경험 동일, 속도 향상 (S3 직접 서빙)

#### CLI update (`internal/cli/update.go`)

- **현재**: `fetchLatestRelease()` → GitHub API, 다운로드 URL → GitHub Releases
- **변경**: S3 LATEST_VERSION 조회, S3 URL로 다운로드
- **영향**: 기존 설치된 CLI는 GitHub URL 사용 → **기존 사용자 업데이트 한 번은 수동 필요** (또는 fallback 로직)

#### Landing Page (`apps/web/`)

- **현재**: GitHub 링크 9곳 (nav, footer, hero, security, cta, layout JSON-LD), llms.txt 1곳
- **변경**: 링크 제거 또는 대체 (docs, install 페이지 등)
- **영향**: UI/SEO만 영향, 기능 영향 없음

#### Vercel (Dashboard + Landing)

- **현재**: GitHub App 통한 자동 배포
- **변경**: private repo에서도 GitHub App 연동 유지 (Vercel에서 private repo 접근 권한 확인 필요)
- **영향**: Vercel Pro 플랜 불필요 (Hobby에서도 private repo 지원)

---

## 5. 실행 계획 (Phase별)

### Phase 1: S3 릴리스 인프라 구축 (Day 1)

#### 1-1. Terraform S3 모듈 확장

**파일: `infra/terraform/modules/s3/variables.tf`**
```hcl
# 추가
variable "create_release_bucket" {
  type    = bool
  default = false
}

variable "release_bucket_name" {
  type    = string
  default = ""
}

variable "release_domain" {
  type    = string
  default = ""
  description = "Custom domain for releases (e.g., releases.tene.sh)"
}
```

**파일: `infra/terraform/modules/s3/main.tf`**
```hcl
# 기존 vault + alb_logs 버킷 아래에 추가

# CLI release artifacts (public read)
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

resource "aws_s3_bucket_policy" "releases_public_read" {
  count  = var.create_release_bucket ? 1 : 0
  bucket = aws_s3_bucket.releases[0].id
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

resource "aws_s3_bucket_public_access_block" "releases" {
  count                   = var.create_release_bucket ? 1 : 0
  bucket                  = aws_s3_bucket.releases[0].id
  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}
```

**파일: `infra/terraform/modules/s3/outputs.tf`**
```hcl
# 추가
output "release_bucket_name" {
  value = var.create_release_bucket ? aws_s3_bucket.releases[0].id : ""
}

output "release_bucket_arn" {
  value = var.create_release_bucket ? aws_s3_bucket.releases[0].arn : ""
}

output "release_bucket_domain" {
  value = var.create_release_bucket ? aws_s3_bucket.releases[0].bucket_regional_domain_name : ""
}
```

**파일: `infra/terraform/environments/prod/main.tf`**
```hcl
# S3 모듈 호출에 추가
module "s3" {
  source = "../../modules/s3"

  project                = local.project
  environment            = local.environment
  aws_region             = local.region
  create_release_bucket  = true
  release_bucket_name    = "tene-releases"
}
```

#### 1-2. GitHub Actions IAM 정책 확장

**파일: `infra/terraform/modules/iam/main.tf`**

`github_actions_deploy` 정책에 S3 릴리스 버킷 업로드 권한 추가:
```hcl
# 기존 Statement 배열에 추가
{
  Effect   = "Allow"
  Action   = ["s3:PutObject", "s3:PutObjectAcl", "s3:ListBucket"]
  Resource = [var.release_bucket_arn, "${var.release_bucket_arn}/*"]
}
```

**파일: `infra/terraform/modules/iam/variables.tf`**
```hcl
# 추가
variable "release_bucket_arn" {
  type    = string
  default = ""
}
```

#### 1-3. (선택) Route53 releases 서브도메인

**파일: `infra/terraform/modules/route53/main.tf`**

`releases.tene.sh` CNAME → S3 버킷 엔드포인트 추가 (향후 CloudFront 전환 용이).
초기에는 S3 직접 URL (`https://tene-releases.s3.ap-northeast-2.amazonaws.com/...`) 사용 가능.

#### 1-4. Terraform Apply

```bash
cd infra/terraform/environments/prod
TF_VAR_db_password="$(tene get DB_PASSWORD --env prod)" \
TF_VAR_jwt_secret="$(tene get JWT_SECRET --env prod)" \
terraform plan   # S3 버킷 + IAM 변경 확인
terraform apply
```

---

### Phase 2: GoReleaser + CI/CD + install.sh 수정 (Day 1-2)

#### 2-1. GoReleaser S3 업로드 설정

**파일: `.goreleaser.yml`**

```yaml
# release 섹션 변경
release:
  github:
    owner: tomo-kay
    name: tene
  draft: false
  prerelease: auto
  name_template: "v{{.Version}}"
  # private repo에서도 GitHub Release는 유지 (내부 참조용)

# S3 업로드 섹션 추가
blobs:
  - provider: s3
    bucket: tene-releases
    region: ap-northeast-2
    directory: "v{{.Version}}"
    extra_files:
      - glob: dist/checksums.txt
```

#### 2-2. release.yml 워크플로우 수정

**파일: `.github/workflows/release.yml`**

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write
  id-token: write  # OIDC for AWS

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

#### 2-3. install.sh S3 전환 + 체크섬 추가

**파일: `apps/web/public/install.sh`**

변경 포인트 3곳:

**1) 상수 변경 (L10-11)**
```sh
# AS-IS
REPO="agent-kay-it/tene"

# TO-BE
RELEASE_BASE="https://tene-releases.s3.ap-northeast-2.amazonaws.com"
```

**2) `get_latest_version` 함수 (L39-49)**
```sh
# AS-IS: GitHub API 호출
get_latest_version() {
  curl -sSfL "https://api.github.com/repos/${REPO}/releases/latest" |
    grep '"tag_name"' | sed 's/.*"tag_name": *"v\([^"]*\)".*/\1/'
}

# TO-BE: S3 LATEST_VERSION 파일 조회
get_latest_version() {
  if command -v curl > /dev/null 2>&1; then
    curl -sSfL "${RELEASE_BASE}/LATEST_VERSION"
  elif command -v wget > /dev/null 2>&1; then
    wget -qO- "${RELEASE_BASE}/LATEST_VERSION"
  else
    error "curl or wget is required"
  fi
}
```

**3) 다운로드 URL + 체크섬 검증 (L75-83)**
```sh
# AS-IS
url="https://github.com/${REPO}/releases/download/v${version}/${filename}"
download "$url" "${tmpdir}/${filename}"
tar xzf "${tmpdir}/${filename}" -C "$tmpdir"

# TO-BE
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

#### 2-4. CLI update 명령어 S3 전환

**파일: `internal/cli/update.go`**

**1) 상수 추가 (파일 상단)**
```go
const releaseBaseURL = "https://tene-releases.s3.ap-northeast-2.amazonaws.com"
```

**2) `fetchLatestRelease` 함수 변경 (L226-252)**
```go
// AS-IS: GitHub API 호출
func fetchLatestRelease() (*githubRelease, error) {
    url := "https://api.github.com/repos/agent-kay-it/tene/releases/latest"
    // ...GitHub API 파싱...
}

// TO-BE: S3 LATEST_VERSION 파일 조회
func fetchLatestRelease() (*githubRelease, error) {
    url := releaseBaseURL + "/LATEST_VERSION"

    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("cannot check for updates: %w", err)
    }
    defer func() { _ = resp.Body.Close() }()

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("version check returned HTTP %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read version: %w", err)
    }

    ver := strings.TrimSpace(string(body))
    return &githubRelease{
        TagName: "v" + ver,
        HTMLURL: releaseBaseURL + "/v" + ver + "/",
    }, nil
}
```

**3) 다운로드 URL 변경 (L119)**
```go
// AS-IS
downloadURL := fmt.Sprintf("https://github.com/agent-kay-it/tene/releases/download/%s/%s", targetVersion, assetName)

// TO-BE
downloadURL := fmt.Sprintf("%s/%s/%s", releaseBaseURL, targetVersion, assetName)
```

**4) (선택) 체크섬 검증 추가**

`downloadBinaryFromTarGz` 호출 전후에 `checksums.txt` 다운로드 + SHA-256 검증 로직 추가 권장.

---

### Phase 3: 랜딩 페이지 GitHub 링크 수정 (Day 2)

총 10곳 수정 필요. GitHub repo 링크를 제거하거나 docs/install 페이지로 대체.

| # | 파일 | 위치 | 현재 | 변경 |
|:-:|------|------|------|------|
| 1 | `apps/web/src/components/nav.tsx` | L55 | GitHub 아이콘 링크 | 제거 또는 docs 링크로 대체 |
| 2 | `apps/web/src/components/nav.tsx` | L70 | 모바일 GitHub 링크 | 제거 또는 docs 링크로 대체 |
| 3 | `apps/web/src/components/footer.tsx` | L17 | GitHub 링크 | 제거 |
| 4 | `apps/web/src/components/footer.tsx` | L25 | Issues 링크 | 이메일/Discord 등 대체 채널로 변경 |
| 5 | `apps/web/src/components/security.tsx` | L78-83 | "소스코드 공개" 링크 + 텍스트 | "감사 가능한 암호화" 등 대체 문구로 변경 |
| 6 | `apps/web/src/components/cta.tsx` | L22 | GitHub 링크 | 제거 또는 docs 링크로 대체 |
| 7 | `apps/web/src/data/hero.ts` | L11 | hero 섹션 GitHub 링크 | docs 또는 install 가이드로 대체 |
| 8 | `apps/web/src/app/layout.tsx` | L150 | JSON-LD installUrl | `https://tene.sh/install.sh` (GitHub Releases 언급 제거) |
| 9 | `apps/web/public/llms.txt` | L77 | `GitHub: https://github.com/agent-kay-it/tene` | 라인 제거 또는 `Website: https://tene.sh`로 변경 |

#### security.tsx 특별 처리

현재 "오픈소스" 관련 마케팅 문구가 있을 수 있음. Private 전환 시 메시지 전략:
- "오픈소스" → "감사 가능한 암호화 (Auditable Cryptography)"
- 암호화 알고리즘 명시 (XChaCha20-Poly1305, Argon2id)는 공개 유지
- 향후 Open Core 모델 검토 시점까지 보류

---

### Phase 4: Private 전환 + 검증 (Day 2-3)

#### 4-1. 사전 검증 (Private 전환 전)

```bash
# 1. S3 릴리스 버킷 접근 확인
curl -I https://tene-releases.s3.ap-northeast-2.amazonaws.com/LATEST_VERSION

# 2. GoReleaser dry-run
goreleaser release --snapshot --clean

# 3. install.sh 로컬 테스트
sh apps/web/public/install.sh

# 4. Vercel 프로젝트 설정 확인
# - tene (Landing): GitHub App 권한에 private repo 포함 여부
# - tene-dashboard: 동일 확인

# 5. GitHub Actions OIDC 조건 확인
# IAM: "repo:agent-kay-it/tene:*" → private 전환 시에도 유효 (repo명 불변)
```

#### 4-2. Private 전환 실행

1. GitHub repo Settings → Danger Zone → Change repository visibility → Private
2. 즉시 확인:
   - `curl https://api.github.com/repos/agent-kay-it/tene` → 404 (정상)
   - Vercel 배포 트리거 (빈 커밋 또는 수동)
   - GitHub Actions 워크플로우 실행 확인

#### 4-3. 전환 직후 검증 체크리스트

| # | 검증 항목 | 방법 | 기대 결과 |
|:-:|----------|------|----------|
| 1 | API 서버 정상 | `curl https://api.tene.sh/health` | `{"status":"ok"}` |
| 2 | install.sh 동작 | `curl -sSfL https://tene.sh/install.sh \| sh` | 설치 성공 |
| 3 | tene update 동작 | `tene update --check` | 버전 정보 출력 |
| 4 | Vercel Landing 배포 | `curl -I https://tene.sh` | 200 OK |
| 5 | Vercel Dashboard 배포 | `curl -I https://app.tene.sh` | 200 OK |
| 6 | GitHub Actions CI | main push 후 확인 | test + lint + deploy 성공 |
| 7 | GitHub Actions Release | 테스트 태그 push 후 확인 | S3 업로드 + LATEST_VERSION 갱신 |
| 8 | OAuth 로그인 | `tene login` | GitHub OAuth 정상 동작 |
| 9 | GitHub repo 404 | 브라우저에서 repo URL 접근 | 404 (비인증 사용자) |

---

### Phase 5: 사후 검증 (Day 3)

1. **기존 사용자 업데이트 경로 확인**: 구버전 CLI (`tene update`)가 GitHub API 404를 받는 경우 에러 메시지 확인
2. **S3 버킷 접근 로그 확인**: 다운로드 정상 기록
3. **Vercel 빌드 로그 확인**: private repo에서 정상 빌드
4. **SEO 영향 확인**: Google Search Console에서 크롤링 에러 모니터링 (GitHub 링크 404)
5. **llms.txt 반영 확인**: AI 크롤러 대응

---

## 6. 변경 파일 목록 (체크리스트)

### Terraform (Phase 1)

| # | 파일 | 변경 내용 | 상태 |
|:-:|------|----------|:----:|
| 1 | `infra/terraform/modules/s3/variables.tf` | `create_release_bucket`, `release_bucket_name`, `release_domain` 변수 추가 | [ ] |
| 2 | `infra/terraform/modules/s3/main.tf` | releases 버킷 리소스 추가 (bucket, versioning, policy, CORS, public access) | [ ] |
| 3 | `infra/terraform/modules/s3/outputs.tf` | `release_bucket_name`, `release_bucket_arn`, `release_bucket_domain` 출력 추가 | [ ] |
| 4 | `infra/terraform/modules/iam/main.tf` | `github_actions_deploy` 정책에 S3 릴리스 버킷 권한 추가 | [ ] |
| 5 | `infra/terraform/modules/iam/variables.tf` | `release_bucket_arn` 변수 추가 | [ ] |
| 6 | `infra/terraform/environments/prod/main.tf` | S3 모듈에 `create_release_bucket=true` 전달, IAM 모듈에 `release_bucket_arn` 전달 | [ ] |

### GoReleaser + CI/CD (Phase 2)

| # | 파일 | 변경 내용 | 상태 |
|:-:|------|----------|:----:|
| 7 | `.goreleaser.yml` | `blobs:` 섹션 추가 (S3 provider, bucket, region, directory) | [ ] |
| 8 | `.github/workflows/release.yml` | OIDC 인증 추가, LATEST_VERSION 갱신 스텝 추가, `id-token: write` 퍼미션 | [ ] |

### CLI (Phase 2)

| # | 파일 | 변경 내용 | 상태 |
|:-:|------|----------|:----:|
| 9 | `internal/cli/update.go` | `releaseBaseURL` 상수 추가, `fetchLatestRelease` S3 전환, 다운로드 URL 변경 | [ ] |
| 10 | `apps/web/public/install.sh` | `RELEASE_BASE` 상수, `get_latest_version` S3 전환, 체크섬 검증 추가 | [ ] |

### Landing Page (Phase 3)

| # | 파일 | 변경 내용 | 상태 |
|:-:|------|----------|:----:|
| 11 | `apps/web/src/components/nav.tsx` | GitHub 링크 2곳 제거/대체 (L55, L70) | [ ] |
| 12 | `apps/web/src/components/footer.tsx` | GitHub 링크 제거 (L17), Issues 링크 대체 (L25) | [ ] |
| 13 | `apps/web/src/components/security.tsx` | 오픈소스 링크/문구 변경 (L78-83) | [ ] |
| 14 | `apps/web/src/components/cta.tsx` | GitHub 링크 제거/대체 (L22) | [ ] |
| 15 | `apps/web/src/data/hero.ts` | GitHub 링크 대체 (L11) | [ ] |
| 16 | `apps/web/src/app/layout.tsx` | JSON-LD installUrl에서 GitHub 언급 제거 (L150) | [ ] |
| 17 | `apps/web/public/llms.txt` | GitHub URL 제거/변경 (L77) | [ ] |

### 총합: 17개 파일

---

## 7. 롤백 계획

### Private → Public 복원 (즉시 가능)

Private 전환 후 문제 발생 시:

1. GitHub repo Settings → Change visibility → Public
2. 즉시 복원됨 (stars, forks 등은 소멸 — 현재 초기 단계이므로 영향 미미)

### S3 릴리스 → GitHub Releases 복원

1. `.goreleaser.yml`에서 `blobs:` 섹션 제거
2. `release.yml`에서 OIDC + LATEST_VERSION 스텝 제거
3. `install.sh`, `update.go` 원복 (git revert)
4. 태그 push로 GitHub Releases 재생성

### 복원 소요 시간

| 시나리오 | 소요 |
|----------|------|
| Repo public 복원 | ~1분 |
| S3 → GitHub Releases 복원 | ~30분 (코드 revert + 태그 push) |
| 전체 원복 | ~1시간 |

---

## 8. 리스크 및 완화 방안

| # | 리스크 | 영향 | 확률 | 완화 방안 |
|:-:|--------|------|:----:|----------|
| 1 | 기존 CLI 사용자의 `tene update` 실패 | 구버전 CLI가 GitHub API 404 수신 | 높음 | install.sh 재설치 안내 에러 메시지 (이미 L131에 fallback 문구 존재), 전환 전 릴리스 노트에 안내 |
| 2 | Vercel private repo 빌드 실패 | Landing + Dashboard 배포 중단 | 낮음 | 전환 전 Vercel GitHub App 권한 확인, Hobby 플랜에서 private repo 지원 확인 |
| 3 | GoReleaser S3 업로드 실패 | 릴리스 바이너리 미배포 | 낮음 | dry-run으로 사전 검증, GitHub Releases는 fallback으로 유지 |
| 4 | S3 버킷 퍼블릭 설정 오류 | install.sh 다운로드 403 | 중간 | Terraform plan에서 policy 확인, 배포 후 curl 테스트 |
| 5 | OIDC 인증 실패 (private repo) | CI/CD 전체 중단 | 낮음 | IAM 조건 `repo:agent-kay-it/tene:*`는 visibility와 무관, 사전 테스트 가능 |
| 6 | SEO 하락 (GitHub 백링크 손실) | 검색 노출 감소 | 중간 | 랜딩페이지 SEO 강화, GitHub 프로필에 tene.sh 링크 유지 |
| 7 | 오픈소스 신뢰도 하락 | 보안 제품 특성상 코드 공개 기대 | 중간 | 암호화 알고리즘 공개 문서화, 향후 crypto 모듈 Open Core 검토 |

---

## 9. 비용 영향

### 추가 비용

| 항목 | 예상 비용 | 비고 |
|------|:---------:|------|
| S3 `tene-releases` 스토리지 | ~$0.01/월 | 5개 플랫폼 × ~10MB = ~50MB/릴리스 |
| S3 GET 요청 | ~$0.01/월 | 초기 트래픽 1,000회 미만 |
| S3 데이터 전송 | ~$0.05/월 | 50MB × 100회 = 5GB |
| Route53 (releases 서브도메인) | ~$0.50/월 | 쿼리 비용 포함 |
| **합계** | **~$0.57/월** | 기존 $55/월 대비 +1% |

### 비용 절감

| 항목 | 절감 | 비고 |
|------|:----:|------|
| GitHub Actions 분 | $0 | private repo도 무료 2,000분/월 |

### 트래픽 증가 시

| 다운로드 수/월 | S3 비용 |
|:--------------:|:-------:|
| 100 | ~$0.05 |
| 1,000 | ~$0.50 |
| 10,000 | ~$5.00 |
| 100,000 | ~$50 (CloudFront 도입 검토 시점) |

---

## 10. 향후 로드맵

### 단기 (전환 직후)

- [ ] CloudFront CDN 도입 검토 (`releases.tene.sh` → CloudFront → S3)
- [ ] Homebrew tap 활성화 (`.goreleaser.yml` 주석 해제, private tap repo)
- [ ] CLI 체크섬 검증 강화 (`tene update`에 SHA-256 검증 추가)

### 중기 (1-3개월)

- [ ] Open Core 전환 검토
  - `internal/crypto/` 모듈을 별도 public repo로 분리
  - CLI 기본 기능(init, set, get, run)은 오픈소스
  - Cloud sync, team, billing은 proprietary
- [ ] In-memory state → Redis/PostgreSQL 마이그레이션 (production blocker)
- [ ] 릴리스 서명 (cosign 또는 GPG) 도입

### 장기 (3-6개월)

- [ ] 가격 구조 개편 (현재 $0 Free / $5 Pro)
- [ ] Enterprise 플랜 추가 (self-hosted, audit, SSO)
- [ ] CLI 플러그인 시스템 (오픈 에코시스템)
- [ ] 보안 감사 보고서 공개 (코드 대신 감사 결과로 신뢰 확보)

---

## 부록: 기존 사용자 마이그레이션 안내 (참고)

Private 전환 후 구버전 CLI 사용자가 `tene update` 실행 시:

```
Error: failed to check for updates: GitHub API returned 404

  Try manually: curl -sSfL https://tene.sh/install.sh | sh
```

이미 `internal/cli/update.go` L131에 fallback 안내 문구가 포함되어 있으므로, `fetchLatestRelease` 실패 시 사용자는 install.sh로 재설치할 수 있다.

전환 전 마지막 public 릴리스에 안내 노트 추가 권장:

```markdown
## Migration Notice
Starting from the next release, CLI binaries are distributed via S3.
To update, run: curl -sSfL https://tene.sh/install.sh | sh
```
