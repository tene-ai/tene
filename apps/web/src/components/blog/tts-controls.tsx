"use client";

// Design Ref: docs/02-design/features/blog-tts.design.md §2.3
// Presentational TTS controls. Pure UI — no state, no side effects.
// Parent (TTSPlayer) owns all state and passes callbacks.
import type { TTSState } from "@/lib/tts/use-tts";

type Props = {
  state: TTSState;
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
  onClose: () => void;
};

const RATE_OPTIONS = [0.5, 0.75, 1, 1.25, 1.5, 2];

export function TTSControls({
  state,
  progress,
  rate,
  voices,
  selectedVoiceURI,
  onPlay,
  onPause,
  onResume,
  onStop,
  onRateChange,
  onVoiceChange,
  onClose,
}: Props) {
  const handleMain = () => {
    if (state === "playing") onPause();
    else if (state === "paused") onResume();
    else onPlay();
  };

  const mainLabel =
    state === "playing" ? "Pause" : state === "paused" ? "Resume" : "Play";

  const percent =
    progress.total > 0
      ? Math.min(100, Math.round((progress.current / progress.total) * 100))
      : 0;

  return (
    <div className="mt-4 rounded-lg border border-border bg-surface/60 p-3 backdrop-blur-sm">
      {/* Row 1: main control + progress */}
      <div className="flex items-center gap-3">
        <button
          type="button"
          onClick={handleMain}
          aria-label={mainLabel}
          className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-accent/40 bg-accent/10 text-accent transition-colors hover:bg-accent/20"
        >
          {state === "playing" ? (
            <PauseIcon className="h-4 w-4" />
          ) : (
            <PlayIcon className="h-4 w-4" />
          )}
        </button>

        <button
          type="button"
          onClick={onStop}
          aria-label="Stop"
          disabled={state === "idle"}
          className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-border text-muted transition-colors hover:border-accent/40 hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40"
        >
          <StopIcon className="h-4 w-4" />
        </button>

        <div className="flex-1">
          <div className="h-1.5 w-full rounded-full bg-border/60">
            <div
              className="h-full rounded-full bg-accent transition-[width] duration-200"
              style={{ width: `${percent}%` }}
            />
          </div>
          <div className="mt-1 text-xs text-muted">
            Chunk {progress.current + 1} / {progress.total} · {percent}%
          </div>
        </div>

        <button
          type="button"
          onClick={onClose}
          aria-label="Close player"
          className="rounded-md px-2 py-1 text-xs text-muted hover:text-foreground"
        >
          Close
        </button>
      </div>

      {/* Row 2: speed + voice */}
      <div className="mt-3 flex flex-wrap items-center gap-3 text-xs">
        <label className="inline-flex items-center gap-2 text-muted">
          Speed
          <select
            value={rate}
            onChange={(e) => onRateChange(parseFloat(e.target.value))}
            className="rounded-md border border-border bg-background px-2 py-1 text-foreground focus:border-accent focus:outline-none"
          >
            {RATE_OPTIONS.map((r) => (
              <option key={r} value={r}>
                {r}x
              </option>
            ))}
          </select>
        </label>

        <label className="inline-flex items-center gap-2 text-muted">
          Voice
          <select
            value={selectedVoiceURI ?? ""}
            onChange={(e) => onVoiceChange(e.target.value)}
            disabled={voices.length === 0}
            className="max-w-[220px] truncate rounded-md border border-border bg-background px-2 py-1 text-foreground focus:border-accent focus:outline-none disabled:cursor-not-allowed disabled:opacity-40"
          >
            {voices.length === 0 ? (
              <option value="">Loading…</option>
            ) : (
              voices.map((v) => (
                <option key={v.voiceURI} value={v.voiceURI}>
                  {v.name} ({v.lang})
                </option>
              ))
            )}
          </select>
        </label>
      </div>
    </div>
  );
}

function PlayIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden
    >
      <path d="M8 5v14l11-7z" />
    </svg>
  );
}

function PauseIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden
    >
      <path d="M6 5h4v14H6zM14 5h4v14h-4z" />
    </svg>
  );
}

function StopIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden
    >
      <path d="M6 6h12v12H6z" />
    </svg>
  );
}
