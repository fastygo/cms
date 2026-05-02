# 02. REST API Contract

This document defines the GoCMS REST compatibility surface.

## Purpose

REST is the base compatibility and control-plane API. It is required even when another delivery API, such as GraphQL, is installed and used as the primary public content API.

## Required Base Paths

An implementation that supports REST compatibility MUST expose:

```text
/go-json
/go-json/go/v2/
```

`GET /go-json` MUST return discovery information for available namespaces.

`GET /go-json/go/v2/` MUST return discovery information for the GoCMS v2 namespace.

## Discovery Shape

Discovery responses MUST include:

```json
{
  "name": "GoCMS",
  "version": "2",
  "routes": {},
  "authentication": [],
  "links": {}
}
```

Additional fields MAY be included. Clients MUST ignore unknown fields.

## Required Endpoints

The v2 namespace MUST define:

```http
GET /go-json/go/v2/posts
GET /go-json/go/v2/posts/{id}
GET /go-json/go/v2/posts/by-slug/{slug}
GET /go-json/go/v2/pages
GET /go-json/go/v2/pages/{id}
GET /go-json/go/v2/pages/by-slug/{slug}
GET /go-json/go/v2/media
GET /go-json/go/v2/media/{id}
GET /go-json/go/v2/taxonomies
GET /go-json/go/v2/taxonomies/{type}
GET /go-json/go/v2/taxonomies/{type}/{id}
GET /go-json/go/v2/menus
GET /go-json/go/v2/menus/{location}
GET /go-json/go/v2/settings
GET /go-json/go/v2/search
```

Authenticated implementations SHOULD define:

```http
POST /go-json/go/v2/posts
PATCH /go-json/go/v2/posts/{id}
DELETE /go-json/go/v2/posts/{id}
POST /go-json/go/v2/pages
PATCH /go-json/go/v2/pages/{id}
DELETE /go-json/go/v2/pages/{id}
POST /go-json/go/v2/media
PATCH /go-json/go/v2/media/{id}
DELETE /go-json/go/v2/media/{id}
```

Plugins MAY register additional routes under their own namespace:

```text
/go-json/{plugin-id}/v1/
```

Plugin namespaces MUST NOT use `go` as their namespace.

## Authentication

REST MUST support unauthenticated reads for public resources unless the implementation is configured as private or headless-private.

REST SHOULD support authenticated requests through at least one documented mechanism:

- Browser session.
- API token.
- Application password.
- Signed request.
- External identity provider token.

Authenticated requests MUST be mapped to a user or service principal before capability checks.

## Authorization

Every endpoint MUST enforce capabilities server-side.

Rules:

- Public users MUST only receive published public content.
- Authenticated users MAY receive drafts, scheduled content, private metadata, or private settings only when capabilities allow it.
- Write operations MUST require explicit capabilities.
- Plugin routes MUST use the same authorization model.

## Pagination

List endpoints MUST support:

```text
page
per_page
```

Responses MUST include pagination metadata:

```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 120,
    "total_pages": 6
  }
}
```

Implementations MAY cap `per_page`. If capped, the effective value MUST be returned.

## Filtering And Sorting

List endpoints SHOULD support relevant filters:

- `status`
- `kind`
- `author`
- `taxonomy`
- `search`
- `locale`
- `after`
- `before`

Sorting SHOULD support:

- `created_at`
- `updated_at`
- `published_at`
- `title`
- `slug`

Invalid filters MUST return a validation error rather than being silently ignored, unless a route explicitly marks them as advisory.

## Error Shape

Errors MUST use this shape:

```json
{
  "error": {
    "code": "validation_error",
    "message": "The request is invalid.",
    "status": 400,
    "details": {},
    "request_id": "req_123"
  }
}
```

Required fields:

- `code`
- `message`
- `status`

Optional fields:

- `details`
- `request_id`

Stable error codes SHOULD include:

- `not_found`
- `validation_error`
- `unauthorized`
- `forbidden`
- `conflict`
- `rate_limited`
- `unsupported_media_type`
- `payload_too_large`
- `internal_error`

## Resource Envelope

Single-resource responses SHOULD use:

```json
{
  "data": {}
}
```

List responses MUST use:

```json
{
  "data": [],
  "pagination": {}
}
```

Implementations MAY include `links` and `meta`.

## Content Resource Fields

Post and page resources MUST include:

- `id`
- `kind`
- `status`
- `slug`
- `title`
- `content`
- `excerpt`
- `author_id`
- `featured_media_id`
- `taxonomy_ids`
- `created_at`
- `updated_at`
- `published_at`
- `links`

Private fields MUST only be returned when authorized.

## Locale Behavior

REST endpoints MUST support locale selection through:

- `locale` query parameter, or
- `Accept-Language` header, or
- an implementation-documented equivalent.

If localized content is requested but missing, fallback behavior MUST be documented and stable.

## Media Upload

Media upload endpoints MUST:

- Validate size.
- Validate type.
- Reject unsafe filenames.
- Return a media resource on success.
- Return validation errors on invalid files.

Upload endpoints MAY use multipart, direct upload tokens, or another documented mechanism.

## Caching

Public GET endpoints SHOULD include cache headers when safe.

Responses that depend on identity, draft content, private settings, or capability checks MUST NOT be cached as public responses.

## Conformance Notes

Conformance tests SHOULD verify:

- Discovery exists.
- Required routes exist.
- Pagination metadata is stable.
- Unauthorized draft access is blocked.
- Error shape is stable.
- Unknown JSON fields do not break clients.
- Plugin namespaces do not collide with reserved namespaces.
