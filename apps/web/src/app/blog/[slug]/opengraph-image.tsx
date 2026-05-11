// Design Ref: blog-seo-enhancements §2.3 — G3 per-article OG image.
// Next.js 16 file-convention: this route auto-wires <meta property="og:image">
// for the [slug] dynamic segment. No manual metadata.openGraph.images needed.
import { ImageResponse } from "next/og";
import { getAllPostSlugs, getPostBySlug } from "@/lib/blog";
import { getTagLabel } from "@/lib/tags";

// Must run on Node.js (not Edge) — getPostBySlug uses fs to read .mdx files.
export const runtime = "nodejs";
export const contentType = "image/png";
export const size = { width: 1200, height: 630 };

export function generateStaticParams() {
  return getAllPostSlugs().map((slug) => ({ slug }));
}

export default async function Image({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const post = getPostBySlug(slug);
  const title = post?.meta.title ?? "tene Tech Blog";
  const tags = post?.meta.tags ?? [];
  const primaryTag = tags[0] ? getTagLabel(tags[0]) : undefined;
  const readingMinutes = post?.meta.readingMinutes;

  return new ImageResponse(
    (
      <div
        style={{
          height: "100%",
          width: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          padding: "60px 72px",
          background: "#0a0a0a",
          color: "#ededed",
          fontFamily: "sans-serif",
        }}
      >{/* noindex header is set on the ImageResponse options below */}
        {/* Dot grid background overlay (subtle) */}
        <div
          style={{
            position: "absolute",
            inset: 0,
            display: "flex",
            backgroundImage:
              "radial-gradient(circle, #2a2a2a 1px, transparent 1px)",
            backgroundSize: "24px 24px",
            opacity: 0.3,
          }}
        />

        {/* Header */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "14px",
            fontSize: "28px",
            color: "#888888",
            zIndex: 1,
          }}
        >
          <div
            style={{
              width: "44px",
              height: "44px",
              background: "#1e1e1e",
              borderRadius: "8px",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              color: "#00ff88",
              fontWeight: 900,
              fontSize: "28px",
            }}
          >
            t
          </div>
          <div style={{ display: "flex", fontFamily: "monospace" }}>
            tene Tech Blog
          </div>
        </div>

        {/* Title */}
        <div
          style={{
            display: "flex",
            fontSize: title.length > 60 ? "56px" : title.length > 40 ? "64px" : "72px",
            fontWeight: 800,
            lineHeight: 1.15,
            letterSpacing: "-0.02em",
            color: "#ededed",
            maxWidth: "1050px",
            zIndex: 1,
          }}
        >
          {title}
        </div>

        {/* Footer: primary tag · reading time · brand */}
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            fontSize: "28px",
            color: "#888888",
            zIndex: 1,
          }}
        >
          <div style={{ display: "flex", gap: "14px", alignItems: "center" }}>
            {primaryTag && (
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  padding: "10px 20px",
                  border: "1px solid #2a2a2a",
                  borderRadius: "9999px",
                  background: "#141414",
                }}
              >
                <span style={{ color: "#00ff88", marginRight: "8px" }}>#</span>
                {primaryTag}
              </div>
            )}
            {readingMinutes && (
              <div style={{ display: "flex", color: "#555555" }}>
                {readingMinutes} min read
              </div>
            )}
          </div>
          <div style={{ display: "flex", color: "#00ff88", fontWeight: 600 }}>
            tene.sh
          </div>
        </div>
      </div>
    ),
    {
      ...size,
      // X-Robots-Tag: noindex — this route returns an image/png stream,
      // not an HTML document. Without this header, GSC indexes the URL
      // as a "Crawled — currently not indexed" page (the dashboard
      // surfaces it under /blog/{slug}/opengraph-image?... entries).
      // The header re-classifies it as "Excluded by 'noindex' tag",
      // which is the correct bucket. See .claude/rules/blog-content.md
      // §10.1 and verify-blog-indexability.mjs route-noindex assertion.
      headers: { "X-Robots-Tag": "noindex" },
    },
  );
}
