# Menu

## APG (W3C)

Menus are specified under **[Menu and Menubar](https://www.w3.org/WAI/ARIA/apg/patterns/menubar/)** in the APG: `role="menu"`, menuitems, typeahead, arrow/Home/End behaviour, and wrapping. That page is the canonical reference for **keyboard interaction once the menu is open**. For the **button that opens** the menu, pair this with the [Menu Button](https://www.w3.org/WAI/ARIA/apg/patterns/menu-button/) pattern (see [menubutton.md](menubutton.md) in this repo).

## `@ui8kit/aria` contract

- Root: `[data-ui8kit="menu"]`
- Item: `[data-menu-item]`
- Arrow keys and Home/End for roving focus inside an open menu.

## In this repository

- Implementation: [`src/patterns/menu/`](../../src/patterns/menu/)
- Selector reference: [`src/patterns/menu/markup.md`](../../src/patterns/menu/markup.md)
- Fixture: [`src/patterns/menu/fixtures/menu.html`](../../src/patterns/menu/fixtures/menu.html)
- Tests: [`src/patterns/menu/__tests__/`](../../src/patterns/menu/__tests__/)
- Often used together with menubutton: [`src/patterns/menubutton/fixtures/menubutton.html`](../../src/patterns/menubutton/fixtures/menubutton.html)
