# PRD — tene.sh SEO / AEO / Analytics (Track 1 only)

- **Feature**: `web-seo-aeo-analytics-strategy`
- **Scope**: **Track 1 = P0 (One-time Setup + Blog Rail) only**
- **Branch**: `feature/web-landing-and-tech-blog`
- **Plan reference**: `docs/01-plan/web-seo-aeo-analytics-strategy.md` (595 lines, comprehensive)
- **Date**: 2026-04-22
- **Purpose of this PRD**: Crystallize Track 1 requirements so `/pdca team` (codebase deep-analysis) and `/pdca design` phases have a precise requirements baseline. The Plan doc already covers market analysis, competitor positioning, content strategy, and P1/P2 roadmap — do NOT re-do those.

---

## Executive Summary

| Perspective | Summary (Track 1 scope) |
|---|---|
| **Problem** | tene.sh has 68% of SEO basics but 0% analytics, 0% Google Search Console verification, no blog route, no dynamic sitemap/robots, and no explicit AI-crawler allowlist. We cannot measure anything, cannot be indexed by GSC, and have zero inventory for AI citation. |
| **Solution** | Ship 18 concrete P0 tasks across three groups: (A) Analytics & Measurement — Vercel Analytics + Speed Insights + GA4 with cookie consent + event instrumentation; (B) SEO/AEO Foundation — dynamic sitemap/robots, expanded JSON-LD (Organization + Person + BreadcrumbList), llms-full.txt, security.txt, AEO answer-capsule rewrites; (C) Blog Rail — `/blog` route + MDX pipeline + first real post + RSS feed + Article schema. |
| **Function / UX Effect** | Landing UX unchanged. Visitors see a small cookie banner and a new `/blog` link. Search crawlers get a dynamic sitemap that auto-includes blog posts. AI crawlers (GPTBot, ClaudeBot, PerplexityBot, etc.) see explicit allow rules and `llms-full.txt`. Maintainers get real-time traffic + conversion funnel in Vercel + GA4. |
| **Core Value** | Turn tene.sh from an invisible static landing into a **measurable, indexable, AI-citable** asset with a working content pipeline — the minimum viable foundation required before any P1 growth work (comparison pages, more posts, syndication) can produce signal. |

---

## 1. Beachhead Segment

Track 1 does not target external customers directly. Its real "users" are three internal/infrastructural personas whose needs must be satisfied for Track 2+ (growth) to work.

| # | Segment | Who | Why they matter for P0 |
|---|---|---|---|
| 1 | **tene maintainers (Kay Kim + future team)** | Internal: need visibility into what traffic/behavior the site actually produces | Without GA4 + Vercel Analytics firing, every P1/P2 decision (which post to double down on, which channel converts) is a guess |
| 2 | **Search & AI crawlers** | Googlebot, Bingbot, GPTBot, ClaudeBot, Claude-User, Claude-SearchBot, Google-Extended, PerplexityBot, OAI-SearchBot, ChatGPT-User, Amazonbot, Applebot-Extended, Meta-ExternalAgent, Perplexity-User | Cannot cite what they can't find. Need dynamic sitemap, robots allowlist, Article schema, llms-full.txt |
| 3 | **First 100 developer visitors (early adopters)** | Devs arriving from GitHub README, install.sh command, or direct share. Technical, skeptical, read docs. | Need a blog to read (validates tene is a real project not abandoned), a launch post explaining the *why*, and an RSS feed they can subscribe to |

**Primary beachhead for P0 success measurement**: Segment 1 (maintainers) — if analytics fires in ≥95% of sessions and GSC shows query data within 8-12 weeks, P0 is successful regardless of visitor count.

---

## 2. JTBD (Jobs To Be Done)

Track 1 solves 4 discrete jobs. Each maps to a subset of the 18 P0 tasks.

### JTBD-1 — "When I ship tene.sh, I want to know what's happening on it, so I can make data-driven growth decisions in Track 2+"
- **Job executor**: Maintainer
- **Tasks solving it**: P0-1, P0-2, P0-3, P0-4, P0-14
- **Success signal**: `window.dataLayer` populated, GA4 real-time shows self-visit, events fire on install-copy / release-click / ClawHub-click / CTA-click

### JTBD-2 — "When a Google or AI crawler arrives at tene.sh, I want it to discover every page I care about and know it has permission to ingest, so tene becomes citable"
- **Job executor**: Search/AI crawler
- **Tasks solving it**: P0-5, P0-6, P0-7, P0-8, P0-10, P0-11
- **Success signal**: GSC verified, sitemap submitted and fetched (200), `robots.txt` served from `src/app/robots.ts` with explicit allow rules, `llms-full.txt` served at 200, security.txt served at 200

### JTBD-3 — "When an AI answer engine decides what to cite for 'secret management for AI agents', I want tene's content to be the most structured, authoritative source it sees"
- **Job executor**: AI answer engine (Perplexity, ChatGPT Search, Claude, Gemini)
- **Tasks solving it**: P0-9, P0-13, P0-18
- **Success signal**: 5+ JSON-LD schemas on homepage (existing SoftwareApplication/FAQPage/HowTo + new Organization/Person/BreadcrumbList); hero and top-3 FAQ items rewritten as 40-60 word answer capsules; blog posts carry Article + BreadcrumbList

### JTBD-4 — "When a developer lands on tene.sh and wants to verify this is a serious project, I want them to find a real, dated, authored article explaining why tene exists"
- **Job executor**: First-visit developer
- **Tasks solving it**: P0-15, P0-16, P0-17, P0-18, P0-12
- **Success signal**: `/blog` index loads, "Introducing tene" post renders with Article schema, `/feed.xml` returns valid RSS 2.0 with full content, `apps/web/package.json` declares homepage + repository

---

## 3. User Stories (Track 1 scope, 14 stories)

INVEST check: each story is independent, negotiable, valuable, estimable, small (≤1 day), testable.

### Maintainer stories (analytics visibility)

**US-1** — As a maintainer, I want Vercel Analytics and Speed Insights running on every page so I can see pageviews, top referrers, and real-user CWV from day one.
- **Acceptance**: `@vercel/analytics` + `@vercel/speed-insights` installed; `<Analytics />` and `<SpeedInsights />` mounted in `src/app/layout.tsx`; Vercel dashboard shows data within 24h of first prod visit.

**US-2** — As a maintainer, I want GA4 configured via `@next/third-parties` so I can link Search Console queries to on-site behavior.
- **Acceptance**: `@next/third-parties` installed; `<GoogleAnalytics gaId={NEXT_PUBLIC_GA_ID} />` mounts only when env var is present; GA4 real-time report shows self-visit.

**US-3** — As a maintainer, I want GA4 loading blocked until the visitor consents so we remain GDPR/ePrivacy-compliant.
- **Acceptance**: Default-deny banner shows on first visit; GA4 script does not execute before "Accept"; consent stored in `localStorage`; GA4 Consent Mode v2 defaults set (`ad_storage=denied`, `analytics_storage=denied`).

**US-4** — As a maintainer, I want four key conversion events tracked (install-script-copy, github-release-click, clawhub-link-click, cta-click) so I can measure what actually moves users.
- **Acceptance**: `sendGAEvent` fires with a named event from the 4 trigger points; events visible in GA4 DebugView.

**US-5** — As a maintainer, I want `NEXT_PUBLIC_GA_ID` stored in the tene vault (not `.env` on disk) so we dogfood our own tool.
- **Acceptance**: `tene set NEXT_PUBLIC_GA_ID ... --env prod` done; Vercel env var set from Vercel dashboard (not committed); `.env.example` documents the var; no `.env` file committed.

### Crawler stories (indexing)

**US-6** — As the Googlebot, I want to fetch `tene.sh/sitemap.xml` and discover every page including future blog posts without a human re-editing an XML file.
- **Acceptance**: `src/app/sitemap.ts` exists and exports the dynamic sitemap; `public/sitemap.xml` deleted; sitemap includes `/`, `/blog`, and every published post slug; lastModified comes from post frontmatter.

**US-7** — As an AI crawler (GPTBot/ClaudeBot/PerplexityBot/etc.), I want an explicit allow rule so I can confirm permission rather than infer it from `Allow: /`.
- **Acceptance**: `src/app/robots.ts` exists and returns explicit rules for the 12 bots listed in Track 1 scope; `public/robots.txt` deleted; sitemap URL + host declared.

**US-8** — As Google Search Console, I want tene.sh domain-verified so I can surface query/impression/CTR data to the maintainer.
- **Acceptance**: DNS TXT record added (Vercel DNS or Cloudflare); GSC verification succeeds; sitemap submitted and in "Success" state; homepage and (if live) `/install` manually request-indexed.

**US-9** — As an AI RAG pipeline (Perplexity/Claude/ChatGPT), I want a single `/llms-full.txt` file containing the full site content plus tene SKILL.md so I can ingest authoritative content without scraping.
- **Acceptance**: `scripts/gen-llms-full.ts` runs at build time and writes `public/llms-full.txt`; output is a markdown concatenation of landing content + SKILL.md + key docs; served 200 at `https://tene.sh/llms-full.txt`.

**US-10** — As a security researcher, I want `/.well-known/security.txt` at tene.sh so I know how to responsibly disclose.
- **Acceptance**: `public/.well-known/security.txt` exists with `Contact:`, `Expires:`, `Preferred-Languages:` fields (RFC 9116); served 200.

### AI-citation stories

**US-11** — As an AI answer engine, I want rich Organization + Person + BreadcrumbList structured data so I can attribute citations to a named author and org with known authority.
- **Acceptance**: Three JSON-LD blocks added to `src/app/layout.tsx` on top of existing three (SoftwareApplication + FAQPage + HowTo), for 6 total; `Person.name = "Kay Kim"`, `Person.url = "https://github.com/agent-kay-it"`; validates in Google Rich Results Test.

**US-12** — As an AI answer engine, I want the first paragraph of the hero and the top-3 FAQ answers to be self-contained 40-60 word capsules so I can quote them cleanly.
- **Acceptance**: `src/components/hero.tsx` lede paragraph is 40-60 words and answers "what is tene?" completely; 3 FAQ entries in `src/components/faq.tsx` rewritten to the same capsule format.

### Visitor stories (blog consumption)

**US-13** — As a first-visit developer, I want to click `/blog` and see an index of posts plus a dated launch post explaining why tene exists, so I can verify this is a real, live project.
- **Acceptance**: `/blog` index lists posts sorted by date desc; "Introducing tene — Local-First Encrypted Secrets for AI Agents" (≥1,500 words) renders at `/blog/introducing-tene`; frontmatter complete (title, description, date, author, tags, canonical); MDX renders with `next-mdx-remote` + `gray-matter`; placeholder post exists in `content/posts/` to prove the index scales.

**US-14** — As a subscriber, I want `/feed.xml` to return a valid RSS 2.0 feed with full post content so I can read in my feed reader or pipe into daily.dev/dev.to.
- **Acceptance**: `src/app/feed.xml/route.ts` returns `Content-Type: application/rss+xml`; feed validates on `validator.w3.org/feed`; `<content:encoded>` contains full HTML (not truncated); `Article` + `BreadcrumbList` JSON-LD present on each post's HTML page.

---

## 4. Success Criteria (Measurable DoD for Track 1)

All 18 P0 tasks complete AND every one of these automated/manual checks passes:

| # | Criterion | How to verify |
|---|---|---|
| SC-1 | Vercel Analytics firing | Vercel dashboard shows ≥1 pageview within 1h of deploy |
| SC-2 | Speed Insights firing | Vercel Speed Insights shows p75 LCP/INP/CLS within 24h |
| SC-3 | GA4 firing ≥95% of sessions | GA4 Realtime matches Vercel Analytics pageviews ±5% over 48h sample |
| SC-4 | Cookie banner default-deny | First-visit network tab: no `google-analytics.com` requests until "Accept" |
| SC-5 | 4 events instrumented | GA4 DebugView shows all 4 custom events triggered from staged clicks |
| SC-6 | GSC domain-verified | GSC "Ownership verified" status on tene.sh property |
| SC-7 | Sitemap submitted + fetched | GSC Sitemaps panel shows status "Success" for `sitemap.xml` |
| SC-8 | `/sitemap.xml` dynamic | curl returns XML containing blog post URLs; static file deleted |
| SC-9 | `/robots.txt` dynamic with AI allowlist | curl returns rules for all 12 listed bots; static file deleted |
| SC-10 | 6 JSON-LD schemas on homepage | View source shows SoftwareApplication + FAQPage + HowTo + Organization + Person + BreadcrumbList; all pass Rich Results Test |
| SC-11 | `llms-full.txt` served 200 | `curl https://tene.sh/llms-full.txt` returns 200 and non-empty markdown |
| SC-12 | `security.txt` served 200 | `curl https://tene.sh/.well-known/security.txt` returns 200 with RFC 9116 fields |
| SC-13 | Hero + 3 FAQ capsules | Each first paragraph: 40-60 words, self-contained answer (manual review) |
| SC-14 | `/blog` renders | `/blog` returns 200, lists ≥1 post |
| SC-15 | Launch post live | `/blog/introducing-tene` returns 200, ≥1,500 words, Article schema present |
| SC-16 | RSS feed valid | `/feed.xml` validates on feed validator; `<content:encoded>` non-truncated |
| SC-17 | package.json fields | `apps/web/package.json` has `homepage: "https://tene.sh"` and `repository` pointing to tene-ai/tene |
| SC-18 | ≥1 page indexed | GSC URL Inspection for homepage returns "URL is on Google" within 8-12 weeks |

SC-18 is deferred (Google indexing lag); SC-1 through SC-17 are **required before Track 1 merge to main**.

---

## 5. Constraints

| # | Constraint | Source | Implication |
|---|---|---|---|
| C-1 | **No backend changes.** All work is in `apps/web/` only. No `cmd/server/`, `internal/api/`, no Go changes. | `.claude/rules/open-core.md` | Purely frontend/static PRD |
| C-2 | **Static-only deployment (Vercel).** Landing remains fully static (no SSR runtime logic that requires a server beyond Vercel Edge). | `.claude/rules/architecture.md`, current `page.tsx` is fully static | Blog uses `generateStaticParams`; RSS route can be static or edge |
| C-3 | **Next.js 16 API must be verified against `node_modules/next/dist/docs/` at Design time.** Plan's §8 snippets were written against Next 15. | `apps/web/AGENTS.md` warning | Design phase MUST verify API shape for sitemap/robots/metadata/MDX integration |
| C-4 | **Founder identity = "Kay Kim"** (Person schema), GitHub URL `https://github.com/agent-kay-it` | Plan §8.4 | Hardcoded in Organization/Person JSON-LD |
| C-5 | **GDPR default-deny cookies.** GA4 must not load before user consent. | Plan §7, Track A | `<CookieBanner>` component required; Consent Mode v2 |
| C-6 | **GA ID via tene vault, not `.env` file.** | `CLAUDE.md` secret rules | `tene set NEXT_PUBLIC_GA_ID ...`; Vercel env var mirrors; `.env.example` documents only |
| C-7 | **MIT-license compatible deps only.** `next-mdx-remote` (MIT), `gray-matter` (MIT), `shiki` (MIT) all OK. | Open-core | Reject any GPL/AGPL dep |
| C-8 | **No CMS. No Contentlayer.** MDX files live in `content/posts/` committed to the repo. | Plan §6 YAGNI | Blog tooling stays minimal |
| C-9 | **Dark-only design system.** Blog typography must use brand palette (`--accent: #00ff88`, Geist Sans/Mono). | `.claude/rules/conventions.md` | `@tailwindcss/typography` configured with brand colors |
| C-10 | **Canonical URLs absolute.** `metadataBase` already set; every blog post sets its canonical. | existing `layout.tsx` | Prevents dev.to canonical-stealing in P1 syndication |

---

## 6. Out of Scope (explicit)

Track 1 explicitly excludes the following. These belong to Track 2 (P1) or Track 3 (P2). If a Design/Do decision would pull any of these in, escalate back to PRD first.

### P1 deferred (will be its own `/pdca` cycle)
- 3 more cornerstone blog posts (comparison, architecture, BIP-39)
- `/compare` data-rich comparison page
- `/guides/secret-management-for-ai-agents` pillar page (3-5k words)
- `/about` author bio page
- FAQ expansion to 15 items
- Dynamic OG images per blog post (`opengraph-image.tsx`)
- Homebrew tap + signed commits + SLSA provenance
- Awesome-list PR submissions (awesome-cli-apps, awesome-go, awesome-security, etc.)
- Marketplace listings beyond ClawHub (LobeHub, SkillHub, claudeskills.info)
- Dev.to / Hashnode syndication (14-day canonical-first cycle)
- Show HN submission

### P2 deferred
- "State of AI Secret Management 2026" original research post
- Reddit seeding (r/commandline, r/golang, r/selfhosted, r/devops)
- Product Hunt launch
- Topic-cluster interlinking pass
- Scoop / AUR / asdf / mise package index listings
- `Review` schema (no public testimonials yet)

### Outside the strategy entirely
- External distribution — delegated to existing `/growth-daily` skill
- Newsletter platform
- Session replay
- A/B testing framework
- Localized (ko/ja) landing pages
- Community forum

---

## 7. Dependencies

### External (user/human actions required)
| # | Dep | Owner | When needed |
|---|---|---|---|
| D-1 | Create GA4 property at analytics.google.com, copy Measurement ID | User (Kay) | Before US-2/SC-3 |
| D-2 | Add Vercel env var `NEXT_PUBLIC_GA_ID` (production + preview) | User | Before first prod deploy |
| D-3 | Add DNS TXT record for GSC verification (Vercel DNS or Cloudflare) | User | Before US-8/SC-6 |
| D-4 | Verify property in GSC UI; submit sitemap; request-index homepage | User | Before SC-7/SC-8 |
| D-5 | Approve launch blog post content (1,500-word "Introducing tene") | User | Before US-13/SC-15 |
| D-6 | Provide PGP key fingerprint OR security contact email for `security.txt` | User | Before US-10/SC-12 |
| D-7 | Approve cookie banner copy (EN only for P0) | User | Before US-3/SC-4 |

### Technical (npm deps to install at Design)
- `@vercel/analytics`, `@vercel/speed-insights`, `@next/third-parties` (peer: Next 16)
- `next-mdx-remote`, `gray-matter`, `shiki`
- `@tailwindcss/typography`
- All MIT-licensed, all compatible with React 19 + Next 16

### Internal (upstream within this repo)
- None — this is a pure frontend feature in `apps/web/`; no changes to `pkg/`, `internal/`, `cmd/`, or `go.mod` required.

---

## 8. Risks

| # | Risk | Severity | Mitigation |
|---|---|---|---|
| R-1 | **Next.js 16 API divergence from Plan §8 snippets (which were written against Next 15).** `sitemap.ts`, `robots.ts`, `generateMetadata` signatures, MDX App Router integration may have breaking changes. | **Critical** | Design phase MUST run `/pdca team` codebase deep-analysis first; Design agent MUST consult `node_modules/next/dist/docs/` before picking APIs; build-and-test on a dev branch before committing. `apps/web/AGENTS.md` warning is load-bearing. |
| R-2 | **Cookie consent banner hurts conversion.** A default-deny banner on a static developer landing is friction that could reduce install-script copies. | High | Keep banner minimal (~60 LOC), single "Accept" button, bottom-right corner, dismissible, non-modal. Do not block the install command copy behind consent. Measure pre/post banner with Vercel Analytics (cookieless) which fires regardless. |
| R-3 | **GSC indexing lag (8-12 weeks).** Team may interpret lack of early traffic as failure. | Medium | Set explicit expectation in Success Criteria: SC-18 is "SC-18 is deferred"; only SC-1 to SC-17 are P0 DoD. Weekly GSC check starting week 4. |
| R-4 | **`llms-full.txt` no proven ROI.** Search Engine Land study (8/9 sites saw no change); Google publicly does not use it. | Low | Already down-scoped — one script, one output file, no ongoing maintenance. Accept as low-cost hygiene. |
| R-5 | **MDX + Next 16 RSC integration gotchas.** `next-mdx-remote/rsc` has had churn; code highlighting (`shiki`) must run at build time, not in the client bundle. | Medium | Design phase picks between `next-mdx-remote/rsc` vs `mdx-bundler` vs native `@next/mdx` — decision documented with trade-offs; build a throwaway spike if unclear. |
| R-6 | **Blog typography collision with existing dark-only palette.** `@tailwindcss/typography` defaults may override brand colors. | Low | Use `prose-invert` + custom `@apply` overrides with tene's neon-green accent; verify in Checkpoint 3 of Design phase. |
| R-7 | **Event tracking fires before consent.** Privacy regression — events trigger before GA loads would queue but mustn't leak PII. | Medium | Ensure `sendGAEvent` is no-op when consent not granted; test with network throttling. |
| R-8 | **Maintainer forgets to add `NEXT_PUBLIC_GA_ID` to Vercel preview env.** GA fails silently in preview deploys. | Low | `layout.tsx` conditionally mounts `<GoogleAnalytics>` only when env var present; document in AGENTS.md. |
| R-9 | **RSS `<content:encoded>` XML-escaping bug.** Unescaped `<` in MDX breaks feed. | Medium | Use a battle-tested escaper (`xmlbuilder2` or manual CDATA wrap); validate against `validator.w3.org/feed` in SC-16. |
| R-10 | **Founder name / photo / bio not finalized.** Blocks Person schema and launch post byline. | Low | Use "Kay Kim" + GitHub URL as minimum viable Person schema (Plan §8.4); bio page deferred to P1. |

---

## Appendix A — Task-to-Story Matrix

| P0 Task | Maps to Stories | Maps to JTBD | Critical path |
|---|---|---|---|
| P0-1 Vercel Analytics | US-1 | JTBD-1 | ✅ |
| P0-2 GA4 via @next/third-parties | US-2 | JTBD-1 | ✅ |
| P0-3 Cookie banner | US-3 | JTBD-1 | ✅ |
| P0-4 GA4 property + vault env | US-5 | JTBD-1 | Dep D-1, D-2 |
| P0-5 GSC verification | US-8 | JTBD-2 | Dep D-3, D-4 |
| P0-6 Dynamic sitemap.ts | US-6 | JTBD-2 | ✅ |
| P0-7 Dynamic robots.ts w/ AI allowlist | US-7 | JTBD-2 | ✅ |
| P0-8 Submit sitemap to GSC | US-8 | JTBD-2 | Dep D-4 |
| P0-9 Organization + Person + BreadcrumbList | US-11 | JTBD-3 | ✅ |
| P0-10 llms-full.txt | US-9 | JTBD-2 | ✅ |
| P0-11 security.txt | US-10 | JTBD-2 | Dep D-6 |
| P0-12 package.json homepage/repository | US-13 | JTBD-4 | ✅ |
| P0-13 Hero + 3 FAQ capsules | US-12 | JTBD-3 | ✅ |
| P0-14 Event instrumentation (4 events) | US-4 | JTBD-1 | ✅ |
| P0-15 Blog scaffold + MDX | US-13 | JTBD-4 | ✅ |
| P0-16 First launch post | US-13 | JTBD-4 | Dep D-5 |
| P0-17 RSS feed | US-14 | JTBD-4 | ✅ |
| P0-18 Article + BreadcrumbList on blog posts | US-14, US-11 | JTBD-3, JTBD-4 | ✅ |

---

## Appendix B — Handoff Checklist to `/pdca team` and `/pdca design`

Before the Design phase starts:

- [ ] Read this PRD entirely
- [ ] Read the Plan doc (`docs/01-plan/web-seo-aeo-analytics-strategy.md`) — especially §8 snippets and §1.1/1.2 audit
- [ ] Read `apps/web/AGENTS.md` warning
- [ ] Run `/pdca team web-seo-aeo-analytics-strategy` to get codebase deep-analysis: current `layout.tsx` structure, existing component inventory, `next.config.ts` state, `package.json` deps
- [ ] Cross-check every Plan §8 snippet against Next 16 docs in `node_modules/next/dist/docs/` — produce a "Next 15 → 16 API delta" note
- [ ] Decide MDX pipeline: `next-mdx-remote/rsc` vs `@next/mdx` vs `mdx-bundler` (trade-off table required)
- [ ] Decide cookie consent architecture: server component wrapper vs client-only, storage key, re-prompt interval
- [ ] Decide where `scripts/gen-llms-full.ts` runs: `prebuild` hook vs manual vs GitHub Action
- [ ] Confirm with user: launch post title and opening paragraph before Do phase starts

After Design is approved, Do phase implements against these 18 tasks + 14 stories. Success Criteria (§4) become the `analyze` phase's match-rate checklist.

---

**End of PRD. Total scope: 18 tasks, 14 user stories, 4 JTBD, 18 success criteria, 10 constraints, 10 risks. Next: `/pdca team web-seo-aeo-analytics-strategy` for codebase deep-analysis, then `/pdca design web-seo-aeo-analytics-strategy`.**
