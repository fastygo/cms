# 06. Hooks Contract

This document defines the GoCMS extension hook model.

## Purpose

Hooks let core modules, themes, and plugins observe events or transform values without depending on a specific implementation.

GoCMS defines two hook categories:

- Actions: observe or react to an event.
- Filters: receive a value and return a value.

## Hook Identifiers

Hook IDs MUST be stable strings.

Recommended format:

```text
domain.event
domain.event.phase
```

Examples:

```text
content.create.before
content.create.after
content.update.before
content.update.after
content.status.before
content.status.after
media.upload.after
theme.activate.after
plugin.activate.before
settings.update.after
```

Plugin-defined hook IDs SHOULD be prefixed by plugin ID:

```text
example-plugin.report.generate.before
```

## Actions

An action handler receives context and arguments, performs side effects, and returns success or failure.

Action handlers MUST NOT mutate arguments unless the hook explicitly allows mutation.

Actions SHOULD be used for:

- Auditing.
- Cache invalidation.
- Notifications.
- Search indexing.
- Webhook dispatch.
- Background job scheduling.

## Filters

A filter receives a value, may transform it, and returns the resulting value.

Filters SHOULD be used for:

- Rendered content transformation.
- Query modification through explicit query objects.
- REST response decoration.
- Theme data decoration.
- Menu item filtering.

Filters MUST preserve the declared value type. If a filter returns an invalid value, the engine MUST reject it or fail the hook according to error policy.

## Priority

Hook handlers MUST support priority ordering.

Rules:

- Lower priority numbers run earlier unless the implementation documents the opposite before version `1.0`.
- Default priority SHOULD be `100`.
- Handlers with the same priority MUST run in deterministic registration order.

## Arguments

Each stable hook MUST document:

- Hook ID.
- Category: action or filter.
- Argument names.
- Argument types.
- Whether arguments are mutable.
- Return type for filters.
- Error behavior.

Hooks MUST NOT change argument order in a compatible release.

## Context

Hooks SHOULD receive execution context containing:

- Request or job context when available.
- Current user or principal when available.
- Locale when relevant.
- Request ID or correlation ID when available.
- Capability checker or authorization context when relevant.

Hooks MUST NOT assume an HTTP request exists.

## Error Policy

Each hook MUST declare one error policy:

- `fail`: abort the operation if a handler fails.
- `log`: record failure and continue.
- `isolate`: disable or quarantine the failing extension after repeated failures.
- `collect`: collect failures and return an aggregated error after all handlers run.

Critical mutation hooks SHOULD use `fail`.

Notification and indexing hooks SHOULD use `log` or `collect`.

## Transaction Boundaries

Hooks MUST document whether they run:

- Before a transaction.
- Inside a transaction.
- After commit.
- After rollback.

External side effects SHOULD run after commit. Hooks inside transactions MUST avoid long-running work.

## Core Hook Set

Implementations SHOULD expose at least these hooks:

```text
content.create.before
content.create.after
content.update.before
content.update.after
content.status.before
content.status.after
content.trash.before
content.trash.after
content.restore.before
content.restore.after
media.upload.before
media.upload.after
media.delete.before
media.delete.after
settings.update.before
settings.update.after
theme.activate.before
theme.activate.after
plugin.activate.before
plugin.activate.after
plugin.deactivate.before
plugin.deactivate.after
rest.response.filter
render.content.filter
```

## Hook Registration

Hook registrations MUST include:

- Hook ID.
- Handler ID.
- Priority.
- Category.
- Owner ID.

Owner ID SHOULD identify the plugin, theme, or core module that registered the handler.

## Observability

Implementations SHOULD expose hook diagnostics:

- Registered handlers.
- Owner.
- Priority.
- Last execution time.
- Last error.
- Disabled or quarantined state.

Sensitive argument values MUST NOT be logged by default.

## Security

Hooks MUST NOT bypass:

- Capabilities.
- Content visibility.
- CSRF checks.
- Upload validation.
- REST auth.
- Private settings visibility.

If a hook modifies a query or response, the engine MUST still enforce final authorization before output.

## Conformance Notes

Conformance tests SHOULD verify:

- Priority order.
- Deterministic same-priority order.
- Filter return value propagation.
- Action failure policy.
- Hook removal on plugin deactivation.
- Hook argument compatibility.
