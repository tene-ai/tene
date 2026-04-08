# Sprint 2: 핵심 경험 — 완료 보고서

> **일자**: 2026-04-08
> **Sprint**: 2/4 (핵심 경험)
> **Match Rate**: 100%

---

## 요약

| 지표 | 결과 |
|------|------|
| 구현 항목 | 9/9 (100%) |
| 신규 파일 | 1개 (helpers.go) |
| 수정 파일 | 10개 (Go 6 + TypeScript 4) |
| Go 빌드/vet | 통과 |
| TypeScript (Dashboard + Web) | 통과 |
| 기존 테스트 | 12 패키지 전체 통과 |

---

## 구현 내역

### 백엔드

| 항목 | 파일 | 설명 |
|------|------|------|
| Auth Code Exchange | auth.go, server.go | 30초 TTL 임시 코드 → POST /auth/exchange로 토큰 교환. URL 토큰 노출 제거 |
| GET /vaults/:id | vault.go, server.go | Vault 상세 조회 endpoint (Pro plan 필수) |
| GET /teams/:id/members | team.go, server.go | 팀 멤버 목록 endpoint (멤버십 체크) |
| CLI plan 사전 체크 | helpers.go, push.go, pull.go, sync_cmd.go, team.go | JWT payload 로컬 디코딩으로 "Pro 필요" 안내 |

### 프론트엔드

| 항목 | 파일 | 설명 |
|------|------|------|
| Auth callback 변경 | auth/callback/page.tsx | code → exchangeAuthCode() → 토큰 교환 |
| exchangeAuthCode() | api.ts | 신규 public API 메서드 |
| Billing 버튼 연결 | billing/page.tsx | createCheckout() → checkout_url redirect |
| Overview 라이브 | page.tsx | TanStack Query로 vault 개수 실시간 표시 |
| Landing Pro CTA | pricing.tsx, login/page.tsx | intent=upgrade → 자동 billing redirect |

---

## P1 이슈 해결 상태

| ID | 이슈 | 상태 |
|----|------|:----:|
| A-04/S-02 | URL에 토큰 노출 | 해결 (auth code exchange) |
| B-05 | CLI plan 사전 체크 없음 | 해결 (JWT 로컬 디코딩) |
| A-01 | Landing Pro CTA → checkout 미연결 | 해결 (intent=upgrade) |
| F-07 | Billing 버튼 onClick 없음 | 해결 (createCheckout) |
| F-01 | Dashboard 개요 하드코딩 | 해결 (TanStack Query) |

---

## 다음: Sprint 3 (기능 완성)
- Dashboard API client 11개 메서드 추가
- Vault/Team/Device/Audit 페이지 라이브 연동
- Team key rotation 구현
