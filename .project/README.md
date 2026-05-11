# GoCMS Project Workspace

This directory contains internal planning material for implementing GoCMS.

It is intentionally separate from the public compatibility and profile documents:

- `../go-codex/en` defines external compatibility.
- `../go-stack/en` defines the Go-native architecture profile.
- `../go-ui8kit/en` defines the UI8Kit admin profile.
- `.project/implementation/en` defines the practical implementation roadmap.
- `.project/progress.md` tracks pass-by-pass implementation progress.

The `.project` directory is ignored by git. Use it for working plans, sequencing, checklists, and implementation notes that may change frequently.

## Current Implementation Track

Start with:

- `progress.md`
- `implementation/en/00-roadmap.md`
- `implementation/en/01-core-kernel.md`
- `implementation/en/02-rest-control-plane.md`
- `implementation/en/03-admin-mvp.md`
- `implementation/en/04-graphql-plugin.md`
- `implementation/en/05-theme-frontend-validation.md`
- `implementation/en/06-entity-coverage-checklist.md`

## Rule

Implementation planning must preserve this dependency direction:

```text
Core Services -> REST Control Plane -> Admin GUI -> GraphQL Plugin -> Theme Frontend Validation
```

Do not let the frontend theme or GraphQL schema define the core model. They validate and consume it.
