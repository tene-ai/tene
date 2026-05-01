import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { Nav } from "@/components/nav";
import { Footer } from "@/components/footer";
import { InteractiveGrid } from "@/components/interactive-grid";
import { ComparisonHero } from "@/components/vs/comparison-hero";
import { ComparisonTable } from "@/components/vs/comparison-table";
import { ComparisonNarrative } from "@/components/vs/comparison-narrative";
import { MigrationGuide } from "@/components/vs/migration-guide";
import { ComparisonFaqList } from "@/components/vs/comparison-faq";
import { RelatedComparisons } from "@/components/vs/related-comparisons";
import { ComparisonJsonLd } from "@/components/seo/software-jsonld";
import { Breadcrumb } from "@/components/breadcrumb";
import {
  getAllComparisonSlugs,
  getComparison,
  getRelatedComparisons,
} from "@/data/comparisons";
import { toIsoDateTime } from "@/lib/iso-date";

// Static at build time — no database, no request-time work. Every slug is
// known ahead of time, so Next.js renders these pages as static HTML with
// Schema.org JSON-LD baked in. Satisfies NFR-02 (Lighthouse Performance ≥ 90)
// and NFR-03 (Lighthouse SEO = 100) via generateStaticParams + metadata.
export const dynamic = "error";

type Params = { slug: string };

export async function generateStaticParams(): Promise<Params[]> {
  return getAllComparisonSlugs().map((slug) => ({ slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<Params>;
}): Promise<Metadata> {
  const { slug } = await params;
  const data = getComparison(slug);
  if (!data) return {};

  const canonical = `https://tene.sh/vs/${slug}`;

  return {
    title: data.metaTitle,
    description: data.metaDescription,
    keywords: data.heroKeywords,
    alternates: { canonical },
    openGraph: {
      title: data.metaTitle,
      description: data.metaDescription,
      url: canonical,
      siteName: "Tene",
      type: "article",
      publishedTime: toIsoDateTime(data.publishedAt),
      modifiedTime: toIsoDateTime(data.updatedAt),
      images: [
        {
          url: "/og-image.webp",
          width: 1200,
          height: 630,
          alt: `tene vs ${data.competitorName}`,
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      title: data.metaTitle,
      description: data.metaDescription,
      images: ["/og-image.webp"],
    },
    robots: {
      index: true,
      follow: true,
      googleBot: { index: true, follow: true },
    },
  };
}

export default async function ComparisonPage({
  params,
}: {
  params: Promise<Params>;
}) {
  const { slug } = await params;
  const data = getComparison(slug);
  if (!data) notFound();

  const related = getRelatedComparisons(slug, 3);
  const canonical = `https://tene.sh/vs/${slug}`;

  return (
    <>
      <ComparisonJsonLd comparison={data} pageUrl={canonical} />

      {/* Same background system as the home page: canvas dot grid on desktop,
          CSS fallback on mobile. Keeps the /vs/* pages visually consistent
          with `/` so a visitor landing here first doesn't see a different
          site. */}
      <InteractiveGrid />
      <div className="dot-grid-fixed sm:hidden" />

      <Nav />
      <main className="relative z-10">
        <Breadcrumb
          items={[
            { label: "Home", href: "/" },
            { label: "Compare", href: "/vs" },
            { label: `vs ${data.competitorName}` },
          ]}
        />
        <ComparisonHero comparison={data} />
        <ComparisonNarrative
          intro={data.intro}
          sections={data.sections.slice(0, 1)}
        />
        <ComparisonTable
          competitorName={data.competitorName}
          rows={data.comparisonRows}
        />
        <ComparisonNarrative
          intro=""
          sections={data.sections.slice(1)}
        />
        {data.migration && <MigrationGuide migration={data.migration} />}
        <ComparisonFaqList faqs={data.faqs} />
        <RelatedComparisons items={related} />
      </main>
      <Footer />
    </>
  );
}
