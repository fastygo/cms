# 05. Theme Frontend Validation

The frontend validation app proves that GoCMS can power a complete marketplace-style CMS theme in headless mode.

## Purpose

This is not the core public renderer. It is a validation frontend that consumes REST or GraphQL and exposes missing API/model capabilities.

## Required Pages

The validation frontend must render:

- Home page.
- Static page.
- Blog archive.
- Single post.
- Category archive.
- Tag archive.
- Custom taxonomy archive.
- Author archive.
- Search results.
- 404 page.

## Required Theme Sections

The frontend must include:

- Header.
- Nested navigation menu.
- Footer.
- Footer menus.
- Hero section.
- Featured posts.
- Card grid.
- Related posts.
- Breadcrumbs.
- Pagination.
- Sidebar-like area or secondary content region.
- Newsletter/contact placeholder if settings/forms support it later.

## Required Data

The API must provide:

- Site identity.
- Public settings.
- Menus by location.
- Content lists.
- Content detail.
- Featured media.
- Media variants.
- Authors.
- Taxonomies and terms.
- Related content data or enough filters to compute it.
- SEO title.
- SEO description.
- Canonical URL.
- Open Graph image.
- Published dates.
- Modified dates.

## Validation Method

Build the frontend as a consumer, not as a privileged internal renderer.

Rules:

- Use public APIs for public pages.
- Use authenticated APIs only for preview.
- Do not query storage directly.
- Record API gaps as backend requirements.
- Record theme data gaps as content model requirements.

## REST And GraphQL Modes

The validation frontend should eventually support both:

- REST mode.
- GraphQL mode.

The first implementation may choose one, but the data model must not be tailored to only one delivery protocol.

## Marketplace-Style Checklist

The frontend is good enough when it can show:

- Rich home page.
- Multiple post cards with images and metadata.
- Category/tag pages.
- Author page.
- Search results.
- Menus from CMS.
- SEO metadata per page.
- Responsive layout.
- Empty states.
- Loading/error states where relevant.

## Exit Checklist

- Frontend can be built without direct database access.
- All public content comes from REST or GraphQL.
- Missing fields are documented.
- API supports enough data for a complete content theme.
- Drafts and private content do not appear.
