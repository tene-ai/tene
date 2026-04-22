"use client";

// Design Ref: §2.8 FR-37 — Fires blog_read_complete when user scrolls past
// article's end sentinel (IntersectionObserver, threshold 0.3, fire-once).
import { useEffect } from "react";
import { track } from "@/lib/track";

type Props = {
  slug: string;
  readingMinutes: number;
};

export function ReadProgressTracker({ slug, readingMinutes }: Props) {
  useEffect(() => {
    const sentinel = document.getElementById("blog-end-sentinel");
    if (!sentinel) return;
    let fired = false;

    const io = new IntersectionObserver(
      (entries) => {
        entries.forEach((e) => {
          if (e.isIntersecting && !fired) {
            fired = true;
            track("blog_read_complete", { slug, readingMinutes });
          }
        });
      },
      { threshold: 0.3 },
    );

    io.observe(sentinel);
    return () => io.disconnect();
  }, [slug, readingMinutes]);

  return null;
}
