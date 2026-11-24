# Preview Feature: Architecture Options

Editors want to preview what a story or page will look like on the Poor Richard Hugo site
before pushing it live.
Almanack currently publishes by committing files to the Poor Richard GitHub repository;
the public site is built from that repository.
All options below must generate a rendered Hugo site from draft content
without exposing that content publicly.

## Key Constraint: poor-richard is a Public Repo

The Poor Richard GitHub repository is public.
Any draft content committed to any branch of that repo — even a temporary one —
is immediately readable by anyone.
Options A, B, and E from the initial plan (push to a preview branch,
per-article branch, or trigger GitHub Actions) are therefore **not viable as described**.

**Important distinction**: the Hugo templates, layouts, assets, and config in poor-richard
are *not* secret — they are already public.
Only the unpublished article *content* is sensitive.
This means we can freely clone or reference the public repo for its build tooling;
we just must never commit draft content to it.

All viable options inject content at build time without it ever touching GitHub.

---

## Option A: Private Mirror of poor-richard + Netlify Branch Deploys

**How it works**

Create a private GitHub repository (`spotlightpa/poor-richard-preview`) that mirrors
poor-richard's templates, assets, and config.
Almanack pushes draft content to a branch of this private repo.
Netlify is configured to build the private repo; branch deploys produce isolated preview URLs.

The private mirror must stay in sync with poor-richard.
The simplest mechanism: a GitHub Actions workflow on poor-richard
that pushes changes to the private mirror on every merge to `main`.

**Pros**
- Full Hugo build; preview is pixel-identical to production.
- Isolated per-article preview URLs; shareable within the org.
- No persistent servers to operate.
- Netlify handles SSL and hosting.

**Cons**
- Two-repo setup: sync lag means previews may use slightly stale templates.
- Sync failures (the Actions mirror job) silently break previews.
- Still a 1–3 min Netlify build wait per preview.
- Branch proliferation in private repo; needs automated cleanup.
- Almanack needs a Netlify API client to poll for deploy status and return the URL.
- Additional Netlify site to configure and monitor.

**Effort**: Medium-High — private repo setup, sync workflow,
Netlify site configuration, Almanack Netlify API client, UI for build status.

---

## Option B: On-Demand Hugo Build in a Netlify Background Function

**How it works**

A new Netlify background function (up to 15 min timeout) is triggered by Almanack
when an editor requests a preview.
The function:
1. Clones the public poor-richard repo (templates only — no secrets).
2. Writes the draft content (passed in the request or fetched from Almanack's API)
   into the working tree.
3. Runs `hugo` to build the full site.
4. Uploads the result to a private S3 or Cloudflare R2 bucket under a content-addressed path.
5. Returns the preview URL to the editor.

The editor's browser loads the preview from the bucket.
CSS and JS already reference the production CDN (poor-richard's assets are deployed there),
so asset paths resolve correctly without rewriting.

**Pros**
- No persistent servers.
- Full Hugo build; production-faithful rendering.
- Draft content never touches any Git repository.
- Each preview is isolated and content-addressed.

**Cons**
- Cold-start: cloning poor-richard + running Hugo takes 1–3 min.
- Netlify Function size budget (50 MB compressed) must accommodate the Hugo binary;
  this is tight but feasible if Hugo is fetched at runtime rather than bundled.
- Preview pages served from S3 need CloudFront or R2's public bucket feature;
  direct S3 URLs are unsigned and clunky.
- Cleanup of stale preview paths needs a separate scheduled job.
- Harder to debug than a persistent server.

**Effort**: Medium-High — function code, S3/R2 bucket + CDN config,
Almanack integration to trigger and poll, UI for status and link.

---

## Option C: Dedicated Preview Server (EC2 / Fly.io)

**How it works**

A small always-on server (e.g. Fly.io free tier, or `t4g.nano` EC2 at ~$3/mo)
maintains a local clone of the public poor-richard repo.
It exposes a small private HTTP API (called by Almanack's backend, not the browser directly).
When an editor requests a preview:
1. Almanack posts the draft content to the preview server's API.
2. The server writes the content into the working tree and runs `hugo`
   (or uses `hugo server`'s watch mode for near-instant incremental rebuilds).
3. The server proxies the rendered page back to Almanack,
   which forwards it to the editor's browser.
   Alternatively, the server hosts it on a password-protected endpoint.

The preview server periodically pulls the latest poor-richard `main`
to stay in sync with template changes.

**Pros**
- Incremental Hugo rebuilds take milliseconds to a few seconds — best editor experience.
- Full Hugo feature support: shortcodes, related content, pagination, etc.
- Draft content never touches any repo.
- Simple mental model; easy to debug by SSHing in.
- Template sync is a plain `git pull` on a cron, not a mirroring pipeline.

**Cons**
- Another service to operate, patch, and monitor.
- Concurrent previews from multiple editors can corrupt the working tree.
  Mitigation: Almanack already has PostgreSQL;
  a `pg_try_advisory_lock` call before invoking the preview server
  serializes builds without any extra table or service.
  Editors who lose the lock get a "build in progress, try again shortly" response.
- Preview server must be reachable from Almanack's backend
  (Netlify Functions → preview server requires a stable URL and auth).
- Single point of failure; needs a health check and restart policy.

**Effort**: Medium — server setup, small API wrapper around Hugo,
Almanack integration, auth between Almanack and the preview server.
Higher if per-editor isolation is required.

---

## Option D: Hugo Embedded in Almanack (Subprocess, No External Calls)

**How it works**

The poor-richard Hugo templates, layouts, assets, and config are vendored
(checked into Almanack or fetched at deploy time from the public repo).
When an editor requests a preview, Almanack's backend:
1. Writes the draft content to a temp directory alongside the vendored templates.
2. Shells out to `hugo` with that temp dir as the content source.
3. Captures the HTML output.
4. Rewrites any asset paths that reference the local build output
   to point to the production CDN (most Hugo themes already output absolute CDN paths,
   so this may be a no-op).
5. Returns the HTML as an HTTP response proxied to the editor's browser.

**Pros**
- Sub-second preview; no network calls, no build queue.
- No external service dependencies at preview time.
- Draft content is processed entirely in-process (or in a temp dir) and never persisted.
- Completely isolated per request.

**Cons**
- Almanack runs as a Netlify Function (15 s timeout, 50 MB limit).
  Bundling Hugo + poor-richard's template tree may push the size limit.
  Hugo can be fetched at cold-start if bundling fails, adding ~2 s.
- Templates must be kept in sync with poor-richard's `main`.
  Out-of-sync templates mean the preview doesn't match production.
  Options: re-fetch at deploy time (adds complexity to Almanack's CI),
  or run a sync job and commit the vendor snapshot.
- Subprocess management under a serverless runtime is unreliable
  (signal handling, temp file cleanup, concurrent requests).
- Full-site Hugo features (related articles, paginated indexes)
  don't work for a single-page build; editors see the article page only,
  not surrounding navigation or sidebar context.

**Effort**: High — vendoring strategy, subprocess management,
path-rewriting logic, keeping templates in sync with poor-richard.

---

## Option E: GitHub Actions on a Private Workflow + R2/S3 Upload

**How it works**

This option avoids committing draft content to any repo by using
GitHub Actions' `workflow_dispatch` with inputs.
Almanack triggers the workflow and passes the draft content as a JSON input
(GitHub Actions inputs are limited to ~65 KB; larger content needs a presigned S3 URL).
The workflow:
1. Clones the public poor-richard repo.
2. Writes the content from the input (or downloads it from a presigned URL).
3. Runs `hugo`.
4. Uploads the built site to Cloudflare R2 or S3 under a content-addressed path.
5. Almanack polls the Actions API for completion and surfaces the URL.

Workflow runs are private (not visible in the public poor-richard repo)
because the workflow lives in a private Almanack-side repo or is triggered
via a repository dispatch on a private repo.

**Pros**
- Full Hugo build in a clean, reproducible GitHub-hosted runner (free tier).
- Draft content is passed as a parameter, not committed to any repo.
- Output is content-addressed and immutable; each preview is linkable.
- No persistent servers to operate.

**Cons**
- 1–4 min latency (Actions runner queue + clone + build).
- Input size limit requires a presigned-URL dance for large articles with images.
- Requires a private GitHub repo to host the workflow
  (can't put it in the public poor-richard repo without exposing it).
- Almanack needs a GitHub Actions API client and a polling loop.
- R2/S3 bucket + CDN setup still required for hosting the output.
- Operational complexity: several failure points with no easy live debugging.

**Effort**: Medium-High — private workflow repo, Actions API client,
presigned URL handling, storage bucket, Almanack polling + UI.

---

## Summary Matrix

| Option | Build Wait | Draft Hits Repo? | Infra Overhead | Isolation | Complexity | Cost |
|--------|-----------|-----------------|----------------|-----------|------------|------|
| A: Private mirror + Netlify | 1–3 min | Private repo only | Low (sync job) | ✅ Per-article | Medium-High | $0 |
| B: Serverless Hugo build + R2 | 1–3 min | Never | Low | ✅ Per-request | Medium-High | Low |
| C: Dedicated Hugo server | Seconds | Never | Medium | ⚠️ Needs care | Medium | ~$3/mo |
| D: Hugo inside Almanack | <1 s | Never | None | ✅ Per-request | High | $0 |
| E: GitHub Actions + R2 | 1–4 min | Never | Low | ✅ Per-run | Medium-High | ~$0 |

---

## Recommendation Notes

- **Best editor experience** (no wait, no extra infra): Option D,
  but only if Almanack moves off the serverless constraint,
  or if the Hugo binary + template bundle fits within Netlify's limits.

- **Best balance of simplicity and correctness**: Option C (dedicated Hugo server).
  A Fly.io Machine costs nothing at idle,
  the template sync is a plain `git pull`,
  and incremental Hugo rebuilds are fast.
  Concurrent-editor isolation is handled by a PostgreSQL advisory lock
  already available in Almanack, so no extra infrastructure is needed for that.

- **If the team prefers zero ops**: Option B (serverless + R2)
  or Option A (private mirror + Netlify) —
  both trade build wait time for no persistent infrastructure.
  Option A is simpler code but adds a sync pipeline;
  Option B is more code but no second repo to maintain.

- **Option E** is the most complex for the least gain over B;
  not recommended unless GitHub Actions is already deeply integrated.
