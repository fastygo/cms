# 01. Domain Model

This document defines recommended Go domain package boundaries for GoCMS implementations.

## Package Boundaries

Recommended domain packages:

```text
content
taxonomy
media
users
authz
settings
menus
themes
plugins
hooks
revisions
search
```

Domain packages should define entities, value objects, typed constants, validation invariants, and domain-level errors. They should not own storage, HTTP handlers, rendering, or session logic.

## Content

The `content` package should represent posts, pages, custom content kinds, statuses, slugs, excerpts, metadata, and publish windows.

Recommended types:

- `ID`
- `Kind`
- `Status`
- `Entry`
- `LocalizedText`
- `Visibility`
- `Metadata`
- `Query`
- `Revision`

Stable kinds:

- `post`
- `page`

Stable statuses:

- `draft`
- `scheduled`
- `published`
- `archived`
- `trashed`

## Taxonomy

The `taxonomy` package should represent taxonomy types, terms, hierarchy, localized names, slugs, and assignments.

Recommended types:

- `Type`
- `Term`
- `Assignment`
- `TermQuery`

Default types:

- `category`
- `tag`

## Media

The `media` package should represent media assets, variants, upload metadata, ownership, captions, alt text, and storage resolver keys.

Recommended types:

- `Asset`
- `Variant`
- `Upload`
- `Attachment`
- `Metadata`

Media URLs should be resolved through a service, not constructed directly in domain objects.

## Users

The `users` package should represent identities, profile data, account state, login metadata, and user-owned settings.

Recommended types:

- `User`
- `UserID`
- `AccountStatus`
- `Profile`

Password hashes, sessions, and external identity bindings should be separated from public profile data.

## Authorization

The `authz` package should represent capabilities, roles, grants, ownership checks, and policy decisions.

Recommended types:

- `Capability`
- `Role`
- `Grant`
- `Decision`
- `Principal`
- `Policy`

Roles are named capability sets. Capabilities are the real permission boundary.

## Settings

The `settings` package should represent typed settings, visibility, validation, groups, defaults, and update events.

Recommended types:

- `Key`
- `Value`
- `Type`
- `Visibility`
- `Group`
- `Definition`

Public and private settings must be distinguishable at the domain level.

## Themes

The `themes` package should represent theme manifests, template roles, slots, assets, settings, and activation state.

Recommended types:

- `ThemeID`
- `Manifest`
- `TemplateRole`
- `Slot`
- `Asset`
- `Activation`

Theme entities should not contain business logic for content visibility.

## Plugins

The `plugins` package should represent plugin manifests, lifecycle state, dependencies, migrations, capabilities, routes, hooks, jobs, and assets.

Recommended types:

- `PluginID`
- `Manifest`
- `State`
- `Dependency`
- `LifecycleEvent`
- `Migration`

Plugin state transitions should be explicit and auditable.

## Hooks

The `hooks` package should define hook IDs, handler descriptors, priority, category, arguments, and failure policy.

Recommended types:

- `ID`
- `Category`
- `Priority`
- `Descriptor`
- `FailurePolicy`
- `Result`

Hook handlers may live outside the domain package, but hook descriptors should be stable.

## Domain Invariants

Domain packages should enforce invariants such as:

- Invalid statuses are rejected.
- Content kind identifiers are normalized.
- Slugs are non-empty when content is publishable.
- Public settings cannot contain private values.
- Plugin IDs are URL-safe.
- Capability identifiers are stable.

Invariants that require storage lookups belong in services, not pure entities.
