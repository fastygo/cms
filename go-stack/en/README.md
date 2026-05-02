# GoCMS Go Stack Profile

This directory defines the Go-native architecture profile for implementing the external compatibility contract in `../../go-codex/en`.

The profile explains how a Go implementation should organize domain models, application services, storage boundaries, delivery adapters, plugin runtime choices, theme rendering, background work, and conformance adapters. It is not a public compatibility standard by itself. The public standard remains `go-codex/en`.

## Scope

This profile covers:

- Go package and boundary guidance.
- Domain models and typed constants.
- Application service boundaries.
- Repository and transaction contracts.
- REST and GraphQL adapter guidance.
- Plugin runtime strategies.
- Theme rendering interfaces.
- Background jobs.
- Testing and conformance adapters.

## Non-Goals

This profile does not require:

- A concrete web framework.
- A concrete user interface toolkit.
- A concrete template engine.
- A concrete database.
- A concrete ORM.
- A concrete GraphQL library.
- A concrete queue or scheduler library.
- A concrete object storage provider.

## Document Map

- `00-go-architecture-profile.md` defines the base Go architecture rules.
- `01-domain-model.md` defines recommended domain package boundaries.
- `02-application-services.md` defines service and command/query boundaries.
- `03-storage-repositories.md` defines repository, transaction, migration, and metadata boundaries.
- `04-rest-graphql-adapters.md` defines REST and GraphQL as adapters over the same services.
- `05-plugin-runtime-strategies.md` defines plugin runtime choices and invariants.
- `06-theme-rendering-interface.md` defines renderer contracts and headless/public rendering modes.
- `07-background-jobs.md` defines background work, scheduling, idempotency, and shutdown.
- `08-testing-and-conformance-adapters.md` defines how Go implementations expose conformance fixtures.

## Relationship To Compatibility

`go-codex/en` defines observable behavior. This profile defines one recommended Go architecture for achieving that behavior.

If this profile and `go-codex/en` conflict, `go-codex/en` wins for compatibility. This profile may be revised without changing the external contract as long as observable behavior remains stable.
