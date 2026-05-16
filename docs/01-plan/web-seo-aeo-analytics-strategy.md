# Plan+ — tene.sh SEO / AEO / Analytics Strategy

- **Feature**: `web-seo-aeo-analytics-strategy`
- **Branch**: `feature/web-landing-and-tech-blog` (from `staging`)
- **Date**: 2026-04-22
- **Method**: `/plan-plus` (Brainstorming-Enhanced PDCA Planning)
- **Target**: apps/web/ (Next.js 15 App Router, Vercel, tene.sh)

---

## Executive Summary

| Perspective | Summary |
|---|---|
| **Problem** | tene.sh has 68% of SEO basics in place (meta/OG/JSON-LD/llms.txt) but **0% analytics**, **0% Google Search Console verification**, no blog, no dynamic sitemap. Google's AI Overviews trigger on ~70% of B2B developer queries, reducing organic CTR by 46.7% — optimizing for *citations* is now more important than ranking for clicks. |
| **Solution** | Three-track rollout: (1) **Analytics** — Vercel Analytics + Speed Insights + GA4 via `@next/third-parties`; (2) **SEO+AEO** — GSC verification, dynamic sitemap/robots, expanded JSON-LD (Organization, Article, BreadcrumbList), FAQ expansion with 40-60 word answer capsules, llms-full.txt, awesome-list syndication; (3) **Blog** — MDX + App Router, topic-cluster strategy around "secret management for AI agents", RSS feed, 4 cornerstone posts in sprint 1. |
| **Function / UX Effect** | Users get unchanged landing UX. Developers get a technical blog surfacing why tene exists + comparisons + deep architecture. Search crawlers get rich structured data. AI answer engines (Perplexity/ChatGPT/Claude/Gemini) get citable content with explicit answer capsules. Team gets measurable traffic attribution from day one. |
| **Core Value** | **Be the source AIs cite when developers ask "how do I manage secrets for AI coding assistants?"** Convert CTR loss into citation share. Establish E-E-A-T signals (named author, security page, reproducible builds) that AI overviews favor. |

---

## 1. Current State Audit

Full inventory completed by subagent (Phase 0 exploration). Summary with file:line evidence:

### 1.1 What's Strong (keep + reinforce)

| Area | Status | Evidence |
|---|---|---|
| Root metadata | ✅ Complete | `src/app/layout.tsx:16-78` — title, description, 36 keywords, OG, Twitter, robots, metadataBase, canonical |
| JSON-LD structured data | ✅ 3 schemas | `layout.tsx:84-176` — `SoftwareApplication`, `FAQPage` (6 Q&A), `HowTo` (4 steps) |
| Static sitemap/robots | ✅ Present | `public/robots.txt`, `public/sitemap.xml` (single URL) |
| llms.txt | ✅ Present | `public/llms.txt:1-64` — comprehensive |
| Semantic HTML | ✅ Present | `<main>`, `<section>`, `<nav>`, `<footer>`, heading hierarchy correct |
| Font optimization | ✅ Present | `layout.tsx:2,6-14` — `next/font` (Geist, Geist_Mono) |
| Static rendering | ✅ Confirmed | `page.tsx` fully static, no SSR overhead |
| Security headers | ✅ Present | `next.config.ts:1-18` — HSTS, CSP, X-Frame-Options |

### 1.2 What's Missing (fill)

| Area | Status | Impact |
|---|---|---|
| Google Analytics / GTM | ❌ None | **Cannot measure anything** — no pageviews, conversions, CTA clicks |
| Google Search Console verification | ❌ None | No visibility into queries, impressions, CTR, indexing status |
| Dynamic sitemap.ts | ❌ None | Blog posts/future routes won't auto-appear |
| Dynamic robots.ts | ❌ None | Can't programmatically allow/block AI crawlers |
| Per-route metadata | ❌ None | Every page inherits same title/description |
| Dynamic OG images | ❌ None | Every share uses identical `/og-image.png` |
| `Organization` + `Person` schema | ❌ None | Missing E-E-A-T authority signals |
| `BreadcrumbList` schema | ❌ None | Reduces rich-result eligibility |
| `Article` schema infra | ❌ None | Blog posts won't qualify for Top Stories / AI Overviews article citations |
| Blog infrastructure | ❌ None | No `/blog` route, no MDX config, no RSS feed |
| `llms-full.txt` | ❌ None | Only slim `llms.txt`; Perplexity/Claude RAG ingestion prefers full dump |
| AI crawler explicit allowlist | ❌ Implicit | `robots.txt` says `Allow: /` — good, but doesn't enumerate GPTBot/ClaudeBot/Google-Extended/PerplexityBot explicitly |
| `security.txt` | ❌ None | `/.well-known/security.txt` expected for credible security tools |
| `Answer capsule` paragraphs | ⚠️ Weak | Hero + FAQ exist but not in the 40-60 word capsule format that gets cited by AI |
| Topic cluster content | ❌ None | No cornerstone "secret management for AI agents" pillar page |
| FAQ breadth | ⚠️ 6 items | Need 12-15 to cover long-tail AI queries |
| Analytics event tracking | ❌ None | No `install.sh` copy tracking, no GitHub release click tracking |
| Search Console sitemap submission | ❌ N/A | Prerequisite: verify domain first |
| Package.json homepage/repository | ⚠️ Missing | `apps/web/package.json` lacks `homepage` and `repository` fields |
| `awesome-*` list listings | ❌ None | Zero entries in awesome-cli-apps / awesome-go / awesome-security / awesome-opensource-alternatives |
| Homebrew tap | ❌ None | `.goreleaser.yml` has no tap config; high-value discovery signal |

### 1.3 Scorecard

- **SEO basics**: 68% covered
- **AEO basics**: 40% covered (llms.txt + schema present; answer capsules + llms-full.txt missing)
- **Analytics**: 0% covered
- **Blog**: 0% covered
- **Authority signals**: 30% covered (GitHub link + MIT license present; Organization/Person/security.txt missing)

---

## 2. 2026 Context (Non-Obvious Facts Driving the Strategy)

Extracted from research brief (see `docs/research/` if archived). Every strategic choice below is traceable to one of these:

1. **AI Overviews now appear on ~70% of B2B Technology queries** (Stackmatix 2026). For dev tools like tene, the search surface has fundamentally changed — ranking #1 is worth less if no one clicks.
2. **46.7% relative decline in organic CTR** when AIOs are present. Recovery path: become the *citation source*, not the click target.
3. **Pages with 3-4 complementary schemas are cited 2x more** by AI overviews than single-schema pages. tene currently has 3 — add `Organization`, `Person`, `Article` (blog), `BreadcrumbList` to reach 6-7.
4. **FAQPage schema: +22% median AI citation lift** (Relixir 50-domain study). tene has 6 Q&A — expanding to 12-15 with question-shaped headings (not noun phrases) is the single cheapest win.
5. **72.4% of ChatGPT-cited pages open every H2 with a 40-60 word answer capsule** (Frase). tene's hero/FAQ are correct in intent but too narrative — rewrite to capsule form.
6. **Original data + named tables earn 4.1x more citations**. We have an obvious candidate: publish "tene vs dotenv vs Doppler vs Infisical vs 1Password CLI" comparison as a data-rich post.
7. **llms.txt is not proven ROI** (8/9 sites saw no measurable traffic change per Search Engine Land). But low-cost to publish, so do it; don't over-invest. Google has publicly said it does not use llms.txt (Mueller).
8. **GSC indexing lag: 8-12 weeks for a new site to reach meaningful organic traffic**. Verify domain *today*, expect first real data late June 2026.
9. **RSS still matters in 2026** — not for humans but for AI crawlers, daily.dev, dev.to auto-import, Hashnode cross-post. Full-content `<content:encoded>` required.
10. **Claude Skills marketplaces (ClawHub/LobeHub/SkillHub) are the new "npm for AI"**. tene already shipped on ClawHub (v1.0.3 published 2026-04-22). Adding LobeHub + SkillHub listings is high-leverage.

---

## 3. Strategy — Three Tracks

### Track A — Analytics & Measurement

**Decision made (Phase 2)**: Vercel Analytics + Speed Insights + GA4 via `@next/third-parties`.

Rationale: Vercel suite gives real-user CWV + cookieless pageviews free; GA4 provides industry-standard metrics and auto-links to Google Search Console for query-level traffic attribution. Yes, we add a GDPR cookie banner — acceptable cost for GSC linkage.

| Component | Purpose | Cost |
|---|---|---|
| `@vercel/analytics` | Cookieless pageviews, referrers, top countries | Free |
| `@vercel/speed-insights` | Real-user LCP/INP/CLS/TTFB from prod | Free |
| `@next/third-parties` `GoogleAnalytics` | Industry-standard events, GSC linkage | Free |
| `NEXT_PUBLIC_GA_ID` env var | Managed via tene vault, not `.env` | — |
| Cookie consent banner | GDPR/ePrivacy compliance for GA4 | Build in-house (~60 LOC) |
| Event instrumentation | `sendGAEvent` on: install.sh copy, GitHub release click, ClawHub link, CTA clicks | — |

### Track B — SEO + AEO Foundation

| Initiative | What | Why |
|---|---|---|
| **B1. GSC verification** | Add DNS TXT OR `google-site-verification` meta in `layout.tsx` metadata | Unlocks query data, sitemap submission, URL Inspector |
| **B2. Dynamic `sitemap.ts`** | Replace static XML with `src/app/sitemap.ts`; auto-include blog slugs | Blog posts surface immediately |
| **B3. Dynamic `robots.ts`** | Explicit allowlist for GPTBot, ClaudeBot, Claude-User, Claude-SearchBot, Google-Extended, PerplexityBot, OAI-SearchBot, Amazonbot | AI crawlers honor robots.txt; explicit allow signals welcome |
| **B4. Expand JSON-LD** | Add `Organization`, `Person` (founder), `BreadcrumbList`, `SoftwareSourceCode` | Reach 6-7 schemas → 2x citation lift |
| **B5. Rewrite answer capsules** | First paragraph of every H2 becomes 40-60 word self-contained answer | +22% FAQ citation lift, 72% of ChatGPT citations had this |
| **B6. Expand FAQ to 12-15 items** | Question-shaped H2s matching actual developer prompts ("How do I stop my AI agent from reading .env files?", "Is tene compatible with Doppler?") | Long-tail AI citations |
| **B7. `llms-full.txt`** | Full markdown dump of all site content + SKILL.md + key docs | Perplexity/Claude RAG preference |
| **B8. `security.txt`** | `/.well-known/security.txt` with PGP key + disclosure policy | E-E-A-T for security tools |
| **B9. Dynamic OG images** | `app/opengraph-image.tsx` per-page variant with post title | Shareable distinct previews per blog post |
| **B10. `hreflang` self-ref** | Self-referencing canonical + `en-US` hreflang | Defensive; allows future i18n |

### Track C — Technical Blog (MDX + Topic Cluster)

| Component | Spec |
|---|---|
| **Routing** | `src/app/blog/page.tsx` (index), `src/app/blog/[slug]/page.tsx` (post), `generateStaticParams` for static export |
| **Content** | `content/posts/*.mdx` with frontmatter: `title`, `description`, `date`, `author`, `tags`, `canonical`, `ogImage` |
| **MDX pipeline** | `next-mdx-remote` or `contentlayer`; pick the leaner option (next-mdx-remote wins for simplicity) |
| **RSS/Atom** | `src/app/feed.xml/route.ts` — full `<content:encoded>`, not truncated |
| **Per-post OG** | Dynamic `app/blog/[slug]/opengraph-image.tsx` |
| **Schema per post** | `Article` + `BreadcrumbList` + `Person` (author) |
| **Cornerstone post** | `/guides/secret-management-for-ai-agents` — 3,000-5,000 words, data tables, embeds, diagrams |
| **Cluster posts** | 4 launch posts (see §5 Content Plan) |
| **Code highlighting** | `shiki` with Geist Mono, dark-only (matches brand) |
| **Syndication** | Canonical-first: publish → wait 14 days → cross-post to dev.to / Hashnode with `canonical_url` |
| **Typography** | `@tailwindcss/typography` with brand dark palette |

---

## 4. Prioritized Action Plan

### P0 — This Sprint (MVP, 3-5 days of focused work)

| # | Action | Est | Owner | Files |
|---|---|---|---|---|
| P0-1 | Add `@vercel/analytics` + `@vercel/speed-insights` to `layout.tsx` | 10m | — | `src/app/layout.tsx`, `package.json` |
| P0-2 | Install `@next/third-parties`, add `<GoogleAnalytics gaId={process.env.NEXT_PUBLIC_GA_ID!}>` | 15m | — | `layout.tsx`, `package.json`, `.env.example` |
| P0-3 | Build minimal cookie consent banner (`<CookieBanner>`) — block GA until consent | 45m | — | `src/components/cookie-banner.tsx` |
| P0-4 | Create GA4 property, put ID in tene vault (`tene set NEXT_PUBLIC_GA_ID ... --env prod`) | 15m | User | Vercel env vars |
| P0-5 | Verify tene.sh in GSC (DNS TXT via Vercel DNS or Cloudflare) | 10m | User | DNS config |
| P0-6 | Replace static sitemap.xml with `src/app/sitemap.ts` (returns array of routes) | 20m | — | delete `public/sitemap.xml`, add `src/app/sitemap.ts` |
| P0-7 | Replace static robots.txt with `src/app/robots.ts` (explicit AI bot allowlist) | 15m | — | delete `public/robots.txt`, add `src/app/robots.ts` |
| P0-8 | Submit sitemap to GSC; request-index homepage + `/install` | 5m | User | GSC UI |
| P0-9 | Add `Organization` + `Person` + `BreadcrumbList` schemas to `layout.tsx` JSON-LD | 30m | — | `src/app/layout.tsx` |
| P0-10 | Generate `public/llms-full.txt` (full site content + SKILL.md merged) | 30m | — | script: `scripts/gen-llms-full.ts` |
| P0-11 | Add `public/.well-known/security.txt` | 10m | — | new file |
| P0-12 | Add `homepage` + `repository` to `apps/web/package.json` | 2m | — | `apps/web/package.json` |
| P0-13 | Rewrite hero + 3 FAQ items as 40-60 word answer capsules | 45m | — | `src/components/{hero,faq}.tsx` |
| P0-14 | Instrument key events: install-script-copy, github-release-click, clawhub-link-click, cta-click | 45m | — | `<InstallCode>`, `<CTA>`, nav links |
| P0-15 | Scaffold `src/app/blog/` routes (index + `[slug]`), MDX pipeline, 1 placeholder post | 90m | — | `src/app/blog/**`, `content/posts/welcome.mdx`, `next.config.ts` |
| P0-16 | First real post published: **"Introducing tene — Local-First Encrypted Secrets for AI Agents"** | 3h | — | `content/posts/introducing-tene.mdx` |
| P0-17 | RSS feed route at `/feed.xml` | 45m | — | `src/app/feed.xml/route.ts` |
| P0-18 | Add `Article` + `BreadcrumbList` schema to `[slug]/page.tsx` | 30m | — | — |

**P0 definition of done**: GA4 firing in prod, GSC verified + sitemap submitted, 1 blog post live with RSS, Organization/Person schemas added, llms-full.txt generated.

### P1 — Next Sprint (expansion, 3-5 days)

| # | Action | Est | Notes |
|---|---|---|---|
| P1-1 | 3 more cornerstone blog posts (§5 content plan) | 2d | See content plan |
| P1-2 | Dynamic OG images per blog post (`app/blog/[slug]/opengraph-image.tsx`) | 90m | Use `next/og` `ImageResponse` |
| P1-3 | Comparison page: `/compare` — table vs dotenv/Doppler/Infisical/1Password CLI + schema | 3h | Data-rich = citable |
| P1-4 | Expand FAQ to 15 items with question-shaped H2s | 90m | Prompts to target AI overviews |
| P1-5 | Submit to awesome-lists (PRs): `awesome-cli-apps`, `awesome-go`, `awesome-security`, `awesome-selfhosted`, `awesome-opensource-alternatives`, `awesome-ai-tools` | 2h | Concurrent PR batch |
| P1-6 | Additional Skill marketplace listings: LobeHub, SkillHub.club, claudeskills.info | 1h | Mirror of ClawHub |
| P1-7 | Cornerstone pillar page: `/guides/secret-management-for-ai-agents` (3-5k words) | 1d | Topic cluster anchor |
| P1-8 | Syndication: cross-post introductory blog post to dev.to + Hashnode with `canonical_url` (wait 14 days) | 30m | After P0-16 is 14d old |
| P1-9 | Hacker News Show HN submission for tene (timed Tue 8-10am PT) | 15m | One-shot |
| P1-10 | Homebrew tap: `homebrew-tap` repo with `Formula/tene.rb`; update `.goreleaser.yml` | 4h | Unlocks `brew install tomo-kay/tap/tene` |
| P1-11 | Signed commits + SLSA provenance on GoReleaser | 2h | Authority signal |
| P1-12 | Author bio page: `/about` with founder photo, GitHub link, security PGP key | 90m | E-E-A-T |

### P2 — Following Sprint (polish + growth, 3-5 days)

| # | Action | Est | Notes |
|---|---|---|---|
| P2-1 | Add 4 more blog posts completing topic cluster (§5) | 3d | |
| P2-2 | "State of AI Secret Management 2026" — original research post with developer survey | 5d | Highest citation bait |
| P2-3 | Reddit seeding (r/commandline, r/golang, r/selfhosted, r/devops) | 2h | Spread across 2 weeks to avoid spam flag |
| P2-4 | Product Hunt launch (second wave, if HN went well) | 4h | Tuesday 00:01 PT |
| P2-5 | Topic-cluster interlinking pass across blog + landing sections | 2h | |
| P2-6 | Performance pass: replace raw `<img>` (Product Hunt badge) with `next/image` where feasible | 45m | |
| P2-7 | Add `Review` schema once we have 3+ public testimonials | 30m | |
| P2-8 | Scoop, AUR, asdf, mise package index listings | 3h | Long-tail discovery |
| P2-9 | Vercel → GA4 → GSC attribution funnel dashboard (internal) | 2h | Measure what moves |
| P2-10 | Monthly GSC query report automation (GitHub Action) | 2h | Trend tracking |

---

## 5. Content Plan — Launch Posts (topic cluster)

### Cornerstone (pillar page — P1)
**`/guides/secret-management-for-ai-agents`** (3,000-5,000 words)
Anchor page for the topic cluster. Sections:
- What changed when AI agents started reading .env files (the problem)
- Threat model: AI context, prompt injection, conversation logs, training data
- What a "local-first secret manager" actually means
- Architecture patterns: keychain-backed, zero-knowledge sync, recovery mnemonics
- Comparison matrix (links to `/compare`)
- How to migrate (3 scenarios: solo dev, team, CI/CD)
- FAQ
- 5+ internal links to cluster posts

### Cluster Posts (P0 + P1)

| # | Title | Length | Sprint | Cluster link |
|---|---|---|---|---|
| 1 | **Introducing tene** — Local-First Encrypted Secrets for AI Agents | 1,500 | P0-16 | Launch announcement + core value prop |
| 2 | **Why `.env` files break AI workflows** (and what to do about it) | 2,000 | P1 | The problem post — linkable by others |
| 3 | **tene vs dotenv vs Doppler vs Infisical vs 1Password CLI** — a data-rich comparison | 2,500 | P1 | The comparison post — citation magnet |
| 4 | **Zero-Knowledge Secret Sync in a Go CLI**: XChaCha20-Poly1305 + Argon2id explained | 2,000 | P1 | Technical depth + E-E-A-T |
| 5 | **BIP-39 for Secrets**: why we borrowed a Bitcoin convention for master password recovery | 1,200 | P2 | Unique angle, highly shareable |
| 6 | **Claude Code + tene**: the setup we wish we had at the start | 1,500 | P2 | Practical tutorial |
| 7 | **State of AI Secret Management 2026** — developer survey + recommendations | 3,000 | P2 | Original data = 4.1x citations |
| 8 | **Publishing our first ClawHub skill**: what we learned | 1,200 | P2 | Community angle + ClawHub cross-link |

Each post:
- Frontmatter with full metadata (title, description, date, author, tags, og image)
- First paragraph is a 40-60 word answer capsule
- Every H2 begins with a 40-60 word capsule
- At least one data table or code block per 500 words
- Internal links to 2-3 other posts
- External links to authoritative sources (RFC, Go docs, Anthropic docs)
- Closing FAQ (3-5 Q&A) → `FAQPage` schema

---

## 6. YAGNI Review (Phase 3)

What we are **deliberately NOT** doing in v1.0:

| Excluded | Why |
|---|---|
| i18n / multilingual content | English-only for 6 months; hreflang self-ref only |
| A/B testing framework | PostHog declined; too early to have meaningful variants |
| Full CMS (Sanity/Contentful) | MDX in-repo is simpler, versioned with code, no external dependency |
| Newsletter platform (ConvertKit/Buttondown) | Waitlist form (Formspree) is enough until 500+ subscribers |
| Session replay (Hotjar/Clarity/FullStory) | Privacy concerns for a security tool's landing page |
| Community forum | Too early; GitHub Discussions is sufficient |
| `Review` schema | No public testimonials yet |
| Sponsored content / ads | Zero budget, distorts early signal |
| Video content (YouTube) | VHS demos on README are enough for now |
| Localized Korean/Japanese landing | Defer until traction in one language |
| `HowTo` schemas per feature | Current single HowTo on landing covers the install flow; more is diminishing return |
| Edge-cached personalized content | Landing is static; personalization adds risk without win |

---

## 7. Success Metrics (Track, report monthly)

### Short-term (4 weeks post-rollout)
- GSC verified + ≥1 query shown in Search Console
- GA4 firing: ≥95% of sessions captured
- Vercel Speed Insights p75: LCP ≤ 2.5s, INP ≤ 200ms, CLS ≤ 0.1
- ≥1 blog post indexed by Google
- llms-full.txt served at 200
- `awesome-*` list PR merged on at least 1 repo

### Mid-term (12 weeks)
- ≥500 monthly organic visitors
- ≥50 install-script copy events
- ≥20 GitHub release clicks from site
- ≥3 AI citation hits (test via Perplexity / ChatGPT search: "how to manage secrets for claude code")
- ≥4 blog posts published, topic cluster interlinked
- Hacker News submission ≥50 points OR ≥1 front-page hit
- ≥1 post syndicated to dev.to (canonical intact)

### Long-term (6 months)
- ≥5,000 monthly organic visitors
- Top-3 ranking for "tene secret manager" and ≥5 long-tail phrases
- ≥10 AI citations per month
- Homebrew tap live + ≥100 `brew install` events
- ClawHub + LobeHub + SkillHub all listed, ≥500 combined downloads
- Named in ≥3 external developer-tool roundups

---

## 8. Implementation — Concrete Next.js 15 Snippets

### 8.1 Vercel Analytics + Speed Insights + GA4 — `src/app/layout.tsx`

```tsx
import { Analytics } from '@vercel/analytics/next';
import { SpeedInsights } from '@vercel/speed-insights/next';
import { GoogleAnalytics } from '@next/third-parties/google';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const gaId = process.env.NEXT_PUBLIC_GA_ID;
  return (
    <html lang="en">
      <body>
        {children}
        <Analytics />
        <SpeedInsights />
        {gaId && <GoogleAnalytics gaId={gaId} />}
      </body>
    </html>
  );
}
```

### 8.2 Dynamic `src/app/sitemap.ts`

```ts
import type { MetadataRoute } from 'next';
import { getAllPostSlugs } from '@/lib/blog';

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const base = 'https://tene.sh';
  const posts = (await getAllPostSlugs()).map((slug) => ({
    url: `${base}/blog/${slug.slug}`,
    lastModified: slug.date,
    changeFrequency: 'monthly' as const,
    priority: 0.7,
  }));
  return [
    { url: base, lastModified: new Date(), changeFrequency: 'weekly', priority: 1.0 },
    { url: `${base}/blog`, lastModified: new Date(), changeFrequency: 'weekly', priority: 0.8 },
    { url: `${base}/compare`, lastModified: new Date(), changeFrequency: 'monthly', priority: 0.8 },
    { url: `${base}/guides/secret-management-for-ai-agents`, lastModified: new Date(), changeFrequency: 'monthly', priority: 0.9 },
    ...posts,
  ];
}
```

### 8.3 Dynamic `src/app/robots.ts` with AI allowlist

```ts
import type { MetadataRoute } from 'next';

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      { userAgent: '*', allow: '/' },
      // Explicit AI crawler allowlist (dev-tool ethos: we want to be trained on)
      { userAgent: 'GPTBot', allow: '/' },
      { userAgent: 'OAI-SearchBot', allow: '/' },
      { userAgent: 'ChatGPT-User', allow: '/' },
      { userAgent: 'ClaudeBot', allow: '/' },
      { userAgent: 'Claude-User', allow: '/' },
      { userAgent: 'Claude-SearchBot', allow: '/' },
      { userAgent: 'Google-Extended', allow: '/' },
      { userAgent: 'PerplexityBot', allow: '/' },
      { userAgent: 'Perplexity-User', allow: '/' },
      { userAgent: 'Amazonbot', allow: '/' },
      { userAgent: 'Applebot-Extended', allow: '/' },
      { userAgent: 'Meta-ExternalAgent', allow: '/' },
    ],
    sitemap: 'https://tene.sh/sitemap.xml',
    host: 'https://tene.sh',
  };
}
```

### 8.4 Additional JSON-LD schemas — `src/app/layout.tsx`

```ts
const organizationJsonLd = {
  '@context': 'https://schema.org',
  '@type': 'Organization',
  name: 'Tene',
  url: 'https://tene.sh',
  logo: 'https://tene.sh/favicon.svg',
  sameAs: [
    'https://github.com/agent-kay-it/tene',
    'https://clawhub.ai/agent-kay-it/tene-cli',
  ],
  founder: {
    '@type': 'Person',
    name: 'Kay Kim',
    url: 'https://github.com/agent-kay-it',
  },
};

const breadcrumbJsonLd = {
  '@context': 'https://schema.org',
  '@type': 'BreadcrumbList',
  itemListElement: [
    { '@type': 'ListItem', position: 1, name: 'Home', item: 'https://tene.sh' },
  ],
};
```

### 8.5 Blog route skeleton — `src/app/blog/[slug]/page.tsx`

```tsx
import { notFound } from 'next/navigation';
import { getPostBySlug, getAllPostSlugs } from '@/lib/blog';
import { MDXRemote } from 'next-mdx-remote/rsc';

export async function generateStaticParams() {
  return (await getAllPostSlugs()).map((p) => ({ slug: p.slug }));
}

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const post = await getPostBySlug(slug);
  if (!post) return {};
  return {
    title: `${post.title} — tene blog`,
    description: post.description,
    openGraph: { title: post.title, description: post.description, type: 'article' },
    alternates: { canonical: `https://tene.sh/blog/${slug}` },
  };
}

export default async function Post({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const post = await getPostBySlug(slug);
  if (!post) notFound();

  const articleJsonLd = {
    '@context': 'https://schema.org',
    '@type': 'Article',
    headline: post.title,
    description: post.description,
    datePublished: post.date,
    author: { '@type': 'Person', name: post.author, url: 'https://github.com/agent-kay-it' },
    publisher: { '@type': 'Organization', name: 'Tene', logo: { '@type': 'ImageObject', url: 'https://tene.sh/favicon.svg' } },
  };

  return (
    <article className="prose prose-invert mx-auto max-w-3xl px-4 py-16">
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(articleJsonLd) }} />
      <h1>{post.title}</h1>
      <MDXRemote source={post.content} />
    </article>
  );
}
```

### 8.6 RSS feed — `src/app/feed.xml/route.ts`

```ts
import { getAllPosts } from '@/lib/blog';

export async function GET() {
  const posts = await getAllPosts();
  const xml = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:content="http://purl.org/rss/1.0/modules/content/">
  <channel>
    <title>tene blog</title>
    <link>https://tene.sh/blog</link>
    <description>Local-first encrypted secrets for AI-native projects.</description>
    <language>en</language>
    ${posts.map((p) => `
    <item>
      <title><![CDATA[${p.title}]]></title>
      <link>https://tene.sh/blog/${p.slug}</link>
      <guid>https://tene.sh/blog/${p.slug}</guid>
      <pubDate>${new Date(p.date).toUTCString()}</pubDate>
      <description><![CDATA[${p.description}]]></description>
      <content:encoded><![CDATA[${p.content}]]></content:encoded>
    </item>`).join('')}
  </channel>
</rss>`;
  return new Response(xml, { headers: { 'Content-Type': 'application/rss+xml' } });
}
```

### 8.7 Cookie consent banner (minimal, GA4-aware)

```tsx
'use client';
import { useEffect, useState } from 'react';

export function CookieBanner() {
  const [visible, setVisible] = useState(false);
  useEffect(() => {
    if (typeof window === 'undefined') return;
    if (localStorage.getItem('tene-consent') === null) setVisible(true);
  }, []);
  if (!visible) return null;
  const decide = (allow: boolean) => {
    localStorage.setItem('tene-consent', allow ? 'granted' : 'denied');
    // Use GA4's Consent Mode v2
    window.gtag?.('consent', 'update', {
      analytics_storage: allow ? 'granted' : 'denied',
      ad_storage: 'denied',
    });
    setVisible(false);
  };
  return (
    <div className="fixed bottom-4 left-1/2 -translate-x-1/2 z-50 max-w-lg rounded-lg border border-[#2a2a2a] bg-[#141414] p-4 text-sm">
      <p className="mb-3 text-[#ededed]">
        We use analytics to understand how tene is used. No tracking without consent.
      </p>
      <div className="flex gap-2">
        <button onClick={() => decide(true)} className="rounded bg-[#00ff88] px-3 py-1 text-black">Allow</button>
        <button onClick={() => decide(false)} className="rounded border border-[#2a2a2a] px-3 py-1 text-[#888]">Deny</button>
      </div>
    </div>
  );
}
```

---

## 9. Risks & Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| GSC indexing takes 8-12 weeks | No organic traffic for 2 months | Launch HN/PH/syndication for referral traffic in meantime |
| Cookie banner hurts conversion | 10-20% drop possible | Minimal banner, default to "deny" (GA4 loses data but privacy preserved); measure impact |
| GA4 blocked by privacy extensions | Under-reports ~15-30% | Vercel Analytics (cookieless) as source of truth |
| Cross-posting triggers duplicate-content penalty | Rankings tank | `canonical_url` in dev.to/Hashnode frontmatter, 14-day lag |
| HN submission bombs | Brand damage if dogpiled | Craft submission title carefully, pick show-HN pattern, don't over-submit |
| Blog content quality dips | Loses citation value | Enforce 40-60 capsule rule + table-per-500-words in review checklist |
| Analytics over-tracking = ethics concern for security tool | Reputation risk | Document exactly what we track in `/privacy`; opt-out default for GA |
| llms.txt ROI unclear | Wasted effort | Automate generation; ~30min upfront, ~0 ongoing maintenance |
| AIO CTR drop continues | Traffic growth stalls | Double down on citation bait (data posts, comparisons) |
| Homebrew tap rejection | Delayed install channel | Ship as `tomo-kay/tap/tene` (personal tap) first; apply to homebrew-core after 1k users |
| SEO crawler infinite loops on dynamic routes | Crawl budget waste | All routes static-exported; no dynamic params except blog slugs |

---

## 10. Rollout Schedule

| Week | Milestone |
|---|---|
| **W1** | P0-1..P0-14 (analytics + GSC + schemas + capsules) |
| **W1** | P0-15..P0-18 (blog infra + post 1 + RSS) |
| **W2** | Verify prod GA4 fire, submit sitemap, check Speed Insights |
| **W2** | P1-1 first cornerstone + P1-3 comparison page |
| **W3** | P1-2 dynamic OG, P1-4 FAQ expansion, P1-5 awesome-list PRs |
| **W3** | P1-9 HN Show HN submission (Tuesday 8-10am PT) |
| **W4** | P1-6..P1-8 marketplace listings + syndication |
| **W4** | P1-10..P1-12 Homebrew + SLSA + /about page |
| **W5-6** | P2 posts (cluster completion) |
| **W7** | P2 survey post launches + Product Hunt wave 2 |
| **W8** | Review metrics, adjust tactics |

---

## 11. Decision Log (Plan+ Phase 2-3)

- **Analytics stack**: Vercel + GA4 (user decision, 2026-04-22). Rationale: GSC linkage + industry-standard metrics. Accepted cost: cookie banner.
- **Scope**: Full P0+P1+P2 (user decision, 2026-04-22). Rationale: speed to market for awareness; all tiers ship within 8 weeks.
- **Blog stack**: MDX + `next-mdx-remote` + App Router. Not contentlayer (abandoned), not Velite (extra dependency).
- **RSS**: full-content, not summary. For daily.dev + dev.to auto-import.
- **llms-full.txt**: automated generation from markdown sources; no manual curation.
- **Cookie consent**: Custom minimal banner, default-deny. Not `cookiebot`/`OneTrust` (overkill).
- **AI crawler policy**: allow all. Dev tool wants to be trained on.
- **i18n**: deferred. English only for ≥6 months.
- **Session replay**: rejected. Privacy-hostile for a security tool's landing.

---

## 12. Next Steps

1. **Review this plan** → user approval gate (HARD-GATE per plan-plus).
2. **On approval**: proceed to `/pdca design web-seo-aeo-analytics-strategy` (Design doc with file structure, dependencies, code interfaces).
3. **Then**: `/pdca do` (implementation starting with P0 items in priority order).
4. **Then**: `/pdca analyze` (gap check against this plan), `/pdca report` (PDCA report with metrics).

---

## Appendix A — References (from research brief)

Key citations driving strategic decisions:

1. [Google AI Overview SEO Impact](https://www.stackmatix.com/blog/google-ai-overview-seo-impact) — 46.7% CTR decline, 70% B2B tech trigger rate
2. [web.dev Core Web Vitals](https://web.dev/articles/vitals) — LCP/INP/CLS thresholds
3. [llmstxt.org](https://llmstxt.org/) — llms.txt spec
4. [Search Engine Land llms.txt analysis](https://searchengineland.com/llms-txt-proposed-standard-453676) — ROI reality check
5. [Mintlify real llms.txt examples](https://www.mintlify.com/blog/real-llms-txt-examples) — Anthropic, Cloudflare, Vercel
6. [Anthropic ClaudeBot split](https://www.searchenginejournal.com/anthropics-claude-bots-make-robots-txt-decisions-more-granular/568253/) — 3-bot naming
7. [Frase AEO guide](https://www.frase.io/blog/what-is-answer-engine-optimization-the-complete-guide-to-getting-cited-by-ai) — 40-60 word capsule, 72.4% citation pattern
8. [Stackmatix structured data](https://www.stackmatix.com/blog/structured-data-ai-search) — schema types & lift data, FAQPage +22%
9. [Next.js third-parties docs](https://nextjs.org/docs/app/guides/third-party-libraries) — GA4 integration
10. [Vercel Speed Insights](https://vercel.com/docs/speed-insights) — real-user CWV
11. [Vercel Web Analytics](https://vercel.com/docs/analytics) — cookieless pageviews
12. [Google Canonical docs](https://developers.google.com/search/docs/crawling-indexing/consolidate-duplicate-urls) — canonical best practices
13. [Google FAQPage schema](https://developers.google.com/search/docs/appearance/structured-data/faqpage) — spec
14. [Draft.dev syndication](https://draft.dev/learn/syndicating-developer-content) — canonical-first cross-posting
15. [Digitalapplied topic clusters](https://www.digitalapplied.com/blog/seo-content-clusters-2026-topic-authority-guide) — cluster ROI data
16. [OpenAIToolsHub marketplace comparison](https://www.openaitoolshub.org/en/blog/claude-skills-marketplace-comparison) — Skill marketplace landscape

---

**End of plan.** Awaiting review + approval before `/pdca design`.
