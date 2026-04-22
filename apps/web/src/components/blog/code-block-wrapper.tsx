"use client";

// Design Ref: §2.4 T4 — Wraps shiki-rendered <pre> with hover copy button.
// Fires blog_copy_code on click (FR-32).
import { useRef, useState } from "react";
import { track } from "@/lib/track";

type Props = {
  children?: React.ReactNode;
  slug?: string;
  "data-language"?: string;
} & React.HTMLAttributes<HTMLPreElement>;

export function CodeBlockWrapper({ children, slug, ...rest }: Props) {
  const ref = useRef<HTMLPreElement>(null);
  const [copied, setCopied] = useState(false);

  async function copy() {
    const code = ref.current?.innerText ?? "";

    // Fire analytics regardless of clipboard success — intent-to-copy is the
    // engagement signal we care about, and clipboard can fail in non-secure
    // contexts or headless automation even when the user clicked.
    const lang =
      ref.current?.className.match(/language-(\w+)/)?.[1] ??
      ref.current?.querySelector("[class*=language-]")
        ?.className.match(/language-(\w+)/)?.[1] ??
      "plaintext";
    if (slug) {
      track("blog_copy_code", { slug, language: lang });
    }

    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard unavailable (e.g. non-secure context). Fail silently.
    }
  }

  return (
    <div className="group relative my-6">
      <pre
        ref={ref}
        className="overflow-x-auto rounded-lg border border-border bg-surface-2 p-4 text-sm"
        {...rest}
      >
        {children}
      </pre>
      <button
        type="button"
        onClick={copy}
        aria-label="Copy code"
        className="absolute right-3 top-3 rounded border border-border bg-surface/80 px-2 py-1 text-xs text-muted opacity-0 backdrop-blur-sm transition-opacity hover:border-accent/40 hover:text-foreground group-hover:opacity-100 focus:opacity-100"
      >
        {copied ? "✓ Copied" : "Copy"}
      </button>
    </div>
  );
}
