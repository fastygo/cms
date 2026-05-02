# 04. Theme Contract

This document defines how GoCMS-compatible themes are described, loaded, resolved, and rendered.

## Purpose

Themes provide public presentation without owning core business rules. A theme receives prepared data and renders public output.

## Theme Independence

Themes MAY be implemented using any rendering technology supported by the engine.

The contract requires:

- A manifest.
- Stable template roles.
- Stable slots.
- Asset declaration.
- Capability boundaries.
- Deterministic template resolution.

It does not require a specific template engine or asset pipeline.

## Theme Manifest

Each theme MUST provide a manifest.

Required fields:

```json
{
  "id": "example-theme",
  "name": "Example Theme",
  "version": "1.0.0",
  "contract": "0.1",
  "description": "A theme.",
  "author": "Example",
  "templates": {},
  "assets": {},
  "slots": []
}
```

Required manifest fields:

- `id`
- `name`
- `version`
- `contract`

Theme IDs MUST be stable, lowercase, and URL-safe.

## Template Roles

Themes SHOULD define templates for:

- `index`
- `front`
- `page`
- `post`
- `archive`
- `taxonomy`
- `search`
- `not_found`
- `error`

Themes MAY define more specific templates:

- `page:{slug}`
- `post:{slug}`
- `content-kind:{kind}`
- `taxonomy:{type}`
- `taxonomy:{type}:{slug}`
- `archive:{kind}`

The syntax above defines role identifiers, not file names.

## Template Resolution

Template resolution MUST be deterministic.

Recommended resolution order for a page:

1. Explicit content template.
2. `page:{slug}`.
3. `page`.
4. `index`.

Recommended resolution order for a post:

1. Explicit content template.
2. `post:{slug}`.
3. `content-kind:post`.
4. `post`.
5. `index`.

Recommended resolution order for taxonomy archive:

1. `taxonomy:{type}:{slug}`.
2. `taxonomy:{type}`.
3. `taxonomy`.
4. `archive`.
5. `index`.

If no template can be resolved, the engine MUST return a controlled server error and MUST NOT expose internal details.

## Slots

Themes MUST declare supported slots if they allow admin or plugin placement.

Common slots:

- `header`
- `footer`
- `sidebar`
- `before_content`
- `after_content`
- `content_top`
- `content_bottom`

Plugins MAY contribute output to slots only through documented hook or slot APIs.

## Assets

Themes MUST declare assets they require.

Asset declarations SHOULD include:

- Identifier.
- Type: `css`, `js`, `font`, `image`, or other documented type.
- Source path or resolver key.
- Dependencies.
- Load location: `head`, `body_end`, or slot-specific.
- Integrity or hash where available.

Theme assets MUST NOT bypass engine security policies.

## Data Contract

Theme templates MUST receive data through stable view models.

Common data:

- Site settings.
- Current request context.
- Current locale.
- Current user public state where relevant.
- Current content entry.
- Menu data.
- Taxonomy data.
- Pagination data.
- SEO metadata.
- Registered slots.

Themes MUST NOT query persistence directly unless the engine explicitly allows a safe theme API. The preferred model is prepared data from rendering services.

## Theme Settings

Themes MAY define settings in the manifest.

Theme settings MUST include:

- Stable key.
- Type.
- Default value.
- Public/private visibility.
- Validation rule.

Theme settings MUST be managed through the same settings and capability model as other settings.

## Activation

Theme activation MUST:

- Validate manifest.
- Validate required templates.
- Validate contract compatibility.
- Preserve previous active theme if validation fails.
- Emit a theme activation event.

Activation MUST require a capability such as `themes.activate`.

## Preview

Implementations SHOULD support theme preview before activation.

Preview MUST NOT change the active theme for normal visitors.

## Security

Themes MUST NOT:

- Bypass content visibility rules.
- Bypass capability checks.
- Execute untrusted uploaded files.
- Expose private settings.
- Expose draft content without preview authorization.

## Conformance Notes

Conformance tests SHOULD verify:

- Manifest validation.
- Active theme rollback after activation failure.
- Template resolution order.
- Slot registration.
- Asset declaration.
- Private content remains hidden during rendering.
