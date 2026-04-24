// Design Ref: docs/02-design/features/blog-tts.design.md §4.3
// Voice enumeration + selection. Chrome returns [] on first getVoices()
// call and populates asynchronously (fires 'voiceschanged'); Safari
// returns them synchronously. This hook subscribes to both paths.
"use client";

import { useEffect, useState } from "react";

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
 * Pick the best voice given the saved voiceURI + the live voice list.
 * Fallback chain:
 *   1. Saved URI exists in current list                 → use it
 *   2. A default voice in the target language prefix    → use it
 *   3. First voice matching the language prefix         → use it
 *   4. First voice overall                              → use it
 *   5. null (let the browser pick its own default)
 */
export function pickBestVoice(
  all: SpeechSynthesisVoice[],
  savedURI: string | null,
  langPrefix = "en",
): SpeechSynthesisVoice | null {
  if (!all.length) return null;
  const matches = filterByLangPrefix(all, langPrefix);
  return (
    (savedURI && matches.find((v) => v.voiceURI === savedURI)) ||
    matches.find((v) => v.default) ||
    matches[0] ||
    all[0] ||
    null
  );
}
