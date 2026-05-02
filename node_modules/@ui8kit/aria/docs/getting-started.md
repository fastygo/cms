# Getting started

## Install

```bash
npm i @ui8kit/aria
# or
bun add @ui8kit/aria
```

## Choose an entry

| Goal | Import / script |
| --- | --- |
| Tree-shakeable API, no globals | `import { registerPattern, getNamespace, dialog } from '@ui8kit/aria'` |
| Auto-register one pattern | `import '@ui8kit/aria/dialog'` |
| Everything + `window.ui8kit` | `import '@ui8kit/aria/all'` or IIFE `dist/all.iife(.min).js` from a CDN |

After side-effect imports (or CDN), patterns initialise on `DOMContentLoaded` unless `window.__UI8KIT_ARIA_AUTO_INIT__ === false` — then call `getNamespace().init()` (or `window.ui8kit.init()`) when your shell is ready.

## Pure ESM example

```ts
import { registerPattern, getNamespace, dialog, tabs } from '@ui8kit/aria'

registerPattern(dialog)
registerPattern(tabs)
getNamespace().init()
```

## Markup

Match the contract for each widget ([pattern index](patterns/README.md)). Wrong markup does not throw at import time; behaviour may degrade or no-op.

## Next

- [Capabilities](capabilities.md) — API surface and lifecycle.
- [Examples](examples.md) — runnable HTML under `examples/`.
