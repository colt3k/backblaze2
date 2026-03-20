# Developer Guide

## Scope

This module wraps a focused slice of the Backblaze B2 v2 API. Public behavior lives in the root `backblaze2` package; JSON payload types live in `b2api/`; non-exported helpers stay under `internal/`.

## Working Layout

- `base.go`: client bootstrap, auth refresh, token cache path helpers, retry backoff helpers
- `bucket.go`: bucket list/create/update/delete calls
- `key.go`: application key list/create/delete calls
- `file.go`: upload, multipart upload, download, large-file lifecycle, and upload manifest persistence
- `internal/caller/`: request marshaling and HTTP execution
- `internal/env/`: token and multipart state on disk

## Development Workflow

1. Format first: `go fmt ./...`
2. Prefer module mode while dependencies are being updated: `go test -mod=mod ./...`
3. If you change dependencies, run `go mod vendor` and commit the refreshed `vendor/` tree.
4. Use `examples/examples.go` for manual API experiments, but confirm the signatures in the root package before copying code from it.

## Adding or Updating an Endpoint

1. Add the URI constant in `internal/uri/b2URI.go`.
2. Add or update request/response structs in `b2api/`.
3. Add a permission gate in `perms/auth.go` if the endpoint maps to a new capability.
4. Add the `*Cloud` method in the root package using `auth.BuildAuthMap`, `caller.UnMarshalRequest`, and `caller.MakeCall`.
5. Preserve the existing retry pattern for `bad_auth_token`, `expired_auth_token`, `misc_error`, and `5xx` responses.

## Auth and Local State

`CloudStore` authorizes immediately and persists the token at `~/<appName>/token.json`. Multipart uploads write a resumable manifest to `~/.cloudstore/<bucket>/upload`. That file is useful for inspection and partial recovery, but there is not yet a dedicated CLI resume command in this repository.

## Known Issues in This Checkout

- `vendor/modules.txt` is out of sync with `go.mod`
- `internal/caller/processor.go` still uses older timeout helper names from `github.com/colt3k/utils/netut/hc`
- `examples/examples.go` lags behind current method signatures
