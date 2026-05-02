# 01. Admin Contract

This document defines the required browser-facing admin behavior for GoCMS-compatible implementations.

## Purpose

The admin contract ensures that editors, administrators, plugin authors, and conformance tools can rely on a stable management surface regardless of implementation details.

## Required Paths

An implementation that supports the Admin compatibility level MUST expose:

```text
/go-admin
/go-login
/go-logout
```

Behavior:

- `GET /go-admin` MUST require an authenticated user with an admin-access capability.
- Unauthenticated access to `/go-admin` MUST redirect to `/go-login` or return an auth challenge in documented non-browser profiles.
- `GET /go-login` MUST render a login surface unless external authentication is configured.
- `/go-logout` MUST end the current browser session through a safe state-changing flow.

Implementations MAY expose additional aliases, but the required paths MUST remain stable.

## Required Screens

The admin MUST provide screens for:

- Dashboard.
- Posts.
- Pages.
- Media.
- Taxonomies.
- Menus.
- Themes.
- Plugins.
- Users.
- Roles or capabilities.
- Settings.
- System health or diagnostics.

Optional screens MAY include comments, forms, redirects, webhooks, jobs, audit logs, and GraphQL settings.

## Dashboard

The dashboard SHOULD show:

- Content counts by status.
- Recent drafts.
- Recently published entries.
- Scheduled entries.
- Recent media uploads.
- Plugin or system warnings.
- Failed jobs or background task warnings.
- Storage or database health.

Dashboard widgets MUST respect capabilities.

## List Screens

List screens for content, users, media, plugins, and themes MUST support:

- Pagination.
- Stable sorting where sortable fields exist.
- Search or filtering where meaningful.
- Empty state.
- Per-row actions.
- Bulk selection.
- Bulk actions where destructive or state transition actions exist.

List screens MUST NOT reveal items the current user is not authorized to view.

## Edit Screens

Content edit screens MUST expose:

- Title.
- Slug/permalink editor.
- Main content editor.
- Excerpt when supported.
- Status controls.
- Publish/schedule controls.
- Visibility controls when supported.
- Author control when authorized.
- Featured media control when supported.
- Taxonomy assignment when supported.
- Revision access.
- Preview action.
- Save draft action.
- Delete or move-to-trash action when authorized.

The editor MAY be rich text, markdown, structured fields, or another UI. It MUST preserve content according to the content contract.

## Required Content States

The admin MUST visibly distinguish:

- `draft`
- `scheduled`
- `published`
- `archived`
- `trashed`

State transitions MUST call the same application rules used by REST, CLI, plugins, and other transports.

## Actions

The admin MUST support these actions where the resource type supports them:

- Create.
- Read.
- Update.
- Delete or trash.
- Restore.
- Publish.
- Unpublish or move to draft.
- Schedule.
- Preview.
- Duplicate or clone SHOULD be supported for content.

Destructive actions MUST require confirmation or an equivalent undo-safe flow.

## Authentication And Session Behavior

Admin sessions MUST:

- Be bound to authenticated identity.
- Expire after a configurable period or inactivity policy.
- Be invalidated on logout.
- Be protected against session fixation.
- Use secure cookie attributes or an equivalent secure session mechanism in browser deployments.

Login attempts MUST be rate-limited or protected by an equivalent abuse defense.

## CSRF And Action Tokens

Browser state-changing admin actions MUST be protected against CSRF.

The protection MAY be implemented through:

- Per-session CSRF tokens.
- Per-action nonces.
- SameSite cookies plus explicit action verification.
- Equivalent signed action tokens.

Tokens MUST be scoped enough to prevent cross-action reuse for high-risk operations such as delete, plugin activation, role changes, and settings changes.

## Capability Enforcement

Every admin screen and action MUST check capabilities. Hiding a button in the UI is not sufficient.

Examples:

- A user without `content.publish` MUST NOT publish content through hidden endpoints.
- A user without `plugins.activate` MUST NOT activate a plugin.
- A user without `settings.manage` MUST NOT update settings.

## Admin Menu

The admin menu MUST support:

- Core menu items.
- Plugin-provided menu items.
- Capability-aware visibility.
- Stable item identifiers.
- Ordering.
- Nested sections or groups.

Plugin-provided menu items MUST NOT override core menu identifiers unless explicitly allowed by the plugin contract.

## Notifications And Errors

Admin actions MUST return clear success and failure states.

Errors SHOULD include:

- Human-readable message.
- Stable error code when available.
- Field-level validation details for forms.
- Request ID or support token when available.

Internal implementation details MUST NOT be exposed to normal admin users.

## Media Admin

The media screen MUST support:

- Upload.
- Search.
- Filtering by media type where supported.
- Preview.
- Metadata edit.
- Deletion with authorization.
- Copying URL or selecting media for content.

Media upload errors MUST distinguish validation errors from server errors.

## Plugins And Themes Admin

Plugin and theme screens MUST show:

- Name.
- Version.
- Author or provider.
- Compatibility requirements.
- Active/inactive state.
- Required capabilities.
- Install, activate, deactivate, update, and uninstall actions where supported.

Activation failures MUST leave the previous stable state intact.

## Headless Profile

In a headless profile, public rendering MAY be disabled. Admin behavior remains required for Admin compatibility unless the implementation declares a lower compatibility level.

The admin SHOULD provide explicit settings for enabled delivery surfaces such as REST, GraphQL, feeds, and public rendering.

## Conformance Notes

Conformance tests SHOULD verify:

- Required paths exist.
- Unauthenticated admin access is blocked.
- Capability checks are enforced server-side.
- Content state transitions work.
- CSRF protection rejects missing or invalid action tokens.
- Plugin and theme activation failures do not corrupt active state.
