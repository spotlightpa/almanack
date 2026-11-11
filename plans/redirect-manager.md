# Plan: Redirect Manager

Allow Spotlight PA editors to manage public URL redirects
through Almanack's admin UI,
instead of editing `redirects.toml` in the GitHub repo by hand.

The file being managed is `redirects.toml` in the GitHub content store (poor-richard).
It uses Netlify's redirect syntax:

```toml
[[redirects]]
from = "/newsletters/howwecare"
to = "/newsletters"
status = 301
force = true
```

Netlify reads this file at deploy time and applies the rules at the CDN edge.
This is distinct from the existing `redirect` table in PostgreSQL,
which is for role-gated SSR redirects served by the Go API
(e.g. `/ssr/redirect/filing-folder`).
This plan does not change SSR redirects.

---

## PR 1 — Backend: TOML parsing, storage, and API

Goal: teach the backend to read, write, and serve `redirects.toml`
as a structured list of redirect rules.

### Data model

Define a Go struct for a single redirect rule
in a new file `internal/almsvc/redirects.go`:

```go
type RedirectRule struct {
    From   string `toml:"from"   json:"from"`
    To     string `toml:"to"     json:"to"`
    Status int    `toml:"status" json:"status"`
    Force  bool   `toml:"force"  json:"force"`
}

type RedirectFile struct {
    Redirects []RedirectRule `toml:"redirects" json:"redirects"`
}
```

Add `github.com/BurntSushi/toml` to `go.mod`
(check if it's already a transitive dependency first with `go mod download -json`).

### Service methods

In `internal/almsvc/redirects.go`, add two methods on `Services`:

- `GetRedirects(ctx) ([]RedirectRule, error)`
  — calls `svc.ContentStore.GetFile(ctx, "redirects.toml")`,
  TOML-decodes it into `RedirectFile`, and returns the slice.
  Returns an empty slice (not an error) if the file is missing (404).

- `UpdateRedirects(ctx, rules []RedirectRule) error`
  — encodes a `RedirectFile{Redirects: rules}` back to TOML bytes,
  then calls `svc.ContentStore.UpdateFile(ctx, "Updating redirects.toml", "redirects.toml", bytes)`.
  Validation before encoding:
  - `From` must be non-empty and start with `/`.
  - `To` must be non-empty.
  - `Status` must be one of 200, 301, 302.
  Returns a 400-level `resperr` on validation failure.

### HTTP handlers

In `internal/almapp/routes-spotlightpa.go`, add two handlers
registered in `router.go` under `spotlightMW`:

```
GET  /api/redirects      → app.getRedirects
POST /api/redirects      → app.postRedirects
```

`getRedirects`: calls `svc.GetRedirects`, replies JSON `{"redirects": [...]}`.

`postRedirects`: reads JSON body `{"redirects": [...]}`,
calls `svc.UpdateRedirects`,
replies with the round-tripped rules as `{"redirects": [...]}`
(re-fetch after save so the response reflects what's in the file).

### Tests

- Unit tests for TOML round-trip in `internal/almsvc/redirects_test.go`.
- Unit test that validation rejects bad `From`/`Status` values.
- The mock `ContentStore` (already used by other tests) supports `GetFile`/`UpdateFile`;
  use it to test `GetRedirects` and `UpdateRedirects` without hitting GitHub.

---

## PR 2 — Frontend: Redirect Manager UI

Goal: a simple admin page for listing, editing, adding, and removing redirects.

### New route

In `src/plugins/router.js`, add under the Spotlight PA section:

```js
{
  path: "/admin/redirects",
  name: "redirects",
  component: load(() => import("@/components/ViewRedirects.vue")),
  meta: { requiresAuth: isSpotlightPAUser },
},
```

### New component: `src/components/ViewRedirects.vue`

Write this in TypeScript (`<script setup lang="ts">`).

State:
- `rules: Ref<RedirectRule[]>` — the working copy of the list.
- Load/save via `get(getRedirects)` / `post(postRedirects, { redirects: rules })`
  using `makeState` (same pattern as `SiteParamsBase.vue`).

UI (Bulma):
- Breadcrumbs: Admin → Redirect Manager.
- A table with columns: From, To, Status, Force, Actions.
- Each row is editable in-place (plain `<input>` fields bound to the rule object).
  Status is a `<select>` with options 301, 302, 200.
  Force is a checkbox.
- "Remove" button per row (splice from array).
- "Add redirect" button appends a blank rule
  (`{ from: "/", to: "/", status: 301, force: true }`).
- "Save" button POSTs the full list.
- "Revert" button re-fetches.
- Standard `SpinnerProgress` and `ErrorReloader` components.

Add API constants to `src/api/client.ts`:

```ts
export const getRedirects = "/api/redirects";
export const postRedirects = "/api/redirects";
```

### Link in admin nav

In `src/components/ViewAdmin.vue`,
add a `<LinkRoute>` entry pointing to the new `redirects` route,
near the existing "Sitewide Settings" link.

---

## Notes for the implementer

- The TOML library's `Encoder` preserves array-of-tables (`[[redirects]]`) correctly
  when encoding `RedirectFile`; verify this in the round-trip unit test.
- Netlify enforces redirect rule order;
  the UI table row order is the canonical order sent to the file.
- The existing `redirect` table and `/ssr/redirect/{slug}` routes are unrelated
  and must not be touched.
- `ContentStore.UpdateFile` is a no-op when the content hasn't changed
  (the GitHub client checks SHA before committing);
  this means saving an unchanged list is safe.
- `MessageForLoc` in `internal/almsvc/site-data.go` does not need to be updated
  because `UpdateRedirects` supplies its own commit message directly.
- Keep validation server-side only for now;
  the frontend should surface the error response string from the API.
