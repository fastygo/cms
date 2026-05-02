# 03. HTML5 And ARIA Policy

This document defines semantic markup and ARIA boundaries for UI8Kit-based GoCMS admin screens.

## HTML5 Semantics

HTML5 semantics are mandatory. Choose tags for meaning first and styling second.

Admin templates should request semantic tags through UI8Kit primitives and props. App templates should not hand-write raw layout tags when UI8Kit provides the semantic primitive.

## Block And Box

Rules:

- Use `Block` for top-level sections, landmarks, and public anchors.
- Use `Box` for internal layout.
- Do not nest `Block` inside `Block`.
- Promote nested anchors to sibling top-level blocks.
- Put public section IDs on `Block`.
- Native form/control IDs remain owned by form primitives.

## Strict Content Models

Use dedicated primitive families for strict HTML5 content models:

- Tables use table primitives.
- Lists use list primitives.
- Description lists use description-list primitives.
- Media uses picture, image, figure, and source primitives.
- Disclosure uses disclosure primitives.
- Forms use form, fieldset, legend, field, and control primitives.

Generic layout containers should not be used to fake strict content-model structures.

## ARIA Ownership

Server-rendered UI8Kit markup owns the initial accessibility contract.

Runtime ARIA behavior owns:

- `hidden` state.
- `data-state`.
- `aria-expanded`.
- `aria-selected`.
- `aria-controls`.
- Focus restoration.
- Keyboard navigation.

Preserve UI8Kit data hooks and roles emitted by composites.

## Interactive Patterns

Interactive admin patterns include:

- Dialog.
- Alert dialog.
- Sheet.
- Tabs.
- Accordion.
- Combobox.
- Tooltip.
- Disclosure.

Do not implement page-local ARIA behavior for these patterns when UI8Kit owns the pattern.

If a required pattern is missing, extend the shared UI layer and runtime behavior together instead of hand-rolling a one-off implementation.

## ARIA References

ARIA references must point to existing elements in the rendered DOM:

- `aria-controls`
- `aria-labelledby`
- `aria-describedby`

Conditional rendering must not leave broken references.

## Admin Forms

Form fields should expose:

- Label.
- Help text where needed.
- Error text.
- Required state.
- Disabled state.
- Described-by references.

Validation summaries should link or refer to invalid fields where practical.

## Admin Tables

Tables should use table semantics for tabular data.

Content management lists should support:

- Header cells.
- Sort state.
- Row selection.
- Bulk action controls.
- Empty state.
- Pagination.

Keyboard and screen-reader behavior should remain usable without custom table-like div structures.

## Conformance Markers

Conformance markers must not replace accessibility attributes. They are testing hooks only.

Use both:

- Semantic markup and ARIA for users.
- `data-gocms-*` markers for tests.
