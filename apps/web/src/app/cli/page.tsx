import type { Metadata } from "next";
import fs from "node:fs";
import path from "node:path";
import { MDXRemote } from "next-mdx-remote/rsc";
import remarkGfm from "remark-gfm";
import rehypeSlug from "rehype-slug";
import rehypeAutolinkHeadings from "rehype-autolink-headings";
import rehypeShiki from "@shikijs/rehype";
import { Nav } from "@/components/nav";
import { Footer } from "@/components/footer";
import { InteractiveGrid } from "@/components/interactive-grid";
import { Breadcrumb } from "@/components/breadcrumb";

export const dynamic = "error";

export const metadata: Metadata = {
  title: "tene CLI Reference — tene",
  description:
    "Complete command-by-command reference for tene: flags, exit codes, JSON schemas, and examples.",
  alternates: { canonical: "https://tene.sh/cli" },
  openGraph: {
    title: "tene CLI Reference",
    description:
      "Every tene command, flag, exit code, and JSON schema in one place.",
    url: "https://tene.sh/cli",
    siteName: "Tene",
    type: "article",
    images: [{ url: "/og-image.webp", width: 1200, height: 630 }],
  },
  twitter: {
    card: "summary_large_image",
    title: "tene CLI Reference",
    description:
      "Every tene command, flag, exit code, and JSON schema in one place.",
    images: ["/og-image.webp"],
  },
  robots: { index: true, follow: true },
};

// Load docs/cli-reference.md at build time. The apps/web workspace sits in
// apps/web/ so we traverse up twice to reach the repo root where `docs/`
// lives.
async function loadReference(): Promise<string> {
  const candidates = [
    path.join(process.cwd(), "..", "..", "docs", "cli-reference.md"),
    path.join(process.cwd(), "docs", "cli-reference.md"),
  ];
  for (const p of candidates) {
    try {
      return fs.readFileSync(p, "utf-8");
    } catch {
      // try next candidate
    }
  }
  throw new Error(
    `cli-reference.md not found. Looked in:\n${candidates.join("\n")}`,
  );
}

export default async function CliReferencePage() {
  const source = await loadReference();

  return (
    <>
      <InteractiveGrid />
      <div className="dot-grid-fixed sm:hidden" />
      <Nav />
      <main className="relative z-10">
        <Breadcrumb
          items={[
            { label: "Home", href: "/" },
            { label: "CLI Reference" },
          ]}
        />
        <article className="mx-auto max-w-3xl px-4 pt-4 pb-12 sm:px-6 prose prose-invert prose-headings:scroll-mt-24">
          <MDXRemote
            source={source}
            options={{
              parseFrontmatter: false,
              mdxOptions: {
                remarkPlugins: [remarkGfm],
                rehypePlugins: [
                  rehypeSlug,
                  [
                    rehypeAutolinkHeadings,
                    {
                      behavior: "append",
                      properties: {
                        className: ["heading-anchor"],
                        ariaLabel: "Link to section",
                      },
                    },
                  ],
                  [rehypeShiki, { theme: "github-dark" }],
                ],
              },
            }}
          />
        </article>
      </main>
      <Footer />
    </>
  );
}
