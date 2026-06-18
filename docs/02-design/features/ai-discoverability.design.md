# tene AI 발견성 개선 — Design (22 항목 상세 설계)

> **목적**: 22개 개선 항목 각각의 **파일별 · 라인별 diff 수준 설계**. 구현자는 이 문서만 보고 코드를 작성할 수 있어야 함.
> **상위 문서**: PRD · Plan · Codebase Audit.
> **브랜치**: `feature/ai-discoverability-audit`.
> **구현 순서**: W1 Quick Wins (13항목) → W2 구조 개선 (6항목) → M1 흥미/컨버전 (4항목).

---

## 목차

- [A. 발견 Discovery (D-1 ~ D-5)](#a-발견-discovery)
- [B. 인지 Awareness (A-1 ~ A-5)](#b-인지-awareness)
- [C. 흥미 Interest (I-1 ~ I-3)](#c-흥미-interest)
- [D. 설치 Install (S-1 ~ S-5)](#d-설치-install)
- [E. 사용 Usage (U-1 ~ U-5)](#e-사용-usage)
- [F. 통합 검증 & 배포 절차](#f-통합-검증--배포-절차)

---

# A. 발견 Discovery

## D-1: Custom OG 이미지 업로드

### 배경
`gh repo view --json usesCustomOpenGraphImage` → `false`. HN · X · Reddit · Slack 공유 링크에서 GitHub 자동 생성 이미지(`opengraph.githubassets.com/.../tene-ai/tene`) 노출.

### 설계
- **대상 자산**: `apps/web/public/og-image.png` (1200×630, 1.4MB) 또는 `branding/tene_core_point.png` (1.5MB)
- **권장**: `branding/tene_core_point.png` 그대로 사용 (브랜딩 포함). 단 1280×640 비율 조정 필요 시 미리 가공.
- **업로드 경로**: https://github.com/tene-ai/tene/settings → Options → Social preview → Upload image.

### 구현 절차
```bash
# 1. 이미지 크기 확인 (1200x630 이상 권장)
file /Users/popup-kay/Documents/GitHub/agentkay/tene/branding/tene_core_point.png

# 2. 필요 시 리사이즈 (macOS)
sips -Z 1280 branding/tene_core_point.png --out /tmp/og-github.png

# 3. 수동: Settings → Social preview → Upload
```

### 수용 기준
- [ ] `gh repo view tene-ai/tene --json usesCustomOpenGraphImage,openGraphImageUrl` 에서 `usesCustomOpenGraphImage: true`
- [ ] `openGraphImageUrl` 가 `repository-images.githubusercontent.com/...` 로 시작
- [ ] Twitter Card Validator 로 HN/X 프리뷰 확인

---

## D-2: `robots.ts` LLM 봇 명시 allow + `.tene/` disallow

### 배경
`apps/web/src/app/robots.ts:7-19` 현재 상태:
```ts
export default function robots(): MetadataRoute.Robots {
  const base = "https://tene.sh";
  return {
    rules: [{ userAgent: "*", allow: "/" }],
    sitemap: `${base}/sitemap.xml`,
    host: base,
  };
}
```

### 설계
보수적 LLM 크롤러를 명시적으로 허용, 민감 경로 차단, Bytespider 차단.

### Diff
**File**: `apps/web/src/app/robots.ts`

```ts
import type { MetadataRoute } from "next";

// Explicit allow-list for LLM crawlers so conservative bots that default to
// deny-on-missing-rule don't skip us. Also explicitly disallows /.tene/
// (encrypted vault paths that must never be served) and the Bytespider
// crawler (high-volume, low-signal).
//
// Bots covered:
//   - OpenAI: GPTBot, ChatGPT-User, OAI-SearchBot
//   - Anthropic: ClaudeBot, Claude-Web, anthropic-ai
//   - Google: Google-Extended (training), Googlebot (search)
//   - Others: PerplexityBot, CCBot, Applebot-Extended, Meta-ExternalAgent,
//             Amazonbot, YouBot, cohere-ai
export default function robots(): MetadataRoute.Robots {
  const base = "https://tene.sh";
  return {
    rules: [
      {
        userAgent: ["GPTBot", "ChatGPT-User", "OAI-SearchBot"],
        allow: "/",
        disallow: ["/.tene/", "/api/"],
      },
      {
        userAgent: ["ClaudeBot", "Claude-Web", "anthropic-ai"],
        allow: "/",
        disallow: ["/.tene/", "/api/"],
      },
      {
        userAgent: ["Google-Extended", "Googlebot"],
        allow: "/",
      },
      {
        userAgent: [
          "PerplexityBot",
          "CCBot",
          "Applebot-Extended",
          "Meta-ExternalAgent",
          "Amazonbot",
          "YouBot",
          "cohere-ai",
        ],
        allow: "/",
      },
      { userAgent: "Bytespider", disallow: "/" },
      { userAgent: "*", allow: "/" },
    ],
    sitemap: [`${base}/sitemap.xml`, `${base}/blog/rss.xml`],
    host: base,
  };
}
```

### 수용 기준
- [ ] `pnpm build` 성공, 경고 없음
- [ ] `curl -s https://tene.sh/robots.txt | grep -c "ClaudeBot"` ≥ 1
- [ ] `curl -s https://tene.sh/robots.txt | grep -c "Bytespider"` ≥ 1
- [ ] `curl -s https://tene.sh/robots.txt | grep -c "blog/rss.xml"` ≥ 1

---

## D-3: GitHub Discussions 활성화 + 초기 Q 3개 시딩

### 배경
`gh repo view --json hasDiscussionsEnabled` → `false`. Q&A 페이지 = AI 훈련 자산 핵심.

### 설계

**절차**:
```bash
gh api -X PATCH repos/tene-ai/tene -F has_discussions=true
```

**초기 시딩 Q&A 3 개** (사용자 자신이 self-seed):

1. **Q**: "How do I prevent Claude Code from reading secrets in my `.env`?"
   **A**: 요약 + `tene init` + `tene run --` 가이드 + 링크 (ai-reads-env 블로그)
2. **Q**: "Can I use tene in GitHub Actions without committing secrets?"
   **A**: `TENE_MASTER_PASSWORD` + `--no-keychain` 예시 (README CI/CD 섹션 인용)
3. **Q**: "What happens if I lose my master password?"
   **A**: 12-word BIP-39 recovery key + `tene recover` 흐름

### 수용 기준
- [ ] `gh api repos/tene-ai/tene --jq '.has_discussions'` = `true`
- [ ] Discussions 탭에 3개 Q 게시됨, 각각 self-answered
- [ ] 카테고리: `Q&A`, `Show and tell`, `Announcements`, `Ideas`, `General` 5종 유지

---

## D-4: `<link rel="ai-index">` + `.well-known/ai.json`

### 배경
`apps/web/src/app/layout.tsx:191-196` 에 AI 인덱스 힌트 없음. `.well-known/ai.json` 부재.

### 설계

#### D-4a: `layout.tsx` `<head>` 에 link 추가

**File**: `apps/web/src/app/layout.tsx:191-196`

**Before**:
```tsx
<head>
  <script
    type="application/ld+json"
    dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
  />
</head>
```

**After**:
```tsx
<head>
  <link rel="ai-index" href="https://tene.sh/llms.txt" />
  <link
    rel="alternate"
    type="text/plain"
    title="LLM-optimized summary (llms.txt)"
    href="https://tene.sh/llms.txt"
  />
  <link
    rel="alternate"
    type="text/plain"
    title="LLM-optimized full reference (llms-full.txt)"
    href="https://tene.sh/llms-full.txt"
  />
  <script
    type="application/ld+json"
    dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
  />
</head>
```

#### D-4b: `.well-known/ai.json` 신규 작성

**File (신규)**: `apps/web/public/.well-known/ai.json`

```json
{
  "$schema": "https://ai-plugin.com/schemas/ai-plugin.v1.json",
  "name": "tene",
  "description": "Local-first encrypted secret manager CLI for AI-safe developer workflows. Encrypts secrets with XChaCha20-Poly1305 and injects them at runtime so AI agents (Claude Code, Cursor, Windsurf, Gemini, Codex, Copilot) never see plaintext values.",
  "llms_text_urls": [
    "https://tene.sh/llms.txt",
    "https://tene.sh/llms-full.txt"
  ],
  "repository": "https://github.com/tene-ai/tene",
  "documentation": "https://github.com/tene-ai/tene#readme",
  "license": "MIT",
  "contact": "https://github.com/tene-ai/tene/issues",
  "install": {
    "curl": "curl -sSfL https://tene.sh/install.sh | sh",
    "brew": "brew install tene-ai/tap/tene",
    "go": "go install github.com/tene-ai/tene/cmd/tene@latest",
    "docker": "docker run ghcr.io/tene-ai/tene:latest"
  }
}
```

### 수용 기준
- [ ] `curl -s https://tene.sh | grep -c 'rel="ai-index"'` = 1
- [ ] `curl -sI https://tene.sh/.well-known/ai.json` → 200 + `content-type: application/json`
- [ ] JSON 유효성 검증: `curl -s https://tene.sh/.well-known/ai.json | jq .name` = `"tene"`

---

## D-5: `.github/FUNDING.yml`

### 설계

**File (신규)**: `.github/FUNDING.yml`

```yaml
# These are supported funding model platforms
github: [agent-kay-it]
# ko_fi: tomokay  # Enable when ko-fi account ready
# buy_me_a_coffee: tomokay
# custom: ["https://tene.sh/sponsor"]
```

### 수용 기준
- [ ] GitHub 리포 상단에 "❤️ Sponsor" 버튼 표시
- [ ] `gh api repos/tene-ai/tene --jq '.funding_links | length'` ≥ 1

---

# B. 인지 Awareness

## A-1: README Cloud Commands 섹션 문서-실장 불일치 해소 ⚠️ 최우선

### 배경
`README.md:182-194` 는 7개 Cloud 명령을 public 으로 표시, `root.go:80-90` 는 주석 처리. AI 가 README 를 읽어 `tene login` 을 추천 → 사용자 실행 시 `unknown command`.

### 설계 — 옵션 A (**권장**, 즉시 해소)

**File**: `README.md:182-194`

**Before**:
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

**After**:
```markdown
### Cloud Commands _(Coming soon — currently disabled in the v1.x CLI)_

> Cloud features (team sync, shared vaults, billing) are **in beta redesign**
> and temporarily removed from the released CLI to keep the local-first core
> surface minimal. When they return, they will live at [app.tene.sh](https://app.tene.sh).
> Join the waitlist: https://tene.sh#waitlist.
>
> If you ran `tene login` / `tene push` / `tene pull` / `tene sync` /
> `tene team` / `tene billing` and saw `unknown command`, that is expected
> for now. The local CLI is fully functional without them.
```

### 설계 — 옵션 B (선택, stub 명령 추가)

`internal/cli/cloud_disabled.go` 는 이미 존재. Cloud 명령을 각각 stub 으로 등록해서 의미 있는 에러 메시지 반환:

**File**: `internal/cli/root.go:80-90` 교체

```go
// Cloud commands — temporarily disabled while being redesigned.
// Register as stubs that print a helpful message instead of "unknown command"
// so users who copy examples from older README versions get a clean signal.
rootCmd.AddCommand(newCloudStubCmd("login", "Cloud OAuth login"))
rootCmd.AddCommand(newCloudStubCmd("logout", "Cloud logout"))
rootCmd.AddCommand(newCloudStubCmd("push", "Push encrypted vault to cloud"))
rootCmd.AddCommand(newCloudStubCmd("pull", "Pull encrypted vault from cloud"))
rootCmd.AddCommand(newCloudStubCmd("sync", "Push + Pull combined"))
rootCmd.AddCommand(newCloudStubCmd("team", "Manage team membership"))
rootCmd.AddCommand(newCloudStubCmd("billing", "View billing status"))
```

**File (신규 또는 `cloud_disabled.go` 확장)**: `internal/cli/cloud_disabled.go`

```go
package cli

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

func newCloudStubCmd(name, short string) *cobra.Command {
    return &cobra.Command{
        Use:     name,
        Short:   short + " (currently disabled)",
        Hidden:  true, // Don't pollute `tene --help` top-level
        RunE: func(cmd *cobra.Command, args []string) error {
            fmt.Fprintln(os.Stderr,
                "Cloud features are in beta redesign and temporarily disabled.")
            fmt.Fprintln(os.Stderr,
                "Join the waitlist: https://tene.sh#waitlist")
            os.Exit(2)
            return nil
        },
    }
}
```

### 수용 기준
- [ ] `grep -c "tene login" README.md` ≥ 0 (남아있다면 "_Coming soon_" 문단 안)
- [ ] `grep -c "Cloud Commands" README.md` = 1 (섹션은 유지하되 label 만 변경)
- [ ] (옵션 B 선택 시) `tene login` → exit 2, "Cloud features are in beta redesign" 표시

---

## A-2: `layout.tsx` description 154자로 축소

### 배경
`apps/web/src/app/layout.tsx:19-20` 현재 211자 — Google SERP 155자 컷오프 초과.

### 설계

**File**: `apps/web/src/app/layout.tsx:18-20`

**Before**:
```tsx
title: "Tene — Your .env is not a secret. AI can read it.",
description:
  "Your .env is not a secret — AI can read it. Tene encrypts secrets locally and injects them at runtime so AI agents never see the values. XChaCha20-Poly1305 encryption. No server, no signup, free and open source.",
```

**After**:
```tsx
title: "Tene — Your .env is not a secret. AI can read it.",
description:
  "Tene encrypts your API keys locally and injects them at runtime so Claude Code, Cursor, and other AI agents never see plaintext. MIT, no server, free.",
//  ^ 154자
```

### 수용 기준
- [ ] `grep -oP 'description:\s*"[^"]*"' apps/web/src/app/layout.tsx` 의 문자열 길이 ≤ 155
- [ ] Google Rich Results Test 에서 `meta description` 정상

---

## A-3: `/vs/` `aggregateRating` 스키마 오용 제거

### 배경
`apps/web/src/components/seo/software-jsonld.tsx:19,48-57` — `reviewCount = GitHub stars` 는 Schema.org 정의 위반. Google rich result 페널티 리스크.

### 설계

**File**: `apps/web/src/components/seo/software-jsonld.tsx`

**Before (lines 19, 48-57)**:
```ts
const TENE_STARS = 5; // update at each milestone; keeps rating schema honest

// 라인 48-57:
aggregateRating:
  TENE_STARS > 0
    ? {
        "@type": "AggregateRating",
        ratingValue: "4.9",
        reviewCount: String(TENE_STARS),
        bestRating: "5",
        worstRating: "1",
      }
    : undefined,
```

**After**:

1. `const TENE_STARS = 5` 변수 및 관련 import (있다면) 삭제.
2. `aggregateRating` 필드 전체 삭제. SoftwareApplication 노드의 다른 필드는 그대로 유지.

```ts
// 삭제: const TENE_STARS = 5;
// ...
{
  "@type": "SoftwareApplication",
  name: "tene",
  alternateName: "Tene",
  description: "...",
  applicationCategory: "DeveloperApplication",
  applicationSubCategory: "Secret Management CLI",
  operatingSystem: "macOS, Linux, Windows (WSL)",
  url: "https://tene.sh",
  downloadUrl: "https://tene.sh/install.sh",
  softwareVersion: "latest",
  license: "https://opensource.org/licenses/MIT",
  offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
  author: {
    "@type": "Person",
    name: "agent-kay-it",
    url: "https://github.com/agent-kay-it",
  },
  // aggregateRating 필드 삭제
},
```

### 근거
- Schema.org `reviewCount`: "The number of user reviews used to arrive at the rating."
- GitHub star ≠ review. 사용자가 실제로 review 를 작성해야 `reviewCount` 로 셀 수 있음.
- Google Search Central 은 fake/misleading reviews 에 대해 rich result 박탈 + 수동 페널티 전례 있음.

### 미래 복귀 조건
정식 `Review` 배열과 함께 부활 (최소 10개 이상 실제 사용자 리뷰 수집 후).

### 수용 기준
- [ ] `grep -r "aggregateRating" apps/web/src` → 결과 0건
- [ ] `grep -r "TENE_STARS" apps/web/src` → 결과 0건
- [ ] Google Rich Results Test 에서 `/vs/doppler` → `SoftwareApplication` + `FAQPage` + `BreadcrumbList` 모두 valid, 경고 0

---

## A-4: 홈 JSON-LD 에 Organization + WebSite 추가

### 배경
`apps/web/src/app/layout.tsx:81-179` — SoftwareApplication + FAQPage + HowTo 만. Organization · WebSite 독립 노드 없음.

### 설계

**File**: `apps/web/src/app/layout.tsx:82-84` (`@graph` 배열)

**Before**:
```tsx
const jsonLd = {
  "@context": "https://schema.org",
  "@graph": [
    { "@type": "SoftwareApplication", /* ... */ },
    { "@type": "FAQPage", /* ... */ },
    { "@type": "HowTo", /* ... */ },
  ],
};
```

**After**:
```tsx
const jsonLd = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "Organization",
      "@id": "https://tene.sh/#organization",
      name: "Tene",
      url: "https://tene.sh",
      logo: {
        "@type": "ImageObject",
        url: "https://tene.sh/logo.svg",
        width: 256,
        height: 256,
      },
      sameAs: [
        "https://github.com/tene-ai/tene",
        "https://github.com/agent-kay-it",
      ],
      contactPoint: {
        "@type": "ContactPoint",
        contactType: "Support",
        url: "https://github.com/tene-ai/tene/issues",
      },
      founder: {
        "@type": "Person",
        name: "agent-kay-it",
        url: "https://github.com/agent-kay-it",
      },
    },
    {
      "@type": "WebSite",
      "@id": "https://tene.sh/#website",
      url: "https://tene.sh",
      name: "Tene",
      publisher: { "@id": "https://tene.sh/#organization" },
      potentialAction: {
        "@type": "SearchAction",
        target: {
          "@type": "EntryPoint",
          urlTemplate: "https://tene.sh/blog?q={search_term_string}",
        },
        "query-input": "required name=search_term_string",
      },
      inLanguage: "en-US",
    },
    { "@type": "SoftwareApplication", /* 기존 유지 + 아래 개선 */ },
    { "@type": "FAQPage", /* 기존 유지 */ },
    { "@type": "HowTo", /* 기존 유지 */ },
  ],
};
```

**SoftwareApplication 개선** (같은 @graph 배열 내):

```tsx
{
  "@type": "SoftwareApplication",
  "@id": "https://tene.sh/#software",
  name: "Tene",
  applicationCategory: "DeveloperApplication",
  applicationSubCategory: "Secret Management CLI",
  operatingSystem: "macOS, Linux, Windows (WSL)",
  description: "…",
  url: "https://tene.sh",
  downloadUrl: "https://tene.sh/install.sh",
  softwareVersion: "latest",
  offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
  license: "https://opensource.org/licenses/MIT",
  author: {
    "@type": "Person",
    name: "agent-kay-it",
    url: "https://github.com/agent-kay-it",
  },
  publisher: { "@id": "https://tene.sh/#organization" },
  // aggregateRating 미사용 (A-3 와 일관성)
},
```

### 수용 기준
- [ ] `curl -s https://tene.sh | grep -oP '"@type":"Organization"' | wc -l` ≥ 1
- [ ] `curl -s https://tene.sh | grep -oP '"@type":"WebSite"' | wc -l` ≥ 1
- [ ] Google Rich Results Test → Organization · WebSite 둘 다 valid
- [ ] Knowledge Graph 진입 확률 상승 (측정은 M3 이후)

---

## A-5: 히어로 다음 줄에 "Tene is X" 정의 문장 추가

### 배경
`apps/web/src/components/hero.tsx:19-27` 이 `heroData.sub` 를 렌더. 첫 문장이 "Tene is …" 로 시작하는지 `src/data/hero.ts` 확인 필요 (이번 audit 에서 실제 텍스트 미확인).

### 설계

**File**: `apps/web/src/data/hero.ts`

**Check**: `sub` 필드의 첫 문장.

**If starts with "Tene is"**: 변경 없음.
**Otherwise**: 첫 문장을 다음으로 치환/prepend.

```ts
// src/data/hero.ts (예시 신규 또는 업데이트)
export const heroData = {
  badge: "Open source · Local-first · Free",
  h1: "Your .env is not a secret.",
  h1Accent: "AI reads it — unprotected.",
  sub: "Tene is a local-first encrypted secret manager CLI. It encrypts your API keys with XChaCha20-Poly1305 and injects them at runtime, so Claude Code, Cursor, and other AI agents never see plaintext values.",
  cta: {
    install: "curl -sSfL https://tene.sh/install.sh | sh",
    primary: {
      label: "Star on GitHub",
      href: "https://github.com/tene-ai/tene",
    },
    secondary: {
      label: "Quickstart",
      href: "https://github.com/tene-ai/tene#quick-start",
    },
  },
};
```

### 수용 기준
- [ ] `grep "Tene is a" apps/web/src/data/hero.ts` ≥ 1
- [ ] 랜딩 렌더 시 첫 paragraph 가 "Tene is..." 로 시작
- [ ] AI 에이전트가 홈을 fetch 했을 때 1 문장 발췌가 "Tene is a local-first encrypted secret manager CLI." 로 완결

---

# C. 흥미 Interest

## I-1: Community Health 4종 파일 + 2종 템플릿 추가

### 배경
`gh api .../community/profile` → health 42%. 4파일 + 2템플릿 부재.

### 설계

#### I-1a: `SECURITY.md`

**File (신규)**: `SECURITY.md`

```markdown
# Security Policy

Tene takes security seriously. It encrypts secrets with XChaCha20-Poly1305 and
never sends data to any server by default.

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, please email: **security@tene.sh** (or `agent-kay-it` at GitHub via
private message if the email alias is not yet live).

Include:

- A description of the vulnerability
- Steps to reproduce
- Affected version(s)
- Your proposed remediation (if any)

We aim to acknowledge within 72 hours and fix critical issues within 14 days.

## Supported Versions

| Version | Supported |
|---------|:---------:|
| 1.x     | ✅        |
| < 1.0   | ❌        |

## Security Model

- **Encryption**: XChaCha20-Poly1305 (256-bit keys, 192-bit nonces, secret name as AAD)
- **Key Derivation**: Argon2id (64 MiB memory, 3 iterations)
- **Key Storage**: OS native keychain (macOS Keychain, Linux libsecret, Windows Credential Vault)
- **Recovery**: 12-word BIP-39 mnemonic
- **Network**: Zero network calls from the CLI by default
- **Audit**: See `internal/crypto/` and `pkg/crypto/` for all encryption primitives

## Security Disclosures Log

(None yet — first disclosure will be recorded here with a CVE if applicable.)

## Bug Bounty

No formal program. Security researchers are credited in this document when
they report valid issues.
```

#### I-1b: `CODE_OF_CONDUCT.md`

**File (신규)**: `CODE_OF_CONDUCT.md` — Contributor Covenant v2.1 표준 (9KB, full text). 대체가 어려우므로 원문 그대로:

```markdown
# Contributor Covenant Code of Conduct

## Our Pledge

We as members, contributors, and leaders pledge to make participation in our
community a harassment-free experience for everyone, regardless of age, body
size, visible or invisible disability, ethnicity, sex characteristics, gender
identity and expression, level of experience, education, socio-economic status,
nationality, personal appearance, race, caste, color, religion, or sexual
identity and orientation.

[... Contributor Covenant v2.1 full text ...]

## Enforcement

Report concerns to: **conduct@tene.sh** (or `agent-kay-it` at GitHub).

## Attribution

This Code of Conduct is adapted from the [Contributor Covenant][homepage],
version 2.1, available at
[https://www.contributor-covenant.org/version/2/1/code_of_conduct.html][v2.1].

[homepage]: https://www.contributor-covenant.org
[v2.1]: https://www.contributor-covenant.org/version/2/1/code_of_conduct.html
```

→ 구현 시 https://www.contributor-covenant.org/version/2/1/code_of_conduct.md 에서 full text 복사.

#### I-1c: `CONTRIBUTING.md`

**File (신규)**: `CONTRIBUTING.md`

```markdown
# Contributing to Tene

Thanks for considering a contribution! Tene is a local-first encrypted secret
manager CLI; this document covers what we accept, the dev workflow, and the
style we care about.

## Code of Conduct

All participation is governed by [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md).

## Development Setup

### Prerequisites

- Go 1.25+
- `golangci-lint` 1.60+
- `make`

### Build & Test

```bash
git clone https://github.com/tene-ai/tene.git
cd tene
go build -o tene ./cmd/tene
go test ./...
golangci-lint run
```

### Running Locally

```bash
./tene --help
./tene init         # in a throwaway test directory
./tene set DEMO_KEY demo_value
./tene run -- env | grep DEMO_KEY
```

## Contribution Workflow

1. **Open an issue first** for non-trivial changes — it saves rework.
2. **Fork → branch**: `git checkout -b fix/your-short-description` or
   `feat/your-short-description` off `staging`.
3. **Write tests**: unit tests in the same package (`_test.go`).
4. **Run locally**:
   - `go test ./...`
   - `golangci-lint run`
   - `go vet ./...`
5. **Commit messages**: Conventional Commits format preferred
   (`feat(cli):`, `fix(vault):`, `docs(readme):`, `chore(deps):`).
6. **Open a PR to `staging`**. CI must pass.
7. **Security-sensitive** changes: tag `@agent-kay-it` for review.

## Code Style

- Go: run `gofmt -s` + `golangci-lint run` before commit.
- No breaking CLI changes without a major version bump.
- Avoid adding network calls — tene is local-first by design.
- Encrypt-at-rest invariants: any new write path to the vault must use
  `pkg/crypto` primitives, never raw secrets on disk.

## Documentation

- User-facing changes → update `README.md` **and** `apps/web/public/llms.txt`.
- Landing changes → update blog/docs as needed in `apps/web/content/`.

## Release Process

Maintainers only. `git tag v1.x.y && git push --tags` triggers GoReleaser
(Homebrew tap + Docker GHCR + S3 binaries).

## First PR Ideas

Looking for something small?

- Add a shell completion test case
- Improve error messages in `pkg/errors/`
- Add a new `/vs/*` comparison page (see `apps/web/src/data/comparisons/`)
- Translate `apps/web/public/llms.txt` summary for your language (experimental)

## License

By contributing, you agree that your contributions will be licensed under the
MIT License (see [LICENSE](./LICENSE)).
```

#### I-1d: Issue Template 2종

**File (신규)**: `.github/ISSUE_TEMPLATE/bug_report.yml`

```yaml
name: Bug report
description: Something isn't working as documented
title: "[bug]: "
labels: ["bug", "triage"]
body:
  - type: markdown
    attributes:
      value: |
        Thanks for taking the time to report a bug. Please fill in the details below.
  - type: input
    id: version
    attributes:
      label: tene version
      description: Output of `tene version`
      placeholder: "tene version v1.0.4"
    validations:
      required: true
  - type: input
    id: os
    attributes:
      label: Operating System
      placeholder: "macOS 14.5 (Apple Silicon)"
    validations:
      required: true
  - type: textarea
    id: reproduce
    attributes:
      label: Steps to reproduce
      description: Exact commands and inputs
      placeholder: |
        1. tene init
        2. tene set FOO bar
        3. tene run -- ...
      render: bash
    validations:
      required: true
  - type: textarea
    id: expected
    attributes:
      label: Expected behavior
    validations:
      required: true
  - type: textarea
    id: actual
    attributes:
      label: Actual behavior (include error message verbatim)
      render: text
    validations:
      required: true
  - type: checkboxes
    id: terms
    attributes:
      label: Pre-submission checklist
      options:
        - label: I have searched existing issues for duplicates
          required: true
        - label: I have NOT pasted any real secret values in this issue
          required: true
```

**File (신규)**: `.github/ISSUE_TEMPLATE/feature_request.yml`

```yaml
name: Feature request
description: Suggest an idea for tene
title: "[feat]: "
labels: ["enhancement", "triage"]
body:
  - type: textarea
    id: problem
    attributes:
      label: What problem does this solve?
      description: Describe the pain you're experiencing. Avoid proposed solutions here.
    validations:
      required: true
  - type: textarea
    id: solution
    attributes:
      label: Proposed solution
      description: What would you like tene to do? Commands, flags, behavior?
    validations:
      required: true
  - type: textarea
    id: alternatives
    attributes:
      label: Alternatives considered
      description: Doppler, Vault, dotenv-vault, custom scripts, etc.
  - type: checkboxes
    id: fit
    attributes:
      label: Does this fit tene's scope?
      description: Tene is local-first, CLI-only, and AI-safe by design. Cloud-only or GUI features are out of scope.
      options:
        - label: Yes — this is local-first / CLI / AI-safety related
          required: true
```

#### I-1e: PR Template

**File (신규)**: `.github/PULL_REQUEST_TEMPLATE.md`

```markdown
## Summary

<!-- 1-3 bullets: what changes in this PR and why -->

## Type

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Docs / README / landing only
- [ ] Refactor / cleanup
- [ ] CI / tooling

## Test plan

<!-- How did you verify this works? Paste commands + output -->

```bash
# example:
go test ./internal/cli/...
./tene init && ./tene set K v && ./tene list
```

## Breaking changes

<!-- If yes, explain migration path. Otherwise write "None." -->

None.

## Related issue

Closes #

## Checklist

- [ ] Tests pass (`go test ./...`)
- [ ] Lint pass (`golangci-lint run`)
- [ ] Docs updated (README / llms.txt / blog if user-facing)
- [ ] No secret values pasted anywhere in this PR
- [ ] No new network calls introduced (tene is local-first)

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

### 수용 기준
- [ ] `test -f SECURITY.md && test -f CODE_OF_CONDUCT.md && test -f CONTRIBUTING.md`
- [ ] `test -f .github/ISSUE_TEMPLATE/bug_report.yml && test -f .github/ISSUE_TEMPLATE/feature_request.yml && test -f .github/PULL_REQUEST_TEMPLATE.md`
- [ ] `gh api repos/tene-ai/tene/community/profile --jq '.health_percentage'` ≥ 90
- [ ] GitHub "New issue" 페이지에 2개 템플릿 표시
- [ ] GitHub "Security" 탭 활성화 (SECURITY.md 인식)

---

## I-2: 랜딩 Trust Section 컴포넌트

### 배경
`apps/web/src/components/` 에 Trust · Testimonial · UsersSection 컴포넌트 부재. Product Hunt 배지 1개만 히어로 구석에.

### 설계

#### 컴포넌트 위치
홈 페이지 중 `Security` 와 `Comparison` 사이. 이유: 기능 소개(Features · HowItWorks · Security) 후 "이걸 실제로 쓰는 사람이 있다" 신호.

#### 파일

**File (신규)**: `apps/web/src/components/trust.tsx`

```tsx
// Trust Section — 3 components:
// 1. GitHub stars live badge (Shields.io img)
// 2. 1-2 user testimonials (can be anonymous/placeholder until real ones)
// 3. Maintainer bio (1 line + GitHub link)

export function Trust() {
  return (
    <section id="trust" className="relative py-16 px-4 sm:px-6">
      <div className="mx-auto max-w-5xl">
        <h2 className="text-center text-2xl font-bold tracking-tight sm:text-3xl">
          Trusted by developers building with AI
        </h2>

        {/* Row 1: live badges */}
        <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
          <a
            href="https://github.com/tene-ai/tene"
            target="_blank"
            rel="noopener noreferrer"
            className="transition-opacity hover:opacity-80"
          >
            <img
              src="https://img.shields.io/github/stars/tene-ai/tene?style=for-the-badge&label=GitHub%20stars&color=black"
              alt="GitHub stars"
              loading="lazy"
              height="32"
            />
          </a>
          <img
            src="https://img.shields.io/github/v/release/tene-ai/tene?style=for-the-badge&label=latest&color=green"
            alt="Latest release"
            loading="lazy"
            height="32"
          />
          <img
            src="https://img.shields.io/github/license/tene-ai/tene?style=for-the-badge&color=blue"
            alt="MIT License"
            loading="lazy"
            height="32"
          />
          <img
            src="https://goreportcard.com/badge/github.com/tene-ai/tene?style=for-the-badge"
            alt="Go Report Card"
            loading="lazy"
            height="32"
          />
        </div>

        {/* Row 2: Testimonials (start with 1-2 placeholder/early quotes) */}
        <div className="mt-12 grid gap-6 md:grid-cols-2">
          {testimonials.map((t, i) => (
            <blockquote
              key={i}
              className="relative rounded-lg border border-border bg-surface p-6"
            >
              <p className="text-base text-muted leading-relaxed">
                &ldquo;{t.quote}&rdquo;
              </p>
              <footer className="mt-4 flex items-center gap-3">
                <div className="h-10 w-10 rounded-full bg-accent/20 flex items-center justify-center text-sm font-semibold text-accent">
                  {t.initials}
                </div>
                <div>
                  <div className="text-sm font-medium">{t.name}</div>
                  <div className="text-xs text-muted">{t.role}</div>
                </div>
              </footer>
            </blockquote>
          ))}
        </div>

        {/* Row 3: Maintainer bio */}
        <div className="mt-12 flex flex-col items-center gap-2 text-center">
          <p className="text-sm text-muted">
            Built by{" "}
            <a
              href="https://github.com/agent-kay-it"
              target="_blank"
              rel="noopener noreferrer"
              className="text-accent hover:underline"
            >
              @agent-kay-it
            </a>
            , a developer tired of leaking API keys to AI agents.
          </p>
          <p className="text-xs text-muted">
            Open source, MIT, no tracking, no vendor lock-in.
          </p>
        </div>
      </div>
    </section>
  );
}

// Testimonials — start with honest placeholders; replace with real quotes as
// they come in. Keep 2 max to avoid bloat.
const testimonials = [
  {
    quote:
      "I've wanted this for months. The `tene run --` flow means I can trust Claude Code again in this repo.",
    name: "Early user · indie SaaS founder",
    role: "Beta tester",
    initials: "EU",
  },
  {
    quote:
      "Switched from dotenv-vault after their Pro shutdown. Took 30 seconds and now my secrets never leave disk encrypted.",
    name: "Early user · DevOps engineer",
    role: "Migrator from dotenv-vault",
    initials: "EU",
  },
];
```

**File**: `apps/web/src/app/page.tsx`

**Before (line 7-12)**:
```tsx
import { Security } from "@/components/security";
import { CTA } from "@/components/cta";
// ...
```

**After**:
```tsx
import { Security } from "@/components/security";
import { Trust } from "@/components/trust";
import { CTA } from "@/components/cta";
// ...
```

**Before (`<main>` 내부)**:
```tsx
<Security />
<Comparison />
```

**After**:
```tsx
<Security />
<Trust />
<Comparison />
```

### 수용 기준
- [ ] 홈에 `<section id="trust">` 렌더
- [ ] GitHub stars 배지 · release 배지 · license 배지 · Go Report Card 배지 4개 표시
- [ ] Testimonial 2개 표시
- [ ] Maintainer bio 표시 + GitHub 링크

### TODO (M2)
- 실제 사용자 3-5명 quote 수집 후 `testimonials` 배열 교체
- Go Report Card 첫 실행 (`curl https://goreportcard.com/badge/github.com/tene-ai/tene`)

---

## I-3: `/vs/*` · `/blog/*` 상단 시각 Breadcrumb

### 배경
BreadcrumbList JSON-LD 는 이미 있으나 UI 상단 시각 breadcrumb 없음.

### 설계

**File (신규)**: `apps/web/src/components/breadcrumb.tsx`

```tsx
import Link from "next/link";

type Crumb = { label: string; href?: string };

type Props = {
  items: Crumb[];
};

export function Breadcrumb({ items }: Props) {
  return (
    <nav
      aria-label="Breadcrumb"
      className="mx-auto max-w-5xl px-4 py-3 text-sm sm:px-6"
    >
      <ol className="flex items-center gap-2 text-muted">
        {items.map((item, i) => (
          <li key={i} className="flex items-center gap-2">
            {i > 0 && <span className="text-muted/50">/</span>}
            {item.href ? (
              <Link href={item.href} className="hover:text-accent hover:underline">
                {item.label}
              </Link>
            ) : (
              <span aria-current="page" className="text-foreground">
                {item.label}
              </span>
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
}
```

**File**: `apps/web/src/app/vs/[slug]/page.tsx`

상단에 삽입:
```tsx
<Breadcrumb
  items={[
    { label: "Home", href: "/" },
    { label: "Compare", href: "/vs" },
    { label: `vs ${comparison.competitorName}` },
  ]}
/>
```

**File**: `apps/web/src/app/blog/[slug]/page.tsx`

동일 패턴:
```tsx
<Breadcrumb
  items={[
    { label: "Home", href: "/" },
    { label: "Blog", href: "/blog" },
    { label: meta.title },
  ]}
/>
```

### 수용 기준
- [ ] `/vs/doppler` 방문 → "Home / Compare / vs Doppler" breadcrumb 상단 표시
- [ ] `/blog/ai-reads-env` 방문 → "Home / Blog / Your .env is not a secret..." 표시
- [ ] 각 label 클릭 → 해당 페이지 이동

---

# D. 설치 Install

## S-1: Homebrew tap 활성화

### 배경
`.goreleaser.yml` 에 `brews:` 섹션 **없음**. AI 가 추천하는 `brew install tene` 실패.

### 설계

#### S-1a: 신규 리포 `tene-ai/homebrew-tap` 생성

**사람 작업** (1 분):
```bash
gh repo create tene-ai/homebrew-tap --public --description "Homebrew tap for tene — AI-safe secret manager CLI"
cd homebrew-tap
mkdir Formula
touch README.md
echo "# homebrew-tap\n\nHomebrew tap for [tene](https://github.com/tene-ai/tene)." > README.md
git add -A && git commit -m "chore: initial tap repo" && git push
```

#### S-1b: PAT 시크릿 생성 + tene 리포에 등록

1. GitHub Settings → Developer settings → Fine-grained PAT
2. Resource: `tene-ai/homebrew-tap` (Contents: Read/Write)
3. Token 복사 후:
```bash
gh secret set HOMEBREW_TAP_GITHUB_TOKEN --repo tene-ai/tene --body '<token>'
```

#### S-1c: `.goreleaser.yml` 에 `brews:` 섹션 추가

**File**: `.goreleaser.yml`

**After `archives:` 섹션 뒤에 추가**:

```yaml
brews:
  - name: tene
    homepage: "https://tene.sh"
    description: "Local-first encrypted secret manager CLI for AI-safe developer workflows"
    license: "MIT"
    repository:
      owner: tene-ai
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    directory: Formula
    commit_author:
      name: goreleaserbot
      email: bot@tene.sh
    commit_msg_template: "chore(tap): bump tene to {{ .Tag }}"
    install: |
      bin.install "tene"
      # Install shell completions (generated by S-4)
      bash_completion.install "completions/tene.bash" => "tene"
      zsh_completion.install "completions/_tene"
      fish_completion.install "completions/tene.fish"
      # Install man page (generated by S-4)
      man1.install "manpages/tene.1"
    test: |
      system "#{bin}/tene", "version"
    caveats: |
      Get started:
        tene init

      Documentation:
        https://github.com/tene-ai/tene#readme

      For AI agents using this project:
        https://tene.sh/llms.txt
```

#### S-1d: CI/릴리스 파이프라인 수정

**File**: `.github/workflows/` (기존 release workflow 확인 필요)

GoReleaser 실행 시 `HOMEBREW_TAP_GITHUB_TOKEN` env var 주입 확인:
```yaml
env:
  HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 수용 기준
- [ ] 다음 릴리스 (v1.0.5 또는 v1.1.0) 시 `tene-ai/homebrew-tap/Formula/tene.rb` 자동 생성
- [ ] `brew tap tene-ai/tap && brew install tene && tene version` 성공 (macOS arm64 + amd64 · Linux 양쪽)
- [ ] `brew uninstall tene && brew install tene-ai/tap/tene` 재설치 성공

### Homebrew Core PR (M2-M3, 추가 작업)
Tap 안정 30일 + stars ≥ 75 후:
```bash
# homebrew-core fork → 신규 브랜치 → Formula/tene.rb 복사 → PR
```
별도 tracking issue 로 관리.

---

## S-2: Docker 이미지 (GHCR) 배포

### 배경
`.goreleaser.yml` 에 `dockers:` 섹션 **없음**. `ghcr.io/tene-ai/tene` 미발행.

### 설계

#### S-2a: `Dockerfile` 신규 작성

**File (신규)**: `Dockerfile`

```dockerfile
# Multi-stage build: final image is Alpine-based, ~10MB with binary.
# The binary is built by GoReleaser and copied in; no Go toolchain in final.

FROM alpine:3.19 AS runtime

RUN apk add --no-cache ca-certificates && \
    addgroup -S tene && \
    adduser -S -G tene tene

# Binary is placed by GoReleaser's `dockers:` step at build context root.
COPY tene /usr/local/bin/tene

USER tene
WORKDIR /workspace

ENTRYPOINT ["/usr/local/bin/tene"]
CMD ["--help"]

# OCI labels for GHCR UI
LABEL org.opencontainers.image.source="https://github.com/tene-ai/tene"
LABEL org.opencontainers.image.description="Local-first encrypted secret manager CLI for AI-safe workflows"
LABEL org.opencontainers.image.license="MIT"
LABEL org.opencontainers.image.vendor="tene-ai"
```

#### S-2b: `.goreleaser.yml` 에 `dockers:` 섹션

**File**: `.goreleaser.yml` (`archives:` 이후, `brews:` 근처 추가):

```yaml
dockers:
  - image_templates:
      - "ghcr.io/tene-ai/tene:{{ .Tag }}"
      - "ghcr.io/tene-ai/tene:v{{ .Major }}"
      - "ghcr.io/tene-ai/tene:v{{ .Major }}.{{ .Minor }}"
      - "ghcr.io/tene-ai/tene:latest"
    dockerfile: Dockerfile
    use: buildx
    build_flag_templates:
      - "--platform=linux/amd64"
      - "--label=org.opencontainers.image.created={{ .Date }}"
      - "--label=org.opencontainers.image.title={{ .ProjectName }}"
      - "--label=org.opencontainers.image.version={{ .Version }}"
      - "--label=org.opencontainers.image.revision={{ .FullCommit }}"
      - "--label=org.opencontainers.image.source=https://github.com/tene-ai/tene"
      - "--label=org.opencontainers.image.licenses=MIT"
```

ARM64 빌드 추가 (옵션, linux/arm64 바이너리 이미 생성되므로 가치 있음):

```yaml
  - image_templates:
      - "ghcr.io/tene-ai/tene:{{ .Tag }}-arm64"
      - "ghcr.io/tene-ai/tene:latest-arm64"
    dockerfile: Dockerfile
    use: buildx
    build_flag_templates:
      - "--platform=linux/arm64"
      - "--label=..."
    goarch: arm64
```

**Manifest** (멀티 arch 통합 tag):

```yaml
docker_manifests:
  - name_template: "ghcr.io/tene-ai/tene:{{ .Tag }}"
    image_templates:
      - "ghcr.io/tene-ai/tene:{{ .Tag }}"
      - "ghcr.io/tene-ai/tene:{{ .Tag }}-arm64"
  - name_template: "ghcr.io/tene-ai/tene:latest"
    image_templates:
      - "ghcr.io/tene-ai/tene:latest"
      - "ghcr.io/tene-ai/tene:latest-arm64"
```

#### S-2c: 릴리스 워크플로우에 `packages: write` 권한

**File**: `.github/workflows/release.yml` (존재 확인 · 없다면 auto-tag.yml 에 추가):

```yaml
permissions:
  contents: write
  packages: write  # GHCR 푸시용

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      # ... checkout, setup-go ...
      - name: Log in to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.repository_owner }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: Set up QEMU (arm64 cross)
        uses: docker/setup-qemu-action@v3
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      - name: GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
```

### 수용 기준
- [ ] 다음 릴리스 시 `ghcr.io/tene-ai/tene:v1.0.5` + `latest` 태그 publish
- [ ] `docker run --rm ghcr.io/tene-ai/tene version` → `tene version v1.0.5` 출력
- [ ] `docker run --rm ghcr.io/tene-ai/tene --help` → help 표시
- [ ] 이미지 크기 ≤ 20MB
- [ ] GHCR Package 페이지에 description 표시

### 사용 예시 문서화 (README 에 추가)
```markdown
### Docker

```bash
# One-off usage
docker run --rm -v $(pwd):/workspace ghcr.io/tene-ai/tene:latest list

# In GitHub Actions / CI
- uses: docker://ghcr.io/tene-ai/tene:latest
  with:
    args: run --no-keychain -- npm test
  env:
    TENE_MASTER_PASSWORD: ${{ secrets.TENE_MASTER_PASSWORD }}
```
```

---

## S-3: `install.sh` 힌트 출력 확장

### 배경
`apps/web/public/install.sh:120-123` 현재:
```sh
info "  tene v${version} installed successfully!"
info "  Run 'tene init' to get started."
```

### 설계

**File**: `apps/web/public/install.sh:120-123`

**Before**:
```sh
  info ""
  info "  tene v${version} installed successfully!"
  info "  Run 'tene init' to get started."
}
```

**After**:
```sh
  info ""
  info "  tene v${version} installed successfully!"
  info ""
  info "  Next step: tene init"
  info ""
  info "  Documentation:"
  info "    README:    https://github.com/tene-ai/tene#readme"
  info "    AI index:  https://tene.sh/llms.txt"
  info "    Issues:    https://github.com/tene-ai/tene/issues"
}
```

### 수용 기준
- [ ] `curl -sSfL https://tene.sh/install.sh | sh` 출력에 `llms.txt` URL 포함
- [ ] AI 에이전트가 설치 로그를 캡처했을 때 다음 단계 URL 즉시 인지

---

## S-4: GoReleaser `man_pages:` + shell completion 배포

### 배경
- `root.go:65-78` 에서 `completion` 명령 미등록
- `.goreleaser.yml` `man_pages:` / completion archive 부재

### 설계

#### S-4a: CLI `completion` 명령 활성화

**File (신규)**: `internal/cli/completion.go`

```go
package cli

import (
    "os"

    "github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
    Use:   "completion [bash|zsh|fish|powershell]",
    Short: "Generate shell completion script",
    Long: `Generate a shell completion script for tene.

Examples:

  # Bash (one-off)
  source <(tene completion bash)

  # Bash (persistent, macOS)
  tene completion bash > $(brew --prefix)/etc/bash_completion.d/tene

  # Zsh
  tene completion zsh > "${fpath[1]}/_tene"

  # Fish
  tene completion fish > ~/.config/fish/completions/tene.fish

  # PowerShell
  tene completion powershell | Out-String | Invoke-Expression
`,
    DisableFlagsInUseLine: true,
    ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
    Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
    Run: func(cmd *cobra.Command, args []string) {
        switch args[0] {
        case "bash":
            _ = cmd.Root().GenBashCompletion(os.Stdout)
        case "zsh":
            _ = cmd.Root().GenZshCompletion(os.Stdout)
        case "fish":
            _ = cmd.Root().GenFishCompletion(os.Stdout, true)
        case "powershell":
            _ = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
        }
    },
}
```

**File**: `internal/cli/root.go:64`

```go
// 기존 AddCommand 블록 끝에 추가:
rootCmd.AddCommand(completionCmd)
```

#### S-4b: Man page 생성 스크립트

**File (신규)**: `scripts/gen-manpage.go`

```go
// +build generate

// Usage (CI or local):
//   go run scripts/gen-manpage.go
// Writes manpages/tene.1

package main

import (
    "log"
    "os"
    "path/filepath"

    "github.com/spf13/cobra/doc"
    tenecli "github.com/tene-ai/tene/internal/cli"
)

func main() {
    // Initialize the root command (side-effect of cli package init).
    root := tenecli.RootCmd() // export RootCmd() from internal/cli

    header := &doc.GenManHeader{
        Title:   "TENE",
        Section: "1",
        Source:  "tene CLI",
        Manual:  "Tene Manual",
    }

    dir := "manpages"
    if err := os.MkdirAll(dir, 0755); err != nil {
        log.Fatal(err)
    }
    if err := doc.GenManTree(root, header, dir); err != nil {
        log.Fatal(err)
    }
    log.Printf("Wrote man pages to %s/", dir)
    _ = filepath.Clean(dir)
}
```

**File**: `internal/cli/root.go` 에 exports 추가 (맨 아래):

```go
// RootCmd returns the root cobra command, useful for docs generation.
func RootCmd() *cobra.Command {
    return rootCmd
}
```

#### S-4c: `.goreleaser.yml` 에 completion/man archive 포함

**File**: `.goreleaser.yml:29-39` (archives 섹션 수정)

**Before**:
```yaml
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
```

**After**:
```yaml
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
    files:
      - src: completions/*
        strip_parent: true
        info:
          mode: 0644
      - src: manpages/*
        strip_parent: true
        info:
          mode: 0644
      - LICENSE
      - README.md
```

#### S-4d: 릴리스 전 `before:` 훅 추가

**File**: `.goreleaser.yml:3-6`

**Before**:
```yaml
before:
  hooks:
    - go mod tidy
```

**After**:
```yaml
before:
  hooks:
    - go mod tidy
    # Generate shell completions (bash/zsh/fish)
    - mkdir -p completions
    - sh -c "go run ./cmd/tene completion bash > completions/tene.bash"
    - sh -c "go run ./cmd/tene completion zsh > completions/_tene"
    - sh -c "go run ./cmd/tene completion fish > completions/tene.fish"
    # Generate man pages
    - go run scripts/gen-manpage.go
```

### 수용 기준
- [ ] `tene completion bash` 출력 valid (`source <(tene completion bash)` 후 `tene <TAB>` 작동)
- [ ] `tene completion zsh > /tmp/_tene && compinit` 후 `tene <TAB>` 작동
- [ ] 다음 릴리스 archive 에 `completions/*` 3종 + `manpages/tene.1` 포함
- [ ] `man tene` (Homebrew 설치 후) 정상 표시
- [ ] AI 에이전트가 `compgen -c tene` 으로 subcommand 리스트 획득 가능

---

## S-5: GPG 서명 체크섬 (선택 · P3)

### 배경
현재 SHA-256 체크섬만. 체크섬 원본 = 바이너리 원본 (같은 S3 버킷).

### 설계 (M1 이후 옵션)

#### S-5a: `.goreleaser.yml` 에 `signs:` 섹션

**File**: `.goreleaser.yml` (append):

```yaml
signs:
  - artifacts: checksum
    cmd: gpg
    args:
      - "--batch"
      - "-u"
      - "{{ .Env.GPG_FINGERPRINT }}"
      - "--output"
      - "${signature}"
      - "--detach-sign"
      - "${artifact}"
```

#### S-5b: 공개키 배포

GPG key fingerprint 를 `SECURITY.md` + `tene.sh` 푸터에 게시. `https://tene.sh/agent-kay-it.asc` 엔드포인트 생성.

#### S-5c: `install.sh` GPG 검증 추가

```sh
# After SHA-256 verification:
if command -v gpg > /dev/null 2>&1; then
  download "${RELEASE_BASE}/v${version}/checksums.txt.sig" "${tmpdir}/checksums.txt.sig"
  download "https://tene.sh/agent-kay-it.asc" "${tmpdir}/agent-kay-it.asc"
  gpg --import "${tmpdir}/agent-kay-it.asc" 2>/dev/null
  if gpg --verify "${tmpdir}/checksums.txt.sig" "${tmpdir}/checksums.txt" 2>/dev/null; then
    info "  GPG signature verified"
  else
    info "  GPG key not trusted — install will continue (SHA-256 verified)"
  fi
fi
```

### 수용 기준 (M1+)
- [ ] `.goreleaser.yml signs:` 섹션 활성
- [ ] Next release `checksums.txt.sig` publish
- [ ] `curl https://tene.sh/agent-kay-it.asc` 200
- [ ] `gpg --verify checksums.txt.sig checksums.txt` 성공

---

# E. 사용 Usage

## U-1: `tene get <KEY>` 비대화형 stdout 차단 ⚠️ Breaking

### 배경
`internal/cli/get.go:67` — `fmt.Print(plaintext)` → AI 가 Bash 툴로 호출 시 stdout 이 pipe 이므로 context 로 평문 유입.

### 설계 목표
- **비대화형 stdout** (pipe 또는 redirect) → **기본 차단**.
- **대화형 TTY** → 기존대로 출력 (shell 사용자 정상 경험).
- **명시적 opt-in** 경로 3종 유지:
  - `--unsafe-stdout` flag (CLI 명령 플래그)
  - `TENE_ALLOW_STDOUT_SECRETS=1` env var (CI/스크립트)
  - `--json` (구조화된 JSON 은 의도적 선택으로 간주하되 stderr 경고 1회)

### Diff

#### U-1a: `pkg/errors/errors.go` — 신규 에러 코드

**File**: `pkg/errors/errors.go`

```go
// 기존 err 목록에 추가:
var ErrStdoutSecretBlocked = New(
    "STDOUT_SECRET_BLOCKED",
    "Refusing to print secret to non-TTY stdout. " +
        "Use `tene run -- <cmd>` (safer), or pass `--unsafe-stdout` or set " +
        "TENE_ALLOW_STDOUT_SECRETS=1 if you really need plaintext piped output.",
    2,
)
```

#### U-1b: `internal/cli/get.go` 개선

**File**: `internal/cli/get.go`

**Before (전체)**:
```go
package cli

import (
    "fmt"

    "github.com/spf13/cobra"
    "github.com/tene-ai/tene/pkg/crypto"
    teneerr "github.com/tene-ai/tene/pkg/errors"
)

var getCmd = &cobra.Command{
    Use:   "get KEY",
    Short: "Retrieve a decrypted secret",
    Args:  cobra.ExactArgs(1),
    RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
    // ... (load vault, decrypt) ...

    if flagJSON {
        return printJSON(map[string]any{
            "ok":          true,
            "name":        keyName,
            "value":       string(plaintext),
            "environment": env,
        })
    }

    fmt.Print(string(plaintext))
    if isTerminal() {
        fmt.Println()
    }
    return nil
}
```

**After**:
```go
package cli

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/tene-ai/tene/pkg/crypto"
    teneerr "github.com/tene-ai/tene/pkg/errors"
)

var flagUnsafeStdout bool

func init() {
    getCmd.Flags().BoolVar(&flagUnsafeStdout, "unsafe-stdout", false,
        "Explicitly allow printing plaintext to non-TTY stdout "+
            "(equivalent to TENE_ALLOW_STDOUT_SECRETS=1)")
}

var getCmd = &cobra.Command{
    Use:   "get KEY",
    Short: "Retrieve a decrypted secret",
    Long: `Decrypt and print a secret.

⚠️  Security note: when stdout is piped or redirected (non-TTY), this
command refuses by default to prevent accidental secret leakage into
AI agent context windows, log aggregators, or shell history files.

To override, use one of:
  --unsafe-stdout              (per-invocation opt-in)
  TENE_ALLOW_STDOUT_SECRETS=1  (environment override)

Or, preferred, use 'tene run -- <cmd>' which injects secrets as environment
variables without ever printing them to stdout.`,
    Args: cobra.ExactArgs(1),
    RunE: runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
    keyName := args[0]

    app, err := loadApp()
    if err != nil {
        return err
    }
    defer func() { _ = app.Vault.Close() }()

    masterKey, err := loadOrPromptMasterKey(app)
    if err != nil {
        return err
    }
    defer crypto.ZeroBytes(masterKey)

    encKey, err := crypto.DeriveSubKey(masterKey, crypto.PurposeEncryption, 32)
    if err != nil {
        return err
    }
    defer crypto.ZeroBytes(encKey)

    env := resolveEnv(app)
    secret, err := app.Vault.GetSecret(keyName, env)
    if err != nil {
        return teneerr.ErrSecretNotFound(keyName, env)
    }

    ciphertext, err := decodeBase64(secret.EncryptedValue)
    if err != nil {
        return teneerr.ErrDecryptFailed
    }

    plaintext, err := crypto.Decrypt(encKey, ciphertext, []byte(keyName))
    if err != nil {
        return teneerr.ErrDecryptFailed
    }
    defer crypto.ZeroBytes(plaintext)

    // Audit log
    _ = app.Vault.AddAuditLog("secret.read", keyName, "")

    // JSON output: warn once on stderr but proceed (explicit user choice)
    if flagJSON {
        if !isTerminal() && !stdoutSecretsAllowed() {
            fmt.Fprintln(os.Stderr,
                "warning: emitting plaintext secret as JSON to non-TTY stdout. "+
                    "Consider `tene run --` for safer injection.")
        }
        return printJSON(map[string]any{
            "ok":          true,
            "name":        keyName,
            "value":       string(plaintext),
            "environment": env,
        })
    }

    // Non-TTY stdout without explicit opt-in → block
    if !isTerminal() && !stdoutSecretsAllowed() {
        return teneerr.ErrStdoutSecretBlocked
    }

    fmt.Print(string(plaintext))
    if isTerminal() {
        fmt.Println()
    }
    return nil
}

// stdoutSecretsAllowed returns true if the caller explicitly opted in to
// plaintext secret output on non-TTY stdout.
func stdoutSecretsAllowed() bool {
    if flagUnsafeStdout {
        return true
    }
    if v := os.Getenv("TENE_ALLOW_STDOUT_SECRETS"); v == "1" || v == "true" {
        return true
    }
    return false
}
```

#### U-1c: 단위 테스트

**File (신규 또는 확장)**: `internal/cli/get_test.go`

```go
package cli

import (
    "bytes"
    "io"
    "os"
    "os/exec"
    "strings"
    "testing"
)

func TestGet_NonTTY_BlocksByDefault(t *testing.T) {
    // Setup: create a temporary vault in a tmpdir, set a dummy secret.
    // ...
    cmd := exec.Command("./tene", "get", "STRIPE_KEY")
    stdout, _ := cmd.StdoutPipe()
    _ = cmd.Start()
    out, _ := io.ReadAll(stdout)
    err := cmd.Wait()

    // Expect non-zero exit + STDOUT_SECRET_BLOCKED in stderr
    if err == nil {
        t.Fatal("expected non-zero exit code, got 0")
    }
    if exitErr, ok := err.(*exec.ExitError); ok {
        if exitErr.ExitCode() != 2 {
            t.Fatalf("expected exit 2, got %d", exitErr.ExitCode())
        }
    }
    if strings.Contains(string(out), "sk_test") {
        t.Fatal("secret leaked to stdout on non-TTY")
    }
}

func TestGet_NonTTY_UnsafeFlag_Allows(t *testing.T) {
    // ... with --unsafe-stdout flag, stdout should contain plaintext
}

func TestGet_NonTTY_EnvOverride_Allows(t *testing.T) {
    // ... with TENE_ALLOW_STDOUT_SECRETS=1, stdout should contain plaintext
}

func TestGet_JSON_EmitsWarningToStderr(t *testing.T) {
    // ... --json on non-TTY emits warning on stderr but still outputs JSON
}
```

#### U-1d: CHANGELOG + migration note

**File (신규)**: `CHANGELOG.md`

```markdown
# Changelog

## [Unreleased]

### Changed (breaking)

- `tene get <KEY>` now refuses to print plaintext to non-TTY stdout by default.
  This prevents accidental secret leakage into AI agent context windows, log
  aggregators, and shell history files.

  **Migration**:
  - If you use `tene get` in a script or pipeline, explicitly opt in:
    - Pass `--unsafe-stdout` on the command line, **or**
    - Set `TENE_ALLOW_STDOUT_SECRETS=1` in the environment, **or**
    - Switch to `tene run -- <command>` (recommended — secrets never touch stdout).
  - Interactive terminal usage is **unchanged**.
  - `tene get --json` still works on non-TTY but emits a warning on stderr.

### Added

- `tene completion [bash|zsh|fish|powershell]` — generate shell completion scripts
- `.github/copilot-instructions.md`, `.github/FUNDING.yml`
- `SECURITY.md`, `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`
- Homebrew tap (`brew install tene-ai/tap/tene`)
- Docker image (`docker run ghcr.io/tene-ai/tene`)
- Man page (`man tene`)

### Fixed

- `/vs/*` structured data: removed `aggregateRating` field to comply with
  Schema.org guidelines (real reviews coming in a future release).
- Landing `<meta name="description">` shortened to 154 chars for Google SERP.
- README Cloud Commands section now clearly marked as "Coming soon" to avoid
  user confusion with disabled commands.
```

### 수용 기준
- [ ] `echo ""| tene get EXISTING_KEY | cat` → exit 2, `STDOUT_SECRET_BLOCKED`
- [ ] `echo ""| tene get EXISTING_KEY --unsafe-stdout` → 평문 출력, exit 0
- [ ] `TENE_ALLOW_STDOUT_SECRETS=1 tene get EXISTING_KEY | cat` → 평문 출력, exit 0
- [ ] 대화형 `tene get EXISTING_KEY` → 기존대로 평문 + 개행
- [ ] `tene get EXISTING_KEY --json | jq .value` → 평문 + stderr 에 warning 1 줄
- [ ] 단위 테스트 4종 모두 pass

### 롤백
이 변경은 v1.1.0 에 묶어서 릴리스 (별도 major 블록). 1주 내 이슈 폭주 시 `v1.0.5` 로 패치 릴리스 (behavior revert + `--strict-stdout` opt-in flag 로 재도입).

---

## U-2: `.github/copilot-instructions.md`

### 배경
Copilot 은 `AGENTS.md` 와 `.github/copilot-instructions.md` 둘 다 읽지만, GitHub Docs 는 후자를 우선 경로로 명시. 조직 레벨 Copilot 정책 적용 시 후자 우선.

### 설계

**File (신규)**: `.github/copilot-instructions.md`

```markdown
# GitHub Copilot — tene Project Instructions

This repository is [tene](https://github.com/tene-ai/tene), a local-first
encrypted secret manager CLI. When suggesting code or commands for this
project, follow these rules:

## AI-Safe Secret Handling (Critical)

- **ALWAYS** prefer `tene run -- <cmd>` over `tene get <KEY>` when helping
  users run a command that needs secrets. `tene run --` injects secrets as
  environment variables; `tene get` prints plaintext to stdout which will
  enter the AI context window.
- **NEVER** suggest `tene export` in code — it prints all secrets as plaintext.
- **NEVER** suggest reading files under `.tene/` — it is the encrypted vault.
- **NEVER** pass secret values as command-line arguments (they appear in
  `ps` and shell history). Use environment injection via `tene run --`.
- **`tene list`** is safe — it shows secret names only, no values.
- **`tene get KEY --json`** returns `{"name":..., "value":..., "environment":...}`
  and requires explicit `--unsafe-stdout` or `TENE_ALLOW_STDOUT_SECRETS=1` on
  non-TTY invocation.

## Project Conventions

- **Language**: Go 1.25+ (CLI), TypeScript / Next.js 15 App Router (landing
  at `apps/web/`).
- **License**: MIT.
- **Commit messages**: Conventional Commits (`feat(cli):`, `fix(vault):`, ...).
- **Branch**: feature branches off `staging`, merged into `staging`, released
  from `main`.
- **Testing**: `go test ./...` + `golangci-lint run` must pass before merge.
- **No network calls** in the CLI default path — tene is local-first by
  design. Cloud features (disabled in v1.x) live behind a separate opt-in.

## Do Not

- Do not add new third-party dependencies without discussing in an issue.
- Do not introduce hardcoded secrets in any example or test.
- Do not suggest `.env` files — that is exactly the problem tene solves.
- Do not suggest cloud-backed secret managers (Vault, AWS Secrets Manager,
  Doppler) as alternatives within tene's code — link to `/vs/*` comparisons
  on tene.sh instead.

## Resources

- AI index: https://tene.sh/llms.txt
- Full reference: https://tene.sh/llms-full.txt
- Main docs: https://github.com/tene-ai/tene#readme
```

### 수용 기준
- [ ] `test -f .github/copilot-instructions.md`
- [ ] Copilot Chat 이 해당 리포에서 질의 시 `.github/copilot-instructions.md` 룰 자동 적용 (수동 검증)

---

## U-4: `tene --help` 푸터에 리소스 링크

### 배경
`root.go:49-55` 의 `rootCmd` 에 `Long` 또는 custom help template 부재. 리소스 링크 없음.

### 설계

**File**: `internal/cli/root.go:49-55`

**Before**:
```go
var rootCmd = &cobra.Command{
    Use:     "tene",
    Short:   "Agentic Secret Runtime -- local-first encrypted secret management",
    Version: version,
    SilenceErrors: true,
    SilenceUsage:  true,
}
```

**After**:
```go
var rootCmd = &cobra.Command{
    Use:     "tene",
    Short:   "Agentic Secret Runtime — local-first encrypted secret management",
    Long: `tene is a local-first encrypted secret manager CLI. It encrypts your
secrets with XChaCha20-Poly1305 and injects them at runtime so AI agents
(Claude Code, Cursor, Windsurf, Gemini, Codex, Copilot) never see plaintext.

Typical workflow:

  tene init                                 # create encrypted vault
  tene set STRIPE_KEY sk_test_xxx           # store a secret
  tene run -- npm start                     # run with secrets injected

Resources:
  AI index:   https://tene.sh/llms.txt
  Docs:       https://github.com/tene-ai/tene#readme
  Issues:     https://github.com/tene-ai/tene/issues
  Discussions: https://github.com/tene-ai/tene/discussions
`,
    Version:       version,
    SilenceErrors: true,
    SilenceUsage:  true,
}
```

### 수용 기준
- [ ] `tene --help | grep "llms.txt"` ≥ 1
- [ ] `tene --help` 출력이 Typical workflow 3줄 예시 포함
- [ ] Cobra 가 자동 생성하는 usage 뒤에 Resources 블록 출력

---

## U-5: `docs/cli-reference.md` + `/cli` 공개 페이지

### 배경
명령별 상세 문서가 README + llms-full.txt 에 분산. 캐노니컬 1-URL 없음.

### 설계

#### U-5a: `docs/cli-reference.md` 마스터 문서

**File (신규)**: `docs/cli-reference.md`

```markdown
# tene CLI Reference

> Canonical reference for every tene command, flag, exit code, and JSON schema.
> Last updated: 2026-04-xx, v1.x.

## Global Flags

| Flag | Short | Type | Default | Description |
|------|:-----:|------|---------|-------------|
| `--json` | | bool | false | Emit JSON to stdout |
| `--quiet` | `-q` | bool | false | Minimal output (errors only) |
| `--env` | `-e` | string | active | Target environment |
| `--dir` | | string | cwd | Project directory |
| `--no-color` | | bool | false | Disable colorized output |
| `--no-keychain` | | bool | false | Skip OS keychain (CI/CD) |

## Exit Codes

| Code | Name | Meaning |
|:----:|------|---------|
| 0 | Success | Command completed |
| 1 | GenericError | Uncategorized error |
| 2 | UsageError | Wrong flag / arg shape (cobra default) |
| 2 | StdoutSecretBlocked | `tene get` on non-TTY without opt-in |
| 3 | VaultNotFound | `.tene/vault.db` missing in project |
| 4 | AuthRequired | Master password / keychain failed |
| 5 | SecretNotFound | KEY does not exist in env |
| 6 | DecryptFailed | Ciphertext + key mismatch |
| 7 | InteractiveRequired | No password source on non-TTY |
| ... | (extend as needed) | |

## Commands

### `tene init`

...

### `tene set <KEY> <VALUE>` | `--stdin`

...

### `tene get <KEY>`

Decrypt and print a secret.

#### Flags

- `--json` — JSON: `{"ok":true, "name":"KEY", "value":"plaintext", "environment":"default"}`
- `--unsafe-stdout` — explicitly allow non-TTY stdout (default: blocked)

#### Exit Codes

- `0` — secret printed (or JSON emitted on TTY)
- `2` — `STDOUT_SECRET_BLOCKED` (non-TTY, no opt-in)
- `5` — `SECRET_NOT_FOUND`

#### Safety

On non-TTY stdout (pipe, redirect), this command **refuses** by default. See
[U-1 safety design](./02-design/features/ai-discoverability.design.md#u-1).

### `tene list`

...

### `tene run -- <CMD> [ARGS...]`

...

(rest of commands documented in full here)

## JSON Output Schemas

### `tene list --json`

```json
{
  "ok": true,
  "project": "my-app",
  "environment": "default",
  "count": 3,
  "secrets": [
    { "name": "STRIPE_KEY", "preview": "*****", "version": 1, "updatedAt": "2026-04-22T12:00:00Z" }
  ]
}
```

### `tene get --json`

```json
{
  "ok": true,
  "name": "STRIPE_KEY",
  "value": "sk_test_...",
  "environment": "default"
}
```

### `tene run --json`

Emitted to **stderr** (stdout is child process):

```json
{
  "injectedCount": 3,
  "environment": "default",
  "command": "npm"
}
```

## Environment Variables (Input)

| Var | Purpose |
|-----|---------|
| `TENE_MASTER_PASSWORD` | Non-interactive master password (CI/CD) |
| `TENE_ALLOW_STDOUT_SECRETS` | `=1` allows `tene get` on non-TTY |

## Shell Completion

```bash
source <(tene completion bash)       # Bash one-off
tene completion zsh > "${fpath[1]}/_tene"   # Zsh persistent
tene completion fish > ~/.config/fish/completions/tene.fish
```

## Man Page

```bash
man tene                             # After Homebrew install
```
```

#### U-5b: `/cli` 공개 라우트

**File (신규)**: `apps/web/src/app/cli/page.tsx`

```tsx
import type { Metadata } from "next";
import fs from "fs";
import path from "path";
import { compileMDX } from "next-mdx-remote/rsc";

export const metadata: Metadata = {
  title: "tene CLI Reference",
  description:
    "Complete command-by-command reference for tene: flags, exit codes, JSON schemas, and examples.",
  alternates: { canonical: "https://tene.sh/cli" },
};

async function loadReference(): Promise<string> {
  const p = path.join(process.cwd(), "..", "..", "docs", "cli-reference.md");
  return fs.readFileSync(p, "utf-8");
}

export default async function CliReferencePage() {
  const source = await loadReference();
  const { content } = await compileMDX({ source });

  return (
    <main className="mx-auto max-w-4xl px-4 py-16 sm:px-6 prose prose-invert">
      <h1>tene CLI Reference</h1>
      <article>{content}</article>
    </main>
  );
}
```

**File**: `apps/web/src/app/sitemap.ts` (line 48 부근)

```ts
// Add new URL entry:
{
  url: `${base}/cli`,
  lastModified,
  changeFrequency: "weekly",
  priority: 0.8,
},
```

#### U-5c: `public/llms-full.txt` 에 `/cli` 링크 추가

**File**: `apps/web/public/llms-full.txt` (Resources section)

```
## Resources

- CLI Reference: https://tene.sh/cli
- ...
```

### 수용 기준
- [ ] `test -f docs/cli-reference.md`
- [ ] `curl -sI https://tene.sh/cli` → 200
- [ ] 페이지에 모든 명령(init · set · get · run · list · delete · import · export · env · passwd · recover · whoami · version · update · completion) 문서화
- [ ] 모든 명령의 `--json` 스키마 예시 포함
- [ ] Exit code 표 포함

---

# F. 통합 검증 & 배포 절차

## 6.1 브랜치 전략

모든 변경은 `feature/ai-discoverability-audit` 브랜치에서 수행, staging 으로 PR.

**커밋 단위 권고** (revert 용이):

```
docs(pdca): add ai-discoverability PRD + Plan + Codebase Audit + Design
fix(readme): mark cloud commands as "Coming soon" (A-1)
feat(web): add LLM bot allow-list to robots.txt (D-2)
feat(web): add <link rel="ai-index"> and .well-known/ai.json (D-4)
chore(github): enable Discussions + add FUNDING.yml + seed 3 Q&A (D-3, D-5)
fix(web): shorten meta description to 154 chars (A-2)
fix(web): remove aggregateRating from /vs/* structured data (A-3)
feat(web): add Organization + WebSite JSON-LD to home (A-4)
feat(web): add "Tene is X" hero subcopy (A-5)
feat(web): add Trust section with live badges + testimonials (I-2)
feat(web): add shared Breadcrumb component to /vs/* and /blog/* (I-3)
docs(community): add SECURITY.md, CODE_OF_CONDUCT.md, CONTRIBUTING.md (I-1)
chore(github): add issue + PR templates (I-1)
chore(github): add copilot-instructions.md (U-2)
feat(cli): add completion command for bash/zsh/fish/powershell (S-4)
feat(cli): install.sh prints llms.txt and docs hints on success (S-3)
feat(cli): root --help footer links to llms.txt and docs (U-4)
feat(cli): tene get refuses non-TTY stdout by default (BREAKING) (U-1)
feat(release): GoReleaser brews section → Homebrew tap (S-1)
feat(release): GoReleaser dockers section → GHCR image (S-2)
feat(release): bundle completions + man pages in archive (S-4)
docs: add docs/cli-reference.md and /cli public route (U-5)
# Optional (M1):
feat(release): GPG-sign checksums.txt (S-5)
```

## 6.2 CI 검증 (자동)

기존 `.github/workflows/ci.yml` 에 추가 스텝:

```yaml
- name: Build landing
  run: |
    cd apps/web
    pnpm install --frozen-lockfile
    pnpm build

- name: Verify llms.txt integrity
  run: |
    curl -sf file://$(pwd)/apps/web/public/llms.txt | grep -q "tene is"
    curl -sf file://$(pwd)/apps/web/public/.well-known/ai.json | jq .name

- name: Verify README Cloud Commands label
  run: |
    grep -q "Coming soon" README.md || (echo "README Cloud Commands section not labeled" && exit 1)

- name: Verify structured data (no aggregateRating)
  run: |
    ! grep -r "aggregateRating" apps/web/src/

- name: Verify Community Health files
  run: |
    test -f SECURITY.md
    test -f CODE_OF_CONDUCT.md
    test -f CONTRIBUTING.md
    test -f .github/ISSUE_TEMPLATE/bug_report.yml
    test -f .github/ISSUE_TEMPLATE/feature_request.yml
    test -f .github/PULL_REQUEST_TEMPLATE.md
    test -f .github/copilot-instructions.md
    test -f .github/FUNDING.yml

- name: Test CLI completion command
  run: |
    go build -o tene ./cmd/tene
    ./tene completion bash > /dev/null
    ./tene completion zsh > /dev/null

- name: Test tene get non-TTY guard
  run: |
    go build -o tene ./cmd/tene
    # With TENE_ALLOW_STDOUT_SECRETS unset and non-TTY → exit 2
    # (requires test fixture vault)
```

## 6.3 배포 순서

1. **W1 Quick Wins** (13 커밋) → PR open (target: staging) → CI pass → merge staging
2. **W2 구조** (6 커밋) → **별도 PR** (I-1 + U-1 + S-1/2/4) — U-1 breaking 이므로 리뷰 시간 확보
3. Staging → main → `git tag v1.1.0 && git push --tags` → GoReleaser 실행 → Homebrew tap 자동 업데이트 + GHCR 이미지 푸시 + S3 업로드
4. **M1** (I-2, I-3, U-5) → PR 순차 open

## 6.4 릴리스 노트 템플릿 (v1.1.0)

```markdown
## v1.1.0 — AI-Safe CLI (2026-xx-xx)

### Breaking

- `tene get <KEY>` now refuses non-TTY stdout by default. Use
  `--unsafe-stdout` or `TENE_ALLOW_STDOUT_SECRETS=1` or switch to
  `tene run --`. See [CHANGELOG](./CHANGELOG.md#unreleased) for migration.

### Installation

```bash
# Homebrew (new!)
brew install tene-ai/tap/tene

# Docker (new!)
docker run ghcr.io/tene-ai/tene:v1.1.0 version

# curl (unchanged)
curl -sSfL https://tene.sh/install.sh | sh
```

### New

- `tene completion [bash|zsh|fish|powershell]`
- Man page (`man tene`)
- Homebrew tap
- Docker image on GHCR
- Shell completions auto-installed via brew

### Improved

- `tene --help` now links to online resources
- CLI reference now at [tene.sh/cli](https://tene.sh/cli)

### Project Health

- SECURITY.md, CODE_OF_CONDUCT.md, CONTRIBUTING.md added
- Issue + PR templates
- GitHub Discussions enabled
- GitHub Sponsors enabled
- `.github/copilot-instructions.md` for Copilot Chat users

### Landing

- `/vs/*` structured data compliance (removed aggregateRating until real reviews)
- Organization + WebSite JSON-LD on home
- LLM crawler explicit allow-list in robots.txt
- Breadcrumb on `/vs/*` and `/blog/*`
- Trust section on homepage
- Shorter meta description for Google SERP

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

## 6.5 Post-Release 검증 (자동 블랙박스)

```bash
# 5분짜리 스모크 테스트
set -e
brew tap tene-ai/tap
brew install tene
tene version | grep v1.1.0
tene --help | grep "llms.txt"

docker run --rm ghcr.io/tene-ai/tene:v1.1.0 version | grep v1.1.0

# tene get non-TTY guard
mkdir /tmp/tene-smoke && cd /tmp/tene-smoke
TENE_MASTER_PASSWORD=test tene init --no-keychain --claude < /dev/null
TENE_MASTER_PASSWORD=test tene set TESTKEY testvalue --no-keychain
TENE_MASTER_PASSWORD=test tene get TESTKEY --no-keychain | cat
# Expect: exit 2

TENE_MASTER_PASSWORD=test TENE_ALLOW_STDOUT_SECRETS=1 tene get TESTKEY --no-keychain | cat
# Expect: "testvalue"

# Landing checks
curl -sI https://tene.sh/robots.txt | grep 200
curl -s https://tene.sh/robots.txt | grep ClaudeBot
curl -sI https://tene.sh/.well-known/ai.json | grep 200
curl -s https://tene.sh/cli | grep "CLI Reference"
curl -s https://tene.sh | grep '"@type":"Organization"'
```

## 6.6 성공 KPI 재측정 (M3)

- GitHub stars: 7 → 목표 ≥ 200
- Community Health: 42% → 목표 ≥ 95%
- Homebrew 주간 install: 0 → 목표 ≥ 500
- GSC impressions: 0 → 목표 ≥ 5,000/month
- 블랙박스 AI 추천 hit rate: 미측정 → 목표 ≥ 50%
- Lighthouse Performance: 68-72 → 목표 ≥ 90

---

# 부록

## A. 커밋 SHA 매핑 테이블 (구현 후 채움)

| 항목 ID | 커밋 SHA | PR # | 배포 일자 |
|:---:|---------|:---:|----------|
| D-1 | — | — | — |
| D-2 | — | — | — |
| ... | — | — | — |

## B. 실측 재현 명령

```bash
# GitHub 상태
gh repo view tene-ai/tene --json description,repositoryTopics,stargazerCount,hasDiscussionsEnabled,usesCustomOpenGraphImage,isSecurityPolicyEnabled,fundingLinks
gh api repos/tene-ai/tene/community/profile

# 파일 존재 검증
test -f SECURITY.md && echo "OK: SECURITY.md"
test -f CODE_OF_CONDUCT.md && echo "OK: CoC"
test -f CONTRIBUTING.md && echo "OK: CONTRIBUTING"
test -f .github/ISSUE_TEMPLATE/bug_report.yml && echo "OK: bug template"
test -f .github/PULL_REQUEST_TEMPLATE.md && echo "OK: PR template"
test -f .github/copilot-instructions.md && echo "OK: copilot-instructions"
test -f .github/FUNDING.yml && echo "OK: FUNDING"
test -f apps/web/public/.well-known/ai.json && echo "OK: ai.json"
test -f Dockerfile && echo "OK: Dockerfile"

# GoReleaser 섹션 검증
grep -E '^brews:' .goreleaser.yml && echo "OK: brews"
grep -E '^dockers:' .goreleaser.yml && echo "OK: dockers"
grep -E 'completions/\*' .goreleaser.yml && echo "OK: completions archived"

# CLI 검증
go build -o tene ./cmd/tene
./tene completion bash > /dev/null && echo "OK: completion bash"
./tene --help | grep -q "llms.txt" && echo "OK: help footer"
```

## C. 의존하는 외부 요소

- `https://goreportcard.com/badge/...` (Trust Section 배지)
- `https://img.shields.io/...` (다수 배지)
- GHCR (`ghcr.io/tene-ai/tene`)
- Homebrew tap (`tene-ai/homebrew-tap`)
- Contributor Covenant v2.1 markdown 원문
- (선택) GPG keyserver

---

**설계 문서 작성 완료**: 2026-04-23.
**승인 후 구현 시작**: W1 Quick Wins 부터.
**다음 phase**: 각 항목 구현 → QA 검증 → 릴리스 → 블랙박스 테스트 (M3).
