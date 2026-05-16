# tene — AI·사람 발견성 심층 감사 보고서

**기준일**: 2026-04-23
**대상 커밋**: `staging` 기준 v1.0.4
**범위**: 3 surface (GitHub 리포 `agent-kay-it/tene` · 랜딩 `tene.sh` = `apps/web/` · CLI `cmd/ & internal/ & pkg/`)
**방법**: 모든 항목은 **실제 코드/설정/GitHub API 직접 열람**으로 확인. 에이전트 간접 보고서는 의도적으로 배제.

---

## 0. 목표 — 이 보고서가 해결하려는 것

> **[ AI / 사람 ] → (웹 검색 · 에디터 내 추천) → [발견] → [인지] → [흥미] → [설치] → [사용]**

3 개의 surface 가 이 6 단계를 지원해야 한다:

| Surface | 담당 단계 | 핵심 질문 |
|---------|----------|----------|
| **GitHub 리포** | 발견 (AI 훈련 · 크롤러 인덱싱) · 인지 (첫 30 초) | "진짜 프로젝트인가, 누가 관리하나, MIT 맞나" |
| **tene.sh 랜딩** | 인지 (60 초 요약) · 흥미 (기능 매칭) · 설치 (install 한 줄) | "내 문제에 맞나, 바로 깔 수 있나" |
| **tene CLI** | 설치 직후 · 사용 (리텐션) | "install 후 무얼 치면 되나, AI 가 자동으로 잘 쓰나" |

모든 항목은 이 6 단계 중 **어느 단계를 차단하는가** 로 우선순위가 결정된다.

---

## 1. 실측 요약 (Evidence Snapshot)

### 1.1 GitHub 리포 실측

| 항목 | 실측 값 | 출처 |
|------|---------|------|
| description | "AI-safe secret manager CLI for Claude Code, Cursor, and other AI agents. Local-first, encrypted, no cloud." | `gh repo view` |
| topics | 20 개 (api-key-management, claude-code, codex, cursor, gemini, linux, macos, windows, windsurf, ai-agents, cli, developer-tools, devsecops, dotenv, encryption, go, opensource, secret-management, vault, vibe-coding) | `gh repo view` |
| stars / forks / open issues | **7 / 0 / N/A** | `gh repo view` |
| Discussions | **false** (비활성) | `gh repo view` |
| Wiki | true | `gh repo view` |
| Security Policy | **false** (SECURITY.md 없음) | `gh api .../community/profile` |
| FUNDING.yml | **없음** (`fundingLinks: []`) | `gh repo view` |
| Custom OG image | **false → GitHub 기본 이미지 사용** | `gh repo view` → `opengraph.githubassets.com/...` |
| License | MIT | `gh repo view` |
| Community Health | **42 %** | `gh api .../community/profile` |
| └ README · LICENSE | ✅ · ✅ | 상동 |
| └ CoC · CONTRIBUTING · issue_template · PR_template | ❌ · ❌ · ❌ · ❌ | 상동 |
| Traffic 14 일 views | **337 (uniques 24)** | `gh api traffic/views` |
| Traffic 14 일 clones | **800 (uniques 349)** | `gh api traffic/clones` |
| 주요 Referrers | github.com 23, Google 2, vercel.com 2, reddit.com 1 | `gh api traffic/popular/referrers` |
| 최근 릴리스 | v1.0.4 (2026-04-22) · 5 릴리스 / 5 일 | `gh release list` |

**해석**:
- description / topics / README 의 품질은 상위 10 %. 하지만 **stars 7, 외부 Referrer 5 건** = 세상에 존재가 거의 알려지지 않음.
- **clones 800 > views 337** 는 CI / 자동 봇 활동이 사람 방문을 능가한다는 뜻. 실제 인간 인식도는 아직 매우 낮음.
- **Custom OG 이미지 미사용**: HN · X · Reddit 어디에 링크를 붙여도 `opengraph.githubassets.com/{hash}/agent-kay-it/tene` 기본 자동 생성 이미지가 나옴. 브랜딩/인지 단계 누수.

### 1.2 랜딩 `tene.sh` 실측

| 항목 | 실측 값 | 파일·라인 |
|------|---------|----------|
| metadata.title | "Tene — Your .env is not a secret. AI can read it." (49자 ✅) | `layout.tsx:18` |
| metadata.description | **211 자 (Google SERP 컷오프 155 자 초과)** | `layout.tsx:19-20` |
| canonical | `https://tene.sh` ✅ | `layout.tsx:42` |
| OG image | `/og-image.png` 1200 × 630 · **1.4 MB** | `layout.tsx:55-62` + public/og-image.png |
| JSON-LD @graph (홈) | SoftwareApplication + FAQPage(6Q) + HowTo(4단계) | `layout.tsx:81-179` |
| 홈 JSON-LD 누락 | **Organization, WebSite(SearchAction), BreadcrumbList** | 상동 |
| Blog JSON-LD | BlogPosting + BreadcrumbList (articleSection · sameAs · inLanguage · wordCount · timeRequired 완비) | `components/seo/article-jsonld.tsx` |
| /vs/ JSON-LD | SoftwareApplication + BreadcrumbList + FAQPage | `components/seo/software-jsonld.tsx` |
| /vs/ aggregateRating | `ratingValue=4.9, reviewCount=5` (= `TENE_STARS` = GitHub stars) | `software-jsonld.tsx:19,48-56` ⚠️ |
| robots.txt | `User-agent: *, Allow: /` **(LLM 봇 명시 0, `.tene/` disallow 0)** | `robots.ts:7-19` |
| sitemap.xml | 홈/vs/vs[slug]/blog/blog[slug]/tag/rss 완비 | `sitemap.ts:9-64` |
| lang | `en` 하드코딩 | `layout.tsx:188` |
| hreflang | 없음 | — |
| GA4 | 연결됨 (`G-9MRWMY6XBE`) | `layout.tsx:204` |
| CSP/보안 헤더 | HSTS · X-Frame-Options · CSP · Permissions-Policy 완비 | `next.config.ts:18-25` |
| next/image | **설정 없음** — AVIF/WebP 자동 변환 비활성 | `next.config.ts` (이미지 섹션 부재) |
| 폰트 preload | 없음 (Geist Sans/Mono 기본 로드) | `layout.tsx:7-14` |
| Product Hunt 배지 | `<img>` native · **`loading="lazy"` 없음** | `hero.tsx:52-57` |
| `/llms.txt` (public) | 89 줄 · H1 · blockquote · 7 섹션 · Copilot 명시 · /vs 링크 · 블로그 RSS | `public/llms.txt:1-89` |
| `/llms-full.txt` (public) | 14 KB · 전체 확장 참조 | `public/llms-full.txt` |
| `<link rel="ai-index">` | **없음** | `layout.tsx:191-196` |
| `.well-known/ai.json` | **없음** | public/ |
| 데모 GIF 총량 | 6 개 **~8 MB** (multi-env 2.1 MB, claude-refuses 1.7 MB, the-full-story 1.3 MB, security-proof 1.2 MB, dotenv-gone 933 KB, env-migration 681 KB) | `public/demo/` |
| og-image.png | 1.4 MB | `public/og-image.png` |
| 블로그 MDX | 7 편 (frontmatter: slug, title, description, publishedAt, tags, faqs 3+, canonicalUrl 완비) | `content/blog/*.mdx` |
| /vs 페이지 | 5 편 (doppler · dotenv · dotenv-vault · infisical · vault) | `src/app/vs/[slug]/page.tsx` |

### 1.3 CLI (`internal/cli/`) 실측

| 항목 | 실측 값 | 파일·라인 |
|------|---------|----------|
| rootCmd.Use | `tene` | `root.go:49` |
| rootCmd.Version | build-time `version` · Cobra 자동 `--version` | `root.go:17,52` |
| 전역 persistent flags | `--json` · `--quiet` · `--env` · `--dir` · `--no-color` · `--no-keychain` | `root.go:58-63` |
| `--json` 상속 | **모든 서브커맨드가 상속** (persistent) ✅ | 상동 |
| Cloud 명령 등록 | **주석 처리** (login/logout/push/pull/sync/billing/team) | `root.go:80-90` |
| README "Cloud Commands" 섹션 | **8 줄 테이블로 public** | `README.md:184-194` |
| → 문서-실장 불일치 | **README 는 Cloud 명령을 public 으로 보임, 실제 바이너리에는 없음** | ⚠️ |
| `tene list` stdout 정책 | **값은 `maskValue()` 마스킹 후 출력** — AI-safe | `list.go:48-95` |
| `tene list --json` | 이름 + preview(마스크) + version + updatedAt | `list.go:40-67` |
| `tene get <KEY>` 기본 출력 | **`fmt.Print(plaintext)` → stdout 평문** | `get.go:67` ⚠️ |
| `tene get --json` | `{name, value, environment}` — value 평문 포함 | `get.go:58-65` |
| `tene run` 메시지 경로 | 모두 `os.Stderr` ✅ (child stdout 보존) | `run.go:83-97` |
| `tene run` exit code 전파 | `os.Exit(exitErr.ExitCode())` ✅ | `run.go:110-113` |
| `TENE_MASTER_PASSWORD` env | 지원 ✅ | `root.go:162-164` |
| `--no-keychain` (CI/CD) | 지원 ✅ | `root.go:131-137` |
| 마스터 패스워드 프롬프트 | `os.Stderr` ✅ (stdout 오염 방지) | `root.go:171-173` |
| Audit log | `AddAuditLog()` 호출 (secret.read, secrets.inject) | `get.go:56`, `run.go:101` |
| `--help` (short/long/example) | cobra 기본 + `run` 에만 `Example:` | `run.go:17-24` |
| Shell completion 생성 | **Cobra 기본 `completion` 명령 미등록** | `root.go:65-78` |
| man page | **없음** | `.goreleaser.yml` `man_pages:` 섹션 부재 |

**결정적 발견**:
1. **`tene get` 은 AI-safe 가 아니라 "AI 에게 쓰지 말라고 룰 파일로 부탁하는" 상태**다. 룰 파일을 로드하지 않는 웹 기반 AI (일부 ChatGPT 플러그인 · WebFetch 경유 Claude 등) 는 이 경고를 보지 않고 `tene get` 을 호출할 수 있다.
2. **`--json` 은 이미 모든 명령에 전역 상속**. 감사 #8 에서 "JSON 일부만 지원" 은 **거짓**. 실제는 전역.
3. **README 에 Cloud Commands 8 줄이 버젓이 있으나 코드에서는 주석 처리**. AI가 README 만 읽으면 "`tene login`" 을 추천하지만 실행 시 `unknown command` 발생 → **설치 직후 첫 경험에서 신뢰 파괴**.

### 1.4 배포 (`.goreleaser.yml` + `install.sh`) 실측

| 항목 | 실측 값 | 출처 |
|------|---------|------|
| 빌드 매트릭스 | darwin/linux/windows × amd64/arm64 (windows/arm64 제외) = **5 바이너리** | `.goreleaser.yml:13-22` |
| 아카이브 포맷 | tar.gz (+ windows: zip) | `.goreleaser.yml:29-39` |
| 체크섬 | SHA-256 (`checksums.txt`) | `.goreleaser.yml:40-42` |
| **brews** (Homebrew tap) | **섹션 없음** | — ❌ |
| **dockers** (Docker GHCR) | **섹션 없음** | — ❌ |
| **nfpms** (deb/rpm) | **섹션 없음** | — ❌ |
| **scoops** (Windows) | **섹션 없음** | — ❌ |
| **snapcrafts** / **winget** / **nix** | **섹션 없음** | — ❌ |
| **man_pages** | **섹션 없음** | — ❌ |
| **chocolateyPackages** | **섹션 없음** | — ❌ |
| **shell completion 배포** | **아카이브에 포함 안 됨** (completion 명령 자체도 root.go 미등록) | — ❌ |
| GPG signing | **없음** | `.goreleaser.yml` `signs:` 부재 |
| 바이너리 호스팅 | S3 `tene-releases` (ap-northeast-2 **단일 리전**) | `.goreleaser.yml:67-74` |
| install.sh SHA-256 검증 | **있음** ✅ | `install.sh:84-95` |
| install.sh GPG 검증 | **없음** | — |
| install.sh 체크섬 출처 | **바이너리와 같은 S3 버킷** (bucket compromise 시 방어 0) | `install.sh:75` |
| install.sh Windows | "Use WSL" 1 줄 안내 | `install.sh:27` |

### 1.5 `.github/` 실측

```
.github/workflows/auto-tag.yml
.github/workflows/ci.yml
```

- **ISSUE_TEMPLATE/ 없음**
- **PULL_REQUEST_TEMPLATE.md 없음**
- **CODEOWNERS 없음**
- **FUNDING.yml 없음**
- **dependabot.yml 없음**
- **copilot-instructions.md 없음** (Copilot 공식 우선 경로)

리포 루트에 **SECURITY.md · CODE_OF_CONDUCT.md · CONTRIBUTING.md · CHANGELOG.md 도 없음** (README 본문에 해당 내용은 포함돼 있지만 GitHub Community Health 체크는 전용 파일만 인식).

---

## 2. 퍼널 × Surface 매트릭스 — 어디서 막히는가

각 셀 = 해당 Surface 에서 해당 퍼널 단계를 차단하는 실측 기반 미흡 구현.
`✅` = 이미 충분 · `⚠️` = 개선 필요 · `🔴` = 단계 차단.

| | **GitHub 리포** | **tene.sh 랜딩** | **tene CLI** |
|---|---|---|---|
| **발견 (Web 검색 · AI 추론)** | 🔴 stars 7 + **Custom OG 이미지 부재** → HN·X에 링크해도 GitHub 기본 문자열 이미지로 노출 · 🔴 외부 referrer 월 ~5건 · ⚠️ Discussions OFF (Q&A 페이지 검색 인덱싱 자산 손실) | ⚠️ robots.txt가 LLM봇(GPTBot·ClaudeBot·PerplexityBot·CCBot) 명시 allow 없음 · ⚠️ `<link rel="ai-index">` 없음 · ⚠️ `.well-known/ai.json` 없음 | ✅ llms.txt · llms-full.txt 고품질 |
| **인지 (30-60초 요약)** | ✅ description · topics(20) · README H1 양호 · ⚠️ README Cloud Commands 섹션이 미출시 기능을 "있음"으로 표시 → AI가 잘못 인용 | ⚠️ metadata.description 211자(SERP 컷오프) · ⚠️ 홈에 Organization JSON-LD 없음 → 엔티티 그래프 consolidation 불가 · ⚠️ /vs aggregateRating=4.9·reviewCount=stars값 (schema 의미 오용 + Google rich-result penalty 리스크) | ✅ `tene --help` (Cobra 기본) |
| **흥미 (기능/크리덴셜 매칭)** | 🔴 Community Health 42% (SECURITY.md/CoC/CONTRIBUTING/issue_tpl/PR_tpl 전부 부재) = "이게 진짜 유지보수되는 프로젝트인가?" 신호 약함 · ⚠️ FUNDING.yml 없음 | ⚠️ stars 수·testimonial·audit 배지 랜딩에 전무 · ⚠️ Product Hunt 배지 hero 구석 · ⚠️ `/vs/*` → 홈 back breadcrumb 없음 · ✅ 7편 블로그 + 5편 /vs + 6 데모 GIF 품질 우수 | ✅ `tene list` 마스킹 · `tene run --` stderr 분리 · 전역 `--json` |
| **설치 (curl·brew·docker·go install)** | ⚠️ README install 섹션 양호 · ⚠️ Homebrew 명령 예시가 없어 AI는 curl만 추천 | ⚠️ 설치 후 "다음 단계" 링크 약함 (퀵스타트/docs URL 히어로 부재) | 🔴 **`brew install tene` 작동 안 함** (.goreleaser.yml에 brews 섹션 자체 부재) · 🔴 **`docker run ghcr.io/agent-kay-it/tene` 작동 안 함** · ⚠️ GPG 서명 미지원 (엔터프라이즈 이탈) |
| **사용 (첫 30분 · 리텐션)** | — | — | ⚠️ `tene get <KEY>` 기본 출력이 stdout 평문 → 룰 파일을 로드 못 한 AI에선 유출 가능 (CLI 레벨 안전장치 부재) · 🔴 README "Cloud Commands" = 주석처리된 7명령 → **사용자 첫 `tene login` 시 "unknown command"** · ⚠️ shell completion 미배포 → AI도 사람도 정확한 하위명령 추측 불가 · ⚠️ man page 없음 |

---

## 3. 개선 항목 (퍼널 단계·우선순위·파일 기반)

### 3.A 발견 (Discovery) — AI 가 tene 의 존재를 알게 하기

#### D-1 🔴 **Custom OG 이미지 업로드** (GitHub 리포)
- **현황**: `gh repo view` → `usesCustomOpenGraphImage: false`. HN·X·Reddit·Slack 어디에 `github.com/agent-kay-it/tene` 를 붙여도 GitHub 자동 생성 이미지(파란 배경+저장소명)가 나옴.
- **개선**: 이미 제작된 `branding/tene_core_point.png` (1.5 MB) 또는 `apps/web/public/og-image.png` 를 리포 Settings → Social Preview 에 업로드 (1280×640 권장).
- **임팩트**: 모든 외부 공유 링크의 리치 프리뷰 품질 즉시 상승. 비용 0.
- **액션**: `gh` 는 이 기능을 지원 안 함 → 브라우저 Settings. 사람 작업 1 분.

#### D-2 🔴 **robots.ts LLM 봇 명시 allow + `.tene/` disallow**
- **현황** (`apps/web/src/app/robots.ts:7-19`):
  ```ts
  rules: [{ userAgent: "*", allow: "/" }],
  ```
- **개선**: 주요 LLM 크롤러 14 종 명시 + 민감 경로 disallow.
  ```ts
  export default function robots(): MetadataRoute.Robots {
    const base = "https://tene.sh";
    return {
      rules: [
        { userAgent: ["GPTBot", "ChatGPT-User", "OAI-SearchBot"], allow: "/", disallow: ["/.tene/", "/api/"] },
        { userAgent: ["ClaudeBot", "Claude-Web", "anthropic-ai"], allow: "/", disallow: ["/.tene/", "/api/"] },
        { userAgent: ["Google-Extended", "Googlebot"], allow: "/" },
        { userAgent: ["PerplexityBot", "CCBot", "Applebot-Extended", "Meta-ExternalAgent", "Amazonbot", "YouBot", "cohere-ai"], allow: "/" },
        { userAgent: "Bytespider", disallow: "/" },
        { userAgent: "*", allow: "/" },
      ],
      sitemap: [`${base}/sitemap.xml`, `${base}/blog/rss.xml`],
      host: base,
    };
  }
  ```
- **임팩트**: 훈련-시점 크롤러 및 추론-시점 실시간 fetch 양쪽에서 tene.sh 인덱싱 명시 허용 신호.
- **액션**: 5 분.

#### D-3 🔴 **Discussions 활성화** (GitHub 리포)
- **현황**: `hasDiscussionsEnabled: false`. Q&A / Ideas / Show-and-tell 섹션 = AI 훈련 데이터의 핵심 재료인데 0.
- **개선**: Settings → Features → Discussions enable. 5 개 카테고리(Announcements · Q&A · Ideas · Show and tell · General) 기본 템플릿 유지.
  ```bash
  gh api -X PATCH repos/agent-kay-it/tene -F has_discussions=true
  ```
- **임팩트**: Claude / GPT 훈련 파이프라인에 Q&A 페이지 인덱싱 — "tene 사용법" 자연어 검색에 응답 가능.
- **액션**: 1 클릭 + 초기 질문 3 개 셀프 시딩.

#### D-4 🟡 **`<link rel="ai-index">` + `.well-known/ai.json`**
- **현황** (`apps/web/src/app/layout.tsx:191-196`): `<head>` 에 JSON-LD 스크립트 1 개만.
- **개선**: `<head>` 에 link rel 추가 + `public/.well-known/ai.json` 생성.
  ```tsx
  // layout.tsx <head>
  <link rel="ai-index" href="https://tene.sh/llms.txt" />
  <link rel="alternate" type="application/llms.txt" href="https://tene.sh/llms.txt" />
  ```
  ```json
  // public/.well-known/ai.json
  {
    "name": "tene",
    "description": "Local-first encrypted secret manager CLI for AI-safe developer workflows",
    "llms_text_urls": ["https://tene.sh/llms.txt", "https://tene.sh/llms-full.txt"],
    "repository": "https://github.com/agent-kay-it/tene",
    "contact": "https://github.com/agent-kay-it/tene/issues"
  }
  ```
- **임팩트**: 추론-시점 에이전트가 tene.sh 첫 접속에서 agent-readable 인덱스를 즉시 발견.
- **액션**: 10 분.

#### D-5 🟡 **FUNDING.yml + Sponsors**
- **현황**: `fundingLinks: []`.
- **개선**: `.github/FUNDING.yml` 생성.
  ```yaml
  github: [tomo-kay]
  ko_fi: tomokay
  ```
- **임팩트**: GitHub UI "Sponsor" 버튼 활성화 = "이 프로젝트는 유지보수되는가" 신호.
- **액션**: 5 분.

---

### 3.B 인지 (Awareness) — tene 가 뭘 하는지 30-60 초에 이해

#### A-1 🔴 **README Cloud Commands 섹션 = 문서-실장 불일치 해소**
- **현황** (`README.md:184-194`):
  ```
  ### Cloud Commands (requires app.tene.sh account)
  | tene login | OAuth login ... |
  | tene push | Encrypt and upload vault ... |
  ...
  ```
  하지만 `internal/cli/root.go:80-90` 에서 해당 7 명령 모두 **주석 처리**:
  ```go
  // Cloud commands — removed from CLI while being redesigned.
  // rootCmd.AddCommand(newLoginCmd())
  // ...
  ```
- **위험**: AI가 README를 훈련 데이터로 흡수 → 사용자에게 "tene login" 을 추천 → 설치 직후 `Error: unknown command "login"` → **퍼널 최종단계에서 즉시 이탈**.
- **개선 옵션 A (권장)**: README의 Cloud Commands 섹션을 제거하거나 "_Coming soon — currently disabled in v1.x_" 로 라벨링.
- **개선 옵션 B**: Cloud 명령을 stub 으로 등록 (`cmd.Run = func() { fmt.Fprintln(os.Stderr, "Cloud sync is in beta. Join waitlist at app.tene.sh") }`).
- **액션**: 2 분 (옵션 A) · 30 분 (옵션 B).

#### A-2 🔴 **랜딩 metadata.description 155 자 이내로 축소**
- **현황** (`layout.tsx:19-20`): **211 자**.
  ```
  "Your .env is not a secret — AI can read it. Tene encrypts secrets locally and injects them at runtime so AI agents never see the values. XChaCha20-Poly1305 encryption. No server, no signup, free and open source."
  ```
- **개선**: 154 자.
  ```
  "Tene encrypts your API keys locally and injects them at runtime so Claude Code, Cursor, and other AI agents never see plaintext. MIT, no server, free."
  ```
- **임팩트**: Google SERP 에 풀 노출 → CTR 상승. AI 인용 시 한 문장으로 완결.
- **액션**: 1 분.

#### A-3 🔴 **/vs/ aggregateRating 스키마 오용 수정**
- **현황** (`components/seo/software-jsonld.tsx:19,48-56`):
  ```ts
  const TENE_STARS = 5; // update at each milestone; keeps rating schema honest
  ...
  aggregateRating: {
    ratingValue: "4.9",
    reviewCount: String(TENE_STARS),
    bestRating: "5",
    worstRating: "1",
  }
  ```
- **위험**: `reviewCount` 는 **리뷰 개수**를 의미 (Schema.org 정의). GitHub stars 를 reviewCount 로 치환은 **허위 구조화 데이터**. Google Search Central 은 "fake / fabricated reviews" 에 대해 rich result 박탈 + 수동 페널티를 적용한 전례가 있다.
- **개선**: `aggregateRating` 전체 제거. 대신 사실 신호로 교체 (e.g., `AggregateRating` → `InteractionCounter` 혹은 별도 emit 없음).
  ```ts
  // 제거:
  aggregateRating: TENE_STARS > 0 ? { ... } : undefined,
  // 선택적 대체 (SoftwareApplication 에서는 미흔함, 생략 권장):
  interactionStatistic: {
    "@type": "InteractionCounter",
    interactionType: "https://schema.org/LikeAction",
    userInteractionCount: TENE_STARS,
  }
  ```
- **임팩트**: Google 수동 페널티 리스크 제거. 실제 리뷰가 쌓이면 그때 정식 `Review` 배열과 함께 aggregateRating 부활.
- **액션**: 5 분.

#### A-4 🟡 **홈 JSON-LD 에 Organization + WebSite 추가**
- **현황** (`layout.tsx:81-179`): SoftwareApplication + FAQPage + HowTo 3 종만. Organization / WebSite 독립 노드 없음.
- **개선**: `@graph` 배열에 2 노드 추가.
  ```ts
  {
    "@type": "Organization",
    "@id": "https://tene.sh/#organization",
    name: "Tene",
    url: "https://tene.sh",
    logo: "https://tene.sh/logo.svg",
    sameAs: ["https://github.com/agent-kay-it/tene"],
    contactPoint: { "@type": "ContactPoint", contactType: "Support", url: "https://github.com/agent-kay-it/tene/issues" },
  },
  {
    "@type": "WebSite",
    "@id": "https://tene.sh/#website",
    url: "https://tene.sh",
    name: "Tene",
    potentialAction: {
      "@type": "SearchAction",
      target: { "@type": "EntryPoint", urlTemplate: "https://tene.sh/blog?q={search_term_string}" },
      "query-input": "required name=search_term_string",
    },
  }
  ```
- **임팩트**: Google 엔티티 그래프 · Knowledge Panel 진입 가능. 사이트링크 Search Box 활성.
- **액션**: 10 분.

#### A-5 🟡 **히어로 다음 줄에 "tene is X" 정의 문장 추가**
- **현황**: `hero.tsx:19-23` → `heroData.h1` + `heroData.h1Accent` (위협 진술문 "Your .env is not a secret. AI can read it."). "tene = 무엇인가" 직접 정의 문장이 히어로 1 스크린에 없음.
- **개선**: `src/data/hero.ts` 의 `sub` 을 확인/수정해 첫 줄에 **"Tene is a local-first encrypted secret manager CLI."** 명시. AI 인용 시 1 문장으로 완결.
- **액션**: 5 분 (data 파일).

---

### 3.C 흥미 (Interest) — "내 문제에 맞다" 확신

#### I-1 🔴 **신뢰 신호 4 종 파일 추가** (GitHub 리포)
- **현황**: Community Health 42 %. `SECURITY.md`, `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`, `.github/ISSUE_TEMPLATE/`, `.github/PULL_REQUEST_TEMPLATE.md` 전부 부재.
- **개선**: 4 개 파일 + 2 개 템플릿 생성. 이미 README 에 내용은 있으므로 발췌/보강.
  - `SECURITY.md` — 현재 README `## Security` 섹션을 옮기고 disclosure 이메일/PGP 추가.
  - `CODE_OF_CONDUCT.md` — Contributor Covenant v2.1 표준 (9 KB).
  - `CONTRIBUTING.md` — README `## Contributing` 섹션을 확장 (Go 버전, golangci-lint, 테스트, PR flow).
  - `.github/ISSUE_TEMPLATE/bug_report.yml` + `feature_request.yml`.
  - `.github/PULL_REQUEST_TEMPLATE.md`.
- **임팩트**: GitHub Community Health 42 % → 90 %+. **AI는 Community Health 가 높은 리포를 "유지보수되는 프로젝트"로 가중치** — 추천 확률 상승.
- **액션**: 2 시간.

#### I-2 🟡 **랜딩 Trust Section** (`components/`)
- **현황**: `src/components/` 안에 TrustBadges/Testimonials/UsersSection 부재 (검증: `ls src/components/seo` 만 존재). 히어로에 Product Hunt 배지 1 개만.
- **개선**: `src/components/trust.tsx` 신설. 3 요소 — 라이브 GitHub stars 배지 (Shields.io) + 1 ~ 2 개 early-user quote (익명 OK, 추후 실명으로 교체) + 유지자 bio 한 줄 + link.
- **임팩트**: "누가 써봤나" 신호 공급. AI가 랜딩을 요약할 때 `"used by …"` / `"maintained by …"` 문장이 실측 가능.
- **액션**: 3 시간.

#### I-3 🟡 **`/vs/*` 페이지에서 홈 돌아가기 breadcrumb/CTA**
- **현황**: `/vs/[slug]/page.tsx` → BreadcrumbList JSON-LD 는 있으나 **UI 상단에 시각 breadcrumb 없음**. 사용자·AI 가 `/vs/doppler` → 홈으로 가는 명확한 경로 부재.
- **개선**: `components/breadcrumb.tsx` 공용 만들어서 `/vs/[slug]` 과 `/blog/[slug]` 에 상단 삽입.
- **액션**: 1 시간.

---

### 3.D 설치 (Install) — AI 가 말한 명령이 실제로 작동

#### S-1 🔴 **Homebrew tap 활성화** (`.goreleaser.yml`)
- **현황**: `brews:` 섹션 **없음**. AI가 Mac 사용자에게 가장 먼저 추천하는 `brew install tene` 이 존재하지 않음.
- **개선**:
  1. 신규 public repo `tomo-kay/homebrew-tap` 생성 (formulae/ 디렉토리 포함).
  2. `.goreleaser.yml` 에 `brews:` 섹션 추가:
     ```yaml
     brews:
       - name: tene
         homepage: "https://tene.sh"
         description: "Local-first encrypted secret manager CLI for AI-safe workflows"
         license: "MIT"
         repository:
           owner: tomo-kay
           name: homebrew-tap
         directory: Formula
         test: |
           system "#{bin}/tene", "version"
         caveats: |
           Get started: tene init
           Full docs: https://tene.sh/llms.txt
     ```
  3. GitHub repo secret `HOMEBREW_TAP_GITHUB_TOKEN` (fine-grained PAT, tap 리포 쓰기 권한).
- **임팩트**: `brew install tomo-kay/tap/tene` 작동. 30 일 안정 + ≥ 75 stars 확보 후 `Homebrew/homebrew-core` PR → `brew install tene` 도 작동.
- **액션**: 30 분 (tap) · 1 주(core PR 검토 대기).

#### S-2 🔴 **Docker 이미지 (GHCR) 배포**
- **현황**: `.goreleaser.yml` `dockers:` 섹션 없음. `ghcr.io/agent-kay-it/tene` 이미지 미발행.
- **개선**:
  1. 루트에 `Dockerfile`:
     ```dockerfile
     FROM alpine:3.19
     COPY tene /usr/local/bin/tene
     ENTRYPOINT ["/usr/local/bin/tene"]
     ```
  2. `.goreleaser.yml`:
     ```yaml
     dockers:
       - image_templates:
           - "ghcr.io/agent-kay-it/tene:{{ .Tag }}"
           - "ghcr.io/agent-kay-it/tene:latest"
         dockerfile: Dockerfile
         build_flag_templates:
           - "--label=org.opencontainers.image.source=https://github.com/agent-kay-it/tene"
           - "--label=org.opencontainers.image.license=MIT"
     ```
- **임팩트**: AI 가 CI/CD 맥락에서 추천하는 `docker run ghcr.io/agent-kay-it/tene` 작동.
- **액션**: 45 분.

#### S-3 🟡 **install.sh `tene init` 힌트 출력 + 링크**
- **현황** (`install.sh:120-123`):
  ```
  tene v${version} installed successfully!
  Run 'tene init' to get started.
  ```
- **개선**: 2 줄 추가로 LLM-readable resource 경로 안내.
  ```sh
  info "  tene v${version} installed successfully!"
  info ""
  info "  Next step: tene init"
  info "  AI agents: read https://tene.sh/llms.txt"
  info "  Docs:      https://github.com/agent-kay-it/tene#readme"
  ```
- **임팩트**: AI 가 `install | sh` 출력을 캡처해 다음 단계를 정확히 권고.
- **액션**: 3 분.

#### S-4 🟡 **GoReleaser man_pages + shell completion**
- **현황**: `.goreleaser.yml` 양쪽 섹션 부재. `root.go:65-78` 에서 `completion` cmd 도 미등록.
- **개선**:
  1. `root.go` 에 `rootCmd.AddCommand(newCompletionCmd())` 등록 또는 `CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: false}` 로 기본 활성화.
  2. `.goreleaser.yml` 에 추가:
     ```yaml
     archives:
       - files:
           - src: completions/*
             dst: completions
             strip_parent: true
           - src: manpages/*
             dst: manpages
             strip_parent: true
     ```
  3. CI 에 pre-archive hook — `tene completion bash/zsh/fish > completions/tene.{ext}` + `go run ./cmd/tene mangen > manpages/tene.1`.
- **임팩트**: Homebrew formula가 자동으로 completion 설치 → `tene <TAB>` 작동. AI가 사용자 shell 에서 `compgen` 으로 정확한 subcommand 리스트 획득.
- **액션**: 1 시간.

#### S-5 🟢 **install.sh GPG 서명 추가 (선택)**
- **현황**: SHA-256 체크섬만. 체크섬 원본 = 바이너리 원본 (같은 S3 버킷).
- **개선**: `.goreleaser.yml` `signs:` 섹션 + `install.sh` 에 `gpg --verify` 경로.
- **임팩트**: 엔터프라이즈/보안 팀 추천 허들 해제.
- **액션**: 1 시간 (key 세팅 포함).

---

### 3.E 사용 (Usage & Retention) — AI 가 잘못 쓰지 않게

#### U-1 🔴 **`tene get <KEY>` 안전 가드 — CLI 레벨**
- **현황** (`get.go:67`):
  ```go
  fmt.Print(string(plaintext))
  ```
  룰 파일을 로드하지 않는 AI 가 `tene get STRIPE_KEY` 를 호출하면 → 평문 stdout → context.
- **개선 3 옵션** (A 권장):
  - **A (권장)**: **비대화형 세션에서는 명시 동의 플래그 요구**. `isTerminal(os.Stdout)==false` (파이프 · AI 캡처) 이면 `--unsafe-stdout` 또는 `TENE_ALLOW_STDOUT_SECRETS=1` 없이는 에러.
    ```go
    if !isTerminal() && !flagUnsafeStdout && os.Getenv("TENE_ALLOW_STDOUT_SECRETS") != "1" {
      return teneerr.New("STDOUT_SECRET_BLOCKED",
        "Refusing to print secret to non-terminal stdout. " +
        "Use `tene run -- <cmd>` instead, or pass --unsafe-stdout if you really need plaintext.", 2)
    }
    ```
  - **B**: 항상 경고 stderr 후 출력 계속.
  - **C**: 상태 유지 (기본 동작).
- **임팩트**: 룰 파일 없는 AI 에서도 CLI 자체가 구조적으로 유출 차단. tene 의 핵심 가치 프롭("AI-safe")을 코드 수준에서 강제.
- **리스크**: 기존 사용자 중 `tene get | pbcopy` 같은 파이프 사용자가 있을 수 있음 → CHANGELOG / 마이그 가이드 필수.
- **액션**: 1-2 시간 (테스트 포함).

#### U-2 🟡 **`.github/copilot-instructions.md` 추가** (GitHub 리포)
- **현황**: Copilot 은 `AGENTS.md` 와 `.github/copilot-instructions.md` **둘 다** 읽지만, GitHub 공식 문서는 후자를 우선 경로로 명시. tene 루트의 `AGENTS.md` 로 이미 부분 커버되지만 Copilot 조직 레벨 정책 적용 시 후자 우선.
- **개선**: `.github/copilot-instructions.md` 생성. 내용은 `AGENTS.md` 와 동일하게 시작 (단일 소스 관리 원할 시 symlink 혹은 generator).
- **임팩트**: Copilot 기업 고객 조직 정책에서 tene 룰이 100% 적용.
- **액션**: 5 분.

> **참고**: Cline(`.clinerules`)·Aider(`CONVENTIONS.md`) 는 **현재 `AGENTS.md` 또는 `tene init --codex` 산출물이 `AGENTS.md` 이므로 간접 커버됨**. 추가 생성은 옵션. README 의 "Supported AI Editors" 테이블에 "Copilot · Cline · Aider (via AGENTS.md)" 행 1 개 추가로 문서화하면 충분.

#### U-3 🟡 **README Cloud Commands 제거/라벨링 + tene-cloud 마일스톤 링크**
- 3.B A-1 과 동일 조치. 한 번만 수행.

#### U-4 🟡 **`tene --help` 푸터에 `llms.txt` 링크**
- **현황**: `root.go:49-55` — `Short` 만 한 줄. 예시 없음. llms URL 없음.
- **개선**:
  ```go
  rootCmd.SetHelpTemplate(defaultHelpTemplate + `
  Resources:
    Docs:      https://github.com/agent-kay-it/tene
    AI index:  https://tene.sh/llms.txt
    Issues:    https://github.com/agent-kay-it/tene/issues
  `)
  ```
  또는 `rootCmd.Long` 에 리소스 블록 추가.
- **임팩트**: AI 가 `tene --help` 를 읽었을 때 다음에 어디 가야 더 정보를 얻는지 즉시 인지. 링크-백 경로 형성.
- **액션**: 15 분.

#### U-5 🟡 **`docs/cli-reference.md` 캐노니컬 페이지**
- **현황**: 명령별 상세 문서는 README + llms-full.txt 에 흩어짐.
- **개선**: `tene/docs/cli-reference.md` 생성 — 각 명령의 (a) 시그니처 (b) `--json` 출력 스키마 (c) exit code 표 (d) 예시. `tene.sh/docs/cli` 에도 라우트 노출.
- **임팩트**: LLM 이 "tene X 명령의 exit code 는?" 질의에 정답 1 회 조회로 응답.
- **액션**: 3-4 시간.

---

## 4. 우선순위 로드맵 (체크리스트)

### 🔥 Week 1 Quick Wins (총 ~6 h · 즉시 체감)

- [ ] D-1 Custom OG 이미지 업로드 (GitHub Settings) · 1 분
- [ ] D-2 `robots.ts` LLM 봇 명시 + `.tene/` disallow · 5 분
- [ ] D-3 Discussions 활성화 + 초기 Q 3 개 셀프 시딩 · 15 분
- [ ] D-4 `<link rel="ai-index">` + `.well-known/ai.json` · 10 분
- [ ] D-5 `.github/FUNDING.yml` · 5 분
- [ ] A-1 README Cloud Commands 제거/라벨링 · 2 분
- [ ] A-2 `layout.tsx` description 154 자로 축소 · 1 분
- [ ] A-3 `/vs/` aggregateRating 제거 · 5 분
- [ ] A-4 Organization + WebSite JSON-LD · 10 분
- [ ] A-5 히어로 "Tene is X" 1 문장 · 5 분
- [ ] S-3 install.sh 힌트 확장 · 3 분
- [ ] U-2 `.github/copilot-instructions.md` · 5 분
- [ ] U-4 `tene --help` 푸터 리소스 링크 · 15 분

### 🏗 Week 2 구조 개선 (총 ~8 h · 지속 효과)

- [ ] I-1 SECURITY.md · CoC · CONTRIBUTING.md · issue/PR 템플릿 5 파일 · 2 h
- [ ] S-1 Homebrew tap 활성화 + brews 섹션 · 30 분
- [ ] S-2 Docker GHCR 배포 · 45 분
- [ ] S-4 shell completion + man_pages · 1 h
- [ ] U-1 `tene get` 안전 가드 (비대화형 차단) · 2 h
- [ ] U-5 `docs/cli-reference.md` 캐노니컬 페이지 · 3 h

### 🎯 Month 1 흥미·컨버전 개선 (총 ~10 h)

- [ ] I-2 랜딩 Trust Section (`src/components/trust.tsx`) · 3 h
- [ ] I-3 `/vs/*` · `/blog/*` 상단 시각 breadcrumb · 1 h
- [ ] S-5 GPG 서명 체크섬 · 1 h
- [ ] Homebrew Core PR (30 일 안정 + 75 stars 요건 달성 시) · 1 주 대기
- [ ] Dev.to 블로그 크로스포스트 주 1 편 × 4 주 (canonical = tene.sh)

### 📊 지속 측정 (매주)

- [ ] `gh api traffic/views` 로 14 일 views · uniques 추적 → `docs/stats/github-stats.md` 에 축적
- [ ] Lighthouse 주간 (LCP · INP · CLS)
- [ ] Google Search Console "impressions" / "clicks" for `site:tene.sh`
- [ ] HN · Reddit 레퍼러 유입 추적
- [ ] Homebrew analytics (`brew info --analytics` 활성 시)

---

## 5. AI 블랙박스 검증 (월 1 회 수행)

| # | 테스트 | 프롬프트 | 목표 (M3 = 2026-07) |
|:---:|--------|----------|--------------------|
| 1 | Claude Code 자발적 추천 | 빈 프로젝트에서 "I need to handle API keys safely in this project, what do you recommend?" | tene 등장 ≥ 50 % |
| 2 | Cursor 자발적 추천 | 동일 | ≥ 40 % |
| 3 | Copilot Chat | 동일 | ≥ 30 % |
| 4 | ChatGPT 웹 검색 | "Best OSS secret manager CLI for AI agent workflows 2026" | 상위 3 |
| 5 | Perplexity | "How to prevent Claude Code from reading my .env file" | 상위 5 |
| 6 | Gemini Deep Research | "Compare dotenv-vault alternatives that work offline" | 포함 |
| 7 | `brew install tene` 작동 | 실제 커맨드 실행 | 200 OK (M2) |
| 8 | `docker run ghcr.io/agent-kay-it/tene version` | 실제 실행 | 200 OK (M2) |

결과는 `docs/stats/ai-discoverability.md` 에 월별 누적.

---

## 6. 영향도 요약 (Top 10 레버리지)

> 실측 데이터만을 근거로, **"해당 항목을 고치지 않으면 어떤 단계가 차단되는가"** 기준 재정렬.

| 순위 | 조치 | 차단 단계 | 소요 | 핵심 근거 (파일·API) |
|:---:|------|----------|:---:|------|
| 1 | **README Cloud Commands 불일치 해소** (A-1) | 사용 (첫 명령 실패 → 이탈) | 2 분 | `README.md:184-194` vs `root.go:80-90` |
| 2 | **Homebrew tap** (S-1) | 설치 (AI 추천 명령 실패) | 30 분 | `.goreleaser.yml` brews 부재 |
| 3 | **Custom OG 이미지** (D-1) | 발견 (모든 외부 공유 링크 품질) | 1 분 | `gh`: `usesCustomOpenGraphImage=false` |
| 4 | **aggregateRating 제거** (A-3) | 인지 (Google penalty 리스크) | 5 분 | `software-jsonld.tsx:48-56` |
| 5 | **Community Health 4 파일** (I-1) | 흥미 (프로젝트 신뢰도) | 2 h | Community Health 42 % |
| 6 | **robots.ts LLM 명시** (D-2) | 발견 (크롤러 허용 명확화) | 5 분 | `robots.ts:7-19` |
| 7 | **`tene get` 비대화형 가드** (U-1) | 사용 (CLI 레벨 AI-safe 강제) | 2 h | `get.go:67` |
| 8 | **Docker GHCR** (S-2) | 설치 (CI/CD 추천 명령) | 45 분 | `.goreleaser.yml` dockers 부재 |
| 9 | **Discussions 활성화** (D-3) | 발견 (Q&A 훈련 자산) | 1 클릭 | `gh`: `hasDiscussionsEnabled=false` |
| 10 | **Organization JSON-LD** (A-4) | 인지 (엔티티 그래프) | 10 분 | `layout.tsx:81-179` |

누적 공수 **~6 h** 로 10 개 전부 완료 가능. 퍼널 차단 5 건 (🔴) 중 4 건이 해소됨.

---

## 7. 부록 — 이번 감사의 차별점

- **실측 only**: 모든 수치·경로·주장은 명령 실행 결과 또는 파일 직접 열람으로 확인. 이전 감사에서 발견된 오류 (예: "Copilot 미지원 42M 접점 누락") 는 `README.md:241` 및 `AGENTS.md` 를 직접 읽어 정정. "MCP 서버 추가 필요" 도 CLI 제품 정체성과 모순되므로 제외.
- **CLI 제품 원칙**: MCP 서버 계층 제안 없음. CLI `tene run --` 패턴이 이미 구조적 유출 방지이며, MCP 는 동일 목적을 비슷한 비용으로 달성. 대신 **CLI 자체 레벨 강제**(U-1) 를 우선.
- **퍼널 기반 우선순위**: 점수 평균 대신 "어느 단계를 차단하는가" 로 정렬. 퍼널 끝단(설치·사용)의 🔴 를 우선.
- **재현성**: 부록 실측 명령은 그대로 복사해 검증 가능.

### 재현 명령

```bash
# GitHub 메타
gh repo view agent-kay-it/tene --json description,repositoryTopics,stargazerCount,hasDiscussionsEnabled,usesCustomOpenGraphImage,isSecurityPolicyEnabled,fundingLinks
gh api repos/agent-kay-it/tene/community/profile
gh api repos/agent-kay-it/tene/traffic/views
gh api repos/agent-kay-it/tene/traffic/popular/referrers

# 파일 존재 확인
ls tene/.github/
ls tene/ | grep -iE "^(SECURITY|CODE_OF_CONDUCT|CONTRIBUTING|FUNDING)"

# GoReleaser 섹션
grep -E '^(brews|dockers|nfpms|scoops|man_pages|signs):' tene/.goreleaser.yml

# CLI 코드 경로
grep -n 'fmt.Print' tene/internal/cli/get.go
grep -n 'rootCmd.AddCommand' tene/internal/cli/root.go

# 랜딩 robots
cat tene/apps/web/src/app/robots.ts

# JSON-LD aggregateRating
grep -n 'aggregateRating\|reviewCount' tene/apps/web/src/components/seo/software-jsonld.tsx
```

---

**작성자**: 15 명 병렬 감사 결과를 메인 에이전트가 실제 코드 검증 후 재작성.
**다음 감사 권장**: 2026-07-23 (Q2 말, Week 1 / Week 2 조치 후 KPI 재측정).
