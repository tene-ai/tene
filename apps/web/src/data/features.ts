// Design Ref: §3.2 — Feature cards with tag-based color coding
export type Feature = {
  icon: string;
  title: string;
  description: string;
  tag: "Problem" | "Solution" | "Coming Soon" | "New" | null;
};

export const features: Feature[] = [
  {
    icon: "eye",
    title: ".env is visible to AI",
    description:
      "AI agents read all your project files \u2014 including .env. Your API keys are sent to AI models as plaintext.",
    tag: "Problem",
  },
  {
    icon: "inject",
    title: "Runtime injection",
    description:
      "tene run injects secrets as environment variables. Your app works normally. AI never sees the values.",
    tag: "Solution",
  },
  {
    icon: "lock",
    title: "Encrypted vault",
    description:
      "XChaCha20-Poly1305 + Argon2id. Secrets stored locally in an encrypted SQLite vault.",
    tag: null,
  },
  {
    icon: "zap",
    title: "One command setup",
    description:
      "Install, init, set \u2014 done. No signup, no config files, no dashboard.",
    tag: null,
  },
  {
    icon: "import",
    title: ".env migration",
    description:
      "tene import .env converts all your existing secrets into the encrypted vault. Zero friction.",
    tag: null,
  },
  {
    icon: "shield",
    title: "AI agent rules",
    description:
      "tene init auto-generates context files for 5 AI editors — Claude Code, Cursor, Windsurf, Gemini, and Codex. AI knows how to use tene safely without configuration.",
    tag: "New",
  },
];
