// Design Ref: docs/02-design/features/blog-tts.design.md §4.3
// Voice enumeration + selection. Chrome returns [] on first getVoices()
// call and populates asynchronously (fires 'voiceschanged'); Safari
// returns them synchronously. This hook subscribes to both paths.
//
// Voice curation (2026-04-24): raw `speechSynthesis.getVoices()` on
// macOS returns ~210 entries including legacy novelty voices — Bells,
// Bubbles, Zarvox, Bad News, Whisper, Cellos, etc. — that play actual
// sound effects instead of natural speech, plus compact/premium
// duplicates of the same name. We filter via a human-curated allowlist
// (NATURAL_VOICE_ALLOWLIST) to show only ~6-10 natural voices across
// 3 male / 3 female tones.
"use client";

import { useEffect, useState } from "react";

/**
 * Known-natural voice names across macOS / Windows / Chrome / Edge / Linux.
 * If a user's OS exposes none of these, we fall back to the lang-prefix
 * filter so the dropdown is never empty.
 */
/**
 * Tight allowlist targeting ~3 male + ~3 female natural voices per
 * platform. We intentionally keep this short: more options is more
 * choice paralysis, and the novelty/compact noise in the raw
 * getVoices() list is what we're trying to eliminate. If a user
 * really wants an obscure voice, they can still pick it from their
 * OS's native TTS settings and install tene elsewhere.
 */
const NATURAL_VOICE_ALLOWLIST: ReadonlySet<string> = new Set([
  // macOS female — 3 distinct accents
  "Samantha", // en-US, F (Apple flagship, decades of tuning)
  "Karen", // en-AU, F
  "Moira", // en-IE, F
  // macOS male — 3 distinct accents
  "Alex", // en-US, M (Apple flagship, if installed)
  "Daniel", // en-GB, M
  "Aaron", // en-US, M (fallback if Alex missing)
  // Chrome / Edge cloud voices — cross-platform fallbacks for users
  // whose OS exposes none of the above (Windows, Linux, ChromeOS).
  "Google US English",
  "Google UK English Female",
  "Google UK English Male",
  "Microsoft Aria Online (Natural) - English (United States)",
  "Microsoft Guy Online (Natural) - English (United States)",
]);

/**
 * macOS new-naming adds a locale suffix — "Daniel (English (UK))" or
 * "Daniel (영어(영국))" depending on the OS language — to the
 * otherwise-same voice. Strip everything from the first " (" so the
 * allowlist match still works.
 */
function voiceBaseName(name: string): string {
  const idx = name.indexOf(" (");
  return idx > 0 ? name.slice(0, idx).trim() : name.trim();
}

export function isNaturalVoice(voice: SpeechSynthesisVoice): boolean {
  return (
    NATURAL_VOICE_ALLOWLIST.has(voice.name) ||
    NATURAL_VOICE_ALLOWLIST.has(voiceBaseName(voice.name))
  );
}

export function useVoices(): SpeechSynthesisVoice[] {
  const [voices, setVoices] = useState<SpeechSynthesisVoice[]>([]);

  useEffect(() => {
    if (typeof window === "undefined" || !("speechSynthesis" in window)) {
      return;
    }
    const update = () => setVoices(window.speechSynthesis.getVoices());
    update();
    window.speechSynthesis.addEventListener("voiceschanged", update);
    return () => {
      window.speechSynthesis.removeEventListener("voiceschanged", update);
    };
  }, []);

  return voices;
}

/**
 * Filter to voices whose BCP-47 language tag starts with `prefix`.
 * e.g. filterByLangPrefix(voices, 'en') matches en-US, en-GB, en-AU.
 */
export function filterByLangPrefix(
  voices: SpeechSynthesisVoice[],
  prefix: string,
): SpeechSynthesisVoice[] {
  const p = prefix.toLowerCase();
  return voices.filter((v) => v.lang.toLowerCase().startsWith(p));
}

/**
 * The UI-facing voice list: lang-filtered, de-duplicated by voice
 * display name, and cut to the allowlist. If the filter leaves us
 * with fewer than 2 voices (exotic OS), falls back to all lang-matching
 * voices minus the well-known novelty names.
 */
const NOVELTY_BLOCKLIST: ReadonlySet<string> = new Set([
  "Albert", "Bad News", "Bahh", "Bells", "Boing", "Bubbles", "Cellos",
  "Deranged", "Good News", "Hysterical", "Jester", "Junior", "Kathy",
  "Organ", "Pipe Organ", "Princess", "Ralph", "Shelley", "Superstar",
  "Trinoids", "Whisper", "Wobble", "Zarvox", "Grandma", "Grandpa",
]);

export function getNaturalVoices(
  voices: SpeechSynthesisVoice[],
  langPrefix = "en",
): SpeechSynthesisVoice[] {
  const langMatches = filterByLangPrefix(voices, langPrefix);
  const curated = langMatches.filter(isNaturalVoice);
  if (curated.length >= 2) return dedupByBaseName(curated);
  // Fallback: return everything lang-matching minus confirmed novelty
  const safeFallback = langMatches.filter(
    (v) => !NOVELTY_BLOCKLIST.has(voiceBaseName(v.name)),
  );
  return dedupByBaseName(safeFallback);
}

/**
 * macOS exposes Compact + Premium variants of the same voice with
 * identical display names. Keep only the first occurrence so the
 * dropdown doesn't have "Samantha" three times.
 */
function dedupByBaseName(
  voices: SpeechSynthesisVoice[],
): SpeechSynthesisVoice[] {
  const seen = new Set<string>();
  const out: SpeechSynthesisVoice[] = [];
  for (const v of voices) {
    const key = voiceBaseName(v.name) + "|" + v.lang;
    if (seen.has(key)) continue;
    seen.add(key);
    out.push(v);
  }
  return out;
}

/**
 * Pick the best voice given the saved voiceURI + the live voice list.
 * Fallback chain:
 *   1. Saved URI exists in the curated natural set        → use it
 *   2. Saved URI exists in the full voice list (legacy)   → use it
 *   3. First natural voice matching the language prefix   → use it
 *   4. Default voice in the language prefix               → use it
 *   5. First voice matching the language prefix           → use it
 *   6. First voice overall                                → use it
 *   7. null (let the browser pick its own default)
 */
export function pickBestVoice(
  all: SpeechSynthesisVoice[],
  savedURI: string | null,
  langPrefix = "en",
): SpeechSynthesisVoice | null {
  if (!all.length) return null;
  const natural = getNaturalVoices(all, langPrefix);
  const langMatches = filterByLangPrefix(all, langPrefix);
  return (
    (savedURI && natural.find((v) => v.voiceURI === savedURI)) ||
    (savedURI && all.find((v) => v.voiceURI === savedURI)) ||
    natural[0] ||
    langMatches.find((v) => v.default) ||
    langMatches[0] ||
    all[0] ||
    null
  );
}
