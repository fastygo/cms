# Alert

## APG (W3C)

The **[Alert pattern](https://www.w3.org/WAI/ARIA/apg/patterns/alert/)** describes a **live region** that announces important, time-sensitive information without stealing focus. APG covers `role="alert"`, `aria-live`, timing, and examples. Read there for when to use `alert` vs other live-region strategies; here we document what this package enforces on init.

## `@ui8kit/aria` contract

- Root: `[data-ui8kit="alert"]` or `[role="alert"]`
- On init, the implementation ensures appropriate **live-region** attributes (`role`, `aria-live`, `aria-atomic`) so assistive technologies can treat the node as an alert surface.

## In this repository

- Implementation: [`src/patterns/alert/`](../../src/patterns/alert/)
- Selector reference: [`src/patterns/alert/markup.md`](../../src/patterns/alert/markup.md)
- Fixture: [`src/patterns/alert/fixtures/alert.html`](../../src/patterns/alert/fixtures/alert.html)
- Tests: [`src/patterns/alert/__tests__/`](../../src/patterns/alert/__tests__/)
