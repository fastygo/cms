# 04. GraphQL Plugin

GraphQL should be implemented as a plugin or extension over the same core services used by REST and admin.

## Prerequisite

Complete `03-5-runtime-profiles-playground.md` before this pass. GraphQL must not decide binary profile, storage profile, playground import behavior, or admin fixture ownership.

## Required Path

Recommended endpoint:

```text
/go-graphql
```

## Core Rule

GraphQL resolvers must not query storage directly. They must use application services or stable read models.

## Schema Coverage

The first GraphQL plugin should cover:

- Posts.
- Pages.
- Custom content types.
- Taxonomies.
- Terms.
- Media.
- Authors.
- Menus.
- Settings.
- Search.
- Revisions where authorized.
- Preview where authorized.

## Queries

Required query groups:

- `posts`
- `post`
- `pages`
- `page`
- `contentTypes`
- `taxonomies`
- `terms`
- `media`
- `authors`
- `menus`
- `settings`
- `search`

## Mutations

Mutations may be added after read coverage.

Candidate mutations:

- Create content.
- Update content.
- Publish content.
- Schedule content.
- Trash/restore content.
- Upload or attach media through media service.
- Update menu.
- Update setting.

All mutations must enforce capabilities.

## Plugin Settings

Admin settings should include:

- Enable/disable endpoint.
- Public introspection policy.
- Auth policy.
- Query depth/complexity limits.
- CORS policy if endpoint is consumed by browser frontend.
- Cache policy.

## Extension Points

Other plugins should be able to extend:

- Object fields.
- Queries.
- Mutations.
- Enums.
- Resolvers.
- Capability rules for fields.

Extensions must not bypass core visibility rules.

## Consistency With REST

GraphQL and REST must agree on:

- IDs.
- Statuses.
- Slugs.
- Visibility.
- Capability behavior.
- Author data visibility.
- Media metadata visibility.
- Taxonomy assignment semantics.

## Exit Checklist

- GraphQL endpoint can be enabled as plugin.
- Public queries return only published public content.
- Authenticated queries can access allowed private data.
- Resolvers use services.
- Draft leakage tests pass.
- Schema covers enough data for the validation frontend.
