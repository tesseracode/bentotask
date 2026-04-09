# Deployment Modes — Reference Guide

**Date**: 2026-04-09  
**Related**: ADR-001 (tech stack), ADR-004 (REST API)

---

## Overview

BentoTask's architecture cleanly separates the Go API server from the SvelteKit frontend. This means the app can be deployed in multiple configurations without code changes — only build/config changes.

---

## Mode 1: Single Binary (Default)

**How it works**: SvelteKit is built with `adapter-static`, output is copied to `internal/api/static/`, and embedded in the Go binary via `go:embed`. One binary serves both the API and the web UI.

```
bt serve → localhost:7878
           ├── /api/v1/*  → Go API handlers
           └── /*         → Embedded static files (SPA)
```

**When to use**: Local-first single-user, Raspberry Pi, self-hosted, demos.

**Build**: `make build` (builds web + Go)

---

## Mode 2: Split Deployment (Static Frontend + API)

**How it works**: Deploy the SvelteKit static build to any CDN/static host (Vercel, Netlify, Cloudflare Pages, S3, nginx). Deploy the Go API separately.

```
CDN (Vercel/Netlify)         Go Server
├── index.html               └── /api/v1/*
├── _app/chunks/*.js
└── _app/immutable/*.css

Browser → CDN for UI → fetch("https://api.example.com/api/v1/*")
```

**What to change**:
1. In `web/src/lib/api.ts`, change `const BASE = '/api/v1'` to read from an environment variable:
   ```typescript
   const BASE = import.meta.env.VITE_API_URL || '/api/v1';
   ```
2. Set `VITE_API_URL=https://api.example.com/api/v1` at build time
3. On the Go API, configure CORS to allow the CDN's origin (update `middleware.go` `AllowedOrigins`)
4. Deploy Go API with `bt serve --host 0.0.0.0` behind a reverse proxy (nginx/Caddy with TLS)

**When to use**: Multiple users, scale frontend independently, CDN edge caching.

---

## Mode 3: SvelteKit SSR (Server-Side Rendering)

**How it works**: Switch from `adapter-static` to `adapter-node` (or `adapter-vercel`, `adapter-cloudflare`, etc.). SvelteKit runs as a Node.js server with full SSR capabilities. The Go API runs separately.

```
Node Server (SvelteKit SSR)     Go Server
├── Server-renders HTML          └── /api/v1/*
├── Streams responses
└── Handles +page.server.ts

Browser → Node SSR → fetch API → Go API
```

**What to change**:
1. `npm install @sveltejs/adapter-node` (or `-vercel`, `-cloudflare`)
2. In `web/svelte.config.js`, swap the adapter:
   ```javascript
   import adapter from '@sveltejs/adapter-node'; // was adapter-static
   ```
3. Remove or update `web/src/routes/+layout.ts`:
   ```typescript
   // Remove these lines (they disable SSR):
   // export const prerender = false;
   // export const ssr = false;
   ```
4. Optionally add `+page.server.ts` files for server-side data loading (replaces client-side `onMount` fetches with server-side `load` functions)
5. Set the API URL via environment variable on the Node server

**When to use**: SEO requirements, faster initial page loads, server-side data fetching, streaming responses.

**Important**: All existing Svelte components work identically with any adapter. The adapter is a build-time choice, not a code-level one. Switching adapters does NOT require rewriting components.

---

## Mode 4: API Only (Headless)

**How it works**: Run only the Go API. No web UI. Consume via CLI, scripts, or external tools (curl, Postman, custom integrations, MCP server).

```
bt serve → localhost:7878/api/v1/*

curl, scripts, MCP → /api/v1/tasks, /api/v1/suggest, etc.
```

**What to change**: Nothing. The API works independently today. The embedded static files are served on `/*` but don't interfere with `/api/v1/*`. You can ignore them entirely.

**When to use**: Backend for mobile apps, third-party integrations, MCP server (M9), headless automation.

---

## API Independence Guarantee

The Go API (`internal/api/`) has **zero dependencies on the frontend**:

- The API package imports: `app`, `model`, `store`, `engine` — no SvelteKit code
- The `go:embed` is in a single file (`static.go`) that embeds pre-built assets — if the directory is empty or contains a placeholder, the API still compiles and runs
- All API endpoints work identically regardless of which frontend (or no frontend) is used
- The API can be tested independently via `httptest` (315 tests, none require the web UI)
- CORS is configurable — can allow any origin for split deployments

If you ever need to separate the API from the frontend, you can:
1. Deploy the Go binary without the web build (use `make build-go` instead of `make build`)
2. The API is fully functional at `/api/v1/*`
3. Point any frontend (SvelteKit, React, mobile app, script) at it

---

## Quick Reference: Adapter Swap

| Adapter | Install | Output | Serves | Best for |
|---------|---------|--------|--------|----------|
| `adapter-static` | `npm i -D @sveltejs/adapter-static` | `web/build/` (HTML+JS+CSS) | Any static host, `go:embed` | Single binary, CDN |
| `adapter-node` | `npm i -D @sveltejs/adapter-node` | `web/build/` (Node server) | Node.js process | SSR, streaming |
| `adapter-vercel` | `npm i -D @sveltejs/adapter-vercel` | `.vercel/` | Vercel platform | Vercel deployment |
| `adapter-cloudflare` | `npm i -D @sveltejs/adapter-cloudflare` | Workers-compatible | Cloudflare Workers | Edge deployment |
| `adapter-auto` | (default) | Auto-detects platform | Depends on platform | Development/prototyping |

All adapters use the same component code. Only the build output and serving mechanism change.
