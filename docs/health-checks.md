# Health checks guide

GoCMS aggregates **provider-neutral health checks** through `internal/application/health` and exposes results on the admin **Runtime status** screen (`/go-admin/runtime`), which requires **`settings.manage`**.

Composition happens in `internal/infra/features/cms/module.go` when constructing `healthService`.

---

## Check registry

Each check has:

- **ID** — stable machine id.
- **Label** — short title shown in admin.
- **Description** — explanatory text when passing.
- **Status** — `ok` or `error`.
- **Error** — message when failing (also surfaced in the runtime list row description).

Implementation runs `Run(ctx)` per check; failures set status to `error` and populate `Error`.

### Built-in checks

| ID | Label | Behavior |
| --- | --- | --- |
| `database` | Database connectivity | Calls `bootstrap.Store.HealthCheck` (SQLite: `PingContext`). |
| `migrations` | Schema migrations | If store implements `MigrationStatus(context.Context) error`, runs it (SQLite: verifies all declared migrations applied — see [migrations.md](./migrations.md)). Otherwise skipped (nil error). |
| `snapshot` | Snapshot capability | Verifies bootstrap `ProviderCapabilities.SupportsSnapshots`; fails if profile disables snapshots. |
| `authn` | Authentication store | Ensures `authnService.Enabled()` — local authn wired with repository and hasher. |
| `audit` | Audit log store | Ensures `auditService.Enabled()` — audit repository available. |
| `error_logs` | Error log store | Ensures `diagnosticsService.Enabled()` — diagnostics repository available ([error-logs.md](./error-logs.md)). |

---

## Where results appear

Admin handler `runtimePage` (`internal/delivery/admin/handler.go`) appends one row per health result after static runtime facts (preset, profiles, plugins, policies).

Operators should treat failing checks as **deployment or startup issues**, not as user-facing CMS configuration toggles.

---

## Relationship to HTTP `/healthz` / `/readyz`

The framework host may expose generic process health endpoints (see root `README.md`). The checks above are **application-level CMS diagnostics** focused on storage, migrations, and operational subsystems — complementing any infrastructure probes.

---

## See also

- [migrations.md](./migrations.md) — what migration failures mean.
- [backup-restore.md](./backup-restore.md) — snapshot capability failures.
- [error-logs.md](./error-logs.md) — diagnostics store failures.
- [operations.md](./operations.md) — summary list.
