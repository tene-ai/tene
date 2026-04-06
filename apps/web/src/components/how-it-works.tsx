import { CopyCommand } from "./copy-command";

const steps = [
  {
    step: "01",
    title: "Install",
    command: "curl -sSfL https://tene.sh/install.sh | sh",
    description: "One command. Auto-detects your OS. No Go required, no account, no server.",
  },
  {
    step: "02",
    title: "Project Initialize",
    command: "tene init",
    description:
      "Navigate to your project folder first. Creates an encrypted vault, generates CLAUDE.md for Claude Code, and issues a 12-word recovery key.",
  },
  {
    step: "03",
    title: "Store secrets",
    command: "tene set STRIPE_KEY sk_test_xxx",
    description:
      "Secrets are encrypted with XChaCha20-Poly1305 and stored in a local SQLite vault. Never leaves your machine.",
  },
  {
    step: "04",
    title: "Develop with secrets",
    command: "tene run -- claude",
    description:
      "Injects all secrets as environment variables into any command. Claude Code reads CLAUDE.md and knows the rest.",
  },
];

export function HowItWorks() {
  return (
    <section id="how-it-works" className="px-4 py-24 sm:px-6">
      <div className="mx-auto max-w-3xl">
        <h2 className="text-center text-3xl font-bold sm:text-4xl">
          Up and running in{" "}
          <span className="text-accent">1 minute</span>
        </h2>

        <div className="mt-16 space-y-12">
          {steps.map((s, i) => (
            <div key={s.step} className="relative flex gap-6">
              <div className="flex flex-col items-center">
                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-accent/40 bg-accent/10 text-sm font-bold text-accent">
                  {s.step}
                </div>
                {i < steps.length - 1 && (
                  <div className="mt-2 h-full w-px bg-border" />
                )}
              </div>
              <div className="pb-12">
                <h3 className="text-xl font-semibold">{s.title}</h3>
                <CopyCommand command={s.command} className="relative mt-3 text-xs sm:text-sm" />
                <p className="mt-3 text-sm leading-relaxed text-muted">
                  {s.description}
                </p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
