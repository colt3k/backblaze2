# Repository Guidelines

## Project Structure & Module Organization
`backblaze2` is a single Go module. Root files such as `base.go`, `bucket.go`, `file.go`, and `key.go` expose the public client for auth, bucket, file, and key operations. `b2api/` contains request/response structs and enums that mirror Backblaze B2 payloads. `internal/` holds non-exported helpers for HTTP calls, auth/token caching, URIs, and environment handling. `errs/` centralizes API error mapping, `perms/` defines capability helpers, and `examples/examples.go` is the runnable integration sample.

## Build, Test, and Development Commands
Use Go 1.24.x; `go.mod` pins `toolchain go1.24.10`.

- `go test -mod=mod ./...` runs all packages in module mode; use this until `vendor/` is resynced with `go.mod`.
- `go build -mod=mod ./...` compiles the library and the example program.
- `go run -mod=mod ./examples` exercises the sample client; set `ACCT_ID`, `APP_ID`, and `BUCKET` first.
- `go fmt ./...` formats packages before review.
- `go mod vendor` refreshes the committed `vendor/` tree after dependency changes.

## Coding Style & Naming Conventions
Follow standard Go formatting with tabs via `gofmt`. Keep package names short and lowercase (`b2api`, `errs`, `perms`); exported types and methods use PascalCase, and unexported helpers use camelCase. Keep B2 request/response structs in `b2api/`, public client methods in the root package, and private transport/cache helpers under `internal/`. Preserve the existing pattern of small focused methods on `*Cloud`.

## Testing Guidelines
There are currently no committed `_test.go` files, so new changes should add table-driven tests alongside the touched package. Name files `*_test.go` and functions `TestXxx`. For API-facing changes, pair unit coverage with a runnable example or a guarded integration path in `examples/`. If you update dependencies, rerun `go test -mod=mod ./...` and refresh `vendor/` before opening a PR.

## Commit & Pull Request Guidelines
Recent history uses short, imperative, lowercase subjects such as `update deps` and `fix dep vuln`. Keep commit titles concise and scoped to one change. PRs should describe the B2 behavior affected, note any required env vars or migration steps, and include the build/test commands you ran. Add request/response snippets only when they clarify API changes.

## Security & Configuration Tips
Never commit Backblaze credentials. The example reads `ACCT_ID`, `APP_ID`, and `BUCKET` from the environment, and auth tokens are cached under the caller's home directory, for example `~/ctexample/token.json`. Clear local token state when debugging authorization issues.
