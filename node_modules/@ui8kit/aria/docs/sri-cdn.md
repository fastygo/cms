# SRI and CDN usage

## What Subresource Integrity is

**Subresource Integrity (SRI)** is a browser mechanism: when you load a script (or stylesheet) with an `integrity` attribute, the browser hashes the **response body** and compares it to the expected value. If the bytes differ — wrong version, corrupted download, or malicious swap — the resource is **not executed**.

Typical tag shape:

```html
<script
  src="https://cdn.jsdelivr.net/npm/@ui8kit/aria@1.2.3/dist/all.iife.min.js"
  integrity="sha384-…base64…"
  crossorigin="anonymous"
  defer
></script>
```

`crossorigin="anonymous"` is required for integrity checks on cross-origin fetches unless the CDN sends appropriate CORS headers for credentialed requests (anonymous is the usual case for public CDNs).

## CDN workflow

1. **Pin the version** in the URL (`@1.2.3`, not floating `@latest` in production).
2. **Build** the same artifact you will ship (`dist/all.iife.min.js` or the file you actually reference).
3. **Compute** the SRI hash for that exact file.
4. Paste `integrity="sha384-…"` into your template or CMS.

Browsers accept several algorithms; this project standardises on **SHA-384** in generated metadata for a sensible balance of digest size and support.

## Role of `scripts/generate-sri.mjs`

The script is **not** part of the runtime library. It is a **release / docs helper** that runs after `tsup` has written `dist/`:

1. Enumerates every `*.js` file in `dist/` (excluding `*.map`).
2. Reads each file as bytes, computes **SHA-384**, base64-encodes the digest, and prefixes with `sha384-` per the SRI format.
3. Writes `dist/sri.json` — a JSON map from **filename → full integrity string**.

Consumers (humans or automation) copy values from `sri.json` into HTML, static site generators, or internal wikis. The root [`README.md`](../README.md) illustrates the hand-off: run `npm run sri` (or `bun run sri`) after build, then substitute the hash in the CDN snippet.

There is **no dependency** on the `object.fromentries` npm package; the script uses built-in `Object.fromEntries`. Any `object.fromentries` folder under `node_modules` comes from other tooling (e.g. ESLint), not from this script.

## Operational notes

- Regenerate SRI whenever the **pinned file** changes (every release that touches `dist/`).
- If you concatenate or minify further **after** `generate-sri`, the hash must be computed on the **final** bytes users load.
- SRI does not help if you load **different** URLs per environment without updating the hash — keep CI and production URLs aligned with the same version pin.
