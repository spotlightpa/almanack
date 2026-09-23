# Plan: Sub-site Sticky Ad Params (Berks & State College)

> **Implementation order:** Phase 0 (consolidation + cleanup) merges to master
> first as a standalone PR. Phase 1 (the new feature) builds on top of it.

## Phase 0: Consolidation & Cleanup (prerequisite, lands on master first)

### 0a — Delete the sidebar items feature

The sidebar items feature (`ViewSidebarItems.vue`, `SidebarItem.vue`, the
`/api/sidebar` route, the `SidebarLoc` constant, and the nav link) is unused and
can be removed entirely. This is the right time to do it since Phase 1 was
originally going to add more sidebar machinery on top of it.

Files to touch:

| File | Change |
|---|---|
| `src/components/ViewSidebarItems.vue` | Delete |
| `src/components/SidebarItem.vue` | Delete |
| `src/api/endpoints.ts` | Remove `getSidebar`, `saveSidebar` |
| `src/plugins/router.js` | Remove `sidebar-items` route |
| `src/components/ViewAdmin.vue` | Remove "Sidebar Items" nav link |
| `internal/almapp/router.go` | Remove `GET/POST /api/sidebar` routes |
| `internal/almsvc/site-data.go` | Remove `SidebarLoc` constant and its `messageForLoc` entry |

### 0b — Consolidate site-data endpoints

Right now `GET/POST /api/sidebar` (deleted above) and `GET/POST /api/site-params`
are special-cased aliases for `GET/POST /api/site-data?location=<loc>` — the
exact same Go handler, with the loc hard-coded in the route. `ViewFrontpageEditor.vue`
already uses the generic `?location=` pattern. Extend that pattern to
`ViewSiteParams.vue` and remove the redundant route pair.

| File | Change |
|---|---|
| `src/components/ViewSiteParams.vue` | Use `getSiteData + "?location=config/_default/params.json"` instead of `getSiteParams`/`postSiteParams` |
| `src/api/endpoints.ts` | Remove `getSiteParams`, `postSiteParams` |
| `internal/almapp/router.go` | Remove `GET/POST /api/site-params` routes |

(`SiteParamsLoc` stays in `site-data.go` — still used by `MessageForLoc`.)

### Phase 0 commit sequence

1. `SidebarItem, ViewSidebarItems: Delete unused sidebar items feature`
2. `endpoints, router: Remove /api/sidebar and sidebar-items route`
3. `ViewSiteParams: Use /api/site-data?location= instead of /api/site-params`
4. `endpoints, router: Remove /api/site-params special-case route`

---

## Phase 1: Sub-site Sticky Ad Params

### Goal

Allow Spotlight PA editors to configure sticky ad params (and potentially other
ad slots) independently for the Berks and State College sub-sites, stored in:

- `data/berks-sidebar.json`
- `data/statecollege-bar.json`

These are analogous to the main site's `config/_default/params.json`, but scoped
to each sub-site and containing only the params those sites need.

### Step 1 — Backend: Add loc constants

**File: `internal/almsvc/site-data.go`**

```go
BerksSiteParamsLoc        = "data/berks-sidebar.json"
StateCollegeSiteParamsLoc = "data/statecollege-bar.json"
```

Add to `messageForLoc`:

```go
BerksSiteParamsLoc:        "Setting Berks County site parameters",
StateCollegeSiteParamsLoc: "Setting State College site parameters",
```

No new backend routes needed — `/api/site-data?location=` handles them
automatically after Phase 0.

### Step 2 — Frontend: Create `ViewSubSiteParams.vue`

Create a new view component modelled on `ViewSiteParams.vue` but:

- Reads `route.meta.location` and `route.meta.title` instead of hard-coding the
  main site params location.
- Renders a sub-set of `SiteParams*.vue` child components appropriate to
  sub-sites. At minimum: `SiteParamsSticky`. Other slots (rail, banner, etc.)
  can be added here as needed — using the same `ref`/`saveData()` pattern
  already established in `SiteParams.vue`.
- The scheduling UI (add scheduled change, revert) is identical to
  `ViewSiteParams.vue` and should be copied as-is.

Because the main site params view includes many homepage and article slots that
sub-sites don't have, a separate view component (rather than parameterizing
`ViewSiteParams.vue`) is the right call — it keeps each view's component tree
explicit and easy to adjust per-site without conditional clutter.

### Step 3 — Frontend: Add router entries

**File: `src/plugins/router.js`**

```js
{
  path: "/admin/berks-params",
  name: "berks-params",
  component: load(() => import("@/components/ViewSubSiteParams.vue")),
  meta: {
    requiresAuth: isSpotlightPAUser,
    location: "data/berks-sidebar.json",
    title: "Berks County Site Params",
  },
},
{
  path: "/admin/statecollege-params",
  name: "statecollege-params",
  component: load(() => import("@/components/ViewSubSiteParams.vue")),
  meta: {
    requiresAuth: isSpotlightPAUser,
    location: "data/statecollege-bar.json",
    title: "State College Site Params",
  },
},
```

### Step 4 — Frontend: Add nav links

**File: `src/components/ViewAdmin.vue`**

Add two `<LinkRoute>` entries near the existing `berks-editor` /
`state-college-editor` links:

```html
<LinkRoute
  label="Berks Site Params"
  to="berks-params"
  :icon="['fas', 'sliders-h']"
></LinkRoute>
<LinkRoute
  label="State College Site Params"
  to="statecollege-params"
  :icon="['fas', 'sliders-h']"
></LinkRoute>
```

### Phase 1 commit sequence

5. `almsvc: Add Berks and State College sub-site params loc constants`
6. `ViewSubSiteParams: New view for sub-site ad param configuration`
7. `router.js: Add berks-params and statecollege-params routes`
8. `ViewAdmin: Add nav links for sub-site params editors`

---

## Full File Change Summary

### Phase 0

| File | Change |
|---|---|
| `src/components/ViewSidebarItems.vue` | **Delete** |
| `src/components/SidebarItem.vue` | **Delete** |
| `src/components/ViewSiteParams.vue` | Use `getSiteData?location=` instead of `getSiteParams` |
| `src/api/endpoints.ts` | Remove `getSidebar`, `saveSidebar`, `getSiteParams`, `postSiteParams` |
| `src/plugins/router.js` | Remove `sidebar-items` route |
| `src/components/ViewAdmin.vue` | Remove "Sidebar Items" nav link |
| `internal/almapp/router.go` | Remove `/api/sidebar` and `/api/site-params` routes |
| `internal/almsvc/site-data.go` | Remove `SidebarLoc` constant and `messageForLoc` entry |

### Phase 1

| File | Change |
|---|---|
| `internal/almsvc/site-data.go` | Add 2 loc constants + 2 `messageForLoc` entries |
| `src/components/ViewSubSiteParams.vue` | **New** — sub-site params view with `SiteParamsSticky` (et al.) |
| `src/plugins/router.js` | Add 2 route entries with `location` and `title` meta |
| `src/components/ViewAdmin.vue` | Add 2 nav links |

## What We Are NOT Doing

- **No DB migrations** — `site_data` keys on `loc` string; new values work automatically.
- **No new backend routes** — `/api/site-data?location=` covers everything after Phase 0.
- **No changes to existing `SiteParams*.vue` components** — they are reused as-is.
- **No content-store schema changes** — same JSON shape as main site params.
