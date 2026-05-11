# GoCMS Implementation Roadmap

This roadmap translates the compatibility documents into practical implementation passes.

## Source Documents

- `../../../go-codex/en`: external compatibility contract.
- `../../../go-stack/en`: Go-native architecture profile.
- `../../../go-ui8kit/en`: UI8Kit admin profile.

## Implementation Strategy

GoCMS should be built in passes:

1. Core kernel.
2. REST control plane.
3. Headless admin MVP.
4. Runtime profiles and playground boundary.
5. GraphQL plugin.
6. Marketplace-style frontend theme validation.
7. CMS admin finalization on the panel core.
8. Full entity coverage and conformance hardening.
9. Production CMS completion.

Each pass should leave the system runnable and testable.

## Engineering Guardrails

Before implementation, read:

- `08-engineering-principles.md`
- `09-architecture-fitness-functions.md`
- `10-technical-debt-policy.md`
- `03-5-runtime-profiles-playground.md`

These documents define how GoCMS protects compatibility, migration readiness, and architecture boundaries before code is written.

## Success Definition

The first complete implementation is successful when:

- The admin can manage all core CMS entities.
- REST exposes the full compatibility surface.
- Runtime profiles separate headless, admin, playground, full, and conformance binaries.
- Playground mode runs a complete isolated CMS sandbox, including admin and public preview/rendering, without server-side persistence.
- GraphQL can be installed as a plugin and use the same services.
- Public rendering can be disabled for headless mode.
- A frontend can render a complete marketplace-style theme using API data.
- Themes and plugins can be mapped through compatibility contracts and migration adapters where possible.
- Existing content-heavy CMS installations can migrate into GoCMS with a one-click workflow that preserves public behavior as much as possible.
- Conformance checks prove drafts, private metadata, capabilities, media, menus, taxonomies, authors, and content types behave correctly.
