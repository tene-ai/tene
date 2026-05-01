import type { MetadataRoute } from "next";

// Explicit allow-list for LLM crawlers so conservative bots that default to
// deny-on-missing-rule don't skip us. Also explicitly disallows /.tene/
// (the encrypted vault directory that must never be served), /api/
// (server actions), and the Bytespider crawler (high-volume, low-signal).
//
// Bot landscape (2026-mid). Each provider runs distinct user agents for
// training-time crawls vs retrieval-time fetches. Both are allowed because
// the goal is being cited in AI answers, not blocking.
//
//   OpenAI
//     - GPTBot           training corpus crawler
//     - ChatGPT-User     ChatGPT browse user-action fetcher
//     - OAI-SearchBot    powers ChatGPT search index
//
//   Anthropic
//     - ClaudeBot        training corpus crawler
//     - Claude-Web       legacy training crawler
//     - anthropic-ai     legacy training crawler
//     - Claude-SearchBot powers claude.ai web search (added 2026-04)
//     - Anthropic-User   Claude tool-call WebFetch (Claude Code etc.)
//
//   Google
//     - Googlebot        search index
//     - Google-Extended  AI training opt-out toggle (Bard/Gemini training)
//     - GoogleOther      generic crawler used by AI Overviews offline
//
//   Microsoft / Bing (powers ChatGPT search per Seer Interactive 2025: 87%
//   citation match with Bing top-5)
//     - Bingbot
//     - BingPreview
//
//   Perplexity
//     - PerplexityBot    index crawler
//     - Perplexity-User  Sonar tool-call fetcher
//
//   Other major crawlers
//     - CCBot              CommonCrawl (foundation for many open LLMs)
//     - Applebot-Extended  Apple Intelligence training opt-out
//     - Meta-ExternalAgent Meta AI fetcher
//     - Amazonbot          Alexa/Rufus
//     - YouBot             you.com
//     - cohere-ai          Cohere training
const COMMON_DISALLOW = ["/.tene/", "/api/"];

export default function robots(): MetadataRoute.Robots {
  const base = "https://tene.sh";
  return {
    rules: [
      // OpenAI — training (GPTBot) + retrieval (ChatGPT-User, OAI-SearchBot)
      {
        userAgent: ["GPTBot", "ChatGPT-User", "OAI-SearchBot"],
        allow: "/",
        disallow: COMMON_DISALLOW,
      },
      // Anthropic — training + retrieval (Claude-SearchBot powers
      // claude.ai web search; Anthropic-User powers Claude Code WebFetch)
      {
        userAgent: [
          "ClaudeBot",
          "Claude-Web",
          "anthropic-ai",
          "Claude-SearchBot",
          "Anthropic-User",
        ],
        allow: "/",
        disallow: COMMON_DISALLOW,
      },
      // Google — training + search (GoogleOther serves AI Overviews offline
      // pipeline and is distinct from Googlebot)
      {
        userAgent: ["Google-Extended", "Googlebot", "GoogleOther"],
        allow: "/",
        disallow: COMMON_DISALLOW,
      },
      // Microsoft Bing — chokepoint for ChatGPT search
      {
        userAgent: ["Bingbot", "BingPreview"],
        allow: "/",
        disallow: COMMON_DISALLOW,
      },
      // Perplexity — index + Sonar retrieval
      {
        userAgent: ["PerplexityBot", "Perplexity-User"],
        allow: "/",
        disallow: COMMON_DISALLOW,
      },
      // Other major crawlers
      {
        userAgent: [
          "CCBot",
          "Applebot-Extended",
          "Meta-ExternalAgent",
          "Amazonbot",
          "YouBot",
          "cohere-ai",
        ],
        allow: "/",
        disallow: COMMON_DISALLOW,
      },
      // Bytespider — high-volume, low-signal training crawler
      { userAgent: "Bytespider", disallow: "/" },
      // Catchall — allow but consistently disallow vault and API paths
      { userAgent: "*", allow: "/", disallow: COMMON_DISALLOW },
    ],
    sitemap: [`${base}/sitemap.xml`, `${base}/blog/rss.xml`],
    host: base,
  };
}
