"use client";

import { useState } from "react";

const faqs = [
  {
    question: "What is Tene?",
    answer:
      "Tene is a local-first, encrypted secret management CLI built in Go. It stores your API keys, tokens, and credentials in an encrypted SQLite vault on your device. Single binary, no runtime needed, no server, no signup.",
  },
  {
    question: "How does Claude Code auto-detection work?",
    answer:
      "When you run tene init, it generates a CLAUDE.md file in your project root. Claude Code reads this file automatically and learns how to use tene to retrieve secrets — no manual configuration needed.",
  },
  {
    question: "How do I install Tene?",
    answer:
      "Run: go install github.com/tomo-kay/tene/cmd/tene@latest. Or download a prebuilt binary from GitHub Releases (https://github.com/tomo-kay/tene/releases). Works on macOS, Linux, and Windows (WSL). Single binary, no runtime required.",
  },
  {
    question: "Is Tene free?",
    answer:
      "Yes, Tene is 100% free and open source under the MIT license. There are no paid tiers, no usage limits, and no hidden costs. It runs entirely on your local machine.",
  },
  {
    question: "How are my secrets encrypted?",
    answer:
      "Tene uses XChaCha20-Poly1305 encryption with 256-bit keys derived from your master password via Argon2id (64MB memory, 3 iterations). Each secret gets a unique 192-bit nonce. Your master key is cached in the OS keychain.",
  },
  {
    question: "Can I migrate from .env files?",
    answer:
      "Yes. Run tene import .env to bring all your existing environment variables into the encrypted vault in one command. Your .env file can then be safely deleted.",
  },
  {
    question: "Does Tene work offline?",
    answer:
      "Tene is 100% offline. It makes zero network calls. Your secrets are encrypted and stored locally in a SQLite database. There is no server, no telemetry, and no internet requirement.",
  },
  {
    question: "What happens if I forget my master password?",
    answer:
      "During tene init, you receive a 12-word BIP-39 recovery key. Store it securely — it is the only way to recover your vault if you forget your master password.",
  },
];

export function FAQ() {
  const [openIndex, setOpenIndex] = useState<number | null>(null);

  return (
    <section id="faq" className="px-4 py-24 sm:px-6">
      <div className="mx-auto max-w-3xl">
        <h2 className="text-center text-3xl font-bold sm:text-4xl">
          Frequently asked <span className="text-accent">questions</span>
        </h2>
        <p className="mx-auto mt-4 max-w-xl text-center text-muted">
          Everything you need to know about Tene.
        </p>

        <div className="mt-12 space-y-2">
          {faqs.map((faq, i) => (
            <div
              key={faq.question}
              className="rounded-xl border border-border bg-surface transition-colors hover:border-accent/20"
            >
              <button
                className="flex w-full items-center justify-between px-6 py-5 text-left"
                onClick={() => setOpenIndex(openIndex === i ? null : i)}
                aria-expanded={openIndex === i}
              >
                <span className="pr-4 font-medium">{faq.question}</span>
                <svg
                  className={`h-5 w-5 shrink-0 text-muted transition-transform duration-200 ${
                    openIndex === i ? "rotate-180" : ""
                  }`}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  strokeWidth="2"
                >
                  <path d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              {openIndex === i && (
                <div className="px-6 pb-5 text-sm leading-relaxed text-muted">
                  {faq.answer}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
