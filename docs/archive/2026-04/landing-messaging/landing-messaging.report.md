# Report: landing-messaging

> README.md + Landing Page 메시징 전면 개편 — 완료 보고서

**Feature**: landing-messaging
**Created**: 2026-04-07
**Completed**: 2026-04-07
**Phase**: Completed
**Final Match Rate**: 100% (after 1 minor fix iteration)

---

## Executive Summary

| Perspective | Result |
|------------|--------|
| **Problem Solved** | 기존 기술 중심 메시지 → .env AI 위험성 공포 훅 + Tene 해결책 + Cloud 확장 경로로 전면 재설계 |
| **Solution Delivered** | Data/UI 분리 아키텍처로 19개 파일 구현. Hero, Features, Terminal, Pricing, FAQ 등 전 섹션 메시지 교체 |
| **Function UX Effect** | README는 기술 개발자 대상 상세 설명 (ASCII 다이어그램 포함), 랜딩페이지는 바이브코더도 즉시 이해 가능한 명료한 메시지 |
| **Value Delivered** | 3대 세일즈 포인트 100% 반영. Pricing Fake Door + Waitlist로 Cloud 수요 검증 기반 마련 |

### 1.3 Value Delivered

| Metric | Before | After |
|--------|--------|-------|
| Hero 메시지 | "Secret management AI agents understand" (기능 설명) | "Your .env is not a secret. AI can read it." (공포 훅) |
| .env 위험성 언급 | 0곳 | Hero + Features + Terminal + Security + FAQ + README (6곳) |
| Pricing 섹션 | 없음 | Free vs Cloud $1/mo + Waitlist form |
| 데이터/UI 분리 | 0 data files | 5 data files (hero, features, faq, comparison, pricing) |
| SEO .env 메시지 | 없음 | meta, OG, Twitter, JSON-LD 모두 .env 위험성 포함 |

---

## 2. PDCA Cycle Summary

| Phase | Status | Duration | Key Output |
|-------|:------:|----------|------------|
| Plan | Done | - | 7 Success Criteria, 3 Sessions 제안 |
| Design | Done | - | Option B Clean Architecture, 19 파일 설계 |
| Do | Done | - | 7 신규 + 12 수정 파일 구현, 빌드 성공 |
| Check | Done | Match Rate 97.4% → 100% | 1 Minor gap (meta description) |
| Act | Done | 1 iteration | meta description + OG + Twitter 수정 |

---

## 3. Files Changed

### New Files (7)

| File | Purpose |
|------|---------|
| `src/data/hero.ts` | Hero 카피 데이터 |
| `src/data/features.ts` | 6개 Feature 카드 데이터 + tag 타입 |
| `src/data/pricing.ts` | Free/Cloud tier 데이터 |
| `src/data/faq.ts` | 7개 FAQ Q&A 데이터 |
| `src/data/comparison.ts` | 비교 테이블 행 + pricing 데이터 |
| `src/components/pricing.tsx` | Pricing 섹션 (2-column, GlowCard) |
| `src/components/waitlist-form.tsx` | Formspree waitlist + mailto fallback |

### Modified Files (12)

| File | Key Change |
|------|------------|
| `src/components/hero.tsx` | data import, "Your .env is not a secret" + "Download binary" 제거 |
| `src/components/features.tsx` | data import, icon mapping, tag 색상 (Problem=red, Solution=accent) |
| `src/components/terminal.tsx` | .env 위험성 데모 (cat .env → tene import → tene run) |
| `src/components/comparison.tsx` | data import, "Secrets hidden from AI" 첫 행 |
| `src/components/faq.tsx` | data import, .env 위험성 Q&A 우선 배치 |
| `src/components/cta.tsx` | "Stop using .env files." 메시지 |
| `src/components/security.tsx` | .env 위험성 텍스트 추가 |
| `src/components/how-it-works.tsx` | Install description 보완 |
| `src/components/nav.tsx` | Pricing 링크 추가 (desktop + mobile) |
| `src/app/page.tsx` | Pricing import + 렌더링 순서 변경 |
| `src/app/layout.tsx` | SEO 전면 업데이트 (title, description, OG, Twitter, JSON-LD) |
| `README.md` | Why Tene 3 섹션 재작성 + ASCII 다이어그램 |

---

## 4. Success Criteria Final Status

| SC | Criterion | Status | Evidence |
|----|-----------|:------:|----------|
| SC1 | Hero에 "Your .env is not a secret" | Met | `data/hero.ts:4` |
| SC2 | Features에 .env 위험, runtime injection, cloud | Met | `data/features.ts` cards 1, 2, 6 |
| SC3 | Pricing 섹션 존재 | Met | `page.tsx:31` `<Pricing />` |
| SC4 | Waitlist 이메일 입력 UI | Met | `waitlist-form.tsx` Formspree + mailto |
| SC5 | README에 .env 위험성 | Met | `README.md` "### .env files are not secrets" |
| SC6 | 빌드 성공 | Met | `npm run build` — 0 errors |
| SC7 | 바이브코더 이해 가능한 메시지 | Met | 전문용어 최소화, 1문장 1개념 |

**Success Rate: 7/7 (100%)**

---

## 5. Key Decisions & Outcomes

| # | Decision | Source | Followed? | Outcome |
|---|----------|--------|:---------:|---------|
| D1 | Data/UI 분리 (`data/*.ts`) | Design | Yes | 5 data 파일로 메시지 콘텐츠 완전 분리 |
| D2 | Formspree for waitlist | Design | Yes | 서버 코드 0, mailto fallback 포함 |
| D3 | Feature tag 색상 분기 | Design | Yes | Problem=red, Solution=accent, Coming Soon=yellow |
| D4 | Pricing을 Comparison 뒤 배치 | Design | Yes | 자연스러운 비교→가격 전환 흐름 |
| D5 | Terminal .env 위험성 시연 | Design | Yes | cat .env(red) → import(green) → run(accent) |
| D6 | Hero "Download binary" 제거 | Design | Yes | curl + GitHub 2개 CTA로 단순화 |

---

## 6. Architecture Note

Option B (Clean Architecture) 선택으로 `data/` 레이어 도입:
- 향후 메시지 변경 시 data 파일만 수정 (컴포넌트 변경 불필요)
- A/B 테스트, i18n 확장 시 data 파일 교체로 대응 가능
- 현재 5개 data 파일, 2개 신규 컴포넌트, 12개 수정 컴포넌트

---

## 7. Remaining Items

| Item | Priority | Note |
|------|----------|------|
| Formspree 계정 설정 | P1 | `NEXT_PUBLIC_FORMSPREE_ID` Vercel env 설정 필요 |
| OG 이미지 업데이트 | P2 | 현재 이미지가 구 메시지 기반. 새 메시지 반영 필요 |
| install.sh Vercel 배포 | P0 | commit + push 후 자동 배포 |
