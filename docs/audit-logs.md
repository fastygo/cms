# Audit logs guide

GoCMS records **audit events** for privileged REST and admin actions. Events are persisted when an audit-capable store is configured; the audit application service redacts sensitive detail keys before storage.

Domain model: `internal/domain/audit`. Application service: `internal/application/audit`.

---

## Event shape

Each event includes:

- **ID** — unique identifier (generated if empty).
- **ActorID** — authenticated principal id when available (failed login attempts may use submitted identifier).
- **Action** — namespaced string such as `api.content.update` or `admin.settings.save`.
- **Resource** — coarse resource type (`content`, `user`, `session`, …).
- **ResourceID** — optional target id.
- **Status** — `success` or `failure`.
- **OccurredAt** — UTC timestamp.
- **RemoteAddr**, **UserAgent** — request metadata when applicable.
- **Details** — structured map, **redacted** before persistence.

### Redaction rules

Detail keys are scanned case-insensitively. If a key contains substrings `password`, `token`, `secret`, `code`, or `hash`, the stored value becomes `[redacted]`.

---

## Where events are recorded

### REST API (`internal/delivery/rest/handler.go`)

Examples (success paths):

- `api.content.create`, `api.content.update`, `api.content.trash`
- `api.content_type.register`
- `api.taxonomy.register`
- `api.term.create`
- `api.terms.assign`
- `api.media.save`

REST list/read/search endpoints generally do **not** emit audit rows — focus is on mutating operations.

### Admin (`internal/delivery/admin/handler.go`)

Includes:

- **Auth**: `auth.login` (success/failure), `auth.logout`
- **Content**: `admin.content.create`, `admin.content.update`, `admin.content.trash`, `admin.content.quick_edit`, `admin.content.bulk`
- **Users / security**: `auth.password.set`, `auth.recovery_codes.generate`, `auth.reset_token.issue`, `auth.app_token.create`, `auth.app_token.revoke`, `admin.user.save`
- **Settings**: `admin.settings.save` (subset of keys listed in handler)

Additional direct `Audit.Record` calls may exist for edge cases — search the codebase for `Audit.Record` / `h.audit`.

---

## Viewing audit data

### Admin Runtime status

`/go-admin/runtime` loads **recent audit events** (limit 10 in current handler) interleaved with health rows — rows prefixed with **`Audit:`** and action name.

### Storage

SQLite persists JSON payloads in `audit_events` (see migration `0002_auth_audit`). Listing is newest-first with a configurable limit (`ListAuditEvents`).

---

## Operational notes

1. Audit logs are **not a full HTTP access log** — only instrumented actions appear.
2. For compliance archives, pair SQLite files or snapshots with external log shipping if required by policy.
3. Failure audits (for example login failure) help detect brute-force attempts alongside login rate limiting ([authentication.md](./authentication.md)).

---

## See also

- [security-model.md](./security-model.md) — redaction philosophy.
- [health-checks.md](./health-checks.md) — audit store health check.
- [cms-capabilities.md](./cms-capabilities.md) — authorization overview.
