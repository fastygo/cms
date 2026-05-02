# GoCMS

GoCMS is a Go-native CMS implementation targeting the compatibility contract in `go-codex/en`.

This repository currently implements **Pass 0: Project Skeleton** only. It provides a runnable process, framework host wiring, health endpoints, structured logging setup, package boundaries, and a smoke test harness.

## Documentation Layers

- `go-codex/en`: external compatibility contract.
- `go-stack/en`: Go-native architecture profile.
- `go-ui8kit/en`: UI8Kit admin profile.
- `.project/implementation/en`: internal implementation roadmap and guardrails.

## What Exists In Pass 0

- Go module: `github.com/fastygo/cms`.
- Local framework replace: `github.com/fastygo/framework => ../@Framework`.
- Composition root: `cmd/server`.
- Runtime config wrapper: `internal/platform/config`.
- Structured logging helper: `internal/platform/logging`.
- Minimal system feature: `internal/infra/features/system`.
- Reserved package boundaries for future domain, application, storage, delivery, and runtime layers.

## What Does Not Exist Yet

- CMS content model.
- REST compatibility resources.
- Admin GUI.
- Storage layer.
- GraphQL plugin.
- Theme rendering.
- Media handling.
- Plugin runtime.

## Run

```bash
go run ./cmd/server
```

Default health endpoints:

```text
/healthz
/readyz
```

Pass-0 system route:

```text
/
```

## Test

```bash
go test ./...
```

## Format

```bash
gofmt -w ./cmd ./internal
```

## Next Pass

The next implementation pass is the Core Compatibility Kernel described in `.project/implementation/en/01-core-kernel.md`.
