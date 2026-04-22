"use client";

// Design Ref: §2.4 FR-18 — Tag cloud on blog index. Fires blog_tag_filter.
import Link from "next/link";
import { getTagLabel } from "@/lib/tags";
import { track } from "@/lib/track";

type Props = {
  tags: Array<{ tag: string; count: number }>;
};

export function TagCloud({ tags }: Props) {
  if (tags.length === 0) return null;

  return (
    <div className="mx-auto max-w-4xl">
      <div className="flex flex-wrap items-center justify-center gap-2">
        {tags.map(({ tag, count }) => (
          <Link
            key={tag}
            href={`/blog/tag/${tag}`}
            onClick={() => track("blog_tag_filter", { tag, from: "index" })}
            className="inline-flex items-center gap-2 rounded-full border border-border bg-surface/60 px-3 py-1.5 text-sm text-muted backdrop-blur-sm transition-colors hover:border-accent/40 hover:text-foreground"
          >
            <span className="text-accent">#</span>
            {getTagLabel(tag)}
            <span className="text-xs text-muted">({count})</span>
          </Link>
        ))}
      </div>
    </div>
  );
}
