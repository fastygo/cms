# Menu button

## APG (W3C)

The **[Menu Button pattern](https://www.w3.org/WAI/ARIA/apg/patterns/menu-button/)** covers the **trigger** (`aria-expanded`, `aria-haspopup`, optional `aria-controls`) and how opening the menu moves focus to the first item. APG links onward to **[Menu and Menubar](https://www.w3.org/WAI/ARIA/apg/patterns/menubar/)** for behaviour **inside** the menu; use both pages when wiring a full control.

(The same content is also available at the alternate path `/patterns/menubutton/`; the hyphenated URL above is the one used in APG navigation.)

## `@ui8kit/aria` contract

- Trigger: `[data-ui8kit="menubutton"]`
- Link to menu: `data-menubutton-target` or `aria-controls` pointing at the menu root
- Menu root: `[data-ui8kit="menu"]` (see [menu.md](menu.md))

## In this repository

- Implementation: [`src/patterns/menubutton/`](../../src/patterns/menubutton/)
- Selector reference: [`src/patterns/menubutton/markup.md`](../../src/patterns/menubutton/markup.md)
- Fixture (menu + button): [`src/patterns/menubutton/fixtures/menubutton.html`](../../src/patterns/menubutton/fixtures/menubutton.html)
- Tests: [`src/patterns/menubutton/__tests__/`](../../src/patterns/menubutton/__tests__/)
