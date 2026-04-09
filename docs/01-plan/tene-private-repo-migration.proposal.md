# Tene Private Repo Migration 기획서 보완 제안서

> 기존 기획서: `docs/01-plan/tene-private-repo-migration.plan.md`
> 작성일: 2026-04-09
> 상태: Proposal

---

## 1. 핵심 제안: "전체 Private" 대신 "전략적 Open Core" 병행 검토

### 왜 전략을 재검토해야 하는가?

기존 기획서는 "전체 Private + S3 릴리스"를 전제로 작성되었다. 이 전략은 즉각적인 보안 리스크를 해소하지만, Infisical 사례 분석 결과 **더 전략적인 접근이 가능하다는 근거**가 확보되었다.

**Infisical의 성공 공식:**

| 단계 | 행동 | 결과 |
|------|------|------|
| 2022 | MIT 라이선스로 오픈소스 공개 | GitHub 바이럴, 초기 사용자 확보 |
| 2023 | 20K+ GitHub stars 달성 | 기업 고객 인바운드 시작 |
| 2024 | $16M Series A (GV, Y Combinator) | 20명 팀, Enterprise 고객 확보 |
| 현재 | MIT Core + `ee/` Enterprise 유료 | Free(5명) → Pro($18/user) → Enterprise(custom) |

이 공식의 핵심: **오픈소스 바이럴 → stars → 신뢰 → 기업 전환 → 매출**

기존 기획서의 "전체 Private" 전략은 이 성장 엔진을 완전히 포기하는 것이다. 그러나 현재 Tene의 상황(1인 개발, 매출 0, 커뮤니티 0)에서 코드 보호가 우선이라는 주장도 충분히 유효하다.

**이 제안서의 목적:** 세 가지 시나리오를 비교하여 의사결정 근거를 제공한다.

---

## 2. 시나리오 A: 전체 Private (기존 기획서)

기존 기획서(`tene-private-repo-migration.plan.md`)의 전략을 그대로 실행한다.

### 장점

- **즉각적 보안 해소**: AWS Account ID, VPC CIDR, IAM Role ARN 등 인프라 정보 즉시 비공개
- **구현 복잡도 최소**: 기존 기획서의 17개 파일 수정으로 완료 (3일)
- **1인 운영 부담 최소**: 커뮤니티 관리, issue 대응, 기여 리뷰 부담 없음
- **수익 보호**: 전체 코드베이스가 비공개이므로 클론/포크 불가
- **집중**: 제품 개발에만 집중 가능

### 단점

- **성장 엔진 없음**: GitHub 검색, stars 기반 유기적 유입 완전 차단
- **신뢰 구축 어려움**: 보안 제품인데 코드를 공개하지 않는다는 인식
- **Infisical 대비 포지셔닝 약화**: "우리가 더 안전하다"고 주장하지만 검증 불가
- **SEO 손실**: GitHub 백링크 소멸, 개발자 커뮤니티 노출 제로
- **기여자 확보 불가**: 1인 개발 상태 지속

### 적합 시기

**지금 즉시.** 보안 리스크가 실재하고 (AWS Account ID 노출), 커뮤니티가 0인 현재 private 전환의 기회비용이 가장 낮다.

---

## 3. 시나리오 B: 전략적 Open Core (Infisical 모델 적용)

Infisical의 구조를 Tene에 맞게 적용한다. 핵심 CLI 엔진은 MIT 오픈소스로 공개하고, 클라우드/인프라/과금은 비공개로 유지한다.

### 구조

```
tene (Public Repo - MIT License)
├── cmd/tene/              CLI entrypoint
├── internal/crypto/       XChaCha20-Poly1305, Argon2id, HKDF, X25519
├── internal/vault/        SQLite vault CRUD
├── internal/keychain/     OS Keychain integration
├── internal/recovery/     BIP-39 mnemonic
├── internal/claudemd/     CLAUDE.md generation
├── internal/cli/          Cobra commands (local-only: init, set, get, run, list, delete, import, export, env, passwd, recover)
├── internal/domain/       Domain models
├── internal/config/       CLI config
└── internal/errors/       CLI error codes

tene-cloud (Private Repo)
├── cmd/server/            Cloud API server
├── internal/api/          Echo server, handlers, middleware
├── internal/auth/         OAuth + JWT
├── internal/sync/         Push/Pull/Merge engine
├── internal/billing/      LemonSqueezy
├── internal/cli/          Cloud commands (login, logout, push, pull, sync, team, billing, whoami)
├── apps/dashboard/        Dashboard (Next.js)
├── apps/web/              Landing (Next.js)
├── infra/terraform/       AWS infrastructure
├── migrations/            PostgreSQL migrations
└── ee/                    Enterprise features (future)
```

### 장점

- **바이럴 성장**: GitHub stars → Hacker News → 개발자 커뮤니티 유기적 유입
- **신뢰 극대화**: Zero-Knowledge 암호화 구현을 누구나 감사 가능
- **Infisical 차별화**: "우리는 코어가 오픈소스이고 Zero-Knowledge" vs "Infisical은 서버가 평문 접근 가능"
- **기여자 확보**: 암호화 모듈, CLI 기능 등 외부 기여 가능
- **SEO + 백링크**: GitHub 프로필, awesome 리스트, 비교 사이트 등 노출

### 단점

- **구현 복잡도 높음**: monorepo 분리, import path 재설계, CI/CD 이중 관리
- **운영 부담**: issue 대응, PR 리뷰, 보안 취약점 보고 대응, 문서 유지
- **수익 보호 약화**: 로컬 기능만으로 충분한 사용자는 유료 전환 동기 없음
- **1인 한계**: 커뮨니티 관리 + 제품 개발 + 비즈니스 운영을 혼자 감당해야 함
- **분리 비용**: Go internal 패키지 특성상 분리 시 public API 설계 필요 (현재 `internal/`은 외부 import 불가)

### 적합 시기

**제품 출시 후 6개월+ 시점.** 최소한의 유료 사용자가 확보되어 커뮤니티 관리에 시간을 쓸 여유가 있을 때.

### Go 모듈 분리 시 기술적 고려사항

현재 Tene는 단일 모듈(`go.mod`)이다. Open Core 분리 시:

```
# Public repo
module github.com/tomo-kay/tene  (또는 go.tene.sh/tene)

# Private repo
module github.com/tomo-kay/tene-cloud
require github.com/tomo-kay/tene v0.x.x  // public 모듈 의존
```

- `internal/` 패키지는 외부 import 불가이므로, 공개할 패키지는 `pkg/` 또는 최상위로 이동 필요
- `internal/crypto/` → `crypto/` (public), `internal/vault/` → `vault/` (public) 등
- 대안: internal을 유지하되, CLI 바이너리만 public repo에서 빌드 (Go SDK 제공 안 함)

---

## 4. 시나리오 C: Private First, Open Later (하이브리드)

**지금은 전체 Private으로 보안을 해결하고, 나중에 전략적으로 일부를 공개한다.**

### 타임라인

| 시점 | 행동 | 목적 |
|------|------|------|
| **지금 (Week 0)** | 전체 Private 전환 (기존 기획서 실행) | 보안 리스크 즉시 해소 |
| **출시 (Month 1-2)** | 제품 공개 출시, 유료 사용자 확보 시작 | 매출 검증 |
| **Month 3** | `tene-crypto` 별도 public repo 생성 | 암호화 모듈만 MIT 공개, 보안 신뢰 확보 |
| **Month 4-5** | CLI core 기능 public repo 추출 | stars 기반 유기적 성장 시작 |
| **Month 6** | Open Core 모델 공식 선언 | Free(로컬) → Pro(클라우드) → Enterprise 구조 확립 |
| **Month 6+** | 커뮤니티 기반 성장 + Enterprise 영업 | Infisical 모델 본격 적용 |

### Stage 1: 전체 Private (지금, 기존 기획서 그대로)

기존 기획서의 Phase 1-5를 그대로 실행한다. 17개 파일 수정, 3일 소요.

### Stage 2: Crypto 모듈 공개 (Month 3)

```
tene-crypto (Public Repo - MIT)
├── argon2id.go      KDF (Argon2id 64MB, 3iter)
├── xchacha20.go     Encrypt/Decrypt (XChaCha20-Poly1305)
├── hkdf.go          Key derivation (HKDF-SHA256)
├── x25519.go        Key exchange (X25519 ECDH)
├── encfile.go       Encrypted file format
└── README.md        알고리즘 설명 + 벤치마크
```

**효과:**
- "Zero-Knowledge 맞냐?"는 질문에 "코드 보세요"로 답변 가능
- 암호화 전문가 리뷰 유도 → 보안 신뢰 향상
- GitHub stars 첫 시드 확보
- 메인 repo는 여전히 private (비즈니스 로직 보호)

### Stage 3: CLI Core 공개 (Month 4-5)

```
tene (Public Repo - MIT)
├── cmd/tene/         CLI entrypoint
├── crypto/           tene-crypto 모듈 inline 또는 import
├── vault/            SQLite vault CRUD
├── keychain/         OS Keychain
├── recovery/         BIP-39 mnemonic
├── claudemd/         CLAUDE.md generation
├── domain/           Domain models
└── cli/              Local commands only (init, set, get, run, list, delete, import, export, env, passwd, recover)
```

**Cloud 명령어(login, push, pull, sync, team, billing)는 빌드 태그로 분리:**
```go
//go:build cloud

package cli

func init() {
    rootCmd.AddCommand(loginCmd)
    rootCmd.AddCommand(pushCmd)
    // ...
}
```

- 공식 릴리스 바이너리는 cloud 태그 포함 빌드
- 오픈소스 빌드는 로컬 기능만 동작
- Infisical의 `ee/` 디렉토리 패턴과 유사하지만 Go 빌드 태그로 더 깔끔하게 분리

### 이 방식의 핵심 장점

1. **지금 보안 해결**: Private 전환으로 인프라 노출 즉시 차단
2. **나중에 성장 엔진 확보**: 검증된 시점에 Open Core 시작
3. **단계적 복잡도**: 한 번에 monorepo 분리하지 않고 점진적으로 추출
4. **실패 비용 최소**: Open Core가 효과 없으면 private 유지하면 됨
5. **Infisical 대비 포지셔닝**: "유일한 Zero-Knowledge + Open Core" → 두 마리 토끼

---

## 5. 기획서 보완 항목

기존 기획서(`tene-private-repo-migration.plan.md`)에 추가/수정이 필요한 항목들이다.

### 5-1. 추가해야 할 Phase

#### Phase 0: Git History 민감 정보 정리 (Phase 1 전에 실행)

현재 git history에 AWS Account ID(`507221376909`), VPC CIDR, IAM Role ARN 등이 남아 있다. Private 전환만으로는 이전 public 기간 동안 클론된 복사본의 리스크를 해소하지 못한다.

**필요 작업:**

```bash
# 1. 현재 노출된 민감 정보 확인
grep -r "507221376909" --include="*.tf" --include="*.yml" --include="*.go" .

# 2. 민감 정보 목록
#    - AWS Account ID: 507221376909
#    - ECR repo URL: 507221376909.dkr.ecr.ap-northeast-2.amazonaws.com/tene-api
#    - IAM Role ARN: arn:aws:iam::507221376909:role/tene-prod-github-actions
#    - RDS endpoint (있다면)
#    - VPC CIDR 대역 (10.0.0.0/16 등)

# 3. 대응 방안 (택 1)
#    a) git filter-repo로 history rewrite → force push (리스크: 기존 클론 무효화)
#    b) Private 전환 후 무시 (이미 클론된 복사본은 방어 불가, 실질 리스크 낮음)
#    c) AWS 리소스 재생성 (Account ID는 변경 불가, IAM Role은 교체 가능)
```

**권장:** 옵션 (b). Private 전환 자체로 추가 노출을 차단하고, AWS 보안 그룹/IAM 정책이 올바르면 Account ID 노출만으로는 실질적 공격이 어렵다. 단, IAM Role 신뢰 정책이 repo명 기반이므로 (`repo:tomo-kay/tene:*`) 외부인이 같은 이름의 public repo를 만들어 OIDC를 시도할 리스크는 점검해야 한다.

**기획서 반영 위치:** 섹션 5 "실행 계획" 앞에 Phase 0으로 추가.

#### Phase 6: Infisical 대비 차별화 강화 (Phase 5 이후)

Private 전환 완료 후, "오픈소스가 아닌 보안 제품"에 대한 신뢰를 확보하기 위한 별도 작업이 필요하다.

**필요 작업:**

1. **암호화 알고리즘 공개 문서 작성** (`tene.sh/security` 또는 `docs.tene.sh/security`)
   - XChaCha20-Poly1305, Argon2id, HKDF, X25519 구현 상세
   - 4-Layer Encryption 도식 (L1-L4)
   - Zero-Knowledge 증명 논리 (서버가 복호화할 수 없는 이유)
   - Infisical과의 아키텍처 비교표

2. **보안 백서(whitepaper) 작성**
   - Threat model 정의
   - 공격 시나리오별 방어 설명
   - 제3자 감사 계획 (향후)

3. **랜딩 페이지 보안 섹션 강화**
   - 기존: "오픈소스 — 코드를 확인하세요"
   - 변경: "Zero-Knowledge 암호화 — 서버가 물리적으로 복호화 불가"
   - 암호화 알고리즘별 아이콘/도식 추가

**기획서 반영 위치:** 섹션 5에 Phase 6으로 추가, 섹션 10 "향후 로드맵 > 단기"에도 연동.

### 5-2. 기존 Phase 보완

#### Phase 2 보완: CLI update 명령어 (`internal/cli/update.go`) 상세

기존 기획서 섹션 5 > Phase 2 > 2-4에서 `update.go`의 변경 사항을 다루고 있지만, 다음 항목이 누락되었다:

**누락 1: 체크섬 검증 구현 상세**

기존 기획서에서 "4) (선택) 체크섬 검증 추가"로 되어 있으나, S3 배포에서는 **필수**로 격상해야 한다. GitHub Releases는 GitHub의 HTTPS + 무결성 보장이 있지만, public S3 버킷은 별도 검증이 필요하다.

```go
// 추가 필요: downloadAndVerify 함수
func downloadAndVerify(binaryURL, checksumsURL, expectedFilename string) (string, error) {
    // 1. checksums.txt 다운로드
    // 2. 바이너리 다운로드
    // 3. SHA-256 비교
    // 4. 불일치 시 에러 반환
}
```

**누락 2: Fallback 로직**

기존 CLI 사용자(구버전)가 `tene update`를 실행하면 GitHub API 404를 받는다. 기존 기획서 섹션 8 리스크 #1에서 언급하지만, 구체적 구현이 없다.

```go
// 제안: fetchLatestRelease에 이중 시도 로직
func fetchLatestRelease() (*releaseInfo, error) {
    // 1차: S3 시도
    info, err := fetchFromS3()
    if err == nil {
        return info, nil
    }
    // 2차: GitHub API fallback (이전 버전 호환)
    info, err = fetchFromGitHub()
    if err != nil {
        return nil, fmt.Errorf("cannot check for updates (try: curl -sSfL https://tene.sh/install.sh | sh): %w", err)
    }
    return info, nil
}
```

**기획서 반영 위치:** 섹션 5 > Phase 2 > 2-4를 확장.

#### Phase 3 보완: GitHub 링크 제거 시 대체 URL 전략

기존 기획서 섹션 5 > Phase 3에서 9곳의 GitHub 링크를 "제거 또는 대체"로 명시하고 있으나, **대체 URL의 구체적 전략**이 없다.

**제안: 대체 URL 매핑표**

| 현재 GitHub 링크 용도 | 대체 URL | 비고 |
|----------------------|----------|------|
| 소스코드 열람 | `tene.sh/security` (보안 문서) | 코드 대신 암호화 설명 |
| Issue 제보 | `tene.sh/feedback` 또는 `feedback@tene.sh` | 이메일 또는 피드백 폼 |
| Star/Fork CTA | `tene.sh/install` (설치 가이드) | 행동 유도를 "설치"로 전환 |
| README (llms.txt) | `tene.sh` (메인 사이트) | AI 크롤러 대응 |
| JSON-LD installUrl | `tene.sh/install.sh` | 이미 기획서에 명시됨 |

**추가 필요 페이지:**
- `tene.sh/security` — 암호화 아키텍처 문서 (Phase 6과 연동)
- `tene.sh/feedback` — 피드백 채널 (이메일 리다이렉트 또는 간단한 폼)
- `tene.sh/install` — 설치 가이드 페이지 (install.sh curl 명령어 + 수동 다운로드 링크)

**기획서 반영 위치:** 섹션 5 > Phase 3 테이블에 "대체 URL" 칼럼 추가.

#### Phase 4 보완: Vercel GitHub App 권한 재설정 구체적 절차

기존 기획서 섹션 5 > Phase 4 > 4-1에서 "Vercel 프로젝트 설정 확인"이라고만 되어 있다. Private 전환 시 Vercel 연동이 끊기면 Landing + Dashboard 배포가 중단되므로, 구체적 절차가 필요하다.

**Vercel Private Repo 연동 확인 절차:**

```
1. Vercel Dashboard → Settings → Git Integration 접속
2. GitHub App 권한 확인:
   - "Repository access" → "Only select repositories" → tene 포함 확인
   - 또는 "All repositories" (이 경우 private 전환 시에도 자동 접근)

3. 만약 "Only select repositories"이고 tene이 private으로 전환되면:
   - GitHub Settings → Applications → Vercel → Configure
   - "Repository access"에서 tene repo가 여전히 선택되어 있는지 확인
   - private repo도 선택 가능 (GitHub App 권한이 이미 있으면)

4. 검증: Private 전환 후 빈 커밋 push
   git commit --allow-empty -m "test: verify Vercel deployment after private"
   git push origin main
   → Vercel 빌드 트리거 확인

5. 실패 시 대응:
   - Vercel Dashboard → Import Project → 다시 연결
   - 또는 Vercel CLI로 수동 배포: vercel --prod
```

**추가 확인 사항:**
- Vercel Hobby 플랜에서 private repo 지원 확인 (지원됨, 2024년 기준)
- 두 프로젝트(tene, tene-dashboard) 모두 확인 필요
- Preview 배포(PR)도 정상 동작하는지 확인

**기획서 반영 위치:** 섹션 5 > Phase 4 > 4-1을 확장.

### 5-3. 누락된 리스크

기존 기획서 섹션 8 "리스크 및 완화 방안"에 다음 항목을 추가해야 한다.

#### 리스크 8: go.mod module path 변경 필요성

**분석:**

```
현재 module path: (go.mod 확인 필요, 추정) github.com/tomo-kay/tene
```

Private 전환 시 `go.mod`의 module path 변경이 필요한가?

- **결론: 불필요.** Tene는 외부에서 `go get`으로 import하는 라이브러리가 아니라, 자체 빌드하는 CLI + 서버 바이너리이다.
- `go.mod`의 module path는 내부 import 경로로만 사용되며, GitHub 접근 가능 여부와 무관하게 빌드된다.
- CI/CD에서 `actions/checkout`으로 소스를 받은 후 로컬에서 `go build`하므로 module proxy 접근 불필요.
- 단, 향후 Open Core로 public 패키지를 제공할 때는 vanity import path(`go.tene.sh/tene`) 도입을 검토할 수 있다.

**기획서 반영:** 섹션 8에 리스크 #8로 추가 (확률: 해당 없음, 결론: 변경 불필요).

#### 리스크 9: GitHub OAuth App Callback URL

**분석:**

GitHub OAuth App의 callback URL은 `https://api.tene.sh/api/v1/auth/github/callback`이다. 이 URL은 GitHub App 설정에 등록된 것이며, **repo visibility와 무관**하다.

- OAuth App은 GitHub 계정 레벨에서 관리된다 (repo 아님).
- callback URL은 `api.tene.sh` 도메인이므로 repo private 전환과 관계없다.
- Local OAuth App(`localhost:8080`)도 동일하게 영향 없다.

**결론: 변경 불필요.**

**기획서 반영:** 섹션 4 "영향도 분석 > 영향 없음" 목록에 명시적으로 추가.

#### 리스크 10: SEO 영향 및 대안 채널

기존 기획서 리스크 #6에서 "SEO 하락"을 언급하지만, 구체적 영향과 대안이 부족하다.

**구체적 영향:**

| 항목 | 현재 | Private 전환 후 |
|------|------|-----------------|
| GitHub 검색 노출 | 가능 (public) | 불가 |
| GitHub "Explore" 추천 | 가능 | 불가 |
| `awesome-*` 리스트 등재 | 가능 | 링크 404 → 제거됨 |
| Google 검색 "tene secret manager" | GitHub README 노출 | 404 → 색인 제거 |
| GitHub 백링크 (도메인 권한) | github.com에서 tene.sh로 링크 | 소멸 |

**대안 채널:**

1. **tene.sh 랜딩 페이지 SEO 강화**: 기술 키워드 타겟 페이지 추가
2. **블로그/기술 포스트**: "Zero-Knowledge Secret Management" 등 키워드 선점
3. **dev.to / Hacker News / Reddit**: 출시 시 기술 포스트 배포
4. **Product Hunt 런칭**: 출시 시점 활용
5. **CLI 디렉토리 등재**: Homebrew, pkgx, 기타 패키지 매니저

**기획서 반영:** 섹션 8 리스크 #6 확장 + 섹션 10 "향후 로드맵"에 마케팅 채널 추가.

### 5-4. 비용 분석 보완

기존 기획서 섹션 9의 비용 분석은 정확하나, 다음 항목을 보완한다.

#### CloudFront 불필요 근거

| 항목 | S3 직접 서빙 | CloudFront |
|------|:-----------:|:----------:|
| 월 비용 | ~$0.07 | ~$1.50+ (최소 요금) |
| 지연 시간 (ap-northeast-2 기준) | ~50ms | ~20ms |
| 글로벌 지연 (us-east-1) | ~200ms | ~50ms |
| 도입 시점 | 지금 | 월 10,000+ 다운로드 시 |

**결론:** 초기에는 S3 직접 서빙으로 충분하다. CloudFront는 월 10,000 다운로드 또는 글로벌 사용자 증가 시점에 도입한다. `releases.tene.sh` 서브도메인을 미리 CNAME으로 설정해두면 나중에 CloudFront 전환 시 URL 변경 없이 가능하다.

#### 전체 추가 비용 요약

| 항목 | 월 비용 | 비고 |
|------|:-------:|------|
| S3 스토리지 | ~$0.01 | ~50MB/릴리스 |
| S3 요청 | ~$0.01 | GET 1,000회 |
| S3 데이터 전송 | ~$0.05 | 5GB |
| Route53 서브도메인 | ~$0.50 | releases.tene.sh |
| **합계** | **<$1/월** | 기존 ~$55/월 대비 +1.1% |

**기획서 반영:** 섹션 9에 CloudFront 불필요 근거 추가.

---

## 6. Infisical 대비 Tene 포지셔닝 전략

### Tene의 방어 가능한 차별점 (Infisical이 따라올 수 없는 것)

| 차별점 | Tene | Infisical | 전환 난이도 |
|--------|------|-----------|:-----------:|
| **Zero-Knowledge** | 서버가 물리적으로 복호화 불가 (XChaCha20 클라이언트 암호화) | 서버가 plaintext 접근 가능 | Infisical이 전환하려면 전체 아키텍처 재설계 필요 |
| **로컬-퍼스트** | 클라우드 없이 완전 동작 (SQLite vault) | 클라우드 필수 (self-host도 서버 필요) | 아키텍처 근본 차이 |
| **AI 에이전트 보안** | `tene run` 패턴 — 에이전트가 시크릿 값을 절대 보지 못함 | MCP 서버가 에이전트에게 plaintext 전달 | 설계 철학 차이 |
| **OS Keychain** | macOS Keychain / Linux libsecret 통합 | 없음 (서버 기반) | 클라이언트 앱 필요 |
| **BIP-39 복구** | 12단어 니모닉으로 마스터키 복구 | 이메일 기반 비밀번호 리셋 | 구현 가능하지만 우선순위 낮음 |
| **단일 바이너리** | `curl | sh` 한 줄로 설치, 의존성 0 | Docker Compose / Helm 또는 CLI 설치 | 아키텍처 차이 |

### 핵심 메시지

**"유일한 Zero-Knowledge Agentic Secret Runtime"**

이 한 문장이 Infisical과의 모든 차별점을 함축한다:

- **Zero-Knowledge**: Infisical은 서버가 시크릿을 볼 수 있다. Tene은 못 본다.
- **Agentic**: Infisical MCP는 에이전트에게 평문을 넘긴다. Tene은 환경변수 주입만 한다.
- **Runtime**: Tene은 `tene run -- CMD` 패턴으로 시크릿을 프로세스 수명에만 존재하게 한다.

### Tene의 기능 갭 (Infisical 대비 미구현)

| 기능 | Infisical | Tene 현황 | 우선순위 |
|------|-----------|-----------|:--------:|
| 시크릿 참조 (`${VAR}`) | 지원 | 미구현 | **높음** |
| 시크릿 스캔 (git pre-commit) | 지원 | 미구현 | 높음 |
| Dynamic Secrets (임시 DB creds) | 지원 | 미구현 | 낮음 |
| SDK (6개 언어) | Node, Python, Go, Ruby, Java, .NET | 미구현 | 중간 |
| Kubernetes Operator | 지원 | 미구현 | 낮음 |
| SSO/SAML | Enterprise | 미구현 | 낮음 |
| IP 제한 | Enterprise | 미구현 | 낮음 |
| 시크릿 로테이션 | 자동 | 팀 키만 | 중간 |
| Import/Export 포맷 | JSON, YAML, .env, PKCS12 등 | .env만 | 중간 |
| 태그/폴더 | 지원 | 환경(env)만 | 낮음 |
| 감사 로그 | 상세 (IP, 시간, 액션) | 기본 구현 | 중간 |
| 버전 관리 (시크릿 히스토리) | 지원 | 미구현 | 중간 |

**핵심 통찰:** Infisical의 기능은 대부분 **엔터프라이즈 요구사항**이다. Tene의 타겟(개인 개발자, 소규모 팀, AI 에이전트 사용자)에게는 Zero-Knowledge + 로컬-퍼스트 + 에이전트 보안이 더 중요하다.

### 컨텐츠 전략 (Private 전환 후)

| # | 제목 | 타겟 | 채널 |
|:-:|------|------|------|
| 1 | "Infisical vs Tene: Zero-Knowledge Secret Management 비교" | 시크릿 관리 도구 비교 검색자 | 블로그, SEO |
| 2 | "Why Your AI Agent Shouldn't See Your Secrets" | AI 에이전트 사용자 | Hacker News, dev.to |
| 3 | "Local-First Secret Management for Solo Developers" | 개인 개발자 | Reddit r/programming |
| 4 | "4-Layer Encryption: How Tene Achieves Zero-Knowledge" | 보안 관심 개발자 | 블로그, 기술 문서 |
| 5 | "From .env to Encrypted Vault in 30 Seconds" | .env 사용자 | Twitter/X, 데모 영상 |

---

## 7. 기능 로드맵 제안 (Infisical 갭 분석 기반)

### 우선순위 1: Private 전환과 함께 (1-2주)

Private 전환 기획서(기존)의 일부로 포함되거나, 전환 직후 바로 실행할 항목.

- [ ] install.sh S3 전환 (기존 기획서 Phase 2-3)
- [ ] GoReleaser S3 업로드 (기존 기획서 Phase 2-1)
- [ ] CLI update 명령어 S3 전환 + 체크섬 검증 (기존 기획서 Phase 2-4 + 보완)
- [ ] Landing 페이지 GitHub 링크 정리 (기존 기획서 Phase 3)
- [ ] 보안 문서 페이지 (`tene.sh/security`) 초안 작성

### 우선순위 2: 출시 전 (1개월)

제품 공개 출시에 필요한 기능 갭 보완.

- [ ] **시크릿 참조 기능** (`${VAR}` 치환) — Infisical 사용자 전환에 필수
  - `tene set DB_URL "postgres://${DB_USER}:${DB_PASS}@${DB_HOST}/db"`
  - `tene get DB_URL` → 참조 해석 후 결과 반환
  - 구현 위치: `internal/vault/` 또는 `internal/cli/run.go`

- [ ] **JSON/YAML export 지원** — `.env` 외 포맷 요구 빈도 높음
  - `tene export --format json`
  - `tene export --format yaml`
  - 구현 위치: `internal/cli/export.go`

- [ ] **Secret scanning** (git pre-commit hook) — 보안 제품의 기본 기능
  - `tene scan` 명령어
  - `.tene/hooks/pre-commit` 자동 설치
  - 구현 위치: `internal/cli/scan.go` + `internal/scanner/`

### 우선순위 3: 출시 후 (3개월)

커뮤니티 성장과 생태계 확장.

- [ ] **Go SDK** (public 패키지) — Open Core Stage 2와 연동
  - `go.tene.sh/sdk` 또는 `github.com/tomo-kay/tene-go`
  - 프로그래매틱 vault 접근 API

- [ ] **GitHub Actions 공식 액션**
  - `uses: tomo-kay/tene-action@v1`
  - CI/CD에서 시크릿 주입 (vault 접근 없이 Sync Envelope 사용)

- [ ] **MCP 서버** (Phase 2, CLAUDE.md 기반)
  - 에이전트가 `tene run` 없이도 시크릿 주입 가능
  - 단, 에이전트에게 값은 노출하지 않는 설계 유지

### 우선순위 4: 성장기 (6개월+)

Enterprise 고객 유치 및 Infisical 직접 경쟁.

- [ ] **Kubernetes Operator** — K8s Secret 자동 동기화
- [ ] **Dynamic Secrets** — 임시 DB 자격증명 (Vault-like)
- [ ] **SSO/SAML** — Enterprise 필수
- [ ] **CLI Open Core 공개** — 시나리오 C Stage 3
- [ ] **IP 제한** — Enterprise 보안 요구사항

---

## 8. 의사결정 매트릭스

### 시나리오별 비교

| 기준 | 가중치 | A: 전체 Private | B: Open Core | C: Private First, Open Later |
|------|:------:|:---------------:|:------------:|:----------------------------:|
| 보안 리스크 해소 속도 | 25% | 즉시 (5/5) | 부분적 (2/5) | 즉시 (5/5) |
| 수익 보호 | 20% | 강함 (5/5) | 중간 (3/5) | 강함→중간 (4/5) |
| 커뮤니티 성장 가능성 | 20% | 불가 (1/5) | 최대 (5/5) | 나중에 가능 (3/5) |
| 구현 복잡도 (낮을수록 좋음) | 15% | 낮음 (5/5) | 높음 (1/5) | 중간 (3/5) |
| Infisical 대비 경쟁력 | 10% | 약함 (2/5) | 강함 (5/5) | 중간→강함 (3/5) |
| 1인 운영 부담 (낮을수록 좋음) | 10% | 낮음 (5/5) | 높음 (1/5) | 단계적 (4/5) |
| **가중 합계** | 100% | **4.05** | **2.85** | **3.85** |

### 시나리오별 상세 평가

#### 시나리오 A (4.05점): 보안과 운영 효율에서 최고점

- 즉각적 효과, 낮은 리스크, 3일 만에 완료
- 그러나 성장 엔진이 없어 장기적으로 Infisical에 밀릴 가능성
- **적합 대상:** 제품 검증 전, 또는 유료 고객이 이미 있어 커뮤니티가 불필요한 경우

#### 시나리오 B (2.85점): 성장 잠재력 최대지만 현실적 부담 과다

- 1인 개발자가 monorepo 분리 + 커뮤니티 관리 + 제품 개발을 동시에 하기 어려움
- 매출 0인 상태에서 오픈소스 관리 비용은 순수 손실
- **적합 대상:** 3명+ 팀, seed 이상 투자 확보, 커뮤니티 매니저 존재 시

#### 시나리오 C (3.85점): 현실적 최선

- A의 즉각적 보안 효과 + B의 장기 성장 가능성을 결합
- 실패 비용 최소 (Open Core가 효과 없으면 private 유지)
- 단계별로 복잡도를 분산하여 1인 운영에서도 실행 가능
- **적합 대상:** 현재 Tene의 상황 (1인, 매출 0, 보안 리스크 존재)

### 최종 추천: 시나리오 C (Private First, Open Later)

**이유:**

1. **지금 당장의 문제(보안)를 즉시 해결**한다 → 기존 기획서 실행
2. **성장 가능성을 포기하지 않는다** → Month 3부터 단계적 공개
3. **실패 시 손실이 최소**이다 → Open Core가 효과 없으면 private 유지
4. **의사결정을 분산**한다 → 지금 모든 것을 결정하지 않아도 됨
5. **Infisical의 검증된 모델을 Tene 상황에 맞게 적용**한다

---

## 9. 기존 기획서 수정 제안 요약

기존 기획서(`tene-private-repo-migration.plan.md`)의 각 섹션별 구체적 수정 사항:

### 섹션 1 (개요)

| 항목 | 현재 | 제안 |
|------|------|------|
| 목적 3번 | 없음 | 추가: "4. 향후 Open Core 전환을 위한 기반 마련" |
| 범위 테이블 | 8개 항목 | 추가: "Git history 민감 정보 검토 \| 보완 (Phase 0)", "보안 문서 페이지 작성 \| 보완 (Phase 6)" |

### 섹션 4 (영향도 분석)

| 항목 | 현재 | 제안 |
|------|------|------|
| "영향 없음" 목록 | GitHub OAuth callback URL 미언급 | 명시적으로 추가: "GitHub OAuth App callback URL (repo visibility와 무관)" |
| "영향 없음" 목록 | go.mod 미언급 | 명시적으로 추가: "go.mod module path (내부 빌드만 사용, 변경 불필요)" |

### 섹션 5 (실행 계획)

| 항목 | 현재 | 제안 |
|------|------|------|
| Phase 0 | 없음 | 추가: "Git History 민감 정보 정리" (본 제안서 섹션 5-1 참조) |
| Phase 2-4 | "체크섬 검증 (선택)" | 변경: "체크섬 검증 (필수)" + 구현 상세 추가 |
| Phase 2-4 | Fallback 로직 없음 | 추가: S3 → GitHub API 이중 시도 로직 |
| Phase 3 | 대체 URL 미지정 | 추가: 대체 URL 매핑표 (본 제안서 섹션 5-2 참조) |
| Phase 4-1 | Vercel 확인 간략 | 확장: 구체적 절차 5단계 (본 제안서 섹션 5-2 참조) |
| Phase 6 | 없음 | 추가: "Infisical 대비 차별화 강화" (본 제안서 섹션 5-1 참조) |

### 섹션 6 (변경 파일 목록)

| 항목 | 현재 | 제안 |
|------|------|------|
| 총 파일 수 | 17개 | 추가: `tene.sh/security` 페이지 (apps/web/src/app/security/page.tsx 또는 유사), `tene.sh/feedback` 리다이렉트 설정 |

### 섹션 8 (리스크)

| 항목 | 현재 | 제안 |
|------|------|------|
| 리스크 수 | 7개 | 추가: #8 go.mod 변경 (불필요 확인), #9 OAuth callback (영향 없음 확인), #10 SEO 대안 채널 상세 |
| 리스크 #6 | "SEO 하락" 간략 | 확장: 구체적 영향 항목 + 대안 채널 5개 |
| 리스크 #7 | "오픈소스 신뢰도 하락" | 확장: Phase 6 보안 문서 + 향후 Open Core 계획 언급 |

### 섹션 9 (비용)

| 항목 | 현재 | 제안 |
|------|------|------|
| CloudFront | 섹션 10 로드맵에만 언급 | 추가: 불필요 근거 (비용 비교표) |

### 섹션 10 (향후 로드맵)

| 항목 | 현재 | 제안 |
|------|------|------|
| 단기 | 3개 항목 | 추가: "보안 문서 페이지 작성 (tene.sh/security)" |
| 중기 | Open Core 언급 있음 | 확장: 시나리오 C Stage 2-3 타임라인 구체화 |
| 장기 | 4개 항목 | 추가: "기능 로드맵 (Infisical 갭 분석 기반)" (본 제안서 섹션 7 참조) |

---

## 부록 A: Infisical 상세 분석 데이터

### 회사 정보

| 항목 | 내용 |
|------|------|
| 설립 | 2022 |
| 투자 | $16M Series A (2024, GV + Y Combinator) |
| 팀 규모 | ~20명 |
| GitHub Stars | 20K+ |
| 다운로드 | 1억+ |
| 기여자 | 216+ |

### 가격 구조

| Plan | 가격 | 주요 기능 |
|------|------|-----------|
| Free | $0 | 5명, 3 프로젝트 |
| Pro | $18/user/월 | 무제한, 감사 로그, IP 제한 |
| Enterprise | Custom | SSO/SAML, 전용 지원, SLA |

### 기술 스택

| 항목 | Infisical | Tene |
|------|-----------|------|
| 서버 | Node.js + Express | Go + Echo |
| DB | PostgreSQL + Redis | PostgreSQL (cloud) + SQLite (local) |
| 암호화 | AES-256-GCM (서버 사이드) | XChaCha20-Poly1305 (클라이언트 사이드) |
| Zero-Knowledge | 아님 (서버가 평문 접근 가능) | 예 (서버가 복호화 불가) |
| Self-host | Docker Compose / Helm | 해당 없음 (로컬-퍼스트) |
| CLI | Node.js 기반 (infisical CLI) | Go 단일 바이너리 |
| SDK | 6개 언어 | 미구현 |
| AI 에이전트 | MCP 서버 (평문 전달) | `tene run` (값 비노출) |
| 라이선스 | MIT Core + EE | Proprietary (현재) |

### Infisical의 약점 (Tene이 공략할 지점)

1. **서버가 시크릿을 본다**: Infisical은 서버 사이드 암호화로, 관리자/해커가 서버 접근 시 전체 시크릿 노출
2. **클라우드 의존**: Self-host도 서버 운영 필요, 진정한 "로컬" 사용 불가
3. **AI 에이전트 보안 미비**: MCP 서버가 에이전트에게 평문 시크릿을 넘겨줌
4. **무거운 설치**: Docker Compose 또는 Helm 필요 (vs Tene `curl | sh`)
5. **Node.js 기반**: 메모리 사용량 높음, cold start 느림 (vs Go 단일 바이너리)

---

## 부록 B: 시나리오 C 실행 시 기존 기획서와의 관계

```
기존 기획서 (Plan)          이 제안서 (Proposal)
═══════════════════        ═══════════════════════════
Phase 1-5 (3일)     ←───  시나리오 C > Stage 1 (그대로 실행)
                    ←───  Phase 0 추가 (git history 검토)
                    ←───  Phase 6 추가 (차별화 강화)
                    ←───  Phase 2-4 보완 (체크섬, fallback, Vercel)
                    ←───  Phase 3 보완 (대체 URL 전략)

섹션 10 로드맵       ←───  시나리오 C > Stage 2 (Month 3, crypto 공개)
  중기: Open Core   ←───  시나리오 C > Stage 3 (Month 4-5, CLI 공개)

(없음)              ←───  기능 로드맵 (섹션 7, Infisical 갭 분석)
(없음)              ←───  포지셔닝 전략 (섹션 6)
(없음)              ←───  컨텐츠 전략 (섹션 6)
```

**요약:** 기존 기획서는 "어떻게 private 전환할 것인가"에 집중한다. 이 제안서는 "왜 이 전략인가", "다음에 무엇을 할 것인가", "Infisical 대비 어떻게 포지셔닝할 것인가"를 보완한다. 기존 기획서의 기술적 실행 계획은 거의 수정 없이 유지되며, Phase 0/6 추가와 일부 Phase 상세화가 핵심 변경 사항이다.
