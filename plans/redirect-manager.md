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

**Why Page and not SiteConfig:**
The Page infrastructure already handles reading and writing files
of multiple formats in the content store.
`FromMD` already auto-detects format on read (tries `+++TOML+++`, falls back to JSON).
The write side currently has two separate methods (`ToTOML`, `ToJSON`)
and two separate service paths (`PublishPage`, `PublishJSONPage`)
that callers choose between by convention.
This plan unifies serialization into a single `(*Page).Serialize()` method
that dispatches on `FilePath` extension,
and collapses `PublishPage`/`PublishJSONPage` into one path.
`redirects.toml` then becomes a plain Page like any other.

---

## PR 1 — Backend: extension-based serialization on `Page`

Goal: teach `Page` to serialize itself correctly based on its file extension,
and unify the two publish paths.

### Add `(*Page).Serialize() ([]byte, error)` in `internal/db/page.go`

Dispatches on `page.FilePath`:

- **`.md`** — existing `ToTOML()` logic (Hugo `+++TOML+++` frontmatter + body).
- **`.json`** — existing `ToJSON()` logic (plain JSON, optional body).
- **`.toml`** — new: encode `page.Frontmatter` as pure TOML with no delimiters,
  using `toml.NewEncoder` (same package already imported).
  `page.Body` is ignored for `.toml` files.
- Unknown extension: return an error.

Leave `ToTOML()` and `ToJSON()` in place for now;
they will be inlined into `Serialize()` or removed in a follow-up.

### Update `PublishPage` in `internal/almsvc/service-page.go`

Replace the `page.ToTOML()` call with `page.Serialize()`.
Remove `PublishJSONPage`;
its callers (YouTube pages and the `postPageJSON` handler in `routes-spotlightpa.go`)
should switch to `PublishPage`.
Verify that the YouTube `FilePath` values end in `.json`
so they will serialize correctly through `Serialize()`.

Also replace the two `ToTOML()` calls in `RefreshPageContents`
(used to diff before/after) with `Serialize()`.

### Tests

- Unit tests for `Serialize()` covering `.md`, `.json`, `.toml`, and unknown extension.
- Confirm the `.toml` output round-trips through `toml.Decode` back to the same map.
- Run `go test ./...`.

---

## PR 2 — Backend: `redirects.toml` as a Page

Goal: expose `redirects.toml` through the standard page API.

Create the row in the database via a one-time migration or seed:

```sql
INSERT INTO page (file_path, source_type, source_id, frontmatter)
VALUES (
  'redirects.toml',
  'redirects',
  'redirects.toml',
  '{"redirects": []}'
)
ON CONFLICT DO NOTHING;
```

This lets the existing `GET /api/page` and `POST /api/page` handlers
manage `redirects.toml` with no new endpoints.
The page is loaded from GitHub on first use via `PageLoadFromContentStore`
if the frontmatter is empty.

---

## PR 3 — Frontend: Redirect Manager page

Goal: a dedicated admin view for editing redirect rules.

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
Fetch and save using the existing `GET /api/page` and `POST /api/page` endpoints
with the `redirects.toml` page ID.
The `frontmatter.redirects` array is the working data.

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
- "Save" button POSTs the updated page.
- "Revert" button re-fetches.
- Standard `SpinnerProgress` and `ErrorReloader` components.

### Link in admin nav

In `src/components/ViewAdmin.vue`,
add a `<LinkRoute>` to the `redirects` route
near the existing "Sitewide Settings" link.

---

## Notes for the implementer

- `BurntSushi/toml` is already a direct dependency (used by `internal/db/page.go`);
  no `go.mod` changes needed.
- For `.toml` pages, `Body` is meaningless; `Serialize()` should ignore it.
  Consider whether `FromMD` should also gain a `.toml` fast-path
  that calls `toml.Decode` directly instead of looking for `+++` delimiters.
- The existing `redirect` table and `/ssr/redirect/{slug}` routes are unrelated
  and must not be touched.
- Netlify enforces redirect rule order;
  the row order in the table is the order written to the file.
