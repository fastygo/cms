# Testing

## Layout

| Layer | Location | Tooling |
| --- | --- | --- |
| Unit | `src/**/__tests__/*.ts` | Vitest + jsdom |
| E2E | `tests/e2e/*.spec.ts` | Playwright |
| HTML | `examples/**/*.html` | `html-validate` with [`.htmlvalidate.json`](../.htmlvalidate.json) |

## Unit tests

Patterns are tested in isolation with a small DOM fixture per file. Helpers in `tests/setup.ts` reset the registry and DOM between tests so global listeners and `WeakMap` scopes do not leak.

Run:

```bash
bun run test:unit
```

## E2E and a11y

Playwright serves the repo root via [`scripts/serve-static.mjs`](static-server.md) and hits `/examples/…/`. Accessibility checks use `@axe-core/playwright` on selected specs (e.g. `@a11y` grep).

Local runs may skip browser installs or full matrices; CI installs browsers and runs the full suite. See [CI & preflight](ci-cd.md).

## HTML validation

`bun run test:html` validates example HTML. Rules are relaxed where appropriate for self-closing void elements and ARIA-heavy demos (see config file).
