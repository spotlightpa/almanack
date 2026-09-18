- Run `./run.sh help` to see managements commands for this repo. Do `./run.sh check-deps` after a new clone to see if you have all the tools needed.

Backend:

- Use `go doc foo.Bar` or `go doc -all foo` to read documentation for packages, types, functions, etc.

- To see source files from a dependency, or to answer questions about a dependency, run `go mod download -json MODULE` and use the returned `Dir` path to read the files.

- Use `go run .` or `go run ./cmd/foo` instead of `go build` to run programs, to avoid leaving behind build artifacts.

- To test if things compile, you can use `go test ./...` which checks if it compiles and also runs the tests. Do not use `go build` just to test compilation.

- Prefer to use `gopls` for minor refactoring where possible.

- This project uses sqlc for type-safe SQL. After modifying files in `sql/queries/` or `sql/schema/`, run `./run.sh sql` to regenerate the Go code in `internal/db/`.

- Database migrations use tern and live in `sql/schema/` with numeric prefixes (e.g., `036_foo.sql`). The format is: SQL for "up" migration, then `---- create above / drop below ----`, then SQL for "down" migration.

- SQL queries are organized by domain in `sql/queries/` (e.g., `news-feed.sql`, `page.sql`). Each query needs a `-- name: QueryName :directive` comment where directive is `:one`, `:many`, `:exec`, or `:execrows`.

- Integration tests in `pkg/integration/` require a real PostgreSQL database. Set `ALMANACK_POSTGRES` env var to run them, or they will be skipped.

- Services are configured via command-line flags. Look for `AddFlags` functions that register flags and return initialization functions.

Frontend:

- Frontend utilities should preferably in TypeScript. If an existing utility is still in JS but needs to be extended, create a commit that can be cherry-picked that just converts it to TS before adding the new functionality in a second commit.

Pull requests should pass the following rubric:

- Is this change atomic? Does it make the minimum viable change to get to a working state?
- Does this change follow the DRY principle? Can it be simplified by abstracting out repeated patterns or using an existing utility function?
- Is the code readable? Will future maintainers be able to tell what the code is doing and why?
- Is this code safe? Does it defend against unexpected inputs or unusual conditions?
- Does this code follow the project’s coding standards? Does it pass all tests and linting?
