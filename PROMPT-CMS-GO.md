# Prompt: Complete WordPress-Style CMS Architecture On Go

Use this document as the master architectural prompt for designing and building a complete CMS on Go. The target product is a finished WordPress-style CMS: it must include an admin GUI, public site rendering, public API, content management, media, users, roles, localization, SEO, themes, menus, settings, extension points, tests, and operational readiness.

The architecture must be Go-first, but it must not depend on any specific web framework, ORM, router, admin package, JavaScript framework, database driver, storage SDK, queue library, or UI kit. Choose implementation tools later. This document defines the required product surface, boundaries, invariants, and workflows.

## Mission

Build a maintainable CMS that can replace a heavy WordPress-style setup while keeping the same product completeness:

- A browser-based admin panel for non-technical users.
- A public website rendering layer.
- A stable public API for external frontends and integrations.
- A content model that supports pages, posts, taxonomies, menus, blocks, media, settings, users, roles, and localization.
- A safe operational model with migrations, tests, backups, audit logs, and performance controls.

The CMS should be small in architecture, not small in capability. Keep the core explicit, modular, and understandable.

## Core Principles

- Keep business rules outside HTTP handlers.
- Keep persistence details outside domain services.
- Keep admin UI actions thin; they should call application services.
- Keep public API responses explicit and versioned.
- Keep public rendering separate from admin management.
- Keep user-generated content sanitized at input and escaped at output.
- Prefer small modules with clear ownership over generic abstractions.
- Prefer stable domain contracts over framework-specific magic.
- Treat localization, permissions, media URLs, SEO metadata, publishing rules, and cache invalidation as first-class requirements.
- Do not expose raw database models directly through API responses.
- Do not hide critical business behavior inside template files.

## High-Level Application Surfaces

### 1. Admin GUI

The admin panel is the main editorial interface. It must be available under a configurable path such as `/admin`.

Required admin areas:

- Dashboard.
- Pages.
- Posts or Articles.
- Taxonomies.
- Media Library.
- Menus.
- Comments or Feedback moderation, if comments are enabled.
- Forms or submissions, if forms are part of the CMS.
- Users.
- Roles and permissions.
- Site settings.
- Localization settings.
- SEO settings.
- Theme and template settings.
- Extensions or modules.
- System health.
- Audit log.

The admin panel must support:

- Login and logout.
- Password reset or secure recovery flow.
- Role-based access control.
- List, create, edit, delete, restore, publish, unpublish, and duplicate actions where appropriate.
- Search, filtering, sorting, pagination, and bulk actions.
- Form validation with user-friendly messages.
- Draft previews.
- Autosave or manual save states for long content forms.
- Clear visual state for draft, scheduled, published, archived, and trashed content.
- Upload progress and media previews.
- Safe confirmation flows for destructive actions.

### 2. Public Website

The CMS must be able to render a normal public website without requiring a separate frontend application.

Required public rendering capabilities:

- Home page.
- Page detail routes.
- Post detail routes.
- Archive routes.
- Taxonomy archive routes.
- Search results route.
- Sitemap.
- RSS or feed output if publishing articles.
- Error pages for 404 and 500.
- Preview routes for unpublished content with signed or permission-checked access.

The public site must support themes and templates:

- A default theme.
- Layout templates.
- Page templates.
- Post templates.
- Archive templates.
- Taxonomy templates.
- Search template.
- Error templates.
- Reusable partials such as header, footer, navigation, breadcrumbs, cards, pagination, and metadata.

### 3. Public API

The CMS must expose a versioned public API under a prefix such as `/api/v1`.

Minimum public endpoints:

- `GET /api/v1/pages`
- `GET /api/v1/pages/{slug}`
- `GET /api/v1/posts`
- `GET /api/v1/posts/{slug}`
- `GET /api/v1/taxonomies`
- `GET /api/v1/taxonomies/{type}/{slug}`
- `GET /api/v1/menus/{location}`
- `GET /api/v1/settings`
- `GET /api/v1/media/{id}` or public media metadata endpoints if needed.
- `GET /api/v1/search`

Optional authenticated API endpoints:

- Content management API for trusted clients.
- Media upload API.
- User profile API.
- Form submission API.
- Webhook management API.

API requirements:

- Version every public contract.
- Validate and normalize every input.
- Return explicit DTO-style responses, not raw persistence records.
- Support locale negotiation.
- Support pagination metadata.
- Support filtering by status only for authenticated users.
- Return stable error shapes.
- Include cache headers where safe.
- Avoid N+1 query behavior.

## Suggested Go Module Boundaries

The final folder names can differ, but the architecture should preserve these responsibilities.

### Domain Layer

Pure business entities and invariants:

- Page.
- Post.
- Taxonomy.
- MediaAsset.
- User.
- Role.
- Permission.
- Setting.
- Menu.
- Comment.
- Block.
- Theme.
- Template.
- Revision.
- AuditEvent.

Domain objects should not know about HTTP, SQL, JSON serialization, templates, sessions, or background workers.

### Application Services

Use services for workflows and rules:

- PageService.
- PostService.
- TaxonomyService.
- MediaService.
- UserService.
- AuthService.
- RoleService.
- SettingService.
- MenuService.
- CommentService.
- SearchService.
- ThemeService.
- RenderService.
- RevisionService.
- AuditService.
- ImportExportService.

Services should coordinate validation, repositories, transactions, events, cache invalidation, and audit logging.

### HTTP Layer

HTTP handlers should:

- Parse request input.
- Authenticate user when required.
- Authorize action.
- Call application services.
- Map service results into response DTOs, redirects, or rendered templates.
- Never contain core business logic.

Separate handler groups:

- Admin handlers.
- Public website handlers.
- Public API handlers.
- Auth handlers.
- Webhook handlers.
- Health and diagnostics handlers.

### Persistence Layer

Persistence should be hidden behind repository interfaces or equivalent boundaries.

Responsibilities:

- Content queries.
- User and role queries.
- Media metadata queries.
- Settings queries.
- Revision history.
- Audit events.
- Migrations.
- Transactions.

Persistence rules:

- Keep migrations deterministic and reversible where possible.
- Keep created, updated, deleted, and published timestamps consistent.
- Use soft deletion for user-facing content where recovery matters.
- Enforce uniqueness where required, especially slugs per content type and locale.
- Store localized fields in a predictable structure.
- Avoid storing computed presentation output as the source of truth.

### Rendering Layer

The rendering layer turns domain content into public HTML.

Responsibilities:

- Resolve incoming URL to a route target.
- Select the active theme.
- Select the correct template.
- Load content, menus, settings, and related media.
- Render blocks and rich content safely.
- Escape output by default.
- Apply SEO metadata.
- Apply cache rules.

Rendering must not mutate content.

### Admin UI Layer

The admin UI layer renders forms, tables, dashboards, and actions.

Responsibilities:

- List screens.
- Edit forms.
- Media browser.
- Block editor or structured content editor.
- Menu builder.
- Settings screens.
- Role and permission screens.
- Audit and system screens.

Admin UI must remain a client of application services. It must not become a second business layer.

### Background Jobs

Use background jobs for work that should not block requests:

- Image resizing and conversion.
- Search indexing.
- Sitemap generation.
- Feed generation.
- Email sending.
- Import and export.
- Webhook delivery.
- Cache warming.
- Cleanup of orphaned media and expired previews.

The architecture must work even if the queue implementation changes later.

## Data Model Requirements

### Pages

Pages are hierarchical evergreen content.

Required fields:

- ID.
- Localized title.
- Localized slug.
- Localized content.
- Localized SEO title.
- Localized SEO description.
- Status.
- Parent ID.
- Template key.
- Featured image ID.
- Author ID.
- Published timestamp.
- Created timestamp.
- Updated timestamp.
- Deleted timestamp.

Required behavior:

- Draft, published, scheduled, archived, and trashed states.
- Parent-child hierarchy.
- Slug generation per locale.
- Slug uniqueness per locale and parent scope.
- Public lookup by localized slug.
- Optional custom template.
- Featured image.
- Taxonomy assignment where configured.
- Revisions.
- Preview for unpublished content.

### Posts

Posts are chronological content similar to WordPress posts.

Required behavior:

- Same localization and SEO fields as pages.
- Draft, published, scheduled, archived, and trashed states.
- Author.
- Excerpt.
- Featured image.
- Categories and tags.
- Archive listing.
- Feed inclusion.
- Comment support if comments are enabled.
- Revisions.

### Taxonomies

Taxonomies classify content.

Required taxonomy types:

- Category.
- Tag.

Architecture must allow custom taxonomy types later.

Required fields:

- ID.
- Type.
- Localized name.
- Localized slug.
- Localized description.
- Parent ID for hierarchical taxonomies.
- Created timestamp.
- Updated timestamp.

Required behavior:

- Assign taxonomies to pages and posts.
- Support hierarchical categories.
- Support non-hierarchical tags.
- Lookup by type and localized slug.
- Public archive pages.
- API serialization.

### Media

Media is a core system, not an afterthought.

Required fields:

- ID.
- Original filename.
- Stored filename.
- Disk or storage location.
- Public URL or resolver key.
- MIME type.
- Size.
- Width and height for images.
- Alt text.
- Caption.
- Credit.
- Localized metadata where useful.
- Owner ID.
- Created timestamp.
- Updated timestamp.

Required behavior:

- Upload from admin.
- Validate file type and size.
- Store original file.
- Generate image variants such as thumbnail, medium, large, and web-optimized formats.
- Attach media to content.
- Support a single featured image per page or post.
- Support gallery-style relationships.
- Detect and clean orphaned media safely.
- Prevent path traversal and unsafe file execution.
- Keep media URLs stable across deployments.

### Blocks And Rich Content

The CMS must support structured content beyond plain HTML.

Required block concepts:

- Paragraph.
- Heading.
- Image.
- Gallery.
- Quote.
- Button.
- Columns.
- Embed.
- Custom HTML with strict permissions.
- Reusable block.

Block requirements:

- Store block type and validated data.
- Render blocks through templates.
- Sanitize rich text.
- Preserve block ordering.
- Allow migration of block schemas over time.
- Provide API representation for headless clients.

### Menus

Menus are editable navigation structures.

Required fields:

- ID.
- Name.
- Location key.
- Items.
- Locale.

Menu item types:

- Internal page.
- Internal post.
- Taxonomy archive.
- Custom URL.
- External URL.

Required behavior:

- Drag-and-drop ordering in admin.
- Nested menu items.
- Active item resolution on public pages.
- Per-locale labels.
- API output by location.

### Settings

Settings store global site configuration.

Required setting groups:

- Site identity.
- Contact.
- Social links.
- SEO defaults.
- Localization.
- Reading and homepage settings.
- Permalink settings.
- Media settings.
- Email settings.
- Comment settings.
- Theme settings.
- Security settings.

Setting requirements:

- Typed values.
- Public and private visibility.
- Cache public settings.
- Invalidate settings cache on writes.
- Audit changes.
- Validate values by type.

Minimum public settings:

- Site name.
- Site description.
- Contact email.
- Social URLs.
- Default locale.
- Supported locales.
- Active theme.

### Users, Roles, And Permissions

Required user roles:

- Administrator.
- Editor.
- Author.
- Contributor.
- Viewer or subscriber.

The exact names can be changed, but the capability model must distinguish:

- Full system access.
- Content editing access.
- Own-content-only access.
- Publish permission.
- Settings permission.
- User management permission.
- Media management permission.
- Theme management permission.
- API token management permission.

User requirements:

- Secure password hashing.
- Optional multi-factor authentication architecture.
- Session management.
- API tokens for integrations.
- Login throttling.
- Password reset flow.
- Email verification where needed.
- Last login tracking.
- Account lock or deactivation.

Authorization must be checked in application services and admin/API handlers.

### Comments And Feedback

If comments are included, they must support:

- Enable or disable globally.
- Enable or disable per post/page.
- Moderation queue.
- Approved, pending, spam, and trashed states.
- Author name, email, URL, IP hash, and user agent metadata.
- Threaded replies.
- Anti-spam hooks.
- Rate limiting.
- Public API or rendered output only for approved comments.

### Forms And Submissions

If the CMS includes forms, they must support:

- Form definitions.
- Field validation.
- Public submission endpoint.
- Spam protection.
- Admin submissions list.
- Email notification.
- Export.
- Private data redaction or deletion.

## Localization

Localization is a core workflow.

Requirements:

- Configurable supported locales.
- Configurable fallback locale.
- Locale selection from route parameter, first URL segment, user preference, cookie, or `Accept-Language`.
- Localized title, slug, content, SEO metadata, taxonomy names, menu labels, and public settings.
- Admin locale tabs or equivalent editing UX.
- API locale negotiation.
- Slug lookup across configured locales when appropriate.
- Tests for localized output and fallback behavior.

URL strategies must be explicit:

- Locale prefix strategy, such as `/en/about` and `/ru/o-nas`.
- Default-locale-without-prefix strategy, if desired.
- Per-locale slug uniqueness.

## Publishing Workflow

Every publishable entity must support:

- Draft.
- Scheduled.
- Published.
- Archived.
- Trashed.

Required rules:

- Public site and public API return only published content unless preview/auth rules allow otherwise.
- Scheduled content becomes visible only after its publish timestamp.
- Unpublishing must remove content from public listings and search indexes.
- Publishing must trigger cache invalidation.
- Every publish, unpublish, schedule, restore, and delete action should be audit logged.

## Revisions And Preview

Required revision behavior:

- Store revisions for pages and posts.
- Track author, timestamp, status, title, content, SEO data, template, and relevant metadata.
- Allow viewing revision history.
- Allow restoring a revision.
- Avoid infinite growth through retention rules or pruning.

Preview behavior:

- Editors can preview drafts.
- Preview URLs must not expose unpublished content publicly.
- Preview links should expire or require authentication.

## SEO Requirements

The CMS must support:

- SEO title.
- SEO description.
- Canonical URL.
- Open Graph metadata.
- Social image.
- Robots directives.
- XML sitemap.
- RSS or content feed.
- Structured data hooks.
- Redirect management.
- Slug/permalink settings.
- 404 tracking or diagnostics.

SEO output must be available in both rendered HTML and API responses where useful.

## Search

Search must support:

- Public search route.
- API search endpoint.
- Indexing pages, posts, taxonomies, and selected metadata.
- Locale-aware search.
- Excluding drafts and private content.
- Reindex command or background job.
- Cache or index invalidation after content changes.

The architecture must allow replacing the search backend without changing domain services.

## Themes And Templates

The CMS must have a theme system, even if the first version ships with one default theme.

Theme requirements:

- Theme metadata.
- Theme assets.
- Layout templates.
- Content templates.
- Block templates.
- Template selection per page/post.
- Menu locations.
- Theme settings.
- Safe template rendering.
- Cache compiled or parsed templates if needed.

Theme management requirements:

- Activate theme.
- Preview theme.
- Validate theme before activation.
- Keep admin UI independent from public theme.

Do not let theme files own business logic. Themes render data; services compute data.

## Extensions And Hooks

The CMS should have extension points without making the core unstable.

Extension concepts:

- Events.
- Hooks.
- Module registration.
- Admin menu contribution.
- Public route contribution.
- Block type registration.
- Settings group registration.
- Background job registration.
- Webhook registration.

Rules:

- Keep extension APIs explicit.
- Do not allow extensions to bypass authorization.
- Do not allow extensions to mutate core state without going through services.
- Validate and sandbox extension inputs where possible.
- Keep extension failures isolated and logged.

## Import, Export, And Migration

A WordPress-style replacement needs migration support.

Required capabilities:

- Import pages.
- Import posts.
- Import taxonomies.
- Import media metadata.
- Import users when safe.
- Import redirects.
- Export content.
- Export settings.
- Export media manifest.
- Dry-run mode.
- Error report.
- Idempotent retry strategy.

Do not assume imported content is trusted. Sanitize and validate it.

## Admin Dashboard

The dashboard should show:

- Content counts by status.
- Recent drafts.
- Recently published content.
- Scheduled content.
- Recent media uploads.
- Moderation queue.
- System health.
- Failed background jobs.
- Storage usage.
- Recent audit events.

Dashboard widgets should be permission-aware.

## Audit Logging

Audit log must record:

- User ID.
- Action.
- Entity type.
- Entity ID.
- Before and after summary where safe.
- IP metadata where appropriate.
- Timestamp.
- Request or correlation ID.

Important actions:

- Login and logout.
- Failed login.
- Password reset.
- User and role changes.
- Content creation, update, publish, unpublish, delete, restore.
- Media upload and delete.
- Settings changes.
- Theme activation.
- API token creation and revocation.

## Caching

Cache boundaries must be explicit.

Cache candidates:

- Public settings.
- Menus.
- Rendered pages.
- API list responses.
- Taxonomy archives.
- Sitemap.
- Search index.

Invalidation must happen after:

- Content writes.
- Publish state changes.
- Taxonomy changes.
- Menu changes.
- Settings changes.
- Media changes.
- Theme activation.

Never cache permission-sensitive responses as public data.

## Security Checklist

Before finishing a feature, check:

- All public input is validated.
- Output is escaped by default.
- Rich text is sanitized.
- File uploads validate type, size, and storage path.
- API errors do not leak internal details.
- Admin routes require authentication.
- Admin actions require authorization.
- Sessions are protected against fixation and CSRF.
- Public forms are rate limited.
- Login is rate limited.
- Passwords are securely hashed.
- API tokens can be revoked.
- Private settings are never exposed through public API.
- Draft or private content cannot be fetched without authorization.
- Media storage does not allow executable uploads.
- Audit log records sensitive administrative actions.

## Performance Checklist

Before finishing a feature, check:

- No N+1 queries in lists, API endpoints, or admin tables.
- Pagination is used for large lists.
- Search does not scan huge tables on every request without a strategy.
- Media variants are generated outside the main request where possible.
- Public settings and menus are cached.
- Expensive rendered pages can be cached and invalidated.
- Indexes exist for slugs, status, published timestamps, locale lookup, parent IDs, taxonomy joins, and user email.
- Background jobs are idempotent.
- Large imports stream data instead of loading everything into memory.

## Testing Expectations

Add focused tests for:

- Public API behavior.
- Public rendering routes.
- Admin authorization.
- Login and session behavior.
- Page and post publishing rules.
- Scheduled publishing.
- Draft preview.
- Localization and slug lookup.
- Taxonomy assignment.
- Menu rendering.
- Settings cache invalidation.
- Media upload and variant metadata.
- Search visibility.
- Revision restore.
- Import/export behavior.
- Permission-sensitive workflows.

Testing style:

- Use factories or builders instead of huge fixtures.
- Make locale and status explicit.
- Test domain services separately from HTTP handlers.
- Test response DTOs separately from persistence models.
- Include regression tests for bugs.

## Operational Requirements

The CMS must include:

- Environment-based configuration.
- Database migrations.
- Seed data for local development.
- Admin bootstrap command or setup flow.
- Health endpoint.
- Readiness endpoint.
- Structured logs.
- Request IDs.
- Background worker process.
- Scheduled task process.
- Backup strategy.
- Restore documentation.
- Media cleanup command.
- Cache clear command.
- Search reindex command.
- Sitemap generation command.
- Safe maintenance mode.

Deployment must document:

- Required environment variables.
- Database setup.
- Storage setup.
- Build step for admin/public assets if any.
- Migration step.
- Worker startup.
- Scheduler startup.
- Backup and restore procedure.

## Minimum Admin Sections

The first complete version must ship with these admin navigation groups.

### Content

- Pages.
- Posts.
- Taxonomies.
- Menus.
- Media Library.
- Comments or Feedback, if enabled.

### Design

- Themes.
- Templates.
- Navigation locations.
- Block library.

### System

- Users.
- Roles and permissions.
- Site settings.
- Localization.
- SEO settings.
- API tokens.
- Webhooks.
- Audit log.
- System health.

## Minimum Public Routes

The first complete version must support:

- `/`
- `/{page-slug}`
- `/posts`
- `/posts/{post-slug}`
- `/category/{category-slug}`
- `/tag/{tag-slug}`
- `/search`
- `/sitemap.xml`
- `/feed.xml`
- Locale-prefixed equivalents when localization is enabled.

Permalink patterns must be configurable without breaking old URLs. If URLs change, redirect management must handle old paths.

## Minimum API Contract

The first complete version must expose:

```http
GET /api/v1/pages
GET /api/v1/pages/{slug}
GET /api/v1/posts
GET /api/v1/posts/{slug}
GET /api/v1/taxonomies
GET /api/v1/taxonomies/{type}/{slug}
GET /api/v1/menus/{location}
GET /api/v1/settings
GET /api/v1/search
```

All list endpoints must support:

- Pagination.
- Locale selection.
- Stable sorting.
- Filtering by taxonomy where appropriate.

All detail endpoints must:

- Resolve localized slugs.
- Return 404 for unpublished content unless preview/auth rules allow access.
- Include SEO metadata.
- Include related media metadata.

## WordPress Feature Parity Checklist

The CMS is not complete until it covers these WordPress-like capabilities:

- Admin dashboard.
- Pages.
- Posts.
- Categories.
- Tags.
- Media library.
- Menus.
- Themes/templates.
- Widgets or reusable blocks.
- Users.
- Roles and permissions.
- Settings.
- Permalinks.
- SEO metadata.
- Comments or a deliberate documented decision to omit comments.
- Revisions.
- Draft preview.
- Scheduled publishing.
- Search.
- RSS/feed.
- Sitemap.
- Import/export.
- Plugin/module extension points.
- Public API.
- Localization.
- Backups and restore guidance.
- Security hardening.
- Performance and cache strategy.

## Implementation Order

Build in this order to avoid gaps:

1. Configuration, migrations, domain entities, and repositories.
2. Auth, users, roles, sessions, and admin bootstrap.
3. Pages, posts, statuses, slugs, revisions, and publishing rules.
4. Taxonomies and content assignment.
5. Settings with cache invalidation.
6. Media upload, metadata, variants, and cleanup.
7. Admin GUI for content, media, users, roles, settings, and taxonomy management.
8. Public rendering with themes, templates, menus, and SEO metadata.
9. Public API with versioned response DTOs.
10. Localization across admin, public site, and API.
11. Search, sitemap, feed, and redirects.
12. Comments/forms if enabled.
13. Import/export and migration tools.
14. Audit logging, system health, backups, and operational commands.
15. Extension points and module boundaries.

## Definition Of Done

A feature is done only when:

- Domain rules are in services or domain objects.
- HTTP handlers are thin.
- Admin UI uses application services.
- API responses use explicit DTOs.
- Validation is complete.
- Authorization is checked.
- Localization behavior is defined.
- Cache invalidation is handled.
- Audit logging is added for important changes.
- Tests cover the main behavior and permission boundaries.
- Documentation explains how to operate the feature.

## Final Goal

The final CMS should feel complete to a WordPress user:

- They can log into `/admin`.
- They can create and publish pages and posts.
- They can upload media.
- They can manage menus, settings, users, SEO, and themes.
- They can preview drafts and restore revisions.
- They can expose the same content through a public site and through API.
- They can run the system in production with backups, logs, health checks, and predictable updates.

The implementation should feel natural to a Go developer:

- Clear package boundaries.
- Explicit dependencies.
- Small handlers.
- Testable services.
- Replaceable infrastructure adapters.
- No hidden framework magic.
- No unnecessary coupling to a specific library.
