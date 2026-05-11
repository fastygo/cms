# 10. Technical Debt Policy

This document defines how GoCMS accepts, records, and pays down technical debt.

## Principle

Technical debt is allowed only when it is explicit, bounded, and scheduled.

Unrecorded debt is not allowed.

## Debt Categories

### Compatibility Debt

Debt that affects public behavior:

- Missing endpoint.
- Missing field.
- Temporary error shape.
- Incomplete status handling.
- Partial capability behavior.
- Missing import mapping.

This is the highest-risk debt because it can block one-click migration or API compatibility.

### Architecture Debt

Debt that affects internal boundaries:

- Adapter bypassing a service.
- Storage detail leaking into domain.
- Temporary duplicated business rule.
- Plugin runtime shortcut.
- Missing transaction boundary.

### Test Debt

Debt that affects confidence:

- Missing conformance test.
- Missing service test.
- Missing security leakage test.
- Missing migration test.
- Missing admin workflow test.

### UI Debt

Debt that affects admin profile quality:

- Missing accessible marker.
- Temporary non-reusable component.
- Missing ARIA runtime test.
- Styling policy exception.
- Missing progressive enhancement fallback.

### Migration Debt

Debt that affects one-click migration:

- Source field not mapped.
- Media import incomplete.
- Slug/redirect mapping incomplete.
- Metadata mapping incomplete.
- Author/user mapping incomplete.
- Plugin/theme migration gap.

## Debt Record Format

Every accepted debt item must record:

```text
ID:
Category:
Context:
Reason:
Risk:
Exit condition:
Owner:
Target milestone:
Blocking before:
```

## Acceptance Rules

Debt may be accepted when:

- It does not leak private data.
- It does not corrupt persisted data.
- It does not break already declared compatibility.
- It has a written exit condition.
- It has a target milestone.

Debt must not be accepted when:

- It weakens authorization.
- It breaks draft/private visibility.
- It prevents migration of core entities.
- It changes public API without versioning.
- It makes future migrations impossible.
- It hides a failed plugin/theme activation state.

## Paydown Rules

Debt must be paid down before:

- Raising compatibility level.
- Declaring a milestone complete.
- Building GraphQL over incomplete REST semantics.
- Building frontend validation over incomplete content model.
- Publishing a migration/import workflow.

## Temporary Shortcuts

Allowed temporary shortcuts:

- In-memory repository for early service tests.
- Simple local media storage for MVP.
- Minimal admin UI before final reusable blocks.
- Limited GraphQL read schema before mutations.
- Fixture-based frontend validation before production frontend.

Forbidden shortcuts:

- Public API fields without stable meaning.
- Direct storage access from GraphQL resolvers.
- Direct storage access from admin handlers.
- Hard-coded content types only as `post` and `page`.
- Ignoring custom taxonomies.
- Ignoring author privacy.
- Ignoring capability checks.

## Debt Register

Create a debt register when the first implementation shortcut is accepted.

Recommended path:

```text
.project/debt/en/debt-register.md
```

Keep the register concise and current. Closed debt should remain recorded with close date and resolution.
