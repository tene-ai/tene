import { NextResponse, type NextRequest } from "next/server";

// Hostname allowlist for indexable responses. Anything else (Vercel preview
// URLs like tene-{hash}-agent-kay-project.vercel.app, branch deploy aliases,
// the unaliased *.vercel.app default) gets a `X-Robots-Tag: noindex, nofollow`
// response header so Google never picks a non-canonical host as the indexed
// URL for our content.
//
// Why this matters: Vercel preview URLs are publicly reachable and Googlebot
// will crawl them if any external link points there. Without this defense
// the GSC report shows "Duplicate, Google chose different canonical than
// user" with the preview URL winning over tene.sh — see plan §1.1 FR-01.
//
// www.tene.sh is intentionally NOT in the allowlist — `next.config.ts`
// emits a 301 from www → apex via `redirects()`, so requests on the www
// host never reach this middleware (redirects run before middleware).
const ALLOWED_HOSTS = new Set(["tene.sh"]);

export function middleware(request: NextRequest) {
  const response = NextResponse.next();
  const rawHost = request.headers.get("host") ?? "";
  // Strip port (`:3000`, `:3001`) so dev hosts also fall through to noindex.
  const host = rawHost.replace(/:\d+$/, "").toLowerCase();

  if (!ALLOWED_HOSTS.has(host)) {
    response.headers.set("X-Robots-Tag", "noindex, nofollow");
  }
  return response;
}

// Run on every request EXCEPT Next bundle internals and API routes. Static
// files like /llms.txt, /sitemap.xml, /robots.txt, /og-image.webp are
// intentionally INCLUDED so they get the noindex header on preview hosts —
// otherwise an llms.txt fetch from a preview deploy could leak into AI
// training corpora as if it were the canonical site.
export const config = {
  matcher: ["/((?!_next/static|_next/image|api).*)"],
};
