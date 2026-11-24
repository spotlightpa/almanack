# Preview Feature: Implementation Plan

Editors want to preview what a draft story will look like on the Poor Richard Hugo site
before pushing it live.

## Constraints and Decisions Reached

- **poor-richard is a public GitHub repo.**
  Draft content must never be committed to any branch of it,
  or it becomes publicly readable before publication.
  Pushing to a preview branch, using Netlify branch deploys,
  or triggering GitHub Actions on that repo are all ruled out.

- **Hugo full-site builds take ~15 seconds** on the site's current content volume.
  Keeping `hugo server` running for incremental builds is unreliable:
  frequent `git pull`s while Hugo is watching the filesystem cause crashes.
  One-shot `hugo` builds with `--contentDir` pointing to a temp directory
  containing only the draft article are fast (under 2 seconds)
  because Hugo has almost no content to process.
  The tradeoff is that related-articles widgets and taxonomy index pages
  will be empty or show only one article;
  this is acceptable for a preview of the article page itself.

- **Asset shearing is not a concern if old hashed assets are not purged.**
  Hugo fingerprints CSS and JS by content hash.
  If the production deployment never deletes old versioned asset files from S3/CDN,
  then a preview HTML page referencing `main.a1b2c3.css` remains valid
  even after subsequent deploys rotate to `main.d4e5f6.css`.
  This is standard practice for content-addressed static assets
  and should be verified/enforced in poor-richard's deployment config.

- **Preview HTML is small** (a few KB per page) once assets are excluded.
  Storing rendered HTML in Postgres is fine;
  the earlier concern about blob size was about bundling the full asset tree,
  which is unnecessary when assets resolve from the production CDN.

- **Hosting: AWS Lightsail**, using existing AWS credits and account
  (Spotlight PA already uses S3 for image storage).
  Lightsail gives a fixed-IP always-on VM with no cold-start,
  managed by systemd, without the VPC/IAM complexity of raw EC2.
  exe.dev is not suitable: its proxy authenticates the VM owner,
  not Spotlight PA editors.

---

## Architecture Overview

```
Editor browser
    │  clicks "Preview"
    ▼
Almanack (Netlify Function)
    │  inserts article_previews row (status: pending)
    │  UI polls status
    ▼
Almanack background cron (every 3 min)
    │  picks up pending rows
    │  POSTs draft content to preview server
    ▼
Preview server (AWS Lightsail)
    │  writes content to temp dir
    │  runs: hugo --contentDir /tmp/preview-{id}/ ...
    │  reads rendered HTML
    │  cleans up temp dir
    │  returns HTML to Almanack
    ▼
Almanack
    │  stores HTML in article_previews row (status: done)
    ▼
Editor browser
    │  UI shows preview link
    ▼
Almanack (Netlify Function, /ssr/preview/{id})
    │  verifies editor is authenticated
    │  reads HTML from DB
    │  serves to browser
    │  assets (CSS, JS, images) load from production CDN
```

---

## Components

### 1. Postgres: `article_previews` table

```sql
CREATE TABLE article_previews (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id  INTEGER NOT NULL REFERENCES articles(id),
    status      TEXT NOT NULL DEFAULT 'pending',
        -- pending | building | done | failed
    html        TEXT,
    error       TEXT,
    requested_by TEXT,  -- editor email, for logging
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    built_at    TIMESTAMPTZ
);

CREATE INDEX ON article_previews (article_id, requested_at DESC);
```

A `pending` row is created immediately when the editor clicks Preview.
The background cron picks it up, sets status to `building`,
calls the preview server, and updates to `done` or `failed`.

If a row has been `building` for more than 5 minutes,
the cron treats it as failed and resets it
(handles preview server crashes or hangs mid-build).

Rows older than 48 hours are deleted by the cron.

### 2. Preview Server (AWS Lightsail)

A small Go HTTP binary running under systemd on a Lightsail `nano` instance
(1 GB RAM, 2 vCPU burst, $5/mo before credits).

**Endpoints:**

`POST /build`
- Authenticated with a shared secret in the `Authorization` header.
- Request body: JSON with article frontmatter and body content.
- Writes a Hugo content file to `/tmp/preview-{uuid}/content/articles/{slug}.md`.
- Runs `hugo --contentDir /tmp/preview-{uuid}/content
         --destination /tmp/preview-{uuid}/output
         --baseURL https://www.spotlightpa.org
         --gc --minify`
  with a 30-second timeout.
- Reads `/tmp/preview-{uuid}/output/articles/{slug}/index.html`.
- Deletes the temp directory.
- Returns the HTML as a JSON string.

`GET /health`
- Returns 200 OK. Polled by Almanack's cron for monitoring.

`POST /sync`
- Authenticated with the same shared secret.
- Runs `git pull` on the poor-richard working tree.
- Called by Almanack after each successful poor-richard deployment
  (via a webhook or build hook), not on a timer.
- Must not run while a build is in progress;
  the server holds an in-process mutex during builds
  and returns 409 Conflict if sync is requested during one.
- Returns 200 when the pull completes or 500 on failure.

### 3. Almanack Changes

**Backend:**
- New sqlc queries and migration for `article_previews`.
- Background cron handler: picks up `pending` rows,
  calls `POST /build` on the preview server,
  stores result.
- New SSR route `/ssr/preview/{id}` that reads HTML from DB
  and serves it to authenticated editors.
- Health check for the preview server
  (ping `/health`; alert via Slack if down).

**Frontend:**
- "Preview" button on the article editor page.
- Polls the preview status every 3 seconds while `pending` or `building`.
- Shows a spinner with estimated wait time.
- When `done`, shows a link that opens `/ssr/preview/{id}` in a new tab.
- When `failed`, shows the error and a "Try again" button.

---

## Auth

**Almanack → Preview server:**
A shared secret (a random 32-byte hex string) stored as an Almanack
environment variable and on the Lightsail instance.
The Almanack cron sets it in the `Authorization: Bearer {secret}` header.
The preview server rejects any request without a valid secret with 401.

The preview server is not firewalled to specific source IPs
because Netlify Functions do not have fixed egress IPs.
Application-level auth over HTTPS is sufficient.

**Editor → Preview HTML:**
The `/ssr/preview/{id}` route in Almanack uses the existing
Netlify Identity auth check (same as other `/ssr/` routes).
Editors must be logged in to Almanack to view a preview.
No separate auth mechanism is needed.

---

## TLS on Lightsail

Assign a Lightsail static IP.
Point a subdomain (e.g., `preview-api.spotlightpa.org`) at it.
Run certbot with the Let's Encrypt HTTP-01 challenge for a free certificate.
The preview server binary terminates TLS directly (using Go's `crypto/tls`)
or runs behind nginx on the same instance.
certbot's cron handles renewal.

---

## Template Sync

The preview server holds a local clone of the poor-richard repo
(public, so no credentials needed).
It is synced by calling `POST /sync` from Almanack
after a poor-richard deployment completes.

Poor Richard's deployment can send a webhook to Almanack's background API,
or Almanack's cron can detect new commits by polling GitHub's API
for the latest commit SHA on `main` and comparing to a stored value.

The sync endpoint is not called on a timer specifically to avoid
running `git pull` while Hugo may be active,
which has been observed to cause Hugo crashes.
The in-process mutex prevents the race if a sync and a build
happen to be requested at the same time.

On first setup and after a manual reset, sync is triggered by SSH'ing in
and running `git clone` or `git pull` manually with the server stopped.

---

## Operational Challenges

### Preview server availability

The preview server is a single point of failure.
If it is down, editors cannot preview — but they can still publish.
Almanack's cron pings `/health` and posts to the Slack error channel if it fails.
Lightsail instances can be configured with automatic restart after failure
via systemd's `Restart=on-failure`.

A Lightsail instance reboots for OS patches occasionally.
The systemd unit ensures the preview server restarts automatically.
In-flight builds at reboot time are detected by the 5-minute
`building` timeout in the cron and re-queued.

### Stuck or leaked temp directories

If the preview server process is killed during a build,
the temp directory is not cleaned up.
On startup, the server deletes any `/tmp/preview-*` directories
left over from previous runs.
In normal operation, `defer os.RemoveAll(tmpDir)` in Go
ensures cleanup even on Hugo subprocess failure.

### Hugo subprocess hangs

Hugo is given a 30-second `context.WithTimeout`.
If it exceeds this, the subprocess is killed, the temp dir is cleaned up,
and the preview row is marked `failed` with an error message.
Hangs are uncommon with a minimal content directory
but can occur if Hugo tries to fetch remote resources.
The `--offline` flag should be passed to prevent network calls during build.

### Disk space

Hugo builds generate output under `/tmp`.
With proper cleanup this is bounded,
but monitoring disk usage on the Lightsail instance is worthwhile.
A simple cron alert if `/tmp` exceeds a threshold catches any leak.

### Article content format

Almanack must send the draft content in the same format Hugo expects:
Markdown with TOML or YAML frontmatter, matching the schema
that poor-richard's templates consume.
If Almanack stores articles in a different internal format (e.g. structured JSON),
a conversion step is needed before sending to the preview server.
This should be the same conversion that runs at publish time,
so preview and publish are always in sync.

### Template staleness

If the sync mechanism fails (GitHub is down, the webhook misfires),
the preview server builds against stale templates.
The preview still works; it just may not reflect recent template changes.
The `/health` endpoint can include the current HEAD SHA of the poor-richard clone,
which Almanack can compare against GitHub's API and surface as a warning.

### Preview staleness for editors

After an editor generates a preview and then edits the article,
the preview is stale but the UI has no way to know.
Show the preview generation timestamp prominently.
Consider invalidating the preview (setting status to `stale`)
when the article is saved after the preview was built,
so the editor sees "Preview is out of date — regenerate?"

### Multiple previews per article

An editor may request several previews as they iterate.
Keep the most recent N previews per article (e.g. 3)
and delete older ones.
The UI shows a list of recent previews with timestamps.

### Queue depth

With a 3-minute cron window and builds taking ~2 seconds,
queue depth is unlikely to be a problem.
If many editors request previews simultaneously,
they will be processed in order within a few cron cycles.
The UI shows "queued" status so editors know to wait.
If this becomes a pain point, the preview build can be triggered
directly from the Almanack background API endpoint
(bypassing the cron delay) using an HTTP call from the frontend
to a new `/api-background/build-preview` route.

### Deployment of the preview server

The preview server binary is built from a small Go package,
probably within the Almanack repo under `cmd/preview-server`.
Deployment: build locally or in CI, `scp` to the Lightsail instance,
`systemctl restart preview-server`.
This is intentionally simple; the preview server changes infrequently.

---

## Open Questions

- **Multiple `--contentDir` mounts via config:**
  Hugo's `--contentDir` flag accepts only one path.
  Hugo's module mount system (configured in `hugo.toml`) supports layering
  multiple content directories and could be used to build with
  all published articles plus the draft
  (giving correct related-articles and homepage context).
  This would require generating a temporary `hugo.toml` per build
  and would bring build time back toward 15 seconds.
  Worth prototyping once the basic single-article preview is working
  to see if the speed penalty is acceptable.

- **Which content types need preview:**
  Articles are the primary use case.
  Pages, newsletters, and other content types may need the same feature
  with different Hugo content paths.
  The preview server API should be designed to accept a content type
  and path alongside the article body from the start.

- **Poor Richard asset purging policy:**
  Confirm that the production deployment does not delete old
  content-hashed CSS/JS files from S3 on deploy.
  If it does, shearing becomes a real problem
  and assets must be included in the preview artifact.
  This should be verified before shipping the feature.
