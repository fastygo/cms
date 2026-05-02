# Tabs

## APG (W3C)

The **[Tabs pattern](https://www.w3.org/WAI/ARIA/apg/patterns/tabs/)** defines **`role="tablist"`**, **`role="tab"`**, and **`role="tabpanel"`**, including activation (automatic vs manual), optional deletion, and keyboard orientation. Use APG examples to validate `aria-selected`, `aria-controls`, and `tabindex` on tabs; this package implements a concrete subset aligned with the `data-ui8kit` contract below.

## `@ui8kit/aria` contract

- Root: `[data-ui8kit="tabs"]` — optional `data-tabs-value` for initial selected tab
- Tab trigger: `[data-tabs-trigger]` with `data-tabs-value` matching a panel
- Panel: `[data-tabs-panel]` with the same `data-tabs-value`
- Arrow Left/Right and Up/Down move focus between tabs (APG-aligned tablist navigation).

## In this repository

- Implementation: [`src/patterns/tabs/`](../../src/patterns/tabs/)
- Selector reference: [`src/patterns/tabs/markup.md`](../../src/patterns/tabs/markup.md)
- Fixture: [`src/patterns/tabs/fixtures/tabs.html`](../../src/patterns/tabs/fixtures/tabs.html)
- Tests: [`src/patterns/tabs/__tests__/`](../../src/patterns/tabs/__tests__/)
- Example: [`examples/plain/index.html`](../../examples/plain/index.html) (tabs section)
