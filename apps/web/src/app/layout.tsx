import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { NoiseOverlay } from "@/components/noise-overlay";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Tene — Your .env is not a secret. AI can read it.",
  description:
    "Your .env is not a secret — AI can read it. Tene encrypts secrets locally and injects them at runtime so AI agents never see the values. XChaCha20-Poly1305 encryption. No server, no signup, free and open source.",
  keywords: [
    "secret management",
    "API key management",
    "Claude Code",
    "AI agent",
    "CLI tool",
    "encryption",
    "XChaCha20-Poly1305",
    "local-first",
    "open source",
    "Go",
    "developer tools",
    "vibe coding",
    "CLAUDE.md",
    "environment variables",
    ".env alternative",
  ],
  authors: [{ name: "tomo-kay", url: "https://github.com/tomo-kay" }],
  creator: "tomo-kay",
  metadataBase: new URL("https://tene.sh"),
  alternates: {
    canonical: "https://tene.sh",
  },
  icons: {
    icon: "/favicon.svg",
    apple: "/apple-touch-icon.png",
  },
  openGraph: {
    title: "Tene — Your .env is not a secret. AI can read it.",
    description:
      "Tene encrypts secrets locally and injects them at runtime so AI agents never see the values. No server, no signup, free.",
    url: "https://tene.sh",
    siteName: "Tene",
    type: "website",
    images: [
      {
        url: "/og-image.png",
        width: 1200,
        height: 630,
        alt: "Tene — Your .env is not a secret. AI can read it.",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Tene — Your .env is not a secret. AI can read it.",
    description:
      "Tene encrypts secrets locally and injects them at runtime so AI agents never see the values. No server, no signup, free.",
    images: ["/og-image.png"],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
    },
  },
};

const jsonLd = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "SoftwareApplication",
      name: "Tene",
      applicationCategory: "DeveloperApplication",
      operatingSystem: "macOS, Linux, Windows (WSL)",
      description:
        "Your .env is not a secret — AI can read it. Tene encrypts secrets locally and injects them at runtime so AI agents never see the values.",
      url: "https://tene.sh",
      offers: {
        "@type": "Offer",
        price: "0",
        priceCurrency: "USD",
      },
      license: "https://opensource.org/licenses/MIT",
    },
    {
      "@type": "FAQPage",
      mainEntity: [
        {
          "@type": "Question",
          name: "What is Tene?",
          acceptedAnswer: {
            "@type": "Answer",
            text: "Tene is a local-first, encrypted secret management CLI. It stores your API keys, tokens, and credentials in an encrypted SQLite vault on your device. No server, no signup, no cloud dependency.",
          },
        },
        {
          "@type": "Question",
          name: "How does Claude Code auto-detection work?",
          acceptedAnswer: {
            "@type": "Answer",
            text: "When you run tene init, it generates a CLAUDE.md file in your project root. Claude Code reads this file automatically and learns how to use tene to retrieve secrets.",
          },
        },
        {
          "@type": "Question",
          name: "Is Tene free?",
          acceptedAnswer: {
            "@type": "Answer",
            text: "Yes, Tene CLI is 100% free and open source under the MIT license. Local encrypted secret management has no limits. Cloud sync for multi-machine access starts at $5/month, and team sharing at $10/user/month.",
          },
        },
        {
          "@type": "Question",
          name: "How are my secrets encrypted?",
          acceptedAnswer: {
            "@type": "Answer",
            text: "Tene uses XChaCha20-Poly1305 encryption with 256-bit keys derived from your master password via Argon2id. Each secret gets a unique 192-bit nonce.",
          },
        },
        {
          "@type": "Question",
          name: "Does Tene work offline?",
          acceptedAnswer: {
            "@type": "Answer",
            text: "Tene is 100% offline. It makes zero network calls. Your secrets are encrypted and stored locally in a SQLite database.",
          },
        },
      ],
    },
    {
      "@type": "HowTo",
      name: "How to use Tene for secret management",
      step: [
        {
          "@type": "HowToStep",
          name: "Install",
          text: "Run: curl -sSfL https://tene.sh/install.sh | sh — or download from GitHub Releases (github.com/tomo-kay/tene/releases)",
        },
        {
          "@type": "HowToStep",
          name: "Initialize",
          text: "Run tene init to create an encrypted vault and generate context files for Claude, Cursor, Windsurf, Gemini, and Codex.",
        },
        {
          "@type": "HowToStep",
          name: "Store secrets",
          text: "Run tene set KEY value to encrypt and store secrets locally.",
        },
        {
          "@type": "HowToStep",
          name: "Develop with secrets",
          text: "Run tene run -- your-command to inject secrets as environment variables.",
        },
      ],
    },
  ],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <head>
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
        />
      </head>
      <body className="min-h-full flex flex-col">
        {children}
        <NoiseOverlay />
      </body>
    </html>
  );
}
