import type { Comparison } from "./types";

export const vaultComparison: Comparison = {
  slug: "vault",
  competitorName: "HashiCorp Vault",
  competitorHomepage: "https://www.hashicorp.com/products/vault",
  metaTitle: "tene vs HashiCorp Vault — enterprise server vs developer CLI",
  metaDescription:
    "HashiCorp Vault is an enterprise server for dynamic, server-side secrets. tene is a local-first CLI for developer machines and AI-safe workflows. They're complements, not substitutes.",
  publishedAt: "2026-04-22",
  updatedAt: "2026-04-22",

  headline: "tene vs HashiCorp Vault",
  subheadline:
    "HashiCorp Vault is the gold standard for enterprise server-side secrets. tene is the local-first CLI for developer machines and AI-editor safety.",
  heroKeywords: [
    "Vault alternative for developers",
    "HashiCorp Vault alternative",
    "tene vs Vault",
    "local secret manager",
  ],

  intro:
    "HashiCorp Vault is the incumbent enterprise secret manager. It handles dynamic secrets, PKI, transit encryption, and policy-driven access at production scale. It is also expensive to run, complex to operate, and wildly overkill for the individual developer workflow.\n\ntene doesn't compete with Vault on Vault's turf. If you already run Vault for production server-side secrets, keep it. tene solves a different problem: encrypted secrets on the developer's workstation that stay out of AI-editor context windows. Run both — they don't conflict.",

  comparisonRows: [
    { dimension: "Target audience", tene: "Individual developers + small teams + AI workflows", competitor: "Platform / SRE / security teams in mid-to-large enterprises" },
    { dimension: "Infrastructure", tene: "None (CLI only)", competitor: "HA cluster, storage backend, unseal workflow, PKI, audit backends" },
    { dimension: "Price", tene: "Free (MIT)", competitor: "Free (Vault OSS) to $1,000+ / month for Vault Enterprise HA" },
    { dimension: "Dynamic secrets", tene: "No (long-lived secrets only)", competitor: "Yes — best-in-class (DB, cloud, PKI, SSH)" },
    { dimension: "AI-editor integration", tene: "Auto-generates CLAUDE.md, .cursor/rules/tene.mdc, .windsurfrules, GEMINI.md, AGENTS.md", competitor: "None" },
    { dimension: "Encryption", tene: "XChaCha20-Poly1305 + Argon2id", competitor: "AES-GCM + Shamir unseal / KMS auto-unseal" },
    { dimension: "Scale", tene: "Single developer / small team", competitor: "Thousands of clients, millions of leases" },
    { dimension: "Operational complexity", tene: "Install binary + `tene init`", competitor: "Cluster ops, seal/unseal, backup, upgrade procedures" },
    { dimension: "Open source", tene: "MIT", competitor: "Vault OSS is BUSL 1.1 (source-available, not OSI-approved)" },
  ],

  sections: [
    {
      heading: "They solve different problems",
      body: "Vault is a production secrets service: policies, audit logs, dynamic leases, and integrations with every major infrastructure layer. If you need AWS IAM credentials that expire in 15 minutes, Vault is the answer.\n\ntene is a developer-workstation and CI tool: encrypted .env replacement, runtime env-var injection, AI-editor safety. If you want your API keys to stay out of Claude Code's context window, tene is the answer. Neither replaces the other.",
    },
    {
      heading: "The common pattern: use both",
      body: "Production services read from Vault (dynamic DB creds, rotating API keys, short-lived cloud credentials). Developer machines use tene for the local developer loop: storing the handful of long-lived secrets (Stripe, OpenAI, webhook keys) that make local dev work, keeping them encrypted, and keeping AI editors from reading them.\n\nA common pattern: `vault kv get -format=json` to bootstrap local secrets from Vault, pipe into `tene import`. Developers then `tene run --` their local stack.",
    },
    {
      heading: "When NOT to use tene",
      body: "If your secret rotation policy is measured in minutes, you need dynamic secrets, or you have a compliance team that requires audit-log-everything: use Vault. tene is intentionally not a replacement for that.",
    },
  ],

  faqs: [
    {
      question: "Why no comparison table for dynamic secrets?",
      answer:
        "Because tene doesn't have them, and it would be misleading to compare. Vault's dynamic secret engines (AWS, database, SSH, PKI) are in a different category — tene focuses on long-lived developer secrets + runtime injection.",
    },
    {
      question: "Can tene bootstrap secrets from Vault?",
      answer:
        "Yes, via a shell pipeline: `vault kv get -format=json secret/myapp | jq -r '.data.data | to_entries[] | \"\\(.key)=\\(.value)\"' | tee .env && tene import .env && rm .env`. This gives your local dev environment the current production secrets without running Vault locally.",
    },
    {
      question: "Is tene a replacement for Vault?",
      answer:
        "No. Vault is the right tool for production server-side secret management. tene is the right tool for developer-workstation + AI-editor workflows. They're complements.",
    },
    {
      question: "What about audit logs?",
      answer:
        "tene's CLI emits local history. Optional cloud audit logs are on the roadmap for Pro (at app.tene.sh). If you need certified audit logs today, stay on Vault.",
    },
  ],
};
