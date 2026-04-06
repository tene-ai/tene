# Tene Market Research v4
## Claude Code 전용 + Go CLI 관점의 경쟁 분석 + GitHub Secrets 차별점

> v4 (2026-04-06) — Go 언어 전환 + Claude Code 전용 MVP 반영
> PM Agent: pm-research | Frameworks: Persona Development, Competitive Analysis, TAM/SAM/SOM, Customer Journey Map
> Architecture: Go CLI Local-Only MVP ($0) → Cloud Phase 2 (수요 검증 후)

---

## 1. 사용자 페르소나 (v4 — MVP 무료 전용)

### Persona 1: "민지" — MVP 사용자 (솔로 바이브코더)

| 항목 | 상세 |
|------|------|
| **이름** | 김민지 |
| **나이/직업** | 27세 / 프리랜서 풀스택 개발자 |
| **기술 수준** | 중급. Claude Code로 빠르게 프로토타입 제작 |
| **사용 도구** | Claude Code, Vercel, Supabase, Stripe |
| **관리 시크릿 수** | 5-15개 (API 키, DB URL, OAuth 토큰) |
| **현재 방법** | .env 파일 + .gitignore |
| **JTBD** | "서버 가입 없이 로컬에서 시크릿을 관리하고, AI가 자동으로 사용법을 아는 도구가 필요하다" |
| **Pain Points** | (1) Claude Code가 시크릿 관리 도구를 몰라서 .env를 직접 읽으려 함 (2) .env가 Git에 올라갈까 불안 (3) 새 도구마다 가입 요구 |
| **Goals** | 가입 없이 즉시 사용, AI가 자동 인식, .env 대체 |
| **Trigger** | "tene init 하면 CLAUDE.md가 자동 생성되어 Claude Code가 바로 인식한다고?" |
| **WTP (지불의향)** | **$0** — 무료 로컬이면 충분 |

**핵심 시나리오**: 민지는 새 SaaS를 Claude Code로 만들고 있다. `brew install tomo-kay/tap/tene` → `tene init` → CLAUDE.md 자동 생성 → `tene set STRIPE_KEY sk_test_xxx` → `tene run -- claude`. Claude Code가 "이 프로젝트는 tene으로 시크릿을 관리합니다"를 자동 인식. 서버 가입? 없다. Node.js 설치? 불필요.

**민지가 Tene를 선택하는 이유**: "brew install 한 줄이면 끝? AI가 알아서 인식하네? 가입도 없고? 이게 왜 이제야 나왔지?"

---

### Persona 2: "재현" — 파워 사용자 (멀티 프로젝트)

| 항목 | 상세 |
|------|------|
| **이름** | 박재현 |
| **나이/직업** | 31세 / 인디 해커, 사이드 프로젝트 다수 |
| **기술 수준** | 상급. 프로젝트 10개+ 동시 운영 |
| **관리 시크릿 수** | 20-50개 |
| **현재 방법** | .env 파일 + 수동 관리 |
| **Pain Points** | (1) 프로젝트마다 .env 반복 설정 (2) 디바이스 간 시크릿 동기화 불편 (3) 백업 없음 |
| **Goals** | 프로젝트별 시크릿 관리, 암호화 백업 |
| **WTP** | `tene export --encrypted`로 수동 백업에 만족하거나, Cloud waitlist에 등록 |
| **Tene 사용** | 로컬 전용, 프로젝트 5-10개, `tene export --encrypted`로 백업 |

**핵심 시나리오**: 재현은 Tene 무료로 프로젝트 10개를 관리한다. `tene export --encrypted > ~/backup/secrets.enc`로 정기 백업. `tene sync` 실행 → "클라우드 동기화 준비 중! waitlist: tene.sh/cloud" → waitlist 등록.

---

### Persona 3: "Sarah" — 잠재적 팀 사용자 (Phase 2+ 가설)

| 항목 | 상세 |
|------|------|
| **이름** | Sarah Chen |
| **나이/직업** | 33세 / 스타트업 CTO (팀 5명) |
| **관리 시크릿 수** | 50-100개 (팀 전체) |
| **Status** | **가설** — 실제 수요 검증 필요 (Fake Door Test) |

> **중요**: Sarah 페르소나는 가설이다. Tene의 MVP는 민지에 집중한다.

---

## 2. 경쟁 제품 분석 (v4 — Claude Code 자동 인식 + GitHub Secrets 차별점)

### 2.1 경쟁사 비교: AI Agent 통합 관점

| 항목 | Tene v4 | Vault | Doppler | Infisical | 1Password | Dotenvx | GitHub Secrets |
|------|:-------:|:-----:|:-------:|:---------:|:---------:|:-------:|:-------------:|
| **AI Agent 자동 인식** | **Yes** | No | No | No | No | No | No |
| **CLAUDE.md 생성** | **Yes** | No | No | No | No | No | No |
| **서버 필요** | **No** | Yes | Yes | Yes | Yes | Partial | Yes |
| **가입 필요** | **No** | Yes | Yes | Yes | Yes | No | Yes |
| **오프라인 동작** | **100%** | No | No | No | No | Yes | No |
| **무료 플랜** | **무제한** | 셀프호스팅 | 3인/3프로젝트 | 셀프호스팅 | 없음 | CLI 무료 | Repo별 제한 |
| **오픈소스** | **Yes (MIT)** | Partial (BSL) | No | Yes (MIT) | No | Yes | No |
| **단일 바이너리** | **Yes (Go)** | Yes (Go) | No | No | No | Yes | N/A |
| **런타임 의존성** | **없음** | 없음 | Node.js | Node.js | 없음 | Node.js | N/A |
| **용도** | **로컬 개발 + AI** | 인프라 | 환경 동기화 | DevOps | 패스워드 | .env 암호화 | **CI/CD** |

### 2.2 GitHub Secrets과의 차별점 (v4 핵심)

| 비교 | GitHub Secrets | Tene |
|------|:-------------:|:----:|
| **용도** | CI/CD 파이프라인 전용 | 로컬 개발 워크플로우 + AI 에이전트 |
| **접근 시점** | 배포/테스트 시 | 개발 중 실시간 |
| **접근 방법** | GitHub Actions workflow YAML | `tene get KEY` / `tene run -- CMD` |
| **오프라인** | 불가 (GitHub 서버 필요) | 100% 오프라인 |
| **AI 에이전트** | 지원 안 함 | **CLAUDE.md 자동 인식** |
| **환경 관리** | Repository/Environment 레벨 | 프로젝트/환경(dev/staging/prod) 레벨 |
| **로컬 개발** | 지원 안 함 (CI/CD 전용) | 핵심 용도 |
| **가격** | 무료 (Public), 유료 (Private 한도 초과) | 무료 |
| **암호화** | GitHub 관리 (서버 측) | 로컬 XChaCha20-Poly1305 (사용자 관리) |

**결론**: GitHub Secrets와 Tene는 **완전히 다른 제품**. GitHub Secrets = 배포 파이프라인의 시크릿 주입. Tene = 로컬 개발 환경 + AI 에이전트의 시크릿 관리. 겹치지 않으며 보완적이다.

### 2.3 경쟁사별 AI Agent 통합 심층 비교

| 제품 | AI Agent 통합 방식 | 자동 인식 | 설정 필요 | Tene 대비 |
|------|-------------------|:---------:|----------|----------|
| **Tene v4** | CLAUDE.md 자동 생성 | **Yes** | **없음** | 기준 |
| 1Password | Unified Access SDK/API | No | SDK 통합 | 설정 복잡 |
| Infisical | Agent Sentinel (MCP) | No | MCP 설정 | MCP 설정 필요 |
| Dotenvx | AS2 (ECIES ID) | No | CLI 학습 | 자동 인식 없음 |
| Vault | MCP 서버 | No | MCP 설정 | 관리자용 |
| Doppler | 없음 | No | `doppler run` | AI 미지원 |
| GitHub Secrets | 없음 | No | Actions YAML | CI/CD 전용 |

**핵심 인사이트**: AI 에이전트 "자동 인식"을 제공하는 시크릿 관리 도구는 Tene가 유일하다. 다른 모든 도구는 수동 설정(SDK, MCP, CLI 학습)이 필요하다.

### 2.4 경쟁사별 심층 분석 (v4)

#### (1) 1Password Unified Access (2026년 3월 출시)

**최신 동향**:
- 2026년 3월 "Unified Access" 플랫폼 정식 출시
- Anthropic Claude, Cursor, GitHub, Vercel, Perplexity와 공식 파트너십
- Discover → Secure → Audit 3단계 프레임워크
- 런타임 자격증명 스코핑 2026년 하반기 출시 예정

**Tene v4 vs 1Password**:
| 비교 | 1Password | Tene v4 |
|------|-----------|---------|
| AI 자동 인식 | No (SDK 필요) | **Yes (CLAUDE.md)** |
| 서버 | 필수 (클라우드) | 불필요 (로컬) |
| 가격 | $7.99/유저/월~ | $0 (무료) |
| 타겟 | 엔터프라이즈 | 솔로 바이브코더 |
| 설치 | 앱 설치 + 가입 | `brew install` 한 줄 |

**결론**: 타겟이 다르다. 1Password = 엔터프라이즈, Tene = 개인/솔로.

#### (2) Dotenvx AS2 (Agentic Secret Storage)

**Tene v4 vs Dotenvx**:
| 비교 | Dotenvx AS2 | Tene v4 |
|------|------------|---------|
| AI 자동 인식 | No | **Yes (CLAUDE.md)** |
| 저장 방식 | 암호화된 .env 파일 | SQLite 볼트 |
| 환경 관리 | .env.dev, .env.prod 파일 분리 | `tene env dev/prod` 명령어 |
| 언어 | Node.js | **Go (단일 바이너리)** |
| 가격 | CLI 무료, 클라우드 미공개 | CLI 무료, Cloud 수요 확인 후 |

**결론**: Dotenvx는 .env의 진화. Tene은 .env의 대체 + AI 자동 인식.

#### (3) Infisical + Agent Sentinel

**Tene v4 vs Infisical**:
| 비교 | Infisical | Tene v4 |
|------|-----------|---------|
| AI 통합 | MCP (Agent Sentinel) | **CLAUDE.md 자동 인식** |
| 서버 | 필수 | 불필요 |
| 복잡도 | 증가 중 (PKI, PAM) | 극도로 단순 |
| 가격 | $0 (셀프) / $6/유저 | $0 |

**결론**: Infisical = "더 나은 Vault". Tene = "AI가 자동 인식하는 .env 대체".

---

## 3. 시장 규모 (TAM/SAM/SOM)

### 3.1 TAM (Total Addressable Market)

| 출처 | 연도 | 시장 규모 | CAGR |
|------|------|-----------|------|
| Mordor Intelligence | 2025 | $4.22B | 13.8% |
| KBV Research | 2032 | $10.09B | 13.4% |

> **TAM = $4.22B** (2025 기준)

### 3.2 SAM (Serviceable Addressable Market)

```
전체 개발자 수 (2026): ~30M
AI 코딩 도구 사용: 30M x 60% = 18M
시크릿 관리 도구 미사용 또는 불만족: 18M x 40% = 7.2M
로컬 도구 선호 (보수적): 7.2M x 30% = 2.16M

SAM = 2.16M (잠재 사용자 수)
```

> **SAM = ~2.16M 사용자** (MVP는 무료이므로 사용자 수 기준)

### 3.3 SOM (Serviceable Obtainable Market)

**Phase 1 (첫 12개월) 현실적 목표:**
```
목표: 설치 50,000 / 활성 사용자 10,000
지역: 한국 + 영어권 바이브코더 커뮤니티
획득 경로: GitHub 오픈소스, DevRel, Homebrew, 입소문

SOM = 10,000 활성 사용자 (Year 1)
수익: $0 (MVP 무료)
```

**Phase 2 (Cloud 수요 확인 시):**
```
waitlist 반응에 따라:
- 가격/수익 모델 결정
- ECS Fargate + NLB + RDS PostgreSQL 구축
```

---

## 4. Customer Journey Map (v4 — Claude Code 자동 인식 여정)

### Primary Persona: 김민지 (무료 MVP 여정)

```
단계     인지          설치          AI 자동인식      습관화          수요 검증
----------------------------------------------------------------------

접점     X/Twitter     brew          CLAUDE.md       매일 tene run   tene sync
         Reddit        curl          Claude Code     프로젝트 추가   Fake Door
         GeekNews      GitHub        (자동 읽기)     시크릿 축적     waitlist

행동     "AI가 자동    brew install  tene init       tene set       $ tene sync
         인식하는      tomo-kay/     → CLAUDE.md     tene run --    → "waitlist
         시크릿 관리?  tap/tene      자동 생성!      claude         등록하세요"
         무료?        (~5초)        "AI가 바로      (매일 반복)
         Node.js                     인식하네!"
         불필요?"

감정     호기심        놀라움        감탄            편안함          "Cloud
         "진짜?       "5초 설치?   "이건 다른     "이제 .env     있으면 좋겠는데"
         무료?"       Node.js도    도구에 없는    안 쓴다"       또는 "이대로
                      불필요?"     기능이네"                     충분해"
```

### 핵심 전환 지점 (Moment of Truth) — v4

| # | 전환 지점 | 성공 기준 | 실패 시나리오 |
|---|-----------|----------|--------------|
| **MoT1** | 첫 설치 → 첫 시크릿 저장 | **1분 이내, 가입 없이, brew 한 줄** | "Node.js가 필요하네?" → 이탈 |
| **MoT2** | **CLAUDE.md 생성 → AI 자동 인식** | "Claude Code가 tene을 바로 알아보네!" | "CLAUDE.md가 뭐지?" |
| **MoT3** | 첫 `tene run` 경험 | "와, .env 없이 되네! 빠르다!" | ".env랑 뭐가 다르지?" |
| **MoT4** | 프로젝트 3개 이상 사용 | "시크릿이 한곳에 정리되네!" | "프로젝트마다 init 귀찮다" |
| **MoT5** | `tene sync` Fake Door | waitlist 등록 | "Cloud 필요 없어" |

---

## 5. 최신 경쟁 동향 심층 분석 (2026년 4월 기준)

### 5.1 GitGuardian State of Secrets Sprawl 2026

| 통계 | 수치 |
|------|------|
| GitHub 공개 커밋에 노출된 시크릿 (2025년) | **2,865만 개** (전년 대비 34% 증가) |
| AI 서비스 관련 시크릿 누출 증가 | **81.5% 급증** (1,275,105개) |
| Claude Code 커밋 시크릿 누출율 | **3.2%** (인간 1.5%의 2배) |
| MCP 설정 파일 시크릿 노출 | **24,008개** (MCP 첫해) |
| 2022년 유출 시크릿 미폐기율 | **64%** (4년 후에도 유효) |

### 5.2 바이브코딩 보안 위기

| 통계 | 출처 |
|------|------|
| 바이브코딩 앱 5,600개 중 취약점 2,000+, 노출 시크릿 400+ | getautonoma.com |
| AI 생성 코드의 53%에서 보안 문제 발견 | 다수 연구 |
| $150K API 키 유출 과금 사고 | chyshkala 보고 |

### 5.3 NHI (Non-Human Identity) 시장

| 통계 | 수치 |
|------|------|
| NHI 시장 규모 (2026) | $12.2B |
| NHI vs 인간 ID 비율 | **25-50:1** |
| CISO 우선순위 "AI 아이덴티티 보안" | 5점 만점 중 **4.46** |

---

## 6. 한국 시장 특수성 (v4)

| 요인 | 설명 | Tene 기회 |
|------|------|-----------|
| 바이브코딩 열풍 | 한국에서 특히 높은 관심 | 한국어 CLI 메시지, 한국어 문서 |
| 가격 민감도 | 유료 도구 저항 높음 | **무료 MVP** — 가격 저항 제로 |
| 보안 규제 | 개인정보보호법 강화 | "서버가 없다" = 규제 부담 감소 |
| 커뮤니티 | GeekNews, Velog, Discord 활발 | 한국어 콘텐츠로 선점 |
| Claude Code 채택 | 한국 개발자 Claude Code 사용 증가 | CLAUDE.md 자동 생성의 즉각적 가치 |
| Homebrew 사용 | macOS 개발자 대다수가 Homebrew 사용 | `brew install` 한 줄 설치 → 높은 전환율 |

---

*Analysis by pm-research agent | Sources: GitGuardian 2026, Mordor Intelligence, 1Password, Dotenvx, Infisical*
*Architecture: Go CLI Local-Only MVP ($0) + Claude Code 자동 인식*
*Tech Stack: Go + cobra + modernc.org/sqlite + golang.org/x/crypto*
*Data collected: 2026-04-06*
