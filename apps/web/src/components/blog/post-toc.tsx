"use client";

// Design Ref: §2.4 FR-11 — Auto TOC. Extracts h2/h3 from rendered article at
// mount time. Desktop only (sticky sidebar). Plan D7.
import { useEffect, useState } from "react";

type Heading = { id: string; text: string; level: 2 | 3 };

export function PostTOC() {
  const [headings, setHeadings] = useState<Heading[]>([]);
  const [active, setActive] = useState<string>("");

  useEffect(() => {
    const article = document.querySelector("article");
    if (!article) return;
    const els = article.querySelectorAll("h2, h3");
    const items: Heading[] = [];
    els.forEach((el) => {
      if (!el.id) return;
      items.push({
        id: el.id,
        text: el.textContent?.replace(/#$/, "").trim() ?? "",
        level: el.tagName === "H2" ? 2 : 3,
      });
    });
    setHeadings(items);

    // Scroll spy via IntersectionObserver
    const io = new IntersectionObserver(
      (entries) => {
        const visible = entries.filter((e) => e.isIntersecting);
        if (visible.length > 0) {
          setActive(visible[0].target.id);
        }
      },
      { rootMargin: "-20% 0px -70% 0px" },
    );
    els.forEach((el) => io.observe(el));
    return () => io.disconnect();
  }, []);

  if (headings.length < 2) return null;

  return (
    <nav aria-label="Table of contents" className="text-sm">
      <div className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted">
        On this page
      </div>
      <ul className="space-y-2">
        {headings.map((h) => (
          <li key={h.id} className={h.level === 3 ? "pl-4" : ""}>
            <a
              href={`#${h.id}`}
              className={`block border-l-2 py-1 pl-3 transition-colors ${
                active === h.id
                  ? "border-accent text-foreground"
                  : "border-transparent text-muted hover:border-border hover:text-foreground"
              }`}
            >
              {h.text}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  );
}
