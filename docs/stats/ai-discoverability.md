# AI Discoverability — KPI Tracking

> Tracks every track of the `ai-discoverability` feature in one place.
> Plan:   `docs/01-plan/features/ai-discoverability.plan.md`
> Design: `docs/02-design/features/ai-discoverability.design.md`
>
> **Update cadence**: monthly on the 1st (run `scripts/ai-discoverability/review.sh`
> and fill in the row). Additional ad-hoc updates are fine; keep them in the
> Run History section.

---

## Latest Snapshot (2026-04-22)

| Track | KPI | Current | Target M3 | Target M6 | Target M12 |
|-------|-----|--------:|----------:|----------:|-----------:|
| T1 | GitHub topics count | **20** ✅ (applied 2026-04-22) | 20 | 20 | 20 |
| T1 | README "For AI Agents" H2 | ✅ shipped | ✅ | ✅ | ✅ |
| T2 | `tene.sh/llms.txt` returns 200 + updated content | ✅ shipped (via `public/llms.txt`) | ✅ | ✅ | ✅ |
| T2 | `tene.sh/llms-full.txt` returns 200 | ✅ shipped | ✅ | ✅ | ✅ |
| T2 | GitHub root `tene/llms.txt` | ✅ shipped | ✅ | ✅ | ✅ |
| T2 | `tene init` links CLAUDE.md → llms.txt | ✅ shipped (template.go) | ✅ | ✅ | ✅ |
| T3 | `/vs/*` pages deployed | 5 *(pending Vercel deploy)* | 5 | 5 | 5 |
| T3 | Schema.org JSON-LD on each page | ✅ shipped | ✅ | ✅ | ✅ |
| T3 | Lighthouse SEO ≥ 95 | pending measurement | ≥ 95 | 100 | 100 |
| T3 | "tene vs X" in Google SERP | 0 indexed | 3 indexed | 5 indexed, first page | first page |
| T4 | Awesome list PRs opened | 0 | 3 | 5 | 6 |
| T4 | Awesome list PRs merged | 0 | 1 | 3 | 5 |
| T5 | Dev.to articles published | 0 | 2 | 4 | 5 |
| T5 | Dev.to cumulative views | 0 | 1,000 | 5,000 | 20,000 |
| — | GitHub stars | 5 | 50 | 200 | 1,500 |
| — | AI-agent recommendation (out of 10 prompts) | 0 / 10 | 2 / 10 | 5 / 10 | 7 / 10 |
| — | tene.sh organic traffic | — | 200 / mo | 1,000 / mo | 5,000 / mo |

---

## AI-Agent Recommendation Test — protocol

Run **monthly** in a fresh private session (no cookies, no memory) on
Claude, ChatGPT, and Gemini. Send each of the 10 prompts below exactly and
record whether tene is mentioned in the answer. Log the aggregated result
(e.g. `4 / 10` = 4 prompts out of 10 where any of the 3 agents mentioned
tene) in the Run History.

1. Recommend a secret manager for CLI developers.
2. How do I stop AI from reading my `.env` file?
3. What's the best `dotenv` alternative?
4. Open source vault CLI?
5. Claude Code + API keys — safe workflow?
6. Cursor secret management best practice?
7. Local-first encrypted secret storage?
8. `dotenv-vault` alternative in 2026?
9. How do I hide secrets from Copilot?
10. MIT licensed secret manager?

Paste raw outputs into `snapshots/YYYY-MM-DD-ai-test.md` for audit.

---

## Verification commands (run after any change)

```bash
# T1 — metadata
gh api repos/tomo-kay/tene --jq '{topics, description}'
gh api repos/tomo-kay/tene --jq '.topics | length'   # expect 20

# T2 — llms.txt routes (post-deploy)
curl -fsS -o /dev/null -w 'llms.txt=%{http_code}\n'       https://tene.sh/llms.txt
curl -fsS -o /dev/null -w 'llms-full.txt=%{http_code}\n'  https://tene.sh/llms-full.txt
curl -fsS https://raw.githubusercontent.com/tomo-kay/tene/main/llms.txt \
  | head -5   # must contain 'tene'

# T3 — comparison pages (post-deploy)
for slug in dotenv doppler dotenv-vault infisical vault; do
  code=$(curl -fsS -o /dev/null -w '%{http_code}' "https://tene.sh/vs/$slug")
  schema=$(curl -fsS "https://tene.sh/vs/$slug" | grep -c '"@type":"SoftwareApplication"' || true)
  printf '%-16s http=%s  schema=%s\n' "$slug" "$code" "$schema"
done

# Sitemap
curl -fsS https://tene.sh/sitemap.xml | grep -oE 'https://tene\.sh/[^<]+' | sort -u
```

---

## Run History

| Date | Event | Detail |
|------|-------|--------|
| 2026-04-22 | Phase 1 shipped | Topics script + README `## For AI Agents` + `public/llms.txt` + `public/llms-full.txt` + `tene/llms.txt` + CLAUDE.md template `Resources` section. Go tests pass. |
| 2026-04-22 | Phase 2 shipped | `/vs/[slug]` dynamic route + 5 comparison data files + Schema.org JSON-LD + UI components + dynamic sitemap/robots. `next build` generates 11 static pages. |
| 2026-04-22 | Phase 3 prepared | `.claude/templates/growth/awesome-pr-template.md` + `tene/docs/01-plan/awesome-lists-plan.md`. Awaiting human to open PRs one at a time. |
| 2026-04-22 | FR-01 + FR-03 applied | Ran `scripts/ai-discoverability/update-github-metadata.sh`. Topics: 11 → 20 (ai-agents, secret-management, cli, devsecops, encryption, developer-tools, dotenv, vault, go, vibe-coding, opensource added; "secrets", "tene" dropped). Description rewritten to AI-agent-friendly phrasing. Verified live via `gh api`. |
| 2026-04-23 | Vercel Analytics wired (M level) | Installed `@vercel/analytics@2.0.1` + `@vercel/speed-insights@2.0.0`. Injected `<Analytics />` and `<SpeedInsights />` in `apps/web/src/app/layout.tsx` body (after `<NoiseOverlay />`). `next build` succeeds (11 static pages); bundle chunk `110.*.js` (10 KB) contains `window.va` bootstrap + `/_vercel/insights/script` path. Pending user actions: (1) enable Analytics + Speed Insights in Vercel dashboard, (2) push → deploy, (3) verify `/_vercel/insights/view` 200 post-deploy, (4) +48 h visitor/vitals data check. |
| 2026-04-23 | **SUPERSEDED: Analytics → GA4** | Swapped out Vercel Analytics stack in favor of Google Analytics 4 (Measurement ID `G-9MRWMY6XBE`, property set up by Seung-jun). Uninstalled `@vercel/analytics` + `@vercel/speed-insights`; installed `@next/third-parties@16.2.4`. Replaced `<Analytics />`/`<SpeedInsights />` with `<GoogleAnalytics gaId="..."/>` in layout.tsx. Rewrote `src/lib/track.ts` to use `sendGAEvent` (track() API unchanged → 0 call-site edits). Updated CSP to allow `googletagmanager.com`, `google-analytics.com`, `analytics.google.com`. Runtime-verified: gtag function loaded, `_ga` client cookie set, `window.va/.vsi` completely removed. See `docs/03-analysis/analytics-ga4-swap.analysis.md` + `docs/05-qa/analytics-ga4-swap.qa-report.md`. |

<!-- BEGIN:runs -->
<!-- Add new runs above this marker; older runs remain in Run History -->
<!-- END:runs -->

---

## Notes

- **awesome-go deferred**: star gate requires ~100+ stars; revisit when the
  repo crosses that threshold.
- **Dev.to Phase 4 not started**: scheduled for M1 onwards. See Plan §4.
- **Google Search Console**: manual submission after first deploy. Add a
  row to Run History when submitted.
