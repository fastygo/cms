# Security

## Subresource Integrity (SRI)

When you load `dist/*.iife*.js` from a CDN, pair the URL with an `integrity` attribute (see [SRI & CDN](sri-cdn.md)). That mitigates **tampering on the CDN or in transit** relative to the hash you pinned. It does **not**:

- Fix XSS in your own page if you pass untrusted HTML.
- Replace a need for `Content-Security-Policy` if you lock down script sources.

## CSP and `unsafe-inline`

IIFE bundles assign `globalThis.ui8kit`. A strict CSP may require `script-src` allowances for your CDN host and possibly `'unsafe-inline'` only if you inline boot scripts — prefer external files and nonces/hashes for inline bootstraps.

## Supply chain

- Pin **exact** package versions in applications that embed vendor JS.
- Prefer **npm provenance** and lockfiles in CI for reproducible builds.
- Review `package.json` `exports` and `sideEffects` when auditing what runs on import.

## Globals

The `all` entry sets `window.ui8kit`. Any third-party script in the same origin can call it — same threat model as any other global. Prefer ESM + explicit `init` in high-trust apps.
