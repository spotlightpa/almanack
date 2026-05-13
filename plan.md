# Punchlist: Remove Arc

Shipped as four PRs, each deployable on its own. Order matters: the
user-visible "no longer available" notice goes first so we can verify Arc
rows are dead before pulling anything out from under the code.

---

## PR 1 — Show "no longer available" for Arc articles

Goal: replace Arc rendering with a stub. No deletions, no API changes,
no backend changes. Easy to roll back.

- [ ] `src/components/ViewSharedArticle.vue`: replace the
  `v-if="article.isArc"` block (renders `ArcArticleAvailable` /
  `ArcArticlePlanned`) with an inline "This article is no longer
  available" notice. Gdocs path unchanged.
- [ ] `src/components/ViewSharedArticleAdmin.vue`:
  - Replace the Arc-only admin sections (around lines 251 and 424) with
    a short "Arc article, no longer rendered" notice. Keep status,
    dates, internal id, page link.
  - In the `emailBody` computed property, drop the
    `a.arc ? "Word count: …" : ""` segment. Keep budget, note, embargo,
    pub date, detail URL.
- [ ] `src/components/ArticleWordCount.vue`: return empty for Arc rows.
- [ ] `src/components/ArticleSlugLine.vue`: remove the
  `isArcUser && article.isArc` Arc-link tag.
- [ ] `src/components/ArticleList.vue`, `src/components/ViewAdmin.vue`:
  remove `article.isArc` branches (other than the row staying visible).
- [ ] `yarn build`; click through a known Arc shared article in dev and
  confirm the stub renders. Confirm gdocs articles still render.
- [ ] Commit + PR: `frontend: show "no longer available" for Arc shared articles`.
- [ ] Deploy. Verify in prod with a real Arc URL before starting PR 2.

Leave alone for now: `src/api/arc-article.js`, all `ArcArticle*` components,
`isArcUser`, the `/articles/:id` legacy route, backend code. They are
unreferenced after this PR but removing them is PR 2.

---

## PR 2 — Remove Arc frontend

Goal: delete the dead Arc UI code. Backend still serves Arc `raw_data`;
frontend just stops reading it.

- [ ] `src/api/shared-article.js`: drop `fromArc`, the
  `arc = new ArcArticle(…)` block, and the `import ArcArticle` line.
  Keep the `isArc` getter — it still drives the PR 1 stub.
- [ ] `src/api/auth.js`: remove `isArcUser` and the `"arc user"` role.
- [ ] `src/components/ArticleSlugLine.vue`: drop the `isArcUser` import
  left over from PR 1.
- [ ] `src/plugins/router.js`: delete the `arc-article` route
  (`/articles/:id`). Keep `shared-articles`, `shared-article`, and
  `shared-article-redirect-from-page`.
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
- [ ] `yarn build`. `rg -i arc src/` should only show `isArc` discriminator.
- [ ] Commit + PR: `frontend: remove Arc components and helpers`.
- [ ] Deploy. Smoke-test admin and partner views.

---

## PR 3 — Remove Arc backend Go code

Goal: rip out the unused Arc Go package, query, and route arm. `arc` table
stays; sqlc still generates code for it (we deal with that in PR 4).

- [ ] `rm -r internal/services/arc/`.
- [ ] In `internal/almapp/routes-spotlightpa.go` `postPageRefresh`
  (~line 694), delete the `case "arc":` arm. The `default` arm returns
  the same conflict error for any straggler.
- [ ] Delete the `UpsertSharedArticleFromArc` query in
  `sql/queries/shared-article.sql`.
- [ ] Delete `sql/queries/arc.sql`.
- [ ] Run `./run.sh sql`. `internal/db/arc.sql.go` regenerates without the
  upsert query but still has `GetArcByArcID` (we drop the table next PR).
- [ ] `go test ./...`.
- [ ] Commit + PR: `go: remove Arc service and queries`.
- [ ] Deploy.

---

## PR 4 — Hide `arc` from sqlc, final cleanup

Goal: stop generating code for the `arc` table without touching prod data.
Same trick as the old `newsletter` table.

- [ ] Append to `sql/schema-overrides/001.sql`:
  ```sql
  DROP TABLE arc;
  ```
- [ ] In `sql/sqlc.json`, remove the override block for `column: arc.raw_data`.
- [ ] Run `./run.sh sql`. `internal/db/arc.sql.go` disappears; `Arc` struct
  is gone from `internal/db/models.go`.
- [ ] `go test ./...`.
- [ ] Search `internal/integration/` and any `testdata/` for fixtures with
  `source_type="arc"` or `arc_id` and update or drop.
- [ ] Scrub Arc references from `README.md` and `ARCHITECTURE.md`.
- [ ] Final audit:
  `rg -i "\barc\b" . -g '!dist' -g '!node_modules' -g '!sql/schema/0*' -g '!sql/one-time'`
  should return only unrelated matches.
- [ ] `./run.sh sql && go test ./... && yarn build`.
- [ ] Commit + PR: `sql: hide arc table from sqlc; tidy docs`.
- [ ] Deploy.

---

## Notes for the implementer

- `shared_article` rows with `source_type='arc'` stay forever; they remain
  visible in list views and the detail page shows the PR 1 stub.
- `page.source_type='arc'` rows are untouched. Page refresh on them already
  errored; after PR 3 the error comes from the `default` arm instead. Same
  user-visible behavior.
- The Netlify Identity `"arc user"` role on existing accounts is harmless
  once `auth.js` stops checking for it. Leave it; no migration needed.
- `shared_article.budget` is shared with gdocs articles — do not remove.
- Prod `arc` table is not dropped. The schema-override trick just hides it
  from sqlc, same as `newsletter`.
