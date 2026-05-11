# 02. REST Control Plane

REST is the first external API surface. It should be implemented before GraphQL and before the full frontend validation theme.

## Required Paths

Base paths:

```text
/go-json
/go-json/go/v2/
```

## Required Resources

Implement endpoints for:

- Posts.
- Pages.
- Content types.
- Taxonomies.
- Terms.
- Media.
- Users.
- Authors.
- Settings.
- Menus.
- Search.
- Revisions where authorized.
- Preview where authorized.

## Public Read Behavior

Public callers may read:

- Published public posts.
- Published public pages.
- Public content type metadata.
- Public taxonomies and terms.
- Public media metadata.
- Public author profiles.
- Public menus.
- Public settings.
- Search results for public published content.

Public callers must not read:

- Draft content.
- Scheduled content before publish time.
- Trashed content.
- Private metadata.
- Private user fields.
- Private settings.

## Authenticated Write Behavior

Authenticated endpoints should allow authorized users to:

- Create content.
- Update content.
- Publish content.
- Schedule content.
- Trash/restore content.
- Manage taxonomies.
- Upload media.
- Manage menus.
- Manage settings.
- Manage users and roles where allowed.

All write behavior must call application services.

## API Shape

Implement:

- Discovery document.
- Resource envelope.
- List envelope.
- Pagination metadata.
- Stable error envelope.
- Locale negotiation.
- Filtering.
- Sorting.
- Capability-aware field visibility.

## Important Filters

Content list filters:

- `kind`
- `status`
- `author`
- `taxonomy`
- `search`
- `locale`
- `after`
- `before`

Taxonomy filters:

- `type`
- `parent`
- `content_type`

Media filters:

- `mime_type`
- `owner`
- `attached`

## Exit Checklist

- `/go-json` returns discovery.
- `/go-json/go/v2/` returns namespace discovery.
- Published content lists work.
- Detail by ID works.
- Detail by slug works.
- Drafts are hidden publicly.
- Authenticated admin can create/update/publish.
- Low-privilege user cannot publish.
- Error shape is stable.
- Pagination is stable.
- API can provide enough data for admin MVP.
