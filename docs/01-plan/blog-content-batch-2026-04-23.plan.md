# Blog Content Batch 2026-04-23 — Plan

| 항목 | 값 |
|------|-----|
| Feature | `blog-content-batch-2026-04-23` |
| 의존성 (선행) | `blog-categories-and-tooling.plan.md` (category 도입 먼저 완료) |
| 브랜치 | `feature/blog-batch-w17` (별도, `feature/blog-categories` merge 이후 생성) |
| 총 공수 | ~6-10 인시간 (선택 조합에 따라) |

---

## 1. 컨텍스트

`blog-categories-and-tooling` plan이 도입될 시점 각 카테고리 콘텐츠 분포:

| Category | 편수 (migration 후) | 상태 |
|---|:-:|---|
| `tools` | 3 | 적정 |
| `engineering` | 1 | **빈약** — 우선 보강 |
| `vibe-coding` | 5 | 풍부 |
| `philosophy` | **0** | **공백** — 최우선 |

### 1.1 목표

오늘 중 2-4편 추가 발행으로:
- `philosophy` 0 → 최소 1편 (카테고리 정체성 확립)
- `engineering` 1 → 2편 이상 (technical depth 층 보강)
- `tools`와 `vibe-coding`은 상호 링크 강화 목적이면 추가 가능

### 1.2 제약

- 모든 글이 **새 2-layer taxonomy**에 맞게 category + tags 설정
- 기존 layout contract (`.claude/rules/blog-content.md §7.1`) 준수
- 각 글 QA: Chrome MCP 레이아웃 검증 + build/sitemap/RSS 확인
- 상호 링크 최소 2개 (내부)

---

## 2. 후보 글 6개

### C1. Philosophy — "Why I built tene: the case for local-first in the AI era"

| 항목 | 값 |
|---|---|
| Category | `philosophy` |
| Tags | `tene`, `founder-story`, `security` |
| 예상 길이 | ~1,400 words |
| 난이도 | 중 (설득력 있게 써야) |
| Hook | "I paid Doppler for 18 months. Then I wrote `tene init` one weekend." |
| 핵심 주장 | Cloud secret manager가 아니라 local-first를 택한 이유 3가지 (AI 위협 모델 + Single founder 관리 가능성 + 데이터 주권) |
| 가치 | **Founder voice 첫 글**. 구글 검색 롱테일 + HN 공유 친화적 |

### C2. Engineering — "Argon2id + OS Keychain: how tene stores the master key"

| 항목 | 값 |
|---|---|
| Category | `engineering` |
| Tags | `cryptography`, `architecture`, `security`, `tene` |
| 예상 길이 | ~1,400 words |
| 난이도 | 고 (기술 정확도 필요) |
| Hook | "Your password is not the key. Your password *derives* the key." |
| 핵심 주장 | KDF (Argon2id) + OS keychain integration (macOS Keychain / libsecret / Credential Vault) 아키텍처 |
| 가치 | XChaCha20 글의 **짝 글**. 암호학 독자 retention. `engineering` 카테고리 강화 |

### C3. Vibe Coding — "Running tene in GitHub Actions"

| 항목 | 값 |
|---|---|
| Category | `vibe-coding` |
| Tags | `tene`, `devsecops`, `tutorial`, `cli` |
| 예상 길이 | ~1,100 words |
| 난이도 | 중 (실제 workflow YAML 필요) |
| Hook | "Your CI has secrets. So does your IDE. They should never meet." |
| 핵심 주장 | GitHub Actions + `TENE_MASTER_PASSWORD` secret + `--no-keychain` 플래그 + `tene run --` 패턴 |
| 가치 | CI/CD 실무 니즈. **"tene in production" SEO 대응**. 현재 CI 관련 글 0편 |

### C4. Tools — "tene + bkit: the full AI-agent dev loop"

| 항목 | 값 |
|---|---|
| Category | `tools` |
| Tags | `tene`, `bkit`, `workflow`, `tutorial` |
| 예상 길이 | ~1,600 words |
| 난이도 | 중 |
| Hook | "bkit plans the feature. tene injects the keys. Claude writes the code — and never sees either." |
| 핵심 주장 | `/pdca plan` → `/pdca do` + `tene run -- <test>` 통합 워크플로우 |
| 가치 | 오늘 올린 bkit 2편의 **3부작 완성편**. 내부 링크 밀도 +30% |

### C5. Philosophy — "The AI-agent secret leak problem in 2026: state of the industry"

| 항목 | 값 |
|---|---|
| Category | `philosophy` |
| Tags | `security`, `devsecops`, `tene` |
| 예상 길이 | ~2,000 words |
| 난이도 | 고 (리서치 + 사실 확인) |
| Hook | "In 2024, AI agents accidentally logged secrets. In 2026, agents accidentally publish them to GitHub." |
| 핵심 주장 | 업계 전반 상태 리뷰: Copilot 유출 사건, Claude Code 업데이트 트렌드, 시장 반응 |
| 가치 | **Pillar / manifesto 성격**. HN Show HN 후보. Dev.to 크로스포스트 최적. 단 **리서치 부담 큼** |

### C6. Vibe Coding — "Windsurf + tene"

| 항목 | 값 |
|---|---|
| Category | `vibe-coding` |
| Tags | `tene`, `security`, `tutorial` |
| 예상 길이 | ~900 words |
| 난이도 | 낮 (기존 Claude Code / Cursor 글 템플릿 복제) |
| Hook | "Windsurf is Cursor's quiet cousin. Here's the same safe pattern for it." |
| 핵심 주장 | `.windsurfrules` 파일 기반 rule + `tene run -- windsurf` |
| 가치 | 편집기 매트릭스 3번째. Landing의 "Supported AI Editors" 표 보강 |

---

## 3. 각 글 의 예상 outline (간략)

### C1 — Why I built tene (philosophy, ~1,400w)

```
H1: Why I built tene (frontmatter title)

H2 1. Eighteen months on Doppler  (hook + 개인 경험)
H2 2. The AI agent changed the threat model
H2 3. Why "local-first" isn't nostalgia — it's operational
H2 4. What I gave up (honest trade-offs)
H2 5. What I'll build next (local-first 확장 로드맵)
H2 6. Closing — if your secrets are for humans, stay cloud. If they're for agents, come local.

Code blocks: 2 (cost table + tene init output)
MDX components: Callout × 1 (투자자/팀 고려사항)
Internal links: /blog/doppler-alternative-journey, /blog/ai-reads-env
```

### C2 — Argon2id + OS Keychain (engineering, ~1,400w)

```
H2 1. The password is not the key
H2 2. Argon2id in one paragraph (Why not PBKDF2/bcrypt/scrypt)
H2 3. Parameters that actually matter (memory 64MB / time 3 / lanes 1)
H2 4. Key cache: where macOS / Linux / Windows store the derived key
H2 5. The handshake: password -> KDF -> cache -> decrypt vault
H2 6. Rotating / revoking (passwd, recover, full rewrap)
H2 7. Summary — 3 invariants tene protects

Code blocks: 4 (Argon2id params, keychain API calls per OS, rotate command, test)
Internal links: /blog/xchacha20-for-devs, /blog/ai-reads-env
```

### C3 — tene in GitHub Actions (vibe-coding, ~1,100w)

```
H2 1. Two secret stores, one leak
H2 2. What GitHub Actions' default setup misses
H2 3. tene + Actions: the 4-step setup
H2 4. Example: Stripe webhook test in CI
H2 5. Rotation without restarting CI
H2 6. Gotchas (keychain absence, masked stdout)

Code blocks: 3 (workflow YAML, tene run in CI, rotation)
Internal links: /blog/claude-code-safe-api-keys, /blog/migrate-env-60s
```

### C4 — tene + bkit full loop (tools, ~1,600w)

```
H2 1. The gap between "plan" and "run"
H2 2. bkit plans the feature (/pdca pm -> design)
H2 3. tene holds the keys (.tene/ + CLAUDE.md rule)
H2 4. /pdca do with runtime injection (tene run -- npm test)
H2 5. /pdca analyze + gap-detector (no secrets in design docs)
H2 6. Completing the loop: ship safely
H2 7. Takeaway — method + primitives = reliable AI-agent dev

Code blocks: 4 (PDCA flow, tene-injected test, CLAUDE.md excerpt, gap-detector output)
Internal links: /blog/bkit-harness-engineering-vibe-coding, /blog/bkit-pdca-for-claude-code, /blog/claude-code-safe-api-keys
```

### C5 — AI-agent secret leak state (philosophy, ~2,000w)

```
H2 1. 2024: the year agents learned to log
H2 2. 2025-early: the GitHub Copilot incidents (cite specific cases)
H2 3. 2025-late: the Claude Code transition
H2 4. 2026 now: three patterns the industry converged on
H2 5. What's still broken (agent output logs, cloud console copy-paste)
H2 6. What tene addresses, what it does not (honest scope)
H2 7. Forecast — where this goes in 2027

Code blocks: 2 (leak reproduction example, scope matrix)
Internal links: /blog/ai-reads-env, /blog/doppler-alternative-journey
External refs: 6-8 news/blog posts (주의: 사실 확인 필요)
```

### C6 — Windsurf + tene (vibe-coding, ~900w)

```
H2 1. Windsurf: Cursor's quieter cousin
H2 2. The .windsurfrules file
H2 3. tene init --windsurf (if command exists, else manual)
H2 4. Running with injected secrets
H2 5. Common mistakes
H2 6. Summary

Code blocks: 3 (install, rule file, run)
Internal links: /blog/cursor-secret-management-2026, /blog/claude-code-safe-api-keys
```

---

## 4. 권장 조합 3종

### 조합 A (최소 — 2편, ~4-5h) ⭐ 권장

- **C1** (philosophy gap 해소 · founder voice 첫 글)
- **C4** (bkit 3부작 완성 · 내부 링크 밀도 강화)

**근거**: 오늘 추가 효과가 가장 큰 2편. 각 카테고리 성격 확립 + 기존 콘텐츠 가치 증폭.

### 조합 B (균형 — 3편, ~7-8h)

- **C1** (philosophy 1편)
- **C2** (engineering 1편)
- **C4** (tools + internal links)

**근거**: 3개 카테고리 모두 +1 만들어서 새 taxonomy 도입 시점에 분포 정돈. `philosophy` 1 · `engineering` 2 · `tools` 4 · `vibe-coding` 5. 고르게 분포.

### 조합 C (풀 세트 — 4편, ~10h)

- **C1** + **C2** + **C3** + **C4**

**근거**: 하루에 네 카테고리 모두 +1. 피로하지만 taxonomy 첫 공개에 최대 임팩트.

### 조합 D (manifesto 도전 — 1편, ~5h)

- **C5** 단독

**근거**: state-of-industry 에세이는 단독 임팩트 글. 다만 리서치 부담 크고 팩트 확인 필요해 실패 리스크 존재.

---

## 5. 공통 품질 기준 (모든 글 공통)

작성 전 확정 (Phase 1 Plan):
- [ ] Category 1개 선택 (4개 중)
- [ ] Tags 2-4개 (15개 vocabulary에서)
- [ ] Slug 승인 (≤ 60자, kebab-case)
- [ ] Length 목표 (800/1,200/2,000 중)
- [ ] Opening hook 1줄

작성 (Phase 3 Do):
- [ ] Frontmatter 완전 (`category` 필수)
- [ ] H2 4-8개, 각 ASCII 문자
- [ ] Code block 실제 tene/bkit 명령만
- [ ] Internal 링크 ≥ 2
- [ ] FAQ ≥ 3 (BlogPosting + FAQPage JSON-LD)
- [ ] OG image (slug별 `/public/blog/<slug>/` 또는 재사용 `/public/demo/*`)

검증 (Phase 4 Check + Phase 6 QA):
- [ ] `next build` 성공 · SSG 라우트 생성
- [ ] `grep "@type\":\"BlogPosting\"" .next/server/app/blog/<slug>.html` → 1
- [ ] `grep "@type\":\"FAQPage\"" .next/...` → 1
- [ ] Chrome MCP layout 16 체크 통과
- [ ] 카테고리 pill에 포함됨 (신설)
- [ ] tag 필터에서 클릭 시 해당 글 표시

---

## 6. 타임라인

### Day 0 (오늘, 선택 조합에 따라)

**조합 A 선택 시**:
- T+0h: C1 Plan (Phase 1-2, 15m)
- T+0.25h: C1 작성 (Phase 3, 90m)
- T+1.75h: C1 검증 + commit (30m)
- T+2.25h: C4 Plan (15m)
- T+2.5h: C4 작성 (120m)
- T+4.5h: C4 검증 + commit (30m)
- T+5h: PR staging

**조합 B 선택 시**: 위 + C2 (60m Plan + 2.5h Do + 30m QA = 4h 추가)

**조합 C 선택 시**: 위 + C3 (45m Plan + 2h Do + 30m QA = 3.5h 추가)

---

## 7. 의존 순서

```
[blog-categories-and-tooling.plan.md 완료]
  ├─ 카테고리 도입 PR merge to staging
  ├─ migrations 완료 (9편 frontmatter)
  └─ /blog-new 스킬 업데이트 (category 질문 추가)
        │
        ▼
[blog-content-batch-2026-04-23.plan.md 시작]
  ├─ 조합 결정 (A/B/C/D)
  ├─ 각 글 /blog-new 호출 (새 스킬로, category 선택됨)
  ├─ 글별 commit
  └─ 단일 PR staging → main
```

**중요**: 카테고리 도입이 먼저 완료돼야 새 글이 정상 분류됨. 두 plan을 동일 PR에 묶지 않음 (스코프 분리 + 리스크 격리).

---

## 8. 측정 (발행 후 7일)

| 지표 | 목표 |
|------|------|
| 신규 글 impressions (Google Search Console) | ≥ 100 per article |
| `/blog/category/*` 4개 URL 인덱싱 | 전부 "Indexed" 상태 |
| 기존 9편 ranking 변동 | ±3위 이내 (유지) |
| HN / Reddit 공유 수 (manual 추적) | 조합 A ≥ 10, 조합 C ≥ 30 |
| 블로그 → tene.sh 홈 CTR | ≥ 3% |

---

## 9. 참조

- 선행 plan: `docs/01-plan/blog-categories-and-tooling.plan.md`
- Content rules: `tene-biz/.claude/rules/blog-content.md`
- PDCA workflow: `tene-biz/.claude/rules/blog-pdca-workflow.md`
- 기존 글: `apps/web/content/blog/*.mdx` (9편)
- Growth routine: `tene-biz/.claude/rules/growth-routine.md` (발행 후 cross-share)

---

**작성**: 2026-04-23 (Claude main, discussion-based)
**대기**: 조합 A/B/C/D 중 결정 + 선행 plan 승인
