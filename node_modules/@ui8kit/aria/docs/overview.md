# Overview

`@ui8kit/aria` is a **TypeScript behaviour layer** for accessible UI patterns. It does not ship Go `templ` or CSS tokens; it assumes markup that follows small, explicit **data-attribute contracts** (see each pattern under [`docs/patterns/`](patterns/README.md) and `src/patterns/*/markup.md` in the repo).

## Architecture

- **Patterns** — plain objects `{ name, init(root?) }` that bind event listeners and ARIA state to the DOM.
- **Registry** — maps pattern names to implementations, tracks per-scope disposers, optional `window.ui8kit` namespace for CDN/IIFE.
- **Entries** — `tsup` entry points: pure `@ui8kit/aria`, side-effectful `@ui8kit/aria/all`, and per-pattern bundles for tree-shaking or partial CDN loads.

## Relationship to UI8Kit

UI8Kit (Go) supplies **atoms** and layout primitives; this package supplies **client-side interaction** for composite widgets (dialog, tabs, combobox, …) in line with W3C APG-style behaviour, so applications can keep markup stable and version the JS independently of the server template layer.
