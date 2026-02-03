- To see source files from a dependency, or to answer questions about a dependency, run `go mod download -json MODULE` and use the returned `Dir` path to read the files.

- Use `go doc foo.Bar` or `go doc -all foo` to read documentation for packages, types, functions, etc.

- Use `go run .` or `go run ./cmd/foo` instead of `go build` to run programs, to avoid leaving behind build artifacts.

- To test if things compile, you can use `go test ./...` which checks if it compiles and also runs the tests.

- Prefer to use `gopls` for minor refactoring where possible.

- Run `./run.sh help` to see managements commands for this repo. Do `./run.sh check-deps` after a new clone to see if you have all the tools needed.
