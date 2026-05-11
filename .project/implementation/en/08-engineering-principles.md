# 08. Engineering Principles

This document defines the engineering principles that protect GoCMS from avoidable technical debt.

## Primary Goal

GoCMS must evolve toward one-click migration from existing content-heavy CMS installations into GoCMS while preserving content, media, authors, taxonomies, slugs, menus, SEO metadata, plugins/themes where possible, and public API compatibility.

That goal makes backward compatibility a core engineering constraint from the first implementation pass.

## Core Principles

### Contract First

Public behavior is designed before implementation.

Required order:

1. Define or update contract.
2. Define DTOs and error shapes.
3. Define capability rules.
4. Define migration impact.
5. Define conformance tests.
6. Implement.

Implementation must not silently redefine public behavior.

### KISS

Prefer the smallest implementation that satisfies the current contract and keeps future extension points open.

Do not build speculative subsystems before a contract, test, or migration need exists.

### DRY

Avoid duplicated business rules.

Allowed duplication:

- DTO mapping for separate delivery protocols.
- Explicit test fixtures.
- Small local UI composition when abstraction would obscure behavior.

Forbidden duplication:

- Separate publish rules in REST and GraphQL.
- Separate authorization rules in admin and services.
- Separate slug lookup rules in frontend and backend.

### SOLID

Use SOLID as practical pressure, not ceremony:

- Single responsibility: domain, services, storage, delivery, and UI do not own each other's work.
- Open/closed: new content types, taxonomies, hooks, plugins, and API fields extend through registries/contracts.
- Liskov: implementations of repositories/services must preserve behavior promised by interfaces.
- Interface segregation: plugins receive narrow service interfaces.
- Dependency inversion: high-level workflows depend on ports, not concrete storage or delivery adapters.

### TDD And Contract Tests

Use tests before or alongside implementation for:

- Domain invariants.
- Application services.
- REST contract behavior.
- GraphQL consistency.
- Admin workflows.
- Plugin lifecycle.
- Theme/frontend validation.
- Migration behavior.

Unit tests are necessary but not sufficient. Compatibility behavior must be tested at API/workflow level.

### Ports And Adapters

External systems must sit behind ports:

- Database.
- Media storage.
- Search.
- Queue/scheduler.
- Email.
- Webhooks.
- REST adapter.
- GraphQL adapter.
- Admin UI.
- Theme renderer.
- Plugin runtime.
- Import/export.

Swapping an adapter must not require rewriting domain or application services.

### 12-Factor Posture

GoCMS should follow 12-factor principles where practical:

- Config in environment or explicit runtime config.
- Logs as event streams.
- Stateless app processes.
- Backing services attached by configuration.
- Build/release/run separation.
- Disposability through graceful shutdown.
- Dev/prod parity.
- Admin tasks as one-off commands.
- Concurrency through processes and workers.
- No hidden local state required for correctness.

### Migration-First Design

Every persisted shape must have an evolution story:

- Core schema migrations.
- Plugin migrations.
- Theme settings migrations.
- Metadata migrations.
- Content type migrations.
- API version migrations.
- Import mapping migrations.

If a field is public, persisted, or imported, changing it requires a compatibility plan.

### Backward Compatibility By Default

Breaking changes are exceptional.

Prefer:

- Adding optional fields.
- Adding new endpoints.
- Adding new capabilities.
- Deprecating before removing.
- Versioning incompatible behavior.
- Migration commands.
- Compatibility adapters.

Do not break existing exported content, slugs, IDs, media URLs, API response fields, hook names, plugin manifests, or theme manifests without an explicit major-version plan.

## One-Click Migration Constraint

One-click migration requires stable mapping for:

- Posts.
- Pages.
- Custom content types.
- Categories.
- Tags.
- Custom taxonomies.
- Authors.
- Users where safe.
- Media.
- Menus.
- Slugs.
- Redirects.
- SEO metadata.
- Excerpts.
- Revisions where available.
- Comments if enabled.
- Custom fields/metadata.
- Public settings.
- Theme settings where compatible.
- Plugin data where compatible.

Do not design core entities in a way that makes these mappings impossible.

## Decision Rule

When choosing between two implementations, prefer the one that:

1. Preserves contracts.
2. Localizes future changes.
3. Is testable through conformance.
4. Has a migration path.
5. Adds the fewest permanent dependencies.
