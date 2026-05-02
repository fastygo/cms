# Capabilities

## Bundles (`tsup`)

- **ESM / CJS** — per-entry `dist/*.mjs` and `dist/*.cjs` with TypeScript declarations.
- **IIFE** — `dist/all.iife.js`, `dist/all.iife.min.js`, and `dist/core.iife(.min).js` for globals (`ui8kit`).

## Registry API

- `registerPattern(pattern)` — idempotent registration by `pattern.name`.
- `init(root?)` / `initPattern(name, root?)` — bind listeners for a document or subtree; duplicate inits for the same scope are skipped.
- `disposePattern` / `resetRegistry` — teardown and test isolation.

## SSR

The pure entry does not touch `window` on import. `getNamespace()` lazily attaches `window.ui8kit` only in a browser context when first used.

## Patterns

See [pattern index](patterns/README.md). Each pattern is intentionally small: no framework adapter, no virtual DOM — only DOM events and attributes.

## Limits (by design)

- No React/Vue/Svelte components; consumers wrap markup themselves.
- Markup contracts are not validated at runtime beyond guards inside each pattern.
