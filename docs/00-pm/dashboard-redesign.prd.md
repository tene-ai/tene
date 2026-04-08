# Dashboard Redesign PRD
## Tene Cloud Dashboard 전면 개편 — Zero-Knowledge 보안 관제탑

> **Version**: v1.0
> **Date**: 2026-04-08
> **Feature**: dashboard-redesign
> **Status**: PRD Complete — Plan Phase Ready
> **PM Agent Team**: pm-research, pm-strategy, codebase-analysis (3-agent synthesis)

---

## Executive Summary

| Perspective | Description |
|-------------|-------------|
| **Problem** | 현재 대시보드는 CLI의 "결과 뷰어"에 불과. 사용자 정보/로그아웃 없음, CLI↔Dashboard 단절, 무료 사용자 진입 경로 결제뿐, 메뉴 구성이 사용 빈도와 불일치 |
| **Solution** | 대시보드를 "Zero-Knowledge 보안 관제탑"으로 재정의. 사용자 프로필, 온보딩 체크리스트, 프로젝트 중심 IA, Cmd+K 커맨드 팔레트로 CLI-native 경험 제공 |
| **Functional UX Effect** | 로그인 → 온보딩 가이드 → Vault sync 상태 실시간 확인 → 팀 초대/권한 관리 → Audit 모니터링을 대시보드 한 곳에서 완결 |
| **Core Value** | 경쟁사(Doppler $10+, Infisical $18+) 대비 $5로 ZK 암호화 감사 + AI 에이전트 인식 + 팀 관제를 제공하는 개발자 도구 대시보드 |

---

## Context Anchor

| Dimension | Content |
|-----------|---------|
| **WHY** | 현재 대시보드가 CLI 보조 도구에 머물러 독자적 가치 부재. Pro 전환율과 사용자 만족도 저하의 원인 |
| **WHO** | 1차: 솔로 개발자(Free→Pro 전환 대상), 2차: 팀 리드(Pro 핵심 고객), 3차: DevOps/SRE(감사/환경 관리) |
| **RISK** | 과도한 기능 추가로 미니멀 정체성 상실, ZK 원칙 위반, 개편 기간 장기화 |
| **SUCCESS** | 온보딩 완료율 80%+, Free→Pro 전환율 15%+, 대시보드 주간 활성 사용자 증가 |
| **SCOPE** | Dashboard 전면 개편 (Next.js 15 유지, 기존 API 활용, 신규 API 최소화) |

---

## 1. 핵심 사용자 피드백 (QA 테스트 발견)

| # | 피드백 | 심각도 | 영향 |
|---|--------|:------:|------|
| F-01 | CLI와 대시보드 왔다갔다 불편 | High | 일상 사용 마찰 |
| F-02 | 로컬 .tene 폴더에서만 push 가능 | Medium | 신규 사용자 혼란 |
| F-03 | 사용자 정보/로그아웃 없음 | Critical | SaaS 기본 요건 미충족 |
| F-04 | 페이지/메뉴 구성 비효율 | Medium | 핵심 기능 접근 저해 |
| F-05 | 랜딩→대시보드 진입이 결제뿐 | High | Free 사용자 이탈 |

---

## 2. 경쟁사 대시보드 벤치마크

| 경쟁사 | 메뉴 구조 | 온보딩 | 핵심 배울 점 |
|--------|----------|--------|------------|
| **Doppler** | Project→Config 2-depth | 5분 온보딩, 30+ 통합 | 프로젝트 중심 계층 구조, 환경별 시크릿 분리 UI |
| **Infisical** | 프로젝트→환경→시크릿 3-depth | 웹 UI에서 시크릿 직접 관리 | 환경별 diff 뷰, 시크릿 버전 히스토리 |
| **Vercel** | 팀/프로젝트 통합 네비게이션 | 프로젝트 연결 중심 | 2026년 사이드바 리디자인, 컨텍스트 전환 최소화 |
| **Linear** | Cmd+K 중심, 키보드 네이티브 | 템플릿 기반 시작 | 커맨드 팔레트, 밀리초 반응성, 미니멀 디자인 |
| **1Password** | 볼트→아이템 계층 | 익숙한 UX | 브랜드 신뢰도, 팀 공유 UX |

---

## 3. 2026 트렌드 적용 방향

| 트렌드 | Tene 적용 |
|--------|----------|
| **Progressive Disclosure** | 첫 방문: 온보딩 체크리스트 → 숙련: 프로젝트 중심 뷰 |
| **Command Palette (Cmd+K)** | CLI 명령어를 대시보드에서 검색/실행 가이드 |
| **Dark + Neon** | 현재 #0a0a0a + #00ff88 유지 (트렌드 정합) |
| **Empty State = 교육** | 상태 기반 동적 안내 (CLI 미설치 / push 안 함 / 팀 미설정) |
| **AI 통합** | CLAUDE.md 생성 상태, AI 에이전트 접근 이벤트 표시 |

---

## 4. 개편 제안: 메뉴 & IA 재설계

### 현재 (6개 동등 메뉴)
```
Overview | Vaults | Devices | Team | Audit Log | Billing
```

### 제안 (4개 메뉴 + 사용자 프로필)
```
Projects | Team | Activity | Settings
                              [avatar ▾] → Profile / Plan / Sign out
```

| 현재 | 개선 | 이유 |
|------|------|------|
| Overview | 제거 (Projects가 메인) | 하드코딩 stat만 보여주는 중복 페이지 |
| Vaults | **Projects** | 사용자 친화적 용어 (Doppler 방식) |
| Devices | Settings 하위 | 사용 빈도 낮음, 설정 성격 |
| Team | **Team** 유지 | 핵심 Pro 기능 |
| Audit Log | **Activity** | 친근한 이름, 범위 확장 |
| Billing | 사용자 메뉴 하위 | 월 1회 접근, 프로필과 함께 |
| ❌ 없음 | **사용자 프로필 메뉴** | 로그아웃, 계정 정보, Plan 상태 |

---

## 5. 핵심 페이지 개편 방향

### 5.1 Projects 페이지 (현재 Vaults)

Doppler/Infisical 방식의 프로젝트 → 환경 → 시크릿 키 계층:
```
Projects
├── my-backend          3 envs · 12 secrets · synced 2m ago
│   ├── development     8 secrets
│   ├── staging         10 secrets
│   └── production      12 secrets
└── my-frontend         1 env · 3 secrets · synced 1h ago
    └── default         3 secrets
```

Zero-Knowledge 유지: 키 이름만 표시, 값은 절대 안 보임.

### 5.2 온보딩 (첫 방문 경험)

```
Welcome to Tene Cloud!
✅ 1. GitHub 계정 연결됨
⬜ 2. CLI 설치 → curl -fsSL tene.sh/install.sh | sh
⬜ 3. 첫 프로젝트 push → tene init && tene push
⬜ 4. 두 번째 디바이스 연결 (선택)
```

### 5.3 사용자 프로필 (우측 상단)

```
[avatar ▾]
├── agent-kay (k99402802@gmail.com)
├── Plan: Free → Upgrade to Pro
├── ─────────
├── Devices (2)
├── Billing
└── Sign out
```

---

## 6. 페르소나별 핵심 Job

| 페르소나 | 핵심 Job | 대시보드 역할 |
|---------|---------|------------|
| **솔로 개발자** | vault sync 상태 확인, 멀티 디바이스 복구 | 상태 모니터링 + CLI 가이드 |
| **팀 리드** | 팀원 접근 권한 관리, 보안 감사 | 팀 관리 + Audit 분석 |
| **DevOps/SRE** | 환경별 시크릿 현황, CI 봇 관리 | 환경 탭 + 디바이스 분류 |

---

## 7. Success Metrics

| Metric | 현재 | 목표 |
|--------|------|------|
| 온보딩 체크리스트 완료율 | N/A | 80%+ |
| Free→Pro 전환율 | N/A (진입 경로 없음) | 15%+ |
| 대시보드 주간 재방문율 | N/A | 40%+ |
| 첫 push까지 시간 | 불명 | 5분 이내 |
| 사용자 정보/로그아웃 접근 | ❌ 불가 | ✅ 1클릭 |

---

## 8. 구현 우선순위

### Phase 1: Quick Wins (1주)
- 사용자 프로필 메뉴 (아바타 + 로그아웃 + Plan 상태)
- Overview 통계 동적 데이터 바인딩
- Empty State 개선 (상태 기반 동적 안내)
- 랜딩 페이지에 "Dashboard" 링크 추가

### Phase 2: IA 재편 (2주)
- 메뉴 재구성: Projects / Team / Activity / Settings
- Projects 페이지 (환경별 시크릿 키 목록)
- Activity 페이지 (필터링 강화 + Pagination)
- Settings 페이지 (Devices + API Keys)

### Phase 3: 고급 기능 (1개월)
- 온보딩 체크리스트 (단계별 완료 추적)
- Cmd+K 커맨드 팔레트
- 팀 초대 직접 처리 (CLI 의존도 제거)
- 모바일 최적화
