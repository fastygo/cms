# 07. Decision Log

This document records implementation decisions that should guide early development.

## Decision 1: REST Before GraphQL

REST is implemented before GraphQL.

Reason:

- REST is the compatibility control plane.
- Admin can use REST immediately.
- Conformance can test REST early.
- GraphQL can be built as a plugin over the same services.

## Decision 2: GraphQL Is A Plugin

GraphQL is not the source of truth.

Reason:

- The core model must support admin, REST, GraphQL, CLI, plugins, and frontend validation.
- GraphQL schema should reflect core contracts.
- Resolvers must not query storage directly.

## Decision 3: Custom Content Types Are First-Class

Custom content types are part of the first core kernel pass.

Reason:

- Adding them later would force refactors across services, REST, GraphQL, admin, and frontend validation.
- Full CMS compatibility requires more than posts and pages.

## Decision 4: Authors Are Public Views Over Users

Public author data is separate from private user data.

Reason:

- Frontends need author archives and bylines.
- Private user fields must not leak through public APIs.

## Decision 5: Frontend Theme Is A Validation Fixture

The marketplace-style frontend is a validation app, not the source of backend requirements by itself.

Reason:

- It proves completeness.
- It reveals missing API fields.
- It must consume APIs like an external client.

## Decision 6: Headless Mode Is First-Class

Public rendering can be disabled while admin, REST, and GraphQL remain available.

Reason:

- The target implementation is headless-ready.
- Public rendering should not be required for content management.

## Decision 7: One-Click Migration Is A Product Goal

GoCMS should evolve toward one-click migration from existing content-heavy CMS installations.

Reason:

- The target users need to preserve content, media, authors, taxonomies, menus, slugs, SEO metadata, custom fields, redirects, and public behavior.
- Migration success depends on contract stability from the first implementation pass.
- Import/export cannot be added safely at the end if core entities are too narrow.

## Decision 8: Backward Compatibility Is The Default

Contracts, persisted data, API responses, plugin manifests, theme manifests, hook IDs, and import mappings should evolve through additive changes, deprecation, versioning, and migrations.

Reason:

- Breaking contracts makes one-click migration and third-party extensions unreliable.
- Compatibility debt is more expensive than implementation debt.
- Early conformance tests should protect public behavior before the implementation grows.

## Decision 9: Runtime Profiles Precede GraphQL

GoCMS must define binary/runtime profiles before GraphQL implementation.

Profiles:

- `headless`
- `admin`
- `playground`
- `full`
- `conformance`

Reason:

- GraphQL should not own runtime shape or storage choices.
- Site-specific binaries need a clean way to mount GoCMS admin and API features.
- Playground demos must not contaminate production storage assumptions.

## Decision 10: Playground Is Browser-Local And Server-Stateless

The playground profile stores mutable demo content in browser storage, not in the binary or backend.

Reason:

- Users can safely play with the admin without server-side persistence.
- A public demo site can expose `/go-admin` without sharing mutable data between users.
- Export/import JSON gives users ownership of their edited snapshot.
- Existing browser-local content must not be silently overwritten by external source refresh.

## Decision 11: Admin Fixtures Are Not Site Content

Admin UI fixtures and labels should be embedded JSON fixtures, while posts, pages, taxonomies, media, menus, and public settings remain site content.

Reason:

- Admin data is readonly runtime/UI configuration.
- Site content is mutable and profile-dependent.
- The playground profile can start with empty site content while still rendering the admin shell.

## Decision 12: Playground Snapshots Follow Compatibility REST Shapes

Playground JSON should stay close to source compatibility REST route payloads where possible.

Reason:

- Exported snapshots can become conformance fixtures.
- Future migration tooling can reuse the same shapes.
- The system avoids drifting into a private demo-only DTO format.

## Decision 13: XML Import Is A Plugin Boundary

WXR/XML import is planned as an import plugin, not as part of core playground bootstrap.

Reason:

- REST bootstrap demonstrates the admin quickly.
- XML import is a migration pipeline with broader entity and metadata coverage.
- Both import paths should produce the same compatibility entities.
