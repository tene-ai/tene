import type { NextConfig } from "next";

// Content-Security-Policy composed from named directives so the analytics
// allowlist is auditable. GA4 needs:
//   - script-src:  https://www.googletagmanager.com (gtag.js loader)
//   - connect-src: https://www.google-analytics.com + https://analytics.google.com
//   - img-src:     https://www.google-analytics.com (GA's 1x1 pixel fallback)
//     (already covered by generic `https:`)
const csp = [
  "default-src 'self'",
  "script-src 'self' 'unsafe-inline' 'unsafe-eval' https://www.googletagmanager.com",
  "style-src 'self' 'unsafe-inline'",
  "img-src 'self' data: https:",
  "connect-src 'self' https://www.google-analytics.com https://analytics.google.com https://www.googletagmanager.com",
  "font-src 'self' data:",
].join("; ");

const securityHeaders = [
  { key: 'Strict-Transport-Security', value: 'max-age=63072000; includeSubDomains; preload' },
  { key: 'X-Frame-Options', value: 'DENY' },
  { key: 'X-Content-Type-Options', value: 'nosniff' },
  { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
  { key: 'Content-Security-Policy', value: csp },
  { key: 'Permissions-Policy', value: 'camera=(), microphone=(), geolocation=()' },
];

const nextConfig: NextConfig = {
  // Hide Next.js fingerprint from response headers (no SEO/AEO benefit, minor
  // security hygiene — bots still detect the framework via other signals).
  poweredByHeader: false,

  async headers() {
    return [{ source: '/:path*', headers: securityHeaders }];
  },

  async redirects() {
    return [
      // www.tene.sh → apex (canonical is no-www). Keeps Google from picking
      // www as the indexed host. Middleware drops www from ALLOWED_HOSTS so
      // any request that slips past this redirect (shouldn't happen on
      // Vercel) still gets noindex via the middleware fallback.
      {
        source: '/:path*',
        has: [{ type: 'host', value: 'www.tene.sh' }],
        destination: 'https://tene.sh/:path*',
        permanent: true,
      },
      // Retired blog tags — see RETIRED_TAG_REDIRECTS in src/lib/tags.ts.
      // External links (Wayback Machine, Daily.dev posts, GeekNews) still
      // reference these. 301 keeps PageRank flowing into the new home.
      // MUST MATCH src/lib/tags.ts:76 — adding a tag to that map requires
      // a redirect entry here too.
      {
        source: '/blog/tag/ai',
        destination: '/blog/category/vibe-coding',
        permanent: true,
      },
      {
        source: '/blog/tag/go',
        destination: '/blog/tag/architecture',
        permanent: true,
      },
      {
        source: '/blog/tag/vibe-coding',
        destination: '/blog/category/vibe-coding',
        permanent: true,
      },
    ];
  },
};

export default nextConfig;
