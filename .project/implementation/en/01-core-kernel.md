# 01. Core Kernel

The core kernel is the first real implementation pass. It creates the domain model and application services used by REST, admin, GraphQL, plugins, and frontend validation.

## Required Domains

Implement these domains first:

- `content`
- `contenttypes`
- `taxonomy`
- `media`
- `users`
- `authz`
- `settings`
- `menus`
- `revisions`
- `preview`

## Content Entries

Content entries must support:

- ID.
- Kind: `post`, `page`, and custom kinds.
- Status.
- Localized title.
- Localized slug.
- Localized content.
- Localized excerpt.
- Author ID.
- Featured media ID.
- Template key.
- Metadata.
- Created timestamp.
- Updated timestamp.
- Published timestamp.
- Deleted timestamp.

## Content Types

Content types are first-class.

They must define:

- ID.
- Label.
- Public visibility.
- REST visibility.
- GraphQL visibility when GraphQL is installed.
- Supports: title, editor, excerpt, featured media, revisions, taxonomies, custom fields, comments if enabled.
- Archive support.
- Slug/permalink rules.
- Capability mapping.

Do not hard-code only `post` and `page` in services.

## Taxonomies

Taxonomies must support:

- `category`.
- `tag`.
- Custom taxonomy types.
- Hierarchical or flat mode.
- Localized name.
- Localized slug.
- Description.
- Parent ID for hierarchical terms.
- Assignment to content types.

## Users And Authors

Users and authors must support:

- User ID.
- Display name.
- Slug.
- Email visibility rules.
- Avatar/media reference.
- Bio.
- Roles/capabilities.
- Active/disabled state.

Public author data must be separated from private user data.

## Capabilities

Capabilities must be granular.

Minimum groups:

- Admin access.
- Content create/read/edit/publish/delete/restore.
- Own vs others content.
- Media upload/edit/delete.
- Taxonomy manage/assign.
- Menus manage.
- Settings manage.
- Users manage.
- Roles manage.
- Plugins manage.
- Themes manage.
- REST private access.

## Services

Minimum services:

- Content service.
- Content type service.
- Taxonomy service.
- Media service.
- User service.
- Authorization service.
- Settings service.
- Menu service.
- Revision service.
- Preview service.

Services must enforce authorization and visibility rules.

## Exit Checklist

- Can create draft post.
- Can create draft page.
- Can register custom content type.
- Can register custom taxonomy.
- Can publish content.
- Can schedule content.
- Can assign author.
- Can assign taxonomy terms.
- Can attach featured media.
- Can create and restore revision.
- Can preview draft with authorization.
- Drafts and scheduled content are hidden from public read queries.
