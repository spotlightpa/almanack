# Plan: Remove Arc & Retire Partner Pages

## Goal

The Arc Publishing integration is dead code — no one uses it. Partners no
longer consume Almanack's shared-article pages either. We want to:

1. Strip out everything Arc-specific (DB, Go, Vue, assets).
2. Replace the partner-facing routes (`/shared-articles`, `/shared-articles/:id`,
   and partner-bound API endpoints) with a "no longer available" stub.
3. **Keep** the Spotlight-PA admin side of shared articles
   (`/admin/shared-articles/:id`, the gdocs ingestion pipeline, and the
   `shared_article` table) intact — Spotlight editors still use it to manage
   internal article metadata.

## What "Arc" currently touches

Backend (Go):
- `internal/services/arc/arc-schema.go` — entire package, only structs.
- `internal/db/arc.sql.go` (`GetArcByArcID`) and `sql/queries/arc.sql`.
- `internal/db/shared-article.sql.go` — `UpsertSharedArticleFromArc` (unused
  outside the generated file).
- `internal/almapp/routes-spotlightpa.go` — `case "arc":` branch in
  `postPageRefresh` (~line 694) that just returns conflict.
- `internal/db/models.go` — generated `Arc` struct.

Database:
- `sql/schema/023_shared_article.sql` created the `arc` table.
- `sql/schema/025_shared_article_props.sql` back-filled shared_article columns
  from arc raw_data.
- The `shared_article.source_type = 'arc'` rows still exist in prod, plus the
  `arc` table.

Frontend (Vue/JS):
- Components: `ArcArticleAvailable`, `ArcArticleDivider`, `ArcArticleHTML`,
  `ArcArticleHeader`, `ArcArticleImage`, `ArcArticleList`, `ArcArticleOEmbed`,
  `ArcArticlePlaceholder`, `ArcArticlePlanned`, `ArcArticleText`, `ThumbnailArc`.
- API: `src/api/arc-article.js`, and `isArc`/`sourceType === "arc"` branches in
  `src/api/shared-article.js`.
- Auth role: `isArcUser` (`"arc user"`) in `src/api/auth.js` and
  `ArticleSlugLine.vue`.
- Router: legacy `/articles/:id` -> `arc-article` redirect entry.
- Sprinkled `article.isArc` branches in `ViewAdmin.vue`, `ViewSharedArticle.vue`,
  `ViewSharedArticleAdmin.vue`, `ArticleWordCount.vue`, `ArticleList.vue`.

Partner-facing routes/components to gut:
- Route `shared-articles` (`/shared-articles`, list) →
  `src/components/ViewSharedArticles.vue`.
- Route `shared-article` (`/shared-articles/:id`, detail) →
  `src/components/ViewSharedArticle.vue`.
- Route `arc-article` (`/articles/:id` legacy redirect).
- Route `shared-article-redirect-from-page` (`/admin/article-redirect`) — partner
  redirect target.
- Home redirect: `isEditor` users currently land on `shared-articles`.
- API endpoints (partner-gated by `editor` role):
  `GET /api/shared-article`, `GET /api/shared-articles` →
  `internal/almapp/routes-editor.go` (`listSharedArticles`, `getSharedArticle`).
- The `editor` role itself in `partnerMW` / `partnerSSRMW` in `router.go`.
  (`partnerSSRMW` also exposes `GET /ssr/download-image`; that is editor-only
  but Spotlight users also have it, since `admin` role bypasses checks.)

## Step-by-step

### Step 1 — Frontend: stub the partner pages
Lowest-risk change, ships visible "no longer available" UI immediately.

- Replace `src/components/ViewSharedArticles.vue` and
  `src/components/ViewSharedArticle.vue` with a single static "This feature is
  no longer available" page (or point both routes at one new
  `ViewPartnerGone.vue`).
- In `src/plugins/router.js`:
  - Keep `/shared-articles` and `/shared-articles/:id` paths but point them at
    the new stub. Drop `requiresAuth: isEditor` so unauthenticated visitors see
    the message too.
  - Remove the `/articles` and `/articles/:id` (legacy arc-article) routes, or
    redirect them to the stub.
  - Remove the `shared-article-redirect-from-page` route (was an Arc-era
    redirect helper).
  - Change the `home` redirect: editors should fall through to
    `unauthorized` (or sign-out) rather than `shared-articles`.
- Drop `ViewArticleRedirect.vue` if no other route uses it after the above.

### Step 2 — Frontend: rip out Arc UI
- Delete components: `ArcArticle*.vue`, `ThumbnailArc.vue`.
- Delete `src/api/arc-article.js`.
- In `src/api/shared-article.js`: drop `fromArc`, `isArc`, `arc` field,
  the `import ArcArticle`. Anything that used `isArc` on the admin side
  (`ViewSharedArticleAdmin.vue`, `ViewAdmin.vue`, `ArticleSlugLine.vue`,
  `ArticleWordCount.vue`, `ArticleList.vue`) becomes simpler — assume
  every shared article is a gdoc.
- In `src/api/auth.js`: remove `isArcUser`/`"arc user"` role.
- In `ArticleSlugLine.vue`: remove the `isArcUser && article.isArc` branch.
- Audit: `rg -i arc src/` should return nothing meaningful afterward (a couple
  of unrelated substring hits like `clear`/`search` are fine).

### Step 3 — Backend: remove partner endpoints
- In `internal/almapp/router.go`:
  - Delete the `partnerMW` block and its two endpoints
    (`GET /api/shared-article`, `GET /api/shared-articles`).
  - Delete `partnerSSRMW` if nothing else needs it.  `GET /ssr/download-image`
    can move to `spotlightMW`-equivalent SSR (admins still need it).
- Delete `listSharedArticles` and `getSharedArticle` from
  `internal/almapp/routes-editor.go`. Rename the file or fold `userInfo` into
  another route file if it's the last function left.
- Remove the `editor`-role check helpers if no longer referenced (likely
  still want them around since `hasRoleMiddleware` is generic — keep, but
  confirm `"editor"` is unreferenced).
- Confirm nothing in `internal/almsvc` still cares about `editor` role.

### Step 4 — Backend: drop Arc Go code
- `rm internal/services/arc/arc-schema.go` and the `internal/services/arc`
  directory.
- Delete `sql/queries/arc.sql`, `internal/db/arc.sql.go`, and the `Arc` struct
  in `internal/db/models.go` (regenerate after Step 5 with `./run.sh sql`).
- Delete `UpsertSharedArticleFromArc` in `sql/queries/shared-article.sql`
  (and the generated companion).
- In `internal/almapp/routes-spotlightpa.go`, drop the `case "arc":` arm in
  `postPageRefresh` (the `default` will catch any straggling rows).
- `grep -rn arc internal/` should only show acronyms ("Architecture",
  "archive", etc.) and `arc` in historical migration files.

### Step 5 — Database migration
Add `sql/schema/040_drop_arc.sql`:

```
DELETE FROM shared_article WHERE source_type = 'arc';
DROP TABLE arc;

---- create above / drop below ----

CREATE TABLE arc (...);  -- copy from 023, no data restore
```

Notes:
- The cascade is clean because `shared_article` only references `arc` via the
  free-form `source_type/source_id` strings; there is no FK.
- Down-migration recreates the empty table for symmetry but does not
  resurrect data. Acceptable since prod uses tern forward-only.
- After migrating, run `./run.sh sql` to regenerate `internal/db/`.

### Step 6 — Tests & docs
- Update / delete any fixture that still references arc shared articles
  (search `testdata/` and `integration/`).
- `go test ./...` (skips integration without `ALMANACK_POSTGRES`).
- With `ALMANACK_POSTGRES` set, run integration to confirm migrations apply
  cleanly to a snapshot.
- README / ARCHITECTURE: remove Arc references; note partner UI sunset.

### Step 7 — Cleanup pass
- `rg -i "\barc\b" .` (excluding `dist/`, `node_modules/`, and migration
  history) — expect only unrelated matches.
- Run `./run.sh check-deps` / `./run.sh build` / `yarn build` to make sure
  nothing is dangling.
- Commit per step with clear messages so the change is bisectable:
  1. `frontend: replace partner pages with sunset notice`
  2. `frontend: remove Arc components and helpers`
  3. `api: remove partner-only endpoints`
  4. `go: remove Arc service and queries`
  5. `sql: drop arc table and shared_article rows`
  6. `docs/tests: tidy after Arc removal`

## Open questions / risks

- **Identity roles**: the Netlify identity "editor" / "arc user" roles still
  exist on user accounts. Removing the middleware is fine, but consider
  documenting that Spotlight admin will not change them (no migration
  needed).
- **Old bookmarks**: external partners might still hit
  `/shared-articles/<id>`; keep the path live with the sunset message rather
  than 404. Step 1 does this.
- **Netlify role gating**: `netlify.toml` / Identity may have role rules for
  `/shared-articles*`. Audit before shipping so unauthenticated visitors can
  see the sunset notice.
- **Apple News / mailchimp** still reference `shared_article` rows — verify
  by inspection that none of them depend on `source_type='arc'` (search
  confirms they don't, but worth eyeballing during PR review).
- **Down-migration**: prod is forward-only, but if a rollback is needed mid
  deploy, dropping the table is destructive. Snapshot or `pg_dump arc` before
  applying.
