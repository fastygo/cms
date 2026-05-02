# 06. Theme Rendering Interface

This document defines Go-native theme rendering boundaries.

## Purpose

Theme rendering turns prepared content, settings, menus, and SEO data into public output. Rendering must not own content business rules.

## Renderer Interface

A rendering boundary should expose operations such as:

- Resolve route target.
- Select active theme.
- Resolve template role.
- Build view model.
- Render output.
- Apply cache policy.

The renderer should accept a request context and a prepared render query. It should return rendered output or a classified error.

## View Models

View models should be explicit and delivery-safe.

Common view model data:

- Site settings.
- Locale.
- Current route.
- Current content entry.
- Taxonomies.
- Menus.
- Pagination.
- SEO metadata.
- Theme settings.
- Slot data.
- Public user state where relevant.

Private fields should never be included in public view models unless preview authorization allows it.

## Template Resolution

Template resolution should implement the order defined by `../../go-codex/en/04-theme-contract.md`.

Resolution should be:

- Deterministic.
- Testable.
- Logged when fallback occurs.
- Independent from storage details.

## Public Rendering Mode

When public rendering is enabled:

- Published public content may be rendered.
- Draft content must be hidden.
- Scheduled content must respect publish time.
- Private content must require authorization.
- SEO metadata should be rendered.
- Menus and settings should come from services or caches.

## Headless Mode

When headless mode disables public rendering:

- Public HTML routes may return `404`, `410`, or a documented disabled response.
- Admin and REST compatibility should remain available according to declared compatibility level.
- Theme management may remain available for preview or future activation, but public rendering is not required.

## Preview

Preview rendering should:

- Require authentication or signed tokens.
- Use unpublished draft state.
- Avoid caching as public output.
- Mark output as preview where possible.

## Slots

Slots should be resolved after content and before final rendering.

Slot contributors should receive:

- Context.
- Slot ID.
- View model subset.
- Principal or public visitor state where relevant.

Slot output should be sanitized or treated as trusted only if it comes from trusted code.

## Cache

Rendered output may be cached when:

- The response is public.
- The content is published.
- The user state does not affect output.
- The locale and theme are part of cache key.

Cache invalidation should be triggered by content, menu, settings, media, theme, and plugin changes.

## Security

Rendering must enforce:

- Escaping by default.
- Rich content sanitization.
- Content visibility.
- Private settings exclusion.
- Preview authorization.

Theme code should not be allowed to query private storage directly.

## Testing

Renderer tests should cover route resolution, template fallback, preview access, headless mode, public cache policy, and private content leakage prevention.
