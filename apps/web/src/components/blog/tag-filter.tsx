"use client";

// Instagram-style tag filter on /blog index. Client-only filtering via URL
// query param (?tags=a,b). AND mode: a post is shown only if it carries
// ALL selected tags. Collapsed by default to avoid clutter.
import { useCallback, useMemo, useState, useTransition } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { getTagLabel, type TagKey } from "@/lib/tags";
import { track } from "@/lib/track";

type TagEntry = { tag: string; count: number };

type Props = {
  allTags: TagEntry[]; // sorted by count desc
  topN?: number; // how many to show before "Show all"
};

function parseTagsParam(raw: string | null): Set<string> {
  if (!raw) return new Set();
  return new Set(
    raw
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean),
  );
}

export function TagFilter({ allTags, topN = 6 }: Props) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [showAll, setShowAll] = useState(false);
  const [, startTransition] = useTransition();

  const selected = useMemo(
    () => parseTagsParam(searchParams.get("tags")),
    [searchParams],
  );

  const filteredTags = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return allTags;
    return allTags.filter(
      ({ tag }) =>
        tag.toLowerCase().includes(q) ||
        getTagLabel(tag).toLowerCase().includes(q),
    );
  }, [allTags, search]);

  const visibleTags = useMemo(() => {
    if (search.trim()) return filteredTags; // search overrides topN
    if (showAll) return filteredTags;
    return filteredTags.slice(0, topN);
  }, [filteredTags, showAll, search, topN]);

  const updateUrl = useCallback(
    (nextSelected: Set<string>) => {
      const params = new URLSearchParams(searchParams.toString());
      if (nextSelected.size === 0) {
        params.delete("tags");
      } else {
        params.set("tags", [...nextSelected].sort().join(","));
      }
      const query = params.toString();
      startTransition(() => {
        router.replace(query ? `/blog?${query}` : "/blog", { scroll: false });
      });
    },
    [router, searchParams],
  );

  const toggle = (tag: string) => {
    const next = new Set(selected);
    if (next.has(tag)) {
      next.delete(tag);
      track("blog_tag_filter", { tag, action: "remove", from: "filter" });
    } else {
      next.add(tag);
      track("blog_tag_filter", { tag, action: "add", from: "filter" });
    }
    updateUrl(next);
  };

  const clearAll = () => {
    track("blog_tag_filter", { action: "clear_all", from: "filter" });
    updateUrl(new Set());
  };

  const hasActive = selected.size > 0;

  return (
    <div className="mx-auto max-w-4xl">
      <details
        open={open || hasActive}
        onToggle={(e) => setOpen((e.currentTarget as HTMLDetailsElement).open)}
        className="rounded-lg border border-border bg-surface/40 backdrop-blur-sm"
      >
        <summary className="flex cursor-pointer items-center justify-between px-4 py-3 text-sm font-medium text-muted hover:text-foreground">
          <span className="flex items-center gap-2">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="16"
              height="16"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              aria-hidden
            >
              <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
            </svg>
            Filter by topic{" "}
            {hasActive && (
              <span className="rounded-full bg-accent/20 px-2 py-0.5 text-xs text-accent">
                {selected.size}
              </span>
            )}
          </span>
          <span className="text-xs opacity-60">
            {open || hasActive ? "Close" : "Open"}
          </span>
        </summary>

        <div className="border-t border-border/60 p-4">
          <div className="flex items-center gap-2">
            <input
              type="text"
              inputMode="search"
              placeholder="Search topics..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="flex-1 rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted focus:border-accent focus:outline-none"
              aria-label="Search tags"
            />
            {hasActive && (
              <button
                onClick={clearAll}
                type="button"
                className="shrink-0 rounded-md border border-border px-3 py-2 text-sm text-muted hover:border-accent/40 hover:text-foreground"
              >
                Clear
              </button>
            )}
          </div>

          {visibleTags.length === 0 ? (
            <p className="mt-4 text-center text-sm text-muted">
              No topics match &quot;{search}&quot;.
            </p>
          ) : (
            <div className="mt-4 flex flex-wrap gap-2">
              {visibleTags.map(({ tag, count }) => {
                const isSelected = selected.has(tag);
                const stateClass = isSelected
                  ? "border-accent bg-accent/10 text-foreground"
                  : "border-border bg-surface/60 text-muted hover:border-accent/40 hover:text-foreground";
                return (
                  <button
                    key={tag}
                    type="button"
                    onClick={() => toggle(tag)}
                    aria-pressed={isSelected}
                    className={`inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-sm transition-colors ${stateClass}`}
                  >
                    <span className="text-accent">#</span>
                    {getTagLabel(tag)}
                    <span className="text-xs opacity-60">({count})</span>
                  </button>
                );
              })}
              {!showAll &&
                !search.trim() &&
                filteredTags.length > topN && (
                  <button
                    type="button"
                    onClick={() => setShowAll(true)}
                    className="inline-flex items-center gap-1 rounded-full border border-border/60 px-3 py-1.5 text-sm text-muted hover:border-accent/40 hover:text-foreground"
                  >
                    Show all ({filteredTags.length})
                  </button>
                )}
            </div>
          )}

          {hasActive && (
            <p className="mt-3 text-xs text-muted">
              Showing posts with{" "}
              <span className="text-foreground">all</span> selected topics.
            </p>
          )}
        </div>
      </details>
    </div>
  );
}
