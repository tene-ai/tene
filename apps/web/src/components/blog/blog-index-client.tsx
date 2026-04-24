"use client";

// Client component that reads ?tags=a,b from the URL and filters the SSR'd
// post list in place. Uses window.location.search rather than
// next/navigation's useSearchParams — that hook combined with a Suspense
// fallback in a statically-prerendered route was causing the initial render
// to show the unfiltered fallback without ever swapping to the filtered
// view. Reading directly from the URL after mount guarantees the filter
// reflects whatever the browser actually loaded.
import { useEffect, useMemo, useState } from "react";
import { PostCard } from "@/components/blog/post-card";
import type { BlogPostMeta } from "@/lib/blog";

type Props = {
  posts: BlogPostMeta[];
};

export const TAG_CHANGED_EVENT = "tene:blog-tags-changed";

function parseSelected(raw: string | null | undefined): Set<string> {
  if (!raw) return new Set();
  return new Set(
    raw
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean),
  );
}

function readTagsFromUrl(): Set<string> {
  if (typeof window === "undefined") return new Set();
  return parseSelected(new URLSearchParams(window.location.search).get("tags"));
}

export function BlogIndexClient({ posts }: Props) {
  const [selected, setSelected] = useState<Set<string>>(new Set());

  useEffect(() => {
    setSelected(readTagsFromUrl());
    const sync = () => setSelected(readTagsFromUrl());
    window.addEventListener("popstate", sync);
    window.addEventListener(TAG_CHANGED_EVENT, sync);
    return () => {
      window.removeEventListener("popstate", sync);
      window.removeEventListener(TAG_CHANGED_EVENT, sync);
    };
  }, []);

  const filtered = useMemo(() => {
    if (selected.size === 0) return posts;
    return posts.filter((p) =>
      [...selected].every((t) => (p.tags as readonly string[]).includes(t)),
    );
  }, [posts, selected]);

  if (filtered.length === 0) {
    return (
      <div className="mx-auto max-w-xl rounded-lg border border-border bg-surface/50 p-8 text-center text-muted">
        <p className="font-medium text-foreground">No articles match.</p>
        <p className="mt-2 text-sm">
          Try removing a tag filter, or browse{" "}
          <a
            href="/blog"
            className="underline underline-offset-4 hover:text-foreground"
          >
            all articles
          </a>
          .
        </p>
      </div>
    );
  }

  return (
    <ul
      className="mx-auto grid max-w-4xl gap-4 sm:grid-cols-2"
      data-filter-active={selected.size > 0 ? "true" : "false"}
    >
      {filtered.map((post) => (
        <li key={post.slug}>
          <PostCard post={post} />
        </li>
      ))}
    </ul>
  );
}
