# 07. Capabilities Contract

This document defines GoCMS authorization capabilities.

## Purpose

Capabilities are granular permissions. Roles are named sets of capabilities. Users receive permissions through roles, direct grants, groups, or an implementation-defined policy engine.

Compatibility depends on capability checks, not role names alone.

## Principles

- Every protected action MUST map to at least one capability.
- Admin UI hiding MUST NOT replace server-side checks.
- REST endpoints MUST enforce capabilities.
- Plugin and theme actions MUST use the same capability model.
- Capability IDs MUST be stable.
- Plugin-defined capabilities SHOULD be prefixed by plugin ID.

## Capability Identifier Format

Recommended format:

```text
domain.action
domain.resource.action
```

Examples:

```text
content.create
content.edit
content.edit_own
content.edit_others
content.publish
content.delete
media.upload
media.delete
settings.manage
themes.activate
plugins.activate
users.manage
roles.manage
rest.access_private
```

## Required Core Capabilities

Implementations MUST define or map equivalents for:

### Admin

- `admin.access`

### Content

- `content.create`
- `content.read_private`
- `content.edit`
- `content.edit_own`
- `content.edit_others`
- `content.publish`
- `content.schedule`
- `content.archive`
- `content.delete`
- `content.restore`
- `content.manage_revisions`

### Media

- `media.upload`
- `media.edit`
- `media.delete`
- `media.read_private`

### Taxonomies

- `taxonomies.manage`
- `taxonomies.assign`

### Menus

- `menus.manage`

### Themes

- `themes.view`
- `themes.activate`
- `themes.manage_settings`

### Plugins

- `plugins.view`
- `plugins.install`
- `plugins.activate`
- `plugins.deactivate`
- `plugins.uninstall`
- `plugins.manage_settings`

### Users And Roles

- `users.view`
- `users.create`
- `users.edit`
- `users.delete`
- `roles.view`
- `roles.manage`

### Settings

- `settings.view`
- `settings.manage`

### API

- `rest.access`
- `rest.access_private`
- `rest.write`

## Roles

Implementations SHOULD provide default roles, but role names are not normative.

Recommended roles:

- Administrator: all core capabilities.
- Editor: content and media management without system settings.
- Author: own content creation and editing, upload if allowed.
- Contributor: draft creation without publish.
- Viewer: authenticated read-only access where needed.

Roles MUST be editable only by principals with role-management capabilities.

## Ownership Rules

Where ownership matters, the authorization model MUST distinguish:

- Own content.
- Others' content.
- System resources with no owner.
- Plugin-owned resources.
- Private content.

Example:

- A user with `content.edit_own` MAY edit their own draft.
- The same user MUST NOT edit another user's draft without `content.edit_others`.

## REST Enforcement

REST endpoints MUST check capabilities for:

- Reading private resources.
- Creating resources.
- Updating resources.
- Deleting resources.
- Publishing or scheduling content.
- Uploading media.
- Managing settings.
- Managing plugins and themes.

REST error behavior:

- Missing authentication SHOULD return `401`.
- Authenticated but unauthorized access SHOULD return `403`.
- Public callers requesting hidden resources MAY receive `404` to avoid revealing existence.

## Admin Enforcement

Admin screens MUST:

- Hide inaccessible menu items.
- Block direct URL access without capability.
- Block submitted actions without capability.
- Show clear authorization errors.

## Plugin-Defined Capabilities

Plugins MAY define capabilities in their manifest.

Rules:

- Capability IDs SHOULD be prefixed by plugin ID.
- Default grants SHOULD be explicit.
- Plugin activation MUST NOT automatically grant high-risk capabilities to low-privilege roles unless documented.
- Plugin deactivation MUST prevent its capabilities from authorizing active behavior.

## Capability Checks In Hooks

Hooks MUST NOT bypass authorization. If a hook changes a query, response, or render payload, the engine MUST apply final capability checks before output.

## Auditing

High-risk capability changes MUST be audit logged:

- Role creation.
- Role deletion.
- Capability grant.
- Capability revoke.
- User role assignment.
- User role removal.

## Conformance Notes

Conformance tests SHOULD verify:

- Required capabilities exist or have documented equivalents.
- Admin direct access is blocked without capability.
- REST write access is blocked without capability.
- Own vs others content checks work.
- Plugin-defined capabilities are enforced.
- Capability changes are audit logged where audit support is enabled.
