# 00. Compatibility Principles

This document defines the compatibility model for GoCMS-compatible engines and extensions.

## Goals

The contract exists to let users, theme authors, plugin authors, API clients, and conformance tools rely on stable behavior across GoCMS-compatible implementations.

The contract defines:

- Required URL surfaces.
- Stable resource names and states.
- Stable API shapes.
- Stable admin behaviors.
- Extension lifecycle expectations.
- Compatibility levels.
- Versioning and deprecation rules.

It does not define:

- Internal package layout.
- Database schema.
- Template engine.
- UI framework.
- Router implementation.
- Authentication storage.
- Plugin loading mechanism.
- Build or deployment tooling.

## Normative Terms

The following terms are normative:

- `MUST` means the behavior is required for the relevant compatibility level.
- `MUST NOT` means the behavior is forbidden.
- `SHOULD` means the behavior is strongly recommended, and deviations require a documented reason.
- `SHOULD NOT` means the behavior is discouraged, and deviations require a documented reason.
- `MAY` means the behavior is optional.
- `REQUIRED` has the same meaning as `MUST`.
- `OPTIONAL` has the same meaning as `MAY`.

## Compatibility Levels

An implementation MUST declare its supported compatibility level.

### Level 0: Core

Level 0 includes:

- Content contract.
- Capabilities contract.
- Required auth states for admin/API access.
- Stable IDs and slugs.
- Stable status names.
- Minimal REST discovery.

### Level 1: REST

Level 1 includes Level 0 plus:

- Required `/go-json` discovery.
- Required `/go-json/go/v2/` namespace.
- Required REST endpoints.
- Stable pagination, filtering, and error shapes.
- Authenticated REST behavior.

### Level 2: Admin

Level 2 includes Level 1 plus:

- Required `/go-admin` and `/go-login` behavior.
- Required admin screens.
- Required admin actions and state transitions.
- Browser security requirements for admin actions.

### Level 3: Extension

Level 3 includes Level 2 plus:

- Theme contract.
- Plugin contract.
- Hooks contract.
- Plugin-defined capabilities.
- Theme and plugin conformance tests.

### Level 4: Full

Level 4 includes all contracts in this directory and passes the full conformance suite.

## Versioning

The compatibility contract uses semantic versioning.

Version format:

```text
MAJOR.MINOR.PATCH
```

- `MAJOR` changes MAY introduce breaking changes.
- `MINOR` changes MAY add new optional or required behavior only if it is backward-compatible for existing compatible implementations.
- `PATCH` changes MUST NOT change required behavior. They may clarify wording, fix examples, or correct non-normative text.

## Breaking Changes

A change is breaking if it:

- Removes a required endpoint.
- Renames a required field.
- Changes the type of a required field.
- Changes the meaning of a required status.
- Changes required authorization behavior.
- Makes a previously valid request invalid without a deprecation period.
- Changes stable error codes.
- Changes hook argument order.
- Changes plugin lifecycle order.
- Changes theme template resolution in a way that selects a different template for existing valid themes.
- Changes required admin action semantics.
- Requires a concrete framework, UI library, database, or runtime not previously required.

## Non-Breaking Changes

A change is non-breaking if it:

- Adds an optional field to a JSON object.
- Adds a new optional endpoint.
- Adds a new optional hook.
- Adds a new optional capability.
- Adds a new optional admin screen.
- Adds a new optional theme slot.
- Clarifies undefined behavior without changing defined behavior.
- Fixes contradictory text by preserving the older stable behavior.

Clients MUST ignore unknown JSON object fields unless the relevant schema explicitly says otherwise.

## Deprecation Policy

Deprecation MUST be documented before removal.

Deprecation notices MUST include:

- Deprecated feature name.
- First version where it is deprecated.
- Earliest version where it may be removed.
- Replacement behavior.
- Migration notes.

Required behavior SHOULD remain available for at least one full minor release after deprecation before removal in a major release.

## Profiles

Implementations MAY define profiles such as:

- `public-rendering`: public HTML routes are enabled.
- `headless`: public rendering is disabled, REST remains available.
- `graphql`: a GraphQL extension is enabled.
- `private-admin`: admin is available only behind network-level access controls.

Profiles MUST NOT silently change core resource semantics. If a profile disables a required public surface, the implementation MUST document the compatibility level impact.

## Stable Identifiers

The following identifiers are stable and reserved:

- URL prefixes: `/go-admin`, `/go-login`, `/go-logout`, `/go-json`, `/go-json/go/v2`.
- Content kinds: `post`, `page`.
- Content statuses: `draft`, `scheduled`, `published`, `archived`, `trashed`.
- Base resource names: `posts`, `pages`, `media`, `taxonomies`, `users`, `settings`, `menus`, `themes`, `plugins`.
- Error object fields: `code`, `message`, `status`, `details`, `request_id`.
- Pagination fields: `page`, `per_page`, `total`, `total_pages`.

Extensions MUST NOT redefine stable identifiers with conflicting meaning.

## Security Compatibility

Compatibility MUST NOT weaken security.

An implementation MUST:

- Enforce authorization before returning private content.
- Protect admin state-changing actions against CSRF.
- Rate-limit authentication attempts or provide an equivalent defense.
- Sanitize user-generated rich content.
- Validate file uploads.
- Prevent draft content from leaking through REST, admin previews, search, or theme rendering.

## Conformance

Conformance tests SHOULD verify observable behavior only. A conformance test MUST NOT require a specific package structure, framework, database schema, template engine, or storage backend.

An implementation passes a contract only if:

- It supports the required behavior.
- It documents optional deviations.
- It rejects invalid input in a stable way.
- It does not expose incompatible behavior under required paths.
