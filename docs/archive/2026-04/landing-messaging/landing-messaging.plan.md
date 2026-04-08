# Plan: landing-messaging

> README.md + Landing Page 메시징 전면 개편

**Feature**: landing-messaging
**Created**: 2026-04-07
**Phase**: Plan
**Status**: Draft

---

## Executive Summary

| Perspective | Description |
|------------|-------------|
| **Problem** | 현재 랜딩페이지가 기술 중심 메시지로, .env의 AI 보안 위험성이라는 킬러 메시지가 부재. 바이브코더에게 어필 부족 |
| **Solution** | "Your .env is not a secret" 공포 훅 → Runtime injection 해결책 → Cloud waitlist 확장 경로로 메시지 구조 전면 재설계 |
| **Function UX Effect** | README는 기술 개발자 대상 상세 설명, 랜딩페이지는 바이브코더도 즉시 이해할 수 있는 명료한 메시지 |
| **Core Value** | AI 시대 .env의 구조적 보안 취약점을 최초로 짚고, Tene가 유일한 해결책임을 포지셔닝 |

## Context Anchor

| Anchor | Content |
|--------|---------|
| **WHY** | CSO 통화에서 도출된 3대 세일즈 포인트가 현재 랜딩페이지에 반영되지 않음. 전환율 개선 필요 |
| **WHO** | AI 에이전트 사용 개발자 (Claude Code, Cursor, Windsurf), 바이브코더 |
| **RISK** | Cloud pricing이 아직 구현 안 됨 → Fake Door/Waitlist로 처리. 메시지 톤 변경이 기존 사용자 혼란 줄 수 있음 |
| **SUCCESS** | 3대 세일즈 포인트가 랜딩페이지에 명확히 전달됨. README와 랜딩페이지 톤이 각 타겟에 맞게 분리됨 |
| **SCOPE** | README.md (1파일) + apps/web/ 컴포넌트 (~10파일). Go CLI 코드 변경 없음 |

---

## 1. Background

### 1.1 현재 상태

- Hero: "Secret management AI agents understand" — 기능 설명, 긴급성 없음
- Features: Claude Code auto-detect이 1순위 — 기술적이지만 "왜 필요한지" 약함
- Comparison: .env 대비 기능 비교만 있고, .env의 근본 위험성 설명 없음
- Pricing: Free($0)만 표시, 유료 확장 경로 없음
- README: go install이 메인 설치법 → curl로 이미 변경 완료

### 1.2 CSO 세일즈 포인트 (이승준, 2026-04-07)

1. **.env는 secret이 아니다**: AI Agent가 프로젝트 파일을 읽으므로 plaintext .env의 API 키가 AI 컨텍스트에 노출
2. **Runtime injection**: `tene run -- claude` → AI가 secret 값을 모르고도 프로그램 정상 동작
3. **Cloud 확장**: 로컬 무료 → $1/user/mo로 여러 프로젝트/기기에서 매번 init/set 반복 없이 사용

### 1.3 메시지 전략

```
Hook (공포)  → ".env는 AI에게 공개된다"
Solution     → "Tene는 암호화 + Runtime injection으로 해결"
Upgrade Path → "로컬 무료 → Cloud $1/mo로 모든 프로젝트 통합 관리"
```

---

## 2. Requirements

### 2.1 README.md (기술 개발자 대상)

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| R1 | Hero 영역에 .env 위험성 + Tene 해결책 메시지 추가 | P0 |
| R2 | "Why Tene?" 섹션을 3대 세일즈 포인트 중심으로 재작성 | P0 |
| R3 | .env vs Tene 보안 비교 다이어그램 (ASCII/텍스트) | P1 |
| R4 | Cloud 로드맵 간략 언급 (Coming soon) | P2 |
| R5 | Install 섹션은 이미 curl로 변경 완료 — 유지 | - |

### 2.2 Landing Page (바이브코더 대상)

| ID | 요구사항 | 우선순위 |
|----|---------|---------|
| L1 | Hero 카피 전면 교체: "Your .env is not a secret. AI can read it." | P0 |
| L2 | Hero sub: "Tene encrypts secrets so AI agents can use them — without ever seeing them." | P0 |
| L3 | Features 재배치: .env 위험성 → Runtime injection → Encryption → 나머지 | P0 |
| L4 | Terminal 데모: .env 위험성 시연 → tene 해결 시연 대비 | P1 |
| L5 | Pricing 섹션 신규: Free (local) vs Cloud ($1/mo, Coming Soon waitlist) | P0 |
| L6 | Comparison 테이블: "AI에 secret 노출" 행 추가 | P1 |
| L7 | FAQ 업데이트: .env 위험성, Cloud 관련 Q&A 추가 | P1 |
| L8 | CTA: curl 설치 (완료) + "Stop using .env" 메시지 | P1 |

---

## 3. Scope

### 3.1 In Scope

| Target | Files | Description |
|--------|-------|-------------|
| README.md | 1 | Hero + Why Tene 재작성, .env 비교, Cloud mention |
| hero.tsx | 1 | 카피 전면 교체 |
| features.tsx | 1 | 6개 feature 카드 내용/순서 재배치 |
| terminal.tsx | 1 | 데모 스크립트 변경 (.env 위험성 포함) |
| comparison.tsx | 1 | "AI secret 노출" 행 추가, Tene pricing 업데이트 |
| security.tsx | 1 | .env vs Tene 비교 강화 |
| faq.tsx | 1 | Q&A 추가/수정 |
| cta.tsx | 1 | 메시지 변경 |
| how-it-works.tsx | 1 | description 텍스트 보완 |
| pricing 섹션 | 1 | **신규 컴포넌트** — Free vs Cloud + Waitlist form |
| page.tsx | 1 | Pricing 섹션 import 추가 |
| layout.tsx | 1 | SEO 구조화 데이터 업데이트 |

### 3.2 Out of Scope

- Go CLI 코드 변경
- Cloud/Sync 백엔드 구현
- 실제 결제 연동
- 다국어 지원

---

## 4. Messaging Architecture

### 4.1 README.md (기술적 톤)

```markdown
## Why Tene?

### .env files are not secrets

Every AI coding agent — Claude Code, Cursor, Windsurf — reads your project files.
That includes `.env`. Your API keys, database passwords, and tokens are sent
to AI models as plaintext context.

### Tene keeps secrets from AI

Tene stores secrets in an encrypted SQLite vault. When you run `tene run -- claude`,
secrets are injected as environment variables at runtime. The AI agent never sees
the actual values — it only knows to use `tene run`.

### Free locally. $1/mo for cloud.

Local CLI is free forever. Cloud sync ($1/user/month) eliminates repeated
`tene init` + `tene set` across projects and machines. (Coming soon)
```

### 4.2 Landing Page Hero (명료한 톤)

```
H1: "Your .env is not a secret."
H1 accent: "AI can read it."

Sub: "Tene encrypts your secrets so AI agents can use them
      — without ever seeing them."

CTA: curl -sSfL https://tene.sh/install.sh | sh
```

### 4.3 Landing Page Features (재배치)

| 순서 | 제목 | 메시지 | Tag |
|------|------|--------|-----|
| 1 | .env is visible to AI | AI agents read all project files. .env is plaintext. Your secrets are exposed. | Danger |
| 2 | Runtime injection | `tene run` injects secrets as env vars. AI never sees the values. | Solution |
| 3 | Encrypted vault | XChaCha20-Poly1305 + Argon2id. Secrets stored locally, encrypted. | Security |
| 4 | One command setup | Install → init → set → done. No signup, no config. | Speed |
| 5 | .env migration | `tene import .env` converts existing secrets. Zero friction. | Migration |
| 6 | Cloud sync | Manage secrets across projects and machines. $1/user/mo. Coming soon. | Coming Soon |

### 4.4 Pricing Section (신규)

```
┌─────────────────────┬─────────────────────┐
│  Free               │  Cloud              │
│  $0 forever         │  $1/user/month      │
│                     │                     │
│  - Local encrypted  │  - Everything Free  │
│  - Unlimited secrets│  - Cross-project    │
│  - AI runtime inject│  - Cross-machine    │
│  - Single project   │  - Team sharing     │
│                     │  - No repeated init │
│                     │                     │
│  [Install now]      │  [Join waitlist]    │
└─────────────────────┴─────────────────────┘
```

---

## 5. Success Criteria

| ID | Criteria | Measurement |
|----|----------|-------------|
| SC1 | Hero 메시지가 .env 위험성을 첫 화면에서 전달 | hero.tsx에 "Your .env is not a secret" 존재 |
| SC2 | 3대 세일즈 포인트가 Features에 모두 반영 | features.tsx에 .env 위험, runtime injection, cloud 포함 |
| SC3 | Pricing 섹션에 Free vs Cloud 비교 존재 | pricing 컴포넌트 존재 + page.tsx에서 렌더링 |
| SC4 | Cloud CTA가 waitlist form으로 동작 | waitlist 이메일 입력 UI 존재 |
| SC5 | README에 .env 위험성 섹션 추가 | "Why Tene?" 섹션에 AI + .env 언급 |
| SC6 | 빌드 성공 | `npm run build` 에러 없음 |
| SC7 | 바이브코더도 이해할 수 있는 명료한 메시지 | 전문 용어 최소화, 1문장 1개념 |

---

## 6. Risk & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Cloud가 미구현인데 pricing 표시 | 사용자 혼란 | "Coming Soon" + Waitlist로 명확히 표시 |
| 공포 훅이 과하면 신뢰 저하 | 이탈 | 해결책을 바로 제시, 오픈소스 신뢰 강조 |
| Waitlist 이메일 수집 백엔드 필요 | 구현 복잡 | 단순 mailto: 또는 외부 서비스(Formspree 등) 사용 |

---

## 7. Implementation Guide

### Module Map

| Module | Scope | Effort |
|--------|-------|--------|
| M1: README | README.md Why Tene 재작성 | Small |
| M2: Hero + CTA | hero.tsx, cta.tsx 카피 교체 | Small |
| M3: Features | features.tsx 6카드 재배치/재작성 | Medium |
| M4: Terminal | terminal.tsx 데모 스크립트 변경 | Medium |
| M5: Comparison + FAQ | comparison.tsx, faq.tsx 업데이트 | Small |
| M6: Pricing (신규) | pricing.tsx 신규 + page.tsx 연결 | Medium |
| M7: SEO | layout.tsx 구조화 데이터 | Small |

### Recommended Session Plan

- **Session 1**: M1 + M2 + M7 (README + Hero + SEO) — 기반 메시지 확정
- **Session 2**: M3 + M4 + M5 (Features + Terminal + Comparison) — 상세 콘텐츠
- **Session 3**: M6 (Pricing 신규 컴포넌트 + Waitlist)
