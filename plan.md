# Punchlist: Remove Arc

## 1. Frontend

- [ ] `src/components/ViewSharedArticle.vue`: replace the `v-if="article.isArc"` branch with an inline "This article is no longer available" notice. Leave the gdocs path alone.
- [ ] `src/components/ViewSharedArticleAdmin.vue`:
  - Replace Arc-only admin fields (around line 251 and line 424) with a short "Arc article, no longer rendered" notice. Keep status, dates, internal id, and any page link.
  - In the `emailBody` computed property, drop the `a.arc ? "Word count: …" : ""` segment. Keep the rest (budget, note, embargo, pub date, detail URL).
- [ ] `src/components/ArticleWordCount.vue`: return empty for `article.isArc` rows.
- [ ] `src/components/ArticleSlugLine.vue`: remove the `isArcUser && article.isArc` Arc-link tag and the `isArcUser` import.
- [ ] `src/components/ArticleList.vue`, `src/components/ViewAdmin.vue`: remove `article.isArc` branches. Shared fields (`budget`, `internalID`, `hed`, byline) still render for Arc rows.
- [ ] `src/api/shared-article.js`: drop `fromArc`, the `arc = new ArcArticle(…)` block, and the `import ArcArticle` line. Keep `isArc` getter — it's the discriminator for the stub.
- [ ] `src/api/auth.js`: remove `isArcUser` and the `"arc user"` role.
- [ ] `src/plugins/router.js`: delete the `arc-article` route (`/articles/:id`). Keep `shared-articles`, `shared-article`, and `shared-article-redirect-from-page`.
- [ ] Delete files:
  - `src/api/arc-article.js`
  - `src/components/ArcArticleAvailable.vue`
  - `src/components/ArcArticleDivider.vue`
  - `src/components/ArcArticleHTML.vue`
  - `src/components/ArcArticleHeader.vue`
  - `src/components/ArcArticleImage.vue`
  - `src/components/ArcArticleList.vue`
  - `src/components/ArcArticleOEmbed.vue`
  - `src/components/ArcArticlePlaceholder.vue`
  - `src/components/ArcArticlePlanned.vue`
  - `src/components/ArcArticleText.vue`
  - `src/components/ThumbnailArc.vue`
- [ ] Run `yarn build`; `rg -i arc src/` should only show `isArc` discriminator usage.
- [ ] Commit: `frontend: remove Arc components and shared-article branches`.

## 2. Backend Go

- [ ] `rm -r internal/services/arc/`.
- [ ] In `internal/almapp/routes-spotlightpa.go` `postPageRefresh` (~line 694), delete the `case "arc":` arm. The `default` arm already returns the same conflict error.
- [ ] Delete the `UpsertSharedArticleFromArc` query block from `sql/queries/shared-article.sql`.
- [ ] Delete `sql/queries/arc.sql`.
- [ ] `go test ./...` (compile check). `rg -n '\barc\b' internal/` should only flag generated `internal/db/arc.sql.go` and the `Arc` struct in `internal/db/models.go` (those go away in step 3).
- [ ] Commit: `go: remove Arc service and queries`.

## 3. SQL: hide `arc` from sqlc

Same pattern used for the old `newsletter` table.

- [ ] Append to `sql/schema-overrides/001.sql`:
  ```sql
  DROP TABLE arc;
  ```
- [ ] In `sql/sqlc.json`, remove the override block for `column: arc.raw_data`.
- [ ] Run `./run.sh sql`.
- [ ] Confirm `internal/db/arc.sql.go` is gone and the `Arc` struct is gone from `internal/db/models.go`.
- [ ] `go test ./...`.
- [ ] Commit: `sql: hide arc table from sqlc`.

No tern migration. The prod `arc` table stays dormant. `shared_article` rows with `source_type='arc'` stay in place.

## 4. Tests & docs

- [ ] Search `internal/integration/` and any `testdata/` for fixtures with `source_type="arc"` or `arc_id` and update or drop.
- [ ] `go test ./...`.
- [ ] If `ALMANACK_POSTGRES` is available, run integration tests.
- [ ] Scrub Arc references from `README.md` and `ARCHITECTURE.md`.
- [ ] Commit: `docs/tests: tidy after Arc removal`.

## 5. Final audit

- [ ] `rg -i "\barc\b" . -g '!dist' -g '!node_modules' -g '!sql/schema/0*' -g '!sql/one-time'` returns only unrelated matches.
- [ ] `./run.sh sql && go test ./... && yarn build` all succeed.
- [ ] Verify in a dev environment: a known Arc shared article shows the "no longer available" stub at `/shared-articles/:id`; gdocs articles still render normally; admin pages still list both kinds.

## Notes for the doer

- `page.source_type='arc'` rows are untouched. Page refresh on them already errored; after step 2 the error comes from `default` instead of the dedicated arm. Same user-visible behavior.
- The Netlify Identity `"arc user"` role on existing accounts is harmless once `auth.js` stops checking for it. Leave it; no migration needed.
- `shared_article.budget` is shared with gdocs articles — do not remove.
- The `arc` table itself is not dropped in prod; the schema-override trick just hides it from sqlc. Same approach as `newsletter`.
