# Plan: Sub-site Sidebar Editors (Berks & State College)

## Goal

Add the ability for Spotlight PA editors to curate sidebar items for the Berks and
State College sub-sites, using the same sidebar item structure already used for the
main site sidebar (`data/sidebar.json`). The new data files will be:

- `data/berks-sidebar.json` (Berks County)
- `data/statecollege-sidebar.json` (State College)

## Background: How the Main Sidebar Works

The existing sidebar flow is a useful template:

1. **Content store constant** - `internal/almsvc/site-data.go` defines
   `SidebarLoc = "data/sidebar.json"` and maps it to a human-readable commit
   message.
2. **Backend routes** - `internal/almapp/router.go` wires up `GET /api/sidebar`
   and `POST /api/sidebar` using the generic `app.siteDataGet(loc)` /
   `app.siteDataSet(loc)` handlers. These handlers read/write scheduled configs
   stored in the DB and publish the active config to the content store.
3. **API endpoint constants** - `src/api/endpoints.ts` exports `getSidebar` and
   `saveSidebar` (both pointing to `/api/sidebar`).
4. **Frontend view** - `src/components/ViewSidebarItems.vue` loads configs via
   `getSidebar`, renders each scheduled config using `SidebarItem.vue`
   sub-components, and saves via `saveSidebar`.
5. **Router entry** - `src/plugins/router.js` has a named route `sidebar-items`
   pointing to `ViewSidebarItems.vue`.
6. **Admin nav** - `src/components/ViewAdmin.vue` has a link to `sidebar-items`
   under "Spotlight PA promotions".

The backend handlers (`siteDataGet` / `siteDataSet`) are already generic - they
take any `loc` string. **No new Go code is needed** beyond adding two constants
and two pairs of route registrations.

## Plan

### Step 1 - Backend: Add constants and routes

**File: `internal/almsvc/site-data.go`**

Add two new location constants and their commit messages:

```go
BerksSidebarLoc        = "data/berks-sidebar.json"
StateCollegeSidebarLoc = "data/statecollege-sidebar.json"
```

Add to `messageForLoc`:

```go
BerksSidebarLoc:        "Setting Berks County sidebar configuration",
StateCollegeSidebarLoc: "Setting State College sidebar configuration",
```

**File: `internal/almapp/router.go`**

Add four new route registrations in the `spotlightMW` block, adjacent to the
existing `/api/sidebar` pair:

```go
HandleFunc(mux, `GET /api/berks-sidebar`, app.siteDataGet(almsvc.BerksSidebarLoc)).
HandleFunc(mux, `POST /api/berks-sidebar`, app.siteDataSet(almsvc.BerksSidebarLoc)).
HandleFunc(mux, `GET /api/statecollege-sidebar`, app.siteDataGet(almsvc.StateCollegeSidebarLoc)).
HandleFunc(mux, `POST /api/statecollege-sidebar`, app.siteDataSet(almsvc.StateCollegeSidebarLoc))
```

### Step 2 - Frontend API: Add endpoint constants

**File: `src/api/endpoints.ts`**

Add alongside the existing sidebar constants:

```ts
export const getBerksSidebar = `/api/berks-sidebar`;
export const saveBerksSidebar = `/api/berks-sidebar`;
export const getStateCollegeSidebar = `/api/statecollege-sidebar`;
export const saveStateCollegeSidebar = `/api/statecollege-sidebar`;
```

### Step 3 - Frontend: Parameterize `ViewSidebarItems.vue` via route meta

Rather than copying the component (which would create three copies to maintain),
we follow the same pattern used by `ViewFrontpageEditor.vue`: read `route.meta`
to determine which endpoint and title to use.

`ViewSidebarItems.vue` currently hard-codes `getSidebar` / `saveSidebar`. Refactor
it to pull these from `route.meta`, falling back to the main sidebar endpoints:

- Import `useRoute` from `vue-router`.
- At setup: `const route = useRoute()`.
- Replace the hard-coded `get(getSidebar)` with `get(route.meta.getEndpoint ?? getSidebar)`
  and likewise for the save call.
- Read `route.meta.title` (default: `"Sidebar Items"`) for the `<MetaHead>` title
  and breadcrumb.

No other changes to `SidebarItem.vue` or any `SiteParams*.vue` component are
needed - those are entirely unrelated to this feature.

### Step 4 - Frontend: Add router entries

**File: `src/plugins/router.js`**

Add two new routes in the Spotlight section, adjacent to the existing
`sidebar-items` route. Import the new endpoint constants from `endpoints.ts`
(check the existing import list at the top of `router.js` and add there):

```js
{
  path: "/admin/berks-sidebar",
  name: "berks-sidebar",
  component: load(() => import("@/components/ViewSidebarItems.vue")),
  meta: {
    requiresAuth: isSpotlightPAUser,
    getEndpoint: getBerksSidebar,
    saveEndpoint: saveBerksSidebar,
    title: "Berks County Sidebar Items",
  },
},
{
  path: "/admin/statecollege-sidebar",
  name: "statecollege-sidebar",
  component: load(() => import("@/components/ViewSidebarItems.vue")),
  meta: {
    requiresAuth: isSpotlightPAUser,
    getEndpoint: getStateCollegeSidebar,
    saveEndpoint: saveStateCollegeSidebar,
    title: "State College Sidebar Items",
  },
},
```

### Step 5 - Frontend: Add nav links

**File: `src/components/ViewAdmin.vue`**

Add two `<LinkRoute>` entries. Group them near the existing sub-site frontpage
editor links (`berks-editor`, `state-college-editor`) rather than with the main
"Sidebar Items" link, since these are sub-site-specific:

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

## File Change Summary

| File | Change |
|---|---|
| `internal/almsvc/site-data.go` | Add 2 loc constants + 2 commit messages |
| `internal/almapp/router.go` | Register 4 new routes (GET+POST x2) |
| `src/api/endpoints.ts` | Add 4 endpoint constants |
| `src/plugins/router.js` | Add 2 route entries with meta; import new endpoint constants |
| `src/components/ViewSidebarItems.vue` | Parameterize endpoint and title via `route.meta` |
| `src/components/ViewAdmin.vue` | Add 2 nav links |

## What We Are NOT Doing

- **No DB migrations** - the existing `site_data` table stores records keyed by
  `loc` string, so new `loc` values work automatically.
- **No new `.vue` components** - `ViewSidebarItems.vue` and `SidebarItem.vue` are
  reused directly via parameterization.
- **No changes to `SiteParams*.vue`** - this feature is about sidebar items, not
  site-wide ad/promo params.
- **No content-store schema changes** - the existing JSON shape (`{ items: [...] }`)
  is reused as-is.

## Commit Sequence

1. `almsvc: Add Berks and State College sidebar loc constants`
2. `router: Register berks-sidebar and statecollege-sidebar API routes`
3. `endpoints: Add berks-sidebar and statecollege-sidebar API constants`
4. `ViewSidebarItems: Parameterize endpoint and title via route meta`
5. `router.js: Add berks-sidebar and statecollege-sidebar routes`
6. `ViewAdmin: Add nav links for sub-site sidebar editors`
