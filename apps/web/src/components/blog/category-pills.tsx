"use client";

// Category pills shown at the top of /blog index. Primary navigation axis
// for the 2-layer taxonomy (category + tag). See blog-content.md §2.
import Link from "next/link";
import { track } from "@/lib/track";
import type { CategoryKey } from "@/lib/tags";

type CategoryEntry = {
  category: CategoryKey;
  count: number;
  label: string;
};

type Props = {
  categories: CategoryEntry[];
  activeCategory?: CategoryKey | null;
};

export function CategoryPills({ categories, activeCategory }: Props) {
  return (
    <div className="mx-auto flex max-w-4xl flex-wrap items-center justify-center gap-2">
      {categories.map(({ category, count, label }) => {
        const isActive = activeCategory === category;
        const stateClass = isActive
          ? "border-accent bg-accent/10 text-foreground"
          : "border-border bg-surface/60 text-muted hover:border-accent/40 hover:text-foreground";
        return (
          <Link
            key={category}
            href={`/blog/category/${category}`}
            aria-current={isActive ? "page" : undefined}
            onClick={() =>
              track("blog_category_filter", {
                category,
                from: isActive ? "active" : "index",
              })
            }
            className={`inline-flex items-center gap-2 rounded-full border px-4 py-2 text-sm backdrop-blur-sm transition-colors ${stateClass}`}
          >
            {label}
            <span className="text-xs opacity-60">({count})</span>
          </Link>
        );
      })}
    </div>
  );
}
