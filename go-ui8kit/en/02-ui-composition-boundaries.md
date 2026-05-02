# 02. UI Composition Boundaries

This document defines UI composition boundaries for UI8Kit-based GoCMS admin screens.

## Package Layers

Recommended app UI layers:

```text
internal/site/views
internal/site/ui/elements
internal/site/ui/blocks
```

Responsibilities:

- `views`: page-level assembly and route-specific view models.
- `elements`: reusable widget-level components.
- `blocks`: reusable top-level sections and admin screen regions.

## Elements

Elements should be:

- Brand-neutral.
- Portable.
- Data-focused.
- Free of hidden request/session coupling.
- Built from UI8Kit primitives and neutral composites.

Examples:

- Status badge.
- Capability badge.
- Field help text.
- Media thumbnail.
- Pagination control.
- Validation summary.
- Action button row.

Elements should not own public anchors or top-level landmarks.

## Blocks

Blocks should represent reusable admin sections.

Examples:

- Content table.
- Publish panel.
- Taxonomy assignment panel.
- Media picker panel.
- Settings card.
- Plugin status panel.
- Theme preview panel.

Blocks may compose UI8Kit primitives, neutral composites, and local elements.

## Views

Views should:

- Assemble shell, blocks, and elements.
- Bind route-specific view models.
- Avoid low-level one-off markup when a reusable block or element is appropriate.
- Avoid direct storage or service access.
- Avoid business rules.

## Archive Mirrors

If archive or shared examples exist, treat them as references only.

Rules:

- Do not import archive blocks or elements directly into app code.
- Copy or adapt reusable candidates into the local UI layer.
- Keep local candidates mirror-friendly.
- Regenerate generated files locally.

## State Boundaries

Reusable UI packages should receive explicit props.

They should not reach into:

- Request context.
- Session state.
- Global config.
- Storage.
- Application services.

Views or handlers prepare data before calling UI components.

## Naming

Names should describe reusable UI concepts, not one-off product copy.

Good:

- `StatusBadge`
- `ContentTable`
- `PublishPanel`
- `MediaPicker`

Avoid:

- Feature-specific decorative names.
- Brand campaign names.
- Hidden business workflow names in low-level elements.

## Conformance Markers

Reusable blocks that represent required admin behavior should expose stable markers such as:

```text
data-gocms-block="publish-panel"
data-gocms-block="content-table"
data-gocms-element="status-badge"
```
