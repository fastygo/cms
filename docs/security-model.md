# Security Model

Related documentation: [CMS capabilities](./cms-capabilities.md) · [Authentication](./authentication.md) · [Audit logs](./audit-logs.md).

## Core Principles

GoCMS core is intentionally local-first and autonomous by default.

- No SMTP server is required.
- No external audit SaaS is required.
- No external identity provider is required.
- Future OIDC support must plug into an adapter boundary instead of replacing local authn internals.

## Password Storage

- Local passwords are stored as Argon2id hashes.
- Plaintext passwords are never persisted.
- Password updates track `PasswordUpdatedAt`, `MustChangePassword`, and `LastLoginAt` on the user record.

## Recovery Material

- Recovery codes are hashed at rest and shown once at generation time.
- Admin reset tokens are hashed at rest, time-bounded, and single-use.
- Core intentionally does not deliver reset links over email.

This design favors offline and single-binary installs where the operator controls the local environment.

## App Tokens

- App tokens are stored only as hashes.
- A prefix/ID is kept for lookup and revocation.
- Tokens may inherit the owning user's role capabilities or use an explicit capability subset.
- Revoked or expired tokens are rejected even if the raw token is presented later.

## Sessions

- Browser auth uses signed cookie sessions.
- Secure cookies are enabled automatically for production-like deployment profiles.
- Sessions include absolute-expiry metadata.
- Login attempts are tracked and can trigger temporary lockout.

## Audit And Redaction

Audit events record:

- actor
- action
- resource and resource ID
- status
- request metadata
- redacted details

Audit detail keys containing `password`, `token`, `secret`, `code`, or `hash` are redacted before persistence.

## Explicit Non-Goals

- Core does not implement SMTP reset delivery.
- Core does not implement OIDC login yet.
- Core does not provide remote trust delegation to third-party auth or logging vendors.
- Core does not bypass local capability checks for convenience paths.

## Operator Guidance

- Rotate seeded local credentials on first use.
- Move one-time recovery material into offline storage immediately.
- Use app tokens instead of static development bearer tokens for real API clients.
- Treat snapshot exports as sensitive backups because they include user/auth metadata needed for autonomous restore.
