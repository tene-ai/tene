// Design Ref: docs/02-design/features/ai-discoverability.design.md §2.3
// TypeScript data files are used instead of MDX to stay consistent with the
// existing landing-page convention (src/data/*.ts). The semantic content
// (frontmatter + body + migration guide) is preserved — only the format
// differs, and the rendered HTML must still satisfy NFR-02/03 (Lighthouse
// SEO = 100 and Performance ≥ 90).

export type ComparisonRow = {
  dimension: string;
  tene: string;
  competitor: string;
};

export type MigrationStep = {
  title: string;
  command?: string;
  note?: string;
};

export type ComparisonFAQ = {
  question: string;
  answer: string;
};

export type Comparison = {
  // --- routing & SEO -----------------------------------------------------
  slug: string; // URL: /vs/{slug}
  competitorName: string; // Display name: "Doppler", "HashiCorp Vault"
  competitorHomepage?: string; // external link for attribution
  metaTitle: string; // <title>
  metaDescription: string; // <meta description>
  publishedAt: string; // ISO date (initial publish)
  updatedAt: string; // ISO date (last revision)

  // --- hero --------------------------------------------------------------
  headline: string;
  subheadline: string;
  heroKeywords?: string[]; // optional long-tail keywords surfaced in H1/H2

  // --- lead paragraph above the comparison table -------------------------
  intro: string; // 1–2 paragraphs of prose, no HTML; Markdown-lite

  // --- comparison table --------------------------------------------------
  comparisonRows: ComparisonRow[];

  // --- migration guide (optional) ----------------------------------------
  migration?: {
    title: string;
    summary: string;
    steps: MigrationStep[];
    postMigrationNote?: string;
  };

  // --- "Why developers migrate" narrative sections -----------------------
  sections: Array<{
    heading: string;
    body: string; // plain text with \n\n paragraph breaks
  }>;

  // --- FAQ ---------------------------------------------------------------
  faqs: ComparisonFAQ[];

  // --- related comparisons (populated by registry; leave empty) ---------
  related?: string[]; // list of slugs
};
