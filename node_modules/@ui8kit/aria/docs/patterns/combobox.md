# Combobox

## APG (W3C)

The **[Combobox pattern](https://www.w3.org/WAI/ARIA/apg/patterns/combobox/)** defines an input with an associated popup (often a listbox), keyboard navigation, and selection semantics. APG lists **several combobox variants** (e.g. popup behaviour, focus vs `aria-activedescendant`); use their examples and keyboard tables to decide which variant your markup follows, then align attributes with that choice.

## `@ui8kit/aria` contract

- Root: `[data-ui8kit="combobox"]`
- Trigger/input: `input` inside the root
- Toggle (optional): `[data-combobox-toggle]`
- List: `ul` or `[role="listbox"]`
- Option: `[data-combobox-option]`
- `data-state` on the root tracks popup open/closed (`open` | `closed`).

## In this repository

- Implementation: [`src/patterns/combobox/`](../../src/patterns/combobox/)
- Selector reference: [`src/patterns/combobox/markup.md`](../../src/patterns/combobox/markup.md)
- Fixture: [`src/patterns/combobox/fixtures/combobox.html`](../../src/patterns/combobox/fixtures/combobox.html)
- Tests: [`src/patterns/combobox/__tests__/`](../../src/patterns/combobox/__tests__/)
- Example: [`examples/htmx/index.html`](../../examples/htmx/index.html) (`data-ui8kit="combobox"` in the swap zone)
