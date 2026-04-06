# Tene Discovery Analysis v4
## Go CLI + Claude Code 전용 MVP: 기회 발견과 검증

> v4 (2026-04-06) — Go 언어 전환 + Claude Code 전용 MVP 반영, 단일 바이너리 배포
> PM Agent: pm-discovery | Framework: Teresa Torres Opportunity Solution Tree + 5-Step Discovery Chain
> Architecture: Go CLI Local-Only MVP ($0) → Cloud Phase 2 (수요 검증 후)

---

## 1. Brainstorm: AI Agent 자동 인식 관점에서의 Pain Point 재발견

### 1.1 아키텍처 전환의 핵심 인사이트

v3에서는 "AI 에이전트가 시크릿 관리 도구를 자동으로 인식하면 어떨까?"가 핵심이었다. v4에서는 여기에 **Go 단일 바이너리의 극단적 간편함**과 **Claude Code 전용 집중**을 더한다:

> **"Node.js 설치도 필요 없는 단일 바이너리 CLI로, Claude Code가 자동 인식하는 시크릿 관리"**

현재 모든 시크릿 관리 도구 — Vault, Doppler, Infisical, 1Password — 에는 이 기능이 없다. AI 에이전트(Claude Code)가 프로젝트를 열 때 시크릿 관리 도구의 존재를 자동으로 인식하고 사용법을 아는 것. 이것이 Tene v4의 진짜 차별점이다.

**핵심 메커니즘:**
- `tene init` → CLAUDE.md 자동 생성 (Claude Code가 프로젝트 열 때 자동 읽음)
- AI 에이전트가 "이 프로젝트는 tene으로 시크릿을 관리한다"를 자동 인식
- `tene get KEY`, `tene run -- CMD` 사용법을 AI가 이미 알고 있음
- Cursor/Windsurf 등 다른 AI 에디터 지원은 Phase 2

**비유:**
- `.gitignore`가 있으면 Git이 무시할 파일을 안다
- `CLAUDE.md`가 있으면 Claude Code가 프로젝트 컨텍스트를 안다
- **Tene이 CLAUDE.md에 시크릿 관리 가이드를 넣으면, AI가 시크릿 사용법을 안다**

### 1.2 Go 전환의 핵심 인사이트

> **"CLI 도구에 Node.js 런타임이 필요한가? 아니다."**

- Go 단일 바이너리: ~10-15MB, 시작 속도 ~5ms
- Node.js 기반 CLI: ~50MB (node_modules), 시작 속도 ~200ms
- 설치: `brew install tomo-kay/tap/tene` (macOS) 또는 `curl -fsSL https://tene.dev/install.sh | sh` (Linux/WSL)
- Node.js 설치 불필요 — 바이브코더에게 의존성 하나를 줄여준다

### 1.3 MVP = CLI 로컬 전용, 서버 비용 $0

> **"시크릿 관리에 서버가 필요한가? 아니다."**
> **"클라우드가 MVP에 필요한가? 아직 모른다. 수요를 먼저 확인하자."**

- Go CLI + modernc.org/sqlite + golang.org/x/crypto
- 서버 없음, 회원가입 없음, 비용 없음
- `brew install tomo-kay/tap/tene` → `tene init` → `tene set` → `tene run`
- Cloud는 Phase 2 — 수요 검증 후에만 구축

### 1.4 바이브코더의 시크릿 관리 Pain Points (v4 재분석)

| # | Pain Point | 심각도 | 빈도 | v4 해결 |
|---|-----------|--------|------|---------|
| P1 | AI 에이전트가 API 키를 코드에 하드코딩 | Critical | 매우 높음 | `tene run` + CLAUDE.md 자동 인식 |
| P2 | .env 파일이 Git에 커밋됨 | Critical | 높음 | 로컬 암호화 볼트, .env 불필요 |
| P3 | AI 에이전트가 시크릿 관리 도구 존재를 모름 | High | 매우 높음 | **CLAUDE.md 자동 생성** |
| P4 | 시크릿 관리 도구가 서버/가입을 요구 | High | 높음 | **서버 없음, 가입 없음** |
| P5 | 기존 도구의 가격이 부담 (Doppler $21/유저/월) | High | 높음 | **무료 (로컬 전용)** |
| P6 | 오프라인에서 시크릿 접근 불가 | High | 중간 | **100% 오프라인 동작** |
| P7 | 새 도구 학습 비용이 높음 | Medium | 높음 | `set/get/run` 3개 명령어면 끝 |
| P8 | 프로젝트마다 .env 파일 반복 설정 | Medium | 매우 높음 | 프로젝트별 로컬 볼트 |
| P9 | CLI 설치에 Node.js 런타임이 필요 | Medium | 높음 | **Go 단일 바이너리, 의존성 제로** |
| P10 | 시크릿 관리 도구 자체가 해킹될 수 있다는 불안 | High | 높음 | **서버가 없으면 해킹 대상 없음** |

### 1.5 P3이 핵심 Pain Point인 이유

**기존 도구의 AI 에이전트 통합 현황:**

| 도구 | AI 에이전트가 자동 인식하는가? | 설정 필요 |
|------|:---------------------------:|----------|
| .env 파일 | 부분적 (파일이 보이면 읽음) | 없음 |
| Vault | No | MCP 설정 필요 |
| Doppler | No | `doppler run` 수동 설정 |
| Infisical | No | Agent Sentinel MCP 설정 |
| 1Password | No | SDK 통합 필요 |
| Dotenvx | No | CLI 학습 필요 |
| **Tene** | **Yes** | **`tene init`이 CLAUDE.md 자동 생성** |

.env 파일이 "부분적으로" AI 에이전트에게 인식되는 이유는 파일 자체가 프로젝트에 보이기 때문이다. 하지만 .env는 암호화되지 않아 보안 위험이 있다. Tene은 AI 에이전트에게 "암호화된 시크릿을 안전하게 사용하는 방법"을 자동으로 알려준다.

### 1.6 Tene가 해결하는 것 / 못 하는 것 (정직하게)

| 구분 | 내용 |
|------|------|
| **해결** | 키값 안전 저장, XChaCha20-Poly1305 암호화, 프로젝트별 관리, 환경변수 주입, AI Agent 자동 인식 연동 |
| **부분적** | 프로젝트 간 재사용 (글로벌 키 공유는 향후 검토) |
| **못 함** | API key 만료 기간 확인, 자동 갱신/로테이션 (Phase 2+ 가설) |

### 1.7 바이브코더의 여정: 설치 → AI 자동 인식 → 수요 검증

```
[Stage 1: 발견] ─────────────────────────────────────────
"AI로 코딩하는데 API 키 관리가 불안하다"
→ 검색 → Tene 발견 → "서버 없이 로컬에서? 무료?"
→ brew install tomo-kay/tap/tene 즉시 설치 (~5초, Node.js 불필요)

[Stage 2: 첫 사용 + AI 자동 인식] ─────────────────────
$ tene init
→ Master Password 설정
→ CLAUDE.md 자동 생성 (Claude Code 프로젝트용)
$ tene set STRIPE_KEY sk_test_xxx
$ tene run -- claude
→ "와, 서버 가입도 없이 바로 되네? 설치도 5초?"
→ Claude Code가 tene 사용법을 이미 알고 있음!

[Stage 3: 습관화] ──────────────────────────────────────
프로젝트 1, 2, 3... 모두 tene으로 관리
→ "이제 .env 파일은 안 쓴다"
→ 로컬 볼트에 시크릿 20-30개 축적

[Stage 4: 수요 검증 모먼트] ─── Fake Door ──────────────
$ tene sync
→ "클라우드 동기화 기능을 준비하고 있습니다!
   관심이 있으시면 waitlist에 등록하세요: tene.dev/cloud"
→ waitlist 가입 수로 수요 확인

[Stage 5: 백업 필요] ──────────────────────────────────
→ 수동 백업: tene export --encrypted > backup.enc
→ "이것만으로도 충분하네"
→ 또는 "클라우드 동기화 있으면 좋겠다" → waitlist 가입
```

---

## 2. Assumptions: 핵심 가정 식별 (v4)

| # | 가정 (Assumption) | 유형 | v3 대비 변경 |
|---|------------------|------|-------------|
| A1 | 바이브코더의 75%+가 시크릿 관리에 불편을 느낀다 | Desirability | 유지 |
| A2 | **AI Agent 자동 인식(CLAUDE.md)이 핵심 차별점이다** | Desirability | 유지 |
| A3 | 로컬 전용 + 서버 없음이 바이브코더에게 매력적이다 | Desirability | 유지 |
| A4 | **Cloud는 MVP에 불필요하다 (수요 검증 후 구축)** | Viability | 유지 |
| A5 | `tene sync` Fake Door로 Cloud 수요를 확인할 수 있다 | Viability | 유지 |
| A6 | Master Password 기반 로컬 암호화가 충분히 안전하다 | Feasibility | 유지 |
| A7 | SQLite 기반 로컬 볼트가 1,000개 이상 시크릿을 빠르게 처리한다 | Feasibility | 유지 |
| A8 | "서버가 없다"는 메시지가 보안 신뢰의 핵심 차별점이 된다 | Desirability | 유지 |
| A9 | CLI 3개 명령어(set/get/run)로 핵심 가치를 전달할 수 있다 | Usability | 유지 |
| A10 | **Go 단일 바이너리가 Node.js CLI 대비 설치 전환율을 높인다** | Feasibility | **신규**: v4 핵심 가정 |
| A11 | **Claude Code 전용 집중이 MVP 완성도를 높인다** | Feasibility | **신규**: v4 타겟 집중 |

---

## 3. Prioritize: 가정 우선순위 (Impact x Risk Matrix)

```
         높은 리스크 (불확실)
              |
    A10 *     |     * A2 (AI 자동 인식)
    A5 *      |     * A4 (Cloud 불필요)
              |     * A8
    ----------+----------
              |     * A1
    A7 *      |     * A9
    A11 *     |     * A3
    A6 *      |
         낮은 리스크 (확실)
              
   낮은 임팩트          높은 임팩트
```

### 최우선 검증 대상 (High Impact + High Risk)

| 순위 | 가정 | 검증 이유 |
|------|------|----------|
| 1 | **A2: AI Agent 자동 인식이 차별점인가** | 핵심 가정. CLAUDE.md 자동 생성이 실제로 사용자에게 가치를 주는지 |
| 2 | **A4: Cloud 없이 MVP가 성립하는가** | 수요 확인 전 Cloud 구축 비용을 피할 수 있는지 |
| 3 | **A8: "서버가 없다"가 차별점인가** | 마케팅 메시지 핵심. 개발자 반응 검증 |

---

## 4. Experiments: 검증 실험 설계 (v4)

### Experiment 1: AI Agent 자동 인식 가치 검증 (A2)

| 항목 | 내용 |
|------|------|
| **방법** | CLI 프로토타입 + 사용자 테스트 (5명) |
| **대상** | Claude Code 사용자 |
| **가설** | `tene init`이 CLAUDE.md를 생성한 후, AI 에이전트가 tene 사용법을 자동 인식하면 5명 중 4명이 "기존 도구에 없는 핵심 기능"으로 평가 |
| **테스트 시나리오** | (1) tene init → CLAUDE.md 생성 확인 (2) Claude Code에서 시크릿 관련 작업 요청 (3) AI가 자동으로 `tene get`/`tene run` 사용 |
| **메트릭** | AI 자동 인식 성공률, 사용자 만족도(1-10), "핵심 차별점" 평가 비율 |
| **기간** | 1주 |
| **비용** | 개발 시간만 |
| **성공 기준** | 5명 중 4명이 AI 자동 인식을 "기존 도구에 없는 핵심 기능"으로 평가 |

### Experiment 2: "서버 없는 시크릿 관리 + AI 자동 인식" 메시지 검증 (A3, A8)

| 항목 | 내용 |
|------|------|
| **방법** | A/B 랜딩 페이지 테스트 |
| **대상** | 바이브코더 커뮤니티 (X/Twitter, Reddit r/vibecoding, GeekNews) |
| **가설A** | "서버 없는 시크릿 관리. 로컬에서 암호화. 무료." |
| **가설B** | "AI 에이전트가 자동으로 인식하는 시크릿 관리. 서버 없이. 무료." |
| **메트릭** | 대기자 등록 전환율 비교 |
| **기간** | 2주 |
| **비용** | $0 |
| **성공 기준** | 가설B가 가설A 대비 20%+ 높은 전환율 |

### Experiment 3: Cloud 수요 Fake Door Test (A4, A5)

| 항목 | 내용 |
|------|------|
| **방법** | `tene sync` 명령어 Fake Door |
| **대상** | MVP CLI 사용자 |
| **가설** | CLI 사용자의 10%+가 `tene sync` 실행 후 waitlist에 등록 |
| **시나리오** | `tene sync` 실행 → "클라우드 동기화 준비 중! waitlist: tene.dev/cloud" 표시 |
| **메트릭** | `tene sync` 실행 횟수, waitlist 등록 수 |
| **기간** | MVP 출시 후 4주 |
| **비용** | $0 |
| **성공 기준** | waitlist 등록 10%+ |

### Experiment 4: Go 바이너리 vs Node.js CLI 설치 전환율 (A10)

| 항목 | 내용 |
|------|------|
| **방법** | 설치 방법별 전환율 비교 |
| **대상** | 랜딩 페이지 방문자 |
| **가설** | `brew install tomo-kay/tap/tene` (~5초)가 `npm install -g @tene/cli` (~15초, Node.js 필요) 대비 설치 전환율 30%+ 높음 |
| **메트릭** | 설치 완료율, 첫 `tene init` 실행까지 시간 |
| **기간** | 2주 |
| **비용** | $0 |
| **성공 기준** | Go 바이너리 설치 전환율이 30%+ 높음 |

---

## 5. Opportunity Solution Tree (OST) v4

```
                    +-------------------------------------------+
                    |          Desired Outcome                  |
                    |  "바이브코더가 시크릿을 서버 없이          |
                    |   로컬에서 무료로, Claude Code가            |
                    |   자동 연동하여 안전하게 관리한다"          |
                    +-------------------+-----------------------+
                                        |
            +---------------------------+---------------------------+
            |                           |                           |
   +--------v--------+       +---------v--------+       +----------v--------+
   |  Opportunity 1   |       |  Opportunity 2   |       |  Opportunity 3    |
   |  Go 단일 바이너리 |       |  Claude Code     |       |  Cloud 확장       |
   |  + 로컬 시크릿    |       |  자동 인식        |       |  (Phase 2, 수요   |
   |  극단적 간소화    |       |  (핵심 차별점)    |       |   검증 후)        |
   +--------+---------+       +---------+--------+       +----------+--------+
            |                           |                           |
   +--------+--------+       +---------+--------+       +----------+--------+
   |                  |       |                  |       |                   |
+--v--+          +---v--+ +--v--+          +---v--+ +---v--+          +---v--+
| S1  |          | S2   | | S3  |          | S4   | | S5   |          | S6   |
|Go   |          |Master| |CLAUDE|         |Cursor| |tene  |          |Cloud |
|단일 |          |Pass  | |.md   |         |rules | |sync  |          |백업  |
|바이 |          |+     | |자동  |         |자동   | |Fake  |          |+동기 |
|너리 |          |XCha20| |생성  |         |생성   | |Door  |          |화    |
|CLI  |          |Poly  | |(MVP) |         |(Ph.2)| |Test  |          |      |
+--+--+          +--+--+ +--+--+          +--+--+ +--+--+          +--+--+
   |                |       |                |       |                 |
+--v--------+  +---v----+ +--v--------+ +---v----+ +--v--------+ +---v----+
| E1: brew   | | E2:    | | E3: AI    | | E4:    | | E5: tene  | | E6:    |
| install 5초| | 마스터  | | 자동 인식 | | Phase  | | sync     | | 수요   |
| 설치 전환율 | | 패스워드| | 가치 테스 | | 2에서  | | Fake Door| | 확인   |
| 테스트     | | UX 테스 | | 트        | | 검증   | | waitlist | | 후     |
|            | | 트      | |           | |        | |          | | 결정   |
+------------+ +--------+ +-----------+ +--------+ +----------+ +--------+

[Phase 1: Free MVP]            [Phase 1 핵심]          [Phase 2: 가설]
Go CLI + 로컬 볼트              Claude Code 자동 인식    Cloud (수요 검증 필요)
```

### OST 상세 설명

#### Opportunity 1: Go 단일 바이너리 + 로컬 시크릿 관리의 극단적 간소화 (MVP 코어)
> "서버 없이, 가입 없이, 런타임 없이, 1분 만에 시작"

- `brew install tomo-kay/tap/tene` → `tene init` → `tene set KEY VALUE` → `tene run -- claude`
- Go 단일 바이너리: ~10-15MB, 시작 속도 ~5ms, Node.js 설치 불필요
- 서버 연결 불필요, 회원가입 불필요, 비용 $0
- modernc.org/sqlite로 로컬에 암호화된 시크릿 저장 (순수 Go, CGo 없음)
- 오프라인 100% 동작
- `tene export --encrypted`로 수동 암호화 백업 제공

#### Opportunity 2: Claude Code 자동 인식 (v4 핵심 차별점)
> "Claude Code가 자동으로 인식하는 시크릿 관리 도구"

**Solution 3 (S3): CLAUDE.md 자동 생성 (MVP)**
- `tene init` 실행 시 프로젝트 루트에 CLAUDE.md 자동 생성
- Claude Code가 프로젝트 열 때 tene 사용법을 자동 인식
- "이 프로젝트는 tene으로 시크릿을 관리합니다. `tene get KEY`로 시크릿을 조회하세요."

**Solution 4 (S4): Cursor/Windsurf 지원 (Phase 2)**
- `--cursor`, `--windsurf` 플래그는 Phase 2로 이동
- MVP에서는 Claude Code 전용에 집중하여 완성도를 높임

**이것이 Vault, Doppler, Infisical에 없는 진짜 차별점:**
- 기존 도구: AI 에이전트가 도구 존재를 모름 → 수동 설정 필요
- Tene: Claude Code가 프로젝트 열면 자동 인식 → 설정 불필요

#### Opportunity 3: Cloud 확장 (Phase 2 — 수요 검증 후)
> "Cloud는 수요가 확인되면 만든다."

- MVP에 Cloud 없음
- `tene sync` 명령어 = Fake Door Test (waitlist 안내만 표시)
- `tene export --encrypted`로 수동 백업 제공
- waitlist 반응 보고 Cloud 구축 여부 결정
- Cloud 구축 시: ECS Fargate + NLB + RDS PostgreSQL + S3 (서버리스 사용 안 함)

---

## 6. 수요 검증 방법 (v4)

### 3단계 검증 프로세스

```
[Step 1: CLI 출시 → 기본 지표 추적]
brew install tomo-kay/tap/tene 출시 (+ curl 설치, go install)
   |
   +-- Homebrew 설치 수 추적
   +-- GitHub Stars 추적
   +-- GitHub Releases 다운로드 수 추적
   +-- 성공 기준: 주간 설치 500+ / Stars 1,000+

[Step 2: tene sync Fake Door → Cloud 수요 확인]
$ tene sync
→ "클라우드 동기화를 준비하고 있습니다!
   관심이 있으시면 waitlist에 등록하세요."
→ tene.dev/cloud → waitlist 가입 수 확인
   |
   +-- 성공 기준: CLI 사용자의 10%+ waitlist 등록

[Step 3: waitlist 반응 → Cloud 구축 결정]
waitlist 등록 100명+ → Cloud 구축 시작
waitlist 등록 < 50명 → Cloud 보류, CLI 기능 강화
   |
   +-- Cloud 구축 시: ECS Fargate + NLB + RDS PostgreSQL + S3
```

---

## 7. GitHub Secrets과의 차별점

| 비교 | GitHub Secrets | Tene |
|------|:-------------:|:----:|
| **용도** | CI/CD 파이프라인 전용 | 로컬 개발 워크플로우 + AI 에이전트 전용 |
| **접근** | GitHub Actions 내에서만 | 로컬 CLI에서 즉시 |
| **오프라인** | 불가 (GitHub 서버 필요) | 100% 오프라인 |
| **AI 에이전트** | 지원 안 함 | CLAUDE.md 자동 인식 |
| **환경** | Repository/Organization 레벨 | 프로젝트/환경(dev/prod) 레벨 |
| **사용 시점** | 배포/테스트 파이프라인 | 개발 중 실시간 |

**결론**: GitHub Secrets와 Tene는 완전히 다른 제품. GitHub Secrets = 배포 파이프라인, Tene = 로컬 개발 + AI 에이전트. 겹치지 않으며 보완적.

---

## 8. 핵심 인사이트 (v4)

### 8.1 왜 "Claude Code 자동 인식"이 최강의 차별점인가

**현재 상황:**
- Claude Code는 CLAUDE.md를 읽고 프로젝트 컨텍스트를 파악한다
- 하지만 어떤 시크릿 관리 도구도 이 메커니즘을 활용하지 않는다

**Tene의 기회:**
- `tene init` → CLAUDE.md에 시크릿 관리 가이드 자동 추가
- Claude Code가 "이 프로젝트는 tene을 사용한다"를 자동 인식
- `process.env.STRIPE_KEY` 대신 AI가 `tene get STRIPE_KEY` 또는 `tene run` 사용법을 이미 알고 있음
- **설정 없이 자동 연동** — 이것이 MCP, SDK, 수동 설정이 필요한 모든 경쟁사와의 근본적 차이

### 8.2 왜 Go 단일 바이너리인가

| 비교 | Node.js CLI (v3) | Go CLI (v4) |
|------|:-----------------:|:-----------:|
| **설치** | `npm install -g @tene/cli` | `brew install tomo-kay/tap/tene` |
| **런타임 의존성** | Node.js 18+ 필요 | **없음** (단일 바이너리) |
| **시작 속도** | ~200ms | **~5ms** (40배 빠름) |
| **바이너리 크기** | ~50MB (node_modules) | **~10-15MB** |
| **크로스 플랫폼** | npm 기반 | **goreleaser** (macOS/Linux/Windows 네이티브) |
| **암호화** | libsodium-wrappers (WASM) | **golang.org/x/crypto/nacl/secretbox** (네이티브) |
| **DB** | better-sqlite3 (네이티브 바인딩) | **modernc.org/sqlite** (순수 Go, CGo 없음) |

### 8.3 시크릿 누출 현황 (2026)

- GitHub에 노출된 시크릿: 연간 **2,865만 개** (전년 대비 34% 증가, GitGuardian 2026)
- AI 서비스 관련 시크릿 누출: **81.5% 급증** (GitGuardian 2026)
- Claude Code 커밋의 시크릿 누출율: **3.2%** (인간 기준 1.5%의 2배)
- MCP 설정 파일에서 노출된 시크릿: **24,008개**
- 바이브코딩 앱 5,600개 분석 결과: **400+ 노출된 시크릿** 발견

### 8.4 Local-First 소프트웨어 트렌드와의 정합성

2026년 Local-First 소프트웨어 운동이 본격화되고 있다:
- FOSDEM 2026에서 Local-First 전용 세션 개최
- PowerSync, ElectricSQL, Automerge 등 프로덕션 레디 도구 성숙
- 개발자들이 클라우드 종속에서 벗어나려는 움직임 가속

Tene v4는 이 트렌드에 완벽히 부합한다:
- **데이터 주권**: 시크릿이 사용자 디바이스에만 존재
- **오프라인 퍼스트**: 인터넷 없이도 100% 동작
- **클라우드 옵셔널**: 수요가 확인되면 그때 구축
- **의존성 제로**: Go 단일 바이너리, 외부 런타임 불필요

---

## 9. 핵심 차별화 방향 (v4)

### v3 vs v4 포지셔닝 변화

| 항목 | v3 (TypeScript + Claude Code/Cursor) | v4 (Go + Claude Code 전용) |
|------|--------------------------------------|---------------------------|
| **코어 메시지** | "AI 에이전트가 자동으로 인식하는 시크릿 관리" | **"Claude Code가 자동 인식. Go 단일 바이너리. 5ms 시작."** |
| **언어/런타임** | TypeScript / Node.js | **Go / 단일 바이너리 (런타임 불필요)** |
| **설치** | `npm install -g @tene/cli` | **`brew install tomo-kay/tap/tene`** |
| **시작 속도** | ~200ms | **~5ms** |
| **AI 타겟** | Claude Code + Cursor | **Claude Code 전용 (MVP)** |
| **Cursor/Windsurf** | `--cursor`, `--windsurf` 플래그 | **Phase 2** |
| **Cloud** | Phase 2 (수요 검증 후) | Phase 2 (수요 검증 후, 유지) |
| **배포** | npm publish | **goreleaser (Homebrew + 바이너리)** |

### 핵심 포지셔닝 선언문 (v4)

> **Tene는 Claude Code가 자동으로 인식하는 시크릿 관리 CLI입니다.**
>
> `tene init` 한 번이면 AI가 시크릿 사용법을 알고 있습니다.
> Go 단일 바이너리. Node.js 설치 불필요. ~5ms 시작.
> 로컬에서 암호화하고, 로컬에 저장합니다. 무료입니다.
> 회원가입이 필요 없습니다. 인터넷이 필요 없습니다.
>
> **코드는 오픈소스입니다. 직접 확인하세요.**

---

*Analysis by pm-discovery agent | Frameworks: Teresa Torres OST, 5-Step Discovery Chain*
*Architecture: Go CLI Local-Only MVP ($0) + Claude Code 자동 인식*
*Tech Stack: Go + cobra + modernc.org/sqlite + golang.org/x/crypto*
*Market data: GitGuardian State of Secrets Sprawl 2026, Mordor Intelligence, Meticulous Research*
