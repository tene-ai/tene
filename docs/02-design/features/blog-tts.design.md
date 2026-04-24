# Blog TTS — Design

| 항목 | 값 |
|------|-----|
| Feature | `blog-tts` |
| Plan | `docs/01-plan/blog-tts.plan.md` |
| 브랜치 | `feature/blog-tts` |
| 리서치 기반 | /pdca team 2 agents (2026-04-24): scroll sync tech + tene 코드베이스 |

---

## 1. 아키텍처 오버뷰

```
apps/web/src/
├── components/blog/
│   ├── post-hero.tsx            [수정] — server component. <TTSButton> import + 렌더
│   ├── tts-button.tsx           [신규] — client. 글 하단 "Listen" 트리거. 처음 누르면 <TTSPlayer> mount
│   ├── tts-player.tsx           [신규] — client. 실제 재생 로직, UI, 상태 관리
│   └── tts-controls.tsx         [신규] — client. play/pause/stop/speed/voice UI (presentational)
├── lib/
│   ├── tts/
│   │   ├── chunks.ts            [신규] — extractChunks + chunkText 순수 함수
│   │   ├── prefs.ts             [신규] — localStorage load/save (namespaced tene:blog-tts)
│   │   └── voices.ts            [신규] — getVoices 구독 hook + en-* 필터
│   └── track.ts                 [수정] — EventMap에 TTS 이벤트 6종 추가
├── hooks/
│   ├── use-tts.ts               [신규] — speech synthesis 생명주기 훅
│   └── use-scroll-sync.ts       [신규] — 문단 scrollIntoView + user-scroll 감지
└── app/
    ├── blog/[slug]/page.tsx     [영향 없음] — 기존 PostHero 호출 유지
    └── globals.css              [수정] — .tts-active-block 클래스 + scroll-margin 추가
```

**요약**: 컴포넌트 3개 + 훅 2개 + 유틸 3개 + 전역 CSS 1블록 + track.ts 확장 + post-hero.tsx 1줄 추가.

예상 client 번들 증가: **~4 KB gzipped** (Plan 목표 < 5 KB 충족).

---

## 2. 컴포넌트 설계

### 2.1 `<TTSButton>` — 진입점

**파일**: `src/components/blog/tts-button.tsx`
**타입**: Client component
**Props**:
```ts
type TTSButtonProps = {
  slug: string;          // analytics용
  readingMinutes: number; // analytics payload
};
```

**동작**:
- 렌더 초기: `<button>Listen</button>` 하나만
- 클릭 시: `<TTSPlayer>` 를 lazy mount + 자동 play 시작
- 이유: unmount 상태에서는 speechSynthesis API에 닿지 않음 → 번들/성능 비용 거의 0

**UI (Tailwind)**:
```tsx
<button
  aria-label="Listen to article"
  className="inline-flex items-center gap-2 rounded-md border border-border bg-surface/60 px-3 py-1.5 text-sm text-muted backdrop-blur-sm transition-colors hover:border-accent/40 hover:text-foreground"
>
  <PlayIcon className="h-4 w-4" />
  Listen
</button>
```

아이콘: 인라인 SVG (lucide-react 같은 런타임 의존성 추가 없음 — 기존 코드베이스 convention).

### 2.2 `<TTSPlayer>` — 재생 오케스트레이터

**파일**: `src/components/blog/tts-player.tsx`
**타입**: Client component
**Props**:
```ts
type TTSPlayerProps = {
  slug: string;
  readingMinutes: number;
  onClose?: () => void;  // 명시적 중단 시 부모에게 unmount 신호
};
```

**책임**:
1. `useTTS()` 훅으로 재생 라이프사이클 관리
2. `useScrollSync()` 훅으로 문단 자동 추적
3. `<TTSControls>` 에 상태 + 콜백 전달
4. 페이지 언마운트 시 cleanup

**상태 (하위 훅이 소유, Player는 orchestrator)**:
- `chunks: TTSChunk[]` — `useRef` (재렌더 트리거 안 함)
- `currentChunkIndex: number` — `useState`
- `playbackState: 'idle' | 'playing' | 'paused'`
- `voices: SpeechSynthesisVoice[]`
- `selectedVoiceURI: string | null`
- `rate: number`

**Chunk 추출 타이밍**:
- Player mount 직후 `useLayoutEffect` 에서 `extractChunks(articleEl)` 1회 실행
- 결과를 `chunksRef.current` 에 저장
- `articleEl = document.querySelector('article.min-w-0')` — `page.tsx:129` 에 이미 안정적 셀렉터 존재

### 2.3 `<TTSControls>` — 프리젠테이셔널 UI

**파일**: `src/components/blog/tts-controls.tsx`
**타입**: Client component (순수 UI, 로직 없음)
**Props**:
```ts
type TTSControlsProps = {
  state: 'idle' | 'playing' | 'paused';
  progress: { current: number; total: number };
  rate: number;
  voices: SpeechSynthesisVoice[];
  selectedVoiceURI: string | null;
  onPlay: () => void;
  onPause: () => void;
  onResume: () => void;
  onStop: () => void;
  onRateChange: (rate: number) => void;
  onVoiceChange: (voiceURI: string) => void;
};
```

**UI 레이아웃** (PostHero 내부, description 아래 inline card):
```
┌─────────────────────────────────────────────────────────────┐
│  ▶ Play    0:00 / 11:32       [●────────○]  Chunk 12/87    │
│                                                             │
│  Speed  [1x ▼]   Voice  [Samantha (en-US) ▼]   ✕ Close     │
└─────────────────────────────────────────────────────────────┘
```

- 상단: Play/Pause 토글 + 시간 표시 (예상; `currentChunk * avgSec`) + progress bar + chunk counter
- 하단: Speed select (0.5 / 0.75 / 1 / 1.25 / 1.5 / 2) + Voice dropdown + Close 버튼

**토큰**: 기존 tene 스타일과 동일 (`bg-surface/60`, `border-border`, `text-accent`).

---

## 3. 훅 설계

### 3.1 `useTTS()` — speech synthesis 생명주기

**파일**: `src/hooks/use-tts.ts`

**Signature**:
```ts
interface UseTTSOptions {
  chunks: React.RefObject<TTSChunk[]>;
  voice: SpeechSynthesisVoice | null;
  rate: number;
  onChunkChange?: (index: number) => void;
  onComplete?: () => void;
  onError?: (err: SpeechSynthesisErrorEvent) => void;
}

interface UseTTSReturn {
  state: 'idle' | 'playing' | 'paused';
  currentChunkIndex: number;
  play: () => void;
  pause: () => void;
  resume: () => void;
  stop: () => void;
}

function useTTS(options: UseTTSOptions): UseTTSReturn;
```

**내부 동작**:

```ts
// 1. 첫 play 클릭 시 동기 호출 (iOS gesture 유지)
const play = useCallback(() => {
  if (!chunks.current?.length) return;
  speechSynthesis.cancel();  // 이전 잔여 utterance 제거
  indexRef.current = 0;
  setState('playing');
  speakNext();  // sync call, no await
}, []);

// 2. chunk 큐 순차 처리
const speakNext = () => {
  const idx = indexRef.current;
  const chunk = chunks.current?.[idx];
  if (!chunk) {
    setState('idle');
    onComplete?.();
    return;
  }
  const u = new SpeechSynthesisUtterance(chunk.text);
  u.voice = voice;
  u.lang = voice?.lang ?? 'en-US';
  u.rate = Math.min(Math.max(rate, 0.5), 2);
  u.onstart = () => {
    setCurrentChunkIndex(idx);
    onChunkChange?.(idx);
  };
  u.onend = () => {
    indexRef.current++;
    speakNext();
  };
  u.onerror = (e) => {
    // Chrome emits 'interrupted' on cancel — not a real error
    if (e.error !== 'interrupted' && e.error !== 'canceled') {
      onError?.(e);
    }
  };
  speechSynthesis.speak(u);
};

// 3. pause/resume — 원자성 보장
const pause = () => {
  speechSynthesis.pause();
  setState('paused');
};
const resume = () => {
  speechSynthesis.resume();
  setState('playing');
};

// 4. stop — queue 플러시
const stop = () => {
  speechSynthesis.cancel();
  indexRef.current = 0;
  setCurrentChunkIndex(0);
  setState('idle');
};

// 5. unmount cleanup
useEffect(() => {
  return () => speechSynthesis.cancel();
}, []);
```

**엣지 케이스**:
- **rate 변경 중 재생**: 현재 재생 중 utterance는 기존 rate 유지. 다음 chunk부터 새 rate 적용. 즉시 반영 원하면: `cancel()` + 현재 `indexRef` 유지 + `speakNext()` 재호출.
- **voice 변경 중 재생**: 동일. 단, 다음 chunk 시점에서 중단감 느껴질 수 있음 → UI에 "Voice will apply from next paragraph" hint.
- **탭 backgrounding**: Chrome은 백그라운드 탭에서 `speechSynthesis` 일시정지함. `visibilitychange` listener 로 UI 상태 sync.

### 3.2 `useScrollSync()` — 문단 추적 + user-scroll 존중

**파일**: `src/hooks/use-scroll-sync.ts`

```ts
interface UseScrollSyncOptions {
  chunks: React.RefObject<TTSChunk[]>;
  currentChunkIndex: number;
  enabled: boolean;  // false면 hook 아무 것도 안 함
}

function useScrollSync(options: UseScrollSyncOptions): void;
```

**구현** (리서치 Agent A 제시 패턴 채택):

```ts
const prevBlockIdxRef = useRef<number | null>(null);
const prevElRef = useRef<HTMLElement | null>(null);
const programmaticRef = useRef(false);
const userScrolledAtRef = useRef(0);
const SUSPEND_MS = 2500;

// User scroll 감지
useEffect(() => {
  const onScroll = () => {
    if (programmaticRef.current) return;
    userScrolledAtRef.current = Date.now();
  };
  window.addEventListener('scroll', onScroll, { passive: true });
  return () => window.removeEventListener('scroll', onScroll);
}, []);

// Chunk 변경에 반응
useEffect(() => {
  if (!enabled) return;
  const chunk = chunks.current?.[currentChunkIndex];
  if (!chunk) return;

  // Highlight 항상 swap
  if (prevElRef.current && prevElRef.current !== chunk.blockEl) {
    prevElRef.current.removeAttribute('data-tts-active');
  }
  chunk.blockEl.setAttribute('data-tts-active', 'true');
  prevElRef.current = chunk.blockEl;

  // 같은 문단 내 chunk 이동이면 scroll skip
  if (prevBlockIdxRef.current === chunk.blockIndex) return;
  prevBlockIdxRef.current = chunk.blockIndex;

  // User scroll 쿨다운
  if (Date.now() - userScrolledAtRef.current < SUSPEND_MS) return;

  const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  programmaticRef.current = true;
  chunk.blockEl.scrollIntoView({
    behavior: reduced ? 'auto' : 'smooth',
    block: 'center',
  });
  const t = setTimeout(() => {
    programmaticRef.current = false;
  }, 800);
  return () => clearTimeout(t);
}, [chunks, currentChunkIndex, enabled]);

// Disabled / unmount cleanup
useEffect(() => {
  if (!enabled && prevElRef.current) {
    prevElRef.current.removeAttribute('data-tts-active');
    prevElRef.current = null;
    prevBlockIdxRef.current = null;
  }
}, [enabled]);
```

---

## 4. 순수 유틸

### 4.1 `src/lib/tts/chunks.ts`

```ts
export interface TTSChunk {
  text: string;
  blockIndex: number;
  blockEl: HTMLElement;
}

/**
 * Split long text into sentence-aligned chunks ≤ maxLen chars.
 * Works around the Chrome 15-second utterance cutoff by ensuring
 * no single utterance exceeds the chunk length.
 */
export function chunkText(text: string, maxLen = 160): string[] {
  const re = new RegExp(
    `[\\s\\S]{1,${maxLen}}[.!?,](?=\\s|$)|[\\s\\S]{1,${maxLen}}`,
    'g',
  );
  return text.match(re) ?? [text];
}

const BLOCK_SELECTOR = 'p, h2, h3, h4, h5, li, blockquote';

/**
 * Extract speakable chunks from the rendered article DOM.
 * Skips code blocks (descendants of <pre>), autolink anchor text
 * (rehypeAutolinkHeadings adds `.heading-anchor`), and empty blocks.
 */
export function extractChunks(articleEl: HTMLElement): TTSChunk[] {
  const blocks = Array.from(articleEl.querySelectorAll<HTMLElement>(BLOCK_SELECTOR))
    .filter((el) => !el.closest('pre'))       // skip inside block code
    .filter((el) => !isEmpty(el));

  const chunks: TTSChunk[] = [];
  blocks.forEach((el, blockIndex) => {
    const text = readableText(el);
    if (!text) return;
    chunkText(text, 160).forEach((t) => {
      chunks.push({ text: t, blockIndex, blockEl: el });
    });
  });
  return chunks;
}

/** Strip autolink anchor "#" text and normalize whitespace. */
function readableText(el: HTMLElement): string {
  const clone = el.cloneNode(true) as HTMLElement;
  clone.querySelectorAll('.heading-anchor, [aria-hidden="true"]').forEach((n) => n.remove());
  return (clone.textContent ?? '').replace(/\s+/g, ' ').trim();
}

function isEmpty(el: HTMLElement): boolean {
  return !(el.textContent ?? '').trim();
}
```

**커버리지 검증** (코드베이스 Agent B 결과 기반):
- ✅ `<p>` (mdx-components.tsx:49)
- ✅ `<h2>`, `<h3>`, `<h4>` (mdx-components.tsx:28-47)
- ✅ `<li>` (mdx-components.tsx:85-88)
- ✅ `<blockquote>` (mdx-components.tsx:90-96)
- ⚠️ `<Callout>` 내부 텍스트: 렌더되면 `<aside>` 블록. 현재 selector에 없음. **후속 결정**: MVP는 Callout 무시, 필요시 Phase B에서 추가.
- ⚠️ `<td>` (테이블 셀): 같은 이유로 MVP 제외.
- ⚠️ `<GifEmbed>` 의 `<figcaption>`: 마찬가지로 MVP 제외.

이 결정은 "핵심 본문 prose는 읽되, 보조 UI 요소는 스킵" 원칙과 일치.

### 4.2 `src/lib/tts/prefs.ts`

단일 JSON object 방식 (리서치 Agent A 권장):

```ts
const STORAGE_KEY = 'tene:blog-tts';
const VERSION = 1;

export interface TTSPrefs {
  v: number;            // schema version
  voiceURI: string | null;
  rate: number;         // clamped [0.5, 2]
}

const DEFAULTS: TTSPrefs = { v: VERSION, voiceURI: null, rate: 1 };

export function loadPrefs(): TTSPrefs {
  if (typeof window === 'undefined') return DEFAULTS;
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return DEFAULTS;
    const parsed = JSON.parse(raw);
    if (parsed?.v !== VERSION) return DEFAULTS; // future-proof schema migration
    return {
      v: VERSION,
      voiceURI: typeof parsed.voiceURI === 'string' ? parsed.voiceURI : null,
      rate: clampRate(parsed.rate),
    };
  } catch {
    return DEFAULTS;
  }
}

let memoryFallback: TTSPrefs | null = null;

export function savePrefs(patch: Partial<Omit<TTSPrefs, 'v'>>): void {
  if (typeof window === 'undefined') return;
  const current = memoryFallback ?? loadPrefs();
  const next: TTSPrefs = {
    ...current,
    ...patch,
    v: VERSION,
    rate: patch.rate != null ? clampRate(patch.rate) : current.rate,
  };
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    memoryFallback = null;
  } catch {
    // iOS Safari private mode → in-memory fallback
    memoryFallback = next;
  }
}

function clampRate(r: unknown): number {
  const n = typeof r === 'number' ? r : parseFloat(String(r));
  if (!Number.isFinite(n)) return 1;
  return Math.min(Math.max(n, 0.5), 2);
}
```

**선택 근거**:
- 원자성 (단일 `setItem` 호출)
- 버전 필드로 후속 schema 변경 안전
- Private mode fallback 명시
- SSR safe

### 4.3 `src/lib/tts/voices.ts`

```ts
import { useEffect, useState } from 'react';

/**
 * Chrome returns [] on first getVoices() call; voices populate
 * asynchronously and fire 'voiceschanged'. Safari returns them
 * synchronously with no event. This hook handles both.
 */
export function useVoices(): SpeechSynthesisVoice[] {
  const [voices, setVoices] = useState<SpeechSynthesisVoice[]>([]);

  useEffect(() => {
    if (typeof window === 'undefined' || !('speechSynthesis' in window)) return;
    const update = () => setVoices(window.speechSynthesis.getVoices());
    update();
    window.speechSynthesis.addEventListener('voiceschanged', update);
    return () => window.speechSynthesis.removeEventListener('voiceschanged', update);
  }, []);

  return voices;
}

/**
 * Voices filtered to the given language prefix (e.g. 'en' matches en-US, en-GB, ...).
 */
export function filterByLangPrefix(
  voices: SpeechSynthesisVoice[],
  prefix: string,
): SpeechSynthesisVoice[] {
  return voices.filter((v) => v.lang.toLowerCase().startsWith(prefix.toLowerCase()));
}

/**
 * Select the best voice given saved URI + available list.
 * Fallback order: saved URI → default en-* voice → first en-* → first voice → null.
 */
export function pickBestVoice(
  all: SpeechSynthesisVoice[],
  savedURI: string | null,
  langPrefix = 'en',
): SpeechSynthesisVoice | null {
  if (!all.length) return null;
  const matches = filterByLangPrefix(all, langPrefix);
  return (
    matches.find((v) => v.voiceURI === savedURI) ??
    matches.find((v) => v.default) ??
    matches[0] ??
    all[0] ??
    null
  );
}
```

---

## 5. 상태 머신

```
        ┌───────┐
        │ idle  │◀─────────────┐
        └───┬───┘              │
            │ play()           │ stop() or complete
            ▼                  │
        ┌───────┐              │
        │playing│──────────────┤
        └──┬─▲──┘              │
           │ │                 │
    pause()│ │resume()         │
           ▼ │                 │
        ┌───────┐              │
        │paused │──────────────┘
        └───────┘
```

**트랜지션 & 부작용 매트릭스**:

| 트리거 | from | to | 부작용 |
|---|---|---|---|
| play() | idle | playing | cancel() → indexRef=0 → speakNext() |
| pause() | playing | paused | speechSynthesis.pause() |
| resume() | paused | playing | speechSynthesis.resume() |
| stop() | any | idle | cancel() → highlight 해제 → currentChunkIndex=0 |
| complete (last chunk onend) | playing | idle | onComplete callback → 이벤트 발사 → highlight 해제 |
| rate change | playing | playing | (next chunk 부터 반영) |
| voice change | playing | playing | (next chunk 부터 반영) |
| unmount | any | N/A | cancel() + highlight 해제 |
| visibilitychange hidden | playing | paused | speechSynthesis.pause() |
| visibilitychange visible | paused(by hidden) | playing | speechSynthesis.resume() |

---

## 6. CSS 변경 (globals.css)

```css
/* ─────────────────────────────────────────────
   TTS — paragraph auto-scroll + active highlight
   ───────────────────────────────────────────── */

/* scroll-margin for scrollIntoView({block:'start'}) fallback.
   Not needed for block:'center' but cheap defense. */
article p,
article h2,
article h3,
article h4,
article h5,
article li,
article blockquote {
  scroll-margin-top: 80px;
}

/* Active paragraph highlight — specificity (0,1,1) */
article [data-tts-active="true"] {
  background: color-mix(in oklab, var(--accent) 8%, transparent);
  border-left: 2px solid var(--accent);
  padding-left: 0.75rem;
  margin-left: -0.875rem;  /* 0.75rem padding + 2px border offset */
  transition:
    background 200ms ease,
    border-color 200ms ease,
    padding-left 200ms ease;
}

@media (prefers-reduced-motion: reduce) {
  article [data-tts-active="true"] {
    transition: none;
  }
}
```

**배치**: `globals.css` 파일 끝에 `/* TTS block */` 섹션 추가.

---

## 7. Analytics events 설계

`src/lib/track.ts` `EventMap` 에 추가:

```ts
blog_tts_play: {
  slug: string;
  readingMinutes: number;
  voice?: string;          // voiceURI
  rate?: number;
};
blog_tts_pause: {
  slug: string;
  percentRead: number;     // Math.round(currentChunkIndex / total * 100)
};
blog_tts_resume: {
  slug: string;
  percentRead: number;
};
blog_tts_stop: {
  slug: string;
  percentRead: number;
};
blog_tts_complete: {
  slug: string;
  readingMinutes: number;
};
blog_tts_rate_change: {
  slug: string;
  rate: number;
};
blog_tts_voice_change: {
  slug: string;
  voiceName: string;
};
```

**발사 지점**:
- `play()` → `blog_tts_play` (prefs + readingMinutes 함께)
- `pause()` → `blog_tts_pause`
- `resume()` → `blog_tts_resume`
- `stop()` → `blog_tts_stop`
- `onComplete` → `blog_tts_complete`
- Speed select `onChange` → `blog_tts_rate_change`
- Voice select `onChange` → `blog_tts_voice_change`

---

## 8. 엣지 케이스 & 대응

| # | 케이스 | 대응 |
|:-:|---|---|
| E1 | Chrome 첫 `getVoices()` 가 빈 배열 반환 | `useVoices()` hook 이 `voiceschanged` 이벤트 구독. 초기 빈 상태에서도 Play 버튼은 활성 — voice=null이면 browser default 사용 |
| E2 | iOS Safari 제스처 끊김 | `play()` 내부에서 **동기 호출 체인** 유지 (async/await/setTimeout 금지) |
| E3 | Chrome 15초 utterance 끊김 | 160자 chunk로 회피 (기존 `chunkText`) |
| E4 | localStorage private mode 차단 | `try/catch` + in-memory fallback (`memoryFallback` 변수) |
| E5 | 저장된 voiceURI 가 OS 업데이트 후 사라짐 | `pickBestVoice()` 의 fallback 체인 |
| E6 | rate > 1.5 에서 scroll animation 어색 | Plan Phase B 고려. MVP는 smooth 유지 (prefers-reduced-motion 에서는 auto) |
| E7 | 사용자가 재생 중 다음 블로그로 네비게이션 | `useEffect(() => () => speechSynthesis.cancel(), [])` |
| E8 | 브라우저 탭 background | `visibilitychange` 리스너: hidden → pause, visible → resume (단, 사용자가 직접 paused 상태였다면 그대로 유지) |
| E9 | Chrome `cancel()` 후 `onerror` event 발사 | `e.error === 'interrupted'` or `'canceled'` 는 무시 (정상 경로) |
| E10 | `scrollIntoView` 가 user scroll 방해 | 2.5s cooldown (`userScrolledAtRef`) |
| E11 | 같은 문단 내 여러 chunk 연속 재생 | `prevBlockIdxRef` 비교로 scroll 1회만 |
| E12 | 코드블록이 `<p>` 안에 인라인으로 (MDX 오작성) | `el.closest('pre')` 로 방어. 인라인 `<code>` 는 textContent에 포함되어 읽힘 — 의도됨 |
| E13 | 헤딩 내부 autolink `#` 문자 | `readableText()` 에서 `.heading-anchor` element 제거 |
| E14 | 매우 긴 문단 (1000자+) | chunkText가 160자씩 분할. blockIndex 같으므로 scroll 1회. UX 영향 없음 |
| E15 | 글 본문이 비어있는 경우 (draft?) | `chunks.length === 0` 이면 Play 버튼 disable + aria-disabled |
| E16 | 키보드로 Play 버튼 focus 상태에서 Space | `<button>` 기본 동작. 별도 단축키 X (MVP 제외) |
| E17 | Preferred voice 가 처음 없지만 나중에 로드 | `useVoices()` + `pickBestVoice()` 가 업데이트 시 재계산. selection이 null→value 로 바뀌면 UI 자동 반영 |

---

## 9. QA 검증 매트릭스

### 9.1 자동 (Chrome MCP)

| 항목 | 기대 | 방법 |
|---|---|---|
| /blog/{slug} 에서 "Listen" 버튼 렌더 | 1개 존재 | `document.querySelector('button[aria-label="Listen to article"]')` |
| 버튼 클릭 후 controls 노출 | play/pause/speed/voice 모두 존재 | DOM query |
| 클릭 후 `speechSynthesis.speaking === true` | true | 300ms wait 후 체크 |
| 현재 문단에 `data-tts-active="true"` 부여됨 | 1개 element | `querySelectorAll('[data-tts-active]').length === 1` |
| Pause 클릭 → `speechSynthesis.paused === true` | true | direct API 확인 |
| Voice select → localStorage에 voiceURI 저장됨 | `localStorage['tene:blog-tts']` 에 voiceURI | |
| 다른 글 navigate 시 이전 음성 종료 | `speechSynthesis.speaking === false` | route change 후 확인 |
| Stop 후 highlight 해제됨 | `[data-tts-active]` 0개 | |

### 9.2 수동 (실기기)

| 브라우저 | OS | 항목 |
|---|---|---|
| Safari 17+ | macOS | 재생 + scroll follow + voice dropdown |
| Chrome 130+ | macOS | 동일 + 15초 chunking 검증 (긴 문단) |
| Firefox 130+ | macOS | 재생 + scroll follow (boundary 없음이지만 chunk 기반 sync 정상) |
| Safari | iOS 17+ | 첫 play 제스처 + 백그라운드 시 pause |
| Chrome | Android | 동일 |

### 9.3 접근성

- [ ] `<button>` aria-label
- [ ] Speed/Voice select 에 `<label>`
- [ ] 활성 문단에 focus 이동 X (screen reader 중복 방지)
- [ ] `prefers-reduced-motion` 존중

---

## 10. 빌드/번들 검증

### 10.1 사이즈 추정 (Agent B 분석)

```
TTSButton              0.5 KB
TTSPlayer + hooks     ~7.5 KB
chunks.ts              1.2 KB
prefs.ts               1.0 KB
voices.ts              0.8 KB
CSS (globals)          0.3 KB
──────────────────────────────
Subtotal (uncompressed) ~11 KB
Gzipped (~35%)          ~4 KB  ✅ Plan 목표 < 5 KB
```

### 10.2 Next.js 호환

- Client component만 client bundle에 편입 (server PostHero 영향 없음)
- SSR 에서 `typeof window === 'undefined'` 가드 전부 적용 (prefs.ts, voices.ts, hooks)
- Code splitting: `TTSButton` → `TTSPlayer` lazy import 옵션 고려 (MVP 아님)

### 10.3 CSP 호환

Agent B 확인: `next.config.ts` CSP는 Web Speech API 미영향 (로컬 API, 외부 네트워크 요청 없음).

---

## 11. Phase B (MVP 외)

- 단어 레벨 하이라이트 (`boundary` 이벤트, Safari/Chrome만)
- 재생 위치 persist (slug + chunkIndex → localStorage)
- 한국어 글 추가 시 다국어 voice auto-select
- 모바일 floating mini-player
- Callout/CopyCommand/Table 읽기 지원
- 키보드 단축키 (Space, Esc, →/←)
- 사용자 개입 감지 고도화 (wheel/touch 이벤트 분리)
- 재생 속도 ≥ 1.75x 시 scroll animation `auto` 강제

---

## 12. 구현 순서 (Do Phase)

1. `src/lib/track.ts` — EventMap 확장 (7개 이벤트 추가)
2. `src/lib/tts/chunks.ts` — 순수 함수 + 테스트 가능한 형태
3. `src/lib/tts/prefs.ts` — localStorage wrapper + SSR guard
4. `src/lib/tts/voices.ts` — voices hook + filter/pick helpers
5. `src/hooks/use-tts.ts` — 재생 lifecycle 훅
6. `src/hooks/use-scroll-sync.ts` — scroll + highlight 훅
7. `src/components/blog/tts-controls.tsx` — 순수 UI
8. `src/components/blog/tts-player.tsx` — orchestrator
9. `src/components/blog/tts-button.tsx` — 진입점
10. `src/components/blog/post-hero.tsx` — TTSButton 통합 (1줄 추가)
11. `src/app/globals.css` — `[data-tts-active]` 스타일
12. Build + Chrome MCP QA
13. 수동 cross-browser 확인 (macOS 3 브라우저 + iOS Safari)
14. CHANGELOG + Report + PR staging

---

## 13. 참조

- Plan: `docs/01-plan/blog-tts.plan.md`
- /pdca team 리서치 2건 (이 세션):
  - Agent A: scroll sync + localStorage 기술 상세
  - Agent B: tene 코드베이스 통합 지점 + DOM 구조 매핑
- 기존 파일:
  - `src/app/blog/[slug]/page.tsx:129` — `<article className="min-w-0">` 컨테이너
  - `src/components/blog/post-hero.tsx` — 서버 컴포넌트, 클라이언트 자식 import 지점
  - `src/mdx-components.tsx:20-164` — 모든 MDX 요소 렌더러
  - `src/components/blog/code-block-wrapper.tsx` — shiki 블록 구분용 `.shiki` class
  - `src/lib/track.ts:6-55` — EventMap 확장 위치
  - `src/app/globals.css:1-140` — CSS 변수 + 애니메이션 위치
- ttsreader 엔진 참조: `github.com/RonenR/ttsreader` (패턴만, import 안 함)
- Web Speech API: https://developer.mozilla.org/en-US/docs/Web/API/Web_Speech_API
- `scroll-margin-top`: https://developer.mozilla.org/en-US/docs/Web/CSS/scroll-margin-top
- `prefers-reduced-motion`: https://developer.mozilla.org/en-US/docs/Web/CSS/@media/prefers-reduced-motion

---

**작성**: 2026-04-24 (Claude main, /pdca team 종합)
**승인 대기**: 사용자 확인 후 Do Phase 진입. 수정 필요한 결정이 있으면 이 문서의 해당 section에 기록하고 Do 재개.
