# Codebase Deep-Analysis — web-seo-aeo-analytics-strategy

- **Feature**: `web-seo-aeo-analytics-strategy`
- **Branch**: `feature/web-landing-and-tech-blog`
- **Date**: 2026-04-22
- **Purpose**: Pre-Design codebase audit. Bridges PRD (what) with Design (how) by verifying Next.js 16 APIs against the Plan's §8 Next 15 snippets, mapping all 18 P0 tasks to concrete file paths, and flagging integration risks.
- **PRD ref**: `docs/00-pm/web-seo-aeo-analytics-strategy.prd.md`
- **Plan ref**: `docs/01-plan/web-seo-aeo-analytics-strategy.md`

---

## 1. Current State Snapshot

### 1.1 Directory Structure (apps/web/)

```
apps/web/
├── AGENTS.md                     # "This is NOT the Next.js you know" warning
├── CLAUDE.md                     # @AGENTS.md, tech stack, design system
├── package.json                  # Next 16.2.2, React 19.2.4, TS 5 — NO homepage/repository
├── next.config.ts                # Security headers only (HSTS, X-Frame, CSP, Referrer, Permissions)
├── tsconfig.json                 # Paths: @/* -> ./src/*
├── postcss.config.mjs            # @tailwindcss/postcss v4
├── eslint.config.mjs             # eslint-config-next flat config
├── public/
│   ├── robots.txt                # ⚠️ STATIC — allow /, sitemap ref (MUST DELETE at Do)
│   ├── sitemap.xml               # ⚠️ STATIC — single URL (MUST DELETE at Do)
│   ├── llms.txt                  # Slim AI discovery file — 64 lines (KEEP; coexists w/ llms-full.txt)
│   ├── install.sh                # CLI installer (unrelated to this feature)
│   ├── favicon.svg, og-image.png, apple-touch-icon.png, logo.svg
│   └── (no .well-known/ dir)     # security.txt will live here
└── src/
    ├── app/
    │   ├── layout.tsx            # Metadata + 3 JSON-LD schemas (84-176) + NoiseOverlay mount
    │   ├── page.tsx              # Home composition: Hero/Features/HowItWorks/Security/Comparison/FAQ/CTA
    │   └── globals.css           # Tailwind v4 + CSS vars (--background #0a0a0a, --accent #00ff88, etc.)
    ├── components/               # 17 components (all client or server pure)
    │   ├── hero.tsx              # Uses heroData, CopyCommand, Terminal
    │   ├── faq.tsx               # 'use client' — 8 FAQ items from data/faq.ts
    │   ├── features.tsx, comparison.tsx, how-it-works.tsx, security.tsx
    │   ├── cta.tsx               # "Stop using .env files" + CopyCommand
    │   ├── footer.tsx, nav.tsx
    │   ├── copy-command.tsx      # 'use client' — clipboard copy button (INSTRUMENT TARGET)
    │   ├── terminal.tsx, interactive-grid.tsx, noise-overlay.tsx, glow-card.tsx
    │   └── pricing.tsx           # commented out in page.tsx
    └── data/                     # Copy separated from UI
        ├── hero.ts               # heroData.badge, h1, h1Accent, sub, cta
        ├── faq.ts                # 8 FAQ items (Q-shape + narrative answers)
        ├── features.ts, comparison.ts, how-it-works.ts, security.ts, pricing.ts
```

### 1.2 Dependencies (current, apps/web/package.json)

```json
{
  "dependencies": {
    "next": "16.2.2",
    "react": "19.2.4",
    "react-dom": "19.2.4"
  },
  "devDependencies": {
    "@tailwindcss/postcss": "^4",
    "@types/node": "^20",
    "@types/react": "^19",
    "@types/react-dom": "^19",
    "eslint": "^9",
    "eslint-config-next": "16.2.2",
    "tailwindcss": "^4",
    "typescript": "^5"
  }
}
```

**Gaps**: `homepage`, `repository`, `description`, `license`, `author` fields all missing. No analytics, no MDX, no typography plugin installed.

### 1.3 Existing SEO/AEO State (file:line evidence)

| Asset | Location | State |
|---|---|---|
| Root Metadata export | `src/app/layout.tsx:16-78` | Full: title, description, 15 keywords, authors, creator, metadataBase, canonical, icons, openGraph (url/siteName/type/images), twitter (card/title/description/images), robots (index:true, follow:true) |
| JSON-LD | `src/app/layout.tsx:80-177` injected via `<script>` in `<head>` (line 191-194) | 3 schemas under `@graph`: `SoftwareApplication` (85-97), `FAQPage` (99-150, 6 Q&As), `HowTo` (152-176, 4 steps) |
| Fonts | `src/app/layout.tsx:2,6-14,188` | `Geist`, `Geist_Mono` via `next/font/google` — keep as-is |
| Semantic HTML | `src/app/page.tsx:14-35` | `<main>`, `<section>`, `<nav>`, `<footer>`, `<h1/h2>` correct hierarchy |
| Security headers | `next.config.ts:3-10` | HSTS, X-Frame DENY, X-Content-Type-Options, Referrer-Policy, CSP (default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'), Permissions-Policy |
| FAQ copy | `src/data/faq.ts:4-45` | 8 items — Q are question-shaped, but answers are narrative (not 40-60 word capsules) |
| Hero copy | `src/data/hero.ts:3-15` | Badge + H1 split + sub + CTA — sub is 18 words (too short for AEO capsule) |

---

## 2. Next.js 16 API Verification (vs Plan §8 Next 15 Snippets)

**Verification source**: `apps/web/node_modules/next/dist/lib/metadata/types/metadata-interface.d.ts` + Next.js official v16 upgrade guide.

### 2.1 `MetadataRoute.Sitemap` — VERIFIED compatible

Type in Next 16 (`metadata-interface.d.ts:562-572`):
```ts
type SitemapFile = Array<{
  url: string;
  lastModified?: string | Date | undefined;
  changeFrequency?: 'always' | 'hourly' | 'daily' | 'weekly' | 'monthly' | 'yearly' | 'never';
  priority?: number | undefined;
  alternates?: { languages?: Languages<string> | undefined } | undefined;
  images?: string[] | undefined;
  videos?: Videos[] | undefined;
}>;
namespace MetadataRoute { type Sitemap = SitemapFile; ... }
```

**Plan §8.2 snippet is COMPATIBLE.** No changes needed. The `changeFrequency` literal must use `as const` (Plan already does this). The exported default function may be `async`.

### 2.2 `MetadataRoute.Robots` — VERIFIED compatible (with nuance)

Type in Next 16 (`metadata-interface.d.ts:547-561`):
```ts
type RobotsFile = {
  rules:
    | { userAgent?: string | string[]; allow?: string | string[]; disallow?: string | string[]; crawlDelay?: number; }
    | Array<{ userAgent: string | string[]; allow?: string | string[]; disallow?: string | string[]; crawlDelay?: number; }>;
  sitemap?: string | string[] | undefined;
  host?: string | undefined;
};
```

**Plan §8.3 snippet is COMPATIBLE.** Array-of-rules form works.

### 2.3 `generateMetadata` + `params` — **BREAKING CHANGE** (v15→v16)

**Next 16 requires `params` to be a Promise.** Plan §8.5 snippet ALREADY uses `Promise<{ slug: string }>` — good, this was a Next 15.x async-params preview. Must be `await params` before property access.

Correct Next 16 shape (matches Plan §8.5):
```ts
export async function generateMetadata(
  { params }: { params: Promise<{ slug: string }> }
): Promise<Metadata> {
  const { slug } = await params;
  const post = await getPostBySlug(slug);
  if (!post) return {};
  return { title: ..., description: ..., alternates: { canonical: `...` } };
}

export async function generateStaticParams() {
  const slugs = await getAllPostSlugs();
  return slugs.map(p => ({ slug: p.slug }));  // NOTE: return type is NOT Promise; array of { slug } objects
}
```

**IMPORTANT**: `generateStaticParams` return values are **NOT** wrapped in Promise in Next 16 — each item is `{ slug: string }`. Only the page's/layout's `params` prop is a Promise. Plan §8.5 snippet is correct on this.

### 2.4 `generateSitemaps` — **BREAKING CHANGE** (v15→v16)

If we decide to split sitemap (NOT planned for P0 — single `sitemap.ts` is fine), Next 16 now passes `id` as Promise: `sitemap({ id }: { id: Promise<string> })`. **Not relevant to P0** (single sitemap for now). Flag for future if we shard to >50k URLs.

### 2.5 `next/og` `ImageResponse` — PRESENT & WORKING

`apps/web/node_modules/next/og.d.ts` exists; ImageResponse class exported. **Dynamic OG per-post (`opengraph-image.tsx`) is deferred to P1** per PRD §6 Out of Scope — not a P0 task.

### 2.6 `@next/third-parties` — NOT INSTALLED

Plan §8 references `@next/third-parties/google` for `<GoogleAnalytics>`. Package is NOT in node_modules. Must install. Next 16 compatibility: `@next/third-parties@16.x` exists and ships the `GoogleAnalytics` component with `sendGAEvent` helper. **Confirm version on install**: `@next/third-parties@16.2.2` (match peer).

### 2.7 `next-mdx-remote/rsc` — REQUIRES DECISION

- next-mdx-remote v5+ supports MDX v3 and React 19, imports via `next-mdx-remote/rsc` for server components.
- **Stability note**: Library docs mark RSC integration as "unstable" (may change) — but widely used in production.
- Next 16 RSC works; `<MDXRemote source={...} />` renders as async server component.
- **Alternative**: `@next/mdx` (native, file-based, `.mdx` pages) — but Plan §4 YAGNI rejects Contentlayer and commits to `next-mdx-remote`. Native `@next/mdx` would mix `content/posts/*.mdx` with `src/app/` which contradicts Plan.
- **Recommendation**: Stick with `next-mdx-remote/rsc` + `gray-matter` + `shiki` as Plan specifies. Record this as a Design Decision (R-5 mitigation).

### 2.8 Summary — Next 16 Deltas from Plan §8

| Plan §8 location | Status | Action |
|---|---|---|
| §8.1 `<Analytics />`, `<SpeedInsights />`, `<GoogleAnalytics>` | Compatible | Install deps; mount in `<body>` (not `<head>`) |
| §8.2 `sitemap.ts` | Compatible | No change; default export may be async |
| §8.3 `robots.ts` | Compatible | No change |
| §8.4 Organization/Person/BreadcrumbList | Compatible | Extend `@graph` array in layout.tsx |
| §8.5 `[slug]/page.tsx` w/ `Promise<{ slug }>` | **Compatible (Next 15 async-params preview became default in 16)** | Keep `await params` — confirmed required |
| §8.6 RSS `/feed.xml` route handler | Compatible | Next 16 route handler API unchanged; return `new Response(xml, { headers })` |
| §8.7 Cookie banner `'use client'` | Compatible | Standard RSC pattern |

**Conclusion**: Plan §8 snippets are mostly safe. Main watch-outs: (1) `await params` in generateMetadata and in the page component itself; (2) `@next/third-parties` version pin; (3) treat `next-mdx-remote/rsc` as unstable-but-accepted.

---

## 3. Task-to-File Mapping (18 P0 Tasks)

### Legend
- **NEW**: file created from scratch
- **MOD**: existing file modified (with line refs)
- **DEL**: existing file deleted
- **DEP**: dependency added to `apps/web/package.json`

### A. Analytics (6 tasks)

| # | P0 Task | Type | Path | Details |
|---|---|---|---|---|
| P0-1 | Vercel Analytics + Speed Insights | MOD | `apps/web/src/app/layout.tsx` | After line 197 (inside `<body>`, before `<NoiseOverlay />`), mount `<Analytics />` and `<SpeedInsights />` |
| P0-1 | | DEP | `apps/web/package.json` | `@vercel/analytics@^1`, `@vercel/speed-insights@^1` |
| P0-2 | GA4 via `@next/third-parties` | MOD | `apps/web/src/app/layout.tsx` | In `<body>`, conditionally mount `{gaId && <GoogleAnalytics gaId={gaId} />}`. Read `process.env.NEXT_PUBLIC_GA_ID` once |
| P0-2 | | DEP | `apps/web/package.json` | `@next/third-parties@16.2.2` |
| P0-2 | | NEW | `apps/web/.env.example` | Document `NEXT_PUBLIC_GA_ID=G-XXXXXXXXXX` (no real value, example only) |
| P0-3 | Cookie consent banner (default-deny) | NEW | `apps/web/src/components/cookie-banner.tsx` | `'use client'`; reads `localStorage["tene-consent"]`; renders default-deny banner; on decision, calls `window.gtag?.('consent', 'update', {...})`. Also sets defaults via `window.gtag?.('consent', 'default', { analytics_storage: 'denied', ad_storage: 'denied' })` in a `useEffect` or via inline pre-consent-default script (see §4 decision) |
| P0-3 | | MOD | `apps/web/src/app/layout.tsx` | Mount `<CookieBanner />` inside `<body>` above `<NoiseOverlay />` |
| P0-4 | GA4 property + vault env | USER | External | User creates GA4 property; runs `tene set NEXT_PUBLIC_GA_ID G-XXX --env prod`; adds var in Vercel dashboard for Production + Preview |
| P0-14 | Event instrumentation (4 events) | MOD | `apps/web/src/components/copy-command.tsx:14-18` | In `handleCopy`, call `sendGAEvent('event', 'install_script_copy', { location: props.location })` after successful clipboard write. Props extended with `eventName?: string` or a per-usage `location` param |
| P0-14 | | MOD | `apps/web/src/components/hero.tsx:32-41` | GitHub anchor: `onClick={() => sendGAEvent('event', 'github_release_click', { source: 'hero' })}` (note: it's a link to the repo, not releases — rename event accordingly to `github_click` OR change `href` to releases page per PRD intent) |
| P0-14 | | MOD | `apps/web/src/components/cta.tsx:20-31` | CTA GitHub link: `sendGAEvent('event', 'cta_click', { location: 'footer-cta' })` |
| P0-14 | | NEW (optional) | `apps/web/src/lib/analytics.ts` | Thin wrapper around `sendGAEvent` that no-ops when consent denied (reads `localStorage["tene-consent"]`). Prevents R-7 event leak pre-consent |
| P0-14 | | MOD | `apps/web/src/components/nav.tsx` | ClawHub link does NOT exist in Nav today — this requires user decision: do we add a ClawHub link to nav/footer? See §7 Open Q |

**Notes on event instrumentation**:
- The CopyCommand component is reused by Hero AND CTA. Pass `location` prop (e.g., `location="hero"` / `location="cta"`) so we can distinguish via GA event parameters.
- The 4 events per PRD are: `install_script_copy`, `github_release_click`, `clawhub_link_click`, `cta_click`. ClawHub link target currently absent — clarify at §7.

### B. SEO / AEO Infrastructure (8 tasks)

| # | P0 Task | Type | Path | Details |
|---|---|---|---|---|
| P0-5 | GSC DNS verification | USER | External (DNS) | Add TXT record at DNS provider (Vercel DNS or Cloudflare). User-only step |
| P0-6 | Dynamic sitemap | NEW | `apps/web/src/app/sitemap.ts` | Default export `async function sitemap(): Promise<MetadataRoute.Sitemap>` — returns `[{ url: 'https://tene.sh', ... }, { url: '.../blog', ... }, ...posts]`. Imports `getAllPostSlugs` from `@/lib/blog` |
| P0-6 | | DEL | `apps/web/public/sitemap.xml` | MUST delete to avoid Next 16 file-vs-route conflict (Next.js prefers app-route over public file but dual presence is a correctness hazard) |
| P0-7 | Dynamic robots w/ AI allowlist | NEW | `apps/web/src/app/robots.ts` | Default export returns `MetadataRoute.Robots` with rules[] including `*` + 12 AI bots per Plan §8.3 + `sitemap: 'https://tene.sh/sitemap.xml'` + `host: 'https://tene.sh'` |
| P0-7 | | DEL | `apps/web/public/robots.txt` | MUST delete for same reason as sitemap.xml |
| P0-8 | Submit sitemap to GSC | USER | External (GSC UI) | After P0-5 completes + sitemap.ts ships to prod, user submits `sitemap.xml` in GSC Sitemaps panel |
| P0-9 | Organization + Person + BreadcrumbList JSON-LD | MOD | `apps/web/src/app/layout.tsx:82-177` | Extend `@graph` array (line 82) from 3 schemas → 6. Keep existing 3. Add per Plan §8.4: Organization (w/ founder→Person), Person (standalone full w/ sameAs array), BreadcrumbList (single item: Home) |
| P0-10 | llms-full.txt auto-gen | NEW | `apps/web/scripts/gen-llms-full.ts` | Node script: reads `public/llms.txt`, `src/app/page.tsx` text content, `content/posts/*.mdx` frontmatter+body, key CLI SKILL.md files, concatenates to `public/llms-full.txt` |
| P0-10 | | MOD | `apps/web/package.json` | Add `"prebuild": "tsx scripts/gen-llms-full.ts"` OR `"build": "tsx scripts/gen-llms-full.ts && next build"` (decision in §4) |
| P0-10 | | DEP | `apps/web/package.json` (dev) | `tsx` (to run TS script without compile) |
| P0-11 | security.txt (RFC 9116) | NEW | `apps/web/public/.well-known/security.txt` | Plain text with `Contact: mailto:security@tene.sh`, `Expires: <ISO date 1y out>`, `Preferred-Languages: en`, optional `Encryption:` (PGP key URL), `Policy:` (URL). User provides contact/PGP per D-6 |
| P0-12 | package.json metadata | MOD | `apps/web/package.json` | Add `"homepage": "https://tene.sh"`, `"repository": { "type": "git", "url": "https://github.com/agent-kay-it/tene" }`, `"description": "Landing page for Tene — local-first encrypted secret manager"`, `"license": "MIT"`, `"author": "Kay Kim <kay@popupstudio.ai>"` (per PRD §7 Open Q) |
| P0-13 | Answer capsule rewrites | MOD | `apps/web/src/data/hero.ts:7` | Rewrite `sub` to 40-60 word self-contained "what is Tene" capsule |
| P0-13 | | MOD | `apps/web/src/data/faq.ts` (top 3 items, lines 4-19) | Rewrite first 3 answers to 40-60 word capsules. Candidates to rewrite: "Why is .env dangerous with AI agents?", "How does Tene keep secrets from AI?", "What is Tene?". See §4.6 for proposed rewrites |

### C. Blog Rail (4 tasks)

| # | P0 Task | Type | Path | Details |
|---|---|---|---|---|
| P0-15 | Blog scaffold + MDX pipeline | NEW | `apps/web/src/app/blog/page.tsx` | Server component index; lists posts from `@/lib/blog#getAllPosts` sorted by date desc; no MDX needed on index |
| P0-15 | | NEW | `apps/web/src/app/blog/[slug]/page.tsx` | Per Plan §8.5 pattern; `generateStaticParams`, `generateMetadata` (async params), default export async page renders `<MDXRemote source={post.content} />`; injects Article + BreadcrumbList JSON-LD (P0-18) |
| P0-15 | | NEW | `apps/web/src/app/blog/[slug]/not-found.tsx` (optional) | Custom 404 for blog |
| P0-15 | | NEW | `apps/web/src/lib/blog.ts` | `getAllPosts()`, `getAllPostSlugs()`, `getPostBySlug(slug)`. Reads `content/posts/*.mdx` via `fs.promises` + `gray-matter` — returns `{ slug, frontmatter, content, date, author, description, title, tags, canonical }` |
| P0-15 | | NEW | `apps/web/content/posts/welcome.mdx` | Placeholder post proving index scalability (per PRD US-13 acceptance) |
| P0-15 | | DEP | `apps/web/package.json` | `next-mdx-remote@^5`, `gray-matter@^4`, `shiki@^1`, `@tailwindcss/typography@^0.5` |
| P0-15 | | MOD | `apps/web/src/app/globals.css` | Add `@plugin "@tailwindcss/typography";` (Tailwind v4 plugin syntax) + custom `.prose-invert` overrides using brand palette (accent on links, border on blockquote) |
| P0-15 | | MOD | `apps/web/src/components/nav.tsx:36-48` | Add `<a href="/blog">Blog</a>` to desktop nav links (AND mobile menu 96-123) |
| P0-16 | First post "Introducing tene" | NEW | `apps/web/content/posts/introducing-tene.mdx` | ≥1,500 words; frontmatter: `title, description, date (2026-04-22 or ship date), author (Kay Kim), tags, canonical, ogImage`; body in MDX format. User approval gate per PRD D-5 |
| P0-17 | RSS feed | NEW | `apps/web/src/app/feed.xml/route.ts` | Per Plan §8.6; `export async function GET()` returns `new Response(xml, { headers: { 'Content-Type': 'application/rss+xml; charset=utf-8' } })`. Use `CDATA` wrap for title/desc/content to prevent XML injection (R-9) |
| P0-17 | | MOD | `apps/web/src/app/layout.tsx` | Add `<link rel="alternate" type="application/rss+xml" title="tene blog" href="/feed.xml">` via Metadata `icons`/`other` or via Metadata.alternates.types. NOTE: Metadata API's `alternates.types` is the canonical Next way; verify in Design |
| P0-18 | Article + BreadcrumbList per post | MOD | `apps/web/src/app/blog/[slug]/page.tsx` | Inject `<script type="application/ld+json">` with Article schema (headline, description, datePublished, author Person, publisher Organization, mainEntityOfPage, image) + BreadcrumbList (Home → Blog → Post) |

### D. Dependency Summary (final DEP add list)

```json
// apps/web/package.json dependencies to add
"@next/third-parties": "16.2.2",
"@vercel/analytics": "^1.3.0",
"@vercel/speed-insights": "^1.1.0",
"next-mdx-remote": "^5.0.0",
"gray-matter": "^4.0.3",
"shiki": "^1.22.0"

// devDependencies to add
"@tailwindcss/typography": "^0.5.15",
"tsx": "^4.19.0"
```

All MIT-licensed (Constraint C-7 honored). Verify exact latest minor versions at Do time.

---

## 4. Architecture Decisions to Resolve (for Design Phase)

These are trade-offs needing explicit user decision at Checkpoint 3 of `/pdca design`.

### 4.1 MDX Pipeline: `next-mdx-remote/rsc` vs `@next/mdx` vs `mdx-bundler`

| Option | Pros | Cons |
|---|---|---|
| **`next-mdx-remote/rsc`** (Plan recommendation) | RSC-native, async component, source from any string | Marked "unstable" in docs; MDXProvider context doesn't work in RSC |
| `@next/mdx` | Native, file-based `.mdx` as pages, stable | Forces MDX files under `src/app/` (collides with Plan's `content/posts/` convention) |
| `mdx-bundler` | Bundles MDX+imports, isolates deps per post | Heavier; build-time only; adds esbuild |

**Recommendation**: `next-mdx-remote/rsc` + `gray-matter` (Plan choice). Accept "unstable" label as a documented risk (R-5). Pin exact version; re-evaluate on v6.

### 4.2 Analytics Client/Server Boundary

- `<Analytics />` and `<SpeedInsights />` from `@vercel/*` are client components — Next mounts them automatically; safe inside server-rendered `<body>`.
- `<GoogleAnalytics gaId />` from `@next/third-parties/google` is a client component (uses `next/script`).
- `<CookieBanner />` must be client (`'use client'`) for localStorage + onClick.

**Decision**: All 4 analytics primitives live in `layout.tsx` body. They are auto-hydrated. No wrapper needed. `layout.tsx` remains a server component; it just renders 4 client children.

### 4.3 Cookie Consent — Consent Mode v2 Integration

**Challenge**: GA4 Consent Mode v2 requires `gtag('consent', 'default', ...)` to run **before** the GA script loads. Otherwise events are tracked pre-consent.

Two patterns:

**Pattern A (simpler)**: `<GoogleAnalytics gaId />` emits a script; we place a tiny inline `<Script strategy="beforeInteractive">` above it that calls `window.gtag('consent', 'default', { analytics_storage: 'denied', ad_storage: 'denied' })`. Then `<CookieBanner>` updates on click.

**Pattern B (conditional mount)**: Only mount `<GoogleAnalytics>` AFTER consent granted. Simpler privacy story but loses Consent Mode's anonymized pings.

**Recommendation**: Pattern A (Plan §7.2 R-7 mitigation). Ship beforeInteractive consent-default shim. Document in component comments.

### 4.4 `llms-full.txt` Generation — Build-time vs On-demand

| Option | Pros | Cons |
|---|---|---|
| **`prebuild` hook** (Plan rec) | Runs once per deploy; output is a static file at `/llms-full.txt`; zero runtime cost | Requires every blog post commit to re-run build (already the case for Vercel) |
| `src/app/llms-full.txt/route.ts` | Dynamic, always fresh | Runtime cost; harder to cache; requires route handler returning text/plain |

**Recommendation**: `prebuild` hook via `tsx scripts/gen-llms-full.ts`. Add `"prebuild"` script in package.json. Output path: `apps/web/public/llms-full.txt` (served directly).

### 4.5 Event Instrumentation — Consent-gated `sendGAEvent` Wrapper

Create `apps/web/src/lib/analytics.ts`:
```ts
'use client';
import { sendGAEvent as gtagSend } from '@next/third-parties/google';

export function trackEvent(name: string, params?: Record<string, unknown>) {
  if (typeof window === 'undefined') return;
  const consent = localStorage.getItem('tene-consent');
  if (consent !== 'granted') return;  // no-op pre-consent
  gtagSend('event', name, params);
}
```
Every call site uses `trackEvent('install_script_copy', { location: 'hero' })`.

**Decision**: Accept wrapper pattern. Prevents R-7 (pre-consent event leak). Simpler than Consent Mode v2 "cookieless ping" semantics.

### 4.6 Answer Capsule Rewrites — Proposed Rewrites

Show 3 FAQ capsules (40-60 words each) for user approval at Design Checkpoint 3.

**FAQ #1 "Why is .env dangerous with AI agents?"** (proposed, 54 words):
> AI coding agents like Claude Code, Cursor, and Windsurf read every file in your project — including `.env`. Your API keys, tokens, and database passwords become plaintext context sent to foundation models. Once sent, you cannot control how that data is logged, cached, or used for future training.

**FAQ #2 "How does Tene keep secrets from AI?"** (proposed, 48 words):
> Tene stores secrets in a local encrypted SQLite vault protected by XChaCha20-Poly1305. When you run `tene run -- claude`, it injects secrets as environment variables at runtime. AI agents see the command in CLAUDE.md and the variable names, but never read the plaintext values themselves.

**FAQ #3 "What is Tene?"** (proposed, 42 words):
> Tene is a local-first, encrypted secret management CLI for developers using AI coding agents. It stores API keys and credentials in a device-only vault, never touches the network, and auto-generates AI editor rules so tools like Claude Code, Cursor, and Windsurf use secrets safely.

**Hero sub rewrite** (proposed, 45 words — replaces current 18-word line):
> Tene encrypts your secrets in a local SQLite vault using XChaCha20-Poly1305 and injects them as environment variables at runtime. AI agents like Claude Code see your commands — not your credentials. No server, no signup, no cloud dependency. MIT licensed and fully open source.

User decides at Design Checkpoint 3 whether these are accepted verbatim or revised.

### 4.7 Blog Content Location

`apps/web/content/posts/*.mdx` — sits OUTSIDE `src/`. Reasoning:
- `src/` is TS source; `content/` is authored prose (convention separation).
- `tsconfig.json` paths only include `src/*`; content files don't need to be type-checked.
- `fs.readdirSync('./content/posts')` at build time (inside `lib/blog.ts`) uses `process.cwd()` which is `apps/web/` during Next build — works.
- Vercel includes the whole repo in the build context; `content/` is copied in.

**Decision**: Commit to `apps/web/content/posts/`. Document in `apps/web/CLAUDE.md` + `AGENTS.md`.

### 4.8 Schema Injection Strategy — Central Utility vs Inline Scripts

Current state (`layout.tsx:191-194`) uses inline `<script type="application/ld+json" dangerouslySetInnerHTML={...}>` directly in JSX.

Two patterns for P0-9 (+ P0-18 per-post):

**Option A (inline, status quo)**: Each JSON-LD object injected via inline `<script>`. Simple.

**Option B (utility helper)**: `src/lib/schema.tsx`:
```tsx
export function JsonLd({ data }: { data: unknown }) {
  return <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(data) }} />;
}
```

**Recommendation**: Option B for consistency. Use in layout (6 schemas merged into `@graph` array still, one `<JsonLd>`) AND in `[slug]/page.tsx` for per-post Article+Breadcrumb. Keep Plan §8.4's `@graph` single-script pattern to minimize inline script count.

---

## 5. Existing Integration Risks

### 5.1 `public/llms.txt` coexistence with `llms-full.txt`

- `llms.txt` (slim, 64 lines) stays; follows llmstxt.org spec.
- `llms-full.txt` is auto-generated, full dump — a different file, different URL.
- Both served from `/public/` → `tene.sh/llms.txt` and `tene.sh/llms-full.txt`.
- **No conflict.** Update `public/llms.txt` to reference `llms-full.txt` at the top? (decision: low-value; skip)

### 5.2 `public/robots.txt` + `public/sitemap.xml` — MUST DELETE

Next 16 `app/robots.ts` and `app/sitemap.ts` output at `/robots.txt` and `/sitemap.xml`. If both static files AND route files exist:
- Next.js typically prefers `app/` routes (build warning).
- Risk: deploy pipeline copies `public/*` over built routes → stale content served.
- **Firm requirement at Do phase**: delete `apps/web/public/robots.txt` and `apps/web/public/sitemap.xml` in the same commit that adds `src/app/robots.ts` and `src/app/sitemap.ts`. Verify deletion in PR.

### 5.3 Existing JSON-LD (`layout.tsx:80-177`) — EXTEND, don't replace

Current `@graph` array has 3 entries. P0-9 adds 3 more:

```ts
// After modification, @graph contains 6 entries:
{ "@graph": [
  existing_SoftwareApplication,
  existing_FAQPage,
  existing_HowTo,
  new_Organization,     // with founder → Person ref
  new_Person,           // full, with sameAs array
  new_BreadcrumbList,   // Home only for root
]}
```

Keep existing indentation and keyword fields. Move full object construction to a const (e.g., `rootJsonLd`) to keep it readable at 100+ lines.

### 5.4 Tailwind v4 + `@tailwindcss/typography`

Tailwind v4 uses the new `@plugin` CSS directive (not the v3 `tailwind.config.js` plugins array). Integration:

```css
/* globals.css */
@import "tailwindcss";
@plugin "@tailwindcss/typography";

@theme inline { ... existing ... }

/* Brand overrides for prose-invert */
.prose-invert {
  --tw-prose-invert-links: var(--accent);
  --tw-prose-invert-bold: var(--foreground);
  --tw-prose-invert-code: var(--accent);
  --tw-prose-invert-quotes: var(--muted);
  /* etc. */
}
```

**Risk**: v4 plugin API may differ; spike at Do start if issues. Mitigation: keep override minimal, test on `welcome.mdx`.

### 5.5 CSP + GA4 / Vercel Analytics — REQUIRES ADJUSTMENT

Current CSP (`next.config.ts:8`):
```
default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self'; font-src 'self' data:
```

**Breaks**:
- GA4: `script-src` needs `https://www.googletagmanager.com`; `connect-src` needs `https://www.google-analytics.com https://*.analytics.google.com https://*.g.doubleclick.net`; `img-src` needs `https://www.google-analytics.com`.
- Vercel Analytics: `connect-src` needs `https://vitals.vercel-insights.com` (or `va.vercel-scripts.com` for the pixel).
- Vercel Speed Insights: same endpoint pattern.

**Action at Do**: Widen CSP in `next.config.ts:8` to:
```
script-src 'self' 'unsafe-inline' 'unsafe-eval' https://www.googletagmanager.com https://va.vercel-scripts.com;
connect-src 'self' https://www.google-analytics.com https://*.analytics.google.com https://*.google-analytics.com https://vitals.vercel-insights.com;
img-src 'self' data: https:;
```
Verify with `next dev` + browser devtools (Console).

### 5.6 Fonts (Geist Sans/Mono via `next/font/google`) — NO CHANGE

`next/font` self-hosts fonts; CSP `font-src 'self' data:` already covers. Blog typography uses same brand fonts. **No change needed.**

### 5.7 `metadataBase` already set — P0 blog MUST reuse

`layout.tsx:39`: `metadataBase: new URL("https://tene.sh")`. Blog `[slug]/page.tsx` `generateMetadata` should use relative `alternates.canonical: /blog/${slug}` — Next 16 resolves against parent `metadataBase`. (Or use absolute URL per Plan §8.5.)

### 5.8 `src/app/layout.tsx` render order (line 185-201)

Current body:
```tsx
<body className="min-h-full flex flex-col">
  {children}
  <NoiseOverlay />
</body>
```

After mods:
```tsx
<body className="min-h-full flex flex-col">
  {children}
  <NoiseOverlay />
  <CookieBanner />
  <Analytics />
  <SpeedInsights />
  {gaId && <GoogleAnalytics gaId={gaId} />}
</body>
```

Order: CookieBanner after content (z-index:50 positioning puts it at bottom-right regardless); Analytics components are scripts, placement inside body is fine and matches Next recommendation.

---

## 6. Session Plan — Module Split for `/pdca design`

Recommend 4 modules. Each is a shippable increment. Module 2 MUST precede Module 4 (sitemap dependency).

### Module 1 — Analytics Foundation (P0-1, P0-2, P0-3, P0-4, P0-14)

- **Files touched**: 5 (layout.tsx, new cookie-banner.tsx, new analytics.ts lib, copy-command.tsx, hero.tsx, cta.tsx)
- **Deps added**: `@vercel/analytics`, `@vercel/speed-insights`, `@next/third-parties`
- **External**: User creates GA4 property + Vercel env var (D-1, D-2)
- **Risks**: R-2 (banner conversion drop), R-7 (pre-consent event leak), CSP changes (§5.5)
- **Success gate**: GA4 DebugView shows events post-consent; Vercel dashboard shows pageviews within 24h
- **Est effort**: 3-4 hours

### Module 2 — SEO Infrastructure (P0-5, P0-6, P0-7, P0-8, P0-11, P0-12)

- **Files touched**: 3 new (sitemap.ts, robots.ts, .well-known/security.txt) + 2 deleted + 1 modified (package.json)
- **Deps added**: none
- **External**: User DNS TXT, GSC UI submission (D-3, D-4, D-6)
- **Risks**: §5.2 static-file coexistence
- **Dependency**: Blog routes (Module 4) will be picked up by sitemap.ts once they exist — sitemap.ts MUST import `getAllPostSlugs` from `@/lib/blog` which is created in Module 4. **Decision**: Ship sitemap.ts in Module 2 with ONLY static routes + `/blog` placeholder; extend in Module 4. OR ship Module 4's `lib/blog.ts` first as a stub returning `[]`. Recommend latter.
- **Success gate**: `/sitemap.xml` and `/robots.txt` serve dynamically; GSC verification passes
- **Est effort**: 2 hours (+ GSC waiting for propagation)

### Module 3 — AEO Enhancements (P0-9, P0-10, P0-13)

- **Files touched**: layout.tsx (JSON-LD extend), new gen-llms-full.ts, data/hero.ts, data/faq.ts
- **Deps added**: `tsx` (dev)
- **External**: User approves 4 answer-capsule rewrites (Checkpoint 3 of Design)
- **Risks**: Rewrite copy might diverge from brand voice — user approval gate required
- **Success gate**: 6 JSON-LD schemas validate in Rich Results Test; llms-full.txt 200
- **Est effort**: 2-3 hours

### Module 4 — Blog Rail (P0-15, P0-16, P0-17, P0-18)

- **Files touched**: 5 new app routes/pages + 1 lib + 2 content posts + 1 nav mod + 1 CSS mod
- **Deps added**: `next-mdx-remote`, `gray-matter`, `shiki`, `@tailwindcss/typography`
- **External**: User approves launch post content (D-5)
- **Risks**: R-5 MDX/RSC stability, R-6 typography collision, R-9 RSS XML-escaping
- **Dependency**: Module 2 sitemap.ts picks up blog posts automatically once `lib/blog.ts` exists
- **Success gate**: `/blog` renders, `/blog/introducing-tene` 200 + Article schema, `/feed.xml` validates on feed validator
- **Est effort**: 8-10 hours (most goes to writing the 1,500-word post)

### Total Estimated Effort
**P0**: ~17 hours of focused dev time (excluding waiting on GSC indexing + user content approval). Matches Plan §4's "3-5 days of focused work" estimate (Plan assumed parallel work + breaks).

### Recommended Execution Order
1. Module 1 (ship, verify analytics fires)
2. Module 2 (ship, verify GSC passes)
3. Module 4 (blog — sitemap auto-updates)
4. Module 3 (AEO polish — last because it depends on blog posts existing for capsule patterns)

### Alternative: Execute Module 2 + 4 in parallel
If user wants to split work across two branches: Module 2 (infra) and Module 4 (blog) are mostly independent. The `lib/blog.ts` stub trick lets Module 2 ship first without blocking on blog content.

---

## 7. Open Questions for User (BEFORE `/pdca design` Checkpoint 1)

Answers drive Design decisions. Present these at Design Checkpoint 1 (Requirements Confirmation) via AskUserQuestion.

### Q1 — GA4 Property
**Question**: Create a new GA4 property specifically for tene.sh, or reuse an existing one?
- A: New property named "Tene Landing"
- B: Existing property — if so, Measurement ID is?
- C: Defer — build code so `NEXT_PUBLIC_GA_ID` is optional; configure in next session

**Recommendation**: A (clean analytics, no cross-project noise).

### Q2 — DNS Provider for GSC TXT
**Question**: Where is tene.sh DNS managed?
- A: Vercel DNS (simplest — add TXT in Vercel dashboard)
- B: Cloudflare
- C: Other (specify)

### Q3 — PGP Key for security.txt
**Question**: For `security.txt Encryption:` field:
- A: Use existing PGP key (provide fingerprint + public key URL)
- B: Skip Encryption field (only Contact mailto: — valid per RFC 9116)
- C: Generate new PGP key for security@tene.sh

**Recommendation**: B for v1; add PGP later.

### Q4 — package.json Metadata
**Confirm**:
- `homepage: "https://tene.sh"` — OK?
- `repository: "https://github.com/agent-kay-it/tene"` — OK?
- `author: "Kay Kim <kay@popupstudio.ai>"` — OK?
- `description`: propose "Landing page for Tene — local-first encrypted secret manager"

### Q5 — Person Schema Author Identity
**Question**: For JSON-LD Person schema (drives AI citation attribution):
- `name`: **"Kay Kim"** (confirmed in PRD Constraint C-4)?
- `url`: `https://github.com/tomo-kay` (confirmed in PRD C-4)?
- `sameAs`: add LinkedIn/X/personal site URLs? If yes, provide.
- `jobTitle`: "Founder, Tene" — OK?

### Q6 — "Introducing tene" Post
**Question**: For the 1,500-word launch post (P0-16):
- A: Claude drafts based on PRD/Plan; user reviews + edits
- B: User writes; Claude only scaffolds MDX frontmatter + structure
- C: Co-draft in Design Checkpoint 3 (outline first, prose second)

**Recommendation**: C (saves time, keeps voice consistent).

### Q7 — Cookie Banner Default
**Confirm**: Default-deny (`analytics_storage: denied` until user clicks "Allow")?
- A: Yes (GDPR-safe; PRD US-3 acceptance; costs ~15-30% of GA data)
- B: Default-allow (more data; legally gray in EU)

**Recommendation**: A (PRD-aligned, dogfood privacy-first brand).

### Q8 — Answer Capsule Rewrites (§4.6)
**Question**: Accept the proposed hero sub + 3 FAQ rewrites in §4.6 as-is? Or revise?
- A: Accept all 4 as proposed
- B: Revise specific ones (indicate which)
- C: User writes new ones

### Q9 — ClawHub Event Instrumentation
**Question**: P0-14 calls for a `clawhub_link_click` event, but the current Nav + Footer have NO ClawHub link. Do we:
- A: Add a ClawHub link to the Nav / Footer (and track clicks on it)
- B: Drop `clawhub_link_click`; replace with another event (e.g., `product_hunt_click` — Hero already has PH badge)
- C: Defer to Module 1 Do-time decision

**Recommendation**: A — add a small ClawHub badge to Footer alongside GitHub/Issues links; instrument.

### Q10 — GitHub Event Naming Precision
**Question**: PRD lists `github-release-click` but all current GitHub links go to the repo homepage (not /releases). Do we:
- A: Change links to `.../releases` — stays true to event name
- B: Rename event to `github_click` (more accurate)
- C: Split: hero stays on repo (rename `github_repo_click`), CTA points to releases (`github_release_click`)

**Recommendation**: C (gives richest signal).

---

## Appendix — Quick Evidence Ledger

| Claim | Evidence |
|---|---|
| Next 16.2.2 actual version | `apps/web/package.json:12`, `apps/web/node_modules/next/package.json:2` |
| 3 JSON-LD schemas currently | `apps/web/src/app/layout.tsx:82-176` (SoftwareApplication, FAQPage with 6 Q&As, HowTo with 4 steps) |
| FAQ count = 8 items (not 6) | `apps/web/src/data/faq.ts:4-45` — Plan §1.1 said 6, actual is 8. Minor discrepancy; not blocking |
| Hero sub = 18 words | `apps/web/src/data/hero.ts:7` — too short for AEO capsule |
| Static robots.txt exists | `apps/web/public/robots.txt:1-5` — 5-line allow-all with sitemap |
| Static sitemap.xml exists | `apps/web/public/sitemap.xml:1-10` — single URL, lastmod 2026-04-06 |
| llms.txt exists | `apps/web/public/llms.txt:1-64` — comprehensive |
| No .env file, `NEXT_PUBLIC_GA_ID` not in vault yet | implied by PRD D-1/D-2 — user must action |
| Current CSP restricts analytics origins | `apps/web/next.config.ts:8` — only `'self'` in connect-src/script-src; MUST widen |
| `MetadataRoute.Sitemap` type matches Plan §8.2 | `apps/web/node_modules/next/dist/lib/metadata/types/metadata-interface.d.ts:562-576` |
| `MetadataRoute.Robots` type matches Plan §8.3 | same file, lines 547-561 |
| Next 16 async params breaking change | Official Next.js v16 upgrade guide — params/searchParams are Promise |
| `@next/third-parties` not installed | no node_modules entry |
| `next-mdx-remote/rsc` RSC support exists but unstable | package README; production-used |

**End of audit.**
