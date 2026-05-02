# Listbox

## APG (W3C)

The **[Listbox pattern](https://www.w3.org/WAI/ARIA/apg/patterns/listbox/)** describes a list of **options** with directional navigation, selection, and optional multi-select semantics. APG distinguishes listbox-only widgets from listboxes **owned by a combobox**; if your list sits inside a combobox, read both the listbox and [combobox](https://www.w3.org/WAI/ARIA/apg/patterns/combobox/) pages.

## `@ui8kit/aria` contract

- Root: `[data-ui8kit="listbox"]`
- Option: `[role="option"]`
- **Single selection** in the current implementation.

## In this repository

- Implementation: [`src/patterns/listbox/`](../../src/patterns/listbox/)
- Selector reference: [`src/patterns/listbox/markup.md`](../../src/patterns/listbox/markup.md)
- Fixture: [`src/patterns/listbox/fixtures/listbox.html`](../../src/patterns/listbox/fixtures/listbox.html)
- Tests: [`src/patterns/listbox/__tests__/`](../../src/patterns/listbox/__tests__/)
