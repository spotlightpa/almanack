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

What we explicitly **do not** delete:
- `shared_article` rows with `source_type='arc'`. They stay in the list views
  (`/shared-articles`, `/admin`) so the historical record is intact. Clicking
  through to such an article shows a "no longer available" message instead of
  rendering Arc content.

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

### Step 1 — Frontend: stub Arc detail view, prune Arc UI
- `src/components/ViewSharedArticle.vue`: replace the `v-if="article.isArc"`
  branch (which renders `ArcArticleAvailable` / `ArcArticlePlanned`) with an
  inline "This article is no longer available" notice. Keep `isArc` on the
  model so we can branch on it. Gdocs path is unchanged.
- `src/components/ViewSharedArticleAdmin.vue`: same treatment — show a
  short "Arc article, no longer rendered" notice instead of the Arc-only
  admin fields. Keep the row visible so admins can still see metadata
  (status, dates, internal id) and the page link if any. Also drop the
  `a.arc ? "Word count: …" : ""` segment from the `emailBody` computed
  property that feeds the Mailchimp send (`/api/message`). The rest of
  the email body — budget, note, embargo, publication date, detail URL —
  is sourced from `shared_article` columns and stays.
- Lists (`ViewAdmin.vue`, `ViewSharedArticles.vue` indirectly via
  `ArticleList.vue`, `ArticleSlugLine.vue`, `ArticleWordCount.vue`): keep
  `isArc` rows visible but drop the now-broken sub-components. For
  example `ArticleWordCount` should just return empty for Arc rows;
  `ArticleSlugLine`'s `isArcUser && isArc` external-link tag goes away.
  Shared fields like `budget`, `internalID`, `hed`, byline, etc. still
  render for Arc rows — they're populated on the row itself (migration
  025 back-filled them) and are also used by gdocs articles.
- `src/api/shared-article.js`: keep `isArc` (it's the discriminator for the
  stub). Drop `fromArc`, the `arc` ArcArticle instance, and the
  `import ArcArticle`. Anything that read `article.arc.*` is gone (only the
  Arc components used it).
- `src/api/auth.js`: drop `isArcUser` / `"arc user"` role.
- `src/plugins/router.js`: remove the legacy `/articles/:id` (`arc-article`)
  route. Keep `shared-articles`, `shared-article`, and
  `shared-article-redirect-from-page` — still needed for gdocs.
- Delete components: `ArcArticleAvailable`, `ArcArticleDivider`,
  `ArcArticleHTML`, `ArcArticleHeader`, `ArcArticleImage`, `ArcArticleList`,
  `ArcArticleOEmbed`, `ArcArticlePlaceholder`, `ArcArticlePlanned`,
  `ArcArticleText`, `ThumbnailArc`.
- Delete `src/api/arc-article.js`.
- Audit: `rg -i arc src/` should only show `isArc` discriminator usage.

### Step 2 — Backend: drop Arc Go code
- `rm -r internal/services/arc/`.
- Delete `sql/queries/arc.sql`; the generated `internal/db/arc.sql.go` and
  the `Arc` model struct will go away when we regenerate.
- Delete the `UpsertSharedArticleFromArc` query in
  `sql/queries/shared-article.sql`.
- In `internal/almapp/routes-spotlightpa.go`, drop the `case "arc":` arm in
  `postPageRefresh` (`default` will catch stragglers and respond with the
  same conflict error).
- The list/get endpoints (`ListSharedArticles*`, `GetSharedArticleByID`,
  `GetSharedArticleBySource`) keep working unchanged; the `raw_data` field
  for Arc rows stays in the response but is unused by the new frontend.
- Verify: `rg -n '\barc\b' internal/ sql/queries/` (case-sensitive) should
  only show generated files we're about to regen.

### Step 3 — Hide `arc` from sqlc via schema override
We don't need to drop the prod table; we just need sqlc to stop generating
code for it. Use the same trick that hid the old `newsletter` table.

- Append to `sql/schema-overrides/001.sql`:
  ```
  DROP TABLE arc;
  ```
- Remove the `column: arc.raw_data` override in `sql/sqlc.json`.
- Delete `sql/queries/arc.sql` and the `UpsertSharedArticleFromArc` query in
  `sql/queries/shared-article.sql` (both reference the now-hidden table).
- Run `./run.sh sql`. The generated `internal/db/arc.sql.go` and the
  `Arc` struct in `internal/db/models.go` disappear.
- No tern migration, no row deletions. The prod `arc` table sits dormant
  forever, same as `newsletter`.

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
  3. `sql: hide arc table from sqlc`
  4. `docs/tests: tidy after Arc removal`

## Open questions / risks

- **Existing `source_type='arc'` shared_article rows**: kept. They remain
  visible in list views; clicking through hits the new "no longer
  available" stub. `raw_data` is still served by the API but ignored by
  the frontend.
- **Linked pages**: `page.source_type='arc'` rows still exist (per migration
  019). They are read-only in the partner UI and the `case "arc":` branch
  was the only refresh handler. After this change, refresh on those pages
  falls into `default` and returns a conflict, which matches today's
  behavior. We are **not** deleting `page` rows.
- **Identity roles**: the Netlify identity "arc user" role lingers on user
  accounts. Harmless after this change; document but no migration needed.

