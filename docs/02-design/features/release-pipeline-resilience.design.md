# Release Pipeline Resilience — Design

| 항목 | 값 |
|------|-----|
| Feature | `release-pipeline-resilience` |
| Plan | `docs/01-plan/release-pipeline-resilience.plan.md` |
| 브랜치 | `fix/release-pipeline-resilience` |

---

## 1. Release 파이프라인 현황 (As-Is)

### 1.1 End-to-end 체인

```
사용자: git push main
   │
   ▼
[auto-tag job]
   - git tag 추출 (latest stable + latest rc)
   - branch 별 next version 계산:
     main   → v{major}.{minor}.{patch+1}  (stable)
     staging → v{major}.{minor}.{patch+1}-rc1 (prerelease)
   - gh api로 tag reference 생성 (Verified 뱃지 획득)
   - outputs.version = 새 태그 이름
   │
   ▼
[goreleaser job]
   steps:
     - Checkout at tag
     - Setup Go / AWS OIDC / GHCR login / QEMU / Buildx
     - Run GoReleaser (5 minute)
       ├ builds (darwin/linux/windows × amd64/arm64)
       ├ archives (tar.gz / zip)
       ├ checksums
       ├ homebrew formula generation
       ├ dockers (GHCR push)
       ├ publishing.blobs (S3 upload)   ← 바이너리 전파
       ├ publishing.dockers (manifests)
       ├ docker_digests
       ├ scm.releases (GitHub Release)  ← 공개 시점
       └ publishing.homebrew (태그 리포 push)  ← 401 실패 지점
     - Update LATEST_VERSION (stable only)   ← 현재 이전 스텝 실패로 스킵됨
```

### 1.2 현재 `Update LATEST_VERSION` 스텝 (auto-tag.yml:150-158)

```yaml
- name: Update LATEST_VERSION (stable only)
  if: "!contains(needs.auto-tag.outputs.version, '-rc')"
  run: |
    VERSION="${{ needs.auto-tag.outputs.version }}"
    VERSION="${VERSION#v}"
    echo "$VERSION" > /tmp/LATEST_VERSION
    aws s3 cp /tmp/LATEST_VERSION s3://${{ env.RELEASE_BUCKET }}/LATEST_VERSION \
      --content-type "text/plain" \
      --cache-control "max-age=60"
```

**취약점**: `if:` 조건에 `always()` 가드가 없어서 GitHub Actions 기본 동작(이전 스텝 실패 시 subsequent step skip)에 노출. goreleaser가 `exit 1` 하면 이 스텝은 `skipped` 상태로 넘어감.

### 1.3 현재 `brews:` 섹션 (.goreleaser.yml:91-120)

```yaml
brews:
  - name: tene
    ...
    repository:
      owner: agent-kay-it
      name: homebrew-tene                      # ← 존재하지 않는 리포
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"  # ← secret 미설정
    skip_upload: auto                           # ← stable 릴리스에서는 skip 안 함
    install: |
      bin.install "tene"
      bash_completion.install ...
    ...
```

**취약점**: `homebrew-tene` 리포가 존재하지 않고 `HOMEBREW_TAP_GITHUB_TOKEN` secret도 없어서 매 stable 릴리스마다 401.

## 2. 설계 (To-Be)

### 2.1 Change #1 — `auto-tag.yml` 의 LATEST_VERSION 스텝 강화

**Before**:
```yaml
- name: Update LATEST_VERSION (stable only)
  if: "!contains(needs.auto-tag.outputs.version, '-rc')"
  run: |
    VERSION="${{ needs.auto-tag.outputs.version }}"
    VERSION="${VERSION#v}"
    echo "$VERSION" > /tmp/LATEST_VERSION
    aws s3 cp /tmp/LATEST_VERSION s3://${{ env.RELEASE_BUCKET }}/LATEST_VERSION \
      --content-type "text/plain" \
      --cache-control "max-age=60"
```

**After**:
```yaml
- name: Update LATEST_VERSION (stable only)
  # Runs even when GoReleaser exited non-zero on a non-critical step
  # (e.g. Homebrew formula push hitting 401 when the tap repo is absent).
  # Safeguard: refuse to flip the pointer unless the darwin/arm64 tarball
  # for this version actually exists on S3 — that single artifact proves
  # the `blobs:` publishing stage completed and the install.sh chain is
  # serviceable.
  if: always() && !contains(needs.auto-tag.outputs.version, '-rc')
  run: |
    VERSION="${{ needs.auto-tag.outputs.version }}"
    VERSION="${VERSION#v}"
    TARBALL="s3://${{ env.RELEASE_BUCKET }}/v${VERSION}/tene_${VERSION}_darwin_arm64.tar.gz"
    if ! aws s3 ls "$TARBALL" > /dev/null 2>&1; then
      echo "::error::Refusing to update LATEST_VERSION — $TARBALL not found on S3"
      exit 1
    fi
    echo "$VERSION" > /tmp/LATEST_VERSION
    aws s3 cp /tmp/LATEST_VERSION s3://${{ env.RELEASE_BUCKET }}/LATEST_VERSION \
      --content-type "text/plain" \
      --cache-control "max-age=60"
    echo "::notice::LATEST_VERSION updated to ${VERSION}"
```

**변경 포인트**:
1. `if:` 조건에 `always()` 추가 → 이전 스텝 실패에도 실행
2. S3 tarball 존재 검증 추가 → 바이너리 없이 포인터 전진 방지
3. `::error::` / `::notice::` GitHub annotation → workflow UI에서 명확한 결과 표시

**엣지 케이스 커버**:
| 시나리오 | 이전 동작 | 새 동작 |
|---------|---------|---------|
| 바이너리 upload 성공 → homebrew 단계만 실패 | skip → LATEST_VERSION 구버전 유지 | 실행 → 가드 통과 → LATEST_VERSION 신버전 |
| 빌드 단계에서 실패 (바이너리 없음) | skip | 실행 → 가드 실패 → LATEST_VERSION 구버전 보존 (의도됨) |
| rc tag | skip (변경 없음) | skip (변경 없음) |
| 정상 (전부 성공) | 실행 | 실행 |

### 2.2 Change #2 — `.goreleaser.yml` 의 `brews:` 섹션 비활성

**Before**: 원본 코드 (위 1.3 참조)

**After** (전체 주석 처리 + 복원 가이드):
```yaml
# Homebrew tap — currently disabled.
#
# Publishing to a Homebrew tap requires:
#   1. A public GitHub repo `agent-kay-it/homebrew-tene` to exist.
#   2. A fine-grained PAT with `Contents: Read/Write` on that repo,
#      stored as the `HOMEBREW_TAP_GITHUB_TOKEN` secret on this repo
#      (agent-kay-it/tene).
#
# Neither is set up yet, and three consecutive stable releases
# (v1.0.5, v1.0.6, v1.0.7) failed on this block with a 401 that
# skipped the downstream `Update LATEST_VERSION` step and left
# install.sh pointing at a stale version.
#
# To re-enable:
#   1. Create https://github.com/agent-kay-it/homebrew-tene (public, MIT).
#   2. Create the PAT and run:
#        gh secret set HOMEBREW_TAP_GITHUB_TOKEN --repo agent-kay-it/tene
#   3. Uncomment the block below.
#
# brews:
#   - name: tene
#     homepage: "https://tene.sh"
#     description: "Local-first encrypted secret manager CLI for AI-safe developer workflows"
#     license: "MIT"
#     repository:
#       owner: agent-kay-it
#       name: homebrew-tene
#       token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
#     directory: Formula
#     commit_author:
#       name: goreleaserbot
#       email: bot@tene.sh
#     commit_msg_template: "chore(tap): bump tene to {{ .Tag }}"
#     skip_upload: false
#     install: |
#       bin.install "tene"
#       bash_completion.install "completions/tene.bash" => "tene"
#       zsh_completion.install "completions/_tene"
#       fish_completion.install "completions/tene.fish"
#       man1.install "manpages/tene.1"
#     test: |
#       system "#{bin}/tene", "version"
#     caveats: |
#       Get started:
#         tene init
#
#       Documentation:
#         https://github.com/agent-kay-it/tene#readme
#
#       For AI agents using this project:
#         https://tene.sh/llms.txt
```

**변경 포인트**:
1. `brews:` 블록 전체 주석 처리
2. 이유 + 복원 가이드 주석으로 명시
3. `skip_upload: auto` → `skip_upload: false` 로 복원 시점에 명확한 동작 (혹시 모를 오해 방지)

**goreleaser 동작 변화**:
- `release --clean` 실행 시 brews 관련 phase 자체가 스킵 (config에 없음)
- `scm releases` (GitHub Release) 이후 바로 완료
- `homebrew formula` 단계의 `check default branch` 401이 발생할 일 자체가 없음

### 2.3 Change #3 — `CHANGELOG.md` 의 Unreleased Fixed 엔트리

```markdown
### Fixed

- `auto-tag.yml` workflow's `Update LATEST_VERSION` step now runs with
  `if: always()` and a S3 tarball existence check. Previously a
  non-critical GoReleaser failure (e.g. Homebrew formula publish) would
  skip this step and leave `install.sh` users on a stale version —
  v1.0.5, v1.0.6, and v1.0.7 each required a manual S3 hotfix.
- Homebrew publishing disabled in `.goreleaser.yml` until the
  `agent-kay-it/homebrew-tene` tap repository and `HOMEBREW_TAP_GITHUB_TOKEN`
  secret are set up. Re-enable instructions are preserved inline in the
  `brews:` comment block.
```

## 3. 실행 순서 (Do Phase)

1. `.github/workflows/auto-tag.yml` 의 `Update LATEST_VERSION` step 수정 (Change #1)
2. `.goreleaser.yml` 의 `brews:` 블록 주석 처리 (Change #2)
3. `CHANGELOG.md` Fixed 엔트리 추가 (Change #3)
4. Single commit: `fix(release): always update LATEST_VERSION + disable Homebrew publish`
5. Push → PR (base: staging)

## 4. 비기능 관점 (NFR)

### 4.1 보안
- S3 tarball 존재 검증은 **OIDC 인증된 GitHub Actions runner**에서만 실행. 외부 접근 불가.
- 가드 실패 시 step 자체가 exit 1 → workflow UI에서 명확한 에러 표시.

### 4.2 관찰성
- `::notice::LATEST_VERSION updated to ${VERSION}` annotation → workflow Summary에 성공 로그
- `::error::Refusing to update...` annotation → 가드 실패 시 명확한 원인 표시

### 4.3 복원 가능성
- `.goreleaser.yml` 주석 블록에 활성화 단계 포함 → 6개월 뒤 사용자가 봐도 즉시 복원 가능
- git log의 이 commit을 reference하면 의사결정 추적 가능

## 5. QA 매트릭스 (Check Phase)

### 5.1 정적 분석

| 검증 | 명령 | 기대 |
|---|---|---|
| YAML 구문 | `yq '.' .github/workflows/auto-tag.yml > /dev/null` | exit 0 |
| YAML 구문 | `yq '.' .goreleaser.yml > /dev/null` | exit 0 |
| goreleaser config | `cd /tmp && git clone ... && goreleaser check` | "config is valid" |
| brews 블록 부재 | `yq '.brews' .goreleaser.yml` | null |

### 5.2 시나리오 매트릭스

| 시나리오 | goreleaser 결과 | version | if 평가 | tarball 가드 | 최종 동작 |
|---------|:-:|:-:|:-:|:-:|---|
| stable, 정상 전체 | success | v1.0.8 | ✓ true | ✓ pass | LATEST_VERSION=1.0.8 |
| stable, 바이너리 upload 후 non-critical 실패 | failure | v1.0.8 | ✓ true | ✓ pass | LATEST_VERSION=1.0.8 |
| stable, 바이너리 upload 전 실패 | failure | v1.0.8 | ✓ true | ✗ fail | LATEST_VERSION 불변 + step exit 1 |
| rc | any | v1.0.8-rc1 | ✗ false | N/A | step skip |

### 5.3 운영 모니터링 (이번 사이클 외부)

다음 stable 릴리스 시 확인:
- Auto Tag & Release 전체 conclusion = success (brews 없어짐)
- `Update LATEST_VERSION` step의 annotation 로그 확인
- S3 `LATEST_VERSION` 파일 Last-Modified가 릴리스 시각과 일치

## 6. 롤백 (Act Phase)

### 6.1 이 fix 자체 문제 발생 시
1. revert this commit
2. push + PR
3. 되돌린 상태에서 수동 S3 hotfix로 운영 계속

### 6.2 Homebrew 재활성 (사용자 결정 시)
이 fix와 관계없이 독립적으로:
1. `gh repo create agent-kay-it/homebrew-tene --public --license MIT --add-readme`
2. Fine-grained PAT 발급 + `gh secret set HOMEBREW_TAP_GITHUB_TOKEN`
3. `.goreleaser.yml` brews 블록 주석 제거
4. 다음 stable 릴리스에서 자동 publish 확인

---

## 7. 참조

- Plan: `docs/01-plan/release-pipeline-resilience.plan.md`
- auto-tag.yml 현재: `tene/.github/workflows/auto-tag.yml:150-158`
- goreleaser brews 현재: `tene/.goreleaser.yml:91-120`
- install.sh 현재: `tene/apps/web/public/install.sh:39-47`
- 과거 실패: GitHub Actions 24826937468, 24888395413, 24889443661
