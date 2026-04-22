"use client";

// Design Ref: §2.8 FR-35 — Visible RSS link with tracked click (Iterate cycle).
import { track } from "@/lib/track";

type Props = {
  location: "footer" | "blog_header";
};

export function RssLink({ location }: Props) {
  return (
    <a
      href="/blog/rss.xml"
      onClick={() => track("blog_rss_click", { location })}
      className="inline-flex items-center gap-2 rounded-full border border-border bg-surface/60 px-4 py-1.5 text-sm text-muted backdrop-blur-sm transition-colors hover:border-accent/40 hover:text-foreground"
      aria-label="RSS feed"
    >
      <svg
        className="h-4 w-4 text-accent"
        fill="currentColor"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <path d="M6.18 15.64a2.18 2.18 0 0 1 2.18 2.18C8.36 19 7.38 20 6.18 20A2.18 2.18 0 0 1 4 17.82a2.18 2.18 0 0 1 2.18-2.18M4 4.44A15.56 15.56 0 0 1 19.56 20h-2.83A12.73 12.73 0 0 0 4 7.27zm0 5.66a9.9 9.9 0 0 1 9.9 9.9h-2.83A7.07 7.07 0 0 0 4 12.93z" />
      </svg>
      RSS feed
    </a>
  );
}
