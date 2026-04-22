import type { Metadata } from "next";
import { Nav } from "@/components/nav";
import { Footer } from "@/components/footer";
import { InteractiveGrid } from "@/components/interactive-grid";
import { VsCardLink } from "@/components/vs/vs-card-link";
import { comparisons } from "@/data/comparisons";

export const metadata: Metadata = {
  title: "Compare tene to every major secret manager",
  description:
    "Head-to-head comparisons of tene vs Doppler, dotenv, dotenv-vault, Infisical, and HashiCorp Vault. Pricing, architecture, AI-agent integration, and migration paths.",
  alternates: { canonical: "https://tene.sh/vs" },
  openGraph: {
    title: "Compare tene to every major secret manager",
    description:
      "Head-to-head comparisons: tene vs Doppler, dotenv, dotenv-vault, Infisical, and HashiCorp Vault.",
    url: "https://tene.sh/vs",
    siteName: "Tene",
    type: "website",
    images: [
      {
        url: "/og-image.png",
        width: 1200,
        height: 630,
        alt: "Compare tene to every major secret manager",
      },
    ],
  },
  robots: {
    index: true,
    follow: true,
  },
};

export default function ComparisonIndex() {
  return (
    <>
      <InteractiveGrid />
      <div className="dot-grid-fixed sm:hidden" />

      <Nav />
      <main className="relative z-10">
        <section className="px-4 pt-28 pb-8 sm:px-6">
          <div className="mx-auto max-w-4xl text-center">
            <h1 className="text-3xl font-bold leading-tight tracking-tight sm:text-4xl md:text-5xl">
              tene vs everything else
            </h1>
            <p className="mx-auto mt-6 max-w-2xl text-base text-muted leading-relaxed sm:text-lg">
              Honest side-by-side comparisons against the major secret
              managers. No FUD, no marketing — just the facts about pricing,
              architecture, AI-agent integration, and migration paths.
            </p>
          </div>
        </section>

        <section className="px-4 py-12 sm:px-6">
          <ul className="mx-auto grid max-w-4xl gap-4 sm:grid-cols-2">
            {comparisons.map((c) => (
              <li key={c.slug}>
                <VsCardLink
                  slug={c.slug}
                  competitorName={c.competitorName}
                  subheadline={c.subheadline}
                />
              </li>
            ))}
          </ul>
        </section>
      </main>
      <Footer />
    </>
  );
}
