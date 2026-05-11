# 09. Architecture Fitness Functions

This document defines automated and review-time checks that keep GoCMS architecture from drifting.

## Purpose

Fitness functions are guardrails. They catch architectural debt before it becomes expensive.

## Required Fitness Functions

### Contract Boundary

Checks:

- `go-codex/en` does not reference concrete implementation stacks.
- `go-stack/en` does not reference UI-specific implementation rules.
- `go-ui8kit/en` does not claim UI8Kit is required for general compatibility.
- Required URLs remain stable: `/go-admin`, `/go-login`, `/go-logout`, `/go-json`, `/go-json/go/v2/`.

### Import Boundaries

Checks:

- Domain packages do not import HTTP, storage drivers, rendering, session, or UI packages.
- Application services do not import delivery adapters.
- REST, GraphQL, admin, and CLI adapters call services instead of storage directly.
- Plugin code receives narrow service interfaces instead of infrastructure internals.

### Compatibility Behavior

Checks:

- Draft content is not returned publicly.
- Scheduled content is hidden before publish time.
- Private metadata is not returned publicly.
- User-private fields are not exposed as author data.
- Low-privilege users cannot publish.
- Admin actions require server-side capabilities.
- REST and GraphQL return consistent content semantics.

### API Stability

Checks:

- REST error shape remains stable.
- Pagination envelope remains stable.
- Resource IDs remain stable.
- Status strings remain stable.
- Hook IDs remain stable.
- Plugin manifest fields remain backward compatible.
- Theme manifest fields remain backward compatible.

### Migration Readiness

Checks:

- Every persisted schema change has a migration.
- Every plugin schema change has plugin migration metadata.
- Every public API breaking change has a version/deprecation note.
- Every import-mapped field has a target model field or documented unsupported reason.

### Runtime Profiles And Playground

Checks:

- Runtime profile choices are explicit and do not leak into core domain code.
- Admin JSON fixtures are readonly and separate from mutable site content.
- Playground content persistence is browser-local only.
- Playground source bootstrap runs only when browser-local content is empty.
- `/go-admin` does not require `?gocms=<source>` after first local import.
- Playground export JSON excludes binary media payloads.
- Missing browser-local media Blobs render deterministic placeholders.
- Compatibility snapshots stay aligned with REST route payload shapes where possible.

### UI Profile

Checks for UI8Kit profile implementations:

- Admin UI uses stable `data-gocms-*` markers.
- Critical admin workflows work without relying solely on client-side JavaScript.
- Dialogs, tabs, comboboxes, and disclosure patterns preserve ARIA state.
- Utility classes pass policy.
- Generated assets are not hand-edited.

## Suggested Automated Commands

Future implementation should provide commands equivalent to:

```bash
go test ./...
go test ./... -run Conformance
go vet ./...
```

Additional project scripts should eventually check:

- Forbidden imports.
- Contract terminology boundaries.
- REST compatibility.
- GraphQL consistency.
- Runtime profile and playground storage boundaries.
- Draft/private leakage.
- Plugin lifecycle.
- Theme/frontend validation.
- UI accessibility markers.

## Pull Request Checklist

Every implementation change should answer:

- Does this change public behavior?
- Does this require a contract update?
- Does this require a migration?
- Does this affect imports/layering?
- Does this affect one-click migration mapping?
- Does this affect REST and GraphQL consistency?
- Does this affect admin capabilities?
- Does this add technical debt?

## Blocking Conditions

The next milestone should be blocked when:

- Draft/private content leaks publicly.
- A service is bypassed by a delivery adapter.
- A persisted shape changes without migration.
- A public API changes without compatibility plan.
- A plugin/theme lifecycle change lacks rollback.
- Conformance tests cannot create required fixtures.
- One-click migration mapping loses required source data.

## Review-Time Fitness

Not all checks are automated at the start. Until automation exists, review manually:

- Layer boundaries.
- Contract compatibility.
- Migration impact.
- Security leakage.
- Extension point stability.
- Future one-click migration impact.
