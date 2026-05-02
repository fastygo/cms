# 05. Plugin Contract

This document defines the GoCMS plugin compatibility contract.

## Purpose

Plugins extend the CMS without modifying core behavior directly. A plugin may add routes, admin screens, hooks, settings, capabilities, content types, taxonomies, media behaviors, REST endpoints, background jobs, or theme slots.

## Implementation Independence

This contract does not require a specific plugin loading mechanism.

An implementation MAY support plugins as:

- Compile-time modules.
- Dynamically discovered modules.
- External processes.
- Sandboxed runtimes.
- Any other documented mechanism.

Regardless of runtime model, observable plugin behavior MUST follow this contract.

## Plugin Manifest

Each plugin MUST provide a manifest.

Required fields:

```json
{
  "id": "example-plugin",
  "name": "Example Plugin",
  "version": "1.0.0",
  "contract": "0.1",
  "description": "A plugin.",
  "author": "Example",
  "requires": {},
  "capabilities": [],
  "routes": [],
  "hooks": [],
  "assets": []
}
```

Required manifest fields:

- `id`
- `name`
- `version`
- `contract`

Plugin IDs MUST be stable, lowercase, URL-safe, and globally unique within an installation.

## Lifecycle

Plugins MUST support these lifecycle states:

- `installed`
- `active`
- `inactive`
- `failed`
- `uninstalled`

Plugins SHOULD support these lifecycle operations:

- Install.
- Activate.
- Deactivate.
- Update.
- Uninstall.

## Install

Install MUST:

- Validate manifest.
- Validate contract compatibility.
- Register plugin metadata.
- Prepare plugin storage if needed.
- Run install migrations if declared.

Install MUST NOT activate the plugin unless the user explicitly requested activation or a documented install-and-activate flow is used.

## Activate

Activate MUST:

- Validate installed state.
- Validate dependencies.
- Run pending migrations.
- Register routes, hooks, capabilities, admin menu items, settings, jobs, and assets.
- Mark plugin active only after successful activation.

If activation fails, the plugin MUST NOT become active.

## Deactivate

Deactivate MUST:

- Stop plugin-provided routes from being exposed.
- Stop plugin-provided hooks from running.
- Stop plugin-provided scheduled jobs.
- Preserve plugin data unless uninstall is requested.

Deactivation MUST be reversible unless the plugin explicitly documents otherwise.

## Uninstall

Uninstall MUST:

- Require explicit confirmation.
- Deactivate the plugin if active.
- Run uninstall cleanup if declared.
- Remove plugin metadata according to configured policy.

Data deletion MUST be explicit. Implementations SHOULD allow preserving plugin data.

## Migrations

Plugin migrations MUST be:

- Versioned.
- Idempotent or guarded against repeated execution.
- Ordered.
- Auditable.
- Reversible where possible.

Migration failures MUST leave plugin state recoverable.

## Routes

Plugins MAY register:

- Public routes.
- Admin routes.
- REST routes under `/go-json/{plugin-id}/vN/`.
- Webhook routes.

Plugin routes MUST:

- Declare required capabilities when protected.
- Avoid reserved path collisions.
- Use stable route identifiers.
- Return stable errors.

## Admin Menu

Plugins MAY add admin menu items.

Menu entries MUST declare:

- Stable ID.
- Label.
- Target path.
- Required capability.
- Ordering.
- Parent group where relevant.

Plugin menu items MUST be hidden from users without the required capability.

## Settings

Plugins MAY register settings.

Plugin settings MUST declare:

- Key.
- Type.
- Default value.
- Validation rule.
- Public/private visibility.
- Required capability for write access.

Plugin setting keys SHOULD be prefixed by plugin ID.

## Assets

Plugins MAY register admin or public assets.

Asset declarations SHOULD include:

- Stable ID.
- Type.
- Source.
- Dependencies.
- Target surface: admin, public, REST docs, or specific plugin screen.

Plugin assets MUST respect security policy and MUST NOT override core assets by default.

## Capabilities

Plugins MAY define capabilities. Capability IDs SHOULD be prefixed by plugin ID.

Plugins MUST NOT grant capabilities directly to users without going through the engine's authorization model.

## Hooks

Plugins MAY register actions and filters. Hook behavior MUST follow `06-hooks-contract.md`.

Plugin hook failures MUST be isolated according to hook error policy. A failing plugin MUST NOT corrupt core state.

## Background Jobs

Plugins MAY register scheduled or background jobs.

Jobs MUST declare:

- Stable ID.
- Schedule or trigger.
- Required capability for manual execution.
- Timeout expectation.
- Retry behavior where applicable.

Jobs SHOULD be idempotent.

## Dependency Rules

Plugins MAY declare dependencies on:

- Core contract version.
- Other plugins.
- Required capabilities.
- Optional features such as REST, themes, or media.

Dependency failures MUST prevent activation and provide a clear error.

## Conformance Notes

Conformance tests SHOULD verify:

- Manifest validation.
- Activation failure rollback.
- Route registration and removal.
- Hook registration and removal.
- Admin menu capability enforcement.
- Migration ordering.
- Uninstall behavior.
