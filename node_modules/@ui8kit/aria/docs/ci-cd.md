# CI and preflight

## GitHub Actions

The workflow runs **`bun run preflight`** (after installing dependencies and Playwright browsers). That single entry point mirrors what maintainers are expected to run before merge or release.

## `scripts/preflight.sh`

Orchestrates, in order:

1. `bun install` (with `--frozen-lockfile` when `bun.lock` exists, `--ignore-scripts` to avoid duplicate `prepare` builds).
2. `lint` — ESLint flat config.
3. `typecheck` — `tsc --noEmit`.
4. `test:unit` — Vitest.
5. `test:html` — `html-validate` on `examples/**/*.html`.
6. `build` — `tsup` (ESM, CJS, IIFE, DTS).
7. **Playwright** — only when `CI=true` (GitHub sets this). Locally the script prints a skip message so machines without browser caches or disk space can still run the rest.
8. `size` — `size-limit` on IIFE artefacts.
9. `sri` — `node scripts/generate-sri.mjs` → `dist/sri.json`.

To force the full suite locally:

```bash
CI=true bun run preflight
```

## Release

Publishing is intentionally manual in this repo; preflight is the quality gate before you version and `npm publish`.
