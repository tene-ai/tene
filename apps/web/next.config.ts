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
  async headers() {
    return [{ source: '/:path*', headers: securityHeaders }];
  },
};

export default nextConfig;
