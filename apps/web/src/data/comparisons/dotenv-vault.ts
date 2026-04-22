import type { Comparison } from "./types";

export const dotenvVaultComparison: Comparison = {
  slug: "dotenv-vault",
  competitorName: "dotenv-vault",
  competitorHomepage: "https://www.dotenv.org",
  metaTitle: "tene vs dotenv-vault — the Pro-tier shutdown migration guide",
  metaDescription:
    "dotenv-vault's Pro tier shut down in February 2026. tene is the local-first, MIT-licensed migration path — one command to import, zero signup, zero cloud. Full migration guide.",
  publishedAt: "2026-04-22",
  updatedAt: "2026-04-22",

  headline: "tene vs dotenv-vault",
  subheadline:
    "dotenv-vault's Pro tier shut down February 2026. tene is the drop-in local-first replacement — MIT, free, no signup.",
  heroKeywords: [
    "dotenv-vault alternative",
    "dotenv-vault shutdown",
    "dotenv-vault Pro discontinued",
    "migrate from dotenv-vault",
  ],

  intro:
    "In February 2026, dotenv-vault announced it was discontinuing its Pro tier. Free-tier users still get the CLI, but the team-sync and encrypted backup features many projects depended on are gone. If you were a Pro subscriber, you need a migration path now.\n\ntene is that path: a local-first encrypted secret manager that imports `.env` files directly, generates AI-editor rule files automatically, and costs $0 forever. No signup, no cloud account, no migration fees.",

  comparisonRows: [
    { dimension: "Current status", tene: "Active, MIT licensed", competitor: "Pro tier discontinued (Feb 2026); Free tier limited" },
    { dimension: "Hosting", tene: "Local-first on your machine", competitor: "Cloud-hosted (dotenv.org)" },
    { dimension: "Signup", tene: "None required", competitor: "Required to use CLI" },
    { dimension: "Encryption", tene: "XChaCha20-Poly1305 + Argon2id KDF", competitor: "AES (encrypted `.env.vault` file)" },
    { dimension: "AI-editor integration", tene: "Yes (CLAUDE.md, .cursor/rules, .windsurfrules, GEMINI.md, AGENTS.md)", competitor: "No" },
    { dimension: "Team sync", tene: "Optional, E2E encrypted (Pro plan at app.tene.sh)", competitor: "Was Pro-only; now discontinued" },
    { dimension: "Recovery", tene: "12-word BIP-39 mnemonic", competitor: "N/A (tied to dotenv.org account)" },
    { dimension: "Export path", tene: "Portable SQLite vault + `tene export --encrypted`", competitor: "`dotenv-vault pull` (while it still works)" },
    { dimension: "Vendor risk", tene: "None — binary is self-contained", competitor: "High — service is shutting down core features" },
    { dimension: "Price", tene: "$0 (MIT)", competitor: "Free tier only; previously paid tier gone" },
  ],

  migration: {
    title: "Migrate from dotenv-vault to tene (urgent)",
    summary:
      "Pull your secrets out of dotenv-vault while the CLI still works, then import into tene. Takes less than a minute for a typical project.",
    steps: [
      { title: "Pull current secrets from dotenv-vault", command: "dotenv-vault pull --no-cache", note: "Writes to .env in your project root." },
      { title: "Install tene", command: "curl -sSfL https://tene.sh/install.sh | sh" },
      { title: "Initialize a local encrypted vault", command: "tene init" },
      { title: "Import the pulled .env", command: "tene import .env" },
      { title: "Delete plaintext files", command: "rm .env .env.vault .env.me" },
      { title: "Run your app through tene", command: "tene run -- npm start" },
    ],
    postMigrationNote:
      "Don't delete your dotenv.org account immediately — keep it as a read-only backup for a month while you confirm every environment works. Then revoke tokens and deactivate.",
  },

  sections: [
    {
      heading: "What dotenv-vault's Pro shutdown actually means",
      body: "dotenv-vault's Pro tier included encrypted team sync, environment separation, and an encrypted `.env.vault` file that was supposed to be committable. With Pro gone, teams that depended on it for sharing secrets across developers have no upgrade path. The Free tier's CLI still works, but the team collaboration story is no longer viable.\n\nMany dotenv-vault users are discovering this the hard way — their CI pipeline fails, their deploy script errors out, their new hire can't spin up a dev environment.",
    },
    {
      heading: "Why tene is a natural replacement",
      body: "tene was designed from day one around the dotenv-vault use case with two twists: (1) the vault lives on your machine, not a third-party server, so nobody can deprecate it out from under you; (2) the vault is AI-agent aware, generating rule files for Claude Code, Cursor, Windsurf, Gemini, and Codex so your secrets stay out of LLM context windows.\n\nFor team sync, tene offers an optional Pro plan at app.tene.sh that uses client-side encryption — the server only sees ciphertext, so a future shutdown would still leave you with a working local vault.",
    },
    {
      heading: "If you were paying for dotenv-vault Pro",
      body: "You were paying roughly $6–12 per developer for encrypted team sync. tene's CLI is free forever and includes the same end-user workflow (`tene import`, `tene run --`, multi-environment). If you need team sync, Pro at app.tene.sh is comparable pricing with stronger crypto (XChaCha20-Poly1305 + per-user X25519 key wrapping) and no lock-in.",
    },
  ],

  faqs: [
    {
      question: "My CI was using dotenv-vault — what now?",
      answer:
        "Replace `dotenv-vault pull` in your CI with `tene run --no-keychain --`. Set `TENE_MASTER_PASSWORD` as a CI secret. The `--no-keychain` flag tells tene to read the password from the env var instead of the OS keychain.",
    },
    {
      question: "Can I keep the `.env.vault` file I committed to git?",
      answer:
        "No. tene's vault is a SQLite file at `.tene/vault.db` and is `.gitignore`d by default. If you need to sync secrets across machines, use `tene push` (Pro) or an encrypted backup via `tene export --encrypted`.",
    },
    {
      question: "What about the DOTENV_KEY environment variable?",
      answer:
        "DOTENV_KEY is a dotenv-vault concept. tene replaces it with `TENE_MASTER_PASSWORD` (for non-interactive use) and the OS keychain (for interactive use). You don't need DOTENV_KEY anymore.",
    },
    {
      question: "Does migration work for monorepo setups?",
      answer:
        "Yes. Run `tene init` once per project root. Each project gets its own `.tene/vault.db`. If you want shared team secrets across projects, use the Pro sync product.",
    },
  ],
};
