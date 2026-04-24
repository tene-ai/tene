// Design Ref: docs/02-design/features/blog-tts.design.md §4.2
// Namespaced, versioned localStorage wrapper for TTS preferences
// (voice + rate). Falls back to in-memory storage when localStorage is
// unavailable (Safari private mode throws QuotaExceededError).

const STORAGE_KEY = "tene:blog-tts";
const VERSION = 1;

export interface TTSPrefs {
  v: number;
  voiceURI: string | null;
  rate: number; // clamped [0.5, 2]
}

const DEFAULTS: TTSPrefs = { v: VERSION, voiceURI: null, rate: 1 };

let memoryFallback: TTSPrefs | null = null;

export function loadPrefs(): TTSPrefs {
  if (typeof window === "undefined") return DEFAULTS;
  if (memoryFallback) return memoryFallback;
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return DEFAULTS;
    const parsed = JSON.parse(raw);
    if (parsed?.v !== VERSION) return DEFAULTS;
    return {
      v: VERSION,
      voiceURI:
        typeof parsed.voiceURI === "string" || parsed.voiceURI === null
          ? parsed.voiceURI
          : null,
      rate: clampRate(parsed.rate),
    };
  } catch {
    return DEFAULTS;
  }
}

export function savePrefs(patch: Partial<Omit<TTSPrefs, "v">>): void {
  if (typeof window === "undefined") return;
  const current = loadPrefs();
  const next: TTSPrefs = {
    v: VERSION,
    voiceURI:
      patch.voiceURI !== undefined ? patch.voiceURI : current.voiceURI,
    rate: patch.rate !== undefined ? clampRate(patch.rate) : current.rate,
  };
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    memoryFallback = null;
  } catch {
    // iOS Safari private mode / quota exceeded → in-memory fallback so
    // preferences at least persist for the current page session.
    memoryFallback = next;
  }
}

export function clampRate(r: unknown): number {
  const n = typeof r === "number" ? r : parseFloat(String(r));
  if (!Number.isFinite(n)) return 1;
  return Math.min(Math.max(n, 0.5), 2);
}
