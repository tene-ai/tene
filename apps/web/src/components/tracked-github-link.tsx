"use client";

// Client wrapper for a GitHub anchor that fires `github_click` with a location
// tag. Use this everywhere a `<a href="https://github.com/tene-ai/tene">`
// appears in the landing page + blog + /vs pages.
import { track } from "@/lib/track";

type Location =
  | "nav"
  | "hero"
  | "footer"
  | "vs_page"
  | "blog_post"
  | "cta"
  | "pricing"
  | "security";

type Props = {
  href: string;
  location: Location;
  className?: string;
  children: React.ReactNode;
  ariaLabel?: string;
  // Some places use target=_self (e.g. same-tab nav); default is _blank.
  openInNewTab?: boolean;
};

export function TrackedGithubLink({
  href,
  location,
  className,
  children,
  ariaLabel,
  openInNewTab = true,
}: Props) {
  return (
    <a
      href={href}
      target={openInNewTab ? "_blank" : undefined}
      rel={openInNewTab ? "noopener noreferrer" : undefined}
      onClick={() => track("github_click", { location })}
      className={className}
      aria-label={ariaLabel}
    >
      {children}
    </a>
  );
}
