# Dialog

## APG (W3C)

For **modal dialogs**, use **[Dialog (Modal)](https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/)**: focus trap, `aria-modal`, labelled dialog, escape to close, and optional focus restoration are all spelled out there with examples.

For **`role="alertdialog"`** (workflow interruption, explicit response), see **[Alert and Message Dialogs](https://www.w3.org/WAI/ARIA/apg/patterns/alertdialog/)** in addition to the modal dialog guidance.

This package also supports a **sheet** variant (`data-ui8kit="sheet"`) as a dialog-shaped surface; keep the same APG principles (focus, label, modal semantics) unless you intentionally build a non-modal panel.

## `@ui8kit/aria` contract

- Root: `[data-ui8kit="dialog"]` | `[data-ui8kit="sheet"]` | `[data-ui8kit="alertdialog"]`
- Open: `data-ui8kit-dialog-open`, `data-ui8kit-dialog-target` (target id fallback)
- Close: `data-ui8kit-dialog-close` or click on `[data-ui8kit-dialog-overlay]`
- State: `data-state="open"` | `data-state="closed"` on the dialog root

See [`src/patterns/dialog/markup.md`](../../src/patterns/dialog/markup.md) for the full selector list.

## In this repository

- Implementation: [`src/patterns/dialog/`](../../src/patterns/dialog/)
- Selector reference: [`src/patterns/dialog/markup.md`](../../src/patterns/dialog/markup.md)
- Fixture: [`src/patterns/dialog/fixtures/dialog.html`](../../src/patterns/dialog/fixtures/dialog.html)
- Tests: [`src/patterns/dialog/__tests__/`](../../src/patterns/dialog/__tests__/)
- Example: [`examples/plain/index.html`](../../examples/plain/index.html) (dialog section)
