# 07. Background Jobs

This document defines background job guidance for GoCMS implementations.

## Purpose

Background jobs handle work that should not block user requests.

Examples:

- Media variant generation.
- Search indexing.
- Sitemap generation.
- Feed generation.
- Webhook delivery.
- Email sending.
- Import/export.
- Plugin scheduled tasks.
- Cache warming.
- Orphan cleanup.

## Job Descriptor

Jobs should have stable descriptors:

- ID.
- Owner.
- Trigger.
- Timeout.
- Retry policy.
- Idempotency key strategy.
- Required capability for manual execution.
- Visibility in admin diagnostics.

Plugin-provided jobs should be namespaced by plugin ID.

## Scheduler

A scheduler should support:

- One-time jobs.
- Recurring jobs.
- Manual trigger.
- Missed schedule handling.
- Locking.
- Graceful shutdown.
- Failure recording.

Scheduler state should survive process restarts when jobs matter for correctness.

## Idempotency

Jobs should be idempotent whenever possible.

Examples:

- Regenerating a media variant should overwrite or skip consistently.
- Reindexing content should be repeatable.
- Webhook delivery should use delivery IDs.
- Import jobs should support resume or dry-run mode.

## Context And Cancellation

Every job should receive `context.Context`.

Jobs must:

- Observe cancellation.
- Respect timeouts.
- Release resources on shutdown.
- Avoid leaking goroutines.

## Error Handling

Job failures should record:

- Job ID.
- Owner.
- Error code.
- Message.
- Attempt count.
- Started timestamp.
- Finished timestamp.
- Request or correlation ID where available.

Failures should be visible to admin diagnostics if the admin profile is implemented.

## Security

Jobs must not bypass:

- Content visibility.
- Capabilities.
- Plugin active state.
- Upload validation.
- Private settings visibility.

Manual job execution should require capabilities.

## Plugin Jobs

Plugin jobs should be registered only while the plugin is active.

Deactivation should:

- Stop future scheduling.
- Allow current execution to finish or cancel according to policy.
- Preserve job history unless cleanup is requested.

## Queue Independence

The architecture should not assume a specific queue library.

Implementations may use:

- In-process workers.
- Database-backed queues.
- External queues.
- Cron-like schedulers.
- Hosted task systems.

The behavior contract is stable job descriptors, lifecycle, retries, cancellation, and diagnostics.

## Testing

Background job tests should verify idempotency, cancellation, retry behavior, plugin deactivation behavior, and no public leakage from derived outputs.
