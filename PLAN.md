# Plan: Sub-site Sticky Controllers (Berks & State College)

> **Implementation order:** Phases 1 & 2 (consolidation + cleanup) merge to master
> first as standalone PRs. Phase 3 (the new feature) builds on top of them.

## Phase 1: Delete Sidebar Items

### Overview

`ViewSidebarItems.vue` and `SidebarItem.vue` are unused and being dropped in
their entirety.

| File | Change |
|---|---|
| `src/components/ViewSidebarItems.vue` | **Delete** |
| `src/components/SidebarItem.vue` | **Delete** |
| `src/api/endpoints.ts` | Remove `getSidebar`, `saveSidebar` |
| `src/plugins/router.js` | Remove `sidebar-items` route |
| `src/components/ViewAdmin.vue` | Remove "Sidebar Items" nav link |
| `internal/almapp/router.go` | Remove `GET/POST /api/sidebar` routes |
| `internal/almsvc/site-data.go` | Remove `SidebarLoc` constant and its `messageForLoc` entry |

## Phase 2: Consolidate Site-Data Endpoints

`GET/POST /api/site-params` is a special-cased alias for
`GET/POST /api/site-data?location=<loc>` — same handler, loc hard-coded in the
route. `ViewFrontpageEditor.vue` already uses the generic `?location=` pattern.
Extend that to `ViewSiteParams.vue` and remove the redundant route pair.

| File | Change |
|---|---|
| `src/components/ViewSiteParams.vue` | Use `getSiteData + "?location=config/_default/params.json"` instead of `getSiteParams`/`postSiteParams` |
| `src/api/endpoints.ts` | Remove `getSiteParams`, `postSiteParams` |
| `internal/almapp/router.go` | Remove `GET/POST /api/site-params` routes |

(`SiteParamsLoc` stays in `site-data.go` — still used by `MessageForLoc`.)

### Phase 1 commit sequence

1. `ViewSidebarItems, SidebarItem: Delete unused sidebar items feature`
2. `endpoints, router, almsvc: Remove /api/sidebar route and SidebarLoc`

### Phase 2 commit sequence

3. `ViewSiteParams: Use /api/site-data?location= instead of /api/site-params`
4. `endpoints, router: Remove /api/site-params special-case route`

---

## Phase 3: Sub-site Sticky Controllers

### Goal

Allow Spotlight PA editors to configure sticky items independently for the Berks
and State College sub-sites, stored in:

- `data/berks-sidebar.json`
- `data/statecollege-bar.json`

The UI is a new view modelled on `ViewSiteParams.vue`, rendering a curated subset
of the existing `SiteParams*.vue` components appropriate to sub-sites.

### Step 1 — Backend: Add loc constants

**File: `internal/almsvc/site-data.go`**

```go
BerksSidebarLoc        = "data/berks-sidebar.json"
StateCollegeSidebarLoc = "data/statecollege-bar.json"
```

Add to `messageForLoc`:

```go
BerksSidebarLoc:        "Setting Berks County sidebar configuration",
StateCollegeSidebarLoc: "Setting State College sidebar configuration",
```

No new backend routes — `/api/site-data?location=` handles them after Phase 2.

### Step 2 — Frontend: Create `ViewSubSiteSidebar.vue`

New view modelled directly on `ViewSiteParams.vue` (scheduling UI, save/revert,
`SiteParamsModel` class, etc.) with two differences:

- Reads `route.meta.location` and `route.meta.title` instead of hard-coding the
  main site params path.
- Renders only the `SiteParams*.vue` child components relevant to sub-site
  stickies. The exact set to confirm with the editor, but at minimum
  `SiteParamsSticky`. Other rail/ad slots can be added later without changing the
  architecture.

A separate component (rather than parameterizing `ViewSiteParams.vue`) is the
right call: the main site view includes homepage and article slots that sub-sites
don't need, and keeping each view's component tree explicit is cleaner than
conditionals.

### Step 3 — Frontend: Add router entries

**File: `src/plugins/router.js`**

```js
{
  path: "/admin/berks-sidebar",
  name: "berks-sidebar",
  component: load(() => import("@/components/ViewSubSiteSidebar.vue")),
  meta: {
    requiresAuth: isSpotlightPAUser,
    location: "data/berks-sidebar.json",
    title: "Berks County Sidebar",
  },
},
{
  path: "/admin/statecollege-sidebar",
  name: "statecollege-sidebar",
  component: load(() => import("@/components/ViewSubSiteSidebar.vue")),
  meta: {
    requiresAuth: isSpotlightPAUser,
    location: "data/statecollege-bar.json",
    title: "State College Sidebar",
  },
},
```

### Step 4 — Frontend: Add nav links

**File: `src/components/ViewAdmin.vue`**

Add two `<LinkRoute>` entries near the existing `berks-editor` /
`state-college-editor` links:

```html
<LinkRoute
  label="Berks Sidebar"
  to="berks-sidebar"
  :icon="['fas', 'sliders-h']"
></LinkRoute>
<LinkRoute
  label="State College Sidebar"
  to="statecollege-sidebar"
  :icon="['fas', 'sliders-h']"
></LinkRoute>
```

### Phase 3 commit sequence

5. `almsvc: Add Berks and State College sidebar loc constants`
6. `ViewSubSiteSidebar: New sub-site sticky controller view`
7. `router.js: Add berks-sidebar and statecollege-sidebar routes`
8. `ViewAdmin: Add nav links for sub-site sidebar editors`

---

## Full File Change Summary

### Phases 1 & 2

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

### Phase 3

| File | Change |
|---|---|
| `internal/almsvc/site-data.go` | Add 2 loc constants + 2 `messageForLoc` entries |
| `src/components/ViewSubSiteSidebar.vue` | **New** — sub-site sticky controller |
| `src/plugins/router.js` | Add 2 route entries with `location` and `title` meta |
| `src/components/ViewAdmin.vue` | Add 2 nav links |

## What We Are NOT Doing

- **No DB migrations** — `site_data` keys on `loc` string; new values work automatically.
- **No new backend routes** — `/api/site-data?location=` covers everything after Phase 2.
- **No content-store schema changes** — same JSON shape as main site params.
