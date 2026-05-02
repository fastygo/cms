# Pattern index

This folder is an **onboarding map** for the [WAI-ARIA Authoring Practices Guide (APG) patterns](https://www.w3.org/WAI/ARIA/apg/patterns/). Each pattern page here:

- Links to the **matching APG pattern** (keyboard interaction, roles, states, and worked **examples** live there).
- Summarises the **`@ui8kit/aria` DOM contract** (`data-ui8kit`, triggers, and related attributes) so you can see how W3C guidance maps to this package.
- Points to **implementation and tests** under `src/patterns/<name>/` and to **examples** under `examples/` where the widget appears.

Implementations live under `src/patterns/<name>/`; unit tests under `src/patterns/<name>/__tests__/`. Where a pattern ships a short selector reference, see `src/patterns/<name>/markup.md`.

| Pattern | Doc | Summary |
| --- | --- | --- |
| Accordion | [accordion.md](accordion.md) | Expand/collapse sections; single vs multiple; trigger/content pairing. |
| Alert | [alert.md](alert.md) | Live region semantics (`role`, `aria-live`, `aria-atomic`). |
| Combobox | [combobox.md](combobox.md) | Editable input + listbox; filtering, keyboard navigation, selection. |
| Dialog | [dialog.md](dialog.md) | Modal/sheet/alertdialog; open/close controls, overlay, focus behaviour. |
| Disclosure | [disclosure.md](disclosure.md) | Simple show/hide with `aria-expanded` on a trigger. |
| Listbox | [listbox.md](listbox.md) | Options list; roving tabindex and arrow-key selection. |
| Menu | [menu.md](menu.md) | Menu container; items and keyboard roving inside an open menu. |
| Menu button | [menubutton.md](menubutton.md) | Button that opens a linked menu; outside click closes. |
| Switch | [switch.md](switch.md) | Toggle with `role="switch"` and `aria-checked`. |
| Tabs | [tabs.md](tabs.md) | Tablist semantics; selected tab and associated panels. |
| Tooltip | [tooltip.md](tooltip.md) | Hover/focus disclosure of auxiliary content. |

For **CDN / IIFE** usage you typically load the full bundle (`all.iife.js`) so every pattern above is registered. For **ESM** you import only the patterns you need and call `registerPattern` + `init` yourself (see [Getting started](../getting-started.md)).
