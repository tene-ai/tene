# Tene Strategy Analysis v4
## "Claude Code가 자동 인식하는 시크릿 관리" Go CLI 포지셔닝 + $0 MVP 전략

> v4 (2026-04-06) — Go 언어 전환 + Claude Code 전용 MVP 반영, 단일 바이너리 배포
> PM Agent: pm-strategy | Frameworks: JTBD 6-Part Value Proposition, Lean Canvas, SWOT, Porter's 5 Forces
> Architecture: Go CLI Local-Only MVP ($0) → Cloud Phase 2 (수요 검증 후)

---

## 1. JTBD (Jobs-to-be-Done) 분석 (v4)

### 1.1 Core Job Statement (v4)

> **Claude Code와 바이브코더를 사용하여 소프트웨어를 개발할 때,**
> **시크릿(API 키, 토큰)을 서버 없이 로컬에서 안전하게 관리하고,**
> **Claude Code가 시크릿 사용법을 자동으로 인식하여,**
> **설정 없이 즉시 안전한 개발을 시작하고 싶다.**

v3와의 핵심 차이: "AI 에이전트" → "Claude Code" 집중, Go 단일 바이너리로 설치 마찰 제거

### 1.2 Job Map (v4 — Claude Code 자동 인식 + Go CLI 관점)

| 단계 | Job Step | 현재 Pain (기존 도구) | Tene v4 |
|------|----------|---------------------|---------|
| 1. 설치 | 도구를 설치하고 시작 | 서버 가입, 이메일 인증, Node.js 필요 | `brew install tomo-kay/tap/tene` → 끝. 가입 없음, 런타임 없음 |
| 2. 초기화 | 프로젝트에 적용 | .env 파일 생성 | `tene init` → 볼트 생성 + **CLAUDE.md 자동 생성** |
| 3. AI 인식 | AI가 도구를 인식 | **수동 설정 필요** | **자동 인식** (CLAUDE.md) |
| 4. 저장 | 시크릿 안전 보관 | 서버로 전송 | `tene set KEY VALUE` → 로컬 암호화 |
| 5. 접근 | 시크릿 사용 | 서버에서 조회 | `tene get KEY` → 로컬에서 즉시 (~5ms) |
| 6. 주입 | 개발 환경에 적용 | `doppler run --` | `tene run -- claude` → 로컬에서 즉시 |
| 7. 백업 | 데이터 보호 | 서버가 관리 | `tene export --encrypted` 수동 백업 |
| 8. 동기화 | 멀티 디바이스 | 서버 중심 | Phase 2 (수요 검증 후) |

### 1.3 JTBD 6-Part Value Proposition (v4)

| 파트 | v4 내용 |
|------|---------|
| **1. Job Performer** | 솔로 바이브코더, Claude Code 사용자 (서버/가입에 거부감이 있는 개발자) |
| **2. Core Functional Job** | 시크릿을 서버 없이 로컬에서 암호화 관리 + Claude Code 자동 연동 |
| **3. Related Jobs** | AI 도구 설정, 프로젝트 환경 관리, 기존 .env 마이그레이션 |
| **4. Emotional Jobs** | "AI가 알아서 시크릿을 안전하게 쓰니까 안심", "서버에 시크릿을 맡기고 싶지 않다" |
| **5. Desired Outcomes** | (1) 5초 내 설치 (brew) (2) Claude Code가 자동 인식 (3) 서버 없이 100% 안전 |
| **6. Constraints** | (1) 서버 가입 거부 (2) 복잡한 설정 거부 (3) 오프라인 환경 지원 필수 (4) Node.js 설치 부담 |

---

## 2. Lean Canvas (v4)

```
+--------------------+--------------------+--------------------+
|  1. PROBLEM         |  4. SOLUTION        |  3. UVP            |
|                     |                     |                     |
| - AI 에이전트가     | - Go 단일 바이너리  | "Claude Code가      |
|   시크릿 도구를     |   CLI + SQLite     |  자동으로 인식하는   |
|   자동 인식 못 함   |   로컬 볼트 (무료)  |  시크릿 관리.       |
|                     |                     |  서버 없이. 무료.   |
| - 시크릿 관리 도구가 | - tene init →      |  5ms 시작."        |
|   서버/가입을 요구   |   CLAUDE.md 자동    |                     |
|                     |   생성              |  High-Level         |
| - CLI 설치에 Node.js| - Master Password  |  Concept:           |
|   런타임 필요       |   + XChaCha20-     |  "AI가 인식하는     |
|                     |   Poly1305         |  시크릿의 Git"      |
|  Existing Alt.:     |                     |                     |
|  .env, Vault,       | - tene export      |                     |
|  Doppler, Infisical |   --encrypted      |                     |
|                     |   수동 백업         |                     |
+--------------------+--------------------+--------------------+
|  8. KEY METRICS     |                     |  9. UNFAIR ADV.     |
|                     |                     |                     |
| - brew/curl 설치 수 |                     | - Claude Code 자동  |
| - GitHub Stars      |                     |   인식 (유일)       |
| - tene sync         |                     | - 서버 없음 = 해킹  |
|   Fake Door 클릭수  |                     |   대상 없음         |
| - waitlist 등록 수   |                     | - Go 단일 바이너리  |
| - 시크릿 주입 횟수  |                     |   의존성 제로       |
|                     |                     | - 오픈소스 + 로컬   |
|                     |                     |   = 최강의 신뢰     |
+--------------------+--------------------+--------------------+
|  7. COST            |                     |  6. REVENUE         |
|                     |  2. CUSTOMER        |                     |
| MVP (로컬 무료):    |  SEGMENTS           | Phase 1 (MVP):      |
| - 인프라 비용: $0   |                     | - $0 (무료)         |
| - 개발: 파운더 1인  | - 1차: 솔로         |                     |
|                     |   바이브코더        | Phase 2 (수요 확인  |
| Phase 2 (Cloud):    |   (Claude Code)     |  후 Cloud):         |
| - ECS Fargate       | - 2차: 멀티 디바이스| - $1/월 (가설,     |
| - NLB + RDS Postgres|   사용 개발자       |   수요에 따라 조정) |
| - S3               | - 3차: AI-first     |                     |
| - 마케팅: $0        |   소규모 팀         | - Team: 가격 미정   |
|   (오픈소스+커뮤니티)|   (Phase 2+)        |   (Phase 2+, 가설)  |
|                     |                     |                     |
|  5. CHANNELS        |                     |                     |
| - Homebrew tap      |                     |                     |
| - GitHub 오픈소스   |                     |                     |
| - Dev 커뮤니티      |                     |                     |
| - 입소문            |                     |                     |
+--------------------+--------------------+--------------------+
```

### v3 vs v4 Lean Canvas 핵심 변화

| 항목 | v3 | v4 |
|------|----|----|
| Problem | AI 인식 못 함, 서버/가입 요구 | + **CLI 설치에 Node.js 런타임 필요** |
| Solution | CLI + SQLite 로컬 볼트 + CLAUDE.md | **Go 단일 바이너리** + SQLite + CLAUDE.md |
| UVP | "AI 에이전트가 자동 인식" | **"Claude Code가 자동 인식. 5ms 시작."** |
| Channels | npm | **Homebrew tap + curl 설치 + go install** |
| Key Metrics | npm 설치, Fake Door | **brew/curl 설치, GitHub Releases 다운로드** |
| Unfair Advantage | AI 자동 인식 + 서버 없음 | + **Go 단일 바이너리 의존성 제로** |

---

## 3. SWOT 분석 (v4 — Claude Code 전용 + Go CLI 관점)

### 3.1 SWOT Matrix

| | **Helpful (긍정적)** | **Harmful (부정적)** |
|---|---|---|
| **Internal** | **Strengths** | **Weaknesses** |
| | S1: **Claude Code 자동 인식** (CLAUDE.md) — 유일 | W1: Cloud 기능 부재 (MVP) |
| | S2: **서버 없음 = 최강 보안** (해킹 대상 자체 부재) | W2: 브랜드 인지도 제로 |
| | S3: **$0 서버 비용** = 무료 제공 가능 | W3: 1인 팀으로 개발 속도 제한 |
| | S4: 오프라인 100% 동작 | W4: Master Password 분실 시 복구 불가 |
| | S5: 오픈소스 + 로컬 = 최강의 신뢰 조합 | W5: 멀티 디바이스 동기화 부재 (Phase 2) |
| | S6: **Go 단일 바이너리** (~5ms 시작, 의존성 제로) | W6: API key 만료 확인/자동 로테이션 불가 |
| | | W7: MVP에서 Cursor/Windsurf 미지원 |
| **External** | **Opportunities** | **Threats** |
| | O1: Local-First 소프트웨어 트렌드 급성장 | T1: 1Password Unified Access (강력한 브랜드+자본) |
| | O2: 바이브코딩 보안 위기 (시크릿 누출 81% 급증) | T2: Infisical (오픈소스 + AI Agent Sentinel) |
| | O3: CLAUDE.md 생태계 확대 | T3: Dotenvx AS2 ("Secrets for agents" 직접 경쟁) |
| | O4: 개발자 클라우드 피로감 증가 | T4: "로컬 전용은 불편하다"는 인식 |
| | O5: GitHub Secrets가 로컬 개발 미지원 | T5: ".env로 충분하다"는 관성 |
| | O6: Go CLI 생태계 성숙 (cobra, goreleaser) | T6: Claude Code 전용 집중 → Cursor 사용자 이탈 |

### 3.2 SO/WT 전략 (v4)

**SO 전략 (강점으로 기회 활용)**

- **SO1**: Claude Code 자동 인식(S1) + CLAUDE.md 생태계(O3) = **"Claude Code가 자동으로 인식하는 유일한 시크릿 도구"** 포지셔닝
- **SO2**: 서버 없음(S2) + 바이브코딩 보안 위기(O2) = **"서버가 없으면 해킹도 없다"** 캠페인
- **SO3**: 오프라인 100%(S4) + 클라우드 피로감(O4) = Local-First 커뮤니티에서 시크릿 관리 표준
- **SO4**: 무료(S3) + GitHub Secrets 미지원(O5) = **"GitHub Secrets는 CI/CD, Tene는 로컬 개발"** 보완 포지셔닝
- **SO5**: Go 단일 바이너리(S6) + Go CLI 생태계(O6) = **극단적 간편 설치** (brew 한 줄)

**WT 전략 (약점 보완 + 위협 대응)**

- **WT1**: Cloud 부재(W1) + 경쟁사(T1,T2) = `tene export --encrypted`로 수동 백업 제공, Cloud는 수요 확인 후
- **WT2**: 로테이션 불가(W6) + Dotenvx AS2(T3) = 정직하게 "못 하는 것" 명시, Phase 2 가설로 관리
- **WT3**: 1인 팀(W3) + ".env 관성"(T5) = Claude Code 자동 인식이라는 .env에 없는 가치로 전환 유도
- **WT4**: Cursor 미지원(W7) + Cursor 이탈(T6) = Phase 2에서 Cursor/Windsurf 지원 추가, MVP는 Claude Code 완성도 집중
- **WT5**: 인지도 부재(W2) = 오픈소스 + Homebrew tap으로 개발자 커뮤니티에서 자연 확산

---

## 4. Porter's 5 Forces 분석 (v4)

### 4.1 산업 경쟁 강도

```
신규 진입 위협: 높음 (4/5) -- 로컬 CLI 도구 진입 장벽 낮음
          |
          v
공급자 교섭력 --> 산업 경쟁 강도 <-- 구매자 교섭력
   낮음 (1/5)       중간 (3/5)       높음 (4/5)
          ^          (Claude Code      (무료이면 전환
          |           자동 인식 니치는   비용 제로)
          |           경쟁 없음)
대체재 위협: 높음 (4/5) -- .env가 여전히 대체재
```

### 4.2 상세 분석

#### Force 1: 기존 경쟁자 간 경쟁 (3/5 - 중간)

| 요인 | v4 분석 |
|------|---------|
| **범용 시크릿 관리** | Vault, Doppler, Infisical — 레드오션 |
| **Claude Code 자동 인식** | **경쟁자 0명**. 완전한 블루오션 |
| **로컬 전용 시크릿** | Dotenvx만 유사 |
| **시사점** | **Claude Code 자동 인식 니치에서는 경쟁이 없다** |

#### Force 2: 신규 진입 위협 (4/5 - 높음)

| 요인 | v4 분석 |
|------|---------|
| 기술 장벽 | 낮음 — CLAUDE.md 생성은 기술적으로 단순 |
| 진입 비용 | 매우 낮음 — 서버 비용 $0 |
| 방어막 | **커뮤니티 + 브랜드 + "처음 인식된 도구"의 선점 효과** |
| 시사점 | **빠른 실행이 핵심. 먼저 CLAUDE.md 생태계에 진입해야** |

#### Force 3: 대체재 위협 (4/5 - 높음)

| 대체재 | 위협 수준 | v4 관점 |
|--------|:---------:|---------|
| .env 파일 | 매우 높음 | "무료이고 로컬이고 익숙" — 가장 큰 대체재 |
| GitHub Secrets | 중간 | **CI/CD 전용이므로 Tene과 겹치지 않음** |
| OS Keychain | 중간 | 로컬이지만 AI 에이전트 통합 불편 |
| Dotenvx | 높음 | 암호화된 .env — 유사하지만 AI 자동 인식 없음 |

#### Force 4: 공급자 교섭력 (1/5 - 매우 낮음)

| 요인 | v4 분석 |
|------|---------|
| 로컬 전용 | **서버 비용 $0** — 클라우드 벤더 종속 없음 |
| 암호화 기술 | 오픈소스 (golang.org/x/crypto, XChaCha20-Poly1305) |
| 배포 | Homebrew + goreleaser — 벤더 독립 |
| 시사점 | **공급자 리스크가 0. MVP 비용 $0** |

#### Force 5: 구매자 교섭력 (4/5 - 높음)

| 요인 | v4 분석 |
|------|---------|
| 가격 | 무료 → 가격 교섭 불필요 |
| 전환 비용 | 낮음 — `tene export` → .env로 쉽게 복귀 |
| 시사점 | **Claude Code 자동 인식 습관이 유일한 전환 비용** |

---

## 5. 경쟁사 포지셔닝 맵 (v4 — Claude Code 통합 축)

### 5.1 2축 포지셔닝 맵: AI Agent 자동 인식 vs 사용 편의성

```
     AI Agent 자동 인식 (높음)
              |
              |  *** Tene v4 (목표 포지션)
              |
              |
     ---------+------------------------------------------
              |
              |  * Dotenvx (AS2)    * 1Password (Unified)
              |
              |  * Infisical         * Doppler
              |    (Sentinel)
              |
              |  * Vault (MCP)
     AI Agent 자동 인식 (없음)
              
   낮은 사용 편의성          높은 사용 편의성
```

### 5.2 포지셔닝 점수 비교 (v4)

| 제품 | AI 자동 인식 | 로컬 자율성 | 사용 편의성 | 오픈소스 | 서버 불필요 | 설치 간편성 |
|------|:----------:|:----------:|:----------:|:-------:|:----------:|:----------:|
| **Tene v4** | **10/10** | **10/10** | **9/10** | **Yes** | **Yes** | **10/10** (brew) |
| Dotenvx | 2/10 | 8/10 | 6/10 | Yes | Partial | 7/10 |
| Vault | 1/10 | 7/10 | 3/10 | Partial | 셀프호스팅 | 3/10 |
| Infisical | 2/10 | 6/10 | 7/10 | Yes | 셀프호스팅 | 5/10 |
| 1Password | 3/10 | 2/10 | 8/10 | No | No | 6/10 |
| Doppler | 0/10 | 1/10 | 7/10 | No | No | 5/10 |
| GitHub Secrets | 0/10 | 0/10 | 7/10 | No | No | N/A |

### 5.3 전략적 포지셔닝 선언문 (v4)

> **Tene는 Claude Code가 자동으로 인식하는 시크릿 관리 CLI입니다.**
>
> `tene init` 한 번이면 Claude Code가 시크릿 사용법을 알고 있습니다.
> Go 단일 바이너리. `brew install` 한 줄. ~5ms 시작.
> .env 파일처럼 **로컬에서 동작**하지만, Vault처럼 **암호화**됩니다.
> **가입이 필요 없고**, **인터넷이 필요 없고**, **무료**입니다.
>
> GitHub Secrets는 CI/CD, Tene는 로컬 개발. 완전히 다른 제품입니다.
> 코드는 **오픈소스**입니다. 서버가 없으니 **해킹 대상도 없습니다.**

---

## 6. 비용 구조 분석: MVP = $0

### 6.1 Phase 1: MVP (Go CLI 로컬 전용)

| 비용 항목 | 금액 | 설명 |
|-----------|------|------|
| 서버 인프라 | **$0** | 로컬 전용, 서버 없음 |
| Homebrew tap 호스팅 | $0 | GitHub 무료 |
| GitHub 호스팅 | $0 | 오픈소스 무료 |
| goreleaser | $0 | 오픈소스 |
| 도메인/DNS | $15/년 | tene.sh |
| **월간 총 비용** | **~$1.25/월** | 사실상 제로 |

**핵심**: MVP에서는 서버 비용이 $0. Homebrew 설치 100만 회 = 비용 $0.

### 6.2 Phase 2: Cloud (수요 검증 후, 서버리스 사용 안 함)

| 구성요소 | 기술 | 비용 (예상) |
|----------|------|------------|
| 컴퓨팅 | **ECS Fargate** | $50-200/월 |
| 로드밸런서 | **NLB** | $20-40/월 |
| 데이터베이스 | **RDS PostgreSQL** | $50-200/월 |
| 저장소 | **S3** | $5-50/월 |
| **총 비용** | | **$125-490/월** |

**왜 서버리스를 사용하지 않는가:**
- Steve가 서버리스를 선호하지 않음
- 트래픽이 늘면 비용 역전 (Lambda/API Gateway 비용 예측 어려움)
- ECS Fargate + NLB + RDS PostgreSQL이 트래픽 증가에 선형적 비용 구조

---

## 7. 전략적 권고사항 (v4)

### 7.1 Phase 1 전략: Go CLI MVP (0-3개월)

**타겟**: 솔로 바이브코더 (Claude Code 사용자)
**핵심 메시지**: "Claude Code가 자동 인식. Go 바이너리. 5ms 시작. 무료."
**비용**: $0

**액션**:
1. Go CLI MVP 출시: set/get/run/list/delete/import/export + Master Password + SQLite
2. **`tene init` → CLAUDE.md 자동 생성** (핵심 차별점)
3. `tene sync` Fake Door Test 내장 (waitlist 안내)
4. `tene export --encrypted` 암호화 수동 백업
5. goreleaser로 Homebrew tap + 멀티 플랫폼 바이너리 배포
6. GitHub 오픈소스 공개
7. 바이브코딩 커뮤니티 DevRel

### 7.2 수요 검증 (3-6개월)

**목표**: Cloud 구축 여부 결정 + Cursor/Windsurf 지원 시기 결정
**방법**:
- Step 1: Homebrew/curl 설치 수, GitHub Stars 추적
- Step 2: `tene sync` Fake Door → waitlist 가입 수 확인
- Step 3: waitlist 반응 보고 Cloud 구축 여부 결정
- Step 4: Cursor/Windsurf 사용자 요청 추적 → Phase 2 시기 결정

### 7.3 Phase 2 전략: Cloud + 멀티 에디터 (수요 검증 후, 6-12개월)

**전제 조건**: waitlist 100명+ 등록
**인프라**: ECS Fargate + NLB + RDS PostgreSQL + S3 (서버리스 X)
**기능**: 암호화 클라우드 백업, 멀티 디바이스 동기화, 웹 대시보드
**멀티 에디터**: `--cursor`, `--windsurf` 플래그 추가 (.cursorrules, .windsurfrules)

### 7.4 Phase 2+ 전략: 팀 기능 (가설 검증 후)

**타겟**: 팀 시크릿 공유 수요가 확인된 경우만 진행
**전제 조건**: Fake Door Test에서 15%+ 관심 확인 후

---

*Analysis by pm-strategy agent | Frameworks: JTBD 6-Part VP, Lean Canvas, SWOT, Porter's 5 Forces*
*Architecture: Go CLI Local-Only MVP ($0) + Claude Code 자동 인식*
*Tech Stack: Go + cobra + modernc.org/sqlite + golang.org/x/crypto*
*Competitive data updated: 2026-04-06*
