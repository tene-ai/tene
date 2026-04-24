# Blog TTS — Plan

| 항목 | 값 |
|------|-----|
| Feature | `blog-tts` (블로그 글 음성 재생 기능) |
| 기반 리서치 | 2026-04-24 /pdca team 병렬 조사 (ttsreader + tene/apps/web) |
| 브랜치 | `feature/blog-tts` (origin/staging 베이스) |
| 후속 | design → do → QA → report |
| 총 공수 | ~6-8 인시간 (MVP 기준) |

---

## 1. 목적 (Why)

**가설**: 블로그 글이 길어지고 있음 (최근 2편 각 ~2,500 단어, 10-11분 read). 이동 중/멀티태스킹 중 **귀로 듣는 옵션**이 있으면 retention + engagement 상승. 특히 `philosophy` 카테고리 에세이성 글은 audio-first 소비에 적합.

**제약 (사용자 확정)**: **비용 0**. 외부 TTS API (OpenAI/ElevenLabs/Google Cloud)를 쓰지 않음. 브라우저 내장 **Web Speech API** 로만 구현.

**영감 source**: [ttsreader.com](https://ttsreader.com) + [RonenR/ttsreader](https://github.com/RonenR/ttsreader) 엔진 — 같은 Web Speech API 패턴 사용.

## 2. 기술 선택 근거

### 2.1 Web Speech API 를 택한 이유

`window.speechSynthesis` + `SpeechSynthesisUtterance` — 모든 주요 브라우저 내장:

- ✅ **Zero cost** — API 키/서버/스토리지 전혀 없음
- ✅ **Zero backend** — Next.js 정적 빌드 그대로 OK
- ✅ **Offline voices** — macOS/Windows/Android 시스템 보이스 사용
- ✅ **MIT-level 자유도** — 표준 웹 API, 종속성 없음
- ⚠️ **품질 편차** — OS/브라우저마다 voice quality 상이
- ⚠️ **Chrome 15초 버그** — 단일 utterance가 15초 후 끊김. 문장 단위 chunking 필요
- ⚠️ **iOS 제스처 제약** — 첫 `speak()`는 user-gesture 핸들러 내에서 동기 호출 필요
- ⚠️ **Firefox boundary 이벤트 불안정** — 단어 하이라이트는 cross-browser 불가. 문장 하이라이트로 폴백

### 2.2 대안 비교 (검토 후 기각)

| 대안 | 품질 | 비용 | 기각 사유 |
|---|:-:|---|---|
| OpenAI TTS (pre-generated) | ⭐⭐⭐⭐⭐ | 편당 ~$0.15 | 사용자가 비용 0 요구 |
| ElevenLabs pre-gen | ⭐⭐⭐⭐⭐ | ~$0.30/편 | 동일 이유 |
| Google Cloud TTS | ⭐⭐⭐⭐ | pay-per-char | 동일 이유 |
| ttsreader 엔진 (RonenR/ttsreader lib) 직접 import | ⭐⭐⭐ | 0 | 라이선스 명시 불분명(LICENSE 파일 미확인) + Next.js 타입 정합성 저하 → **패턴만 차용** |
| **Web Speech API 직접 구현** | ⭐⭐⭐ | **0** | ✅ **선택** |

### 2.3 ttsreader 엔진에서 배운 핵심 패턴 (인용하되 직접 import 안 함)

1. **문장 단위 chunking** — 정규표현식 `[\s\S]{1,160}[.!?,](?=\s|$)|[\s\S]{1,160}` 로 ~160자 단위로 자름. Chrome 15초 버그 우회.
2. **순차 재생 큐** — 각 chunk의 `utterance.onend` 에서 다음 chunk `speak()` 호출.
3. **Voice filtering by language prefix** — `v.lang.startsWith('en')` 로 필터링.
4. **iOS 호환 제스처 유지** — 클릭 핸들러 내부에서 비동기 없이 `speechSynthesis.speak()` 호출.
5. **Rate clamp** — Chrome 한계 때문에 `[0.5, 2.0]` 범위로 제한.

## 3. 스코프

### 3.1 MVP 포함

| # | 항목 | 설명 |
|:-:|---|---|
| M1 | **Listen 버튼** | `PostHero` 하단 — 설명 다음 줄에 "Listen" 버튼 추가 |
| M2 | **TTS player 컴포넌트** | `<TTSPlayer>` client component — play/pause/stop/speed slider(0.5-2x)/voice dropdown |
| M3 | **문장 chunking + 순차 재생** | Chrome 15초 버그 회피. 160자 단위로 자르고 큐 처리 |
| M4 | **Voice selection** | `getVoices()` 에서 en-* 로 시작하는 보이스 필터링, 사용자 선택 dropdown |
| **M4b** | **Voice persistence (localStorage)** | 선택한 voice + rate 를 `localStorage` 에 저장. 다음 방문/다음 글에서 같은 선택 유지. 키 namespace: `tene:blog-tts:{voice,rate}` |
| M5 | **Code block skip** | MDX `<pre>` / `<code>` 블록은 읽지 않음. 텍스트 추출 시 제외 + 스크롤 sync 시에도 건너뜀 |
| M6 | **Progress indicator** | 현재 chunk 번호 / 총 chunk 수 표시 |
| **M6b** | **Auto-scroll sync (paragraph-level)** | TTS 가 현재 읽는 문단(`<p>`/`<h2>`/`<li>`)을 `scrollIntoView({behavior:'smooth', block:'center'})` 로 화면 중앙에 배치 + 시각 하이라이트 CSS 클래스 토글. 코드블록은 skip이므로 다음 문단으로 건너뜀. 사용자 스크롤 개입 감지 시 auto-scroll 일시정지 |
| M7 | **Analytics events** | `blog_tts_play`, `blog_tts_pause`, `blog_tts_complete`, `blog_tts_rate_change`, `blog_tts_voice_change` 등록 + 발사 |
| M8 | **Unload cleanup** | 페이지 이탈 시 `speechSynthesis.cancel()` — 다음 페이지에서 이전 글 음성이 남는 현상 방지 |

### 3.2 MVP 미포함 (후속 phase B)

- ❌ **단어 레벨 하이라이트** (`boundary` 이벤트 기반) — Firefox/Android 지원 불안정. 복잡도 큼. 문단 레벨(M6b)로 충분히 follow-along UX 제공
- ❌ **재생 위치 저장 (seek/resume)** — 첫 릴리스 후 사용자 피드백 보고 판단
- ❌ **다국어 자동 감지** — 현재 모든 글이 `en-US`. 한국어 글 추가 시 재검토
- ❌ **키보드 단축키** (Space, Esc) — a11y 개선 사이클에서
- ❌ **모바일 floating mini-player** — MVP는 단일 Listen 버튼

### 3.3 미포함 (의도적 제외)

- 외부 TTS API 연동 — 비용 제약으로 완전 배제
- 빌드 타임 mp3 pre-generation — 동일 이유
- 오디오 파일 다운로드 — 사용자 요구 없음

## 4. 설계 결정 요약 (Design phase 에서 상세화 예정)

### 4.1 Integration point

리서치 결과 4개 후보 중 **#1 (PostHero 하단 Listen 버튼)** 채택:

- 발견성 최고: 글 상단 fold 안에 보임
- 모바일+데스크톱 커버
- 공수 최저 — PostHero 수정 1곳 + 신규 컴포넌트 1개
- 리스크 최저 — 기존 레이아웃 계약 (`max-w-3xl`, `pt-4`, 블로그 §7.1) 불변

### 4.2 Component split

```
<PostHero>  (server component)
  └─ <TTSButton>  (client component, "Listen" 버튼 트리거)
         ↓ 클릭 시 mount
      <TTSPlayer>  (client component, 실제 재생 로직)
```

글 상세 DOM 에서 텍스트 추출은 `<article className="min-w-0">` 컨테이너(page.tsx:129)를 query selector로 잡음. `<pre>` 자식은 제외.

### 4.3 Text extraction + chunk↔DOM mapping strategy

**Chunking이 DOM 블록과 맞물려야 scroll sync 가능**:

```ts
interface TTSChunk {
  text: string;          // 실제 읽을 160자 이내 문장 단위
  blockEl: HTMLElement;  // 속한 <p>/<h2>/<h3>/<li>
  blockIndex: number;    // 전체 블록 순번 (scroll-into-view 대상)
}

function extractChunks(articleEl: HTMLElement): TTSChunk[] {
  const blocks = articleEl.querySelectorAll('p, h2, h3, h4, li, blockquote');
  const chunks: TTSChunk[] = [];
  let blockIndex = 0;
  for (const blockEl of blocks) {
    // 코드블록 포함 요소는 skip
    if (blockEl.querySelector('pre, code')) continue;
    const text = blockEl.textContent?.trim() ?? '';
    if (!text) continue;
    const sentences = chunkText(text, 160);  // 기존 chunking 함수
    for (const s of sentences) {
      chunks.push({ text: s, blockEl: blockEl as HTMLElement, blockIndex });
    }
    blockIndex++;
  }
  return chunks;
}
```

### 4.4 Scroll sync 전략

**접근**: 문단 단위. chunk 의 onstart 에서 `blockEl.scrollIntoView({behavior:'smooth', block:'center'})`.

**사용자 스크롤 개입 감지**:
- user-initiated scroll (wheel, touch, keyboard arrow) 감지 → auto-scroll 2초 suspend
- 2초 내 재개입 없으면 재개
- 구현: `scroll` 이벤트 listener + debounced flag

**블록 전환 감지**:
- 각 utterance 의 `onstart` 이벤트에서 `chunk.blockIndex` 가 직전 chunk 와 다를 때만 scroll 실행
- 같은 문단 내 여러 chunk 가 연속이면 scroll 1회만

**시각 하이라이트**:
```css
.tts-active-block {
  background: color-mix(in oklab, var(--accent) 8%, transparent);
  border-left: 2px solid var(--accent);
  padding-left: 1rem;
  margin-left: -1.125rem;  /* 기존 position 유지 */
  transition: background 200ms ease;
}
```

### 4.5 Voice/Rate 지속성 (M4b)

```ts
const STORAGE_KEYS = {
  voice: 'tene:blog-tts:voice',  // voiceURI 저장
  rate: 'tene:blog-tts:rate',     // string ("1.25") 저장
} as const;

function loadPrefs() {
  if (typeof window === 'undefined') return { voiceURI: null, rate: 1 };
  return {
    voiceURI: localStorage.getItem(STORAGE_KEYS.voice),
    rate: parseFloat(localStorage.getItem(STORAGE_KEYS.rate) ?? '1') || 1,
  };
}

function savePrefs({ voiceURI, rate }: { voiceURI?: string; rate?: number }) {
  if (typeof window === 'undefined') return;
  if (voiceURI) localStorage.setItem(STORAGE_KEYS.voice, voiceURI);
  if (rate != null) localStorage.setItem(STORAGE_KEYS.rate, String(rate));
}
```

**기본값 선정 로직**:
1. `localStorage` 에 저장된 `voiceURI` 가 있고 현재 `getVoices()` 결과에도 존재 → 그것 사용
2. 없으면 `en-*` 보이스 중 첫 번째
3. 그것도 없으면 `null` (browser default)

### 4.4 Utterance chunking

```ts
function chunkText(text: string, max = 160): string[] {
  const re = new RegExp(`[\\s\\S]{1,${max}}[.!?,](?=\\s|$)|[\\s\\S]{1,${max}}`, 'g');
  return text.match(re) ?? [text];
}
```

~160자 chunk. 2,500단어 글 기준 약 80-100개 chunk. Chunk당 평균 12-15초 재생 → Chrome 버그 회피.

### 4.6 Analytics event 설계

`src/lib/track.ts` EventMap 확장:
```ts
blog_tts_play:          { slug: string; readingMinutes: number };
blog_tts_pause:         { slug: string; percentRead: number };
blog_tts_resume:        { slug: string; percentRead: number };
blog_tts_complete:      { slug: string; readingMinutes: number };
blog_tts_rate_change:   { slug: string; rate: number };
blog_tts_voice_change:  { slug: string; voiceName: string };
```

## 5. 성공 기준

### 5.1 Functional

- [ ] macOS Safari/Chrome/Firefox 에서 11편 모두 재생됨 (맨 앞 30초 + 맨 뒤 30초 샘플)
- [ ] 재생 중 Pause → Resume 정상 동작
- [ ] Stop 시 state 초기화
- [ ] Speed 0.75x/1x/1.25x/1.5x 전환 실시간 반영
- [ ] Voice dropdown에 최소 1개 이상 en-* voice 노출
- [ ] **Voice 선택 후 재방문 시 동일 voice 자동 선택됨** (localStorage)
- [ ] **Rate 선택 후 재방문 시 동일 rate 자동 선택됨** (localStorage)
- [ ] **현재 읽는 문단이 smooth scroll 로 화면 중앙 배치됨** (M6b)
- [ ] **현재 문단에 시각적 하이라이트 (`.tts-active-block`) 적용됨**
- [ ] **사용자 수동 scroll 시 auto-scroll 2초간 suspend 후 재개**
- [ ] **코드블록은 skip + auto-scroll 에서도 건너뜀**
- [ ] 페이지 나갈 때 음성 자동 종료 (`cancel()`) + `tts-active-block` class 정리
- [ ] 인터넷 끊긴 상태에서도 재생 가능 (offline voice 사용)

### 5.2 UX

- [ ] "Listen" 버튼이 PostHero 하단에 명확히 보임 (모바일 375px, 데스크톱 1920px 모두)
- [ ] 버튼 클릭 후 1초 이내 첫 utterance 시작
- [ ] player UI가 기존 Tailwind 토큰 (`bg-surface`, `text-accent`, `border-border`) 일관 사용
- [ ] 버튼의 accessible label (`aria-label="Listen to article"`) 부여

### 5.3 Non-functional

- [ ] 번들 크기 추가 < 5 KB gzipped (client component만 client 번들로 split)
- [ ] 첫 `speak()` 호출 전까지 `speechSynthesis` 초기화 없음 (lazy)
- [ ] TypeScript 컴파일 에러 0
- [ ] Next.js build 성공 (기존 SSG 페이지 영향 없음)
- [ ] Console errors 0

## 6. 타임라인

| 단계 | 예상 | 산출물 |
|---|:--:|---|
| Plan (본 문서) | 30m | `docs/01-plan/blog-tts.plan.md` |
| Design | 45m | `docs/02-design/features/blog-tts.design.md` — 컴포넌트 API + 상태 머신 + 엣지 케이스 |
| Do | 2-3h | `src/components/blog/tts-*.tsx` 2-3 파일, `post-hero.tsx` + `track.ts` 수정 |
| QA | 1h | Chrome MCP 자동 재생 + 수동 실측 (여러 브라우저) |
| Report | 30m | `docs/03-report/blog-tts-YYYY-MM-DD.md` |
| Commit + PR | 15m | PR → staging |
| **Total** | **~5-6h** | |

## 7. 리스크 & 완화

| 리스크 | 확률 | 영향 | 완화 |
|---|:-:|:-:|------|
| Chrome 15초 버그로 중간 끊김 | 중 | 고 | 160자 chunk + onend 큐 (ttsreader 패턴) |
| iOS Safari 첫 `speak()` 차단 | 중 | 고 | user-gesture handler 내부에서 동기 호출. 제스처 끊김 피하도록 async/await 사용 X |
| `getVoices()` 빈 배열 반환 (Chrome 초기) | 높 | 중 | `onvoiceschanged` 리스너 + useEffect 재시도 |
| 긴 글 (2,500단어) 재생 시 배터리 소모 | 낮 | 저 | 경고 표시 불필요 (사용자가 능동 재생) |
| `boundary` 이벤트 Firefox에서 미발사 | 높 | 저 | MVP는 단어 하이라이트 안 하므로 무관 |
| MDX 이미지 alt text가 읽힘 | 낮 | 저 | `textContent` 추출은 alt text 포함 안 함 (표준 DOM 동작) |
| 독자가 끄는 걸 잊고 다음 글로 이동 → 두 음성 겹침 | 낮 | 중 | Next.js route change listener + `cancel()` |
| CSP 위반 | 낮 | 중 | Web Speech API는 외부 요청 없음. CSP 무관. `next.config.ts` 검증 완료 (code analysis §G) |

## 8. 배포 플로우

```
feature/blog-tts  (staging 기준)
    ├─ commit 1: analytics EventMap 확장
    ├─ commit 2: TTSPlayer 컴포넌트 + TTSButton 컴포넌트
    ├─ commit 3: PostHero 통합 + build verify
    └─ PR → staging → CI → merge
         ↓
      main PR 은 사용자 검토 후 직접
```

## 9. Check phase (QA) 계획 미리보기

Design 완료 후 상세화 예정. 핵심 항목:

### 9.1 Chrome MCP 자동 검증

- /blog/{slug} navigate → "Listen" 버튼 렌더 확인
- 버튼 클릭 → `speechSynthesis.speaking` true 확인
- Pause 클릭 → `paused` true 확인
- 다른 글로 이동 → 이전 글 음성 cancel 확인

### 9.2 수동 cross-browser

- macOS Safari 17+
- macOS Chrome (Apple Silicon)
- macOS Firefox
- iOS Safari (실기기 권장)

각 브라우저에서:
- 첫 30초 재생 OK
- Chunk 경계에서 끊김 없음
- Speed slider 실시간 반영

### 9.3 Layout regression

- Chrome MCP `blog-content.md §7.1` 계약 유지 — hero max-w-3xl, articleWidth 720, gridCols "720px 220px"
- 모바일 375px 가로 스크롤 0
- Listen 버튼이 네 카테고리 글 모두에서 동일 위치

## 10. 의존성 / 선행 조건

- **선행**: 없음 (blog-categories, release-pipeline-resilience 모두 staging에 머지 완료, main 기다림)
- **병행 가능**: homebrew 재활성 결정 — 별개 이슈
- **후속**:
  - Phase B — 단어/문장 하이라이트 + 재생 위치 저장 (MVP 사용 데이터 보고 결정)
  - Phase C — 한국어 글 추가 시 다국어 voice auto-detect
  - Phase D — 모바일 floating mini-player (사용자 요청 들어오면)

## 11. Out-of-scope 명시

이 PDCA 에서 다루지 않는 것:

- 오디오 파일 다운로드 기능 (MP3 export)
- 재생 속도 외 pitch/volume 조정
- 재생 중 광고 삽입 / 후원 음성
- 소셜 공유용 오디오 snippet 생성
- Podcast RSS feed (팟캐스트 앱 구독 대응)

필요해지면 각각 별도 plan 문서로.

## 12. 참조 (리서치 원본)

- ttsreader hosted: https://ttsreader.com/
- ttsreader 웹사이트 repo (Hugo): https://github.com/ttsreader/ttsreader-web
- ttsreader 엔진 repo (RonenR, vanilla JS): https://github.com/RonenR/ttsreader
- Chromium 15초 버그: #679437, #335907
- 3rd-party 폴리필 레퍼런스: https://github.com/leaonline/easy-speech
- Web Speech API 스펙 draft: https://webaudio.github.io/web-speech-api/
- tene 기존 PostHero: `apps/web/src/components/blog/post-hero.tsx`
- tene 기존 page: `apps/web/src/app/blog/[slug]/page.tsx:129` (article container)
- tene track EventMap: `apps/web/src/lib/track.ts:6-55`
- tene next.config CSP: `apps/web/next.config.ts:9-16` — Web Speech 무관, 변경 불필요
- tene tsconfig: dom/dom.iterable 포함 — `speechSynthesis` 타입 자동 resolve

---

**작성**: 2026-04-24 (Claude main + /pdca team 병렬 리서치 2건)
**승인 대기**: 사용자 확인 후 Design 단계 진입
