# Static file server (`serve-static.mjs`)

## Purpose

`scripts/serve-static.mjs` is a **zero-dependency** (beyond Node itself) static HTTP server used to:

- Serve `examples/` and `dist/` from the repository root during **Playwright** runs (`webServer` in `playwright.config.ts`).
- Offer a lightweight alternative to Vite for “open this folder as a site” during local review (`package.json` script `docs-serve`).

It is **not** a production application server — no routing framework, no compression middleware, no HTTPS.

## Behaviour

- Binds to `127.0.0.1` and a configurable `--port` (default `4173`).
- Serves files under `--root` (default `.`), with path traversal rejected.
- Maps extensions to `Content-Type`; serves `index.html` inside directories.
- Sets `Cache-Control: no-store` for predictable local testing.

## Example

```bash
node scripts/serve-static.mjs --port 4173 --root .
```

Then browse to `/examples/plain/` relative to that root.
