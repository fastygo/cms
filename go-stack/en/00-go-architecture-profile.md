# 00. Go Architecture Profile

This document defines the general Go architecture rules for a GoCMS-compatible implementation.

## Principles

The implementation should be explicit, testable, and replaceable at boundaries.

Rules:

- Pass `context.Context` through all request, job, storage, and service calls.
- Prefer typed constants over unstructured string literals for stable identifiers.
- Keep domain objects independent from HTTP, storage drivers, rendering, sessions, and background workers.
- Keep application services independent from delivery protocols.
- Keep persistence behind interfaces owned by the application.
- Keep composition in a small root where dependencies are wired deliberately.
- Avoid hidden global registries for mutable application state.
- Use package-level registries only for immutable constants or documented extension descriptors.
- Return typed or classified errors that can be mapped to compatibility error responses.

## Suggested Layering

```text
cmd/
internal/domain/
internal/application/
internal/storage/
internal/delivery/
internal/runtime/
internal/platform/
pkg/gocms/
```

Recommended responsibilities:

- `cmd`: composition roots and process startup.
- `internal/domain`: content, media, users, settings, hooks, themes, plugins, and capability entities.
- `internal/application`: services, commands, queries, validation, transactions, and authorization orchestration.
- `internal/storage`: repository implementations and migration adapters.
- `internal/delivery`: REST, GraphQL, admin, CLI, webhook, and public rendering adapters.
- `internal/runtime`: background jobs, scheduling, plugin runtime, cache warming, and shutdown coordination.
- `internal/platform`: cross-cutting adapters such as logging, metrics, tracing, and configuration.
- `pkg/gocms`: public extension SDK if plugin authors need stable Go imports.

## Public SDK Boundary

If plugin authors are expected to compile against GoCMS APIs, expose only stable extension-facing contracts under a public package tree.

Public contracts should include:

- Content read/write service interfaces.
- Hook registration interfaces.
- Plugin manifest types.
- Capability constants.
- Theme extension interfaces.
- REST route registration contracts.
- Background job descriptors.

Do not expose internal repository implementations as the plugin SDK.

## Error Model

Application errors should carry:

- Stable code.
- Human-readable message.
- Optional cause.
- Optional field errors.
- Optional conflict metadata.

Delivery adapters map errors to REST, GraphQL, admin, CLI, or job failure shapes.

## Configuration

Configuration should be environment-driven and explicit.

The implementation should support:

- Base URL.
- Admin path.
- REST path.
- Public rendering mode.
- Storage configuration.
- Session configuration.
- Locale configuration.
- Upload limits.
- Cache settings.
- Plugin and theme directories or registries.

Configuration should be validated at startup.

## Concurrency

Services and handlers should be safe for concurrent use unless documented otherwise.

Long-running resources must have shutdown paths:

- Database pools.
- Cache cleaners.
- Schedulers.
- Media processors.
- Search indexers.
- Plugin processes.
- Webhook dispatchers.

## Compatibility Rule

Internal architecture may change freely as long as the observable behavior in `../../go-codex/en` remains compatible.
