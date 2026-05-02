# @ui8kit/aria — documentation

This directory is the **canonical documentation** for the package. The repository root [`README.md`](../README.md) stays a short quick-start; deeper topics live here.

## Contents

| Topic | Description |
| --- | --- |
| [Overview](overview.md) | What the library is, how it fits into UI8Kit, and the mental model. |
| [Getting started](getting-started.md) | Install, entries (`/`, `/all`, per-pattern), init, and first integration. |
| [Why @ui8kit/aria?](why.md) | Motivation: versioning, CDN, tree-shaking, separation from Go `templ`. |
| [Capabilities](capabilities.md) | Bundles, registry, patterns, SSR, and extension points. |
| [Examples](examples.md) | Static HTML examples under [`examples/`](../examples/) and how to run them. |
| [Security](security.md) | Supply chain, CSP, globals, and what SRI does *not* solve. |
| [SRI & CDN](sri-cdn.md) | Subresource Integrity, CDN usage, and the `generate-sri` script. |
| [Testing](testing.md) | Unit tests, E2E, HTML validation, and where they live. |
| [CI & preflight](ci-cd.md) | GitHub Actions vs local `preflight.sh`. |
| [Static file server](static-server.md) | `serve-static.mjs` for local examples and Playwright. |
| [FAQ](faq.md) | Common questions and trade-offs. |

## Pattern reference

Behaviour contracts and markup hints are documented per pattern:

**[Patterns index →](patterns/README.md)**

Source-of-truth for DOM shape during development remains [`src/patterns/<name>/markup.md`](../src/patterns/) in the tree; `docs/patterns/*.md` is the curated, reader-facing summary.
