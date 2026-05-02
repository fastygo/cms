# 00. UI8Kit Profile

This document defines the UI8Kit profile for GoCMS admin surfaces.

## Purpose

The profile provides one concrete UI implementation strategy for the admin behavior required by `../../go-codex/en/01-admin-contract.md`.

It standardizes how GoCMS admin screens should be composed when using UI8Kit:

- UI8Kit primitives first.
- Neutral composites before local equivalents.
- Local reusable elements and blocks for app-level patterns.
- No ad-hoc raw markup in admin templates.
- Static, policy-approved utility classes.
- Generated runtime assets.
- Accessible server-rendered markup with runtime ARIA synchronization.

## Compatibility Boundary

UI8Kit is not required for GoCMS compatibility in general.

An implementation can pass `go-codex` conformance with another UI approach. This profile only applies when a GoCMS implementation declares the `go-ui8kit` admin profile.

## Composition Order

Admin UI should be composed in this order:

1. UI8Kit primitives.
2. UI8Kit neutral composites.
3. Local reusable elements.
4. Local reusable blocks.
5. Page views that assemble blocks and elements.

Page views should not become one-off component libraries.

## Required Admin Surfaces

The UI profile should cover:

- Dashboard.
- Posts.
- Pages.
- Media.
- Taxonomies.
- Menus.
- Themes.
- Plugins.
- Users.
- Roles and capabilities.
- Settings.
- System diagnostics.

These surfaces map to `go-codex` admin requirements.

## Stable Markers

Admin screens should expose stable testing markers:

```text
data-gocms-screen
data-gocms-action
data-gocms-resource
data-gocms-state
```

Conformance tests should use stable GoCMS markers rather than fragile CSS classes.

## Generated Files

Generated files are outputs, not source truth.

Rules:

- Regenerate generated UI files through project scripts.
- Do not manually edit generated runtime JS.
- Do not copy generated files from archive mirrors as source truth.

## Progressive Enhancement

The admin should remain usable through server-rendered links and forms for critical actions.

JavaScript may enhance:

- Dialogs.
- Sheets.
- Tabs.
- Accordions.
- Comboboxes.
- Upload progress.
- Bulk action selection.
- Toasts.

JavaScript must not be the only enforcement layer for authorization, validation, or state transitions.
