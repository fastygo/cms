# 00. Implementation Roadmap

This document defines the implementation sequence for GoCMS headless CMS compatibility.

## Core Dependency Rule

Build in this direction:

```text
Core Kernel -> REST Control Plane -> Admin MVP -> Runtime Profiles + Playground -> GraphQL Plugin -> Theme/Public API Foundation -> CMS Admin Finalization -> Conformance -> Production CMS Completion
```

Do not start with GraphQL or the frontend theme. They should consume the same core services as REST and admin. Runtime profile and playground storage boundaries must be fixed before GraphQL so the plugin does not assume one binary shape or one storage mode.

## Pass 0: Project Skeleton

Goal: create a runnable empty application with basic process boundaries.

Deliverables:

- Go module.
- Application entrypoint.
- Configuration loading.
- Health endpoints.
- Structured logging.
- Basic test harness.
- Empty domain/application/storage/delivery directories.
- Seed/test fixture strategy.

Exit criteria:

- The server starts.
- Health checks respond.
- Tests run.
- No CMS business logic exists yet.

## Pass 1: Core Compatibility Kernel

Goal: implement the minimum CMS domain and application services.

Deliverables:

- Users/authors.
- Capabilities.
- Content entries: `post`, `page`, and custom content types.
- Statuses: `draft`, `scheduled`, `published`, `archived`, `trashed`.
- Slugs and localized fields.
- Taxonomies: categories, tags, custom taxonomies.
- Media metadata.
- Settings.
- Menus.
- Revisions and preview contracts.

Exit criteria:

- Services can create, update, publish, schedule, trash, restore, and query content.
- Drafts and scheduled content are not public.
- Custom content types and taxonomies are first-class, not bolted on later.

## Pass 2: REST Control Plane

Goal: expose the compatibility REST surface.

Deliverables:

- `/go-json`.
- `/go-json/go/v2/`.
- Resource endpoints for posts, pages, content types, taxonomies, media, users/authors, settings, menus, and search.
- Pagination, filtering, sorting, locale, auth, and error envelopes.
- Authenticated write endpoints for admin operations.

Exit criteria:

- Public callers can read published content.
- Authenticated admin/editor callers can manage resources through services.
- REST conformance can begin.

## Pass 3: Headless Admin MVP

Goal: implement `/go-admin` as the first complete management GUI.

Deliverables:

- Login/logout.
- Dashboard.
- Posts.
- Pages.
- Content types.
- Taxonomies.
- Media.
- Menus.
- Users/authors.
- Roles/capabilities.
- Settings.
- API/headless settings.

Exit criteria:

- Editors can manage all core entities from the admin.
- Admin uses the same services as REST.
- UI enforces capabilities server-side.
- Public rendering can remain disabled.

## Pass 3.5: Runtime Profiles And Playground Boundary

Goal: define how GoCMS is assembled into binaries and how playground mode can run a complete isolated CMS experience without server-side persistence.

Deliverables:

- Runtime profiles: `headless`, `admin`, `playground`, `full`, and `conformance`.
- Storage profiles: browser IndexedDB, memory, JSON fixtures, SQLite, MySQL, and PostgreSQL.
- Clear split between readonly admin JSON fixtures and mutable site content.
- Browser-local playground storage with IndexedDB.
- Playground as an isolated full CMS sandbox: admin plus public preview/rendering, inspired by WordPress Playground.
- One-time source bootstrap through `?gocms=<source>`.
- Demo login policy for playground (`admin` / `admin`).
- Import/export JSON policy for playground snapshots.
- Media Blob policy for user-uploaded images in IndexedDB.
- Blueprint-style startup configuration and embeddable preview goals.
- Compatibility snapshot shape aligned with source REST routes where possible.
- Future XML import plugin boundary.

Exit criteria:

- `/go-admin` can be reasoned about independently from production storage.
- Playground does not persist imported content in the binary or backend.
- Playground is not a reduced admin-only profile; it is a sandboxed full CMS runtime target.
- Existing browser-local content is never overwritten by silent source refresh.
- GraphQL can be implemented after this pass without owning runtime or storage decisions.

## Pass 4: GraphQL Plugin

Goal: install GraphQL as a plugin over the same core contracts.

Deliverables:

- `/go-graphql`.
- Schema generated from core resources.
- Queries for posts, pages, media, taxonomies, authors, menus, settings, and search.
- Mutations for authorized management operations where desired.
- Plugin settings screen.
- Capability-aware resolvers.
- Plugin extension points for schema additions.

Exit criteria:

- GraphQL does not query storage directly.
- GraphQL respects draft/private visibility.
- GraphQL and REST return consistent resource semantics.

## Pass 5: Marketplace-Style Frontend Validation

Goal: prove that the API can power a complete theme-like frontend.

Deliverables:

- Separate frontend validation app or fixture.
- Home page.
- Header and nested menus.
- Footer menus.
- Blog archive.
- Single post.
- Page.
- Category archive.
- Tag archive.
- Author archive.
- Search.
- Featured/related posts.
- Breadcrumbs.
- Pagination.
- Media/featured images.
- SEO/Open Graph metadata.

Exit criteria:

- The frontend looks and behaves like a full marketplace CMS theme.
- Missing API/model capabilities are recorded as implementation gaps.

## Pass 6: Conformance And Hardening

Goal: move from MVP to compatibility confidence.

Deliverables:

- Conformance fixture runner.
- REST compatibility tests.
- Admin workflow tests.
- Capability tests.
- Plugin lifecycle tests.
- Theme/rendering tests where enabled.
- GraphQL extension tests.
- Security and leakage tests.

Exit criteria:

- Drafts/private data do not leak through any surface.
- All required compatibility levels are tested.
- Known deviations are documented.
