# GoCMS UI8Kit Admin Profile

This directory defines the GoCMS admin user interface profile for implementations that choose UI8Kit for server-rendered admin surfaces.

This profile is not required for GoCMS compatibility in general. The external compatibility contract is `../../go-codex/en`. The Go architecture profile is `../../go-stack/en`.

## Scope

This profile covers:

- Admin shell and navigation.
- UI8Kit-first composition.
- Local reusable elements and blocks.
- HTML5 semantic boundaries.
- ARIA ownership and runtime behavior.
- Admin forms, tables, dialogs, tabs, filters, and bulk actions.
- Runtime assets.
- Styling, token, and ui8px policy.
- Accessibility and conformance markers.

## Non-Goals

This profile does not define:

- The external GoCMS compatibility contract.
- Domain model behavior.
- Storage architecture.
- Plugin runtime architecture.
- A requirement that all GoCMS-compatible implementations use UI8Kit.

## Document Map

- `00-ui8kit-profile.md` defines the profile and its relationship to the other layers.
- `01-admin-shell-and-navigation.md` defines shell, navigation, account, theme, and locale behavior.
- `02-ui-composition-boundaries.md` defines views, local elements, local blocks, and archive boundaries.
- `03-html5-and-aria-policy.md` defines semantic tag rules and ARIA ownership.
- `04-forms-tables-and-admin-patterns.md` defines admin UI patterns for common CMS workflows.
- `05-runtime-assets.md` defines generated assets, subset runtime JS, and progressive enhancement.
- `06-styling-and-ui8px-policy.md` defines utility class, token, and styling rules.
- `07-accessibility-and-conformance.md` defines accessibility and admin UI conformance expectations.

## Layer Relationship

```text
go-codex/en -> go-stack/en -> go-ui8kit/en -> implementation
```

If this profile conflicts with `go-codex/en`, the external compatibility contract wins. If this profile conflicts with `go-stack/en`, the stack profile wins for non-UI architecture.
