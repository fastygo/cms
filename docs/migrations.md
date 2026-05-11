# Schema migrations guide

GoCMS applies **versioned migrations** for durable SQLite-backed deployments. Migrations run during store initialization (`Store.Init` → `applyMigrations` in `internal/storage/sqlite`).

---

## What gets migrated

### Migration ledger

Table `schema_migrations` stores:

- `version` — unique migration id (string).
- `description` — human-readable summary.
- `applied_at` — UTC timestamp when the migration committed.

Each migration runs in a **single transaction**: all statements execute, then the ledger row is inserted. Failed migrations roll back and surface on next startup or via health checks.

### Current SQLite migration chain

Defined in `internal/storage/sqlite/migrations.go`:

| Version | Purpose |
| --- | --- |
| `0001_core_schema` | Core CMS tables: content, content types, taxonomy, media, users, settings, menus, revisions, preview access, indexes. |
| `0002_auth_audit` | Authentication artifacts (recovery codes, reset tokens, app tokens, login attempts) and `audit_events`. |
| `0003_error_logs` | Bounded **error log** table used by diagnostics ([error-logs.md](./error-logs.md)). |

Startup is **idempotent**: already-applied versions are skipped.

---

## Provider-specific behavior

### SQLite (implemented)

- Default file DSN normalization: empty or `fixture` becomes `file:gocms.db` unless overridden by bootstrap/preset `DataSource`.
- `MigrationStatus(ctx)` verifies every declared migration version exists in `schema_migrations`; missing versions return an error such as `pending migration 0003_error_logs`.

### Memory / JSON fixtures / browser-indexeddb (transitional)

These profiles still use the SQLite driver with **in-memory** file names for tests or playground-style runs. Migrations apply the same SQL chain so behavior stays aligned with durable SQLite.

### MySQL, PostgreSQL, Bolt (`bbolt`)

Declared in `internal/infra/bootstrap` but **not implemented** — bootstrap returns an error if selected. Future adapters should implement:

- The shared `bootstrap.Store` interface (including `Init`, `HealthCheck`, and migration reporting compatible with health checks).
- The same operational boundaries listed in [operations.md](./operations.md).

---

## Operational checklist

1. **Deploy new binary** that contains additional migration versions.
2. **Restart** the process so `Init` runs migrations before serving traffic.
3. Confirm **health**: admin **Runtime status** (`/go-admin/runtime`) includes check `migrations` — failures show pending migration errors ([health-checks.md](./health-checks.md)).
4. For **multi-instance** deployments, ensure only one instance runs migrations at a time or use external orchestration; SQLite deployments are typically single-writer.

---

## Changing durable providers

Bootstrap storage profile is **not** an admin toggle. Moving between durable backends requires export → configuration → restart → import. See [backup-restore.md](./backup-restore.md) and runtime messaging on `/go-admin/runtime`.

---

## See also

- [cms-capabilities.md](./cms-capabilities.md) — runtime and storage overview.
- [health-checks.md](./health-checks.md) — migration status in health registry.
- [operations.md](./operations.md) — durable provider notes.
