# Tene Landing Page (tene.sh)

@AGENTS.md

## Overview
Marketing landing page for Tene -- Agentic Secret Runtime.
Deployed to Vercel at tene.sh.

## Tech Stack
- Next.js 15 (App Router)
- Tailwind CSS v4
- Geist Sans + Geist Mono fonts
- Vercel deployment

## Design System
- Dark-only theme: background #0a0a0a, foreground #ededed
- Accent: #00ff88 (neon green), dim: #00cc6a
- Surface: #141414, #1e1e1e
- Border: #2a2a2a, Muted: #888888

## Development
```bash
cd apps/web
npm install
npm run dev  # port 3000
```

## Public Assets
- /og-image.png -- Open Graph image (1200x630)
- /favicon.svg -- Browser favicon
- /logo.svg -- Tene logo
- /llms.txt -- AI agent discoverability
- /install.sh -- CLI installer script
- /robots.txt + /sitemap.xml -- SEO

## SEO
- Structured data: SoftwareApplication + FAQPage + HowTo (JSON-LD in layout.tsx)
- Meta tags in layout.tsx metadata export
- Domain: tene.sh
