# 01. Admin Shell And Navigation

This document defines the GoCMS admin shell and navigation profile for UI8Kit-based implementations.

## Shell

The admin shell should own:

- Document layout.
- Sidebar or primary navigation.
- Mobile navigation surface.
- Header actions.
- Account menu.
- Locale switcher where localization is enabled.
- Theme appearance control where supported.
- Global notifications.
- Main content region.

The shell should receive data from application view models. It must not query storage or perform authorization checks by itself.

## Required Paths

The shell must support the admin paths defined by `../../go-codex/en/01-admin-contract.md`:

```text
/go-admin
/go-login
/go-logout
```

The UI may render additional nested admin routes such as:

```text
/go-admin/posts
/go-admin/pages
/go-admin/media
/go-admin/settings
/go-admin/plugins
```

## Navigation Items

Navigation items should include:

- Stable ID.
- Label.
- Path.
- Icon identifier.
- Order.
- Required capability.
- Active matching rules.
- Group or parent.

Navigation must be capability-aware. Hidden navigation does not replace server-side authorization.

## Navigation Groups

Recommended groups:

- Dashboard.
- Content.
- Design.
- Extensions.
- Users.
- System.

Plugins may add groups or items, but they must not override core identifiers without explicit support.

## Account Menu

The account menu should include:

- Current user display name or email.
- Profile link where supported.
- Logout action.
- Locale switcher where supported.
- Theme mode control where supported.

Logout must use a safe state-changing flow.

## Active State

The shell should calculate active navigation state from route metadata, not from brittle string matching alone.

When route metadata is unavailable, path matching should be deterministic and documented.

## Breadcrumbs

Admin screens should provide breadcrumbs for nested resources:

- List.
- Create.
- Edit.
- Detail.
- Settings group.

Breadcrumb items should be generated from stable route metadata.

## Notifications

The shell should render:

- Success messages.
- Validation summaries.
- Authorization failures.
- System warnings.
- Plugin/theme activation failures.

Notifications should be accessible to screen readers.

## Responsive Behavior

Mobile navigation may use a sheet or equivalent disclosure surface.

Responsive behavior must preserve:

- Keyboard navigation.
- Focus restoration.
- Correct ARIA state.
- Server-rendered fallback links.

## Conformance Markers

Recommended markers:

```text
data-gocms-shell
data-gocms-nav
data-gocms-nav-item
data-gocms-account-menu
data-gocms-main
data-gocms-breadcrumbs
```
