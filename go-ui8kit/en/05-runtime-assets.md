# 05. Runtime Assets

This document defines runtime asset rules for UI8Kit-based GoCMS admin screens.

## Generated Runtime

Runtime UI assets should be generated through project scripts.

Rules:

- Do not hand-edit generated UI runtime JavaScript.
- Do not use full CDN bundles for app-level admin behavior.
- Include only the runtime patterns actually used by rendered UI.
- Keep generated files out of source-of-truth discussions.

## Pattern Subsetting

The admin should declare required interactive patterns.

Examples:

- `dialog`
- `tabs`
- `accordion`
- `combobox`
- `tooltip`
- `disclosure`

The generated runtime should include only declared patterns unless a documented full-build profile is used for development.

## Progressive Enhancement

Admin screens should remain usable through:

- Server-rendered links.
- Server-rendered forms.
- Redirects.
- Accessible fallback states.

JavaScript may enhance:

- Focus management.
- Keyboard shortcuts.
- Upload progress.
- Dialog behavior.
- Toasts.
- Client-side filtering hints.

JavaScript must not be the only layer enforcing authorization, validation, CSRF, publish status, or plugin activation state.

## Asset Locations

Implementations should keep app-owned assets in an app-owned static directory.

Recommended categories:

- Generated UI runtime JS.
- Theme mode script.
- Admin CSS.
- Brand overrides.
- Font assets.
- Icon assets.

The public theme asset pipeline may be separate from admin assets.

## Cache Busting

Production assets should support cache busting through:

- Content hashes.
- Manifest files.
- Versioned URLs.
- Deployment-specific asset resolvers.

Admin pages should reference the resolved asset URL rather than hard-coded build output names.

## CSS Layering

CSS should be layered:

1. UI8Kit base styles.
2. Token and component styles.
3. App admin CSS.
4. Brand-specific overrides.

If brand CSS is disabled, the admin should remain usable and accessible with base styles.

## Avoid Extra Scripts

App-specific JavaScript should be:

- Small.
- Explicit.
- Progressively enhanced.
- Covered by tests when it changes accessibility state.

Do not add broad runtime frameworks just to implement one admin interaction.

## Verification

Asset verification should check:

- Generated runtime matches declared patterns.
- Required files exist.
- No generated runtime was hand-edited.
- Admin pages reference built assets.
- Accessibility patterns have matching markup.
