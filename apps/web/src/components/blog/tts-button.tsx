"use client";

// Design Ref: docs/02-design/features/blog-tts.design.md §2.1
// Entry button. Stays collapsed until clicked; first click mounts the
// full TTSPlayer (which then immediately begins playback on the click's
// user gesture). Keeps the bundle cost zero for readers who don't use
// TTS and preserves the iOS Safari "gesture must be live" constraint.
import { useState } from "react";
import dynamic from "next/dynamic";

const TTSPlayer = dynamic(
  () =>
    import("@/components/blog/tts-player").then((m) => ({
      default: m.TTSPlayer,
    })),
  { ssr: false },
);

type Props = {
  slug: string;
  readingMinutes: number;
};

export function TTSButton({ slug, readingMinutes }: Props) {
  const [mounted, setMounted] = useState(false);

  if (mounted) {
    return (
      <TTSPlayer
        slug={slug}
        readingMinutes={readingMinutes}
        onClose={() => setMounted(false)}
      />
    );
  }

  return (
    <button
      type="button"
      onClick={() => setMounted(true)}
      aria-label="Listen to article"
      className="mt-4 inline-flex items-center gap-2 rounded-md border border-border bg-surface/60 px-3 py-1.5 text-sm text-muted backdrop-blur-sm transition-colors hover:border-accent/40 hover:text-foreground"
    >
      <ListenIcon className="h-4 w-4" />
      Listen
    </button>
  );
}

function ListenIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      {/* Headphones icon */}
      <path d="M3 12a9 9 0 0 1 18 0" />
      <path d="M21 19a2 2 0 0 1-2 2h-1a2 2 0 0 1-2-2v-3a2 2 0 0 1 2-2h3zM3 19a2 2 0 0 0 2 2h1a2 2 0 0 0 2-2v-3a2 2 0 0 0-2-2H3z" />
    </svg>
  );
}
