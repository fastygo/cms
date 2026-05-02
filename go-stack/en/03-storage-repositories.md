# 03. Storage And Repositories

This document defines storage boundaries for GoCMS implementations.

## Purpose

Storage is an implementation detail. Domain and application services should depend on repository interfaces or equivalent ports, not concrete storage drivers.

## Repository Ownership

Repository interfaces should be owned by the package that consumes them.

Examples:

- `ContentRepository` is defined near content application services.
- `MediaRepository` is defined near media application services.
- `SettingsRepository` is defined near settings application services.

Concrete implementations live in infrastructure packages.

## Required Repository Capabilities

Repositories should support:

- Lookup by ID.
- Lookup by slug where applicable.
- List with filters and pagination.
- Create.
- Update.
- Soft delete or trash where applicable.
- Restore where applicable.
- Permanent delete where applicable.
- Transaction participation where needed.

## Transactions

The application should provide a transaction boundary abstraction.

The abstraction should:

- Accept `context.Context`.
- Support nested call safety or reject nesting clearly.
- Commit on success.
- Roll back on error.
- Propagate transaction-scoped repositories.

Do not leak concrete driver transaction types into domain packages.

## IDs And Timestamps

Entities should have stable IDs and timestamps.

Rules:

- IDs must be stable across API responses.
- Creation timestamps should not change.
- Update timestamps should change on mutation.
- Published timestamps should reflect visibility rules.
- Deleted timestamps should support restoration when soft deletion is used.

ID type may be string, integer, UUID-like, or another stable type, but API serialization must be stable.

## Metadata

Metadata storage should support:

- Content metadata.
- User metadata.
- Term metadata.
- Media metadata.
- Plugin-owned metadata.
- Private/public visibility flags.

Frequently queried metadata should have a documented indexing strategy.

## Migrations

Migrations should be:

- Versioned.
- Ordered.
- Repeat-safe.
- Logged.
- Recoverable on failure where possible.

Plugin migrations should be separated from core migrations but integrated into activation/update workflows.

## Search Storage

Search indexes should be treated as derived data.

Rules:

- Search can be rebuilt from source content.
- Draft and private content must not leak into public search indexes.
- Index updates should be triggered by service workflows or background jobs.

## Media Storage

Media storage should separate:

- Binary object storage.
- Media metadata.
- Variant metadata.
- URL resolution.

The repository should not construct public URLs with string concatenation. Use a resolver service so deployment changes do not corrupt stored metadata.

## Backup And Restore

Storage implementation should document:

- Database backup.
- Media backup.
- Plugin data backup.
- Theme settings backup.
- Restore order.
- Consistency expectations.

## Conformance Adapter

Conformance tests should use public APIs, admin workflows, or documented setup commands. Direct database writes should be avoided unless a test profile explicitly exposes a storage fixture adapter.
