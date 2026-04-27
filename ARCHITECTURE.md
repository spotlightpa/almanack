# Almanack Architecture

Almanack is Spotlight PA's content management system for managing articles, images, and site configuration. [Vue.js](https://vuejs.org/guide/introduction.html) frontend + [Go](https://go.dev) backend deployed on [Netlify](https://docs.netlify.com/).

## High-Level Structure

```
netlify.toml          # Deployment config, plumbing for URL routing
run.sh                # Task runner (build, test, format, db commands)

src/                  # Vue.js frontend ([Vite](https://vite.dev/guide/) + [Bulma](https://bulma.io/documentation/))
  main.js             # Entry point
  components/         # Vue components (View*.vue are pages)
  api/auth.js         # Netlify Identity auth wrapper
  api/client-v2.js    # API client
  plugins/router.js   # Route definitions with role guards

funcs/almanack-api/   # Go backend entry point
internal/almapp/      # HTTP handlers, routing, CLI entry
internal/almservices/ # Business logic, Services container
internal/db/          # Database layer ([sqlc](https://docs.sqlc.dev/en/stable/)-generated)
internal/services/    # Third-party integrations (github, aws, google, mailchimp, youtube, etc.)
internal/utils/       # General-purpose helpers (httpx, iterx, slicex, stringx, timex, etc.)
internal/convert/     # Document conversion tools (blocko, tableaux)
internal/layouts/     # Server-rendered HTML layouts
internal/integration/ # Integration tests (require ALMANACK_POSTGRES)

sql/schema/           # Database migrations ([tern](https://github.com/jackc/tern))
sql/queries/          # SQL queries (sqlc)
```

## Backend

**Entry**: `funcs/almanack-api/main.go` → `internal/almapp.CLI()` → routes in `internal/almapp/router.go` (split across `routes-public.go`, `routes-editor.go`, `routes-spotlightpa.go`, `routes-background.go`, `routes-ssr.go`).

**Services**: `internal/almservices.Services` is the dependency container holding database, GitHub, S3, Google, Mailchimp, Slack, YouTube clients, etc. Configured via CLI flags/environment variables in `internal/almservices/flags.go`.

**Routes** are grouped by auth level:
- Public: `/api/healthcheck`, `/api/identity-hook`
- Partner ("Editor") role: `/api/shared-article(s)`
- Spotlight PA role: `/api/page*`, `/api/images`, `/api/site-*`, etc.
- Background tasks: `/api-background/cron` (cron scheduled every 3 min, can have longer execution time than the 15 second window for normal lambda functions)

## Frontend

**Auth**: `src/api/auth.js` wraps [Netlify Identity](https://docs.netlify.com/security/secure-access-to-sites/identity/). Roles: `admin`, `Spotlight PA`, `editor` (meaning partner).

**Routing**: `src/plugins/router.js` - routes have `meta.requiresAuth` for role guards.

**API calls**: `src/api/client-v2.js` - `get(url, params)` and `post(url, body)` with auth headers.

## Database

PostgreSQL with **sqlc** for type-safe queries:
- Schema migrations in `sql/schema/` (numbered, use tern)
- Queries in `sql/queries/` with `-- name: QueryName :one/:many/:exec` directives
- Run `./run.sh sql` to regenerate `internal/db/`

## Configuration

The backend uses CLI flags for configuration. Flags can be set via environment variables using [flagx.ParseEnv](https://pkg.go.dev/github.com/earthboundkid/flagx/v2#ParseEnv): flag names are prefixed with `ALMANACK_` and converted to `SCREAMING_SNAKE_CASE`. In production on Netlify, all config and secrets are set via environment variables.

Examples:
- `-postgres` flag → `ALMANACK_POSTGRES` env var
- `-slack-hook-url` flag → `ALMANACK_SLACK_HOOK_URL` env var

Run `./run.sh api -h` to see all available flags.

For local development, create a `.env` file with your secrets. The `./run.sh api` command sources it before running:

```bash
# .env example
export ALMANACK_POSTGRES="postgres://..."
export ALMANACK_SLACK_HOOK_URL="https://hooks.slack.com/..."
```

## Development

```bash
./run.sh              # Run Go API locally (sources .env, port 33160)
./run.sh backend      # Same as plain ./run.sh
./run.sh frontend     # Run Vite dev server (port 33159, proxies to API)
./run.sh test         # Run all tests
./run.sh sql          # Regenerate sqlc code after SQL changes
./run.sh format       # Format all code
./run.sh db:migrate   # Run database migrations
```

## Netlify Deployment

- Build: `./run.sh build:prod` compiles Go to `functions/`, Vue to `dist/`
- `/api/*` routes to `almanack-api` function
- `/api-background/*` routes to same binary with longer timeout
- SPA routes (`/admin/*`, `/shared-articles/*`) serve `index.html`
- `functions/schedule.mts` triggers `/api-background/cron` every 3 minutes
