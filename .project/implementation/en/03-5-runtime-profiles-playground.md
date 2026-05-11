# 03.5. Runtime Profiles And Playground Boundary

This pass fixes the runtime boundary before GraphQL work begins. GoCMS must be able to run as different binary profiles without mixing admin fixtures, mutable site content, playground demo data, public rendering, and production storage.

## Why This Pass Exists

Pass 3 proved the admin GUI. Before Pass 4, the system needs a clear answer to:

- Which binary profile is running?
- Where does admin UI data come from?
- Where does site content data live?
- Can playground users run a complete isolated CMS experience without server persistence?
- How does a demo site initialize from an external compatibility source without overwriting local edits?

GraphQL must build on this boundary instead of becoming another place that assumes one storage mode.

## Runtime Profiles

GoCMS binaries should be assembled from explicit runtime profiles:

- `headless`: REST and future GraphQL surfaces, no public rendering requirement and no production admin surface.
- `admin`: `/go-admin`, `/go-login`, admin workflows, configured server-side storage, fixtures, or site package management.
- `playground`: isolated full CMS sandbox inspired by WordPress Playground: admin, public preview/rendering, import/export, browser-local or ephemeral content storage, and no backend persistence.
- `full`: site rendering, admin, REST, plugins, durable production storage.
- `conformance`: minimal runtime for contract and fixture checks.

A site-specific binary, such as a BrandOSS demo binary, may mount GoCMS as a feature in `playground` profile. The same domain can then serve the public preview and `/go-admin` without requiring a second process or port.

`playground` is not a reduced admin profile. It is a safe isolated runtime where unauthenticated visitors may receive admin-level capabilities because the environment is disposable, browser-local, or explicitly sandboxed. It should eventually support the same core CMS experience as `full`, but with demo auth, browser-local storage, snapshot import/export, blueprint-style startup configuration, and embeddable preview use cases.

## Storage Profiles

Storage profile must be separate from runtime profile:

- `browser-indexeddb`: playground-only mutable content storage.
- `memory`: tests or disposable local runs.
- `json-fixtures`: readonly admin/demo fixtures embedded into the binary.
- `sqlite`: local durable development or small deployments.
- `mysql`: production adapter candidate.
- `postgres`: production adapter candidate.

The binary may embed fixtures and schemas, but it must not pretend that mutable content can be safely written into the binary itself.

## Admin Fixtures Versus Site Content

Admin fixtures are not site content.

Admin fixtures include:

- Admin navigation labels.
- Dashboard text.
- Empty states.
- Form labels and helper text.
- Settings screen copy.
- Demo notices.
- UI component data required by the admin shell.

Admin fixtures should be pure JSON and may be embedded into the binary. They are readonly at runtime.

Site content includes:

- Posts.
- Pages.
- Custom content types.
- Taxonomies and terms.
- Authors/users public projections.
- Media metadata.
- Menus.
- Settings that affect public delivery.

In `playground`, site content starts empty or from an explicit blueprint/source import and is populated in browser-local or sandbox storage only.

## Playground Persistence Rule

The `playground` profile is stateless on the server:

- The backend does not persist imported content.
- The binary does not mutate itself.
- The browser owns mutable playground data.
- Reloading `/go-admin` or the public preview on the same domain reads from browser storage.
- Clearing browser storage removes playground content.

The preferred browser storage is IndexedDB. It should contain:

- A compatibility REST snapshot store.
- A media metadata store.
- A media Blob store for user-uploaded files.
- A small settings store for source/import metadata and snapshot version.

## One-Time Source Bootstrap

The query parameter is only an initializer:

```text
/?gocms=example.com
```

Bootstrap policy:

```text
if browser content storage has data:
    use browser storage
    do not call the external source
else if a source is known from the query parameter:
    import once from the external compatibility REST source
    save the normalized snapshot into IndexedDB
else:
    show the empty playground admin state and import prompts
```

The admin may later open as:

```text
/go-admin
```

without the `gocms` query parameter. It must still load existing browser-local content.

No automatic refresh may overwrite existing browser content. Refresh from source must be an explicit user action with an overwrite warning.

## First Playground Scenario

The first implementation may target one exact demo source site. It should:

- Use simple demo login: `admin` / `admin`.
- Expose the admin and public preview/rendering inside the same isolated sandbox.
- Import at most the latest 10 posts.
- Import related categories, tags, authors, and required media metadata.
- Import front page and blog page when available from the source.
- Preserve the 10-post limit in playground editing workflows.
- Show admin capabilities on real content.
- Allow users to export the edited snapshot to their device.
- Allow users to import a snapshot from their device and replace browser storage.

If a user needs a new post while the 10-post limit is reached, the UI should require deleting or trashing an existing post first.

## Compatibility Snapshot Shape

Playground JSON export should remain close to the compatibility REST surface. Prefer a route-keyed snapshot over a private ad-hoc DTO:

```json
{
  "snapshot_version": "gocms.playground.v1",
  "source": {
    "kind": "wp-json",
    "base_url": "https://example.com/wp-json",
    "imported_at": "2026-05-02T00:00:00Z"
  },
  "routes": {
    "/wp-json/wp/v2/posts": [],
    "/wp-json/wp/v2/pages": [],
    "/wp-json/wp/v2/categories": [],
    "/wp-json/wp/v2/tags": [],
    "/wp-json/wp/v2/media": []
  },
  "local": {
    "media_blobs": "excluded"
  }
}
```

GoCMS routes may expose equivalent data under `/go-json/go/v2/...`, but the playground snapshot should preserve the source compatibility shape where possible. This keeps import/export, conformance fixtures, and future migration tooling aligned.

## Media Blob Policy

User-uploaded media in playground is browser-local:

- Store uploaded files as Blob values in IndexedDB.
- Preview works only in the browser that has the Blob.
- JSON export does not include binary data or base64 payloads.
- JSON export includes metadata only.

Media metadata should include:

- Stable local media ID.
- Filename.
- MIME type.
- Width.
- Height.
- Byte size.
- Alt text.
- Caption.
- Created timestamp.
- Attachment relationships.
- Blob status.

If the Blob is missing after browser storage cleanup or after JSON import on another device, the UI should render a gray placeholder with the original filename and preserved dimensions/aspect ratio. Later media work can add thumbnail, medium, large, and original derivative metadata.

## Import And Export UX

Playground settings must eventually include:

- Blueprint-style startup configuration.
- Embeddable playground launch, for example through an iframe or one-click preview URL.
- Import from compatibility REST source.
- Import JSON from device.
- Export JSON to device.
- Refresh from source with explicit overwrite warning.
- Reset local playground storage.

The first version can expose only import/export/reset needed to demonstrate the admin safely.

## Future XML Import Plugin

XML import belongs to a plugin boundary, not the playground bootstrap itself.

Future plugin:

```text
Plugin: wordpress-xml-importer
Input: WXR XML export
Output: GoCMS compatibility entities and REST-compatible snapshots
```

The plugin should preserve posts, pages, custom content types, terms, authors, attachment metadata, postmeta where supported, parent/child relationships, and later comments if comments become part of the compatibility contract.

## Exit Checklist

- Runtime profile names and storage profile names are documented.
- Admin fixtures are separated from mutable site content.
- Playground is a full isolated CMS runtime target, not a reduced admin-only profile.
- Playground is browser-local or sandbox-backed and server-stateless.
- Playground can expose both admin and public preview/rendering safely inside the isolated environment.
- IndexedDB is selected for playground content and media Blob storage.
- One-time source bootstrap policy is documented.
- `?gocms=<source>` is initializer-only and never required after first import.
- JSON export excludes binary media and preserves metadata.
- Missing local Blobs have a defined placeholder behavior.
- Snapshot format follows the compatibility REST shape where possible.
- Future XML import is documented as a plugin boundary.
