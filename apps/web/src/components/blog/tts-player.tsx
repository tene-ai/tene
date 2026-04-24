"use client";

// Design Ref: docs/02-design/features/blog-tts.design.md §2.2
// Orchestrator: glues useTTS + useScrollSync + useVoices + prefs
// together, owns the chunks ref and the analytics event emission.
import { useCallback, useEffect, useRef, useState } from "react";
import { TTSControls } from "@/components/blog/tts-controls";
import { extractChunks, type TTSChunk } from "@/lib/tts/chunks";
import { loadPrefs, savePrefs } from "@/lib/tts/prefs";
import { pickBestVoice, useVoices } from "@/lib/tts/voices";
import { useTTS } from "@/lib/tts/use-tts";
import { useScrollSync } from "@/lib/tts/use-scroll-sync";
import { track } from "@/lib/track";

type Props = {
  slug: string;
  readingMinutes: number;
  onClose: () => void;
};

const ARTICLE_SELECTOR = "article.min-w-0";

export function TTSPlayer({ slug, readingMinutes, onClose }: Props) {
  const voices = useVoices();
  const chunksRef = useRef<TTSChunk[] | null>(null);
  const [chunkCount, setChunkCount] = useState(0);

  const initialPrefs = loadPrefs();
  const [selectedVoiceURI, setSelectedVoiceURI] = useState<string | null>(
    initialPrefs.voiceURI,
  );
  const [rate, setRate] = useState<number>(initialPrefs.rate);

  // Pick best voice from live voice list + saved preference
  const voice = pickBestVoice(voices, selectedVoiceURI, "en");

  // Once voices load and we haven't committed a selection yet, sync the
  // UI dropdown to the chosen voice.
  useEffect(() => {
    if (!selectedVoiceURI && voice) {
      setSelectedVoiceURI(voice.voiceURI);
    }
    // If saved URI no longer exists, fall back + persist the new choice
    if (
      selectedVoiceURI &&
      voices.length > 0 &&
      !voices.some((v) => v.voiceURI === selectedVoiceURI) &&
      voice
    ) {
      setSelectedVoiceURI(voice.voiceURI);
      savePrefs({ voiceURI: voice.voiceURI });
    }
  }, [voices, voice, selectedVoiceURI]);

  // Extract chunks once on mount — the article DOM is available now
  // because PostHero renders before the MDX body in the same tree.
  useEffect(() => {
    const articleEl =
      typeof document !== "undefined"
        ? (document.querySelector<HTMLElement>(ARTICLE_SELECTOR) ?? null)
        : null;
    if (!articleEl) {
      chunksRef.current = [];
      setChunkCount(0);
      return;
    }
    const chunks = extractChunks(articleEl);
    chunksRef.current = chunks;
    setChunkCount(chunks.length);
  }, []);

  // Analytics helper — compute percentRead at call time.
  const percentRead = useCallback(
    (idx: number) => {
      const total = chunksRef.current?.length ?? 0;
      if (total === 0) return 0;
      return Math.min(100, Math.round((idx / total) * 100));
    },
    [],
  );

  const ttsRef = useRef<ReturnType<typeof useTTS> | null>(null);

  const tts = useTTS({
    chunksRef,
    voice,
    rate,
    onComplete: () => {
      track("blog_tts_complete", { slug, readingMinutes });
    },
    onError: () => {
      // Silent — onerror usually means the user cancelled mid-utterance.
      // Non-interrupted errors are already filtered in useTTS.
    },
  });
  ttsRef.current = tts;

  useScrollSync({
    chunksRef,
    currentChunkIndex: tts.currentChunkIndex,
    enabled: tts.state !== "idle",
  });

  const handlePlay = useCallback(() => {
    tts.play();
    track("blog_tts_play", {
      slug,
      readingMinutes,
      voice: voice?.name,
      rate,
    });
  }, [tts, slug, readingMinutes, voice, rate]);

  const handlePause = useCallback(() => {
    tts.pause();
    track("blog_tts_pause", {
      slug,
      percentRead: percentRead(tts.currentChunkIndex),
    });
  }, [tts, slug, percentRead]);

  const handleResume = useCallback(() => {
    tts.resume();
    track("blog_tts_resume", {
      slug,
      percentRead: percentRead(tts.currentChunkIndex),
    });
  }, [tts, slug, percentRead]);

  const handleStop = useCallback(() => {
    const pct = percentRead(tts.currentChunkIndex);
    tts.stop();
    track("blog_tts_stop", { slug, percentRead: pct });
  }, [tts, slug, percentRead]);

  const handleRateChange = useCallback(
    (r: number) => {
      setRate(r);
      savePrefs({ rate: r });
      track("blog_tts_rate_change", { slug, rate: r });
    },
    [slug],
  );

  const handleVoiceChange = useCallback(
    (voiceURI: string) => {
      setSelectedVoiceURI(voiceURI);
      savePrefs({ voiceURI });
      const v = voices.find((x) => x.voiceURI === voiceURI);
      if (v) {
        track("blog_tts_voice_change", { slug, voiceName: v.name });
      }
    },
    [slug, voices],
  );

  const handleClose = useCallback(() => {
    tts.stop();
    onClose();
  }, [tts, onClose]);

  return (
    <TTSControls
      state={tts.state}
      progress={{ current: tts.currentChunkIndex, total: chunkCount }}
      rate={rate}
      voices={voices}
      selectedVoiceURI={selectedVoiceURI}
      onPlay={handlePlay}
      onPause={handlePause}
      onResume={handleResume}
      onStop={handleStop}
      onRateChange={handleRateChange}
      onVoiceChange={handleVoiceChange}
      onClose={handleClose}
    />
  );
}
