"use client";

import { useState, useEffect, useRef, useCallback } from "react";

type Line =
  | { type: "command"; text: string }
  | { type: "output"; text: string; style?: "green" | "accent" | "dim" }
  | { type: "blank" };

const SCRIPT: Line[] = [
  { type: "command", text: "go install github.com/tomo-kay/tene/cmd/tene@latest" },
  { type: "blank" },
  { type: "command", text: "tene version" },
  { type: "output", text: "  tene v0.1.0 (darwin/arm64)", style: "green" },
  { type: "blank" },
  { type: "command", text: "tene init" },
  { type: "output", text: "  Master Password: ********", style: "dim" },
  { type: "output", text: "  ✓ .tene/vault.db created", style: "green" },
  { type: "output", text: "  ✓ CLAUDE.md created — Claude Code will auto-detect tene", style: "green" },
  { type: "output", text: "  ✓ .tene/ added to .gitignore", style: "green" },
  { type: "blank" },
  { type: "output", text: "  Recovery Key:", style: "dim" },
  { type: "output", text: "  apple banana cherry dolphin eagle frost", style: "accent" },
  { type: "output", text: "  grape harbor island jungle kite lemon", style: "accent" },
  { type: "blank" },
  { type: "command", text: "tene set STRIPE_KEY sk_test_51Hxxxxx" },
  { type: "output", text: "  ✓ STRIPE_KEY saved (encrypted)", style: "green" },
  { type: "blank" },
  { type: "command", text: "tene set OPENAI_API_KEY sk-proj-xxxxx" },
  { type: "output", text: "  ✓ OPENAI_API_KEY saved (encrypted)", style: "green" },
  { type: "blank" },
  { type: "command", text: "tene run -- claude" },
  { type: "output", text: "  ✓ 2 secrets injected as environment variables", style: "green" },
  { type: "output", text: "  ✓ Starting: claude", style: "green" },
  { type: "blank" },
  { type: "output", text: '  // Claude Code reads CLAUDE.md and knows:', style: "dim" },
  { type: "output", text: '  // "This project uses tene for secret management."', style: "dim" },
  { type: "output", text: '  // "Use tene get <KEY> to retrieve secrets."', style: "dim" },
];

const TYPING_SPEED = 30;
const OUTPUT_DELAY = 150;
const COMMAND_PAUSE = 400;

export function Terminal() {
  const [visibleLines, setVisibleLines] = useState<
    { text: string; style?: string; typing?: boolean }[]
  >([]);
  const [started, setStarted] = useState(false);
  const sectionRef = useRef<HTMLElement>(null);
  const hasRun = useRef(false);

  const runAnimation = useCallback(async () => {
    if (hasRun.current) return;
    hasRun.current = true;

    const delay = (ms: number) => new Promise((r) => setTimeout(r, ms));

    for (const line of SCRIPT) {
      if (line.type === "blank") {
        setVisibleLines((prev) => [...prev, { text: "", style: "blank" }]);
        await delay(100);
      } else if (line.type === "command") {
        // Type command character by character
        const prefix = "$ ";
        for (let i = 0; i <= line.text.length; i++) {
          const partial = prefix + line.text.slice(0, i);
          setVisibleLines((prev) => {
            const next = [...prev];
            // Replace last typing line or add new
            if (next.length > 0 && next[next.length - 1].typing) {
              next[next.length - 1] = { text: partial, style: "command", typing: true };
            } else {
              next.push({ text: partial, style: "command", typing: true });
            }
            return next;
          });
          await delay(TYPING_SPEED);
        }
        // Finalize command (remove cursor)
        setVisibleLines((prev) => {
          const next = [...prev];
          if (next.length > 0) {
            next[next.length - 1] = {
              text: prefix + line.text,
              style: "command",
              typing: false,
            };
          }
          return next;
        });
        await delay(COMMAND_PAUSE);
      } else {
        // Output line — appears instantly
        setVisibleLines((prev) => [
          ...prev,
          { text: line.text, style: line.style || "" },
        ]);
        await delay(OUTPUT_DELAY);
      }
    }
  }, []);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !started) {
          setStarted(true);
          runAnimation();
        }
      },
      { threshold: 0.3 }
    );

    const el = sectionRef.current;
    if (el) observer.observe(el);
    return () => {
      if (el) observer.unobserve(el);
    };
  }, [started, runAnimation]);

  return (
    <section ref={sectionRef} className="px-4 py-20 sm:px-6">
      <div className="mx-auto max-w-3xl">
        <div className="overflow-hidden rounded-xl border border-border bg-surface">
          <div className="flex items-center gap-2 border-b border-border px-4 py-3">
            <div className="h-3 w-3 rounded-full bg-[#ff5f57]" />
            <div className="h-3 w-3 rounded-full bg-[#febc2e]" />
            <div className="h-3 w-3 rounded-full bg-[#28c840]" />
            <span className="ml-3 text-xs text-muted font-mono">~/my-project</span>
          </div>
          <div className="overflow-x-auto p-4 font-mono text-xs leading-7 sm:p-6 sm:text-sm min-h-[420px] sm:min-h-[480px]">
            {visibleLines.map((line, i) => {
              if (line.style === "blank") return <div key={i} className="h-4" />;

              const colorClass =
                line.style === "green"
                  ? "text-[#28c840]"
                  : line.style === "accent"
                    ? "text-accent font-semibold"
                    : line.style === "dim"
                      ? "text-muted"
                      : "";

              return (
                <div key={i} className={`whitespace-nowrap ${colorClass}`}>
                  {line.text}
                  {line.typing && (
                    <span className="inline-block w-2 h-4 bg-accent ml-0.5 animate-pulse align-middle" />
                  )}
                </div>
              );
            })}
          </div>
        </div>

        <p className="mt-6 text-center text-sm text-muted">
          From install to first secret injection — under 1 minute. No account needed.
        </p>
      </div>
    </section>
  );
}
