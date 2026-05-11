# Backup and restore guide

GoCMS uses a **snapshot service** to export and import the core CMS dataset as a single JSON bundle. This is the supported autonomous backup path for operators who cannot rely on external backup SaaS.

Implementation: `internal/application/snapshot`.

---

## Bundle format

- **Version string**: `gocms.snapshot.v1` (`SnapshotVersion` constant).
- **Top-level fields**: exported timestamp, `backup` metadata, arrays for content, content types, taxonomy definitions, taxonomy terms, media, users, settings, menus.

### Backup metadata (`BackupMetadata`)

Included in each export:

- `generated_at`
- `provider_profile` — storage profile string passed via `WithProviderProfile` (from bootstrap).
- `item_counts` — counts per entity group.
- `validation_result` — export validation summary (currently `ok` when export completes).

Import order inside `Import` applies dependencies safely: content types and taxonomy definitions before terms; users and settings before content lists; menus and content entries last within the bundle logic.

---

## Admin workflows (`json-import-export` plugin)

Plugin id: `json-import-export`. Registered from `internal/plugins/jsonimportexport/plugin.go`.

Requires capability **`settings.manage`** for all snapshot routes.

### Export JSON to workstation

- **GET** `/go-admin/plugins/json-import-export/export`
- Response: `application/json` attachment `gocms-content-snapshot.json`.

### Import JSON from workstation

- **POST** `/go-admin/plugins/json-import-export/import`
- Multipart form field **`snapshot`** — file upload (max 8 MiB parsed by handler).
- Success: HTTP 204 No Content.

### Site package (directory on server)

When `jsondir` site package provider is configured (`SitePackageDir` / plugin setting `json-import-export.site_package_dir`):

- **POST** `/go-admin/plugins/json-import-export/export-site-package` — writes current snapshot via site package provider.
- **POST** `/go-admin/plugins/json-import-export/import-site-package` — loads bundle from site package then imports.

Admin UI actions also appear on **Settings** and **Headless** screens (screen ids `settings`, `headless`) when the plugin registers screen actions.

---

## Operational guidance

1. **Treat exports as sensitive**: they include user records, password hashes, recovery metadata, and settings — protect files like database dumps.
2. **Restore strategy**: stop writes if possible → import snapshot → verify health and audit trail → rotate secrets if the backup might be stale or exposed.
3. **Cross-provider migration**: export from old profile → deploy new binary/config → import into new store ([migrations.md](./migrations.md)).
4. **Media binaries**: snapshot carries **media metadata** and URLs; large binaries may live outside the JSON bundle depending on provider — verify asset URLs and blob stores separately after restore.

---

## Health integration

The core registers a **snapshot capability** health check: if the active storage profile disables snapshots in bootstrap capabilities, the check fails ([health-checks.md](./health-checks.md)).

---

## See also

- [cms-capabilities.md](./cms-capabilities.md) — plugins and presets.
- [migrations.md](./migrations.md) — schema evolution alongside restores.
- [audit-logs.md](./audit-logs.md) — auditing administrative restores if extended in your deployment.
