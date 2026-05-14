# Plan: Migrate from Netlify JWT auth to scs-managed sessions

## Goal

Today, every authenticated request to Almanack carries a Netlify Identity JWT — either in the `Authorization: Bearer …` header (API calls from the SPA) or in the `nf_jwt` cookie (SSR routes proxied by Netlify's role-gated redirects). The Go server decodes that JWT on every request to learn the user's email and roles.

We want to keep Netlify Identity as the *login* provider (it owns signup, password reset, email verification, and the GoTrue user database), but stop treating its JWT as the request-time credential. Instead, after a successful Netlify login, the SPA hands the JWT to our server exactly once. The server validates it, looks up roles, and creates a server-side session managed by [`github.com/alexedwards/scs/v2`](https://github.com/alexedwards/scs). All subsequent requests authenticate via a normal opaque session cookie.

Benefits:
- Roles and other claims live in our database, not in a JWT the user could be holding stale.
- We can revoke a session immediately.
- Same auth mechanism works for API (`/api/*`), SSR (`/ssr/*`), and background routes — no more cookie-vs-header split.
- We can drop the Netlify role-conditioned redirects in `netlify.toml`; authorization is enforced in Go.
- Easier to add non-Netlify identity providers later.

## Current state (for reference)

- `internal/services/netlifyid/` decodes the Netlify JWT.
  - `AuthFromHeader` reads the lambda client context (`netlify` custom field — base64 JSON, already verified by Netlify).
  - `AuthFromCookie` reads `nf_jwt` and **does not verify the signature** (it trusts Netlify's CDN redirect rules in `netlify.toml`).
  - `HasRole` reads the `*JWT` stashed in the request context.
- `internal/almapp/router.go` wires three middleware stacks: `authHeaderMiddleware` (API), `authCookieMiddleware` (SSR), and `hasRoleMiddleware(role)`.
- `internal/almapp/routes-public.go:postIdentityHook` handles the `signup` webhook from Netlify Identity: validates the HMAC via `internal/services/jwthook`, then enriches the user with roles from `db.GetRolesForEmail`.
- The SPA (`src/api/auth.js`) wraps `netlify-identity-widget`, pulls a JWT via `user.jwt()` on demand, and attaches `Authorization: Bearer <jwt>` to every API call.
- `netlify.toml` has `[[redirects]]` blocks that gate `/ssr/*` on Netlify Identity roles (`admin`, `Spotlight PA`, `editor`) before the request ever reaches the function.

## Target architecture

```
  Browser ──login──> Netlify Identity (GoTrue) ──JWT──> Browser
     │
     │ POST /api/auth/session   { Authorization: Bearer <netlify JWT> }   (one time)
     ▼
  Almanack Go server
     ├── validates the Netlify JWT (signature + exp + iss)
     ├── loads/refreshes roles from Postgres (db.GetRolesForEmail)
     ├── scs.SessionManager.RenewToken + Put("user", User{…})
     └── responds 200 with Set-Cookie: almanack_session=<opaque>; HttpOnly; Secure; SameSite=Lax

  Browser ──every subsequent request──> sends almanack_session cookie
     ▼
  Almanack Go server
     ├── scs LoadAndSave middleware materializes session data into ctx
     ├── authMiddleware pulls the User struct out, attaches to almlog ctx
     └── hasRoleMiddleware checks roles from session, not from a JWT
```

Session storage: scs's pgxstore against the existing pgx pool. Sessions table lives in our Postgres database — added via a new tern migration.

## Work plan

### 1. Dependencies

- `go get github.com/alexedwards/scs/v2`
- `go get github.com/alexedwards/scs/pgxstore` (uses `*pgxpool.Pool` — already in `internal/db/handle.go`).
- Verify no version conflict with our pgx v5.

### 2. Database

Add `sql/schema/041_sessions.sql` (tern migration):

```sql
CREATE TABLE sessions (
    token text PRIMARY KEY,
    data bytea NOT NULL,
    expiry timestamptz NOT NULL
);
CREATE INDEX sessions_expiry_idx ON sessions (expiry);

---- create above / drop below ----

DROP TABLE sessions;
```

(Schema matches what scs's pgxstore expects.) The sessions table is intentionally *not* managed by sqlc — scs owns it.

### 3. New `internal/services/session` package

Replaces (and ultimately deletes) `internal/services/netlifyid` for request-time use. Responsibilities:

- Owns a configured `*scs.SessionManager`.
  - Store: `pgxstore.New(pool)`.
  - Cookie name: `almanack_session`.
  - `Lifetime: 30 * 24 * time.Hour`, `IdleTimeout: 7 * 24 * time.Hour` (tweakable).
  - `Cookie.HttpOnly = true`, `Cookie.Secure = true` in lambda mode, `SameSite = http.SameSiteLaxMode`, `Path = "/"`.
- Exposes:
  - `Middleware(http.Handler) http.Handler` — wraps `LoadAndSave`.
  - `Login(ctx, User) error` — calls `RenewToken` then `Put`.
  - `Logout(ctx) error` — calls `Destroy`.
  - `FromContext(ctx) *User` — typed accessor.
  - `HasRole(ctx, role) error` — replaces `netlifyid.HasRole` at the request level.
- `User` struct lives here:
  ```go
  type User struct {
      Email    string
      FullName string
      Roles    []string
      // possibly: Provider, NetlifySub, LoggedInAt time.Time
  }
  func (u *User) HasRole(role string) bool { /* same logic as JWT.HasRole, admin override */ }
  ```
- `gob.Register(User{})` in `init()` so scs can serialize it.

### 4. New login/logout endpoints

In `internal/almapp/routes-auth.go` (new file):

- `POST /api/auth/session` — accepts `Authorization: Bearer <netlify JWT>`, exchanges it for a session.
  1. Pull the bearer token.
  2. Verify it:
     - Production (lambda): if Netlify already verified and passed us the decoded identity via `lambdacontext.ClientContext.Custom["netlify"]`, prefer that path (no key needed).
     - Otherwise: verify the JWT signature against Netlify Identity's JWT secret. Add `ALMANACK_NETLIFY_JWT_SECRET` to `almsvc` flags. Use `internal/services/jwthook`'s primitives or a small wrapper (HS256, check `exp`, `iss`).
  3. Pull email + sub + full_name from the verified payload.
  4. Call `db.GetRolesForEmail` to compute roles fresh.
  5. Build `session.User{…}`, call `session.Login(ctx, user)`.
  6. Respond 204.
- `POST /api/auth/logout` — calls `session.Logout(ctx)`, responds 204.
- `GET /api/user-info` and `GET /ssr/user-info` keep working but now read from `session.FromContext`.

### 5. Middleware swap in `internal/almapp/router.go`

- Mount the scs middleware in `baseMW` (it must be present on every route that touches sessions, including the public ones, so the Set-Cookie path round-trips correctly on login). It is cheap when no cookie is present.
- Replace `authHeaderMiddleware` and `authCookieMiddleware` with a single `requireSessionMiddleware` that errors 401 if `session.FromContext(ctx) == nil`.
- Apply `requireSessionMiddleware` to both `/api/*` (formerly `authMW`) and `/ssr/*` (formerly `ssrMW`).
- `hasRoleMiddleware` continues to exist but reads from `session.FromContext` instead of `netlifyid.FromContext`.
- Keep `POST /api/identity-hook` and `POST /api/auth/session` (and healthchecks etc.) in the public/unauthenticated stack.

### 6. `internal/almapp/httputil.go` and friends

- `logErr` currently uses `netlifyid.FromContext` for Sentry tags. Switch to `session.FromContext`.
- Anywhere else in `internal/almapp/` that imports `netlifyid` for the *user* — port to `session`.
- The identity-hook handler (`postIdentityHook`) continues to use `netlifyid.User` as the JSON shape that GoTrue sends in the webhook payload — that's not request auth, it's webhook payload. Leave it (or move just the `User` struct type to a smaller package like `internal/services/gotrue` to make the layering crisp).

### 7. Frontend changes (`src/api/auth.js`, callers)

- After Netlify Identity emits `login` (or on `init` when a session is already present):
  1. `let token = await user.jwt();`
  2. `await fetch("/api/auth/session", { method: "POST", headers: { Authorization: "Bearer " + token }, credentials: "include" });`
  3. From this point forward, every API call uses `credentials: "include"` (or `"same-origin"`) and **omits** the `Authorization` header. The browser supplies the `almanack_session` cookie automatically.
- `methods.headers()` no longer needs to return an `Authorization` header. It can disappear or return `{}`.
- On `logout`: call `POST /api/auth/logout` *before* `netlifyIdentity.logout()` so the server-side session is destroyed.
- If the SPA gets a 401 from any API call, force a re-login (Netlify JWT may have expired between sessions; user goes through the widget again).
- Keep using `netlify-identity-widget` for the UI flows.

### 8. `netlify.toml`

- The `[[redirects]]` rules that gate `/ssr/*` on Netlify Identity roles are no longer load-bearing — our middleware enforces authorization. We can simplify to a single rule that always proxies `/ssr/*` to the function. (Keep the role-conditioned ones for a deploy or two during the transition, then delete.)

### 9. Roles refresh

Roles can change in the DB without the user logging out. Options:

- Cheap: refresh roles every time `session.HasRole` is called if `LastRoleRefresh > 5 min`. Implement inside `session.HasRole` by calling `db.GetRolesForEmail` and `Put`ing the updated user back. Avoids stale-role surprises with negligible DB load.
- Heavier: re-verify on every request. Skip — defeats the point of a session.

Go with the cheap option, gated by a `roleTTL` knob.

### 10. Tests

- Unit-test `session.User.HasRole` (mirrors existing `netlifyid.JWT.HasRole` semantics — admin override).
- Integration test (in `pkg/integration/`, gated by `ALMANACK_POSTGRES`):
  - Run migrations.
  - POST to `/api/auth/session` with a freshly minted HS256 JWT signed with a test secret; assert `Set-Cookie: almanack_session=…`.
  - GET `/api/user-info` with that cookie; assert 200 and correct payload.
  - POST `/api/auth/logout`; assert subsequent `/api/user-info` returns 401.
  - Assert `hasRoleMiddleware` 401s for missing roles.
- Keep `MockAuthService` semantics in some form for local dev — likely a `session.LoginDevUser(ctx)` helper or a flag-gated handler that grants a mock session without a JWT.

### 11. Rollout

1. Land migration 041 and the new `session` package behind a feature flag (`-session-auth` boolean on `almsvc`); default off.
2. When off: existing `netlifyid` middleware runs (current behavior).
3. When on: new middleware runs; `/api/auth/session` is live.
4. Deploy with flag off, run migration, verify sessions table exists and is empty.
5. Flip flag on in a preview deploy; smoke test SSR + API.
6. Flip flag on in production.
7. Ship the frontend change that calls `/api/auth/session`.
8. After a week of stability: delete `internal/services/netlifyid`'s request-time code paths (`AuthFromHeader`, `AuthFromCookie`, `HasRole`, `MockAuthService`, `FromCookie`, `FromLambdaContext`), keep only the `User`/`AppMetadata`/`UserMetadata` JSON structs that the identity-hook webhook still uses, or move them to `internal/services/gotrue`.
9. Remove the role-conditioned `[[redirects]]` from `netlify.toml`.

### 12. Open questions

- **JWT verification key in lambda**: can we trust `lambdacontext.Custom["netlify"]` as authoritative (Netlify already verified it before invoking the function), or do we still want to double-check the bearer token's signature ourselves? Leaning yes-trust-it in lambda, verify-the-signature elsewhere.
- **Cookie domain across `*.spotlightpa.org`**: do we need `Cookie.Domain` set? Probably not — `almanack.data.spotlightpa.org` only.
- **Background functions** (`/api-background/*`) currently bypass `authMW`. Keep them as-is (they're invoked by Netlify cron, not users).
- **CSRF**: with `SameSite=Lax` and the existing `http.CrossOriginProtection` already wired into `baseMW`, we should be covered, but worth a once-over for the new login/logout endpoints (they should require same-origin POSTs).
- **scs and lambda cold starts**: scs's pgxstore opens no long-lived goroutines by default (we'll pass `WithCleanupInterval(0)` and instead clean up via a periodic background cron) — confirm before shipping.

## File-by-file summary of changes

| Path | Change |
| --- | --- |
| `go.mod` / `go.sum` | + scs/v2, scs/pgxstore |
| `sql/schema/041_sessions.sql` | new migration |
| `internal/services/session/` | new package: manager, User, middleware, Login/Logout/HasRole |
| `internal/services/netlifyid/` | shrink: keep JSON structs for the webhook payload; delete request-time auth code after rollout |
| `internal/almsvc/flags.go` | + `-netlify-jwt-secret` flag, + session manager construction |
| `internal/almapp/cli.go` | replace `app.auth = netlifyid.NewService(...)` with `app.session = session.New(...)` |
| `internal/almapp/router.go` | mount scs middleware in `baseMW`; swap auth middleware; add `/api/auth/session` and `/api/auth/logout` |
| `internal/almapp/routes-auth.go` | new: session login + logout handlers |
| `internal/almapp/routes-editor.go` | `userInfo` reads from `session.FromContext` |
| `internal/almapp/routes-public.go` | unchanged except imports |
| `internal/almapp/httputil.go` | `logErr` Sentry tags read from session; middleware helpers updated |
| `src/api/auth.js` | exchange JWT for session on login; drop `Authorization` header; call logout endpoint |
| `netlify.toml` | simplify `/ssr/*` redirects after rollout |
| `pkg/integration/` | new auth flow tests |
