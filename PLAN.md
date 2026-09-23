# Plan: Sub-site Sidebar Sticky Items (Berks & State College)

> **Implementation order:** Phase 0 (consolidation + cleanup) merges to master
> first as a standalone PR. Phase 1 (the new feature) builds on top of it.

## Phase 0: Consolidation & Cleanup (prerequisite, lands on master first)

### 0a — Remove main-site sidebar items wiring

The main-site sidebar items feature is unused and being dropped. However,
`ViewSidebarItems.vue` and `SidebarItem.vue` are the right components for Phase 1
(sub-site sidebar stickies), so **the components themselves are kept** — only
their main-site wiring is removed:

| File | Change |
|---|---|
| `src/api/endpoints.ts` | Remove `getSidebar`, `saveSidebar` |
| `src/plugins/router.js` | Remove `sidebar-items` route |
| `src/components/ViewAdmin.vue` | Remove "Sidebar Items" nav link |
| `internal/almapp/router.go` | Remove `GET/POST /api/sidebar` routes |
| `internal/almsvc/site-data.go` | Remove `SidebarLoc` constant and its `messageForLoc` entry |

### 0b — Consolidate site-data endpoints

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

### Phase 0 commit sequence

1. `endpoints, router, almsvc: Remove main-site sidebar items wiring`
2. `ViewSiteParams: Use /api/site-data?location= instead of /api/site-params`
3. `endpoints, router: Remove /api/site-params special-case route`

---

## Phase 1: Sub-site Sidebar Sticky Items

### Goal

Allow Spotlight PA editors to curate sidebar sticky items independently for the
Berks and State College sub-sites, using the same `ViewSidebarItems.vue` /
`SidebarItem.vue` components already built for the main site. Stored in:

- `data/berks-sidebar.json`
- `data/statecollege-bar.json`

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

No new backend routes needed — `/api/site-data?location=` handles them after
Phase 0.

### Step 2 — Frontend: Parameterize `ViewSidebarItems.vue` via route meta

After Phase 0, `ViewSidebarItems.vue` still hard-codes
`get(getSidebar)`/`post(saveSidebar, ...)`. Update it to use
`getSiteData?location=` and read the location from `route.meta`, following the
same pattern as `ViewFrontpageEditor.vue`.

- Import `useRoute` from `vue-router`.
- Use `route.meta.location` (required — no main-site fallback needed now that the
  main-site wiring is gone) for the `?location=` query param on both fetch and
  save calls.
- Use `route.meta.title` for the `<MetaHead>` title and breadcrumb label.

### Step 3 — Frontend: Add router entries

**File: `src/plugins/router.js`**

Add two routes adjacent to where `sidebar-items` used to be:

```js
{
  path: "/admin/berks-sidebar",
  name: "berks-sidebar",
  component: load(() => import("@/components/ViewSidebarItems.vue")),
  meta: {
    requiresAuth: isSpotlightPAUser,
    location: "data/berks-sidebar.json",
    title: "Berks County Sidebar",
  },
},
{
  path: "/admin/statecollege-sidebar",
  name: "statecollege-sidebar",
  component: load(() => import("@/components/ViewSidebarItems.vue")),
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
  :icon="['fas', 'check-circle']"
></LinkRoute>
<LinkRoute
  label="State College Sidebar"
  to="statecollege-sidebar"
  :icon="['fas', 'check-circle']"
></LinkRoute>
```

### Phase 1 commit sequence

4. `almsvc: Add Berks and State College sidebar loc constants`
5. `ViewSidebarItems: Parameterize location and title via route meta`
6. `router.js: Add berks-sidebar and statecollege-sidebar routes`
7. `ViewAdmin: Add nav links for sub-site sidebar editors`

---

## Full File Change Summary

### Phase 0

| File | Change |
|---|---|
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
| `src/components/ViewSidebarItems.vue` | Parameterize location and title via `route.meta` |
| `src/plugins/router.js` | Add 2 route entries with `location` and `title` meta |
| `src/components/ViewAdmin.vue` | Add 2 nav links |

## What We Are NOT Doing

- **No DB migrations** — `site_data` keys on `loc` string; new values work automatically.
- **No new backend routes** — `/api/site-data?location=` covers everything after Phase 0.
- **No deleting `ViewSidebarItems.vue` or `SidebarItem.vue`** — they're reused as-is for Phase 1.
- **No changes to `SiteParams*.vue`** — the sticky slider in those components is a different UI element; sub-site sidebar stickies use the sidebar items pattern.
- **No content-store schema changes** — same JSON shape as `data/sidebar.json`.
