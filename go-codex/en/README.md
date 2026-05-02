# GoCMS Compatibility Contract

This directory defines the GoCMS Compatibility Contract. It is a public, implementation-neutral specification for engines, admin clients, themes, plugins, REST clients, and conformance tools that want to interoperate with GoCMS-compatible systems.

The contract describes observable behavior, resource shapes, URL surfaces, lifecycle events, extension points, and compatibility tests. It does not prescribe a database, router, UI framework, template engine, build system, storage adapter, queue system, or plugin runtime.

## Status

Version: draft `0.1`

This is the first English draft. A GoCMS-compatible implementation may use this version as a design target, but a production conformance badge should only be issued once the test suite described in `08-conformance-tests.md` exists as executable checks.

## Normative Language

The key words `MUST`, `MUST NOT`, `REQUIRED`, `SHOULD`, `SHOULD NOT`, `RECOMMENDED`, `MAY`, and `OPTIONAL` are to be interpreted as described in `00-compatibility-principles.md`.

## Required URL Surface

The base URL surface is GoCMS-native:

```text
/go-admin
/go-login
/go-logout
/go-json
/go-json/go/v2/
```

Implementations MAY expose additional URLs, but they MUST NOT break the required GoCMS paths unless explicitly operating in a documented compatibility profile that disables them.

## Documents

- `00-compatibility-principles.md` defines the compatibility model, versioning, deprecation, and breaking-change policy.
- `01-admin-contract.md` defines required admin screens, actions, states, and browser-facing behavior.
- `02-rest-api-contract.md` defines REST discovery, resources, auth, pagination, filtering, errors, and response shape rules.
- `03-content-contract.md` defines posts, pages, statuses, slugs, excerpts, revisions, preview, autosave, and metadata.
- `04-theme-contract.md` defines theme manifests, template roles, slots, assets, and resolution rules.
- `05-plugin-contract.md` defines plugin manifests, lifecycle, migrations, routes, hooks, admin menu entries, settings, and assets.
- `06-hooks-contract.md` defines abstract actions and filters with priorities, argument contracts, error handling, and observability.
- `07-capabilities-contract.md` defines granular permissions and how roles, users, admin actions, REST endpoints, themes, and plugins use them.
- `08-conformance-tests.md` defines compatibility levels and expected tests.

## Compatibility Layers

GoCMS compatibility is split into layers:

- Core resource compatibility: content, users, media, taxonomies, settings, menus, themes, plugins, and capabilities.
- Admin compatibility: required screens, URLs, actions, states, and error behavior.
- REST compatibility: required endpoints, schemas, auth behavior, pagination, and error shape.
- Extension compatibility: plugin manifests, hook behavior, theme behavior, and capability registration.
- Conformance compatibility: repeatable tests that verify observable behavior.

An implementation can be partially compatible only if it declares which layers it supports.

## Implementation Independence

This contract is intentionally independent from any concrete Go stack. The same contract can be implemented by:

- A server-rendered admin interface.
- A single-page admin interface.
- A hybrid admin interface.
- A monolith.
- A modular binary.
- A multi-process deployment.
- A headless-only deployment with admin GUI and public rendering disabled by policy.

The implementation details are not part of this contract. Public behavior is.

## Headless And GraphQL

GraphQL support is considered an extension profile. A GraphQL endpoint SHOULD use the same content, capability, plugin, and hook contracts defined here. A GraphQL plugin MUST NOT bypass core authorization, lifecycle, or resource invariants.

REST remains the base compatibility and control-plane contract even when GraphQL is the primary content delivery API.
