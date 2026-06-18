# tene AI 발견성 개선 — Codebase Audit (수정 대상 현황)

> **목적**: 22개 개선 항목이 영향을 미치는 **모든 파일의 현재 상태**를 증거 수준으로 기록. Design 문서(`ai-discoverability.design.md`)의 diff 기준점.
> **감사 방법**: 실제 파일 열람 + `gh` CLI + grep. 에이전트 간접 보고 없음.
> **기준일**: 2026-04-23, `staging` 브랜치, v1.0.4 시점.

---

## 1. 수정 대상 파일 전수 (File Inventory)

### 1.1 tene CLI Go 코드

| 파일 | 용도 | 수정 항목 | 줄수 |
|------|-----|---------|:---:|
| `internal/cli/root.go` | CLI 루트 + 전역 flag + 서브커맨드 등록 | A-1 (Cloud cmd stub 옵션), U-4 (help footer), S-4 (completion 등록) | 225 |
| `internal/cli/get.go` | `tene get <KEY>` 구현 | **U-1 (비대화형 stdout 가드)** | 73 |
| `internal/cli/list.go` | `tene list` 구현 | 변경 없음 (참조용) | 135 |
| `internal/cli/run.go` | `tene run --` 구현 | 변경 없음 (참조용) | 164 |
| `internal/cli/version.go` | version 명령 | 변경 없음 | 미확인 |
| `cmd/tene/main.go` | 엔트리 포인트 | 변경 없음 | 미확인 |
| (신규) `internal/cli/completion.go` | Shell completion 생성 명령 | S-4 | ~50 (신규) |
| (신규) `internal/cli/manpage.go` 또는 별도 generator | man page 생성 | S-4 | ~30 (신규) |
| `pkg/errors/errors.go` | 에러 코드 정의 | U-1 (`STDOUT_SECRET_BLOCKED` 신규) | 미확인 |

### 1.2 tene CLI 배포 설정

| 파일 | 용도 | 수정 항목 |
|------|-----|---------|
| `.goreleaser.yml` | 릴리스 파이프라인 | **S-1 `brews:`**, **S-2 `dockers:`**, S-4 `man_pages:` · completion archive, S-5 `signs:` (선택) |
| (신규) `Dockerfile` | Docker 이미지 빌드 | S-2 |
| `apps/web/public/install.sh` | curl install 스크립트 | S-3 (힌트 확장), S-5 (GPG verify, 선택) |
| (신규 리포) `tene-ai/homebrew-tap/Formula/tene.rb` | Homebrew formula | S-1 (GoReleaser 자동 생성) |

### 1.3 GitHub 리포 루트 / `.github/`

| 파일 | 현재 | 수정 항목 |
|------|:---:|---------|
| `README.md` | 있음 (474줄) | **A-1 (Cloud Commands 섹션 수정)** |
| `LICENSE` | 있음 | 변경 없음 |
| `SECURITY.md` | **없음** | **I-1a (신규)** |
| `CODE_OF_CONDUCT.md` | **없음** | I-1b (신규) |
| `CONTRIBUTING.md` | **없음** | I-1c (신규) |
| `CHANGELOG.md` | **없음** | U-1 배포 시 신규 (releasing 후) |
| `.github/FUNDING.yml` | **없음** | D-5 (신규) |
| `.github/ISSUE_TEMPLATE/bug_report.yml` | **없음** | I-1d (신규) |
| `.github/ISSUE_TEMPLATE/feature_request.yml` | **없음** | I-1d (신규) |
| `.github/PULL_REQUEST_TEMPLATE.md` | **없음** | I-1e (신규) |
| `.github/copilot-instructions.md` | **없음** | U-2 (신규) |
| `.github/CODEOWNERS` | **없음** | (선택, CONTRIBUTING 부속) |
| `.github/dependabot.yml` | **없음** | (선택, 별도 feature) |
| `.github/workflows/ci.yml` | 있음 | 변경 없음 (기존 활용) |
| `.github/workflows/auto-tag.yml` | 있음 | 변경 없음 |

### 1.4 랜딩 `apps/web/`

| 파일 | 용도 | 수정 항목 |
|------|-----|---------|
| `apps/web/src/app/layout.tsx` | 루트 metadata + JSON-LD | **A-2 description**, A-4 Organization + WebSite, D-4 link rel ai-index |
| `apps/web/src/app/page.tsx` | 홈 페이지 구성 | 변경 없음 (컴포넌트 조합) |
| `apps/web/src/app/robots.ts` | robots.txt 생성 | **D-2 LLM 봇 명시 + `.tene/` disallow** |
| `apps/web/src/app/sitemap.ts` | sitemap.xml 생성 | 변경 없음 (이미 완비) |
| `apps/web/src/components/hero.tsx` | 히어로 컴포넌트 | A-5 (히어로 데이터 참조 경로) |
| `apps/web/src/data/hero.ts` | 히어로 텍스트 데이터 | **A-5 ("Tene is X" 문장 추가)** |
| `apps/web/src/components/seo/software-jsonld.tsx` | /vs/ 전용 JSON-LD | **A-3 (aggregateRating 제거)** |
| `apps/web/src/components/seo/article-jsonld.tsx` | 블로그 JSON-LD | 변경 없음 (이미 우수) |
| `apps/web/src/components/seo/faq-jsonld.tsx` | FAQ JSON-LD | 변경 없음 |
| `apps/web/src/components/seo/blog-index-jsonld.tsx` | 블로그 index JSON-LD | 변경 없음 |
| (신규) `apps/web/src/components/trust.tsx` | Trust Section 컴포넌트 | **I-2** |
| (신규) `apps/web/src/components/breadcrumb.tsx` | 공용 Breadcrumb | I-3 |
| `apps/web/src/app/vs/[slug]/page.tsx` | /vs/ 동적 라우트 | I-3 (breadcrumb 삽입) |
| `apps/web/src/app/blog/[slug]/page.tsx` | 블로그 동적 라우트 | I-3 (breadcrumb 삽입) |
| `apps/web/next.config.ts` | Next 설정 | (선택) next/image AVIF/WebP 추가 |
| `apps/web/public/llms.txt` | LLM 전용 인덱스 | 변경 없음 (이미 우수) |
| `apps/web/public/llms-full.txt` | LLM 확장 참조 | U-5 완료 후 CLI 레퍼런스 링크 추가 |
| (신규) `apps/web/public/.well-known/ai.json` | AI 에이전트 발견 표준 | **D-4** |
| (신규) `apps/web/src/app/cli/page.tsx` | /cli CLI 레퍼런스 공개 페이지 | U-5 |

### 1.5 문서 (`docs/`)

| 파일 | 수정 항목 |
|------|---------|
| `docs/00-pm/ai-discoverability.prd.md` | (이 PRD, 이미 생성) |
| `docs/01-plan/ai-discoverability.md` | (이 Plan, 이미 생성) |
| `docs/02-design/features/ai-discoverability.codebase-audit.md` | (이 Audit, 이 문서) |
| `docs/02-design/features/ai-discoverability.design.md` | Design (다음 단계) |
| `docs/03-report/ai-discoverability-2026-04-23.md` | 이미 존재 (기반 감사) |
| (신규) `docs/cli-reference.md` | **U-5** |
| (신규) `CHANGELOG.md` | U-1 release 시 |

---

## 2. 현재 상태 증거 (Before-state Evidence)

### 2.1 CLI — `get.go` (U-1 대상)

**파일**: `internal/cli/get.go:18-72` (73줄 전체)

```go
func runGet(cmd *cobra.Command, args []string) error {
    // ... (load vault, decrypt)
    
    if flagJSON {
        return printJSON(map[string]any{
            "ok":          true,
            "name":        keyName,
            "value":       string(plaintext),
            "environment": env,
        })
    }

    fmt.Print(string(plaintext))        // ⚠️ LINE 67 — 비대화형에서도 stdout 평문 출력
    if isTerminal() {
        fmt.Println()
    }
    return nil
}
```

**문제**: `isTerminal()` 은 있지만 출력 차단 판단이 아닌 **개행 문자 추가** 용도로만 쓰임. AI 에이전트가 Bash 툴로 `tene get X` 호출 시 stdout 이 pipe 이므로 `isTerminal()==false` → 평문이 tool_result 로 유입.

### 2.2 CLI — `root.go` (A-1, U-4, S-4)

**파일**: `internal/cli/root.go:57-91`

```go
func init() {
    rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output in JSON format")
    // ... 다른 전역 flag 5종

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
    rootCmd.AddCommand(whoamiCmd)
    rootCmd.AddCommand(versionCmd)
    rootCmd.AddCommand(updateCmd)

    // Cloud commands — removed from CLI while being redesigned.        ⚠️
    // Code preserved in: login.go, logout.go, push.go, pull.go,
    // sync_cmd.go, billing.go, team.go
    // Re-enable by uncommenting:
    // rootCmd.AddCommand(newLoginCmd())
    // ... (7개 명령 주석 처리)
}
```

**문제**:
1. README.md:184-194 에 이 7개 명령이 public 으로 표시됨 → AI 훈련 데이터 왜곡.
2. `completion` 명령 미등록 — `tene completion bash` 불가능.
3. `rootCmd.Long` 에 리소스 링크 없음 → `tene --help` 에서 llms.txt 안내 부재.

### 2.3 CLI — README.md Cloud Commands 섹션 (A-1)

**파일**: `README.md:182-194`

```markdown
### Cloud Commands (requires [app.tene.sh](https://app.tene.sh) account)

| Command | Description |
|---------|-------------|
| `tene login` | OAuth login to Tene Cloud |
| `tene push` | Encrypt and upload vault to cloud |
| `tene pull` | Download and decrypt remote vault |
| `tene sync` | Push + Pull combined (Pro plan) |
| `tene team create` | Create team + generate project key |
| `tene team invite` | Invite member with X25519 key wrapping |
| `tene billing` | View subscription status |
| `tene billing upgrade` | Open checkout page |
```

**현실**: `root.go:80-90` 에서 전부 주석 처리. 실행 시 `Error: unknown command "login"` 발생.

### 2.4 랜딩 — `layout.tsx` metadata (A-2, A-4, D-4)

**파일**: `apps/web/src/app/layout.tsx:17-79`

```tsx
export const metadata: Metadata = {
  title: "Tene — Your .env is not a secret. AI can read it.",
  description:
    "Your .env is not a secret — AI can read it. Tene encrypts secrets locally and injects them at runtime so AI agents never see the values. XChaCha20-Poly1305 encryption. No server, no signup, free and open source.",
  //   ^^^ 211자 — Google SERP 155자 컷오프 초과 ⚠️
  keywords: [ /* 15 items */ ],
  authors: [{ name: "agent-kay-it", url: "https://github.com/agent-kay-it" }],
  // ... canonical, icons, openGraph, twitter, robots 완비
};

const jsonLd = {
  "@context": "https://schema.org",
  "@graph": [
    { "@type": "SoftwareApplication", /* ... */ },
    { "@type": "FAQPage", /* 6 Q&A */ },
    { "@type": "HowTo", /* 4 steps */ },
    // ⚠️ 누락: Organization, WebSite + SearchAction
  ],
};
```

### 2.5 랜딩 — `robots.ts` (D-2)

**파일**: `apps/web/src/app/robots.ts:1-19` (19줄 전체)

```ts
export default function robots(): MetadataRoute.Robots {
  const base = "https://tene.sh";
  return {
    rules: [
      { userAgent: "*", allow: "/" },
    ],
    sitemap: `${base}/sitemap.xml`,
    host: base,
  };
}
```

**문제**: GPTBot · ClaudeBot · PerplexityBot 등 LLM 봇 **명시 allow 없음**. `.tene/` disallow 없음.

### 2.6 랜딩 — `software-jsonld.tsx` (A-3)

**파일**: `apps/web/src/components/seo/software-jsonld.tsx:19,48-57`

```ts
const TENE_STARS = 5; // update at each milestone; keeps rating schema honest
//                                                    ^^^ 주석은 "honest" 라고 하지만 실제로는 schema 의미 오용

// 라인 48-57
aggregateRating:
  TENE_STARS > 0
    ? {
        "@type": "AggregateRating",
        ratingValue: "4.9",
        reviewCount: String(TENE_STARS),  // ⚠️ reviewCount 는 리뷰 개수, stars 는 별점/star 수
        bestRating: "5",
        worstRating: "1",
      }
    : undefined,
```

**문제**: `reviewCount` 는 Schema.org 에서 **명시적으로 "리뷰 개수"**. GitHub stars ≠ review. Google 가이드라인 위반 → rich result 박탈 + 수동 페널티 리스크.

### 2.7 랜딩 — `hero.tsx` (A-5)

**파일**: `apps/web/src/components/hero.tsx:19-27`

```tsx
<h1 className="text-3xl font-bold leading-tight tracking-tight whitespace-nowrap sm:text-4xl md:text-5xl lg:text-[3.25rem] xl:text-6xl">
  {heroData.h1}
  <br />
  <span className="text-accent">{heroData.h1Accent}</span>
</h1>

<p className="mx-auto mt-6 max-w-xl text-base text-muted leading-relaxed sm:text-lg lg:mx-0">
  {heroData.sub}
</p>
```

→ `heroData.sub` 의 실제 내용을 `src/data/hero.ts` 에서 확인 필요. "Tene is X" 정의 첫 단어 유무 확인.

### 2.8 배포 — `.goreleaser.yml` (S-1, S-2, S-4, S-5)

**파일**: `.goreleaser.yml` (74줄 전체)

- `builds:` ✅, `archives:` ✅, `checksum:` ✅, `blobs:` (S3) ✅
- ❌ **`brews:` 섹션 없음**
- ❌ **`dockers:` 섹션 없음**
- ❌ **`nfpms:` 섹션 없음**
- ❌ **`scoops:` 섹션 없음**
- ❌ **`man_pages:` 섹션 없음**
- ❌ **`signs:` 섹션 없음**
- ❌ **`completions/*` archive 포함 안 됨**

### 2.9 GitHub — Community Health 42% 증거

**실측 (`gh api repos/tene-ai/tene/community/profile`)**:
```json
{
  "health_percentage": 42,
  "files": {
    "readme": true,
    "license": true,
    "code_of_conduct": false,
    "contributing": false,
    "issue_template": false,
    "pull_request_template": false
  }
}
```

**실측 (`gh repo view`)**:
```json
{
  "hasDiscussionsEnabled": false,
  "isSecurityPolicyEnabled": false,
  "usesCustomOpenGraphImage": false,
  "fundingLinks": []
}
```

### 2.10 `.github/` 디렉토리 현황

**실측 (`ls -la tene/.github/`)**:
```
drwxr-xr-x@  3 popup-kay  staff    96B 4  6 21:21 .
drwxr-xr-x@ 42 popup-kay  staff  1344B 4 23 02:03 ..
drwxr-xr-x@  4 popup-kay  staff   128B 4 22 15:09 workflows
```

→ `workflows/` 만 존재. 다른 모든 파일 부재.

---

## 3. 수정 영향도 매트릭스

| 수정 항목 | 영향 파일 수 | 외부 breaking? | CI 변경 | 새 리포 필요 |
|:---:|:---:|:---:|:---:|:---:|
| D-1 OG 이미지 | 0 (Settings UI) | No | No | No |
| D-2 robots.ts | 1 | No | No | No |
| D-3 Discussions | 0 (Settings UI) | No | No | No |
| D-4 ai-index + ai.json | 2 | No | No | No |
| D-5 FUNDING.yml | 1 (신규) | No | No | No |
| A-1 README Cloud 섹션 | 1 | **Yes (문서 수정, 독자 혼란 해소)** | No | No |
| A-2 description | 1 | No (SEO 개선) | No | No |
| A-3 aggregateRating 제거 | 1 | No (Google penalty 회피) | No | No |
| A-4 Organization + WebSite | 1 | No | No | No |
| A-5 "Tene is X" | 1 (data) | No | No | No |
| I-1 Community Health 4+2 | 5 (신규) | No | No | No |
| I-2 Trust Section | 2 (신규 컴포 + 홈 조합) | No | No | No |
| I-3 Breadcrumb | 3 (신규 + /vs + /blog) | No | No | No |
| S-1 Homebrew tap | 1 (.goreleaser) + 신규 리포 | No | **Yes** | **Yes (`tene-ai/homebrew-tap`)** |
| S-2 Docker GHCR | 2 (Dockerfile + .goreleaser) | No | **Yes (packages:write)** | No |
| S-3 install.sh 힌트 | 1 | No | No | No |
| S-4 man/completion | 2 (.goreleaser + CLI completion.go) | No (opt-in) | **Yes (archive 포함)** | No |
| S-5 GPG 서명 | 2 (.goreleaser + install.sh) + 키 관리 | No | **Yes (GPG 환경)** | No |
| **U-1 tene get 가드** | 1 (`get.go`) | **YES — breaking** | No | No |
| U-2 copilot-instructions | 1 (신규) | No | No | No |
| U-4 tene --help 푸터 | 1 (root.go) | No | No | No |
| U-5 cli-reference.md + /cli | 2 (신규) | No | No | No |

**핵심 위험**: **U-1** 만 유일한 breaking change. 나머지는 additive 또는 bugfix.

---

## 4. 테스트 전략 요약 (Design 문서에서 상세)

### 4.1 자동 테스트

- CLI: `go test ./internal/cli/...` 기존 suite + U-1 TDD 신규 테스트
- 빌드 검증: `go build ./cmd/tene`
- Goreleaser dry run: `goreleaser release --snapshot --clean --skip=publish`
- Docker 빌드: `docker build -t tene:test . && docker run tene:test version`
- 랜딩: `pnpm build` + `pnpm lint`
- 구조화 데이터 검증: Google Rich Results Test API (CI 추가 선택)

### 4.2 수동 테스트 (체크리스트)

- `curl -I https://tene.sh/robots.txt` → LLM bots 포함 확인
- `curl https://tene.sh/.well-known/ai.json` → 200
- `brew install tene-ai/tap/tene && tene version` → 성공
- `docker run ghcr.io/tene-ai/tene version` → 성공
- `tene get X | cat` → exit 2 + `STDOUT_SECRET_BLOCKED`
- `tene get X` (터미널) → 정상 출력
- `tene completion bash` → 스크립트 출력
- Rich Results Test (Google) — SoftwareApplication · FAQPage · HowTo · Organization · BlogPosting 전부 valid
- PageSpeed Insights → LCP ≤ 2.5s, Performance ≥ 90

---

## 5. 롤백 전략

각 항목 독립 커밋 → 문제 시 `git revert <sha>` 로 개별 롤백 가능.

특히 **U-1 (get.go 가드)** 는 별도 태그(`v1.1.0-rc1`)로 테스트 배포 → 1주 모니터링 후 stable 승격. 이슈 발생 시 `v1.0.4` 로 sed-back 가능.

Homebrew tap / Docker 이미지는 태그 단위 immutable. 문제 시 다음 tag 로 hotfix (`v1.1.1`).

---

## 6. 다음 단계

이 Audit 로 "**어디를**" 수정하는지 확인 완료. 다음 문서에서 "**어떻게**" 수정하는지:

→ `docs/02-design/features/ai-discoverability.design.md`
