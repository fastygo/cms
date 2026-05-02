# 03. Content Contract

This document defines GoCMS content resources and lifecycle behavior.

## Purpose

Content resources are the shared model used by admin screens, REST endpoints, themes, plugins, search, imports, previews, feeds, and optional GraphQL endpoints.

## Content Kinds

The required content kinds are:

- `post`
- `page`

Implementations MAY support custom content kinds. Custom kinds MUST use stable identifiers and MUST NOT redefine `post` or `page`.

## Statuses

Publishable content MUST support:

- `draft`
- `scheduled`
- `published`
- `archived`
- `trashed`

Status rules:

- `draft` content MUST NOT appear in public unauthenticated output.
- `scheduled` content MUST NOT appear publicly before its publish timestamp.
- `published` content MAY appear publicly if visibility allows it.
- `archived` content MUST NOT appear in normal public listings unless explicitly requested by an authorized user.
- `trashed` content MUST NOT appear publicly and SHOULD be restorable until permanently deleted.

## Required Fields

Content entries MUST have:

- `id`
- `kind`
- `status`
- `slug`
- `title`
- `content`
- `excerpt`
- `author_id`
- `created_at`
- `updated_at`
- `published_at`
- `deleted_at`
- `featured_media_id`
- `template`
- `metadata`

Implementations MAY expose additional fields.

## Localized Fields

Implementations that support localization SHOULD localize:

- `title`
- `slug`
- `content`
- `excerpt`
- `seo_title`
- `seo_description`

Localized fields MUST have deterministic fallback behavior.

## Slugs

Slugs MUST be stable public identifiers.

Rules:

- Slugs MUST be normalized consistently.
- Slugs MUST be unique within their content kind and routing scope.
- Page slugs MAY be unique within parent scope if hierarchical routing is enabled.
- Slug changes SHOULD create redirects when public rendering is enabled.
- Slug lookup MUST respect locale where localization is enabled.

## Pages

Pages are hierarchical evergreen content.

Pages MUST support:

- Parent-child relationship.
- Optional template selection.
- Optional menu inclusion.
- Preview.
- Revisions.
- Featured media.

Pages SHOULD support custom ordering.

## Posts

Posts are chronological content.

Posts MUST support:

- Author.
- Excerpt.
- Featured media.
- Taxonomy assignment.
- Archive listing.
- Feed inclusion when feeds are enabled.
- Preview.
- Revisions.

Posts SHOULD support comments if the implementation includes comments.

## Metadata

Content metadata is a key-value extension surface.

Rules:

- Metadata keys MUST be stable strings.
- Metadata values SHOULD be typed.
- Private metadata MUST NOT be exposed publicly.
- Metadata exposed through REST MUST declare visibility.
- Plugin-owned metadata keys SHOULD use plugin-prefixed names.

Implementations SHOULD support indexing strategy for frequently queried metadata.

## Excerpt

Content MAY have a manually authored excerpt.

If no excerpt exists, implementations MAY generate one. Generated excerpts MUST be clearly distinguishable from stored excerpts in internal APIs or metadata.

## Revisions

Content MUST support revisions for `post` and `page`.

A revision SHOULD capture:

- Title.
- Slug.
- Content.
- Excerpt.
- Status.
- Author.
- Template.
- Featured media.
- Taxonomies.
- Metadata.
- Timestamp.

Implementations MUST allow authorized users to view revision history and restore a revision.

Retention limits MAY be configured. If revisions are pruned, pruning rules MUST be documented.

## Autosave

Admin-compatible implementations SHOULD support autosave or an equivalent draft preservation mechanism.

Autosaves MUST NOT become public content unless explicitly published.

## Preview

Preview MUST allow authorized users to view unpublished changes.

Preview URLs or tokens MUST:

- Be protected by authentication, signed tokens, or equivalent checks.
- Expire or be revocable.
- Not grant broad access to other unpublished content.

## Visibility

Implementations MAY support additional visibility states such as private or password-protected content. If supported, visibility MUST be enforced consistently across admin, REST, themes, search, and feeds.

## Taxonomy Assignment

Posts MUST support taxonomy assignment. Pages MAY support taxonomy assignment.

Taxonomy assignment MUST preserve:

- Taxonomy type.
- Term ID.
- Content ID.
- Ordering where relevant.

## Lifecycle Events

The content lifecycle MUST emit or expose hooks for:

- Before create.
- After create.
- Before update.
- After update.
- Before status transition.
- After status transition.
- Before trash/delete.
- After trash/delete.
- Before restore.
- After restore.

Hook names are defined by the hooks contract.

## Conformance Notes

Conformance tests SHOULD verify:

- Drafts are not public.
- Scheduled content remains hidden before publish time.
- Published content appears in REST.
- Slug lookup is stable.
- Revisions can be restored.
- Preview does not leak unrelated drafts.
- Unauthorized metadata is hidden.
