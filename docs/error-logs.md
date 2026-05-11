# Error logs and operational diagnostics

GoCMS maintains a **bounded local error log** for operator visibility without requiring an external logging stack. Records are stored through the diagnostics application service and surfaced on the admin **Runtime status** page.

Domain type: `internal/domain/diagnostics`. Service: `internal/application/diagnostics`. SQLite persistence: `internal/storage/sqlite/store.go` (migration `0003_error_logs`).

---

## Record model

`ErrorRecord` contains:

- **ID** — UUID string.
- **Source** — short subsystem label (for example `admin.login`).
- **Message** — human-readable error text.
- **Severity** — caller-defined string (`warning`, `error`, …).
- **OccurredAt** — UTC timestamp.
- **Details** — optional JSON-compatible map.

---

## Bounded retention

After each insert, SQLite deletes older rows beyond **200** retained entries (ordered by `occurred_at` descending). This keeps disk usage predictable for embedded deployments.

---

## Recording paths

`internal/delivery/admin/handler.go` defines `logError`, which calls `diagnostics.Record` when the service is enabled.

Typical sources today:

- **admin.login** — login failures, lockouts, or validation issues (severities `warning` / `error`).
- **admin.user.\*** — failures while rotating passwords, generating recovery codes, issuing reset tokens, creating/revoking app tokens.

Core REST handlers may be extended similarly — search for `diagnostics.Record` / `logError` when auditing coverage.

---

## Viewing entries

### Runtime status screen

`/go-admin/runtime` (requires **`settings.manage`**) lists recent error rows (limit **10** alongside health and audit snippets). Rows appear with prefix **`Error:`** plus source; description shows message; badge shows severity.

### Health check

Health id **`error_logs`** verifies diagnostics service is configured (`repo` and clock present). Failure means operators lose this diagnostic channel ([health-checks.md](./health-checks.md)).

---

## Operational guidance

1. Use error logs for **first-line triage** — authentication friction, admin action failures — not as a substitute for structured server logs in production.
2. Pair with **audit logs** for security-sensitive workflows ([audit-logs.md](./audit-logs.md)).
3. When reporting bugs, capture relevant rows from Runtime status plus surrounding audit entries.

---

## See also

- [health-checks.md](./health-checks.md) — error log store check.
- [migrations.md](./migrations.md) — `error_logs` table migration.
- [cms-capabilities.md](./cms-capabilities.md) — operational features overview.
