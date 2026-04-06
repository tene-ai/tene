// Design Ref: §3.4 — FAQ data with .env risk questions first
export type FAQ = { question: string; answer: string };

export const faqs: FAQ[] = [
  {
    question: "Why is .env dangerous with AI agents?",
    answer:
      "AI coding agents like Claude Code, Cursor, and Windsurf read all files in your project directory \u2014 including .env. This means your API keys, database passwords, and tokens are sent to AI models as plaintext context. You have no control over how that data is processed or stored.",
  },
  {
    question: "How does Tene keep secrets from AI?",
    answer:
      "Tene stores secrets in an encrypted SQLite vault (.tene/vault.db). When you run tene run -- claude, secrets are injected as environment variables at runtime. The AI agent sees the tene run command in CLAUDE.md, but never sees the actual secret values.",
  },
  {
    question: "What is Tene?",
    answer:
      "Tene is a local-first, encrypted secret management CLI. It stores your API keys, tokens, and credentials in an encrypted vault on your device. Single binary, no runtime needed, no server, no signup.",
  },
  {
    question: "How do I install Tene?",
    answer:
      "Run: curl -sSfL https://tene.sh/install.sh | sh \u2014 it auto-detects your OS and installs the latest binary. Works on macOS, Linux, and Windows (WSL). No Go required.",
  },
  {
    question: "Is Tene free?",
    answer:
      "Yes, Tene CLI is 100% free and open source under the MIT license. Local encrypted secret management has no limits. Cloud sync for teams and multi-project is coming at $1/user/month.",
  },
  {
    question: "Which AI Agents does Tene support?",
    answer:
      "Tene supports Claude Code, Cursor, Windsurf, Gemini, and Codex. When you run tene init, it auto-generates context files for each editor (CLAUDE.md, .cursor/rules/tene.mdc, .windsurfrules, GEMINI.md, AGENTS.md). Each AI editor reads its file and knows how to use tene automatically — no manual setup needed.",
  },
  {
    question: "What encryption does Tene use?",
    answer:
      "XChaCha20-Poly1305 for secret encryption with 192-bit random nonces. Argon2id (64MB memory, 3 iterations) for key derivation. Master key stored in your OS keychain. 12-word BIP-39 recovery key.",
  },
  {
    question: "What is Cloud sync?",
    answer:
      "Cloud sync (coming soon, $1/user/month) lets you manage secrets across multiple projects and machines without running tene init and tene set every time. Your secrets stay encrypted end-to-end.",
  },
];
