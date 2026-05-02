# 06. Styling And ui8px Policy

This document defines styling policy for UI8Kit-based GoCMS admin screens.

## Utility Class Rules

Utility classes should be:

- Explicit.
- Static.
- Policy-approved.
- Reviewable.
- Free of hidden runtime-only assembly.

Avoid broad safelists and dynamic class strings that cannot be checked by policy tools.

## Spacing

Admin layouts should use the strict layout spacing scale.

Compact controls and primitive internals may use finer spacing only when the relevant files are in a reviewed control scope.

## Colors

Reusable admin UI should use semantic token classes.

Recommended token categories:

- Background.
- Foreground.
- Card.
- Popover.
- Primary.
- Secondary.
- Muted.
- Accent.
- Destructive.
- Border.
- Input.
- Ring.

Do not use raw palette utilities in reusable elements and blocks.

## Variants And Patterns

Prefer:

- Existing UI8Kit variants.
- Existing semantic classes.
- Reviewed local elements.
- Reviewed local blocks.

Avoid repeated one-off utility compositions. Repeated compositions should become:

- A variant.
- A local reusable element.
- A local reusable block.
- A reviewed semantic pattern.

## Policy Files

If the project owns explicit policy, commit reviewed policy files.

Policy may define:

- Allowed utilities.
- Denied utilities.
- File path scopes.
- Class groups.
- Semantic patterns.

Telemetry and learned proposals are review inputs, not automatic approvals.

## Generated CSS

Generated CSS should be rebuilt through project scripts.

Do not hand-edit generated CSS output.

## Brand Overrides

Brand CSS may add:

- Brand tone.
- Marketing decoration.
- Token overrides.
- Product-specific surface adjustments.

Brand CSS must not be required for basic admin usability.

## Review Gates

UI work should not be considered complete until:

- Utility policy passes.
- Color token policy passes.
- Repeated patterns are reviewed.
- Accessibility checks pass.
- Build output is regenerated.

## Conformance

Conformance tests should not depend on styling classes. They should use stable `data-gocms-*` markers and accessible roles/names.
