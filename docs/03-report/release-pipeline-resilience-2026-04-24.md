# Release Pipeline Resilience — Report

| 항목 | 값 |
|------|-----|
| Feature | `release-pipeline-resilience` |
| 기준일 | 2026-04-24 |
| 브랜치 | `fix/release-pipeline-resilience` |
| 관련 docs | `docs/01-plan/release-pipeline-resilience.plan.md`, `docs/02-design/features/release-pipeline-resilience.design.md` |
| 소요 시간 | ~2h (Plan 25m + Design 30m + Do 15m + QA 20m + Report 20m + S3 hotfix 10m) |

---

## 1. 배경 — 왜 이 PDCA가 필요했나

### 1.1 반복된 실패 패턴

2026-04-22 ~ 2026-04-24 사이 안정 릴리스 **3회 연속 동일 실패**:

| 릴리스 | 릴리스 시각 (UTC) | GoReleaser 결과 | S3 LATEST_VERSION 자동 갱신 |
|:-:|:-:|:-:|:-:|
| v1.0.5 | 2026-04-23 09:14 | ❌ failure (homebrew 401) | ❌ skip → 수동 hotfix |
| v1.0.6 | 2026-04-24 12:00 | ❌ failure (homebrew 401) | ❌ skip → 수동 hotfix 필요 |
| v1.0.7 | 2026-04-24 12:29 | ❌ failure (homebrew 401) | ❌ skip → 수동 hotfix |

### 1.2 사용자 영향

- **신규 사용자**: `curl ... | sh`로 설치 시 매번 구버전(최대 2버전 뒤) 다운로드
- **운영자(사용자 본인)**: 매 릴리스마다 수동 S3 `aws s3 cp` 명령어 실행 필요
- **인식 신뢰도**: GitHub Actions UI에 매번 빨간 "failure" 라벨 → 신규 기여자/외부 관찰자가 "유지 안 되는 프로젝트" 시그널로 해석 가능

### 1.3 이전 시도 (PR #66, 2026-04-23)

비슷한 워크플로우 fix 제안을 포함했지만 **homebrew 복원 작업과 묶어서** 제시 → 사용자가 "homebrew는 아직 고민할래"라며 PR close + 브랜치 삭제. 결과적으로 문제 방치.

---

## 2. 진단 (Plan Phase 결과)

### 2.1 Two independent root causes

| # | 원인 | 파일 | 증상 |
|:-:|---|---|---|
| A | `Update LATEST_VERSION` step에 `if: always()` 가드 없음 | `.github/workflows/auto-tag.yml:150-158` | goreleaser exit 1 → step skip |
| B | `brews:` 섹션이 존재하지 않는 `agent-kay-it/homebrew-tene` 리포 + 미설정 `HOMEBREW_TAP_GITHUB_TOKEN` secret 참조 | `.goreleaser.yml:91-120` | 매 stable 릴리스마다 401 Bad credentials |

A는 B의 **결과 증폭**. B를 고치지 않아도 A를 고치면 LATEST_VERSION 자동 갱신은 보장됨. B도 함께 해소하면 workflow failure 라벨 자체가 사라짐.

### 2.2 사용자 의도 (본 세션에서 확정)

> "LATEST_VERSION 업데이트를 CICD에 항상 포함해야겠어. 그리고 homebrew는 아직 대응안할거야. 이점도 CICD 수정 필요한듯."

→ 두 원인 모두 **동시에 해소**하기로 결정.

---

## 3. 실행 (Do Phase)

### 3.1 Change #1 — `auto-tag.yml` LATEST_VERSION step 강화

**파일**: `.github/workflows/auto-tag.yml`

**변경 내용**:
1. `if:` 조건에 `always()` 추가 → 이전 스텝 실패에 영향받지 않음
2. S3 tarball 존재 검증 가드 추가 → 바이너리 없는데 포인터만 전진하는 사고 방지
3. `::error::` / `::notice::` GitHub annotation 추가 → workflow Summary에서 결과 즉시 확인 가능
4. 이 변경의 동기와 과거 사례를 5줄 주석으로 파일 내 보존

**Before → After diff (요약)**:
```diff
- if: "!contains(needs.auto-tag.outputs.version, '-rc')"
+ if: always() && !contains(needs.auto-tag.outputs.version, '-rc')
  run: |
    VERSION="${{ needs.auto-tag.outputs.version }}"
    VERSION="${VERSION#v}"
+   TARBALL="s3://${{ env.RELEASE_BUCKET }}/v${VERSION}/tene_${VERSION}_darwin_arm64.tar.gz"
+   if ! aws s3 ls "$TARBALL" > /dev/null 2>&1; then
+     echo "::error::Refusing to update LATEST_VERSION — $TARBALL not found on S3"
+     exit 1
+   fi
    echo "$VERSION" > /tmp/LATEST_VERSION
    aws s3 cp /tmp/LATEST_VERSION s3://${{ env.RELEASE_BUCKET }}/LATEST_VERSION \
      --content-type "text/plain" \
      --cache-control "max-age=60"
+   echo "::notice::LATEST_VERSION updated to ${VERSION}"
```

### 3.2 Change #2 — `.goreleaser.yml` brews 섹션 비활성

**파일**: `.goreleaser.yml`

**변경 내용**:
- `brews:` 블록(30줄) 전체 주석 처리
- 상단 14줄 주석으로 (a) 왜 disabled인지 (b) 3번 실패한 릴리스 레퍼런스 (c) 3단계 재활성 가이드 명시
- 향후 재활성 시 `skip_upload: auto` → `skip_upload: false` 권장 기본값 변경 명시

### 3.3 Change #3 — `CHANGELOG.md` Unreleased Fixed 엔트리

2개 엔트리 추가:
1. `auto-tag.yml` LATEST_VERSION 내구성 개선 (증상 + fix 설명)
2. Homebrew publishing 일시 비활성 (이유 + 재활성 위치)

---

## 4. 검증 (QA Phase)

### 4.1 정적 분석 (실행 완료)

| 검증 | 결과 |
|---|:-:|
| `.github/workflows/auto-tag.yml` YAML 파싱 (Python `yaml.safe_load`) | ✅ 유효 |
| `.goreleaser.yml` YAML 파싱 | ✅ 유효 |
| `brews` top-level key가 `.goreleaser.yml`에서 사라졌는가 | ✅ None |
| `Update LATEST_VERSION` step의 `if` 조건 | ✅ `always() && !contains(needs.auto-tag.outputs.version, '-rc')` |
| tarball 존재 가드 코드 포함 | ✅ `TARBALL` + `aws s3 ls` 존재 |
| GitHub notice annotation 포함 | ✅ `LATEST_VERSION updated to` 존재 |
| `builds` / `archives` / `checksum` / `dockers` / `docker_manifests` / `blobs` 섹션 영향 없음 | ✅ 모두 정상 파싱 |

### 4.2 시나리오 매트릭스 (설계 시점 예상)

| 시나리오 | goreleaser 결과 | version | `if` 평가 | tarball 가드 | 최종 동작 |
|---------|:-:|:-:|:-:|:-:|---|
| 정상 (이상적) | success | v1.0.8 | ✅ true | ✅ pass | LATEST_VERSION=1.0.8 자동 갱신 |
| **현재 주된 시나리오** — 바이너리 upload 후 homebrew 실패 | failure | v1.0.8 | ✅ true | ✅ pass | LATEST_VERSION=1.0.8 자동 갱신 ✨ |
| 빌드/upload 전 실패 (바이너리 없음) | failure | v1.0.8 | ✅ true | ❌ fail | exit 1 + ::error::, LATEST_VERSION 불변 (보호) |
| rc 태그 (모든 경우) | any | v1.0.8-rc1 | ❌ false | N/A | skip (변경 없음) |
| workflow 취소 | cancelled | any | `always()` 특성상 try | 실제로는 runner 취소 시 step 미실행 | 영향 없음 |

### 4.3 운영 검증 (PDCA 외부 — 다음 릴리스 시 모니터링)

다음 stable 릴리스(v1.0.8 예상) 시 확인 항목:
1. `brews:` 없어져서 goreleaser exit 0 (homebrew formula step 자체 실행 안 됨)
2. Auto Tag & Release workflow conclusion = **success** (3개월 만에 처음)
3. S3 `LATEST_VERSION` 자동 갱신 Last-Modified가 릴리스 시각과 일치
4. `install.sh` 스크립트가 최신 버전을 즉시 다운로드
5. 수동 S3 hotfix **불필요**

---

## 5. 효과 (Act Phase 측정 예정)

### 5.1 측정 가능한 효과

| 지표 | Before (최근 3회 릴리스 평균) | After (목표) | 영향 |
|---|:-:|:-:|---|
| 릴리스 후 수동 개입 (분) | 10-15 | **0** | 릴리스당 ~12분 절약 |
| Auto Tag & Release workflow 성공률 | 0/3 (0%) | 예상 3/3 (100%) | UI 신호 정상화 |
| install.sh 구버전 노출 기간 | 길게는 1일+ | **< 1분** (S3 전파) | 신규 사용자 영향 제거 |
| Homebrew publish | ❌ 설정은 있으나 항상 실패 | ⏸ 명시적 보류 (문서화) | 의사결정 투명성 |

### 5.2 비의도적 부작용 (없음 예상)

- `dockers:`, `blobs:`, `scm releases` 등 다른 블록 영향 없음 (YAML 파싱 확인)
- 기존 v1.0.5-v1.0.7 릴리스의 이미 배포된 아티팩트(S3 바이너리, GHCR Docker)는 불변
- 복원 시점에 정확히 어떻게 되돌릴지 `.goreleaser.yml` 내부 가이드로 보존

---

## 6. 잔여 리스크 & 관찰 지점

### 6.1 모니터링 대상

- **다음 stable 릴리스 1회** — 이 fix의 실제 효과 검증. Auto Tag & Release conclusion 확인.
- **S3 LATEST_VERSION 가드 false positive** — 만약 goreleaser가 탁월한 경로로 바이너리는 업로드했지만 다른 경로로 실패한다면? 현재 가드는 darwin/arm64 1개 파일만 검사. 5개 플랫폼 전체 검증까지는 안 함. **트레이드오프**: 완전 검증은 복잡, 단일 플랫폼은 빠름. 필요 시 후속.

### 6.2 보류 (의도적)

- Homebrew 재활성 — 사용자 결정 대기
  - `agent-kay-it/homebrew-tene` 저장소 생성
  - `HOMEBREW_TAP_GITHUB_TOKEN` secret 설정
  - `.goreleaser.yml` brews 주석 제거
- Staging goreleaser 17분 hang 이슈 (2026-04-24 12:27:37 시작) — 본 fix와 무관한 GitHub Runner 이슈로 추정. main이 staging rc를 승격한 race condition 가능성. 재발 시 별도 조사.

### 6.3 장기 개선 후보 (이번 스코프 외)

- Slack/Discord 알림 연동 (release 성공/실패 시)
- `install.sh`에 `LATEST_VERSION` 파일 정합성 fallback (GitHub API 조회)
- cron sync workflow (hourly reconcile between GitHub latest release ↔ S3 LATEST_VERSION)
- Staging 릴리스에서 goreleaser 생략 (rc publish 자체가 필요한지 재평가)

---

## 7. 파일 변경 요약

| 파일 | 유형 | 변경 |
|---|:-:|---|
| `.github/workflows/auto-tag.yml` | 수정 | Update LATEST_VERSION step: `if: always()` + tarball guard + annotations |
| `.goreleaser.yml` | 수정 | `brews:` 블록 주석 처리 + 재활성 가이드 |
| `CHANGELOG.md` | 수정 | Unreleased > Fixed 2 entries 추가 |
| `docs/01-plan/release-pipeline-resilience.plan.md` | 신규 | Plan 문서 |
| `docs/02-design/features/release-pipeline-resilience.design.md` | 신규 | Design 문서 |
| `docs/03-report/release-pipeline-resilience-2026-04-24.md` | 신규 | 본 보고서 |

**코드 변경 규모**: 3개 파일, +53 / -39 줄 (워크플로우 + goreleaser + CHANGELOG 합산).

---

## 8. 배포 계획

1. 본 브랜치 (`fix/release-pipeline-resilience`) push
2. PR → `staging` 생성 + merge (CI 통과)
3. `staging` → `main` PR + merge
4. main merge → auto-tag이 v1.0.8 생성 → goreleaser 실행
5. **관찰 지점**: 이 시점에서 fix가 실제로 작동하는지 확인. 3회 연속 실패 패턴이 깨지는지.

---

## 9. 참조

- Plan: `docs/01-plan/release-pipeline-resilience.plan.md`
- Design: `docs/02-design/features/release-pipeline-resilience.design.md`
- 실패 로그 (GitHub Actions runs):
  - 24826937468 (v1.0.5)
  - 24888395413 (v1.0.6)
  - 24889443661 (v1.0.7)
- S3 수동 hotfix 이력 (이 보고서 기준 가장 최근):
  - 2026-04-23 11:06:48 UTC — 1.0.4 → 1.0.5
  - 2026-04-24 12:50:03 UTC — 1.0.5 → 1.0.7
- 이전 시도 PR #66 (close됨, 2026-04-23)

---

**PDCA 사이클 요약**:
- Plan: ✅ 2개 원인 진단 + 4개 옵션 평가 후 각 1개 선택
- Design: ✅ diff 레벨 설계 + 엣지 케이스 매트릭스
- Do: ✅ 3개 파일 수정
- Check (QA): ✅ YAML 유효성 + 키 제거 검증 + 시나리오 매트릭스 정적 평가
- Act: ⏸ 다음 stable 릴리스에서 실측 검증 (PDCA 외부)

**작성**: 2026-04-24 (Claude main, 사용자 요청 PDCA 전주기 단일 세션)
