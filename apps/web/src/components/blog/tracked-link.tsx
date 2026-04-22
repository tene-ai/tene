"use client";

// Design Ref: §2.8 FR-36 — External link wrapper used by MDX `a` renderer.
// Fires blog_external_link on click.
import { track } from "@/lib/track";

type Props = {
  href: string;
  slug?: string;
  children: React.ReactNode;
} & Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, "href">;

export function TrackedLink({ href, slug, children, ...rest }: Props) {
  function handleClick() {
    try {
      const domain = new URL(href).hostname;
      if (slug) {
        track("blog_external_link", { slug, domain });
      }
    } catch {
      // URL parsing failed — skip tracking.
    }
  }

  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      onClick={handleClick}
      className="text-accent underline underline-offset-4 hover:text-accent-dim"
      {...rest}
    >
      {children}
    </a>
  );
}
