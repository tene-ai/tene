# Sprint 1: 기반 구축 — 완료 보고서

> **일자**: 2026-04-08
> **Sprint**: 1/4 (기반 구축)
> **브랜치**: fix/staging-qa
> **Match Rate**: 100%

---

## 요약

| 지표 | 결과 |
|------|------|
| 구현 항목 | 12/12 (100%) |
| 신규 파일 | 11개 (Go 8 + Migration 2 + 없음 1) |
| 수정 파일 | 6개 (Go 3 + TypeScript 2 + 없음 1) |
| Go 빌드 | 통과 |
| Go vet | 통과 |
| TypeScript 체크 | 통과 |
| 기존 테스트 | 14 패키지 전체 통과 |

---

## 구현 내역

### 1. PostgreSQL Repository Layer (8파일)

`internal/repository/postgres/`에 pgx v5 기반 레포지토리 생성:

| 파일 | 인터페이스 | 메서드 수 |
|------|-----------|:--------:|
| postgres.go | DB 래퍼 (New, Close, Ping) | 3 |
| user_repo.go | UserStore + billing.UserStore | 8 |
| vault_repo.go | handler.VaultStore | 7 |
| team_repo.go | handler.TeamStore | 9+ |
| device_repo.go | handler.DeviceStore | 3+ |
| audit_repo.go | handler.AuditStore | 1 |
| refresh_token_repo.go | RefreshTokenStore (신규) | 4 |
| waitlist_repo.go | handler.WaitlistStore | 1 |

### 2. server.go 교체
- `Config.DatabaseURL` 추가
- `NewServer` 시그니처: `(*echo.Echo, func(), error)`
- DatabaseURL 존재 시 PgStore, 비어있으면 MemStore fallback
- `billing.NewService(..., userStore)` — nil 제거

### 3. Migration 추가
- `000009_add_refresh_token_family.up.sql` — family 칼럼 + 인덱스
- `000009_add_refresh_token_family.down.sql` — 롤백

### 4. Dashboard 토큰 갱신
- `auth-store.ts`: refreshToken state 저장, login 시 getMe() 호출
- `api.ts`: 401 인터셉터 (doRefresh + isRefreshing 플래그로 무한루프 방지)

### 5. CLI Keychain 이전
- `login.go`: go-keyring 우선, 파일 fallback, 자동 마이그레이션
- `logout.go`: clearAuthTokens()로 keychain + 파일 모두 삭제

---

## P0 Critical 해결 상태

| ID | 이슈 | 상태 |
|----|------|:----:|
| B-07 | PostgreSQL 레포지토리 미구현 | 해결 |
| A-02 | billing webhook panic (nil UserStore) | 해결 |
| S-01 | Auth token 평문 저장 | 해결 |
| F-08 | Dashboard 토큰 갱신 미구현 | 해결 |

---

## 다음: Sprint 2 (핵심 경험)
- Auth code exchange
- CLI plan 사전 체크
- Landing Pro CTA intent=upgrade
- Dashboard billing 버튼 + overview 라이브 데이터
