# Tene Dashboard

## Overview
Next.js 15 App Router dashboard for Tene Cloud (app.tene.sh).
Zero-Knowledge: never shows secret values, only metadata.

## Tech Stack
- Next.js 15 (App Router)
- Tailwind CSS v4 with Tene design system
- TanStack Query v5 for data fetching
- Zustand for auth state
- Vercel deployment

## Design System
- Dark-only theme (bg: #0a0a0a, fg: #ededed)
- Accent: #00ff88 (neon green)
- Fonts: Geist Sans + Geist Mono
- Components match landing page (apps/web/) branding

## Development
```bash
cd apps/dashboard
npm install
npm run dev  # port 3001
```

## Environment Variables
```
NEXT_PUBLIC_API_URL=https://api.tene.sh
```
