# 08. Conformance Tests

This document defines expectations for GoCMS compatibility testing.

## Purpose

Conformance tests verify observable behavior. They do not inspect source code, package layout, database schema, router implementation, UI framework, or template engine.

An implementation is compatible only when it passes the tests for its declared compatibility level and documents any optional profiles.

## Test Principles

Conformance tests MUST:

- Test public behavior.
- Use required URLs and resource contracts.
- Avoid implementation-specific assumptions.
- Be repeatable.
- Be isolated.
- Provide actionable failure messages.
- Support authenticated and unauthenticated scenarios.

Conformance tests MUST NOT:

- Require a specific database.
- Require a specific Go framework.
- Require a specific UI library.
- Require a specific template engine.
- Require internal package names.
- Require a specific build system.

## Compatibility Levels

The suite SHOULD support these levels:

- Level 0: Core.
- Level 1: REST.
- Level 2: Admin.
- Level 3: Extension.
- Level 4: Full.

Each level MUST include all tests from lower levels.

## Required Test Fixtures

The test suite SHOULD be able to create or load fixtures for:

- Admin user.
- Editor user.
- Low-privilege user.
- Public visitor.
- Draft post.
- Published post.
- Scheduled post.
- Draft page.
- Published page.
- Media item.
- Taxonomy term.
- Menu.
- Theme.
- Plugin.

Fixtures SHOULD use public APIs, admin workflows, or documented setup commands rather than direct database writes.

## Level 0: Core Tests

Core tests SHOULD verify:

- Required content statuses exist.
- Draft content is not public.
- Scheduled content is hidden before publish time.
- Published content is visible through authorized read paths.
- Slugs are stable.
- Content IDs are stable.
- Capabilities exist or documented equivalents exist.
- Users without capabilities cannot perform protected actions.
- Private metadata is not exposed publicly.

## Level 1: REST Tests

REST tests SHOULD verify:

- `GET /go-json` returns discovery.
- `GET /go-json/go/v2/` returns namespace discovery.
- Required REST endpoints exist.
- Public list endpoints return envelope shape.
- Pagination metadata includes `page`, `per_page`, `total`, and `total_pages`.
- Invalid pagination returns stable validation errors.
- Draft content is hidden from unauthenticated callers.
- Authenticated users with capabilities can read allowed private data.
- Unauthorized users receive `401`, `403`, or documented `404` masking.
- Error shape includes `code`, `message`, and `status`.
- Unknown response fields do not break schema compatibility.
- Plugin namespaces cannot collide with reserved namespace `go`.

## Level 2: Admin Tests

Admin tests SHOULD verify:

- `GET /go-admin` blocks unauthenticated access.
- `GET /go-login` is available.
- Login creates an authenticated session.
- Logout invalidates the session.
- Required admin screens are discoverable or reachable.
- Admin menu items are capability-aware.
- Content can be created as draft.
- Draft can be published by an authorized user.
- Low-privilege users cannot publish without capability.
- Invalid CSRF/action token is rejected for state-changing actions.
- Bulk actions require authorization.
- Media upload rejects invalid files.
- Plugin activation requires capability.
- Theme activation requires capability.

Browser automation MAY be used, but tests SHOULD prefer stable HTML markers, REST calls, or documented test hooks over fragile CSS selectors.

## Level 3: Extension Tests

Extension tests SHOULD verify:

### Themes

- Theme manifest is validated.
- Invalid theme cannot be activated.
- Failed activation preserves previous active theme.
- Template resolution follows contract order.
- Theme assets are registered.
- Theme slots are available.

### Plugins

- Plugin manifest is validated.
- Plugin can be installed.
- Plugin can be activated.
- Activation failure does not mark plugin active.
- Plugin routes become available only while active.
- Plugin hooks run only while active.
- Plugin admin menu items are capability-aware.
- Plugin settings validate input.
- Plugin deactivation removes routes and hooks.
- Plugin uninstall follows data retention policy.

### Hooks

- Hook priority order is deterministic.
- Same-priority order is deterministic.
- Filters propagate returned values.
- Action failure policy is honored.
- Hook diagnostics identify owner and handler.

## Level 4: Full Tests

Full tests SHOULD verify:

- All lower levels pass.
- Public rendering profile works when enabled.
- Headless profile disables public rendering without breaking REST.
- Theme rendering enforces content visibility.
- Search excludes draft/private content.
- Revisions can be created and restored.
- Preview access is protected.
- Settings cache invalidates after write where cache is enabled.
- Import/export preserves stable fields.
- Audit log records high-risk actions when audit support is enabled.

## Required Test Outputs

The conformance runner SHOULD output:

- Contract version tested.
- Implementation name.
- Declared compatibility level.
- Enabled profiles.
- Passed tests.
- Failed tests.
- Skipped optional tests.
- Warnings for SHOULD-level recommendations.

Machine-readable output SHOULD be available as JSON.

## Profiles

Tests MUST account for declared profiles.

Examples:

- `headless`: public rendering tests may be skipped, REST tests remain required.
- `graphql`: GraphQL extension tests may run in addition to REST tests.
- `private-admin`: admin tests may require network or token configuration.
- `no-comments`: comment-related tests are skipped if comments are optional and disabled.

Profiles MUST NOT be used to skip core security or authorization tests.

## Seed And Cleanup

The conformance suite SHOULD use isolated test data and cleanup after itself.

Test-created resources SHOULD use a recognizable prefix such as:

```text
gocms_conformance_
```

Cleanup failures SHOULD be reported but SHOULD NOT hide the original test failure.

## Compatibility Report

An implementation SHOULD publish a compatibility report including:

- Contract version.
- Compatibility level.
- Profiles.
- Known deviations.
- Unsupported optional features.
- Date tested.
- Test suite version.

## Failure Classification

Failures SHOULD be classified as:

- `required`: violates a MUST.
- `recommended`: violates a SHOULD.
- `optional`: optional behavior missing.
- `security`: violates required security behavior.
- `unstable`: behavior changes between repeated runs.

Security failures MUST fail the relevant compatibility level.

## Future Test Areas

Future conformance suites MAY add:

- GraphQL extension tests.
- CLI tests.
- Scheduler tests.
- Import/export format tests.
- Accessibility tests for admin screens.
- Performance budget tests.
- Internationalization tests.
- Multi-site or multi-tenant profile tests.
