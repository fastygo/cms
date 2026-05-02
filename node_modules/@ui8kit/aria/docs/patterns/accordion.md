# Accordion

## APG (W3C)

Start with the **[Accordion pattern](https://www.w3.org/WAI/ARIA/apg/patterns/accordion/)** in the APG. There you will find the intended **heading / trigger / panel** relationships, keyboard behaviour (including optional shortcuts), and **example implementations** linked from that page. Use APG as the source of nuance (e.g. one vs many sections open, focus management); this doc only anchors how that maps to `@ui8kit/aria`.

## `@ui8kit/aria` contract

- Root: `[data-ui8kit="accordion"]`
- Item: `[data-accordion-item]`
- Trigger: `[data-ui8kit-accordion-trigger]`
- Content: `[data-ui8kit-accordion-content]`
- `data-state` on each item reflects open/closed (`open` | `closed`).

Match your markup to APG roles and labels (`aria-controls`, `aria-labelledby`, `role="region"` on panels where appropriate); the library wires behaviour to the `data-ui8kit` contract above.

## In this repository

- Implementation: [`src/patterns/accordion/`](../../src/patterns/accordion/)
- Fixture (reference markup): [`src/patterns/accordion/fixtures/accordion.html`](../../src/patterns/accordion/fixtures/accordion.html)
- Tests: [`src/patterns/accordion/__tests__/`](../../src/patterns/accordion/__tests__/)
- Example: [`examples/plain/index.html`](../../examples/plain/index.html) (accordion section)

There is no separate `markup.md` for accordion; the fixture and this page are the contract reference.
