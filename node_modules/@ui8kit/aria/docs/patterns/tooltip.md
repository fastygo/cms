# Tooltip

## APG (W3C)

The **[Tooltip pattern](https://www.w3.org/WAI/ARIA/apg/patterns/tooltip/)** describes a **popup** tied to a trigger, shown on hover or keyboard focus, with `role="tooltip"` and an accessible relationship to the owning element. APG emphasises persistence, pointer hover bridges, and not using tooltips for critical-only information; read their examples before relying on tooltips for essential content.

## `@ui8kit/aria` contract

- Root: `[data-ui8kit="tooltip"]` (container associating trigger and popup)
- Content node: `[role="tooltip"]`
- Open on hover/focus of the trigger; close on blur/leave as implemented.

## In this repository

- Implementation: [`src/patterns/tooltip/`](../../src/patterns/tooltip/)
- Selector reference: [`src/patterns/tooltip/markup.md`](../../src/patterns/tooltip/markup.md)
- Fixture: [`src/patterns/tooltip/fixtures/tooltip.html`](../../src/patterns/tooltip/fixtures/tooltip.html)
- Tests: [`src/patterns/tooltip/__tests__/`](../../src/patterns/tooltip/__tests__/)
