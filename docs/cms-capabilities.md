# GoCMS capabilities overview

This document summarizes what the GoCMS core binary provides today: surfaces, domain features, configuration, and where to read deeper guides.

External contracts (authoritative for compatibility wording):

- `go-codex/en` — observable REST URLs, admin behavior, hooks, capabilities.
- `go-stack/en` — Go-native layering (domain, application, storage, delivery).
- `go-ui8kit/en` — UI8Kit admin presentation profile (when applicable).

Implementation roadmap and pass status: `.project/progress.md` and `.project/implementation/en/`.

---

## 1. Documentation map

| Topic | Guide |
| --- | --- |
| Capability overview | This document |
| Schema migrations | [migrations.md](./migrations.md) |
| Backup and restore | [backup-restore.md](./backup-restore.md) |
| Health checks | [health-checks.md](./health-checks.md) |
| Audit logs | [audit-logs.md](./audit-logs.md) |
| Error logs and diagnostics | [error-logs.md](./error-logs.md) |
| Authentication flows | [authentication.md](./authentication.md) |
| Security properties | [security-model.md](./security-model.md) |
| Operations summary | [operations.md](./operations.md) |

---

## 2. Runtime assembly

### Presets

Deployment selects a **preset** (for example via `GOCMS_PRESET`) resolved by `internal/platform/preset`. Presets combine:

- Runtime profile (`headless`, `admin`, `playground`, `full`, `conformance`).
- Storage profile (`sqlite`, `memory`, `json-fixtures`, `browser-indexeddb`, and declared-but-not-implemented `mysql`, `postgres`, `bbolt`).
- Deployment profile (`local`, `browser`, `serverless`, `container`, `ssh`).
- Active compiled plugins, optional site package directory, login/admin policies, dev bearer toggles.

Named presets include `full`, `headless`, `playground`, `offline-json-sql`, `ssh-fixtures`. Defaults and overrides are documented in `.project/implementation/en/03-6-configurator-plugin-foundation.md`.

### Bootstrap providers

`internal/infra/bootstrap` opens the **storage implementation** before application services and plugin activation. Storage is not a casual runtime toggle: switching durable providers is an export → configuration change → restart/redeploy → import operation (see [migrations.md](./migrations.md) and [backup-restore.md](./backup-restore.md)).

---

## 3. Core domain features

### Content

- **Kinds**: built-in `post` and `page`; custom kinds via content types.
- **Statuses**: `draft`, `scheduled`, `published`, `archived`, `trashed`.
- **Fields**: localized title, slug, body, excerpt; author; featured media; template; **metadata** (with typed registry / Meta API in Pass 7); taxonomy term references; timestamps; visibility.

Application logic: `internal/application/content`.

### Content types

Registered types define labels, permalink hints, supports flags (editor, taxonomies, revisions, etc.), REST/GraphQL visibility. Built-ins are installed at startup (`internal/application/contenttype`).

### Taxonomies

Definitions (mode flat/hierarchical, assignment to kinds) and terms with localized name/slug/description and optional hierarchy. Built-in types include `category` and `tag`; custom types are supported.

### Media

Core emphasizes **metadata**: filename, MIME (validated allowlist), dimensions, public URL, optional provider blob reference, variants, alt/caption. Binary upload pipelines beyond registering metadata are explicitly plugin/provider follow-ups (see `.project/progress.md` Pass 7).

### Users and authors

Users carry login, email, roles, password hash, session-related flags, and a **public author profile** projection for the REST author endpoint. Capabilities derive from roles (`admin`, `editor`, `viewer`) and optional explicit scopes on app tokens.

### Settings

Typed **definitions** (groups, validation, public/private, capability requirements, autoload policy) merged with stored values. Core keys cover site title, public rendering, permalinks, theme activation, operational auth tuning, admin screen preferences; plugins and themes register additional keys.

### Menus

Menus have id, name, **location**, and nested items (label, URL, kind, target, children).

### Revisions and preview

Revision storage and preview tokens exist in the domain model; dedicated HTTP preview URLs may be extended in later passes (roadmap notes in `.project/progress.md`).

---

## 4. Authorization

**Capabilities** are granular strings (for example `content.edit`, `settings.manage`, `users.manage`). A **principal** holds a set of capabilities; **roles** bundle defaults (`BuiltInRoles` in `internal/domain/authz`).

Every admin route and REST mutation checks capabilities through application services — the UI hiding a button is not sufficient.

---

## 5. Authentication (summary)

- **Local** password auth with Argon2id hashes, recovery codes, admin-issued reset tokens, app passwords / API tokens, login rate limiting, signed cookie sessions with idle and absolute TTL metadata.
- **Playground** demo login only in isolated profiles (see [authentication.md](./authentication.md)).
- **REST** accepts browser session cookie, `Authorization: Bearer` with hashed app tokens (resolved via authn service), optional **dev bearer** principals when explicitly enabled (never for production-like deployment presets).

Details: [authentication.md](./authentication.md), [security-model.md](./security-model.md).

---

## 6. REST API (`/go-json`)

Discovery:

- `GET /go-json` — root discovery (links namespace, authentication modes).
- `GET /go-json/go/v2/` — resource index.

Representative routes (see `internal/delivery/rest/handler.go`):

- Posts and pages: list, create, get by id, patch, delete (trash), get by slug.
- Content types: list, register.
- Taxonomies: list definitions, register definition, list/get terms, create term, assign terms to content.
- Media: list, get, save metadata, attach featured media to content.
- Menus: list, get by location.
- Settings: **public** settings only on `GET /go-json/go/v2/settings`.
- Authors: public profile by id.
- Search: published content search.

Authenticated callers receive capability-aware projections; public callers only see published, non-private data.

---

## 7. Admin UI (`/go-admin`)

Built on `github.com/fastygo/panel` descriptors in `internal/platform/cmspanel`:

| Area | Path (typical) | Capability theme |
| --- | --- | --- |
| Dashboard | `/go-admin` | Core dashboard |
| Posts | `/go-admin/posts` | Content private read / edit |
| Pages | `/go-admin/pages` | Same |
| Content types | `/go-admin/content-types` | Settings |
| Taxonomies / terms | `/go-admin/taxonomies`, `/go-admin/taxonomies/{type}/terms` | Taxonomies |
| Media | `/go-admin/media` | Media upload/edit |
| Menus | `/go-admin/menus` | Menus |
| Users | `/go-admin/users` | Users |
| Authors | `/go-admin/authors` | Read private |
| Roles / capabilities | `/go-admin/capabilities` | Roles |
| Settings | `/go-admin/settings` | Settings |
| Themes | `/go-admin/themes` | Themes |
| Permalinks | `/go-admin/permalinks` | Settings |
| API / headless | `/go-admin/headless` | Settings |
| **Runtime status** | `/go-admin/runtime` | Settings (operators) |

Login and logout: `GET/POST /go-login`, `POST /go-logout`.

---

## 8. Public site rendering

When the runtime profile includes public rendering (`full`, `playground`, etc.), `internal/delivery/publicsite` serves `/` with theme resolution, permalink routing (home, blog, posts, pages, taxonomies, author, search), and optional **preview query** overrides for themes. Public rendering can be disabled via `public.rendering` setting.

---

## 9. GraphQL plugin

Compiled plugin exposes `GET/POST /go-graphql` (plus options) when activated. Resolvers use the same application services as REST. Settings cover introspection, depth/length limits, CORS, cache headers. See `.project/implementation/en/04-graphql-plugin.md`.

---

## 10. Plugins (compiled)

Plugins are **compiled into the binary** and activated via preset/plugin set:

- **graphql** — schema over core services.
- **json-import-export** — JSON snapshot download/upload and optional site package paths ([backup-restore.md](./backup-restore.md)).
- **playground** — browser-local sandbox UX aligned with Pass 3.5 boundary.

Registry supports manifests, capabilities, settings, hooks (actions/filters), routes per surface (admin, REST, public), assets, editor providers, admin menu and screen actions (`internal/platform/plugins`).

---

## 11. Operational features

| Feature | Description |
| --- | --- |
| Migrations | Versioned SQL migrations with `schema_migrations` (SQLite today). |
| Snapshots | Export/import core entities as JSON (`internal/application/snapshot`). |
| Health | Aggregated checks in admin runtime view ([health-checks.md](./health-checks.md)). |
| Audit | Append-only style events for sensitive actions ([audit-logs.md](./audit-logs.md)). |
| Error logs | Bounded local store for operator diagnostics ([error-logs.md](./error-logs.md)). |

---

## 12. Related README

The repository root `README.md` may still describe early Pass 0 scaffolding; for current behavior rely on this folder, `.project/progress.md`, and implementation plans under `.project/implementation/en/`.
