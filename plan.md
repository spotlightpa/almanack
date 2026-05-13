# Plan: Remove Arc

## Goal

The Arc Publishing integration is dead code — no one uses it. Strip out
everything Arc-specific (DB, Go, Vue, assets) while leaving the rest of the
shared-article pipeline intact.

What **stays**:
- The `shared_article` table and all gdocs-sourced rows.
- Spotlight admin pages (`/admin`, `/admin/shared-articles/:id`).
- **Partner pages** (`/shared-articles`, `/shared-articles/:id`) and their
  API endpoints (`GET /api/shared-article`, `GET /api/shared-articles`) —
  partners still use these to read gdocs-sourced shared articles.
- The `editor` role and `partnerMW` middleware.

What **goes**:
- Anything that only exists to render or fetch Arc content.
- The legacy `/articles/:id` -> arc redirect route.
- The "arc user" role.
- The `arc` Postgres table and the `arc_id` flavored helpers.
- Any `shared_article` row whose `source_type='arc'` (there should only be
  gdocs rows after this).

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

## Step-by-step

### Step 1 — Frontend: prune Arc from shared-article views
- `src/components/ViewSharedArticle.vue`: remove the `v-if="article.isArc"`
  branch (both `ArcArticleAvailable` and `ArcArticlePlanned`). The component
  becomes "render `GDocsDoc`" unconditionally; if `article.isArc` would be
  true (legacy data only) we can show a small "no longer available"
  fallback inline rather than crash.
- `src/components/ViewSharedArticleAdmin.vue`, `ViewAdmin.vue`,
  `ArticleSlugLine.vue`, `ArticleWordCount.vue`, `ArticleList.vue`: delete
  every `article.isArc` branch and Arc-only column (`budget` is Arc-only —
  confirm before stripping, since it's also a gdoc-era column on
  `shared_article`).
- `src/api/shared-article.js`: drop `fromArc`, `isArc`, the `arc` field,
  and the `import ArcArticle`.
- `src/api/auth.js`: drop `isArcUser` / `"arc user"` role.
- `src/plugins/router.js`: remove the legacy `/articles/:id` (`arc-article`)
  route. Keep `shared-articles`, `shared-article`, and
  `shared-article-redirect-from-page` — they're partner-visible and serve
  gdocs.
- Delete components: `ArcArticleAvailable`, `ArcArticleDivider`,
  `ArcArticleHTML`, `ArcArticleHeader`, `ArcArticleImage`, `ArcArticleList`,
  `ArcArticleOEmbed`, `ArcArticlePlaceholder`, `ArcArticlePlanned`,
  `ArcArticleText`, `ThumbnailArc`.
- Delete `src/api/arc-article.js`.
- Audit: `rg -i arc src/` should return nothing meaningful.

### Step 2 — Backend: drop Arc Go code
- `rm -r internal/services/arc/`.
- Delete `sql/queries/arc.sql`; the generated `internal/db/arc.sql.go` and
  the `Arc` model struct will go away when we regenerate.
- Delete the `UpsertSharedArticleFromArc` query in
  `sql/queries/shared-article.sql`.
- In `internal/almapp/routes-spotlightpa.go`, drop the `case "arc":` arm in
  `postPageRefresh` (`default` will catch stragglers and respond with the
  same conflict error).
- Verify: `rg -n '\barc\b' internal/ sql/queries/` (case-sensitive) should
  only show generated files we're about to regen.

### Step 3 — Database migration
Add `sql/schema/040_drop_arc.sql`:

```
DELETE FROM shared_article WHERE source_type = 'arc';
DROP TABLE arc;

---- create above / drop below ----

CREATE TABLE arc (...);  -- copy from 023, no data restore
```

Notes:
- No FK between `shared_article` and `arc`; the link is via the free-form
  `source_type/source_id` strings.
- Down-migration recreates the empty table for symmetry; tern is
  forward-only in prod so this is a courtesy.
- After applying the schema, run `./run.sh sql` to regenerate
  `internal/db/`.

### Step 4 — Tests & docs
- Update / delete any fixture that still references arc shared articles
  (search `testdata/` and `internal/integration/`).
- `go test ./...` (skips integration without `ALMANACK_POSTGRES`).
- With `ALMANACK_POSTGRES` set, run integration to confirm the migration
  applies cleanly against a recent snapshot.
- README / ARCHITECTURE: remove Arc references.

### Step 5 — Cleanup pass
- `rg -i "\barc\b" .` (excluding `dist/`, `node_modules/`, and migration
  history) — expect only unrelated matches.
- `./run.sh sql` + `go test ./...` + `yarn build`.
- Commit per step:
  1. `frontend: remove Arc components and shared-article branches`
  2. `go: remove Arc service and queries`
  3. `sql: drop arc table and arc-sourced shared_article rows`
  4. `docs/tests: tidy after Arc removal`

## Open questions / risks

- **Existing `source_type='arc'` shared_article rows**: deleted by the
  migration. Confirm via `SELECT count(*) FROM shared_article WHERE
  source_type='arc';` and snapshot before applying — there are likely
  hundreds of historical rows.
- **Linked pages**: `page.source_type='arc'` rows still exist (per migration
  019). They are read-only in the partner UI and the `case "arc":` branch
  was the only refresh handler. After this change, refresh on those pages
  falls into `default` and returns a conflict, which matches today's
  behavior. We are **not** deleting `page` rows.
- **Identity roles**: the Netlify identity "arc user" role lingers on user
  accounts. Harmless after this change; document but no migration needed.
- **Apple News / mailchimp**: also read from `shared_article`. Confirmed by
  grep that none of them filter on `source_type='arc'`; still worth a
  reviewer eyeball.
- **Budget field on `shared_article`**: present in the schema and used in
  the admin UI for Arc only. Decide during step 1 whether to leave the
  column or drop it in a follow-up migration. (Recommendation: leave it,
  it's just `text NOT NULL DEFAULT ''`.)
