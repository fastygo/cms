# 07. Accessibility And Conformance

This document defines accessibility and admin UI conformance expectations for UI8Kit-based GoCMS screens.

## Accessibility Baseline

Admin screens should support:

- Keyboard navigation.
- Visible focus.
- Correct heading structure.
- Form labels.
- Field-level error announcements.
- Dialog focus trapping and restoration.
- Tabs with keyboard behavior.
- Combobox keyboard behavior where used.
- Non-color status indicators.
- Screen-reader accessible notifications.

## Stable Test Markers

Use stable GoCMS markers for conformance tests:

```text
data-gocms-screen
data-gocms-resource
data-gocms-action
data-gocms-state
data-gocms-field
data-gocms-error
data-gocms-block
data-gocms-element
```

Markers should describe GoCMS behavior, not visual implementation.

## Accessible Names

Interactive controls should have accessible names through:

- Visible text.
- Associated labels.
- ARIA label props when visible text is not available.

Icon-only buttons must have accessible labels.

## Focus Management

Focus should be managed for:

- Dialog open and close.
- Sheet open and close.
- Validation failure.
- Route-level form submission result where practical.
- Bulk action confirmation.
- Media picker selection.

Focus should not be lost to the document body after interactive flows.

## Error Presentation

Errors should be presented as:

- Field-level messages.
- Form-level summary for multi-field forms.
- Toast or notification where appropriate.
- Stable machine marker for tests.

Error text should be human-readable and should not expose internal implementation details.

## Admin Conformance Tests

UI conformance tests should verify:

- Login form is reachable.
- Protected admin screens require authentication.
- Required screens expose stable markers.
- Content list supports selection and bulk actions.
- Content edit screen exposes publish controls.
- Validation errors are visible and linked to fields where practical.
- Unauthorized action controls are hidden or disabled, and server rejects direct submission.
- Dialogs and tabs expose expected ARIA state.

## Visual Independence

Conformance tests must not rely on:

- Raw CSS class names.
- Pixel-perfect layout.
- Color values.
- Generated asset filenames.

They should rely on:

- Required URLs.
- Accessible names.
- Roles.
- Stable GoCMS markers.
- Server responses.

## Regression Areas

High-risk areas for regression:

- Bulk actions.
- Publish panel.
- Plugin activation.
- Theme activation.
- Media picker.
- Settings validation.
- Capability-gated menus.
- Locale switching.
- Dialog focus behavior.

These areas should receive focused tests when changed.
