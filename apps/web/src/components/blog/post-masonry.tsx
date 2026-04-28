"use client";

// Row-major masonry grid for post cards. CSS Grid with grid-auto-rows: 1px
// + per-card grid-row: span N (computed from measured height via
// ResizeObserver). Reading order is DOM order: top-row left-to-right is
// newest-first, then row 2, etc.
//
// Used by /blog (wrapped with tag filter), /blog/category/[c], and
// /blog/tag/[t]. Server-side renders the cards with a default span; client
// hook re-measures on mount and on resize.
import { useEffect, useRef } from "react";
import { PostCard } from "@/components/blog/post-card";
import type { BlogPostMeta } from "@/lib/blog";

type Props = {
  posts: BlogPostMeta[];
};

const ROW_PX = 1;
const GAP_PX = 24; // matches Tailwind `gap-6`

function calcSpan(heightPx: number): number {
  return Math.ceil((heightPx + GAP_PX) / (ROW_PX + GAP_PX));
}

export function PostMasonry({ posts }: Props) {
  const gridRef = useRef<HTMLUListElement>(null);

  useEffect(() => {
    const grid = gridRef.current;
    if (!grid) return;

    const updateSpans = () => {
      const items = grid.querySelectorAll<HTMLLIElement>("li[data-card]");
      items.forEach((li) => {
        const card = li.firstElementChild as HTMLElement | null;
        if (!card) return;
        const h = card.getBoundingClientRect().height;
        if (h > 0) li.style.setProperty("--row-span", String(calcSpan(h)));
      });
    };

    updateSpans();

    const ro = new ResizeObserver(updateSpans);
    grid.querySelectorAll("li[data-card] > *").forEach((card) => {
      ro.observe(card);
    });
    window.addEventListener("resize", updateSpans);

    return () => {
      ro.disconnect();
      window.removeEventListener("resize", updateSpans);
    };
  }, [posts]);

  return (
    <ul
      ref={gridRef}
      className="mx-auto grid max-w-4xl grid-cols-1 gap-6 [grid-auto-rows:1px] sm:grid-cols-2 lg:max-w-6xl lg:grid-cols-3 xl:max-w-[1500px] xl:grid-cols-4 2xl:max-w-[1800px]"
    >
      {posts.map((post) => (
        <li
          key={post.slug}
          data-card
          style={{
            ["--row-span" as string]: 200,
            gridRow: "span var(--row-span)",
          }}
        >
          <PostCard post={post} />
        </li>
      ))}
    </ul>
  );
}
