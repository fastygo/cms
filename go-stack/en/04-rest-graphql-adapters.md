# 04. REST And GraphQL Adapters

This document defines delivery adapter guidance for REST and GraphQL.

## Core Rule

REST and GraphQL are adapters over the same application services. They must not own separate business rules, separate authorization rules, or direct storage shortcuts.

```text
Admin -> Services
REST -> Services
GraphQL -> Services
CLI -> Services
Plugins -> Services
```

## REST

REST is the base compatibility and control-plane surface described by `../../go-codex/en/02-rest-api-contract.md`.

REST adapters should:

- Parse HTTP input.
- Authenticate principal.
- Build commands or queries.
- Call services.
- Map results to stable DTOs.
- Map service errors to compatibility errors.
- Enforce cache headers after visibility decisions.

REST adapters should not:

- Query storage directly.
- Reimplement publish visibility.
- Bypass capability checks.
- Return storage models as responses.

## GraphQL

GraphQL should be treated as an optional extension adapter.

GraphQL resolvers should:

- Use the same services as REST.
- Respect the same capabilities.
- Respect content visibility.
- Reuse the same ID and status semantics.
- Expose private fields only with authorization.
- Support pagination with stable cursors or documented page arguments.

GraphQL schema extensions should be registered through plugin or extension descriptors.

## Headless Mode

In headless mode:

- Public HTML rendering may be disabled.
- REST compatibility remains available as the control plane unless explicitly declared otherwise.
- GraphQL may be the primary content delivery surface.
- Admin should expose which delivery surfaces are enabled.

Headless mode must not weaken authorization.

## DTOs And View Models

Adapters should define DTOs close to delivery code.

Application service result types should be stable enough to map into:

- REST DTOs.
- GraphQL types.
- Admin view models.
- Public render view models.

Do not use one delivery DTO as the internal domain model.

## Error Mapping

Service errors should map consistently:

- Not found -> REST `not_found`, GraphQL not found error.
- Validation -> field errors.
- Unauthorized -> authentication error.
- Forbidden -> capability error.
- Conflict -> conflict error.
- Internal -> generic internal error with request ID.

Internal causes should be logged, not exposed.

## Uploads

Media uploads may use REST even when GraphQL is enabled.

Recommended approach:

- REST handles upload transport.
- Media service validates and stores asset.
- GraphQL references uploaded media by ID.

GraphQL upload transport may be supported, but it must call the same media service.

## Schema Discovery

REST discovery and GraphQL schema introspection should both reflect active plugins and capabilities where appropriate.

Private routes, fields, or mutations should not be advertised to principals that cannot use them unless the implementation explicitly exposes public schema metadata.

## Testing

Adapter tests should verify:

- Same service rules through REST and GraphQL.
- Same authorization outcomes.
- Same content visibility outcomes.
- Same plugin extension visibility.
- Same error classification.
