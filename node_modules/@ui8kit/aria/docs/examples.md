# Examples

Static demos live under [`examples/`](../examples/) at the repository root (not under `docs/`).

| Path | Intent |
| --- | --- |
| `examples/plain/` | Accordion, tabs, dialog with the full IIFE bundle — good baseline for manual testing and axe runs. |
| `examples/spa/` | `__UI8KIT_ARIA_AUTO_INIT__` off; manual `init` on a subtree — mimics a client island. |
| `examples/htmx/` | Partial-root `init` — useful for swap-in markup. |

## Running locally

Use the minimal static server (see [Static file server](static-server.md)) from the repo root:

```bash
node scripts/serve-static.mjs --port 4173 --root .
```

Then open `http://127.0.0.1:4173/examples/plain/` (and siblings). Ensure `bun run build` has produced `dist/` so script `src` paths in the HTML resolve.

Playwright in CI uses the same server command with a health URL under `/examples/…/` (see `playwright.config.ts`).
