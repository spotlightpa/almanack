# Plan: Sub-site Sidebar Editors (Berks & State College)

> **Implementation order:** Phase 0 (consolidation) merges to master first as a
> standalone PR. Phase 1 (the new feature) builds on top of it.

## Phase 0: Consolidation (prerequisite, lands on master first)

### The Problem

Right now there are three parallel mechanisms for reading/writing site data from
the content store, each with its own dedicated URL and named endpoint constants:

| Frontend constants | Backend route | Handler | What it does |
|---|---|---|---|
| `getSidebar` / `saveSidebar` | `GET/POST /api/sidebar` | `siteDataGet(SidebarLoc)` | Main sidebar |
| `getSiteParams` / `postSiteParams` | `GET/POST /api/site-params` | `siteDataGet(SiteParamsLoc)` | Site-wide ad params |
| `getSiteData` / `postSiteData` | `GET/POST /api/site-data` | `getSiteData` / `postSiteData` (reads `?location=` query param) | Frontpage editors (Berks, State College, Homepage) |

The `/api/site-data` route already solves the general case: it accepts a
`?location=` query parameter and delegates to the same generic handlers. The
other two routes (`/api/sidebar`, `/api/site-params`) are just special-cased
wrappers around the exact same handler with a hard-coded loc. Every new loc we
add (e.g. the two new sidebar files) would normally mean a new pair of named
endpoint constants and a new pair of named routes — that's what the plan before
this revision proposed.

### The Solution

Consolidate everything through `/api/site-data?location=<loc>`. This means:

**Backend (`internal/almapp/router.go`):**

Remove the four special-cased routes:
```
GET/POST /api/sidebar
GET/POST /api/site-params
```
The `GET/POST /api/site-data` routes (which already exist and use the same
handler logic via `getSiteData`/`postSiteData`) absorb their traffic.

**Frontend (`src/api/endpoints.ts`):**

Remove the four now-redundant named constants:
```ts
getSidebar / saveSidebar
getSiteParams / postSiteParams
```
They are replaced by the existing `getSiteData` / `postSiteData` constants, used
with a `?location=` query param — exactly as `ViewFrontpageEditor.vue` already
does.

**Frontend callers:**

- `ViewSidebarItems.vue` — change `get(getSidebar)` →
  `get(getSiteData, { location: "data/sidebar.json" })` and
  `post(saveSidebar, ...)` → `post(postSiteData + "?location=data/sidebar.json", ...)`.
- `ViewSiteParams.vue` — same treatment with `"config/_default/params.json"`.

Follow the same URL-construction pattern already used in `ViewFrontpageEditor.vue`
(`getSiteData + "?location=" + dataFile`). Check whether `client.ts`'s `post()`
helper accepts a `params` arg for query params; if not, use string concatenation
as the existing code does.

**Backend (`internal/almsvc/site-data.go`):**

No changes needed. The loc constants (`SidebarLoc`, `SiteParamsLoc`, etc.) stay
as-is — they're still used by `MessageForLoc`.

### How the generic handlers already work

From `routes-spotlightpa.go`:

```go
func (app *appEnv) getSiteData(w http.ResponseWriter, r *http.Request) http.Handler {
    loc := r.URL.Query().Get("location")
    return app.siteDataGet(loc)
}
func (app *appEnv) postSiteData(w http.ResponseWriter, r *http.Request) http.Handler {
    loc := r.URL.Query().Get("location")
    return app.siteDataSet(loc)
}
```

No backend changes needed.

### Consolidation commit sequence (master PR)

1. `ViewSiteParams: Use /api/site-data?location= instead of /api/site-params`
2. `ViewSidebarItems: Use /api/site-data?location= instead of /api/sidebar`
3. `endpoints: Remove getSidebar, saveSidebar, getSiteParams, postSiteParams`
4. `router: Remove /api/sidebar and /api/site-params special-case routes`

Steps 1–2 can be verified against the running backend before steps 3–4 remove
the old routes, making rollback easy if needed.

---

## Phase 1: The Feature (builds on Phase 0)

### Goal

Add the ability for Spotlight PA editors to curate sidebar items for the Berks
and State College sub-sites. Controlled by:

- `data/berks-sidebar.json`
- `data/statecollege-sidebar.json`

### Step 1 — Backend: Add loc constants

**File: `internal/almsvc/site-data.go`**

Add two new location constants and their commit messages. No new routes needed —
`/api/site-data?location=` handles them automatically after Phase 0.

```go
BerksSidebarLoc        = "data/berks-sidebar.json"
StateCollegeSidebarLoc = "data/statecollege-sidebar.json"
```

Add to `messageForLoc`:

```go
BerksSidebarLoc:        "Setting Berks County sidebar configuration",
StateCollegeSidebarLoc: "Setting State College sidebar configuration",
```

### Step 2 — Frontend: Parameterize `ViewSidebarItems.vue` via route meta

After Phase 0, `ViewSidebarItems.vue` already uses `getSiteData`/`postSiteData`
with a location string. Make that location configurable via `route.meta` so the
same component serves all three sidebars — following the same pattern as
`ViewFrontpageEditor.vue`.

- Import `useRoute` from `vue-router`.
- At setup: `const route = useRoute()`.
- Use `route.meta.location ?? "data/sidebar.json"` as the location string.
- Use `route.meta.title ?? "Sidebar Items"` for the `<MetaHead>` title and
  breadcrumb.

### Step 3 — Frontend: Add router entries

**File: `src/plugins/router.js`**

Add two new routes adjacent to the existing `sidebar-items` route. No new
imports needed — `getSiteData`/`postSiteData` are already imported after Phase 0.

```js
{
  path: "/admin/berks-sidebar",
  name: "berks-sidebar",
  component: load(() => import("@/components/ViewSidebarItems.vue")),
  meta: {
    requiresAuth: isSpotlightPAUser,
    location: "data/berks-sidebar.json",
    title: "Berks County Sidebar Items",
  },
},
{
  path: "/admin/statecollege-sidebar",
  name: "statecollege-sidebar",
  component: load(() => import("@/components/ViewSidebarItems.vue")),
  meta: {
    requiresAuth: isSpotlightPAUser,
    location: "data/statecollege-sidebar.json",
    title: "State College Sidebar Items",
  },
},
```

### Step 4 — Frontend: Add nav links

**File: `src/components/ViewAdmin.vue`**

Add two `<LinkRoute>` entries grouped near the existing sub-site frontpage editor
links (`berks-editor`, `state-college-editor`):

```html
<LinkRoute
  label="Berks Sidebar Items"
  to="berks-sidebar"
  :icon="['fas', 'check-circle']"
></LinkRoute>
<LinkRoute
  label="State College Sidebar Items"
  to="statecollege-sidebar"
  :icon="['fas', 'check-circle']"
></LinkRoute>
```

### Phase 1 commit sequence

5. `almsvc: Add Berks and State College sidebar loc constants`
6. `ViewSidebarItems: Parameterize location and title via route meta`
7. `router.js: Add berks-sidebar and statecollege-sidebar routes`
8. `ViewAdmin: Add nav links for sub-site sidebar editors`

---

## File Change Summary

### Phase 0 (consolidation PR, master)

| File | Change |
|---|---|
| `src/components/ViewSiteParams.vue` | Use `getSiteData?location=` instead of `getSiteParams` |
| `src/components/ViewSidebarItems.vue` | Use `getSiteData?location=` instead of `getSidebar` |
| `src/api/endpoints.ts` | Remove `getSidebar`, `saveSidebar`, `getSiteParams`, `postSiteParams` |
| `internal/almapp/router.go` | Remove `/api/sidebar` and `/api/site-params` special-case routes |

### Phase 1 (feature PR, based on Phase 0)

| File | Change |
|---|---|
| `internal/almsvc/site-data.go` | Add 2 loc constants + 2 commit messages |
| `src/components/ViewSidebarItems.vue` | Parameterize location and title via `route.meta` |
| `src/plugins/router.js` | Add 2 route entries with `location` and `title` meta |
| `src/components/ViewAdmin.vue` | Add 2 nav links |

## What We Are NOT Doing

- **No DB migrations** — the `site_data` table keys on `loc` string; new values
  work automatically.
- **No new `.vue` components** — `ViewSidebarItems.vue` and `SidebarItem.vue` are
  reused via parameterization.
- **No changes to `SiteParams*.vue`** — this is about sidebar items, not ad params.
- **No new backend routes** — `/api/site-data?location=` covers everything after
  Phase 0.
- **No content-store schema changes** — existing JSON shape `{ items: [...] }`
  reused as-is.
