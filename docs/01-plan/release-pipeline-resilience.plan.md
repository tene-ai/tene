# Release Pipeline Resilience — Plan

| 항목 | 값 |
|------|-----|
| Feature | `release-pipeline-resilience` |
| 기반 컨텍스트 | v1.0.5, v1.0.6, v1.0.7 세 번 연속 동일 실패 |
| 브랜치 | `fix/release-pipeline-resilience` (origin/staging 베이스) |
| 관련 이전 시도 | PR #66 (비슷한 fix 제안, 사용자가 close/revert함) |
| 총 공수 | ~2-3 인시간 |

---

## 1. 문제 요약 (Why now)

지난 3번의 안정 릴리스(v1.0.5, v1.0.6, v1.0.7) 모두 **동일한 증상**으로 실패 처리됨:

1. GoReleaser가 바이너리 빌드 + S3 업로드 + Docker GHCR publish + GitHub Release 모두 **성공**
2. 마지막 `homebrew formula` 단계에서 `tomo-kay/homebrew-tene` 리포(존재하지 않음) 401
3. goreleaser 전체 exit 1
4. 다음 스텝 `Update LATEST_VERSION (stable only)`가 **GitHub Actions 기본 동작(이전 스텝 실패 시 skip)** 으로 스킵됨
5. 결과: `install.sh`가 참조하는 S3 `LATEST_VERSION` 파일이 **2번 연속 수동 핫픽스** 필요했음 (4/23, 4/24)

### 1.1 실측 증거

```bash
# main v1.0.7 Auto Tag & Release (2026-04-24 12:27) 로그 마지막:
• release published url=https://github.com/tomo-kay/tene/releases/tag/v1.0.7
• homebrew formula
• error checking for default branch projectID=tomo-kay/homebrew-tene statusCode=401
⨯ release failed after 1m47s
error=homebrew formula: could not get default branch: 401 Bad credentials

# S3 LATEST_VERSION 당시 상태:
curl https://tene-releases.s3.ap-northeast-2.amazonaws.com/LATEST_VERSION
# → 1.0.5  (수동 hotfix 그대로, 1.0.6 / 1.0.7 배포됐음에도)
```

## 2. 사용자 확정 방향

사용자 의도 (2026-04-24 세션):

1. **"LATEST_VERSION 업데이트를 CICD에 항상 포함해야겠어"** — 어떤 스텝 실패에도 install.sh 포인터는 갱신되어야
2. **"homebrew는 아직 대응안할거야"** — 당분간 비활성화. 나중에 `tomo-kay/homebrew-tene` 저장소 + PAT 준비 시 재활성화
3. 두 조치 모두 CICD 코드 변경 필요

## 3. 스코프

### 3.1 변경 대상 2개 파일

| 파일 | 변경 이유 | 변경 요약 |
|------|----------|----------|
| `.github/workflows/auto-tag.yml` | LATEST_VERSION 항상 실행 | `if: always()` 가드 + S3 tarball 존재 검증 |
| `.goreleaser.yml` | Homebrew publish 비활성 | `brews:` 섹션 주석 처리 + 복원 가이드 |

### 3.2 변경 비대상 (의도적 제외)

- `install.sh` — 변경 없음 (S3 `LATEST_VERSION` 파일 갱신만 신경 쓰면 됨)
- `dockers:` 섹션 — 정상 작동 중, 변경 없음
- `blobs:` 섹션 — 정상 작동 중, 변경 없음
- `tomo-kay/homebrew-tene` 리포 — 생성 보류 (사용자 결정)
- `HOMEBREW_TAP_GITHUB_TOKEN` secret — 설정 보류

### 3.3 문서 업데이트

- `docs/01-plan/release-pipeline-resilience.plan.md` (이 문서)
- `docs/02-design/features/release-pipeline-resilience.design.md` (설계 상세)
- `CHANGELOG.md` — Unreleased 섹션의 Fixed 엔트리
- `docs/03-report/release-pipeline-resilience-YYYY-MM-DD.md` (완료 후)

## 4. 의사결정 요약

### 4.1 LATEST_VERSION 갱신 보장 방식 (4가지 검토 → 1 택)

| 옵션 | 내용 | 평가 |
|------|-----|------|
| A. `if: always()` + tarball 존재 가드 | goreleaser가 어느 단계에서 실패하든 LATEST_VERSION 시도. 단, 실제 바이너리가 S3에 없으면 update 거부. | ✅ **선택** — 기회적으로 항상 시도하되 잘못된 포인터 방지 |
| B. goreleaser job 성공/실패와 무관하게 별도 job으로 분리 | 전체 release job 완료 후 `needs:` + `if:` 로 분기 | 과한 구조 변경 |
| C. post-install 체크 스크립트 | install.sh가 이상 감지 시 GitHub API fallback | install.sh 변경 필요 + 사용자 환경에 curl 의존 |
| D. Cron job으로 매시간 sync | GitHub latest release ↔ S3 LATEST_VERSION 동기화 | 새 cron workflow 추가 + 지연 시간 발생 |

### 4.2 Homebrew 비활성 방식 (3가지 검토 → 1 택)

| 옵션 | 내용 | 평가 |
|------|-----|------|
| A. `brews:` 섹션 전체 주석 처리 | goreleaser 실행 시 brews 단계 자체가 없음 | ✅ **선택** — 실수 여지 최소, 복원 시 주석 제거 1줄 |
| B. `skip_upload: true` | Formula 파일 로컬 생성하되 push 안 함 | 실험 결과 `check default branch` API 호출은 skip_upload 전에 발생 → 여전히 401 |
| C. `brews:` 블록 삭제 | 히스토리 상실, git log에서 복원 의도 추적 어려움 | 비대체 권장 |

## 5. 검증 계획 (QA Phase)

실제 v1.0.8 태그를 잘라보는 건 너무 파괴적. 대신 **정적 분석 + 과거 실패 시나리오 재현**:

### 5.1 YAML 문법 검증
```bash
yq '.' .github/workflows/auto-tag.yml > /dev/null  # 구문 에러 없음
```

### 5.2 의존성 그래프 분석
- `Update LATEST_VERSION` step의 `if:` 조건이 모든 주요 시나리오에서 올바르게 평가되는지 매트릭스 검증

| 시나리오 | goreleaser 결과 | auto-tag.outputs.version | if 평가 | 기대 동작 |
|---------|----------|--------|--------|----------|
| main stable, 바이너리 S3 업로드 성공 후 homebrew 실패 | failure | `v1.0.8` | `always() && !contains(v1.0.8, '-rc')` = true | 실행 → S3 ls 통과 → LATEST_VERSION=1.0.8 |
| main stable, 바이너리 업로드 전 실패 | failure | `v1.0.8` | true | 실행 → S3 ls 실패 → exit 1 (포인터 보호) |
| staging rc tag | any | `v1.0.8-rc1` | false | skip |
| 전부 성공 (가상 시나리오) | success | `v1.0.8` | true | 실행 → LATEST_VERSION 갱신 |

### 5.3 goreleaser dry-run
```bash
# goreleaser check로 config 유효성 확인
cd tene && goreleaser check  # brews 주석 처리 후에도 config 유효
```

### 5.4 실제 운영 검증 (릴리스 후)

현재 PDCA 사이클에서 실행 가능한 건 3번째까지. 실제 동작은 **다음 stable 릴리스(v1.0.8 또는 v1.0.6 재배포)** 시점까지 확인 보류.

## 6. 타임라인

| 단계 | 예상 소요 | 산출물 |
|------|:--:|------|
| Plan (본 문서) | 30m | `docs/01-plan/release-pipeline-resilience.plan.md` |
| Design | 30m | `docs/02-design/features/release-pipeline-resilience.design.md` |
| Do | 20m | 2개 파일 수정 + CHANGELOG + commit |
| QA | 30m | YAML 검증 + 시나리오 매트릭스 |
| Report | 20m | `docs/03-report/release-pipeline-resilience-2026-04-24.md` |
| PR + merge (staging → main) | 20m | 브랜치 push + PR #XX |
| **Total** | **~2.5h** | |

## 7. 리스크 & 완화

| 리스크 | 확률 | 영향 | 완화 |
|---|:-:|:-:|------|
| `if: always()` 추가로 인해 goreleaser 실패해도 LATEST_VERSION이 갱신되어 깨진 버전 공개 | 낮 | 고 | S3 tarball 존재 검증 가드로 차단 |
| brews: 주석 처리 후 다른 섹션(dockers/blobs) 영향 | 낮 | 저 | goreleaser config는 독립 블록. QA에서 goreleaser check로 검증 |
| 다음 릴리스에서 또 다른 예외 케이스 노출 | 중 | 중 | Report에 모니터링 지점 명시. 재발 시 재플랜 |
| 사용자가 언젠가 homebrew 켤 때 복원 과정에서 혼란 | 낮 | 저 | `.goreleaser.yml` 주석 블록에 복원 가이드 명시 |

## 8. 성공 기준

- 다음 stable 릴리스 1회 경과 후:
  - goreleaser exit 0 (brews 단계 없어짐)
  - Auto Tag & Release workflow 상태 = success
  - S3 `LATEST_VERSION` 파일 자동 갱신됨 (수동 hotfix 불필요)
  - `install.sh`로 설치 시 최신 버전 도달
- Homebrew publish는 여전히 안 됨 (의도적)

---

## 9. 참조

- 과거 실패 로그: GitHub Actions run 24826937468 (v1.0.5), 24888395413 (v1.0.6), 24889443661 (v1.0.7)
- 수동 hotfix 이력:
  - 2026-04-23 11:06:48 UTC — S3 LATEST_VERSION 1.0.4 → 1.0.5
  - 2026-04-24 12:50:03 UTC — S3 LATEST_VERSION 1.0.5 → 1.0.7
- 이전 PR #66 (close됨) — 유사 제안 포함했으나 homebrew-tene 생성 경로와 함께 묶어서 제시하여 사용자가 거부
- 관련 파일: `.github/workflows/auto-tag.yml:150-158`, `.goreleaser.yml:91-120`, `apps/web/public/install.sh:39-47`

---

**작성**: 2026-04-24 (Claude main, 사용자 요청)
**승인 대기**: 사용자 확인 후 Design → Do 진입
