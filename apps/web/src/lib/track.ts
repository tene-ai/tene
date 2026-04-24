// Design Ref: analytics-ga4-swap — Type-safe analytics event wrapper.
// In production, forwards to GA4 via @next/third-parties `sendGAEvent`.
// In dev/SSR, logs to console.debug so nothing hits analytics endpoints.
import { sendGAEvent } from "@next/third-parties/google";

type EventMap = {
  // Site-wide conversions / navigation (legacy names kept stable).
  install_copy: {
    source: "hero" | "cta" | "vs_page" | "blog_post" | "pricing";
  };
  vs_card_click: { competitor: string };
  github_click: {
    location:
      | "nav"
      | "hero"
      | "footer"
      | "vs_page"
      | "blog_post"
      | "cta"
      | "pricing"
      | "security";
  };
  related_vs_click: { from: string; to: string };
  // Which nav menu item was clicked — answers "어떤 메뉴" analysis question.
  nav_click: {
    item:
      | "features"
      | "how_it_works"
      | "security"
      | "compare"
      | "blog"
      | "faq";
    device: "desktop" | "mobile";
  };

  // tech-blog Phase 2 (FR-32 ~ FR-37)
  blog_copy_code: { slug: string; language: string };
  // blog-categories-and-tooling — tag filter now fires from the multi-select
  // filter widget too, and sometimes without a specific tag (clear all).
  blog_tag_filter: {
    tag?: string;
    action?: "add" | "remove" | "clear_all";
    from: "index" | "post_header" | "card" | "filter";
  };
  // blog-categories-and-tooling — category pill click on /blog index.
  blog_category_filter: {
    category: string;
    from: "index" | "active";
  };
  blog_related_click: { fromSlug: string; toSlug: string };
  blog_rss_click: { location: "footer" | "blog_header" };
  blog_external_link: { slug: string; domain: string };
  blog_read_complete: { slug: string; readingMinutes: number };
  blog_share_click: { slug: string; channel: string };
  // blog-tts — Web Speech API based article TTS player.
  blog_tts_play: {
    slug: string;
    readingMinutes: number;
    voice?: string;
    rate?: number;
  };
  blog_tts_pause: { slug: string; percentRead: number };
  blog_tts_resume: { slug: string; percentRead: number };
  blog_tts_stop: { slug: string; percentRead: number };
  blog_tts_complete: { slug: string; readingMinutes: number };
  blog_tts_rate_change: { slug: string; rate: number };
  blog_tts_voice_change: { slug: string; voiceName: string };
};

export function track<E extends keyof EventMap>(
  event: E,
  payload: EventMap[E],
): void {
  if (typeof window === "undefined") return; // SSR safety
  if (process.env.NODE_ENV !== "production") {
    // eslint-disable-next-line no-console
    console.debug("[analytics]", event, payload);
    return;
  }
  // GA4 accepts an event name + a flat record of parameters.
  sendGAEvent("event", event, payload as Record<string, unknown>);
}
