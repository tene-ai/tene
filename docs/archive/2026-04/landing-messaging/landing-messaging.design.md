# Design: landing-messaging

> README.md + Landing Page 메시징 전면 개편 — Option B Clean Architecture

**Feature**: landing-messaging
**Created**: 2026-04-07
**Phase**: Design
**Architecture**: Option B — Clean Architecture (data/UI separation)

---

## Context Anchor

| Anchor | Content |
|--------|---------|
| **WHY** | CSO 통화에서 도출된 3대 세일즈 포인트가 현재 랜딩페이지에 반영되지 않음 |
| **WHO** | AI 에이전트 사용 개발자, 바이브코더 |
| **RISK** | Cloud pricing 미구현 → Fake Door/Waitlist 처리 |
| **SUCCESS** | 3대 세일즈 포인트 명확 전달. README/랜딩페이지 톤 분리 |
| **SCOPE** | README.md (1) + apps/web/ (~14파일). Go CLI 변경 없음 |

---

## 1. Overview

Option B: 메시지 데이터를 `data/` 폴더로 분리하여 UI 컴포넌트와 콘텐츠를 완전 분리.
향후 i18n, A/B 테스트, 메시지 변경 시 데이터 파일만 수정하면 됨.

### 1.1 Architecture Decision

| 결정 | 선택 | 이유 |
|------|------|------|
| 데이터 분리 | `data/*.ts` | 메시지 콘텐츠를 UI에서 분리, 향후 확장성 |
| Pricing 컴포넌트 | 독립 `pricing.tsx` | CTA와 역할 분리 명확 |
| Waitlist form | 독립 `waitlist-form.tsx` | 재사용 가능 (Pricing, CTA 등) |
| Nav 업데이트 | Pricing 링크 추가 | 새 섹션 접근성 |

---

## 2. File Structure

```
apps/web/src/
├── data/                          # NEW — 메시지 데이터 분리
│   ├── hero.ts                    # Hero 카피 (h1, sub, cta)
│   ├── features.ts                # 6개 feature 카드 데이터
│   ├── faq.ts                     # FAQ Q&A 데이터
│   ├── comparison.ts              # 비교 테이블 행 데이터
│   └── pricing.ts                 # Pricing tier 데이터
├── components/
│   ├── hero.tsx                   # MODIFY — data import로 전환
│   ├── features.tsx               # MODIFY — data import로 전환
│   ├── terminal.tsx               # MODIFY — 데모 스크립트 변경
│   ├── comparison.tsx             # MODIFY — data import + AI 노출 행 추가
│   ├── security.tsx               # MODIFY — .env 위험성 강화 텍스트
│   ├── faq.tsx                    # MODIFY — data import로 전환
│   ├── cta.tsx                    # MODIFY — "Stop using .env" 메시지
│   ├── how-it-works.tsx           # MODIFY — description 보완
│   ├── nav.tsx                    # MODIFY — Pricing 링크 추가
│   ├── pricing.tsx                # NEW — Free vs Cloud 비교 카드
│   ├── waitlist-form.tsx          # NEW — 이메일 수집 폼
│   └── (기존 유지: glow-card, copy-command, footer, interactive-grid, noise-overlay)
├── app/
│   ├── page.tsx                   # MODIFY — Pricing import 추가
│   └── layout.tsx                 # MODIFY — SEO 구조화 데이터
README.md                          # MODIFY — Why Tene 재작성
```

---

## 3. Data Layer Design

### 3.1 `data/hero.ts`

```typescript
export const heroData = {
  badge: "Open source · Local-first · Free",
  h1: "Your .env is not a secret.",
  h1Accent: "AI can read it.",
  sub: "Tene encrypts your secrets so AI agents can use them — without ever seeing them.",
  cta: {
    install: "curl -sSfL https://tene.sh/install.sh | sh",
    primary: { label: "View on GitHub", href: "https://github.com/tene-ai/tene" },
  },
};
```

### 3.2 `data/features.ts`

```typescript
export type Feature = {
  icon: string;       // SVG icon identifier
  title: string;
  description: string;
  tag: string | null;  // "Danger" | "Solution" | "Coming Soon" | null
};

export const features: Feature[] = [
  {
    icon: "eye",
    title: ".env is visible to AI",
    description: "AI agents read all your project files — including .env. Your API keys are sent to AI models as plaintext.",
    tag: "Problem",
  },
  {
    icon: "inject",
    title: "Runtime injection",
    description: "tene run injects secrets as environment variables. Your app works normally. AI never sees the values.",
    tag: "Solution",
  },
  {
    icon: "lock",
    title: "Encrypted vault",
    description: "XChaCha20-Poly1305 + Argon2id. Secrets stored locally in an encrypted SQLite vault.",
    tag: null,
  },
  {
    icon: "zap",
    title: "One command setup",
    description: "Install, init, set — done. No signup, no config files, no dashboard.",
    tag: null,
  },
  {
    icon: "import",
    title: ".env migration",
    description: "tene import .env converts all your existing secrets into the encrypted vault. Zero friction.",
    tag: null,
  },
  {
    icon: "cloud",
    title: "Cloud sync",
    description: "Manage secrets across all your projects and machines. No more repeated init and set. $1/user/month.",
    tag: "Coming Soon",
  },
];
```

### 3.3 `data/pricing.ts`

```typescript
export type PricingTier = {
  name: string;
  price: string;
  period: string;
  description: string;
  features: string[];
  cta: { label: string; action: "install" | "waitlist" };
  highlighted: boolean;
};

export const pricingTiers: PricingTier[] = [
  {
    name: "Free",
    price: "$0",
    period: "forever",
    description: "Local encrypted secrets for individual projects.",
    features: [
      "Unlimited secrets",
      "XChaCha20-Poly1305 encryption",
      "AI runtime injection",
      "OS keychain integration",
      "12-word recovery key",
    ],
    cta: { label: "Install now", action: "install" },
    highlighted: false,
  },
  {
    name: "Cloud",
    price: "$1",
    period: "per user / month",
    description: "Sync secrets across projects and machines.",
    features: [
      "Everything in Free",
      "Cross-project sync",
      "Cross-machine access",
      "Team sharing",
      "No repeated tene init",
    ],
    cta: { label: "Join waitlist", action: "waitlist" },
    highlighted: true,
  },
];
```

### 3.4 `data/faq.ts`

```typescript
export type FAQ = { question: string; answer: string };

export const faqs: FAQ[] = [
  {
    question: "Why is .env dangerous with AI agents?",
    answer: "AI coding agents like Claude Code, Cursor, and Windsurf read all files in your project directory — including .env. This means your API keys, database passwords, and tokens are sent to AI models as plaintext context. You have no control over how that data is processed or stored.",
  },
  {
    question: "How does Tene keep secrets from AI?",
    answer: "Tene stores secrets in an encrypted SQLite vault (.tene/vault.db). When you run tene run -- claude, secrets are injected as environment variables at runtime. The AI agent sees the tene run command in CLAUDE.md, but never sees the actual secret values.",
  },
  {
    question: "What is Tene?",
    answer: "Tene is a local-first, encrypted secret management CLI. It stores your API keys, tokens, and credentials in an encrypted vault on your device. Single binary, no runtime needed, no server, no signup.",
  },
  {
    question: "How do I install Tene?",
    answer: "Run: curl -sSfL https://tene.sh/install.sh | sh — it auto-detects your OS and installs the latest binary. Works on macOS, Linux, and Windows (WSL). No Go required.",
  },
  {
    question: "Is Tene free?",
    answer: "Yes, Tene CLI is 100% free and open source under the MIT license. Local encrypted secret management has no limits. Cloud sync for teams and multi-project is coming at $1/user/month.",
  },
  {
    question: "What encryption does Tene use?",
    answer: "XChaCha20-Poly1305 for secret encryption with 192-bit random nonces. Argon2id (64MB memory, 3 iterations) for key derivation. Master key stored in your OS keychain. 12-word BIP-39 recovery key.",
  },
  {
    question: "What is Cloud sync?",
    answer: "Cloud sync (coming soon, $1/user/month) lets you manage secrets across multiple projects and machines without running tene init and tene set every time. Your secrets stay encrypted end-to-end.",
  },
];
```

### 3.5 `data/comparison.ts`

```typescript
export type ComparisonRow = {
  feature: string;
  tene: boolean;
  env: boolean;
  doppler: boolean;
  vault: boolean;
  infisical: boolean;
};

export const comparisonRows: ComparisonRow[] = [
  { feature: "Secrets hidden from AI", tene: true, env: false, doppler: false, vault: false, infisical: false },
  { feature: "Local-first", tene: true, env: true, doppler: false, vault: false, infisical: false },
  { feature: "No server required", tene: true, env: true, doppler: false, vault: false, infisical: false },
  { feature: "Encrypted at rest", tene: true, env: false, doppler: true, vault: true, infisical: true },
  { feature: "AI agent auto-detect", tene: true, env: false, doppler: false, vault: false, infisical: false },
  { feature: "Runtime injection", tene: true, env: false, doppler: true, vault: true, infisical: true },
  { feature: "No signup required", tene: true, env: true, doppler: false, vault: false, infisical: false },
  { feature: "Open source", tene: true, env: true, doppler: false, vault: false, infisical: true },
];

export const comparisonPricing = {
  tene: "$0",
  env: "$0",
  doppler: "$21/mo",
  vault: "$1,152+",
  infisical: "$6/mo",
};
```

---

## 4. Component Design

### 4.1 `pricing.tsx` (NEW)

- 2-column 카드 레이아웃 (Free / Cloud)
- Free: accent border, "Install now" → CopyCommand로 curl 명령
- Cloud: highlighted card, "Coming Soon" 배지, "Join waitlist" → WaitlistForm 연결
- 기존 디자인 토큰 사용 (surface, border, accent)
- `GlowCard` 래퍼 재사용
- 모바일: 1-column 스택

### 4.2 `waitlist-form.tsx` (NEW)

- 이메일 input + submit 버튼
- 외부 서비스 사용: Formspree (무료 tier, 월 50건) 또는 단순 `mailto:` fallback
- 제출 후 "You're on the list!" 확인 메시지
- 최소한의 상태: email, submitted, error
- 서버 사이드 코드 불필요 (Next.js API route 안 만듦)

### 4.3 `hero.tsx` (MODIFY)

- `data/hero.ts`에서 import
- H1: "Your .env is not a secret." + accent span "AI can read it."
- Sub: 현재와 동일 구조, 텍스트만 변경
- CTA: curl 명령 (이미 변경 완료) + GitHub 버튼
- "Download binary" 버튼 제거 → GitHub 버튼 하나로 단순화

### 4.4 `features.tsx` (MODIFY)

- `data/features.ts`에서 import
- icon은 SVG를 icon string 기반 매핑 함수로 렌더
- tag 타입에 따라 색상 분기: Problem(red), Solution(accent), Coming Soon(yellow)
- 6카드 순서: Problem → Solution → Security → Speed → Migration → Cloud

### 4.5 `terminal.tsx` (MODIFY)

기존 데모에 .env 위험성 시연 추가:

```
$ cat .env
  STRIPE_KEY=sk_test_51Hxxxxx     ← AI can see this
  OPENAI_API_KEY=sk-proj-xxxxx    ← AI can see this

$ tene import .env
  ✓ 2 secrets imported and encrypted
  ✓ .env can now be deleted

$ tene run -- claude
  ✓ 2 secrets injected as environment variables
  ✓ Starting: claude
  ✓ AI cannot see secret values
```

### 4.6 `comparison.tsx` (MODIFY)

- `data/comparison.ts`에서 import
- 첫 행: "Secrets hidden from AI" — Tene만 체크, 나머지 모두 X
- pricing 행도 data에서 import

### 4.7 `faq.tsx` (MODIFY)

- `data/faq.ts`에서 import
- 첫 2개 Q&A가 .env 위험성 관련 (순서 재배치)

### 4.8 `cta.tsx` (MODIFY)

- H2: "Stop using .env files."
- Sub: "Encrypt your secrets. Inject at runtime. AI never sees them."
- curl CTA (이미 변경 완료)

### 4.9 `nav.tsx` (MODIFY)

- "Pricing" 링크 추가 (desktop + mobile)

### 4.10 `security.tsx` (MODIFY)

- 서브 텍스트에 .env 위험성 한 줄 추가: "While .env files expose secrets to every AI agent in your project, Tene ensures..."

### 4.11 `how-it-works.tsx` (MODIFY)

- Step 1 description 업데이트 (이미 부분 변경됨, data 분리는 안 함 — 3 step 고정)

### 4.12 `page.tsx` (MODIFY)

- Pricing import + 렌더링 순서: Hero → Terminal → Features → HowItWorks → Security → Comparison → Pricing → FAQ → CTA

### 4.13 `layout.tsx` (MODIFY)

- SEO 구조화 데이터: description 업데이트, Pricing 정보 추가

---

## 5. README.md Design

### 5.1 구조 변경

```markdown
# Tene
(badges — 유지)
(OG image — 유지)
(한줄 소개 — 변경)

## Why Tene?

### .env files are not secrets
(AI가 .env를 읽는 문제 설명)

### Tene keeps secrets from AI
(Runtime injection 해결책)

### Free locally. $1/mo for cloud.
(Cloud 확장 경로)

## Install
(curl — 이미 변경 완료, 유지)

## Quick Start
(기존 유지)

## How It Works
(기존 유지, .env 위험성 ASCII 다이어그램 추가)

(나머지 섹션 유지)
```

### 5.2 .env vs Tene 비교 다이어그램

```
## The Problem with .env

  .env (plaintext)          AI Agent (Claude, Cursor...)
  ┌──────────────┐         ┌──────────────────────┐
  │ STRIPE_KEY=  │────────>│ Reads all project     │
  │ sk_test_xxx  │         │ files including .env  │
  │              │         │                       │
  │ DB_PASS=     │────────>│ Secrets sent to AI    │
  │ s3cur3p@ss   │         │ model as plaintext    │
  └──────────────┘         └──────────────────────┘

## How Tene Solves This

  .tene/vault.db (encrypted)    tene run -- claude
  ┌──────────────────┐         ┌──────────────────────┐
  │ ████████████████ │         │ Secrets injected as   │
  │ ████████████████ │───X───> │ env vars at runtime   │
  │ (XChaCha20-Poly) │         │                       │
  └──────────────────┘         │ AI sees: tene run     │
                               │ AI gets: $ENV_VARS    │
                               │ AI knows: nothing     │
                               └──────────────────────┘
```

---

## 6. Waitlist Implementation

### 6.1 Options 비교

| Option | Pros | Cons |
|--------|------|------|
| Formspree | 무료 50건/월, API 호출만 | 외부 의존성 |
| mailto: link | 제로 의존성 | UX 나쁨, 이메일 앱 열림 |
| Google Forms embed | 무료, 무제한 | 디자인 커스텀 어려움 |
| Next.js API + DB | 완전 제어 | Over-engineering |

### 6.2 선택: Formspree

- 가입 후 form endpoint 1개 생성
- Client-side fetch POST로 제출
- 서버 코드 불필요
- 무료 tier 충분 (초기 수요 검증 단계)
- Fallback: Formspree 미설정 시 `mailto:` 링크로 graceful degradation

### 6.3 Form 스펙

```typescript
// POST https://formspree.io/f/{form_id}
// Content-Type: application/json
// Body: { email: string }
// Response: 200 OK | 422 Validation Error
```

환경변수: `NEXT_PUBLIC_FORMSPREE_ID` (빈 값이면 mailto fallback)

---

## 7. Section Order

```
page.tsx 렌더링 순서:

Nav
├── Hero              ← "Your .env is not a secret. AI can read it."
├── Terminal          ← .env 위험성 시연 → tene 해결 시연
├── Features          ← 6카드 (Problem → Solution → ...)
├── HowItWorks       ← Install → Init → Set → Run
├── Security          ← 암호화 아키텍처 (기존 + .env 언급)
├── Comparison        ← 기존 + "Secrets hidden from AI" 행
├── Pricing           ← NEW: Free vs Cloud
├── FAQ               ← .env 위험성 Q&A 우선
├── CTA               ← "Stop using .env files."
Footer
```

---

## 8. Test Plan

| ID | Test | Method |
|----|------|--------|
| T1 | `npm run build` 성공 | CLI |
| T2 | 모든 data import가 정상 resolve | 빌드 시 TypeScript 검증 |
| T3 | Pricing 섹션 렌더링 확인 | 브라우저 검증 |
| T4 | Waitlist form 제출 동작 | Formspree test mode |
| T5 | 모바일 반응형 확인 | 브라우저 resize |
| T6 | README 마크다운 렌더링 확인 | GitHub preview |

---

## 9. Dependencies

- Formspree 계정 (waitlist form용) — 선택사항, 없으면 mailto fallback
- `NEXT_PUBLIC_FORMSPREE_ID` 환경변수 — Vercel에 설정 또는 빈 값

신규 npm 패키지: 없음

---

## 10. Design Decisions Log

| # | Decision | Rationale |
|---|----------|-----------|
| D1 | 데이터를 `data/*.ts`로 분리 | UI 변경 없이 메시지 수정 가능. 향후 A/B 테스트, i18n 대비 |
| D2 | Formspree for waitlist | 서버 코드 0. 무료. API route 불필요 |
| D3 | Feature tag 색상: Problem=red | .env 위험성을 시각적으로 강조. Solution=accent로 대비 |
| D4 | Pricing을 Comparison 뒤에 배치 | 비교 후 가격 → 자연스러운 전환 흐름 |
| D5 | Terminal에 .env 위험성 시연 추가 | "보여주기"가 "말하기"보다 강력. cat .env → tene import → tene run |
| D6 | "Download binary" 버튼 제거 | curl이 메인. 버튼 2개는 선택 피로 |

---

## 11. Implementation Guide

### 11.1 Implementation Order

1. `data/*.ts` 5개 파일 생성 (메시지 데이터 확정)
2. `pricing.tsx` + `waitlist-form.tsx` 신규 컴포넌트
3. `hero.tsx` → data import + 카피 교체
4. `features.tsx` → data import + icon 매핑 + tag 색상
5. `terminal.tsx` → 데모 스크립트 변경
6. `comparison.tsx` → data import + AI 노출 행
7. `faq.tsx` → data import
8. `cta.tsx`, `security.tsx`, `how-it-works.tsx` → 텍스트 변경
9. `nav.tsx` → Pricing 링크
10. `page.tsx` → Pricing import + 순서 변경
11. `layout.tsx` → SEO 업데이트
12. `README.md` → Why Tene 재작성

### 11.2 Key File Changes

| File | Action | Key Change |
|------|--------|------------|
| `data/hero.ts` | CREATE | Hero 카피 데이터 |
| `data/features.ts` | CREATE | 6 feature 카드 + tag |
| `data/faq.ts` | CREATE | 7 Q&A |
| `data/comparison.ts` | CREATE | 비교 행 + pricing |
| `data/pricing.ts` | CREATE | 2 tier 데이터 |
| `pricing.tsx` | CREATE | 2-column 카드 |
| `waitlist-form.tsx` | CREATE | Formspree email form |
| `hero.tsx` | MODIFY | data import, 카피 교체 |
| `features.tsx` | MODIFY | data import, icon map, tag 색상 |
| `terminal.tsx` | MODIFY | .env 시연 스크립트 |
| `comparison.tsx` | MODIFY | data import, AI 행 추가 |
| `faq.tsx` | MODIFY | data import |
| `cta.tsx` | MODIFY | "Stop using .env" |
| `security.tsx` | MODIFY | .env 위험성 추가 |
| `how-it-works.tsx` | MODIFY | description 보완 |
| `nav.tsx` | MODIFY | Pricing 링크 |
| `page.tsx` | MODIFY | Pricing import + 순서 |
| `layout.tsx` | MODIFY | SEO 데이터 |
| `README.md` | MODIFY | Why Tene 재작성 |

### 11.3 Session Guide

| Session | Modules | Files | Description |
|---------|---------|-------|-------------|
| S1 | Data Layer | 5 data files | 메시지 데이터 확정 — 이후 모든 작업의 기반 |
| S2 | New Components | pricing.tsx, waitlist-form.tsx | 신규 컴포넌트 생성 |
| S3 | Core Messaging | hero, features, terminal, cta | 핵심 메시지 컴포넌트 교체 |
| S4 | Supporting | comparison, faq, security, how-it-works, nav | 보조 섹션 업데이트 |
| S5 | Integration | page.tsx, layout.tsx, README.md | 통합 + SEO + 빌드 검증 |
