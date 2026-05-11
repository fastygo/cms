# 09. Go-native Developer And Plugin APIs

This document describes the practical extension surface for GoCMS when plugins are compiled into the binary.

It is not a PHP compatibility layer. The goal is to preserve the observable compatibility contract while exposing extension points in a Go-native form.

## Core idea

GoCMS plugins are descriptors plus registrations.

A compiled plugin should:

1. Expose a validated `Manifest`.
2. Register capabilities, settings, routes, assets, hooks, and editor providers through `internal/platform/plugins.Registry`.
3. Receive narrow application services through constructor injection.
4. Avoid global mutation at import time.

## Mapping familiar concepts

WordPress-style concept to GoCMS equivalent:

- Plugin header to `plugins.Manifest`
- Activation state to `plugins.Runtime` plus `plugins.StateRepository`
- `add_action()` to `Registry.AddActionHandlers(...)`
- `add_filter()` to `Registry.AddFilterHandlers(...)`
- Admin menu registration to `Registry.AddAdminMenu(...)`
- Admin screen actions to `Registry.AddScreenActions(...)`
- REST route registration to `Registry.AddRoutes(...)` with `SurfaceREST`
- Public route registration to `Registry.AddRoutes(...)` with `SurfacePublic`
- Admin route registration to `Registry.AddRoutes(...)` with `SurfaceAdmin`
- Capability declaration to `Registry.AddCapabilities(...)`
- Setting schema declaration to `Registry.AddSettings(...)`
- Script/style enqueue metadata to `Registry.AddAssets(...)`
- Editor integration to `Registry.AddEditorProviders(...)`

## Descriptor shape

A compiled plugin implements:

```go
type Descriptor interface {
    Manifest() Manifest
    Register(context.Context, *Registry) error
}
```

The manifest is stable metadata. The `Register` method is the executable part.

## Hooks and filters

GoCMS keeps hook behavior deterministic:

- Handlers are ordered by ascending priority.
- Same-priority handlers run in registration order.
- Action handlers can fail fast or collect failures.
- Filter handlers transform typed values sequentially.

Use hook metadata for discoverability and executable handlers for behavior.

Typical pattern:

```go
registry.AddHooks(plugins.HookRegistration{
    HookID:    "render.content.filter",
    HandlerID: "my-plugin.render",
    OwnerID:   "my-plugin",
    Category:  plugins.HookCategoryFilter,
    Priority:  20,
})

registry.AddFilterHandlers(plugins.FilterHandlerRegistration{
    Hook: plugins.HookRegistration{
        HookID:    "render.content.filter",
        HandlerID: "my-plugin.render",
        OwnerID:   "my-plugin",
        Category:  plugins.HookCategoryFilter,
        Priority:  20,
    },
    Handle: func(ctx context.Context, hookCtx plugins.HookContext, value any) (any, error) {
        html := value.(string)
        return `<div class="my-plugin-marker"></div>` + html, nil
    },
})
```

## Safe output boundaries

Pass 7 intentionally exposes filters only at boundaries that do not bypass authorization:

- `rest.content.filter` receives already projected REST content DTOs.
- `graphql.content.filter` receives already projected GraphQL content values.
- `render.content.filter` receives the final public content HTML/string payload.

These hooks are for presentation and projection shaping, not for loading private data.

## Routes and surfaces

Every plugin route declares its surface explicitly:

- `SurfaceAdmin`
- `SurfaceREST`
- `SurfacePublic`

This keeps activation, exposure, and conformance testing explicit.

## Capabilities and settings

Plugins should declare new capabilities and settings in the manifest, then register them through the registry.

Guidelines:

- Capability IDs should be namespaced by plugin.
- Settings should declare type, default, visibility, and managing capability.
- Public setting visibility must remain explicit.

## Editor providers

Editor integrations should be registered as providers, not by patching admin pages ad hoc.

Use `Registry.AddEditorProviders(...)` when adding:

- A custom editor implementation
- A higher-priority editor override
- A specialized editor for one content workflow

## Assets

Plugin assets remain declarative metadata.

Register them with `Registry.AddAssets(...)` and bind them to a surface. Theme assets remain a theme concern and are resolved through the theme registry and active preset.

## Lifecycle expectations

Compiled plugins still participate in lifecycle behavior:

- Manifest validation happens before runtime activation.
- Activation failure marks the plugin failed.
- Inactive plugins do not contribute routes, hooks, or assets to the active registry.
- Lifecycle hooks can observe activation and deactivation events.

## Constructor guidance

Prefer constructors like:

```go
func New(content appcontent.Service, settings appsettings.Service) Plugin
```

Avoid handing plugins a raw storage implementation or a full service locator unless the plugin is part of core infrastructure.

## What this is not

This pass does not provide:

- Runtime installation of arbitrary code
- PHP-style global function shims
- Marketplace packaging
- Unrestricted plugin access to core internals

The extension surface is intentionally compiled, explicit, typed, and testable.
