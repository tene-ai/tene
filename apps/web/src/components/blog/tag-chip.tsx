"use client";

// Design Ref: §2.4 FR-17 — Clickable tag chip. Fires blog_tag_filter on
// click (FR-33). Used in post header, cards, and index.
import Link from "next/link";
import { getTagLabel } from "@/lib/tags";
import { track } from "@/lib/track";

type Props = {
  tag: string;
  from?: "index" | "post_header" | "card";
  size?: "sm" | "md";
};

export function TagChip({ tag, from = "card", size = "sm" }: Props) {
  const label = getTagLabel(tag);
  const padding = size === "md" ? "px-3 py-1.5" : "px-2.5 py-1";
  const fontSize = size === "md" ? "text-sm" : "text-xs";

  return (
    <Link
      href={`/blog/tag/${tag}`}
      onClick={() => track("blog_tag_filter", { tag, from })}
      className={`inline-flex items-center gap-1 rounded-full border border-border bg-surface ${padding} ${fontSize} text-muted transition-colors hover:border-accent/40 hover:text-foreground`}
    >
      <span className="text-accent">#</span>
      {label}
    </Link>
  );
}
