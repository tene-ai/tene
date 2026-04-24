// Design Ref: docs/02-design/features/blog-tts.design.md §3.2
// Auto-scroll the article so the currently-spoken paragraph stays in
// view, toggle a visual highlight, and respect the user if they scroll
// away manually (2.5 s cooldown before auto-scroll resumes).
"use client";

import { useEffect, useRef } from "react";
import type { TTSChunk } from "./chunks";

export interface UseScrollSyncOptions {
  chunksRef: React.RefObject<TTSChunk[] | null>;
  currentChunkIndex: number;
  enabled: boolean;
}

const USER_SCROLL_SUSPEND_MS = 2500;
const PROGRAMMATIC_SCROLL_SETTLE_MS = 800;

export function useScrollSync(options: UseScrollSyncOptions): void {
  const { chunksRef, currentChunkIndex, enabled } = options;

  const prevBlockIdxRef = useRef<number | null>(null);
  const prevElRef = useRef<HTMLElement | null>(null);
  const programmaticRef = useRef(false);
  const userScrolledAtRef = useRef(0);

  // Track user scrolls so auto-scroll doesn't fight the reader.
  useEffect(() => {
    if (typeof window === "undefined") return;
    const onScroll = () => {
      if (programmaticRef.current) return;
      userScrolledAtRef.current = Date.now();
    };
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => {
      window.removeEventListener("scroll", onScroll);
    };
  }, []);

  // React to chunk changes: swap highlight class, scroll on paragraph
  // transitions only (not on sentence transitions inside the same block).
  useEffect(() => {
    if (!enabled) return;
    const chunks = chunksRef.current ?? [];
    const chunk = chunks[currentChunkIndex];
    if (!chunk) return;

    // Swap highlight via inline styles. Inline + !important is the most
    // reliable path on top of Tailwind v4's base layer; we also keep
    // the data attribute + class for QA querySelector convenience.
    if (prevElRef.current && prevElRef.current !== chunk.blockEl) {
      stripHighlight(prevElRef.current);
    }
    chunk.blockEl.setAttribute("data-tts-active", "true");
    chunk.blockEl.classList.add("tts-active");
    // No transition — when we setProperty the destination values, some
    // browsers were interpolating from the default state forever rather
    // than settling. Applying values directly keeps the highlight crisp.
    chunk.blockEl.style.setProperty(
      "background-color",
      "rgba(0, 255, 136, 0.12)",
      "important",
    );
    chunk.blockEl.style.setProperty("border-left-width", "3px", "important");
    chunk.blockEl.style.setProperty("border-left-style", "solid", "important");
    chunk.blockEl.style.setProperty(
      "border-left-color",
      "#00ff88",
      "important",
    );
    chunk.blockEl.style.setProperty("padding-left", "0.75rem", "important");
    chunk.blockEl.style.setProperty("margin-left", "-0.875rem", "important");
    chunk.blockEl.style.setProperty(
      "border-radius",
      "0 2px 2px 0",
      "important",
    );
    prevElRef.current = chunk.blockEl;

    // Scroll only when we cross a block boundary.
    if (prevBlockIdxRef.current === chunk.blockIndex) return;
    prevBlockIdxRef.current = chunk.blockIndex;

    // Respect user scroll: if they interacted recently, skip the scroll
    // for this block (highlight still updates so they can catch up).
    if (Date.now() - userScrolledAtRef.current < USER_SCROLL_SUSPEND_MS) {
      return;
    }

    const reduced =
      typeof window !== "undefined" &&
      window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    programmaticRef.current = true;
    chunk.blockEl.scrollIntoView({
      behavior: reduced ? "auto" : "smooth",
      block: "center",
    });
    const t = window.setTimeout(() => {
      programmaticRef.current = false;
    }, PROGRAMMATIC_SCROLL_SETTLE_MS);
    return () => {
      window.clearTimeout(t);
    };
  }, [chunksRef, currentChunkIndex, enabled]);

  // When disabled (stop/unmount), strip the highlight so the paragraph
  // returns to its normal style.
  useEffect(() => {
    if (!enabled && prevElRef.current) {
      stripHighlight(prevElRef.current);
      prevElRef.current = null;
      prevBlockIdxRef.current = null;
    }
  }, [enabled]);

  // Final cleanup on unmount.
  useEffect(() => {
    return () => {
      if (prevElRef.current) {
        stripHighlight(prevElRef.current);
        prevElRef.current = null;
      }
    };
  }, []);
}

/** Remove the TTS highlight data/class/inline-style from an element. */
function stripHighlight(el: HTMLElement): void {
  el.removeAttribute("data-tts-active");
  el.classList.remove("tts-active");
  el.style.removeProperty("background-color");
  el.style.removeProperty("border-left-width");
  el.style.removeProperty("border-left-style");
  el.style.removeProperty("border-left-color");
  el.style.removeProperty("padding-left");
  el.style.removeProperty("margin-left");
  el.style.removeProperty("border-radius");
}
