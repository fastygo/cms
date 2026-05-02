# 08. Testing And Conformance Adapters

This document defines how Go implementations should expose tests and adapters for the external conformance suite.

## Purpose

The conformance suite in `../../go-codex/en/08-conformance-tests.md` verifies observable behavior. A Go implementation should provide stable ways to create fixtures, authenticate test actors, run migrations, and reset state without relying on private database writes.

## Test Surfaces

Recommended test surfaces:

- HTTP server test harness.
- Admin fixture setup command.
- REST fixture setup endpoint in test mode.
- CLI fixture setup command.
- Storage reset command for isolated test environments.
- Plugin install/activate test harness.
- Theme install/activate test harness.

Test-only surfaces must never be enabled accidentally in production.

## Fixture Adapter

A fixture adapter should support:

- Create admin user.
- Create editor user.
- Create low-privilege user.
- Create draft post.
- Create published post.
- Create scheduled post.
- Create draft page.
- Create published page.
- Upload media fixture.
- Create taxonomy term.
- Create menu.
- Install test theme.
- Install test plugin.

The adapter should return stable IDs and credentials for the conformance runner.

## Authentication Helpers

Tests should be able to obtain:

- Anonymous client.
- Admin client.
- Editor client.
- Low-privilege client.
- API token where supported.
- Session cookie where browser admin tests are used.

Authentication setup should use public or documented test interfaces.

## Determinism

Tests should control:

- Clock or publish timestamps.
- Locale.
- Base URL.
- Storage root.
- Plugin registry.
- Active theme.
- Cache state.

Where a real clock is used, tests should allow safe time windows.

## Isolation

Each test run should isolate:

- Database state.
- Media files.
- Search index.
- Cache.
- Plugin state.
- Theme state.
- Background job state.

Parallel test runs should use unique prefixes or isolated environments.

## Contract Assertions

Go tests should assert both:

- Internal service behavior.
- External conformance behavior.

Internal tests are useful but do not replace external compatibility tests.

## Golden Data

Golden API responses may be used for:

- REST error shapes.
- Discovery documents.
- Resource DTOs.
- Plugin manifests.
- Theme manifests.

Golden files should allow additional unknown fields where the external contract allows them.

## Security Tests

Security tests should verify:

- Draft content is hidden.
- Private metadata is hidden.
- Missing capabilities are rejected.
- Invalid action tokens are rejected.
- Upload validation works.
- Plugin routes disappear after deactivation.

## Reporting

The implementation should produce test reports that include:

- Contract version.
- Implementation version.
- Enabled profiles.
- Passed tests.
- Failed tests.
- Skipped optional tests.
- Known deviations.

## Continuous Compatibility

Compatibility tests should run in CI for:

- Core services.
- REST endpoints.
- Admin state transitions.
- Plugin lifecycle.
- Theme resolution.
- Authorization.
- Headless profile.
