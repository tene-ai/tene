import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { GoogleAnalytics } from "@next/third-parties/google";
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
    "Tene encrypts your API keys locally and injects them at runtime so Claude Code, Cursor, and other AI agents never see plaintext. MIT, no server, free.",
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
        url: "/og-image.webp",
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
    images: ["/og-image.webp"],
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
      "@type": "Organization",
      "@id": "https://tene.sh/#organization",
      name: "Tene",
      url: "https://tene.sh",
      logo: {
        "@type": "ImageObject",
        url: "https://tene.sh/logo.svg",
        width: 256,
        height: 256,
      },
      sameAs: [
        "https://github.com/tomo-kay/tene",
        "https://github.com/tomo-kay",
      ],
      contactPoint: {
        "@type": "ContactPoint",
        contactType: "Support",
        url: "https://github.com/tomo-kay/tene/issues",
      },
      founder: {
        "@type": "Person",
        name: "tomo-kay",
        url: "https://github.com/tomo-kay",
      },
    },
    {
      "@type": "WebSite",
      "@id": "https://tene.sh/#website",
      url: "https://tene.sh",
      name: "Tene",
      publisher: { "@id": "https://tene.sh/#organization" },
      potentialAction: {
        "@type": "SearchAction",
        target: {
          "@type": "EntryPoint",
          urlTemplate: "https://tene.sh/blog?q={search_term_string}",
        },
        "query-input": "required name=search_term_string",
      },
      inLanguage: "en-US",
    },
    {
      "@type": "SoftwareApplication",
      "@id": "https://tene.sh/#software",
      name: "Tene",
      applicationCategory: "DeveloperApplication",
      applicationSubCategory: "Secret Management CLI",
      operatingSystem: "macOS, Linux, Windows (WSL)",
      description:
        "Tene encrypts your API keys locally and injects them at runtime so Claude Code, Cursor, and other AI agents never see plaintext. MIT, no server, free.",
      url: "https://tene.sh",
      downloadUrl: "https://tene.sh/install.sh",
      softwareVersion: "latest",
      offers: {
        "@type": "Offer",
        price: "0",
        priceCurrency: "USD",
      },
      license: "https://opensource.org/licenses/MIT",
      author: {
        "@type": "Person",
        name: "tomo-kay",
        url: "https://github.com/tomo-kay",
      },
      publisher: { "@id": "https://tene.sh/#organization" },
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
            text: "Yes, Tene is 100% free and open source under the MIT license. All local features — encryption, runtime injection, multi-environment, AI editor rules — are free forever with no limits.",
          },
        },
        {
          "@type": "Question",
          name: "Will there be team features?",
          acceptedAnswer: {
            "@type": "Answer",
            text: "Team sync and collaboration features are being designed. The goal is encrypted team sync without a central server. Join the waitlist at tene.sh to get notified when it launches.",
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
        <link rel="ai-index" href="https://tene.sh/llms.txt" />
        <link
          rel="alternate"
          type="text/plain"
          title="LLM-optimized summary (llms.txt)"
          href="https://tene.sh/llms.txt"
        />
        <link
          rel="alternate"
          type="text/plain"
          title="LLM-optimized full reference (llms-full.txt)"
          href="https://tene.sh/llms-full.txt"
        />
        {/* RSS feed auto-discovery — emitted on EVERY page (root layout) so
            search-engine crawlers and RSS readers find the feed from the
            homepage, comparison pages, tag pages, etc. Next.js Metadata API
            shallow-replaces `alternates` per page, which would otherwise drop
            the link on routes that set their own canonical. Direct <head>
            injection is the single source of truth. */}
        <link
          rel="alternate"
          type="application/rss+xml"
          title="tene Tech Blog RSS"
          href="https://tene.sh/blog/rss.xml"
        />
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
        />
      </head>
      <body className="min-h-full flex flex-col">
        {children}
        <NoiseOverlay />
      </body>
      {/* Google Analytics 4 via @next/third-parties. Managed async script
          load, gtag() global, dataLayer[]. Measurement ID is injected via
          the production property set up by the team. */}
      <GoogleAnalytics gaId="G-9MRWMY6XBE" />
    </html>
  );
}
