# Why @ui8kit/aria?

## Versioning and delivery

ARIA behaviour in a **separate npm package** (with optional CDN artifacts) gives you:

- Semver for breaking interaction changes without retagging the Go UI layer.
- **Pinned CDN URLs** + Subresource Integrity for static sites and sandboxes.
- A single place to fix keyboard bugs and ship patches to all consumers.

## Tree-shaking vs full bundle

Applications that bundle with esbuild/Rollup/Vite benefit from the **pure** entry (`@ui8kit/aria`): only imported patterns end up in the chunk.

Drop-in demos, CodePen-style pages, and legacy stacks often want one script tag — the **IIFE** `all` bundle registers every pattern and exposes `window.ui8kit`.

## Separation of concerns

Go `templ` + UI8Kit describe **structure and styling contracts**; this library owns **focus management, keyboard routing, and ARIA state sync**. That split keeps server templates declarative and keeps behavioural complexity in one typed codebase with tests.
