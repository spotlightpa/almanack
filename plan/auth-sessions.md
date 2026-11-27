# Plan: Migrate from Netlify JWT auth to scs-managed sessions

## Goal

Today, every authenticated request to Almanack carries a Netlify Identity JWT — either in the `Authorization: Bearer …` header (API calls from the SPA) or in the `nf_jwt` cookie (SSR routes proxied by Netlify's role-gated redirects). The Go server decodes that JWT on every request to learn the user's email and roles.

We want to keep Netlify Identity as the *login* provider (it owns signup, password reset, email verification, and the GoTrue user database), but stop treating its JWT as the request-time credential. Instead, after a successful Netlify login, the SPA hands the JWT to our server exactly once. The server reads the identity (Netlify already verified it at the CDN — we trust the lambda-injected client context, same as today), looks up roles, and creates a server-side session managed by [`github.com/alexedwards/scs/v2`](https://github.com/alexedwards/scs). All subsequent requests authenticate via a normal opaque session cookie.

Benefits:
- Roles and other claims live in our database, not in a JWT the user could be holding stale.
- We can revoke a session immediately.
- Same auth mechanism works for API (`/api/*`), SSR (`/ssr/*`) — no more cookie-vs-header split.
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

## Package layout

- `internal/services/netlifyid/` **stays**. It still owns:
  - The JSON shape of Netlify's GoTrue user (`User`, `AppMetadata`, `UserMetadata`) used by the identity-hook webhook.
  - `FromLambdaContext` — the one-time exchange handler uses it to read the verified identity that Netlify injects into the lambda invocation.
  - `FromCookie` — likely still useful for the same one-time exchange when the SPA arrives at `/ssr/*` without a session yet (Netlify already validated the role at the edge before letting the request through).
  - Removing/shrinking later: the request-time middleware methods (`AuthFromHeader`, `AuthFromCookie`, `HasRole`, `MockAuthService`) move into the session layer.
- `internal/almuser/` **new**. Owns our application's notion of a logged-in user:
  - `type User struct { Email, FullName string; Roles []string; … }` with `HasRole(role) bool` (admin override, same semantics as today's `netlifyid.JWT.HasRole`).
  - Context get/put helpers (`FromContext`, `NewContext`) plus almlog enrichment (currently in `netlifyid/context.go`).
  - Constructor `FromNetlifyID(ctx, q, *netlifyid.JWT) (User, error)` that pulls roles from the DB via `db.GetRolesForEmail` and merges them with whatever Netlify already attached.
  - No HTTP or scs imports — this is the pure domain type. `internal/almapp` and `internal/services/session` both depend on it.
- `internal/services/session/` **new**. Owns scs configuration, the login/logout helpers, the request middleware, and the typed accessor for `*almuser.User` out of the scs session.

## Target architecture

```
  Browser ──login──> Netlify Identity (GoTrue) ──JWT──> Browser
     │
     │ POST /api/auth/session   { Authorization: Bearer <netlify JWT> }   (one time)
     ▼
  Almanack Go server (running on Netlify Functions)
     ├── Netlify already verified the JWT at the edge and base64-injected the
     │   identity payload into lambda client context. We read that via
     │   netlifyid.FromLambdaContext — no HMAC verification of our own.
     ├── db.GetRolesForEmail → build almuser.User
     ├── session.Login(ctx, user): RenewToken + Put("user", user)
     └── responds 204 with Set-Cookie: almanack_session=<opaque>; HttpOnly; Secure; SameSite=Lax

  Browser ──every subsequent request──> sends almanack_session cookie
     ▼
     ├── scs LoadAndSave middleware materializes session data into ctx
     ├── requireSessionMiddleware: 401 if no almuser.User in ctx
     └── hasRoleMiddleware checks almuser.User.HasRole, not a JWT
```

Session storage: scs's pgxstore against the existing pgx pool. Sessions table lives in our Postgres database — added via a new tern migration.

## Auth verification: we don't roll our own

In production, Netlify Functions receives the request **after** Netlify Identity has already validated the JWT signature and expiry. The decoded identity comes to us through `lambdacontext.ClientContext.Custom["netlify"]` (for header-auth API calls) or, for SSR routes, has been gated by the role-conditioned redirect at the CDN. That's what we trust today and what we'll keep trusting at the moment of session creation. No HS256 verification, no JWT secret flag, no extra dependency on a JWT library for the auth path.

`internal/services/jwthook` keeps doing HMAC verification — but only for the `POST /api/identity-hook` webhook payload, which is a totally separate concern (server-to-server, signed by Netlify with our webhook secret).

## Dev login

Locally, lambda client context is empty and there's no Netlify edge to gate cookies. Solve this with a dev-only endpoint mounted *only when `app.isLambda == false`*:

- `POST /api/auth/dev-login` — body picks an email (or just hardcodes `dev@spotlightpa.org`); the handler calls `db.GetRolesForEmail` (or grants `admin` outright in dev), builds an `almuser.User`, and `session.Login`s the request. Returns 204.
- The SPA shows a "dev login" button on the login page when `import.meta.env.DEV` is true (or when a `?dev` flag is present); the button hits `/api/auth/dev-login` and then proceeds as if Netlify Identity had logged the user in.
- In lambda mode, the handler is not registered at all (router check) — not just refused, so there's no chance of accidentally exposing it.
- **Belt and suspenders**: the handler function itself begins with `if app.isLambda { app.logErr(ctx, errors.New("dev-login invoked in production")); replyErr 404 }` so that even if a routing mistake exposes it, the call is rejected *and* fires a Sentry error we'll actually notice. The handler also unconditionally logs a warn-level Sentry breadcrumb on every use so we have a paper trail in dev too.

This replaces today's `MockAuthService`.

## Work plan

### 1. Dependencies

- `go get github.com/alexedwards/scs/v2`
- `go get github.com/alexedwards/scs/pgxstore` (uses `*pgxpool.Pool` — already in `internal/db/handle.go`).
- Verify no version conflict with pgx v5.

### 2. Database migration

`sql/schema/041_sessions.sql`:

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

Not managed by sqlc — scs owns the table.

### 3. `internal/almuser` (new package)

```go
package almuser

type User struct {
    Email    string
    FullName string
    Roles    []string
}

func (u *User) HasRole(role string) bool { /* admin override + match */ }
func FromContext(ctx context.Context) *User
func NewContext(ctx context.Context, u *User) context.Context  // also enriches almlog
func FromNetlifyID(ctx context.Context, q *db.Queries, j *netlifyid.JWT) (User, error)
```

`gob.Register(User{})` in `init()` so scs can serialize it.

### 4. `internal/services/session` (new package)

- Owns a configured `*scs.SessionManager` (cookie name `almanack_session`, lifetime ~30d, idle 7d, HttpOnly, Secure in lambda, SameSite=Lax, Path=/).
- Store: `pgxstore.NewWithCleanupInterval(pool, 0)` — disable the goroutine, rely on a periodic background cron (we already have `/api-background/cron`) to call `pgxstore.Cleanup`. Avoids goroutine-in-lambda hazards.
- Exposes:
  - `Middleware(http.Handler) http.Handler` — wraps `scs.LoadAndSave` and also lifts the stored `almuser.User` into ctx via `almuser.NewContext` if present.
  - `Login(ctx, almuser.User) error` — `RenewToken` then `Put("user", u)`.
  - `Logout(ctx) error` — `Destroy`.
  - `Touch(ctx, almuser.User) error` — `Put` without renewing, used by the role-TTL refresh.
- No JWT or HMAC code here.

### 5. New auth endpoints

`internal/almapp/routes-auth.go`:

- `POST /api/auth/session` (always mounted):
  1. `j, err := netlifyid.FromLambdaContext(r.Context())` — fail 401 if not present.
  2. `user, err := almuser.FromNetlifyID(ctx, app.svc.Queries, j)`.
  3. `app.session.Login(ctx, user)`.
  4. 204.
- `POST /api/auth/logout`: `app.session.Logout(ctx)`; 204.
- `POST /api/auth/dev-login` (only registered when `!app.isLambda`): first guards `if app.isLambda` and, if true, calls `app.logErr` with a distinct "dev-login invoked in production" error so Sentry alerts and returns 404. Otherwise: build a fake user (admin or DB-resolved roles), `Login`, 204.
- `GET /api/user-info` and `GET /ssr/user-info` read from `almuser.FromContext`.

### 6. Router rewiring (`internal/almapp/router.go`)

- Add `app.session.Middleware` to `baseMW` so every route (public or not) gets `LoadAndSave` semantics — needed for Set-Cookie to round-trip on login.
- Replace `authHeaderMiddleware` and `authCookieMiddleware` with a single `requireSessionMiddleware` (401 if no user in ctx).
- Apply `requireSessionMiddleware` to both `/api/*` and `/ssr/*` route groups.
- `hasRoleMiddleware` reads roles from `almuser.FromContext`.
- Keep public stack for healthchecks, identity-hook, and the new auth endpoints.

### 7. `httputil.go` and other consumers

- `logErr` Sentry tags switch from `netlifyid.FromContext` to `almuser.FromContext`.
- Anywhere in `internal/almapp/` that calls `netlifyid.FromContext` for user info — switch to `almuser`. (`netlifyid.FromContext` itself can stay for the one-time exchange handler that reads the lambda-injected JWT.)

### 8. `internal/services/netlifyid/`

Keep:
- `JWT`, `User`, `AppMetadata`, `UserMetadata`, `Identity` structs.
- `FromLambdaContext`, `FromCookie`.
- `jwt.go` helpers like `Email`, `Username`, `HasRole` on `*JWT` (still useful when constructing an `almuser.User`).

Remove (after rollout):
- `service.go`'s `AuthService` interface, `NetlifyAuth`, `NewService`.
- `mock.go` (replaced by dev-login endpoint).
- `context.go` (moves to `almuser`).

Nothing in `netlifyid/` should be imported by request-time middleware after the migration; only by the auth-exchange endpoint and the identity-hook webhook.

### 9. Frontend (`src/api/auth.js`, callers)

- On `netlifyIdentity.on("init")` and `on("login")`: after `user.jwt()` resolves, `POST /api/auth/session` with `Authorization: Bearer <jwt>` and `credentials: "include"`. From then on, all API calls use `credentials: "include"` and **do not** set `Authorization`.
- `methods.headers()` becomes a no-op (or is deleted; callers drop the `await auth.headers()` line).
- On `logout`: `POST /api/auth/logout` (cookie included), then `netlifyIdentity.logout()`.
- On a 401 from any API call: clear local state and re-open the Netlify Identity widget.
- Dev login button (vite `import.meta.env.DEV`): a `<button @click="devLogin">` calls `POST /api/auth/dev-login` and refreshes the user-info query. Not bundled into production builds.

### 10. `netlify.toml`

- The `/ssr/*` role-conditioned redirects are still useful during transition because they keep `nf_jwt` semantics alive while users have stale tabs open. Plan: leave them in place; after the rollout window, collapse to a single `/ssr/*` redirect with no role conditions and let Go enforce authz.

### 11. Role refresh

Roles can change in the DB without the user logging out. In `requireSessionMiddleware` (or in `hasRoleMiddleware`), if `almuser.User.LastRoleRefresh` is older than e.g. 5 minutes, call `db.GetRolesForEmail`, update the user, `session.Touch`. Cheap and avoids stale-role surprises.

### 12. Tests

- Unit-test `almuser.User.HasRole` (admin override).
- Integration test (`pkg/integration/`, gated by `ALMANACK_POSTGRES`):
  - Migrate.
  - Hit `/api/auth/dev-login` (dev mode). Assert `Set-Cookie: almanack_session=…`.
  - `GET /api/user-info` with that cookie → 200 with expected payload.
  - `POST /api/auth/logout` → subsequent `/api/user-info` returns 401.
  - `hasRoleMiddleware` 401s for missing roles, 200s when role is granted.

### 13. Rollout

1. Land migration 041 + new packages behind a feature flag (`-session-auth`, default off).
2. Off → today's `netlifyid` middleware runs.
3. On → new middleware runs; `/api/auth/session` and `/api/auth/logout` live.
4. Deploy with flag off, run migration, verify `sessions` table.
5. Flip on in preview deploy; smoke test SSR + API.
6. Flip on in production.
7. Ship the frontend change that calls `/api/auth/session`.
8. After a stability window: delete the unused request-time bits of `internal/services/netlifyid/` and collapse `netlify.toml` redirects.

## File-by-file summary

| Path | Change |
| --- | --- |
| `go.mod`/`go.sum` | + scs/v2, scs/pgxstore |
| `sql/schema/041_sessions.sql` | new migration |
| `internal/almuser/` | **new** package: `User`, `HasRole`, ctx helpers, `FromNetlifyID` |
| `internal/services/session/` | **new** package: scs config, `Middleware`, `Login`, `Logout` |
| `internal/services/netlifyid/` | stays; eventually shrinks to JWT structs + lambda/cookie readers |
| `internal/almsvc/flags.go` | + session manager construction (using existing pgx pool) |
| `internal/almapp/cli.go` | construct `app.session`; keep `app.auth` until the old paths are deleted |
| `internal/almapp/router.go` | mount session middleware in `baseMW`; swap auth/role middleware; register `/api/auth/*` |
| `internal/almapp/routes-auth.go` | **new**: session, logout, dev-login handlers |
| `internal/almapp/routes-editor.go` | `userInfo` reads from `almuser.FromContext` |
| `internal/almapp/httputil.go` | Sentry tags + middleware updated |
| `src/api/auth.js` | exchange JWT once; drop `Authorization` header; logout endpoint; dev-login button |
| `netlify.toml` | simplify `/ssr/*` redirects after rollout |
| `pkg/integration/` | new auth flow tests |

## Open questions

- **Cookie domain**: do we need `Cookie.Domain` set, or is host-only fine? Almanack only runs on `almanack.data.spotlightpa.org`. Default to host-only.
- **CSRF**: `SameSite=Lax` plus the existing `http.CrossOriginProtection` should cover us, but verify it also guards the new `/api/auth/*` POSTs.
- **scs cleanup in lambda**: confirm `pgxstore.NewWithCleanupInterval(pool, 0)` disables the goroutine; wire `pgxstore.Cleanup` into the `/api-background/cron` job.
- **Long-lived tabs**: when the scs cookie expires but Netlify still has a session, the 401 → widget re-login flow should transparently re-establish the scs session. Verify the SPA path on first 401.
