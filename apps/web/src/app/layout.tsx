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
  authors: [{ name: "agent-kay", url: "https://agentkay.it" }],
  creator: "agent-kay",
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
        "https://agentkay.it",
      ],
      contactPoint: {
        "@type": "ContactPoint",
        contactType: "Support",
        url: "https://github.com/tomo-kay/tene/issues",
      },
      founder: {
        "@type": "Person",
        name: "agent-kay",
        url: "https://agentkay.it",
      },
    },
    {
      "@type": "WebSite",
      "@id": "https://tene.sh/#website",
      url: "https://tene.sh",
      name: "Tene",
      publisher: { "@id": "https://tene.sh/#organization" },
      // SearchAction intentionally NOT declared — /blog has no ?q= search
      // endpoint yet. Asserting one would fail Google's Rich Results
      // validator and trigger "구조화된 데이터 오류". Add SearchAction back
      // once a real search route exists.
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
        name: "agent-kay",
        url: "https://agentkay.it",
      },
      publisher: { "@id": "https://tene.sh/#organization" },
    },
    // FAQPage + HowTo intentionally moved to <HomeJsonLd /> (rendered only
    // on `/`). Emitting them in the root layout duplicated FAQPage on every
    // /blog/{slug} and /vs/{slug} page, which GSC flags as "FAQPage 입력란
    // 중복" and drops from rich results. Site-wide entities only here.
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
