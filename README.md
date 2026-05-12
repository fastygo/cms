# GoCMS

GoCMS is a Go-native CMS platform aligned with the compatibility contract in `go-codex/en`.

The repository is now in an implemented operational state and includes both content management and hardening features, not only bootstrap scaffolding.

## Architecture and Documentation Layers

- `go-codex/en`: external compatibility contract (public behavior, APIs, runtime modes, capabilities).
- `go-stack/en`: internal Go architecture (modules, services, repositories, adapters, plugin runtime).
- `go-ui8kit/en`: admin UI profile and runtime assets.
- `.project/implementation/en`: implementation plan and progress notes.

## Current Feature Set (Pass 7+)

- Core CMS domains and services:
  - Content types, taxonomies, content, revisions, media, menus, settings, users/authors.
- REST control plane and admin shell:
  - Admin routes (`/go-admin`, `/go-login`, `/go-logout` depending on profile).
  - JSON API under `/go-json`.
- Authentication and authorization hardening:
  - Local password provider with Argon2id verification.
  - Session policy and secure cookie handling.
  - Login rate limiting and temporary lockout.
  - Recovery codes and admin reset tokens.
  - App/API tokens with scope, expiry, and revocation support.
  - Capability-based RBAC (`admin`, `editor`, `viewer`) with action/resource checks.
- Operational hardening:
  - Versioned schema migrations with `schema_migrations` ledger.
  - Health registry (`database`, `migrations`, `snapshot`, `authn`, `audit`, `error logs`).
  - Snapshot-based backup and restore (site backup workflows).
  - Audit logs for privileged admin/API actions with secret redaction.
  - Bounded local error log storage (diagnostics for admin visibility).
- Plugin and extension model:
  - Runtime plugins for routes, capabilities, and settings.
  - GraphQL plugin.
  - JSON import/export plugin for backup/restore UX.

## Storage and Profiles

GoCMS supports multiple storage profiles, including:

- `memory`
- `sqlite`
- `bbolt`
- `mysql`
- `postgres`
- `json-fixtures`
- `browser-indexeddb`

Runtime profiles include:

- `admin`
- `headless`
- `playground`
- `full`
- `conformance`

## Getting Started

Run locally:

```bash
go run ./cmd/server
```

Health checks:

```text
/healthz
/readyz
```

Admin + API entry points:

```text
/go-login
/go-admin
/go-json
```

Run tests:

```bash
go test ./...
```

Format source:

```bash
gofmt -w ./cmd ./internal
```

## Docker

Build and run with Compose (SQLite data under `./data`; see [Operations — Docker Compose](./docs/operations.md#docker-compose-local-and-ssh-hosts) for host permissions and session keys):

```bash
mkdir -p data
docker compose up -d --build cms
```

## Documentation Map

- [CMS capability overview](./docs/cms-capabilities.md)
- [Onboarding 101 (English/Russian guide in `docs/onboarding-101.md`)](./docs/onboarding-101.md)
- [Authentication](./docs/authentication.md)
- [Security model](./docs/security-model.md)
- [Operations](./docs/operations.md)
- [Migrations](./docs/migrations.md)
- [Backup and restore](./docs/backup-restore.md)
- [Health checks](./docs/health-checks.md)
- [Audit logs](./docs/audit-logs.md)
- [Error logs](./docs/error-logs.md)

## Project History

Current implementation progress is tracked in:

- `.project/progress.md`
- `.project/implementation/en`
