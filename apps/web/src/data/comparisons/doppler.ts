import type { Comparison } from "./types";

export const dopplerComparison: Comparison = {
  slug: "doppler",
  competitorName: "Doppler",
  competitorHomepage: "https://www.doppler.com",
  metaTitle: "tene vs Doppler — local-first vs cloud secret management",
  metaDescription:
    "Doppler is a cloud secret manager at $21/user/month. tene is local-first, MIT licensed, and free — with AI-editor integration Doppler does not offer. Full comparison + migration guide.",
  publishedAt: "2026-04-22",
  updatedAt: "2026-04-22",

  headline: "tene vs Doppler",
  subheadline:
    "Doppler is cloud-based and proprietary. tene is local-first and MIT. Same developer workflow, very different ops story.",
  heroKeywords: [
    "tene vs Doppler",
    "Doppler alternative",
    "open source Doppler",
    "local-first secret manager",
  ],

  intro:
    "Doppler is a well-engineered cloud secret manager aimed at teams that want a central dashboard and granular access controls. tene is a different product for a different audience: a local-first CLI for individual developers and small teams who want zero infrastructure, zero signup, and AI-editor safety out of the box.\n\nIf you already pay for Doppler and need SOC 2, dashboards, and audit logs on a managed service, stay on Doppler. If you want your secrets to live on your own machine, cost $0, and stay out of AI context windows, tene is the answer.",

  comparisonRows: [
    { dimension: "Hosting", tene: "Local-first (your machine)", competitor: "Cloud (Doppler servers)" },
    { dimension: "Price", tene: "$0 (MIT, unlimited secrets)", competitor: "$21 / user / month (Team); Free tier has workplace limits" },
    { dimension: "AI-editor integration", tene: "Auto-generates CLAUDE.md, .cursor/rules, .windsurfrules, GEMINI.md, AGENTS.md", competitor: "Not AI-aware" },
    { dimension: "Secrets reach AI context", tene: "No (runtime env injection only)", competitor: "Depends on how you inject — `doppler run` also injects, but the dashboard is a separate attack surface" },
    { dimension: "Vendor lock-in", tene: "Zero — SQLite vault is portable", competitor: "Requires Doppler account + sync; migration off requires `doppler secrets download`" },
    { dimension: "Open source", tene: "Yes (MIT)", competitor: "No (proprietary SaaS)" },
    { dimension: "Offline", tene: "100% offline", competitor: "Requires network for CLI (caching available, but online is the happy path)" },
    { dimension: "Team sync", tene: "Optional E2E encrypted sync via app.tene.sh (Pro)", competitor: "Core product — strong team features" },
    { dimension: "Dashboard UI", tene: "CLI only (dashboard coming via Cloud Pro)", competitor: "Full web dashboard with access controls" },
    { dimension: "Audit log", tene: "Local CLI history + optional cloud audit on Pro", competitor: "Comprehensive audit log (enterprise plan)" },
  ],

  migration: {
    title: "Migrate from Doppler to tene",
    summary:
      "Export from Doppler as a .env, then import into a tene vault. Doppler stays available the whole time, so you can roll back freely.",
    steps: [
      { title: "Export Doppler secrets as .env", command: "doppler secrets download --no-file --format env > .env" },
      { title: "Install tene", command: "curl -sSfL https://tene.sh/install.sh | sh" },
      { title: "Initialize a local vault", command: "tene init" },
      { title: "Import the .env", command: "tene import .env" },
      { title: "Delete the plaintext .env", command: "rm .env" },
      { title: "Run your app through tene", command: "tene run -- npm start" },
    ],
    postMigrationNote:
      "Run both side-by-side for a week to validate. Once you are confident, revoke the Doppler tokens and cancel the subscription.",
  },

  sections: [
    {
      heading: "When Doppler is the right answer",
      body: "Large teams with strong compliance requirements (SOC 2, audit logs, access controls) will find Doppler's managed service a better fit. The dashboard, RBAC, and integrations ecosystem (CI, k8s, dynamic envs) are genuinely strong. If you already have a Doppler contract and engineers trained on it, the switching cost is not worth it for a single-developer workflow.",
    },
    {
      heading: "When tene is the right answer",
      body: "Individual developers, early-stage startups, and open-source maintainers often don't need $250/month per seat for a secret manager. They need: encryption at rest, AI-editor safety, multi-environment support, and a runtime that injects secrets without writing them to disk. That's exactly tene's scope.\n\nThe decision tree: if your team has a compliance team, pick Doppler. If your team is you (or you + a few collaborators), pick tene.",
    },
    {
      heading: "What you give up by moving to tene",
      body: "There is no web dashboard (yet — it's being designed for Pro). There is no native k8s operator. Audit logs are local CLI history on the free tier. If those are dealbreakers, Doppler is the better pick. For the majority of developer-workstation + CI workflows, tene's encrypted vault + `tene run --` covers the actual daily needs without an account to manage.",
    },
  ],

  faqs: [
    {
      question: "Can I run both for a week to compare?",
      answer:
        "Yes. tene is local-first and doesn't touch Doppler; nothing breaks. Export Doppler secrets as a .env, import into tene, and run both in parallel until you're confident.",
    },
    {
      question: "What happens to my Doppler audit log?",
      answer:
        "It stays in Doppler. If you need a historical record, keep your Doppler account on a free tier as a read-only archive. tene Pro will add an encrypted audit log in a future release.",
    },
    {
      question: "Does tene have team features?",
      answer:
        "The CLI is local-first and free forever. Optional E2E encrypted team sync is available at app.tene.sh (Pro). Server-side, only ciphertext is stored.",
    },
    {
      question: "How is tene different for AI agents?",
      answer:
        "tene generates rule files for five AI editors (CLAUDE.md, .cursor/rules/tene.mdc, .windsurfrules, GEMINI.md, AGENTS.md) that teach the agent to call `tene run --` instead of reading `.env`. Doppler does not ship any AI-editor integration.",
    },
  ],
};
