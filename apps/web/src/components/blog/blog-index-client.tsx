"use client";

// Client component that reads ?tags=a,b from the URL and filters the SSR'd
// post list in-place. Renders the full list initially so users without JS
// still see all articles. Kept separate from page.tsx to keep the server
// component pure (no client hooks).
import { useMemo } from "react";
import { useSearchParams } from "next/navigation";
import { PostCard } from "@/components/blog/post-card";
import type { BlogPostMeta } from "@/lib/blog";

type Props = {
  posts: BlogPostMeta[];
};

function parseSelected(raw: string | null): Set<string> {
  if (!raw) return new Set();
  return new Set(
    raw
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean),
  );
}

export function BlogIndexClient({ posts }: Props) {
  const searchParams = useSearchParams();
  const selected = useMemo(
    () => parseSelected(searchParams.get("tags")),
    [searchParams],
  );

  const filtered = useMemo(() => {
    if (selected.size === 0) return posts;
    // AND mode: a post matches only if it contains every selected tag.
    return posts.filter((p) =>
      [...selected].every((t) => p.tags.includes(t as never)),
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
    <ul className="mx-auto grid max-w-4xl gap-4 sm:grid-cols-2">
      {filtered.map((post) => (
        <li key={post.slug}>
          <PostCard post={post} />
        </li>
      ))}
    </ul>
  );
}
