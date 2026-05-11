# 03.6. Configurator And Plugin Foundation

This document records the foundation delivered during Pass 3.6 so GraphQL, site packages, and future plugins can build on a stable preset/bootstrap/runtime model instead of adding new core-only entrypoints.

## Why This Pass Exists

Pass 3.5 fixed the runtime and playground boundary. Pass 3.6 turns that boundary into a deployment model:

- Deployment selects a preset instead of editing core code.
- Storage providers are resolved before services and plugin activation.
- Plugins are compiled into the binary through descriptors, then activated through lifecycle state.
- Admin extensions, plugin assets, plugin actions, and plugin routes are registered through shared registries instead of one-off wiring.

This keeps GraphQL, snapshot tooling, playground UX, and future extensions on the same application-service baseline.

## Delivered Preset Model

GoCMS now resolves runtime assembly through `internal/platform/preset` and `internal/platform/config`.

Preset entrypoint:

```text
GOCMS_PRESET=offline-json-sql
GOCMS_PRESET=ssh-fixtures
GOCMS_PRESET=full
GOCMS_PRESET=headless
GOCMS_PRESET=playground
```

Current preset behavior:

- `offline-json-sql`: local admin binary with SQLite and `json-import-export`.
- `ssh-fixtures`: admin-oriented preset for server-side site package workflows with SQLite and `json-import-export`.
- `full`: full runtime with SQLite and `json-import-export`.
- `headless`: REST-first runtime without admin exposure by default.
- `playground`: browser-stateless runtime with browser IndexedDB storage conventions, demo auth, and the `playground` plugin enabled.

Current lower-level overrides remain available for:

- Runtime profile.
- Storage profile.
- App bind address.
- Data source.
- Active plugin set.
- Site package directory.
- Dev bearer auth.
- Login policy.
- Admin policy.

The default preset can also be baked into a binary at build time with linker flags. `scripts/build-presets.mjs` exists to produce preset-specific binaries from the same codebase.

## Bootstrap Provider Boundary

GoCMS now resolves bootstrap providers before application services and plugin activation through `internal/infra/bootstrap`.

Current bootstrap provider responsibilities:

- Open the storage implementation for the selected storage profile.
- Resolve the plugin state repository.
- Resolve the external site package provider.
- Report the selected content provider name to the runtime composition root.

Currently declared storage profiles:

- `sqlite`
- `memory`
- `browser-indexeddb`
- `json-fixtures`
- `bbolt`
- `mysql`
- `postgres`

The first four are wired today. `bbolt`, `mysql`, and `postgres` are declared as bootstrap providers but still intentionally return not-implemented errors.

This establishes the enforced rule: durable storage providers are bootstrap-time deployment choices, not casual runtime plugin toggles. Changing `sqlite`, `bbolt`, `mysql`, `postgres`, or another durable provider requires an export/migration handoff, deployment configuration change, process restart or redeploy, and import/verification on the new provider. The admin UI may display the effective provider and migration rule, but it must not expose a live provider toggle.

## Delivered Plugin Foundation

GoCMS now has an initial compile-time plugin runtime in `internal/platform/plugins`.

Delivered foundation:

- Compile-time plugin descriptors expose `Manifest()` and `Register(...)`.
- Manifest validation exists for plugin ID, name, version, and contract version.
- Lifecycle states exist: `installed`, `active`, `inactive`, `failed`, `uninstalled`.
- A plugin state repository exists and is resolved by bootstrap.
- Activation writes plugin state only after successful registration.
- Activation failures are recorded as `failed`.
- Deactivation marks compiled plugins `inactive`.
- Only plugins compiled into the current binary can be activated.

Current registry surfaces support:

- Capabilities.
- Settings.
- Hooks.
- Assets.
- Routes.
- Admin menu items.
- Admin screen actions.

This is enough foundation for GraphQL to start as a plugin over the same core services instead of becoming a separate core runtime path.

## Admin Registration Foundation

The admin handler now consumes core and plugin registrations for:

- Admin routes.
- Admin navigation items.
- Screen actions.
- Extra admin assets.

Core screens now register stable screen IDs such as `dashboard`, `posts`, `pages`, `settings`, `headless`, and `runtime`. Fixtures supply localized labels and copy for those IDs; they no longer define the active admin topology. Core still owns the admin shell, layout, and capability enforcement. That is intentional: plugins extend the admin surface, but core remains responsible for shell integrity and server-side authorization.

## Operator Runtime Status

The admin now exposes `/go-admin/runtime` as an operator-facing status screen. It reports:

- Preset name.
- Runtime profile.
- Storage profile.
- Resolved content provider.
- Site package status.
- Active plugins.
- Browser-stateless mode.
- Playground auth.
- Dev bearer auth.
- Login policy.
- Admin policy.
- Durable provider switch rule.

This status screen is informational. Durable provider changes remain deployment operations handled through export/migration, restart or redeploy, and import/verification.

## First Extraction Candidates

Two concrete extraction candidates now exist as compile-time plugin descriptors:

- `json-import-export`: content snapshot import/export, site package import/export, settings registration, admin actions, admin routes, admin asset, and example hook/capability registration.
- `playground`: browser-local playground actions, admin asset registration, and hook registration for playground-specific UX.

These candidates prove the descriptor model against real features instead of placeholder examples.

## GraphQL Readiness

GraphQL is still a later pass, but the foundation now exists for it to start as a plugin:

- Presets can keep headless binaries REST-first until GraphQL is activated.
- Bootstrap providers are resolved before plugin activation.
- Plugin lifecycle and state handling already exist.
- Route, asset, capability, setting, and hook registration paths already exist.

Pass 4 should therefore add GraphQL as another plugin descriptor over the same application services already used by REST and admin.

## Remaining Follow-Up

Pass 3.6 is now complete enough to hand off to Pass 4. Remaining follow-up after Pass 3.6 includes:

- Stronger manifest/dependency validation beyond the initial required fields.
- Durable implementations for the declared non-SQLite bootstrap providers.
- Richer provider migration tooling around the already-enforced export/restart/import boundary.
