# GoCMS Pass Progress

This file tracks implementation progress across the GoCMS pass roadmap.

Source roadmap:

- [`implementation/en/00-roadmap.md`](implementation/en/00-roadmap.md)

Detailed pass plans:

- [`implementation/en/01-core-kernel.md`](implementation/en/01-core-kernel.md)
- [`implementation/en/02-rest-control-plane.md`](implementation/en/02-rest-control-plane.md)
- [`implementation/en/03-admin-mvp.md`](implementation/en/03-admin-mvp.md)
- [`implementation/en/03-5-runtime-profiles-playground.md`](implementation/en/03-5-runtime-profiles-playground.md)
- [`implementation/en/03-6-configurator-plugin-foundation.md`](implementation/en/03-6-configurator-plugin-foundation.md)
- [`implementation/en/04-graphql-plugin.md`](implementation/en/04-graphql-plugin.md)
- [`implementation/en/05-theme-frontend-validation.md`](implementation/en/05-theme-frontend-validation.md)
- [`implementation/en/06-entity-coverage-checklist.md`](implementation/en/06-entity-coverage-checklist.md)
- [`implementation/en/08-engineering-principles.md`](implementation/en/08-engineering-principles.md)
- [`implementation/en/09-architecture-fitness-functions.md`](implementation/en/09-architecture-fitness-functions.md)
- [`implementation/en/10-technical-debt-policy.md`](implementation/en/10-technical-debt-policy.md)

## Current Status

Current pass: **Pass 7: Production CMS Completion**.

Completed:

- [x] Pass 0: Project Skeleton.
- [x] Pass 1: Core Compatibility Kernel.
- [x] Pass 2: REST Control Plane.
- [x] Pass 3: Headless Admin MVP.
- [x] Pass 3.5: Runtime Profiles And Playground Boundary.
- [x] Pass 3.6: Configurator And Plugin Foundation.
- [x] Pass 4: GraphQL Plugin.
- [x] Pass 5: Theme/Public API Foundation.
- [x] Pass 5.5: CMS Admin Finalization On Panel Core.

Pass 5 completion details:

- [x] Hybrid theme foundation slice is complete: theme contract, built-in default theme, permalink settings/resolver, `Themes` and `Permalinks` admin screens, and `full`-mode public rendering foundation.
- [x] Render checklist slice is complete: native installed themes (`gocms-default`, `blank`), public render contract/page assembler, blog/category-tag/author/search/content rendering, menus, media, breadcrumbs, SEO, admin activation, style preset selection, preview state, focused tests, and `bun run verify`.
- [x] Theme folders now keep compiled Go/templ logic and static asset conventions aligned at the theme/package level, while still separating code from `web/static` assets so styles can be changed through static files, tokens, and presets.
- [x] Public frontend validation slice is complete: REST/GraphQL projections and the first compiled project/company theme package validate the structure needed for future marketplace-grade frontend work.

In progress:

- [ ] Pass 7: Production CMS Completion.

Dependency direction:

```text
Core Kernel -> REST Control Plane -> Admin GUI -> Runtime Profiles + Playground -> Configurator + Plugin Foundation -> GraphQL Plugin -> Theme/Public API Foundation -> CMS Admin Finalization -> Conformance -> Marketplace Theme Finish
```

GraphQL, public rendering, playground, and the admin must consume the same core services. They must not redefine the core model. Plugin and theme marketplaces are intentionally excluded from the CMS core because plugins and themes are compiled into the binary and will be handled as separate distribution work.

## Pass 0: Project Skeleton

Status: **Done**.

Reference:

- [`implementation/en/00-roadmap.md#pass-0-project-skeleton`](implementation/en/00-roadmap.md#pass-0-project-skeleton)

Checklist:

- [x] Go module exists.
- [x] Application entrypoint exists.
- [x] Configuration loading exists.
- [x] Structured logging exists.
- [x] Health endpoints are wired.
- [x] Basic system feature is registered.
- [x] Package boundary directories exist.
- [x] Basic tests run with `go test ./...`.

Notes:

- Pass 0 intentionally contains no CMS business logic.

## Pass 1: Core Compatibility Kernel

Status: **Done**.

Reference:

- [`implementation/en/01-core-kernel.md`](implementation/en/01-core-kernel.md)

Checklist:

- [x] Domain packages exist for content, content types, taxonomies, media, users/authors, authz, settings, menus, revisions, and preview.
- [x] Content supports `post`, `page`, and custom kinds.
- [x] Content statuses include `draft`, `scheduled`, `published`, `archived`, and `trashed`.
- [x] Content entries include localized title, slug, content/body, excerpt, author, featured media, template, metadata, timestamps, and term references.
- [x] Custom content types are first-class.
- [x] Built-in `post` and `page` content types are registered through the same registry path.
- [x] Taxonomy definitions support category, tag, and custom taxonomies.
- [x] Terms can be created and assigned to content.
- [x] Public author projection is separated from private user data.
- [x] Capabilities are granular and checked by write workflows.
- [x] Media metadata can be attached as featured media.
- [x] Revisions can be created and restored.
- [x] Preview access model exists without HTTP URLs.
- [x] Public content queries hide drafts and scheduled content before publish time.
- [x] Focused domain and service tests exist.
- [x] `go test ./...` passes.

Follow-up gaps intentionally deferred:

- [x] Durable persistence adapters.
- [x] Query filtering by slug, taxonomy, status, author, search, locale, and date range.
- [x] Stable public/private projection DTOs for API delivery.
- [x] Seed fixtures for integration and conformance tests.

## Pass 2: REST Control Plane

Status: **Done**.

Reference:

- [`implementation/en/02-rest-control-plane.md`](implementation/en/02-rest-control-plane.md)

Goal:

Expose the compatibility REST surface over the Pass 1 services.

Checklist:

- [x] Decide first durable storage adapter for local development and tests.
- [x] Wire application services into a runtime CMS feature.
- [x] Add `/go-json` discovery document.
- [x] Add `/go-json/go/v2/` namespace discovery document.
- [x] Add stable resource envelope and list envelope.
- [x] Add stable error envelope.
- [x] Add pagination metadata.
- [x] Add locale negotiation.
- [x] Add public read endpoints for published posts and pages.
- [x] Add detail by ID.
- [x] Add detail by slug.
- [x] Add content list filters: `kind`, `status`, `author`, `taxonomy`, `search`, `locale`, `after`, `before`.
- [x] Add content types endpoint.
- [x] Add taxonomy definitions endpoint.
- [x] Add terms endpoint.
- [x] Add media metadata endpoint.
- [x] Add public author profile endpoint.
- [x] Add public settings endpoint.
- [x] Add public menus endpoint.
- [x] Add authenticated write endpoints for create/update/publish/schedule/trash/restore.
- [x] Add authenticated taxonomy and term management endpoints.
- [x] Add authenticated media metadata and featured media operations.
- [x] Ensure all writes call application services.
- [x] Ensure public REST does not expose drafts, scheduled content, private metadata, private user data, or private settings.
- [x] Add REST tests for public reads, authenticated writes, capability failures, errors, and pagination.
- [x] Run `go test ./...`.

Exit criteria:

- [x] `/go-json` returns discovery.
- [x] `/go-json/go/v2/` returns namespace discovery.
- [x] Published content lists work.
- [x] Detail by ID works.
- [x] Detail by slug works.
- [x] Drafts are hidden publicly.
- [x] Authenticated admin/editor can create, update, and publish.
- [x] Low-privilege user cannot publish.
- [x] Error shape is stable.
- [x] Pagination is stable.
- [x] API can provide enough data for the admin MVP.

## Pass 3: Headless Admin MVP

Status: **Done**.

Reference:

- [`implementation/en/03-admin-mvp.md`](implementation/en/03-admin-mvp.md)

Goal:

Implement `/go-admin` as the first complete management GUI while public rendering can remain disabled.

Checklist:

- [x] Login screen.
- [x] Logout flow.
- [x] Dashboard with counts and status.
- [x] Posts list/create/edit/publish workflow.
- [x] Pages list/create/edit/publish workflow.
- [x] Content type management.
- [x] Taxonomy definition management.
- [x] Term management.
- [x] Media metadata/select/featured-media attachment fields.
- [x] Menu management.
- [x] User and author management.
- [x] Role and capability management screen.
- [x] Settings screen.
- [x] API/headless settings screen.
- [x] Content edit screen includes title, slug, editor, excerpt, status, publish/schedule controls, author, featured media, taxonomy assignment, metadata/custom fields, save draft, publish, trash, and restore workflow coverage.
- [x] UI hides inaccessible actions.
- [x] Server rejects forbidden direct submissions.
- [x] Low-privilege users cannot access forbidden actions.

Exit criteria:

- [x] Admin can manage posts.
- [x] Admin can manage pages.
- [x] Admin can manage custom content types.
- [x] Admin can manage taxonomies.
- [x] Admin can register media metadata and select featured media by ID.
- [x] Admin can manage menus.
- [x] Admin can manage authors/users.
- [x] Admin can review capability groups.
- [x] Admin can configure headless mode.

Verification:

- [x] `bun run verify` passes.
- [x] `templ generate ./...` passes.
- [x] UI architecture tests enforce no raw app `.templ` tags, import boundaries, `@apply` app CSS, and raw palette bans.
- [x] `bun run lint:ui8px` (pinned `ui8px` in `package.json`) passes.
- [x] `go test ./...` passes.

Notes:

- Persistent custom role storage, revision browser UI, and preview URL workflows remain candidates for later hardening after the GraphQL/plugin pass. Pass 3 establishes the protected headless admin surface, reusable UI layers, workflows, action-token checks, and conformance gates.

## Pass 3.5: Runtime Profiles And Playground Boundary

Status: **Done**.

Reference:

- [`implementation/en/03-5-runtime-profiles-playground.md`](implementation/en/03-5-runtime-profiles-playground.md)

Goal:

Define the binary/runtime boundary before GraphQL: how GoCMS runs as headless, admin, playground, full, or conformance profile; how admin fixtures differ from mutable site content; and how playground targets a complete isolated CMS experience without server-side persistence.

Checklist:

- [x] Define runtime profiles: `headless`, `admin`, `playground`, `full`, and `conformance`.
- [x] Define storage profiles: browser IndexedDB, memory, JSON fixtures, SQLite, MySQL, and PostgreSQL.
- [x] Clarify that `playground` is a full isolated CMS sandbox target, not a reduced admin-only profile.
- [x] Separate admin UI fixtures from site content data.
- [x] Move admin-only fixture data toward pure embedded JSON files.
- [x] Ensure playground site content starts empty until source import or user import.
- [x] Define browser-local IndexedDB content store for playground.
- [x] Define IndexedDB Blob store for user-uploaded images.
- [x] Define one-time source bootstrap through `?gocms=<source>`.
- [x] Ensure `/go-admin` works without the query parameter after first browser-local import.
- [x] Ensure existing browser-local content skips external `wp-json`/compat source calls.
- [x] Define demo auth profile with `admin` / `admin` for playground only because the environment is isolated.
- [x] Define playground's target as admin plus public preview/rendering inside a browser-local or sandbox-backed runtime.
- [x] Enforce playground maximum of 10 posts unless the user deletes/trashes old content.
- [x] Add import JSON from device behavior.
- [x] Add export JSON to device behavior.
- [x] Ensure exported JSON excludes media binary payloads and includes only media metadata.
- [x] Define missing local Blob placeholder behavior with filename and preserved dimensions.
- [x] Keep playground JSON snapshots aligned with compatibility REST route shapes where possible.
- [x] Document future XML/WXR import as a plugin boundary, not as core playground bootstrap.

Exit criteria:

- [x] Runtime profile and storage profile decisions are documented.
- [x] Admin fixtures and mutable site content are separate concepts.
- [x] Playground is defined as a browser-local or sandbox-backed full CMS runtime target and is server-stateless.
- [x] `?gocms=<source>` is initializer-only and never required after first import.
- [x] Existing browser-local content is never overwritten by silent source refresh.
- [x] Media Blob export/import behavior is documented.
- [x] Future XML/WXR import has a documented plugin boundary.

## Pass 3.6: Configurator And Plugin Foundation

Status: **Done**.

Reference:

- [`implementation/en/03-6-configurator-plugin-foundation.md`](implementation/en/03-6-configurator-plugin-foundation.md)
- `go-codex/en/05-plugin-contract.md`
- `go-stack/en/05-plugin-runtime-strategies.md`
- `go-codex/en/06-hooks-contract.md`
- `go-codex/en/07-capabilities-contract.md`

Goal:

Create the deployment configurator and plugin foundation before GraphQL so GoCMS can behave like a WordPress-style core: deployment selects presets and active plugins/providers without modifying core code.

Configurator model:

- [x] Add a preset resolver so deployment can select one constant such as `GOCMS_PRESET=full`, `GOCMS_PRESET=headless`, or `GOCMS_PRESET=playground`.
- [x] Keep lower-level overrides available for runtime profile, storage profile, plugin set, and site package.
- [x] Add remaining lower-level policy overrides for auth policy and admin policy where presets need them.
- [x] Define first-class presets for the five admin/runtime modes:
  - Offline JSON fixtures binary + SQL import/export.
  - GUI over SSH tunnel for server-side site package management.
  - Typical full-stack CMS: admin + public site + DB.
  - Headless CMS: REST by default, GraphQL after plugin activation.
  - Playground: browser-local demo admin.
- [x] Ensure presets assemble runtime, storage, bind, plugin, and site-package plans through explicit typed configuration rather than ad-hoc env branching.

Bootstrap provider boundary:

- [x] Treat SQLite, Bbolt, Postgres, MySQL, JSON fixtures, memory, and browser IndexedDB as bootstrap providers, not ordinary runtime plugins.
- [x] Resolve bootstrap providers before application services, admin screens, routes, and plugin activation.
- [x] Add a storage provider registry for content, settings, media metadata, media blobs, plugin state, and site packages.
- [x] Require migration/restart/redeploy workflows for changing durable storage providers; storage provider changes must not be a casual admin toggle.
- [x] Keep JSON fixtures immutable when embedded into the binary; writable JSON content lives in an external site package provider.

Plugin foundation:

- [x] Add compile-time plugin descriptors as the first supported strategy.
- [x] Add initial plugin manifest validation for ID, name, version, and contract version.
- [x] Add plugin lifecycle state: installed, active, inactive, failed, uninstalled.
- [x] Add a plugin state repository.
- [x] Add activation/deactivation flow with rollback-safe registration.
- [x] Add plugin capability registration using the existing granular capability model.
- [x] Add plugin settings registration with plugin-prefixed keys and visibility rules.
- [x] Add plugin asset registration for admin/public surfaces.
- [x] Add plugin route registration for admin, public, and REST surfaces while keeping GraphQL as a future plugin surface.
- [x] Add hook registration foundations for actions and filters using the hooks contract.

Admin foundation:

- [x] Add an admin registry foundation so menu items, screen actions, routes, and assets can be registered by core modules or plugins.
- [x] Move hardcoded admin route registration toward registered admin screens.
- [x] Move fixture-backed admin navigation toward active plugin/core screen registration.
- [x] Keep the admin shell/layout in core.
- [x] Keep capability checks server-side for every registered screen and action.

First extraction candidates:

- [x] Extract `json-import-export` as the first plugin candidate for content/site snapshots.
- [x] Extract `playground` as a plugin candidate: `playground.js`, IndexedDB snapshot, `?gocms=...`, import/export/reset UX, browser Blob policy.
- [x] Keep GraphQL implementation explicitly deferred until after the plugin foundation exists.

Remaining 3.6 focus before Pass 4:

- [x] Add auth/admin policy override fields so presets can choose stricter login or operator-only admin behavior without code edits.
- [x] Move hardcoded core admin route registration toward registered core admin screens.
- [x] Move core admin navigation from fixture-only assembly toward core/plugin screen registration.
- [x] Document and enforce migration/restart/redeploy rules for durable bootstrap provider switches.
- [x] Add the first operator-facing bootstrap-provider status UX so the admin can display provider mode without pretending provider switches are a casual runtime toggle.

Keep in core:

- [x] Content model and application services.
- [x] Taxonomies, users/authors, capabilities, settings service, and core REST compatibility baseline.
- [x] Admin shell, plugin lifecycle engine, preset resolver, bootstrap provider registry, and stable storage ports.

Exit criteria:

- [x] Deployment can select a preset without changing core code.
- [x] Bootstrap providers are resolved before plugins and are documented as not-normal runtime plugins.
- [x] Compile-time plugin descriptors can register admin menu/actions, routes, settings, capabilities, assets, and hooks.
- [x] `json-import-export` is ready to become the first extracted plugin.
- [x] `playground` is ready to be gated behind plugin activation.
- [x] GraphQL work can start as a plugin rather than a new core surface.

## Pass 4: GraphQL Plugin

Status: **Done**.

Reference:

- [`implementation/en/04-graphql-plugin.md`](implementation/en/04-graphql-plugin.md)

Goal:

Install GraphQL as a plugin over the same core services used by REST and admin.

Checklist:

- [x] Add `/go-graphql` endpoint behind plugin enablement.
- [x] Add plugin settings definitions for endpoint status, introspection, auth policy, query limits, CORS, and cache policy.
- [x] Add schema coverage for posts, pages, content types, taxonomies, terms, media, authors, menus, settings, and search.
- [x] Add public queries for published public data.
- [x] Add authenticated queries for allowed private data.
- [x] Add focused mutations for create/update/publish/schedule/trash/restore workflows.
- [x] Add taxonomy assignment, media attach, menu save, and setting save mutations.
- [x] Ensure all resolvers use application services or stable read models.
- [x] Ensure GraphQL does not query storage directly.
- [x] Expose GraphQL plugin status/actions through existing admin and runtime surfaces.
- [x] Add draft/private leakage tests.
- [x] Add REST/GraphQL consistency tests for IDs, statuses, slugs, capability behavior, author data, media metadata visibility, and taxonomy assignments.
- [x] Run focused package tests and `bun run verify`.

Exit criteria:

- [x] GraphQL endpoint can be enabled as plugin.
- [x] Public queries return only published public content.
- [x] Authenticated queries can access allowed private data when capabilities allow it.
- [x] Mutations are capability-gated and service-backed.
- [x] Resolvers use services.
- [x] Draft leakage tests pass.
- [x] Schema covers enough data for the validation frontend.
- [x] `go test ./...` passes.
- [x] `bun run verify` passes.

## Pass 5: Theme/Public API Foundation

Status: **Done**.

Reference:

- [`implementation/en/05-theme-frontend-validation.md`](implementation/en/05-theme-frontend-validation.md)

Goal:

Prove that the API and public renderer can power a theme-like frontend, while keeping the first hybrid foundation slice for built-in themes, permalink settings, and safe `full`-mode public rendering inside the same pass instead of introducing a separate Pass 4.5. The final marketplace-grade public theme finish is now deferred to Pass 7.

Pass 5 should converge on a Go-native theme model:

- Native themes are installed at build time as compiled Go/templ packages, for example `internal/themes/gocmsdefault`, `internal/themes/blank`, and project/company themes.
- Admin activates installed themes and edits theme settings; it does not upload or execute arbitrary Go code.
- Each theme package can own its renderer, view components, section components, and manifest.
- Theme static assets live with the deployable static tree under clear theme/preset paths, for example `web/static/themes/<theme-id>/...` and `web/static/presets/<preset-id>/...`.
- Theme code and static assets remain separated, but are packaged together in the same preset/dist bundle so a theme can be deployed as one unit.
- Styling should be swappable through static assets, CSS variables, tokens, and BrandOSS-style preset JSON rather than through runtime code upload.
- Brand/style presets can be imported and previewed by admin later; they should configure visual identity, not replace the native theme renderer.
- Marketplace-grade visual polish, advanced theme sections, and final public theme behavior are deferred to Pass 7 so Pass 5 can remain the public API/rendering foundation.

Checklist:

- [x] Add theme contract foundation: manifest types, validation, activation model, and built-in default theme declaration.
- [x] Add permalink foundation for home, pages, posts, taxonomy archives, search, and controlled 404 resolution.
- [x] Add `Themes` admin screen under `/go-admin/themes`.
- [x] Add `Permalinks` admin screen under `/go-admin/permalinks`.
- [x] Persist first-pass theme/permalink settings through the existing settings service.
- [x] Add `full`-mode public renderer mounted after admin/API/plugin routes.
- [x] Add minimal public home, content detail, search/archive, and 404 rendering.
- [x] Ensure public rendering uses application services and public-only data paths.
- [x] Add focused route, capability, permalink, and leakage tests for the hybrid foundation slice.
- [x] Run `bun run verify` after the hybrid foundation slice.
- [x] Add native installed theme packages: `gocms-default` and `blank`.
- [x] Extend the theme registry to register native theme implementations, not only manifests.
- [x] Add theme activation in admin for installed themes.
- [x] Add theme style preset selection in admin.
- [x] Define the theme/static folder convention: compiled theme package plus static assets packaged under `web/static/themes/<theme-id>/`.
- [x] Define brand preset import/preview/apply contract for BrandOSS-generated JSON.
- [x] Ensure style preset changes do not require changing theme business logic.
- [x] Add tests for theme activation, static asset resolution, preset selection, and public render output.
- [x] Render home page.
- [x] Render header and nested menus.
- [x] Render footer menus.
- [x] Render blog archive.
- [x] Render single post.
- [x] Render page.
- [x] Render category archive.
- [x] Render tag archive.
- [x] Render author archive.
- [x] Render search.
- [x] Render featured and related posts.
- [x] Render breadcrumbs.
- [x] Render pagination.
- [x] Render media and featured images.
- [x] Render SEO/Open Graph metadata.
- [x] Add a first project/company public theme package to validate compiled theme structure without coupling it to CMS admin internals. `internal/themes/company` is registered as a native compiled theme with its static entrypoint under `web/static/themes/company/`.
- [x] Create a separate public frontend validation fixture that consumes `cmspanel` content projections instead of admin handler internals.
- [x] Validate the public frontend through REST projections exposed by `cmspanel.PostsResource`, `cmspanel.PagesResource`, media, taxonomy, and menu contracts.
- [x] Validate the public frontend through GraphQL projections exposed by the GraphQL plugin for the same `cmspanel` resource contracts.
- [x] Record missing public API, resource projection, theme, and model capabilities as implementation gaps for `cmspanel` and future `fastygo/panel` extraction.

Implementation gaps recorded by the public frontend validation slice:

- `cmspanel` now exposes public projection contracts for posts, pages, media, taxonomies, and menus, but only posts and pages are backed by full `panel.Resource` descriptors. Media, taxonomies, and menus still need dedicated `cmspanel` resources/pages before those contracts can move toward reusable `fastygo/panel` resource extraction.
- REST public projections cover posts, pages, media, taxonomies, and menus with stable fixture validation.
- GraphQL public projections cover posts, media, taxonomies, and menus with richer field validation; pages currently validate identity fields only because querying richer page body/excerpt/link projection fields collapses the response to `data:null`, which should be fixed before treating GraphQL page projections as complete.
- The `company` theme validates native compiled project theme structure and public rendering independence from admin handlers; it is still a first structural theme, not the final marketplace-grade visual/theme capability target.

Exit criteria:

- [x] `full` mode can serve a minimal public site at `/` without breaking admin, REST, GraphQL, or static routes.
- [x] The admin declares themes and permalink settings in a WordPress-style direction.
- [x] Public rendering uses application services and public projections, not storage.
- [x] Missing public API, resource projection, theme, and model capabilities are recorded as implementation gaps.

## Pass 5.5: CMS Admin Finalization On Panel Core

Status: **Done**.

Reference:

- `github.com/fastygo/framework` remains the lightweight application, HTTP, auth/session, locale, middleware, and rendering foundation.
- `github.com/fastygo/panel` is the reusable Filament-like control-plane package that GoCMS consumes through CMS-specific descriptors.
- GoCMS should finish its WordPress-style admin through `cmspanel`, while theme/plugin marketplaces remain separate compiled-distribution work.

Goal:

Finish the GoCMS admin as a coherent WordPress-style management surface on top of `cmspanel` and `github.com/fastygo/panel`, while keeping CMS domain behavior in GoCMS and generic control-plane contracts in `panel`.

Scope:

- `cmspanel.PostsResource`, `cmspanel.PagesResource`, `cmspanel.MediaResource`, `cmspanel.TaxonomiesResource`, `cmspanel.MenusPage`, `cmspanel.ThemesPage`.
- `cmspanel.SettingsPage`, `cmspanel.HeadlessPage`, `cmspanel.RuntimePage`, and related CMS-only pages should describe admin intent and navigation without moving rendering or domain services into `panel`.
- Multi-panel products outside GoCMS are out of scope for this roadmap.

Completed foundation:

- [x] Define the first `github.com/fastygo/panel` boundary through the standalone `github.com/fastygo/panel` module: surfaces, routes, navigation/menu items, assets, editor providers, panel descriptors, resources, pages, widgets, actions, schemas, policies, and optional data-source contracts.
- [x] Keep `github.com/fastygo/framework` lightweight and make `panel` consume framework primitives instead of merging panel code into the framework.
- [x] Move generic admin registry concepts out of GoCMS naming: routes, menu items, assets, and editor providers now live in `github.com/fastygo/panel`; CMS-specific screen actions, capabilities, hooks, and settings descriptors remain in `internal/platform/plugins` until their contracts are split.
- [x] Introduce typed resource contracts for table schema, form schema, record actions, bulk actions, filters, sorting, search, validation, and relation fields.
- [x] Introduce page and widget contracts for arbitrary screens, dashboards, reports, setup wizards, and runtime status panels.
- [x] Add action contracts for header actions, row actions, bulk actions, modal actions, confirmation flows, and action-local forms.
- [x] Add optional data-source contracts for typed service-backed resources.
- [x] Document what remains CMS-specific in GoCMS and what belongs to reusable `fastygo/panel`.
- [x] Convert Posts and Pages to real `cmspanel` resources using `panel.Resource`, `panel.TableSchema`, `panel.FormSchema`, `panel.Action`, and `panel.ResourceRoute` metadata.
- [x] Add GoCMS-only adapters that consume Posts/Pages panel descriptors without moving admin rendering into `panel`.
- [x] Remove duplicate admin screen titles/descriptions after moving page-level metadata into the common `Screen` header.
- [x] Add public projection contracts for posts, pages, media, taxonomies, and menus so the public frontend/API validation can track the same CMS resource boundaries.

CMS admin panelization slice:

- [x] Panelize media admin as a `cmspanel.MediaPage` with table/form/action descriptors.
- [x] Panelize taxonomy definitions and terms as `cmspanel.TaxonomiesPage` and `cmspanel.TermsPage` descriptors.
- [x] Panelize menus as a `cmspanel.MenusPage` with menu item form/table/action descriptors.
- [x] Panelize settings, headless status, runtime status, themes, and permalinks as CMS-specific `cmspanel.Page` descriptors.
- [x] Reduce hardcoded core admin route/menu registration so `admin.Handler` consumes `cmspanel` resources/pages for every core CMS section.
- [x] Keep compiled plugins and compiled themes visible in admin, but leave plugin/theme marketplace installation outside the CMS core roadmap.

Exit criteria:

- [x] GoCMS admin can be described as `cmspanel` resources/pages on top of the panel core for posts, pages, media, taxonomies, terms, menus, settings, headless, runtime, themes, and permalinks.
- [x] Panel core stays independent from CMS domain services and UI packages.
- [x] All CMS admin sections are functional, capability-gated, and covered by focused admin workflow tests.

## Pass 6: Conformance And Hardening

Status: **Done**.

Reference:

- [`implementation/en/06-entity-coverage-checklist.md`](implementation/en/06-entity-coverage-checklist.md)
- [`implementation/en/09-architecture-fitness-functions.md`](implementation/en/09-architecture-fitness-functions.md)

Goal:

Move from MVP to compatibility confidence.

Checklist:

- [x] Add conformance fixture runner.
- [x] Add REST compatibility tests.
- [x] Add admin workflow tests.
- [x] Add capability tests.
- [x] Add plugin lifecycle tests.
- [x] Add theme/rendering tests where enabled.
- [x] Add GraphQL extension tests.
- [x] Add security and leakage tests.
- [x] Verify all required compatibility levels.
- [x] Document known deviations.

Exit criteria:

- [x] Drafts and private data do not leak through any surface.
- [x] All required compatibility levels are tested.
- [x] Known deviations are documented.

## Pass 7: Production CMS Completion

Status: **In progress**.

Goal:

Finish the remaining production-grade CMS capabilities after conformance baseline is stable. This pass keeps the GoCMS core Go-native: plugins and themes remain compiled into the binary, while extensibility is exposed through descriptors, typed registries, hooks, manifests, and stable public/admin surfaces rather than runtime marketplace installation.

Scope:

- Align playground with the WordPress Playground-style goal: a one-click isolated CMS sandbox with admin, public preview/rendering, browser-local or ephemeral storage, blueprints, import/export, and embeddable demos.
- Finish Go-native developer/plugin APIs that cover plugin descriptors, route/assets/settings/capability registration, hook registration and dispatch, content render filters, theme assets through manifests, editor providers, and public/admin surfaces.
- Add a typed custom fields / Meta API foundation on top of existing `content.Metadata`: meta key registry, schemas, validation, field UI, ownership/capability rules, public/private rules, typed metadata registry, and service-layer hooks.
- Keep media core lightweight: validated metadata, public URLs, provider refs, and BlobStore/plugin boundaries in core, while binary uploads, crop/edit, attachment pages, and responsive renditions remain plugin/provider follow-up work.
- Improve admin list UX across CMS sections: bulk actions, quick edit, filters, sorting, search, screen options, and pagination preferences.
- Add a typed Options/Settings API: option registry, autoload/public/private behavior, validation, settings groups, and plugin/theme option ownership.
- Harden authentication: real password storage/login through the users repository, password reset, app passwords/API tokens, rate limiting, and safer session/auth policies.
- Harden operations: migrations, backup/restore, provider health checks, audit logs, and admin-visible error logs.
- Finish the public-facing marketplace-style theme experience on top of compiled native theme packages and BrandOSS-style visual presets.

Out of scope:

- Runtime marketplace installation of arbitrary plugin or theme code.
- WXR/XML import unless introduced by a dedicated plugin.
- Full WordPress REST edge-case parity beyond the compatibility levels selected in Pass 6.
- Comments and advanced revisions UI unless required by a concrete product slice.

Checklist:

Architecture readiness for offline/browser/serverless targets:

- [x] Add a deployment profile layer that separates deployment target (`local`, `browser`, `serverless`, `container`, `ssh`) from runtime and storage profiles.
- [x] Document current `browser-indexeddb` as a transitional browser-local shim and define the future IndexedDB/WASM provider boundary.
- [x] Add provider capability planning for repositories, migrations, health checks, snapshots, and blob storage without replacing the current `bootstrap.Store` yet.
- [x] Restrict fixture login and dev bearer auth from leaking into serverless/container production-style deployments.
- [x] Introduce a media `BlobStore` boundary for local filesystem, browser IndexedDB blobs, object storage, and memory-backed sandbox media.

Go-native developer/plugin APIs:

- [x] Add a hook dispatcher for registered plugin hooks, with action/filter categories and deterministic priority ordering.
- [x] Add content render filters for public rendering and REST/GraphQL projections where extension points are safe.
- [x] Stabilize plugin descriptors for routes, assets, settings, capabilities, hooks, editor providers, and public/admin surfaces.
- [x] Keep theme assets resolved through compiled theme manifests and preset/static asset manifests.
- [x] Document Go-native equivalents for WordPress developer APIs without copying PHP template tags, shortcodes, or enqueue APIs one-to-one.

Playground finish:

- [x] Enable playground to expose both admin and public preview/rendering inside the isolated runtime.
- [x] Add blueprint-style startup configuration for playground demos and QA scenarios.
- [x] Add embeddable playground launch support for one-click demos, previews, and iframe-style integrations.
- [x] Keep unauthenticated playground admin access safe by confining it to browser-local or ephemeral sandbox storage.
- [x] Keep playground import/export aligned with compatibility snapshots and future plugin importers.

Custom fields / Meta API:

- [x] Add a typed metadata key registry with owner, scope, schema, public/private visibility, capability requirements, and validation rules.
- [x] Add field UI generation for registered meta keys in admin editors and relevant `cmspanel` descriptors.
- [x] Add service-layer hooks around metadata validation, persistence, and public projection.
- [x] Preserve existing flexible `content.Metadata` storage while enforcing registered meta behavior when schemas exist.

Media library:

- [x] Extend media metadata with validated public URLs, dimensions, MIME policy, and provider/ref fields aligned with BlobStore boundaries.
- [x] Keep admin and REST media workflows metadata-only in core for CDN/object-storage/external assets.

### Deferred for the plugin

- [>] Add plugin/provider binary upload support for local, object storage, and CDN-backed media.
- [>] Add plugin/provider generated variants, responsive image sizes, and image crop/edit workflows.
- [>] Add plugin/provider attachment page/read model support where public rendering enables it.

Admin list UX:

- [x] Add bulk actions for content and users, with descriptor-backed selection flows and capability re-checks server-side where current services can safely execute them.
- [x] Add quick edit forms for content, media metadata, taxonomies/terms, menus, and users on top of existing protected save paths.
- [x] Add filters, sorting, search controls, pagination, and real list state parsing to descriptor-backed lists.
- [x] Add screen options and pagination preferences backed by typed admin option keys.

Options/settings API:

- [x] Add typed option definitions with validation, defaults, public/private visibility, and autoload policy.
- [x] Add option groups for core, plugins, themes, public rendering, headless/API, and operational settings.
- [x] Route existing settings screens and public setting projection through the typed option registry.

Authentication hardening:

- [x] Persist password hashes through the users repository and replace fixture login for non-demo profiles.
- [x] Add password reset flow and account recovery policy.
- [x] Add app passwords or API tokens for authenticated API use.
- [x] Add login rate limiting and session policy hardening.

Operational hardening:

- [x] Add migrations for durable providers.
- [x] Add backup and restore workflows.
- [x] Add provider-specific health checks.
- [x] Add audit logs for privileged admin/API actions.
- [x] Add admin-visible error logs and operational diagnostics.

Public theme finish:

- [ ] Build the final marketplace-style public theme experience on top of compiled native theme packages.
- [ ] Expand theme sections, visual presets, responsive behavior, and public UX beyond the structural `company` validation theme.
- [ ] Keep public themes independent from `github.com/fastygo/panel` and admin internals.

Exit criteria:

- [x] Go-native plugin/developer APIs can support compiled plugins without runtime code installation.
- [x] Playground behaves as a full isolated CMS sandbox rather than an admin-only demo.
- [x] Registered meta fields can be validated, edited, persisted, and projected safely.
- [x] Media core supports validated metadata, public URLs, and provider refs, while upload and rendition pipelines remain explicit plugin/provider follow-up work.
- [x] Admin list screens support production-grade list workflows.
- [x] Settings/options are typed, validated, grouped, and visibility-aware.
- [x] Authentication no longer depends on fixture login outside demo/playground profiles.
- [x] Operations have migrations, backup/restore, health, audit, and error-log visibility.
- [ ] The public frontend looks and behaves like a full marketplace CMS theme while remaining independent from the reusable panel core.

## Update Rules

- Update this file at the end of every pass.
- Keep statuses limited to `Done`, `In Progress`, `Next`, `Planned`, or `Blocked`.
- Do not mark a pass as `Done` until `go test ./...` passes and exit criteria are checked.
- When a checklist item becomes too large, split it into the pass-specific implementation plan instead of overloading this tracker.
