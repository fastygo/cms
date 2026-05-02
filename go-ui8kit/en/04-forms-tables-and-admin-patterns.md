# 04. Forms, Tables, And Admin Patterns

This document defines common admin UI patterns for UI8Kit-based GoCMS screens.

## Forms

Admin forms should use shared form primitives and reusable field elements.

Required form behavior:

- Label every control.
- Show field-level validation errors.
- Show form-level validation summary where useful.
- Preserve submitted values after validation failure.
- Distinguish required, optional, disabled, and read-only fields.
- Include CSRF or action token fields for state-changing browser actions.

Forms should work as server-submitted forms for critical actions. JavaScript may enhance the experience.

## Content Editor Layout

Content edit screens should include:

- Main editor area.
- Slug/permalink control.
- Status and publish panel.
- Featured media panel.
- Taxonomy panel.
- Excerpt panel where supported.
- Revision panel.
- Preview action.
- Save draft action.
- Publish or schedule action.
- Trash/delete action where authorized.

Panels should be reusable blocks when repeated across content kinds.

## Tables

Admin tables should support:

- Column headers.
- Sort state.
- Row actions.
- Bulk selection.
- Bulk actions.
- Search/filter controls.
- Empty state.
- Pagination.
- Loading or disabled state where asynchronous enhancement exists.

Tables should expose stable markers for conformance tests.

## Filters

Filter controls should:

- Reflect current state.
- Be bookmarkable where possible.
- Submit through normal links or forms.
- Preserve unrelated query parameters where appropriate.
- Show active filters clearly.

## Bulk Actions

Bulk actions should:

- Require selected rows.
- Require authorization.
- Require confirmation for destructive actions.
- Include action token protection.
- Return a clear success/failure summary.

Partial failure should be reported item by item when possible.

## Dialogs And Confirmation

Use shared dialog or alert-dialog components for:

- Delete confirmation.
- Plugin activation failure details.
- Theme preview warning.
- Media picker.
- Unsaved changes warning where supported.

Destructive dialogs should identify the target resource and action.

## Media Picker

The media picker should support:

- Upload.
- Search.
- Type filtering.
- Selection.
- Preview.
- Alt text display/edit where supported.
- Clear selected media action.

Upload progress may be enhanced with JavaScript, but validation and final state must be enforced server-side.

## Settings Screens

Settings screens should group fields by domain:

- Site identity.
- Reading/public rendering.
- Localization.
- Media.
- SEO.
- REST/API.
- Plugins.
- Themes.
- Security.

Settings groups should expose capability requirements.

## Plugin And Theme Cards

Plugin and theme cards should show:

- Name.
- Version.
- Description.
- Author/provider.
- Compatibility status.
- Active/inactive state.
- Actions.
- Last error where applicable.

Activation and deactivation actions must use action tokens and server-side capability checks.

## Status Badges

Use consistent badges for:

- `draft`
- `scheduled`
- `published`
- `archived`
- `trashed`
- plugin active/inactive/failed
- theme active/available/invalid

Badges should not rely on color alone.
