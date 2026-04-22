import type { Comparison } from "./types";

export const infisicalComparison: Comparison = {
  slug: "infisical",
  competitorName: "Infisical",
  competitorHomepage: "https://infisical.com",
  metaTitle: "tene vs Infisical — self-hosted server vs local-first CLI",
  metaDescription:
    "Infisical is a self-hosted or SaaS secret manager with a web dashboard. tene is a local-first CLI with zero infrastructure. AI-editor safety out of the box. Full comparison.",
  publishedAt: "2026-04-22",
  updatedAt: "2026-04-22",

  headline: "tene vs Infisical",
  subheadline:
    "Infisical needs a server (self-hosted or SaaS). tene runs on your machine. Same encryption guarantees, completely different ops surface.",
  heroKeywords: [
    "Infisical alternative",
    "tene vs Infisical",
    "local-first secret manager",
    "secret manager without server",
  ],

  intro:
    "Infisical is a solid open-source secret manager — but it is a server-side product. You either run their hosted SaaS or self-host the PostgreSQL-backed service. Either way, there's infrastructure to manage, a web dashboard to secure, and a user database to worry about.\n\ntene takes a different approach: the secret manager is the CLI. There is no server. There is no database. The vault is a SQLite file on your machine encrypted with XChaCha20-Poly1305. You get the same encryption-at-rest guarantee as Infisical, without the ops overhead or the signup.",

  comparisonRows: [
    { dimension: "Architecture", tene: "CLI + local SQLite vault (no server)", competitor: "Server + PostgreSQL + web dashboard (self-hosted or SaaS)" },
    { dimension: "Deployment", tene: "Single Go binary", competitor: "Docker Compose / Kubernetes / managed SaaS" },
    { dimension: "AI-editor integration", tene: "Auto-generates rules for Claude, Cursor, Windsurf, Gemini, Codex", competitor: "None" },
    { dimension: "Encryption at rest", tene: "XChaCha20-Poly1305 + Argon2id KDF", competitor: "AES-256 (server-side, with KMS on hosted)" },
    { dimension: "Signup required", tene: "No (CLI is fully local)", competitor: "Yes (for hosted); not for self-hosted" },
    { dimension: "Runtime injection", tene: "`tene run -- <cmd>`", competitor: "`infisical run -- <cmd>` (similar)" },
    { dimension: "Open source license", tene: "MIT", competitor: "MIT (core) + Infisical License (enterprise features)" },
    { dimension: "Team sync", tene: "Optional E2E sync (Pro at app.tene.sh)", competitor: "Built-in (workspaces, RBAC, approval flows)" },
    { dimension: "Offline usability", tene: "100% offline CLI", competitor: "Requires connectivity to the Infisical server" },
    { dimension: "Best for", tene: "Individual devs + small teams + AI-heavy workflows", competitor: "Mid-to-large teams wanting a central dashboard + RBAC" },
  ],

  migration: {
    title: "Migrate from Infisical to tene",
    summary:
      "Export Infisical secrets as .env, then import into tene. Infisical stays available for rollback.",
    steps: [
      { title: "Export secrets from Infisical", command: "infisical export --format dotenv > .env" },
      { title: "Install tene", command: "curl -sSfL https://tene.sh/install.sh | sh" },
      { title: "Initialize a local vault", command: "tene init" },
      { title: "Import the .env", command: "tene import .env" },
      { title: "Delete the plaintext export", command: "rm .env" },
      { title: "Run your app through tene", command: "tene run -- npm start" },
    ],
    postMigrationNote:
      "If your team depends on Infisical's dashboard + RBAC, keep Infisical as the shared source and use tene just for local developer runtime + AI-editor safety. The two are complementary — tene doesn't try to replace server-side workspaces.",
  },

  sections: [
    {
      heading: "Where Infisical shines",
      body: "Infisical's workspaces, approval flows, and dashboard are excellent for teams that want governance over secrets. Their integrations (k8s, Vercel, Terraform) and SDK are first-class. For a company of 20+ engineers with compliance requirements, Infisical is a better fit than tene.",
    },
    {
      heading: "Where tene shines",
      body: "tene owns two niches Infisical doesn't serve: (1) individual developers who don't want to stand up or pay for a server, and (2) AI-editor workflows where the threat model is secret leakage through LLM context windows. tene's auto-generated rule files (CLAUDE.md, .cursor/rules/tene.mdc, .windsurfrules, GEMINI.md, AGENTS.md) teach every major AI editor to call `tene run --` instead of reading `.env`.",
    },
    {
      heading: "You can use both",
      body: "tene and Infisical are complements. Use Infisical as the organization-wide source of truth, and `infisical export | tene import` on developer machines so local workflows pick up AI-editor safety without losing the team's central dashboard. `tene run --` then becomes the local runtime wrapper that keeps secrets out of LLM context.",
    },
  ],

  faqs: [
    {
      question: "Can I self-host tene's team sync?",
      answer:
        "The CLI's team sync goes through app.tene.sh by default, with client-side encryption so the server never sees plaintext. Self-hosting the sync server is on the roadmap. In the meantime, you can use tene locally and Infisical server-side — they don't conflict.",
    },
    {
      question: "Is tene open source like Infisical?",
      answer:
        "Yes, the entire CLI is MIT licensed at github.com/tomo-kay/tene. Infisical's core is MIT as well, but some enterprise features use the Infisical License. Both are auditable.",
    },
    {
      question: "What if I need RBAC?",
      answer:
        "tene doesn't have RBAC (there's nothing to authorize on a local-only tool). If you need RBAC, use Infisical server-side and tene only for developer-machine runtime safety.",
    },
    {
      question: "How does tene handle dynamic secrets (DB credentials with short TTL)?",
      answer:
        "tene is optimized for long-lived developer secrets (API keys, OpenAI tokens, webhook keys). For short-TTL dynamic secrets, Infisical or HashiCorp Vault are the right tools.",
    },
  ],
};
