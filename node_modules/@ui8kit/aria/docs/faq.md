# FAQ

## Why not embed this in UI8Kit’s Go repo?

Independent versioning, npm/CDN distribution, and TypeScript tooling (Vitest, ESLint, `tsup`) are simpler in a dedicated JS package. Go apps can still vendor `dist/*.iife.js` via embed.

## Can I use only one pattern from the CDN?

The published **IIFE** entries are `all` and `core` — `all` registers every pattern. For a single pattern from npm, use ESM `import '@ui8kit/aria/dialog'` or the pure API with `registerPattern`.

## Does the library validate my HTML?

No. It binds behaviour to selectors and attributes you provide. Use `html-validate`, axe, and manual keyboard testing.

## `window.ui8kit` vs ESM `getNamespace()`

Same object in the browser after the namespace is created. The pure entry avoids touching globals until you call `getNamespace()`; the `all` IIFE registers patterns and triggers auto-init (unless disabled).

## Why SHA-384 in SRI output?

Strong enough for integrity tags; widely supported. You can recompute with other tools if your organisation mandates SHA-512.

## Playwright fails locally but passes in CI

CI installs browsers and sets `CI=true`. Locally run `bunx playwright install` if you need E2E, or rely on unit + HTML checks until you have disk and bandwidth for browser binaries.
