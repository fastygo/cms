# 03. Headless Admin MVP

The admin MVP proves that non-technical users can manage every core CMS entity while public rendering remains optional or disabled.

## Required Paths

```text
/go-admin
/go-login
/go-logout
```

## Required Screens

Implement:

- Login.
- Dashboard.
- Posts.
- Pages.
- Content types.
- Taxonomies.
- Media.
- Menus.
- Users.
- Authors.
- Roles/capabilities.
- Settings.
- API/headless settings.

Optional early screens:

- Revisions.
- Preview.
- System health.
- Plugin registry.
- Theme registry.

## Admin Workflow Priority

Build workflows in this order:

1. Login/logout.
2. Dashboard with counts.
3. Posts list/create/edit/publish.
4. Pages list/create/edit/publish.
5. Taxonomy management.
6. Media upload/select/attach.
7. Menus.
8. Users/authors.
9. Capabilities.
10. Settings.
11. API/headless switches.

## Content Edit Screen

Must include:

- Title.
- Slug.
- Content editor.
- Excerpt.
- Status.
- Publish/schedule controls.
- Author.
- Featured media.
- Taxonomy assignment.
- Metadata/custom fields.
- Revision access.
- Preview.
- Save draft.
- Publish.
- Trash/restore.

## Headless Settings

Admin should expose:

- Public rendering enabled/disabled.
- REST public access policy.
- GraphQL plugin status when installed.
- Search indexing status.
- Sitemap/feed status.
- CORS policy if needed for frontend delivery.

## Capability Rules

Admin UI must:

- Hide inaccessible actions.
- Reject direct submissions without capability.
- Distinguish own vs others content.
- Require stronger permissions for settings, plugins, themes, roles, and users.

## UI Implementation

If using UI8Kit, follow `../../../go-ui8kit/en`.

If using another UI, it must still satisfy the admin behavior in `../../../go-codex/en/01-admin-contract.md`.

## Exit Checklist

- Admin can manage posts.
- Admin can manage pages.
- Admin can manage custom content types.
- Admin can manage taxonomies.
- Admin can upload media and select featured media.
- Admin can manage menus.
- Admin can manage authors/users.
- Admin can manage capabilities.
- Admin can configure headless mode.
- Low-privilege users cannot access forbidden actions.
