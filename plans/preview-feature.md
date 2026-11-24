# Preview Feature: Architecture Options

Editors want to preview what a story or page will look like on the Poor Richard Hugo site
before pushing it live.
Almanack currently publishes by committing files to the Poor Richard GitHub repository;
the public site is built from that repository.
All options below must generate a rendered Hugo site from draft content
without touching the production branch.

---

## Option A: Netlify Deploy Previews via a Preview Branch

**How it works**

Almanack pushes draft content to a dedicated `preview` branch of the Poor Richard repo
instead of `main`.
Netlify's branch-deploy feature automatically builds and hosts that branch at a stable
subdomain (e.g. `preview--poor-richard.netlify.app`).
The Almanack UI links editors to that URL.

**Pros**
- Minimal new infrastructure: Netlify already builds Poor Richard.
- The preview is a full Hugo build, identical to production.
- Netlify handles SSL, CDN, auth options (Netlify Identity).
- No new services to operate or pay for separately.
- Branch deploys are retained and linkable per commit.

**Cons**
- Multiple editors writing simultaneously share one preview branch,
  so their changes will collide or bleed into each other's previews.
- Each preview requires a full Netlify build (usually 1–3 min); no instant feedback.
- Editors see each other's unpublished drafts (potential information leak within the org).
- Requires configuring branch-deploy rules and protecting the branch on GitHub.

**Effort**: Low — a flag swap (`-github-branch preview`) for a preview-mode publish,
plus a link in the UI.

---

## Option B: Per-Article Netlify Deploy Previews (One Branch per Story)

**How it works**

For each preview request, Almanack creates or force-pushes a short-lived branch
(e.g. `preview/article-slug`) to the Poor Richard repo.
Netlify builds it and returns a unique deploy URL
(Netlify's API exposes `deploy.deploy_ssl_url` per branch deploy).
Almanack stores that URL and shows it to the editor.

**Pros**
- Isolated per article: editors cannot see each other's drafts.
- Each preview URL is stable and shareable (useful for editor-in-chief sign-off).
- Still a full Hugo build with no new runtime infrastructure.

**Cons**
- Branch proliferation in the repo (needs automated cleanup).
- Still a 1–3 min build wait.
- Netlify API integration needed to poll for the deploy URL and status.
- More complex Almanack code: branch create/delete, deploy status polling.

**Effort**: Medium — Almanack needs Netlify API client code,
branch lifecycle management, and UI to show build status + link.

---

## Option C: On-Demand Hugo Build in a Netlify Function (Serverless)

**How it works**

A new Netlify background function (or AWS Lambda) receives the draft content,
clones the Poor Richard repo, runs `hugo` to build a single page or the full site,
then uploads the resulting HTML to a temporary S3/R2 bucket path
or returns it inline.
The editor sees the rendered page served from that path.

**Pros**
- No persistent servers to run.
- Could be nearly instant for a single-page build
  (Hugo can render one page in well under a second).
- Fully isolated per request.

**Cons**
- Netlify Functions have a 50 MB deployment size limit and a 15 s timeout
  (background functions: 15 min), making a full Hugo + git clone marginal.
- Hugo binary + Poor Richard repo together are large; cold starts are slow.
- Partial builds (single page) require Hugo's `--contentDir` tricks and
  may miss cross-page data (related articles, nav).
- Output needs to be hosted somewhere with working asset references
  (CSS/JS paths will break if served from an S3 key).
- Higher operational complexity; debugging is hard.

**Effort**: High — requires custom Lambda/function code, asset-path rewriting,
temporary hosting, and cleanup.

---

## Option D: Long-Running EC2 (or Fly.io) Preview Server

**How it works**

A small always-on server (e.g. `t4g.nano` EC2, ~$3/mo, or Fly.io free tier)
runs Hugo in server mode (`hugo server`) with a checkout of the Poor Richard repo.
When an editor requests a preview, Almanack calls this server's API to:
1. Write the draft content to the working tree.
2. Trigger `hugo` to rebuild.
3. Return a URL to the rendered page (served by the preview server).

Access is gated by Almanack's existing auth or an HTTP basic-auth header.

**Pros**
- Incremental Hugo rebuilds are very fast (milliseconds to seconds).
- Full Hugo feature support: shortcodes, templates, related content, pagination.
- Simple mental model; easy to debug by SSHing in.
- Can serve previews indefinitely for review workflows.

**Cons**
- Another service to operate, patch, and monitor.
- `hugo server` is not designed for concurrent multi-user writes;
  simultaneous previews can corrupt the working tree.
  (Mitigated by a queue or per-editor checkout directories.)
- Preview server is a single point of failure.
- Needs to be accessible from Almanack's backend (or the editor's browser directly).

**Effort**: Medium — server setup + a small HTTP API wrapper around Hugo +
Almanack integration. High if isolation per editor is required.

---

## Option E: GitHub Actions Preview (CI-triggered Build + Pages/R2 Upload)

**How it works**

Almanack pushes draft content to a preview branch and triggers a GitHub Actions workflow
(via `workflow_dispatch` or a branch-push trigger).
The workflow runs `hugo`, then uploads the built site to a Cloudflare R2 bucket
(or AWS S3 + CloudFront) under a path keyed by article slug or commit SHA.
Almanack polls the Actions API for completion and surfaces the resulting URL.

**Pros**
- Uses existing GitHub relationship; no new SaaS account needed beyond storage.
- Full Hugo build in a clean, reproducible environment.
- Artifacts are content-addressed (SHA-keyed), so each preview is immutable and linkable.
- GitHub Actions free tier is generous; unlikely to cost anything.

**Cons**
- Workflow dispatch + poll adds 1–4 min of latency.
- Requires R2/S3 bucket configuration and asset-path handling
  (same path problem as Option C).
- More moving parts: branch push → Actions → upload → URL → Almanack display.
- Cleanup of old preview paths needs a separate job.

**Effort**: Medium-High — GitHub Actions workflow, storage bucket setup,
Almanack polling loop, URL display in UI.

---

## Option F: Hugo Running Inside Almanack Itself (Embedded Preview)

**How it works**

The Poor Richard repo (or a subset: templates, assets, config) is vendored or
cloned into Almanack's server at deploy time.
When an editor requests a preview, Almanack writes the draft content to a temp directory,
shells out to `hugo` (or calls a Hugo library if one exists), captures the HTML output,
and returns it as a proxied response to the editor's browser.
CSS and JS are rewritten to point to the production CDN so assets resolve correctly.

**Pros**
- No external service calls or network round-trips; response in under a second.
- No persistent preview infrastructure.
- Completely isolated per request (temp dirs).
- No build wait time for the editor.

**Cons**
- Almanack runs as a Netlify Function (15 s timeout, 50 MB limit);
  bundling Hugo + Poor Richard templates may exceed the size limit.
- The Poor Richard template set must be kept in sync with Almanack's deployment.
- Asset-path rewriting is fiddly; inline styles, `srcset`, and JS imports all need handling.
- Hugo is not designed to be called as a library; subprocess management under a
  serverless runtime is unreliable.

**Effort**: High — significant engineering for the subprocess, path rewriting,
and keeping template snapshots in sync.

---

## Summary Matrix

| Option | Build Wait | Infra Overhead | Isolation | Complexity | Cost |
|--------|-----------|----------------|-----------|------------|------|
| A: Shared Netlify branch | 1–3 min | None | ❌ Shared | Low | $0 |
| B: Per-article Netlify branch | 1–3 min | None | ✅ Per-article | Medium | $0 |
| C: Serverless Lambda build | Marginal | Low | ✅ Per-request | High | Low |
| D: EC2/Fly.io Hugo server | Seconds | Medium | ⚠️ Needs care | Medium | ~$3/mo |
| E: GitHub Actions + R2 | 1–4 min | Low | ✅ Per-commit | Medium-High | ~$0 |
| F: Hugo inside Almanack | <1 s | None | ✅ Per-request | High | $0 |

---

## Recommendation Notes

- **Quickest to ship**: Option A (shared branch).
  Good enough if only one editor previews at a time and information leakage between
  editors is acceptable.
- **Best balance**: Option B (per-article Netlify branch).
  Builds on existing Netlify infrastructure, gives isolated and shareable URLs,
  and the main cost is the build wait time.
- **Best editor experience** (no wait): Option D (dedicated Hugo server)
  if the team is comfortable operating a small VM.
- **Most adventurous / high-reward**: Option F,
  but only if Almanack moves off the serverless constraint
  (e.g. to a long-running Fly.io or EC2 deployment).
