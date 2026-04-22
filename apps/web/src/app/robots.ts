import type { MetadataRoute } from "next";

// Allows every major crawler (Google, Bing) and LLM crawler that respects
// robots.txt (GPTBot, ClaudeBot, PerplexityBot, etc.). The llms.txt / llms-full.txt
// files are explicitly linked so LLM crawlers that discover robots.txt first
// can jump straight to the agent-readable index.
export default function robots(): MetadataRoute.Robots {
  const base = "https://tene.sh";
  return {
    rules: [
      {
        userAgent: "*",
        allow: "/",
      },
    ],
    sitemap: `${base}/sitemap.xml`,
    host: base,
  };
}
