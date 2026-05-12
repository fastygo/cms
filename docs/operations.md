# Operations

Focused guides:

- [CMS capabilities overview](./cms-capabilities.md)
- [Schema migrations](./migrations.md)
- [Backup and restore](./backup-restore.md)
- [Health checks](./health-checks.md)
- [Audit logs](./audit-logs.md)
- [Error logs](./error-logs.md)

---

## Schema Migrations

SQLite initialization runs versioned migrations and records them in `schema_migrations`.

Current baseline:

- `0001_core_schema`
- `0002_auth_audit`
- `0003_error_logs`

Repeated startup is idempotent. Health diagnostics surface pending migration failures through the runtime status screen. Details: [migrations.md](./migrations.md).

---

## Health Registry

The runtime status screen combines static runtime facts with live health checks for:

- database connectivity
- schema migration status
- snapshot capability
- local authn store availability
- audit store availability
- bounded local error-log availability

These checks are intended to stay provider-neutral even though SQLite is the first durable implementation. Details: [health-checks.md](./health-checks.md).

---

## Backup And Restore

GoCMS uses snapshot export/import as the autonomous backup workflow.

The exported bundle includes backup metadata:

- generation timestamp
- provider profile
- per-entity item counts
- validation result

Admin operators use the JSON import/export plugin actions for backup and restore. Details: [backup-restore.md](./backup-restore.md).

---

## Audit Visibility

Privileged admin and API mutations emit audit events. The runtime status screen shows the most recent audit rows so operators can confirm sensitive actions without leaving core.

Examples include:

- login and logout
- content create/update/trash/bulk actions
- settings writes
- user writes
- app token creation and revocation
- recovery code generation and reset-token issuance

Details: [audit-logs.md](./audit-logs.md).

---

## Error Logs

Core keeps a bounded local error log. The runtime status screen shows recent entries so operators can inspect local auth/admin failures without requiring an external logging vendor. Details: [error-logs.md](./error-logs.md).

---

## Local Recovery Runbook

If local admin access is lost:

1. Sign in with a stored recovery code, then rotate the password immediately.
2. If recovery codes are unavailable, have another administrator issue a reset token from the `Users` screen.
3. Copy the one-time token through a trusted channel.
4. Reset the password and sign in again.
5. Rotate or revoke any old app tokens if compromise is suspected.

See also [authentication.md](./authentication.md).

---

## Durable Provider Notes

SQLite is the first full provider for this pass. Future MySQL/Postgres/Bolt adapters should implement the same operational boundaries:

- versioned migrations
- health reporting
- backup/export compatibility
- authn state persistence
- audit persistence

Details: [migrations.md](./migrations.md), [cms-capabilities.md](./cms-capabilities.md).

---

## Static assets and CI checks

JavaScript tooling uses **Bun** (`bun.lock`, `packageManager` in `package.json`). To match CI locally:

```bash
bun install --frozen-lockfile
bun run verify
```

The same pipeline runs in GitHub Actions and inside the multi-stage **Docker** build (Tailwind, esbuild, UI8Kit `sync-assets`, **`append-gocms-locale-sync`** so public locale SPA updates header/footer menus, pinned **ui8px** lint, then `go test` when you run `make verify`).

---

## Docker Compose (local and SSH hosts)

The default [`docker-compose.yml`](../docker-compose.yml) binds SQLite data to **`./data`** on the host (`GOCMS_DATA_DIR` overrides the host path). The runtime image uses the distroless **non-root** user (**UID/GID 65532**).

`make deploy` creates the host data directory, runs a one-shot **BusyBox** container to fix bind-mount permissions, then starts `cms`. The helper is intentionally kept out of `docker-compose.yml` so IaC scanners only evaluate the long-running application service. It mounts the same data path and applies:

```bash
chown -R ${GOCMS_DATA_UID:-65532}:${GOCMS_DATA_GID:-65532} /data
chmod u+rwX /data
```

This prevents SQLite startup failures such as **`unable to open database file (14)`** when `./data` was created as `root:root`. Override `GOCMS_DATA_UID` / `GOCMS_DATA_GID` only if you build a runtime image with a different container user.

Set a strong **`APP_SESSION_KEY`** in production; do not rely on the development default.

Docker **`HEALTHCHECK`** and Compose **`healthcheck`** call the **`/healthcheck`** helper, which requests **`/readyz`** inside the container (override with **`HEALTHCHECK_URL`** if needed).
