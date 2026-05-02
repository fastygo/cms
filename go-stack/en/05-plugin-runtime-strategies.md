# 05. Plugin Runtime Strategies

This document defines Go-native plugin runtime strategies and invariant behavior.

## Purpose

The external plugin contract defines observable lifecycle behavior. This profile explains implementation strategies for Go systems while preserving that behavior.

## Runtime Options

Supported strategies may include:

- Compile-time modules linked into the binary.
- Dynamically registered modules discovered at startup.
- External plugin processes communicating over an internal protocol.
- Sandboxed modules.
- Hosted plugin services.

An implementation may support one or more strategies.

## Required Invariants

Every strategy must preserve:

- Manifest validation.
- Install, activate, deactivate, update, and uninstall lifecycle.
- Capability registration.
- Hook registration and removal.
- Route registration and removal.
- Settings registration.
- Migration ordering.
- Failure rollback.
- Auditability.

Runtime choice must not change compatibility behavior.

## Compile-Time Modules

Compile-time modules are the simplest and safest strategy for early versions.

Benefits:

- Strong Go type checking.
- Simple deployment.
- No runtime binary loading.
- Easy service injection.

Costs:

- Plugin changes require rebuild.
- Marketplace-style installation requires build automation.

Compile-time modules should expose descriptors rather than mutate global state at import time.

## Dynamic Modules

Dynamic modules may be loaded at startup from a registry or directory.

They must:

- Validate manifest before activation.
- Be isolated from core state until activation succeeds.
- Provide clear failure diagnostics.
- Support deterministic ordering.

## External Processes

External process plugins can isolate failures and dependencies.

They should:

- Communicate through a stable protocol.
- Authenticate calls.
- Have timeouts.
- Have health checks.
- Have restart policy.
- Be unable to bypass core authorization.

External plugin process failure should not corrupt core state.

## Sandboxed Modules

Sandboxed modules may be useful for untrusted extensions.

They should:

- Limit CPU and memory.
- Limit file and network access.
- Expose only approved host functions.
- Treat all inputs as untrusted.

Sandbox strategy is optional.

## Plugin Registry

The plugin registry should track:

- Plugin ID.
- Version.
- Manifest.
- State.
- Runtime strategy.
- Dependencies.
- Registered routes.
- Registered hooks.
- Registered capabilities.
- Registered settings.
- Migration version.
- Last error.

Registry writes should be transactional where possible.

## Activation Flow

Recommended activation flow:

1. Load descriptor.
2. Validate manifest.
3. Validate dependencies.
4. Validate required contract version.
5. Prepare runtime.
6. Run pending migrations.
7. Register capabilities.
8. Register settings.
9. Register routes, hooks, jobs, and assets.
10. Mark active.
11. Audit activation.

If any step fails before marking active, the plugin remains inactive or failed.

## Deactivation Flow

Recommended deactivation flow:

1. Stop scheduled jobs.
2. Remove routes.
3. Remove hook handlers.
4. Remove admin menu items.
5. Mark inactive.
6. Preserve data.
7. Audit deactivation.

## Service Access

Plugins should receive narrow service interfaces, not full storage or runtime handles.

Plugin access should be scoped by:

- Capabilities.
- Declared dependencies.
- Active state.
- Runtime strategy.

## Security

Plugins must not:

- Bypass authorization.
- Mutate core registries outside lifecycle operations.
- Access secrets unless explicitly granted.
- Expose private content through routes.
- Register conflicting reserved identifiers.

## Testing

Plugin runtime tests should verify lifecycle, rollback, route visibility, hook registration, capability registration, migration ordering, and failure isolation for each supported strategy.
