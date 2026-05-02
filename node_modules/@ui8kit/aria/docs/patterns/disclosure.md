# Disclosure

## APG (W3C)

The **[Disclosure (Show/Hide) pattern](https://www.w3.org/WAI/ARIA/apg/patterns/disclosure/)** is the minimal expand/collapse widget: a control with **`aria-expanded`** linked to the shown or hidden content. APG examples show button vs other controls and heading integration; read there for authoring details.

## `@ui8kit/aria` contract

- Root: `[data-ui8kit="disclosure"]`
- Trigger: `[data-disclosure-trigger]`
- Content: `[data-disclosure-content]`

## In this repository

- Implementation: [`src/patterns/disclosure/`](../../src/patterns/disclosure/)
- Selector reference: [`src/patterns/disclosure/markup.md`](../../src/patterns/disclosure/markup.md)
- Fixture: [`src/patterns/disclosure/fixtures/disclosure.html`](../../src/patterns/disclosure/fixtures/disclosure.html)
- Tests: [`src/patterns/disclosure/__tests__/`](../../src/patterns/disclosure/__tests__/)
