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
which is for role-gated SSR redirects served by the Go API.
This plan does not change SSR redirects.

**Why SiteConfig and not Page:**
The existing `GET /api/site-data` and `POST /api/site-data` endpoints
already read and write arbitrary files in the content store via `?location=`.
`redirects.toml` is configuration data, not a Hugo content page,
so the SiteConfig path is the right fit.
The Page path is ruled out because `(*Page).ToTOML()` wraps output
in `+++ ... +++` Hugo frontmatter delimiters — wrong structure for `redirects.toml`.

**Why not `SiteParamsBase`:**
`SiteParamsBase` exists to surface the scheduling UI.
We don't need scheduling for redirects
(redirect rules take effect at the next Netlify deploy, not at a wall-clock time),
so the view is built as a simple dedicated component
that calls `getSiteData`/`postSiteData` directly,
the same way `ViewFrontpageEditor.vue` does.

---

## PR 1 — Backend: TOML serialization in `PublishSiteConfig`

Goal: make the site-data publish path write TOML when the file location ends in `.toml`.

- In `internal/almsvc/service-site-data.go`,
  extract a helper `marshalForLoc(loc string, data db.Map) ([]byte, error)`
  that returns `toml.Marshal(data)` when `strings.HasSuffix(loc, ".toml")`
  and `json.MarshalIndent(data, "", "  ")` otherwise.
  Replace the `json.MarshalIndent` call in `PublishSiteConfig` with this helper.

- Add `"redirects.toml"` to `MessageForLoc` in `internal/almsvc/site-data.go`:
  ```go
  "redirects.toml": "Updating redirect rules",
  ```

- Add a unit test in `internal/almsvc/` that calls `marshalForLoc` with a `.toml`
  location and verifies the output parses back to the same data.
  Confirm existing JSON locations are unaffected.

- Run `go test ./...`.

---

## PR 2 — Frontend: Redirect Manager page

Goal: an admin page for listing, editing, adding, and removing redirect rules,
built on the existing `SiteParamsBase` component.

### New route

In `src/plugins/router.js`, add:

```js
{
  path: "/admin/redirects",
  name: "redirects",
  component: load(() => import("@/components/ViewRedirects.vue")),
  meta: { requiresAuth: isSpotlightPAUser },
},
```

### New view: `src/components/ViewRedirects.vue`

Write in TypeScript (`<script setup lang="ts">`).
Fetch and save with `getSiteData`/`postSiteData` and `?location=redirects.toml`,
the same way `ViewFrontpageEditor.vue` calls those endpoints directly
without going through `SiteParamsBase`.
Use `makeState` for loading/error state.

The data shape returned by the API is `{ configs: [{ data: { redirects: [...] } }] }`.
Read `configs[0].data.redirects` on load;
on save, POST `{ configs: [{ schedule_for: <now>, data: { redirects: rules } }] }`.

```ts
interface RedirectRule {
  from: string;
  to: string;
  status: number;
  force: boolean;
}
```

UI (Bulma):
- Breadcrumbs: Admin → Redirect Manager.
- A table with columns: From, To, Status, Force, Remove.
- Each row has editable `<input>` fields.
  Status is a `<select>` with options 301, 302, 200.
  Force is a checkbox.
- Rows can be reordered; Netlify applies rules in order.
  Add up/down buttons, or drag-and-drop if straightforward.
- "Add redirect" button appends `{ from: "/", to: "/", status: 301, force: true }`.
- "Remove" button splices the row.

### Link in admin nav

In `src/components/ViewAdmin.vue`,
add a `<LinkRoute>` to the `redirects` route
near the existing "Sitewide Settings" link.

---

## Notes for the implementer

- `BurntSushi/toml` is already a direct dependency (used by `internal/db/page.go`);
  no `go.mod` changes needed.
  Follow the same `toml.NewEncoder` pattern from `(*Page).ToTOML()`.
- The TOML and JSON paths in `marshalForLoc` should both live in
  `service-site-data.go` to keep the change self-contained.
- Netlify enforces redirect rule order;
  the row order in the table is the order written to the file.
- The existing `redirect` table and `/ssr/redirect/{slug}` routes are unrelated
  and must not be touched.
- No new API endpoints are needed.
  `getSiteData` / `postSiteData` in `src/api/client.ts` already exist.
