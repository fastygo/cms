# 02. Application Services

This document defines application service boundaries for a GoCMS-compatible implementation.

## Purpose

Application services coordinate domain rules, validation, authorization, transactions, repositories, hooks, cache invalidation, audit logging, and delivery-independent workflows.

Delivery adapters call services. Services should not know whether a request came from admin, REST, GraphQL, CLI, public rendering, or a background job.

## Recommended Services

Core services:

- `ContentService`
- `TaxonomyService`
- `MediaService`
- `UserService`
- `AuthService`
- `AuthorizationService`
- `SettingsService`
- `MenuService`
- `ThemeService`
- `PluginService`
- `HookService`
- `RevisionService`
- `SearchService`
- `AuditService`

Optional services:

- `CommentService`
- `RedirectService`
- `ImportExportService`
- `WebhookService`
- `SchedulerService`

## Command And Query Shape

Write workflows should be represented as commands. Read workflows should be represented as queries.

Commands should include:

- Principal or actor.
- Locale where relevant.
- Idempotency key where relevant.
- Input payload.
- Request metadata for audit when available.

Queries should include:

- Principal or visitor state.
- Locale.
- Pagination.
- Filters.
- Visibility context.

This is a conceptual pattern. It does not require a specific dispatcher package.

## Authorization

Services must authorize protected operations before mutating state or returning private data.

Rules:

- Admin adapters must not be trusted to authorize alone.
- REST adapters must not be trusted to authorize alone.
- GraphQL resolvers must not be trusted to authorize alone.
- Plugin calls must not bypass service authorization.

## Transactions

Services should own transaction boundaries for workflows that mutate multiple resources.

Examples:

- Create content and assign taxonomies.
- Publish content and update search visibility.
- Activate plugin and register migrations.
- Update settings and invalidate cache.
- Upload media and attach variants.

External side effects should run after commit whenever possible.

## Hooks And Events

Services should trigger hook points around important workflows.

Recommended pattern:

1. Validate input.
2. Authorize.
3. Run before hooks.
4. Start transaction.
5. Mutate state.
6. Commit.
7. Run after-commit hooks.
8. Invalidate cache.
9. Audit.

Hooks that can block mutation should run before commit. Hooks that perform external side effects should run after commit.

## Cache Invalidation

Services that mutate state must invalidate affected caches.

Examples:

- Content changes invalidate content detail, lists, menus, feeds, sitemaps, and search.
- Settings changes invalidate public settings and rendering context.
- Theme changes invalidate render caches.
- Plugin changes invalidate route, hook, and admin menu registries.

## Result Types

Services should return typed results rather than storage records.

Results should be safe for delivery adapters to map to:

- Admin view models.
- REST DTOs.
- GraphQL objects.
- CLI output.
- Job logs.

## Validation

Input validation should be close to the command/query boundary.

Validation should distinguish:

- Missing required fields.
- Invalid formats.
- Unsupported state transitions.
- Capability failures.
- Conflicts.
- Storage failures.

Delivery adapters may perform shallow parsing validation, but service validation is authoritative.

## Auditing

High-risk workflows should emit audit records:

- Login and logout.
- User and role changes.
- Content publish and delete.
- Settings changes.
- Plugin and theme activation.
- API token changes.

Audit emission should not expose secrets or large payloads.
