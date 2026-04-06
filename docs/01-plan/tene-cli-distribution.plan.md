# Tene CLI 배포/버전관리 계획

> 작성일: 2026-04-06
> Go 오픈소스 프로젝트 심층 조사 기반

---

## 1. 오픈소스 프로젝트 배포 방식 비교표

| 프로젝트 | 빌드 도구 | GitHub Releases | Homebrew | go install | install.sh | 패키지 매니저 | 버전 규칙 |
|----------|----------|----------------|----------|------------|------------|-------------|----------|
| **gh** (GitHub CLI) | GoReleaser | tar.gz/zip + 체크섬 | homebrew-core | X (복잡한 빌드) | X | apt, scoop, winget, snap | semver (v2.x.x) |
| **lazygit** | GoReleaser | tar.gz/zip + checksums.txt | homebrew-core | O | X | scoop, choco, pacman | semver |
| **fzf** | 자체 Makefile | tar.gz + checksums | homebrew-core | O (fallback) | O (git clone + install) | apt, pacman, scoop | semver |
| **cobra-cli** | 없음 (라이브러리) | GitHub 태그만 | X | O (주요 배포) | X | X | semver |
| **age** | 자체 빌드 | tar.gz + Sigsum 증명 | homebrew-core | O | X | apt, pacman, winget | semver |
| **direnv** | GoReleaser | tar.gz + checksums | homebrew-core | O | O | apt, pacman, nix | semver |
| **chezmoi** | GoReleaser | tar.gz/zip + checksums | homebrew tap + core | O | O | snap, scoop, pacman | semver |

### 핵심 발견사항

1. **GoReleaser가 사실상 표준**: gh, lazygit, chezmoi, direnv 등 대부분의 Go CLI 프로젝트가 GoReleaser 사용
2. **GitHub Releases가 기본**: 모든 프로젝트가 GitHub Releases를 1차 배포 채널로 사용
3. **go install 지원이 일반적**: 간단한 프로젝트는 `go install`을 주요 설치 방법으로 안내
4. **Homebrew는 2단계**: 처음에는 homebrew tap, 인기 후 homebrew-core로 이동
5. **install.sh는 선택적**: fzf, chezmoi, direnv 정도만 제공. 필수는 아님
6. **Semantic Versioning 필수**: 모든 프로젝트가 semver (vX.Y.Z) 사용

---

## 2. Tene CLI 배포 전략

### Phase 1 — MVP (지금 즉시)
- `go install github.com/tomo-kay/tene/cmd/tene@latest`
- GitHub Releases (GoReleaser로 자동화)
- 크로스 컴파일 바이너리 + 체크섬

### Phase 2 — 안정화 후 (v0.5.0+)
- Homebrew tap (`brew install tomo-kay/tap/tene`)
- install.sh 스크립트 (curl 한 줄 설치)

### Phase 3 — 인기 후 (v1.0.0+)
- homebrew-core 등록
- scoop (Windows), snap (Linux) 등록
- apt/yum 패키지

---

## 3. goreleaser 설정

### 3.1 `.goreleaser.yml` (바로 사용 가능)

```yaml
# .goreleaser.yml
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
    # Windows arm64는 아직 사용자가 적어 제외
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

# Phase 2: Homebrew tap (주석 해제하여 활성화)
# brews:
#   - name: tene
#     homepage: "https://github.com/tomo-kay/tene"
#     description: "Agentic Secret Runtime Platform - Secure secret management for AI agents"
#     license: "MIT"
#     directory: Formula
#     commit_author:
#       name: tomo-kay
#       email: tomo-kay@users.noreply.github.com
#     commit_msg_template: "Brew formula update for {{ .ProjectName }} version {{ .Tag }}"
#     repository:
#       owner: tomo-kay
#       name: homebrew-tap
#       branch: main
#       token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
#     install: |
#       bin.install "tene"
#     test: |
#       system "#{bin}/tene", "version"
```

### 3.2 ldflags 버전 주입 설명

GoReleaser는 빌드 시 자동으로 다음 변수를 주입한다:

| 템플릿 변수 | 설명 | 예시 |
|------------|------|------|
| `{{.Version}}` | Git 태그 (v 접두사 제거) | `0.1.0` |
| `{{.ShortCommit}}` | Git 커밋 SHA (짧은 형태) | `abc1234` |
| `{{.Date}}` | 빌드 시간 (RFC3339) | `2026-04-06T12:00:00Z` |

현재 `cmd/tene/main.go`의 `var version, commit, date`에 주입된다.

### 3.3 크로스 컴파일 대상 (5개 바이너리)

| OS | Arch | 바이너리 이름 | 아카이브 형식 |
|----|------|-------------|-------------|
| darwin | amd64 | tene | tar.gz |
| darwin | arm64 | tene | tar.gz |
| linux | amd64 | tene | tar.gz |
| linux | arm64 | tene | tar.gz |
| windows | amd64 | tene.exe | zip |

---

## 4. GitHub Actions 워크플로우

### 4.1 `.github/workflows/release.yml` (바로 사용 가능)

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

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
          go-version: '1.22'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v7
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          # Phase 2: Homebrew tap 배포 시 필요
          # HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
```

### 4.2 `.github/workflows/ci.yml` (바로 사용 가능)

```yaml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Download dependencies
        run: go mod download

      - name: Vet
        run: go vet ./...

      - name: Test
        run: go test ./... -v -count=1 -race

  lint:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  build:
    runs-on: ubuntu-latest
    needs: [test, lint]
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build
        run: go build ./cmd/tene

      - name: GoReleaser Check
        uses: goreleaser/goreleaser-action@v7
        with:
          distribution: goreleaser
          version: "~> v2"
          args: check
```

---

## 5. Homebrew tap 설정 (Phase 2)

### 5.1 homebrew-tap 리포지토리 생성

```bash
# 1. GitHub에 새 리포 생성: tomo-kay/homebrew-tap
# 2. 리포에 빈 Formula 디렉토리 생성
mkdir -p Formula
touch Formula/.gitkeep
git add . && git commit -m "Initial commit" && git push
```

### 5.2 GitHub Personal Access Token 생성

1. GitHub Settings > Developer settings > Personal access tokens > Fine-grained tokens
2. 권한: `homebrew-tap` 리포에 대한 `Contents: Read and write` 권한
3. 토큰을 `tene` 리포의 Secrets에 `HOMEBREW_TAP_GITHUB_TOKEN`으로 저장

### 5.3 goreleaser에서 brews 섹션 활성화

위 `.goreleaser.yml`의 `brews` 섹션 주석을 해제하면, 태그 push 시 자동으로:
1. GoReleaser가 바이너리 빌드 + GitHub Release 생성
2. Homebrew Formula 파일을 자동 생성
3. `tomo-kay/homebrew-tap` 리포에 Formula를 커밋

### 5.4 사용자 설치 플로우

```bash
# 최초 1회: tap 등록
brew tap tomo-kay/tap

# 설치
brew install tomo-kay/tap/tene

# 또는 한 줄로
brew install tomo-kay/tap/tene

# 업그레이드
brew upgrade tene
```

### 5.5 자동 생성되는 Formula 예시

GoReleaser가 자동으로 생성하는 `Formula/tene.rb`:

```ruby
class Tene < Formula
  desc "Agentic Secret Runtime Platform - Secure secret management for AI agents"
  homepage "https://github.com/tomo-kay/tene"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/tomo-kay/tene/releases/download/v0.1.0/tene_0.1.0_darwin_arm64.tar.gz"
      sha256 "<자동생성>"
    end
    if Hardware::CPU.intel?
      url "https://github.com/tomo-kay/tene/releases/download/v0.1.0/tene_0.1.0_darwin_amd64.tar.gz"
      sha256 "<자동생성>"
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/tomo-kay/tene/releases/download/v0.1.0/tene_0.1.0_linux_arm64.tar.gz"
      sha256 "<자동생성>"
    end
    if Hardware::CPU.intel?
      url "https://github.com/tomo-kay/tene/releases/download/v0.1.0/tene_0.1.0_linux_amd64.tar.gz"
      sha256 "<자동생성>"
    end
  end

  def install
    bin.install "tene"
  end

  test do
    system "#{bin}/tene", "version"
  end
end
```

---

## 6. install.sh 스크립트 (Phase 2)

### `install.sh` (바로 사용 가능)

```bash
#!/bin/sh
# Tene CLI 설치 스크립트
# 사용법: curl -fsSL https://raw.githubusercontent.com/tomo-kay/tene/main/install.sh | sh

set -e

REPO="tomo-kay/tene"
BINARY_NAME="tene"
INSTALL_DIR="/usr/local/bin"

# --- 색상 출력 ---
info() { printf "\033[0;34m[info]\033[0m %s\n" "$1"; }
error() { printf "\033[0;31m[error]\033[0m %s\n" "$1" >&2; exit 1; }
success() { printf "\033[0;32m[ok]\033[0m %s\n" "$1"; }

# --- OS 감지 ---
detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) error "지원하지 않는 OS: $(uname -s)" ;;
    esac
}

# --- 아키텍처 감지 ---
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) error "지원하지 않는 아키텍처: $(uname -m)" ;;
    esac
}

# --- 최신 버전 조회 ---
get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" |
        grep '"tag_name"' |
        sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
}

# --- 메인 ---
main() {
    OS=$(detect_os)
    ARCH=$(detect_arch)
    VERSION=${TENE_VERSION:-$(get_latest_version)}

    if [ -z "$VERSION" ]; then
        error "최신 버전을 가져올 수 없습니다. GitHub API 확인 필요."
    fi

    # v 접두사 제거
    VERSION_NUM=$(echo "$VERSION" | sed 's/^v//')

    info "Tene ${VERSION} 설치 중... (${OS}/${ARCH})"

    # 아카이브 확장자 결정
    EXT="tar.gz"
    if [ "$OS" = "windows" ]; then
        EXT="zip"
    fi

    FILENAME="${BINARY_NAME}_${VERSION_NUM}_${OS}_${ARCH}.${EXT}"
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"
    CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

    # 임시 디렉토리
    TMP_DIR=$(mktemp -d)
    trap "rm -rf ${TMP_DIR}" EXIT

    info "다운로드 중: ${URL}"
    curl -fsSL -o "${TMP_DIR}/${FILENAME}" "${URL}"

    # 체크섬 검증
    info "체크섬 검증 중..."
    curl -fsSL -o "${TMP_DIR}/checksums.txt" "${CHECKSUM_URL}"
    EXPECTED=$(grep "${FILENAME}" "${TMP_DIR}/checksums.txt" | awk '{print $1}')
    if command -v sha256sum > /dev/null 2>&1; then
        ACTUAL=$(sha256sum "${TMP_DIR}/${FILENAME}" | awk '{print $1}')
    elif command -v shasum > /dev/null 2>&1; then
        ACTUAL=$(shasum -a 256 "${TMP_DIR}/${FILENAME}" | awk '{print $1}')
    else
        info "sha256sum/shasum을 찾을 수 없어 체크섬 검증을 건너뜁니다."
        ACTUAL="$EXPECTED"
    fi

    if [ "$EXPECTED" != "$ACTUAL" ]; then
        error "체크섬 불일치! 예상: ${EXPECTED}, 실제: ${ACTUAL}"
    fi
    success "체크섬 검증 완료"

    # 압축 해제
    info "압축 해제 중..."
    if [ "$EXT" = "tar.gz" ]; then
        tar -xzf "${TMP_DIR}/${FILENAME}" -C "${TMP_DIR}"
    else
        unzip -q "${TMP_DIR}/${FILENAME}" -d "${TMP_DIR}"
    fi

    # 설치
    if [ -w "$INSTALL_DIR" ]; then
        mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        info "sudo 권한이 필요합니다."
        sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    success "Tene ${VERSION} 설치 완료! (${INSTALL_DIR}/${BINARY_NAME})"
    info "확인: tene version"
}

main
```

---

## 7. 버전 관리 워크플로우

### 7.1 Semantic Versioning 규칙

```
v<MAJOR>.<MINOR>.<PATCH>[-<pre-release>]

예시:
  v0.1.0        - 최초 릴리스 (불안정)
  v0.1.1        - 버그 수정
  v0.2.0        - 새 기능 추가
  v0.2.0-beta.1 - 베타 릴리스
  v1.0.0        - 첫 안정 릴리스 (breaking change 가능성 명시)
```

- **MAJOR**: 호환되지 않는 API 변경 (0.x.x 동안은 불안정)
- **MINOR**: 하위 호환 기능 추가
- **PATCH**: 하위 호환 버그 수정
- **Pre-release**: `-beta.1`, `-rc.1` 등

### 7.2 릴리스 워크플로우 (구체적 명령어)

```bash
# 1. 개발 완료 후 최종 확인
go test ./... -v -count=1
go vet ./...
go build ./cmd/tene

# 2. CHANGELOG 업데이트 (선택)
# CHANGELOG.md에 변경사항 기록

# 3. 변경사항 커밋
git add -A
git commit -m "준비: v0.1.0 릴리스"

# 4. 태그 생성
git tag -a v0.1.0 -m "Release v0.1.0: 최초 릴리스"

# 5. 태그 push (GitHub Actions 자동 트리거)
git push origin main
git push origin v0.1.0

# 6. 결과 확인
# GitHub > Releases 페이지에서 자동 생성된 릴리스 확인
# - tene_0.1.0_darwin_amd64.tar.gz
# - tene_0.1.0_darwin_arm64.tar.gz
# - tene_0.1.0_linux_amd64.tar.gz
# - tene_0.1.0_linux_arm64.tar.gz
# - tene_0.1.0_windows_amd64.zip
# - checksums.txt
```

### 7.3 Pre-release 릴리스

```bash
# 베타 릴리스
git tag -a v0.2.0-beta.1 -m "Beta: v0.2.0-beta.1"
git push origin v0.2.0-beta.1
# goreleaser의 prerelease: auto 설정으로 자동으로 pre-release로 표시됨
```

### 7.4 전체 플로우 다이어그램

```
개발 → commit → tag push → GitHub Actions 트리거
                                    │
                          ┌─────────┴─────────┐
                          │  GoReleaser 실행   │
                          └─────────┬─────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
              크로스 컴파일    GitHub Release    체크섬 생성
              (5개 바이너리)    (자동 생성)     (checksums.txt)
                    │               │               │
                    └───────────────┼───────────────┘
                                    │
                          (Phase 2) Homebrew tap 업데이트
```

---

## 8. go install 지원

### 8.1 현재 상태

```bash
go install github.com/tomo-kay/tene/cmd/tene@latest
```

이 명령어가 동작하려면:
1. `go.mod`에 **`replace` directive가 없어야 한다** (현재 없음 -- OK)
2. GitHub에 **Git 태그가 있어야 한다** (`v0.1.0` 등)
3. 모듈 경로가 정확해야 한다 (`github.com/tomo-kay/tene`)

### 8.2 주의사항

- `replace` directive가 있으면 `go install <pkg>@latest`가 실패한다. 로컬 개발 시 replace를 사용했다면 커밋 전 반드시 제거해야 한다.
- `go install`은 `CGO_ENABLED=0`이 아닐 수 있으므로 CGO 의존성이 있으면 문제가 될 수 있다. Tene는 `modernc.org/sqlite`를 사용하고 있어 pure Go SQLite이므로 CGO 없이 빌드 가능하다.
- 태그가 없으면 `@latest`가 최신 커밋을 가져온다. 안정적 배포를 위해 반드시 semver 태그를 붙여야 한다.

### 8.3 go install vs GitHub Releases 비교

| 항목 | go install | GitHub Releases |
|------|-----------|----------------|
| 사전 조건 | Go 설치 필요 | 없음 (바이너리) |
| 빌드 | 사용자 로컬에서 빌드 | 미리 빌드된 바이너리 |
| 속도 | 느림 (컴파일 필요) | 빠름 (다운로드만) |
| 버전 지정 | `@v0.1.0`, `@latest` | 릴리스 페이지에서 선택 |
| 대상 사용자 | Go 개발자 | 모든 사용자 |

---

## 9. README 설치 명령어 (Phase별)

### Phase 1 (지금 즉시)

```markdown
## Installation

### Pre-built binaries (recommended)
Download from [GitHub Releases](https://github.com/tomo-kay/tene/releases/latest).

### Go developers
```bash
go install github.com/tomo-kay/tene/cmd/tene@latest
```

### Build from source
```bash
git clone https://github.com/tomo-kay/tene.git
cd tene
make build
# Binary: ./bin/tene
```
```

### Phase 2 (v0.5.0+ 추가)

```markdown
### macOS / Linux (Homebrew)
```bash
brew install tomo-kay/tap/tene
```

### Quick install (macOS / Linux)
```bash
curl -fsSL https://raw.githubusercontent.com/tomo-kay/tene/main/install.sh | sh
```
```

### Phase 3 (v1.0.0+ 추가)

```markdown
### macOS (Homebrew)
```bash
brew install tene
```

### Windows (Scoop)
```powershell
scoop install tene
```
```

---

## 10. 즉시 실행 체크리스트

### 지금 바로 해야 할 것

- [ ] `.goreleaser.yml` 파일을 프로젝트 루트에 생성 (위 3.1 섹션 복사)
- [ ] `.github/workflows/release.yml` 생성 (위 4.1 섹션 복사)
- [ ] `.github/workflows/ci.yml` 생성 (위 4.2 섹션 복사)
- [ ] Makefile ldflags 경로 확인: `main.version` vs `internal/cli.version`
- [ ] 첫 번째 릴리스 태그: `git tag -a v0.1.0 -m "Initial release"`
- [ ] 태그 push: `git push origin v0.1.0`
- [ ] GitHub Releases 페이지에서 결과 확인

### Makefile ldflags 주의사항

현재 Makefile의 ldflags:
```
-X github.com/tomo-kay/tene/internal/cli.version=$(VERSION)
```

GoReleaser의 ldflags:
```
-X main.version={{.Version}}
```

`cmd/tene/main.go`에서 `var version`이 선언되어 있으므로 `main.version`이 올바르다.
단, `internal/cli/root.go`에도 `var version`이 있고 `SetVersion()`으로 전달하는 구조이므로,
**GoReleaser에서는 `main.version`을 사용하는 것이 맞다** (main.go에서 cli.SetVersion()으로 전달).

### v0.5.0 이후 해야 할 것

- [ ] `tomo-kay/homebrew-tap` 리포지토리 생성
- [ ] Personal Access Token 생성 및 시크릿 등록
- [ ] `.goreleaser.yml`의 brews 섹션 주석 해제
- [ ] `install.sh` 스크립트 프로젝트 루트에 추가
- [ ] README 업데이트

---

## 참고 자료

- [GoReleaser 공식 문서 - GitHub Actions](https://goreleaser.com/ci/actions/)
- [GoReleaser 공식 문서 - Homebrew Taps](https://goreleaser.com/customization/homebrew/)
- [GoReleaser 공식 문서 - ldflags](https://goreleaser.com/cookbooks/using-main.version/)
- [GoReleaser 공식 문서 - Builds](https://goreleaser.com/customization/builds/)
- [GitHub CLI .goreleaser.yml](https://github.com/cli/cli/blob/trunk/.goreleaser.yml)
- [lazygit .goreleaser.yml](https://github.com/jesseduffield/lazygit/blob/master/.goreleaser.yml)
- [chezmoi .goreleaser.yaml](https://github.com/twpayne/chezmoi/blob/master/.goreleaser.yaml)
- [Go Modules Reference - replace directive](https://go.dev/ref/mod)
