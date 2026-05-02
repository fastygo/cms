# Switch

## APG (W3C)

The **[Switch pattern](https://www.w3.org/WAI/ARIA/apg/patterns/switch/)** defines an on/off control with **`role="switch"`** and **`aria-checked`**, distinct from checkbox and toggle button semantics. APG documents keyboard support and labelling; follow it for accessible names and state text.

## `@ui8kit/aria` contract

- Root: `[data-ui8kit="switch"]` or `[role="switch"]`
- Control: `[data-switch-control]`
- Toggle via pointer, Space, or Enter as implemented for the control.

## In this repository

- Implementation: [`src/patterns/switch/`](../../src/patterns/switch/)
- Selector reference: [`src/patterns/switch/markup.md`](../../src/patterns/switch/markup.md)
- Fixture: [`src/patterns/switch/fixtures/switch.html`](../../src/patterns/switch/fixtures/switch.html)
- Tests: [`src/patterns/switch/__tests__/`](../../src/patterns/switch/__tests__/)
- Example: [`examples/spa/index.html`](../../examples/spa/index.html) (`data-ui8kit="switch"`)
