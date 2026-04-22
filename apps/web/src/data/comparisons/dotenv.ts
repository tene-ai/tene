import type { Comparison } from "./types";

export const dotenvComparison: Comparison = {
  slug: "dotenv",
  competitorName: ".env files",
  competitorHomepage: "https://github.com/motdotla/dotenv",
  metaTitle: "tene vs .env — stop AI agents from reading your secrets",
  metaDescription:
    "Why plaintext .env files are a liability in the AI coding era. tene encrypts your secrets and injects them at runtime so Claude Code, Cursor, and Copilot never see the values.",
  publishedAt: "2026-04-22",
  updatedAt: "2026-04-22",

  headline: "tene vs .env",
  subheadline:
    "Plaintext .env files are now an AI liability. tene encrypts them and injects values at runtime — agents never see the secrets.",
  heroKeywords: [
    "dotenv alternative",
    "hide env from AI",
    "Claude Code secret management",
    "Cursor env file",
  ],

  intro:
    "The `.env` file was invented in 2012, before AI coding agents existed. Every modern AI assistant — Claude Code, Cursor, Windsurf, Gemini, Copilot — reads your project files as context. That means your API keys, database passwords, and OpenAI tokens get loaded into the LLM context window the moment you open your editor.\n\ntene keeps `.env` out of the equation: your secrets live in an encrypted SQLite vault, and `tene run -- <cmd>` injects them as environment variables at runtime. The AI agent sees the command, never the values.",

  comparisonRows: [
    { dimension: "Storage", tene: "Encrypted SQLite vault (XChaCha20-Poly1305)", competitor: "Plaintext file on disk" },
    { dimension: "AI agent reads values", tene: "No — values never enter context window", competitor: "Yes — .env is part of project context" },
    { dimension: "Accidental git commit", tene: "Safe — vault is ciphertext", competitor: "Common leak vector (even with .gitignore)" },
    { dimension: "Runtime injection", tene: "`tene run -- <cmd>` injects env vars", competitor: "Requires an in-code loader (dotenv package)" },
    { dimension: "Multi-environment", tene: "Built-in (`--env dev/staging/prod`)", competitor: "Manual file juggling (.env.dev, .env.prod)" },
    { dimension: "Recovery", tene: "12-word BIP-39 mnemonic", competitor: "Whatever backup you remembered to make" },
    { dimension: "Vendor lock-in", tene: "None — portable SQLite file", competitor: "None" },
    { dimension: "Price", tene: "Free (MIT)", competitor: "Free" },
  ],

  migration: {
    title: "Migrate from .env to tene in 30 seconds",
    summary:
      "`tene import` converts an existing .env file into the encrypted vault. After that, you delete `.env` and use `tene run --` in its place.",
    steps: [
      { title: "Install tene", command: "curl -sSfL https://tene.sh/install.sh | sh" },
      { title: "Initialize a vault", command: "tene init" },
      { title: "Import your existing .env", command: "tene import .env" },
      { title: "Delete the plaintext .env", command: "rm .env" },
      { title: "Run your app through tene", command: "tene run -- npm start" },
    ],
    postMigrationNote:
      "Your build/test/CI commands stay the same — just prefixed with `tene run --`. Add `TENE_MASTER_PASSWORD` to your CI secrets to run non-interactively.",
  },

  sections: [
    {
      heading: "Why .env is not a secret in 2026",
      body: "AI coding assistants index every file in your workspace so they can answer questions like \"why did this build fail?\". That index includes `.env`. In practice, your Stripe key or database password will appear in tool_result blocks, session transcripts, completion suggestions, and anything that gets logged.\n\nThe traditional advice — `.gitignore` the file — only stops the secret from reaching GitHub. It does nothing about the LLM context window.",
    },
    {
      heading: "Why runtime injection beats a plaintext file",
      body: "tene never writes plaintext values to disk. `tene run -- <cmd>` decrypts secrets in memory, sets them as environment variables on the child process, and exits. Child processes — including whatever AI editor you launch through it — see the same env vars they would have read from .env, but no file on disk contains the plaintext.\n\nThat means: no `.env` to accidentally commit, no plaintext for an AI agent to index, and no leftover backup file after a rotation.",
    },
    {
      heading: "What you keep, what you drop",
      body: "You keep: `process.env.STRIPE_KEY` in your Node app, `os.Getenv(\"STRIPE_KEY\")` in your Go service, `os.environ[\"STRIPE_KEY\"]` in Python. You drop: the plaintext `.env`, the `dotenv` npm package import (no longer needed), and the mental overhead of remembering whether you rotated staging after you last edited prod.",
    },
  ],

  faqs: [
    {
      question: "Does my app need code changes?",
      answer:
        "No. `tene run -- <cmd>` sets the same environment variables your code already reads via `process.env.*`, `os.Getenv`, `os.environ`, etc. You typically remove the `dotenv` import because it is no longer needed.",
    },
    {
      question: "What about my CI pipeline?",
      answer:
        "Add `TENE_MASTER_PASSWORD` to your CI secrets. Then run `tene run --no-keychain -- npm test`. The `--no-keychain` flag tells tene to read the master password from the environment variable instead of prompting.",
    },
    {
      question: "Can I still share secrets with my team?",
      answer:
        "The CLI is strictly local. Optional end-to-end encrypted team sync is available via app.tene.sh (Pro). Each teammate runs `tene pull` and decrypts the vault with their own master password.",
    },
    {
      question: "Is this really safer than .env?",
      answer:
        "For the AI-agent threat model — values ending up in an LLM context window — yes, because values never exist as plaintext on disk. For the general threat of a compromised machine, tene is as strong as the master password + Argon2id KDF + OS keychain, which is much stronger than a plaintext file.",
    },
  ],
};
