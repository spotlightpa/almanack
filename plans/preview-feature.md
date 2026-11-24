# Preview Feature: Implementation Plan

Editors want to preview what a draft story will look like on the Poor Richard Hugo site
before pushing it live.
This plan describes a new standalone service, **Groundhog**
(Groundhog Day — it forgets everything every morning),
and the corresponding changes to Almanack.

## Constraints

- **poor-richard is a public GitHub repo.**
  Draft content must never be committed to any branch of it.

- **Hugo full-site builds take ~15 seconds** on the site's current content volume.
  Building with a single article via `--contentDir` takes ~3 seconds and produces ~5 MB of output.
  Related-articles widgets and taxonomy index pages will be empty or contain only the draft article;
  this is acceptable for checking article layout, typography, and shortcode rendering.

- **The zip is self-contained.**
  Hugo's output includes compiled, fingerprinted CSS and JS.
  Bundling them in the zip avoids any dependency on the production CDN
  and eliminates asset shearing when the production site redeploys.

---

## Groundhog: Architecture Overview

Groundhog is a small Go HTTP server running on an AWS Lightsail instance
(existing AWS account; credits available).
It does two things: build previews, and serve them.

```
Almanack backend
    │  POST /_groundhog/build  (HMAC-signed)
    ▼
Groundhog (groundhog.spotlightpa.org)
    │  acquires in-memory mutex
    │  writes draft .md to tempdir
    │  runs: hugo --contentDir tempdir --baseURL https://groundhog.spotlightpa.org/
    │  zips hugo output
    │  uploads zip to S3 (random key, 24h lifecycle)
    │  releases mutex
    ▼
Almanack frontend
    │  polls GET /_groundhog/build-status?id=XXX every 300ms
    ▼
Groundhog
    │  when done: 302 to /articles/{slug}/
    │            Set-Cookie: preview_zip={s3key}; Path=/; HttpOnly; Max-Age=86400
    ▼
Editor browser
    │  lands on /articles/{slug}/  (served from zip via cookie)
    │  browser fetches /css/main.a1b2c3.css
    ▼
Groundhog
    │  reads preview_zip cookie
    │  fetches zip from S3 (cached in memory)
    │  extracts and serves css/main.a1b2c3.css
```

---

## Groundhog: API

### `POST /_groundhog/build`

Authenticated by HMAC-SHA256 signature over the request body,
using a shared secret (same pattern as GitHub webhooks:
`X-Hub-Signature-256: sha256={hex}`).

Request body (JSON):
```json
{
  "article_path": "content/articles/my-story.md",
  "article_content": "---\ntitle: My Story\n...\n---\n\nBody here."
}
```

Response (immediately, before build starts):
```json
{"build_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479"}
```

Groundhog runs the build in a goroutine and returns immediately.
If a build is already in progress (mutex held), Groundhog returns 409 Conflict.
Almanack should surface this to the editor as "a preview is already building, try again shortly."

### `GET /_groundhog/build-status?id=XXX`

Polled by the Almanack frontend every 300ms.

While building:
```json
{"status": "building"}
```

On failure:
```json
{"status": "failed", "error": "hugo exited with code 1: ..."}
```

On success: `302 Found` to `/articles/{slug}/`
with `Set-Cookie: preview_zip={s3key}; Path=/; HttpOnly; SameSite=Lax; Max-Age=86400`.

Build state is kept in an in-memory map (`map[string]*buildState` guarded by a sync.RWMutex).
Since there is exactly one Groundhog instance, no external state store is needed.
Entries are pruned after 1 hour.

### `GET /*` (preview serving)

All other requests check for the `preview_zip` session cookie.
If present, Groundhog fetches the named zip from S3
(LRU cache in memory, holds up to ~100 MB worth of zips),
extracts the requested path, and serves it with an appropriate `Content-Type`.

If the cookie is absent or the key does not exist in S3: 404.

This means multiple editors can browse their respective previews concurrently;
each editor's browser carries their own cookie pointing to their zip.

### `GET /_groundhog/health`

Returns 200 with the current poor-richard HEAD SHA and whether a build is in progress.
Polled by Almanack's cron for monitoring.

---

## Groundhog: Build Process

```
1. Acquire sync.Mutex (one build at a time)
2. Create tempdir: /tmp/groundhog-{buildID}/content/articles/
3. Write article_content to /tmp/groundhog-{buildID}/content/articles/{slug}.md
4. Run with 60s timeout:
     hugo \
       --source /srv/poor-richard \
       --contentDir /tmp/groundhog-{buildID}/content \
       --destination /tmp/groundhog-{buildID}/output \
       --baseURL https://groundhog.spotlightpa.org/ \
       --gc --minify --quiet
5. Walk /tmp/groundhog-{buildID}/output and write a zip to /tmp/groundhog-{buildID}/output.zip
6. Upload output.zip to S3: previews/{random-uuid}.zip
7. Store s3key in buildState; mark done
8. defer os.RemoveAll("/tmp/groundhog-{buildID}") — runs even on failure
9. Release mutex
```

The S3 bucket has a lifecycle rule that expires objects under `previews/` after 24 hours.

---

## Groundhog: Template Sync

Groundhog maintains a local clone of `github.com/spotlightpa/poor-richard` at `/srv/poor-richard`.
Sync is triggered externally via a dedicated endpoint rather than a background goroutine,
keeping Groundhog's internal state simple.

### `POST /_groundhog/sync`

Authenticated with the same shared secret as `/_groundhog/build`.

```
1. If build mutex is held: return 409 Conflict
2. Acquire mutex
3. git -C /srv/poor-richard pull --ff-only
4. If yarn.lock changed since last pull: yarn --cwd /srv/poor-richard install
5. Release mutex
6. Return 200 with new HEAD SHA
```

Almanack's cron calls this endpoint once daily at 3am.
This is handled by a dedicated Netlify scheduled function
(separate from the existing every-3-minutes cron)
with a `0 3 * * *` schedule expression in `netlify.toml`.

For an out-of-schedule sync (e.g. after a significant poor-richard template change),
the endpoint can be called manually with curl:
```bash
curl -X POST https://groundhog.spotlightpa.org/_groundhog/sync \
  -H "Authorization: Bearer {secret}"
```

Node and yarn must be installed on the Lightsail instance alongside Hugo,
since poor-richard uses a node-based asset pipeline.

---

## Groundhog: Deployment

**Instance**: AWS Lightsail `nano_3_0` (1 GB RAM, 2 vCPU burst, $5/mo before credits).

**DNS**: `groundhog.spotlightpa.org` → Lightsail static IP.

**TLS**: nginx on the instance terminates HTTPS via Let's Encrypt (certbot).
Groundhog listens on `localhost:8080`; nginx proxies to it.

**Systemd unit** (`/etc/systemd/system/groundhog.service`):
```ini
[Unit]
Description=Groundhog Preview Server
After=network.target

[Service]
Type=simple
User=groundhog
WorkingDirectory=/srv/groundhog
ExecStart=/srv/groundhog/groundhog
Restart=on-failure
RestartSec=5
EnvironmentFile=/srv/groundhog/.env

[Install]
WantedBy=multi-user.target
```

**`.env`** (on instance, not in repo):
```
GROUNDHOG_SHARED_SECRET=...
GROUNDHOG_AWS_ACCESS_KEY_ID=...
GROUNDHOG_AWS_SECRET_ACCESS_KEY=...
GROUNDHOG_S3_BUCKET=spotlightpa-previews
GROUNDHOG_POOR_RICHARD_DIR=/srv/poor-richard
GROUNDHOG_LISTEN=127.0.0.1:8080
```

**Deploying a new Groundhog binary**:
```bash
GOOS=linux GOARCH=amd64 go build -o groundhog ./cmd/groundhog
scp groundhog user@groundhog.spotlightpa.org:/srv/groundhog/groundhog.new
ssh user@groundhog.spotlightpa.org \
  'mv /srv/groundhog/groundhog.new /srv/groundhog/groundhog && systemctl restart groundhog'
```

Groundhog changes infrequently; manual deploy is fine to start.

**IAM**: A dedicated IAM user with a policy allowing only
`s3:PutObject`, `s3:GetObject` on `arn:aws:s3:::spotlightpa-previews/previews/*`.
Lightsail instances do not support IAM instance roles directly,
so static credentials in the env file are used.

---

## Almanack Changes

**New environment variables**: `ALMANACK_GROUNDHOG_URL` and `ALMANACK_GROUNDHOG_SECRET`.

**Editor action**: "Preview" button on the article editor page.

**Flow**:
1. Frontend POSTs to Almanack API: `POST /api/article/{id}/preview`.
2. Almanack serializes the article to Hugo frontmatter + Markdown
   (same conversion used at publish time — must be identical).
3. Almanack POSTs to `{GROUNDHOG_URL}/_groundhog/build` with HMAC signature.
4. If 409: return "preview already building" to frontend.
5. If 200: return `{build_id}` to frontend.
6. Frontend opens a new tab to `{GROUNDHOG_URL}/_groundhog/build-status?id={build_id}`
   and polls it at 300ms.
   When Groundhog returns the redirect, the browser follows it;
   the editor lands on their preview page with the cookie set.

No Almanack database changes are required.
Build state lives in Groundhog's memory.
Almanack does not need to poll or store preview status;
the frontend tab handles it directly against Groundhog.

**Health monitoring**: Almanack's existing background cron
pings `{GROUNDHOG_URL}/_groundhog/health` and posts to Slack if it fails.

---

## Operational Considerations

**Mutex contention**: Only one build at a time.
If two editors click Preview simultaneously, the second gets a 409
and Almanack tells them to try again in a few seconds.
With ~3s build time this is a minor inconvenience.

**Hugo subprocess hangs**: Builds run with a 60-second `context.WithTimeout`.
On timeout, the subprocess is killed, the temp dir is cleaned up via `defer`,
and the build state is marked failed.

**Groundhog crash during build**: On restart, in-memory build state is lost.
Any frontend tab polling a build that was in-progress at crash time
will eventually get a 404 on `/_groundhog/build-status`.
Almanack should treat a 404 from the status endpoint as a failure
and prompt the editor to retry.
In-flight temp directories are cleaned up on startup via `filepath.Glob("/tmp/groundhog-*")`.

**Instance reboot**: systemd restarts Groundhog automatically.
Same recovery path as a crash.

**Disk space**: Hugo builds go to `/tmp` and are always removed by `defer`.
Monitor with a simple cron alert if `/tmp` exceeds a threshold.

**Zip cache eviction**: The in-memory LRU zip cache is bounded by size
(e.g., 200 MB, ~40 zips).
S3 is the source of truth; cache misses re-fetch transparently.
An editor returning to a preview after the cache has been evicted
will see a slightly slower first load (S3 fetch) but no error.

**Preview after article edit**: The session cookie points to a zip
built from a previous version of the article.
If the editor edits the article after previewing, the preview is stale.
The Almanack frontend should show the preview generation time
and offer a "Rebuild preview" button.
The old cookie/zip remains valid until the lifecycle policy expires it;
triggering a new build sets a new cookie, replacing the old one.

**Template staleness**: The 10-minute sync cycle means previews may use
templates up to 10 minutes old.
This is acceptable; poor-richard template deploys are infrequent.
If a template change is urgent, SSH in and run
`git -C /srv/poor-richard pull && systemctl restart groundhog`
(the restart flushes in-memory state and re-syncs cleanly on next cycle).

**Yarn install duration**: `yarn install` with a warm cache takes a few seconds.
Running it in the sync goroutine while the mutex is held means
a build request during yarn install will block briefly.
This is preferable to running yarn and hugo simultaneously.

**Multiple editors viewing concurrently**: Each editor's browser carries
its own `preview_zip` cookie from when they triggered their own build.
Groundhog serves each correctly from their respective zip.
No per-editor state is needed beyond the cookie.

---

## Open Questions

**Article content format**: The conversion from Almanack's internal representation
to Hugo-ready Markdown+frontmatter must be identical to the publish-time conversion.
If they diverge, editors approve a preview that doesn't match what goes live.
This should be a shared function called by both paths.

**Full-site build option**: Hugo's module mount system supports layering
multiple content directories via a generated `hugo.toml`.
Building with all published content alongside the draft
would give correct related-articles and homepage context,
at the cost of returning to ~15s build times.
Worth prototyping as a "high fidelity" mode once the basic preview ships.

**Slug derivation**: The redirect from `/_groundhog/build-status` needs to know
the article's URL path. Almanack should include the expected Hugo output path
(`article_output_path`) in the `/_groundhog/build` request so Groundhog can construct
the correct redirect target.
