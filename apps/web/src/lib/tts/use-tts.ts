// Design Ref: docs/02-design/features/blog-tts.design.md §3.1
// Speech-synthesis playback lifecycle hook. Owns the utterance queue,
// state transitions (idle/playing/paused), and browser quirk
// mitigations (Chrome 15-sec chunking, cancel-emits-error filter,
// visibility-based auto-pause, iOS gesture preservation).
"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import type { TTSChunk } from "./chunks";
import { clampRate } from "./prefs";

export type TTSState = "idle" | "playing" | "paused";

export interface UseTTSOptions {
  chunksRef: React.RefObject<TTSChunk[] | null>;
  voice: SpeechSynthesisVoice | null;
  rate: number;
  onChunkChange?: (index: number) => void;
  onComplete?: () => void;
  onError?: (err: SpeechSynthesisErrorEvent) => void;
}

export interface UseTTSReturn {
  state: TTSState;
  currentChunkIndex: number;
  play: () => void;
  pause: () => void;
  resume: () => void;
  stop: () => void;
}

export function useTTS(options: UseTTSOptions): UseTTSReturn {
  const { chunksRef, voice, rate, onChunkChange, onComplete, onError } =
    options;

  const [state, setState] = useState<TTSState>("idle");
  const [currentChunkIndex, setCurrentChunkIndex] = useState(0);
  const indexRef = useRef(0);

  // Keep latest voice/rate in refs — utterance factory needs them at
  // speak time without triggering re-render.
  const voiceRef = useRef(voice);
  const rateRef = useRef(rate);
  useEffect(() => {
    voiceRef.current = voice;
  }, [voice]);
  useEffect(() => {
    rateRef.current = rate;
  }, [rate]);

  const onChunkChangeRef = useRef(onChunkChange);
  const onCompleteRef = useRef(onComplete);
  const onErrorRef = useRef(onError);
  useEffect(() => {
    onChunkChangeRef.current = onChunkChange;
    onCompleteRef.current = onComplete;
    onErrorRef.current = onError;
  });

  const isTTSAvailable = (): boolean =>
    typeof window !== "undefined" && "speechSynthesis" in window;

  const speakNext = useCallback(() => {
    if (!isTTSAvailable()) return;
    const chunks = chunksRef.current ?? [];
    const idx = indexRef.current;
    const chunk = chunks[idx];
    if (!chunk) {
      setState("idle");
      setCurrentChunkIndex(0);
      onCompleteRef.current?.();
      return;
    }
    const u = new SpeechSynthesisUtterance(chunk.text);
    u.voice = voiceRef.current;
    u.lang = voiceRef.current?.lang ?? "en-US";
    u.rate = clampRate(rateRef.current);
    u.onstart = () => {
      setCurrentChunkIndex(idx);
      onChunkChangeRef.current?.(idx);
    };
    u.onend = () => {
      indexRef.current = idx + 1;
      speakNext();
    };
    u.onerror = (e) => {
      // Chrome emits 'interrupted' / 'canceled' on normal cancel().
      // Treat as non-error; otherwise propagate.
      if (e.error !== "interrupted" && e.error !== "canceled") {
        onErrorRef.current?.(e);
      }
    };
    window.speechSynthesis.speak(u);
  }, [chunksRef]);

  const play = useCallback(() => {
    // Must be called synchronously inside a user-gesture handler on iOS
    // Safari. Don't await anything here — keep the gesture "live".
    if (!isTTSAvailable()) return;
    const chunks = chunksRef.current ?? [];
    if (chunks.length === 0) return;
    window.speechSynthesis.cancel();
    indexRef.current = 0;
    setCurrentChunkIndex(0);
    setState("playing");
    speakNext();
  }, [chunksRef, speakNext]);

  const pause = useCallback(() => {
    if (!isTTSAvailable()) return;
    window.speechSynthesis.pause();
    setState("paused");
  }, []);

  const resume = useCallback(() => {
    if (!isTTSAvailable()) return;
    window.speechSynthesis.resume();
    setState("playing");
  }, []);

  const stop = useCallback(() => {
    if (!isTTSAvailable()) return;
    window.speechSynthesis.cancel();
    indexRef.current = 0;
    setCurrentChunkIndex(0);
    setState("idle");
  }, []);

  // Unmount cleanup — always cancel so the voice doesn't bleed into
  // the next route (Next.js client navigation preserves window but
  // unmounts the component tree).
  useEffect(() => {
    return () => {
      if (isTTSAvailable()) {
        window.speechSynthesis.cancel();
      }
    };
  }, []);

  // Background tab handling — pause when the tab is hidden, resume when
  // it becomes visible again, but only if the user didn't manually pause.
  const pausedByVisibilityRef = useRef(false);
  useEffect(() => {
    if (typeof document === "undefined") return;
    const onVisibility = () => {
      if (!isTTSAvailable()) return;
      if (document.hidden && state === "playing") {
        window.speechSynthesis.pause();
        pausedByVisibilityRef.current = true;
        setState("paused");
      } else if (!document.hidden && pausedByVisibilityRef.current) {
        window.speechSynthesis.resume();
        pausedByVisibilityRef.current = false;
        setState("playing");
      }
    };
    document.addEventListener("visibilitychange", onVisibility);
    return () => {
      document.removeEventListener("visibilitychange", onVisibility);
    };
  }, [state]);

  return { state, currentChunkIndex, play, pause, resume, stop };
}
