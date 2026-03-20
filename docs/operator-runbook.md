# Operator Runbook

## Prerequisites

- Go 1.24.x
- Backblaze account ID in `ACCT_ID`
- Backblaze application key ID or master key in `APP_ID`
- Bucket name in `BUCKET` when using `examples/examples.go`

## Safe Local Commands

- `go fmt ./...`
- `go test -mod=mod ./...`
- `go build -mod=mod ./...`
- `go run -mod=mod ./examples`

Use module mode until `vendor/` is resynced. In the current checkout, expect build failures until the renamed timeout helpers in `internal/caller/processor.go` and the stale example signatures are fixed.

## Common Operations

### Clear Cached Authorization

Delete `~/<appName>/token.json` or set `Cloud.AuthConfig.Clear = true` before calling `AuthAccount`.

### Inspect Multipart Upload State

Read `~/.cloudstore/<bucket>/upload` to see:

- the B2 `fileId`
- the local source path
- the remote override path
- uploaded parts and their SHA1 values

### Exercise Failure Paths

Export one of the Backblaze test modes before auth:

- `X-Bz-Test-Mode=expire_some_account_authorization_tokens`
- `X-Bz-Test-Mode=fail_some_uploads`
- `X-Bz-Test-Mode=force_cap_exceeded`

The library currently forwards this header during authorization; upload-only scenarios may need a temporary local patch to send the header with upload requests.

## Troubleshooting

### `not allowed`

The token lacks the needed capability. Compare the failed method with `AuthorizationResp.Allowed.Capability` and the checks in `perms/auth.go`.

### `429 too_many_requests`

Back off and retry. Backblaze may include a `Retry-After` header, but the client does not parse it explicitly yet.

### `expired_auth_token` or `bad_auth_token`

Clear the cached token and rerun the operation. Most public methods already reauthorize automatically.

### Partial Multipart Upload

Use the saved upload manifest to inspect completed parts. The repository persists enough state for manual recovery analysis, but it does not yet ship a dedicated resume tool or operator command.
