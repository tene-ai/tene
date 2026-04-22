"use client";

import { useState, useCallback } from "react";
import { track } from "@/lib/track";

type InstallSource = "hero" | "cta" | "vs_page" | "blog_post" | "pricing";

export function CopyCommand({
  command,
  className = "",
  source,
}: {
  command: string;
  className?: string;
  // Only instances representing the INSTALL command pass a source. Others
  // (e.g. how-it-works step copies) omit it so the install_copy conversion
  // metric is not polluted.
  source?: InstallSource;
}) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    // Fire analytics regardless of clipboard success — intent is the signal.
    if (source) {
      track("install_copy", { source });
    }
    try {
      await navigator.clipboard.writeText(command);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard may be unavailable (non-secure context, headless) — ignore.
    }
  }, [command, source]);

  return (
    <button
      onClick={handleCopy}
      className={`group relative flex items-center gap-3 rounded-lg border border-border bg-surface px-5 py-3 font-mono text-sm transition-all hover:border-accent/50 active:scale-[0.98] ${className}`}
      title="Click to copy"
    >
      <span className="text-accent">$</span>
      <code>{command}</code>
      <span className="relative ml-1 h-4 w-4 shrink-0">
        {copied ? (
          <svg
            className="h-4 w-4 text-accent"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            strokeWidth="2"
          >
            <path d="M5 13l4 4L19 7" />
          </svg>
        ) : (
          <svg
            className="h-4 w-4 text-muted transition-colors group-hover:text-accent"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <rect x="9" y="9" width="13" height="13" rx="2" ry="2" strokeWidth="2" />
            <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" strokeWidth="2" />
          </svg>
        )}
      </span>
      {copied && (
        <span className="absolute -top-8 left-1/2 -translate-x-1/2 rounded bg-accent px-2 py-1 text-xs font-medium text-background">
          Copied!
        </span>
      )}
    </button>
  );
}
