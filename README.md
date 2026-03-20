# backblaze2

`backblaze2` is a Go client for Backblaze B2 Cloud Storage. The root package exposes account authorization plus bucket, key, file, multipart upload, and multipart download helpers, while `b2api/` contains the request and response models used on the wire.

## Documentation Map

- [Developer Guide](docs/developer-guide.md)
- [API Reference](docs/api-reference.md)
- [Schema Reference](docs/schema-reference.md)
- [Data Flow](docs/dataflow.md)
- [Operator Runbook](docs/operator-runbook.md)

## Repository Layout

- `base.go`, `bucket.go`, `key.go`, `file.go`: public client methods on `*Cloud`
- `b2api/`: request and response structs plus the `BucketType` enum
- `internal/`: auth header construction, HTTP caller, URI constants, and token/file-state helpers
- `perms/`: capability checks derived from `AuthorizationResp.Allowed`
- `errs/`: B2 error normalization
- `examples/examples.go`: scratch integration example driven by environment variables

## Quick Start

```go
package main

import (
	"fmt"
	"os"

	"github.com/colt3k/backblaze2"
)

func main() {
	c := backblaze2.CloudStore(os.Getenv("ACCT_ID"), os.Getenv("APP_ID"), "ctexample")

	buckets, err := c.ListBuckets("", "", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("bucket count:", len(buckets.Buckets))
}
```

For uploads, build an `Upload` and let `Process` choose single-part or multipart handling:

```go
up := backblaze2.NewUploader("my-bucket", "/tmp/report.csv", "")
fileID, err := up.Process(c)
```

## Development Commands

- `go fmt ./...`: format the module
- `go test -mod=mod ./...`: run tests without trusting the stale `vendor/` tree
- `go build -mod=mod ./...`: compile packages in module mode
- `go run -mod=mod ./examples`: run the integration example with `ACCT_ID`, `APP_ID`, and `BUCKET`
- `go mod vendor`: resync `vendor/` after dependency changes

## Current Caveats

- `vendor/modules.txt` is behind `go.mod`, so plain `go test ./...` and `go build ./...` fail until `vendor/` is refreshed.
- The current checkout also has compile drift after dependency updates. `internal/caller/processor.go` references timeout helpers that were renamed upstream, and `examples/examples.go` still uses older method signatures.

## Local State

The client caches the auth token under `~/<appName>/token.json`. Multipart upload state is written to `~/.cloudstore/<bucket>/upload`, which captures the generated `fileId`, part ranges, and uploaded part SHA1 values.

## Backblaze Test Modes

Backblaze exposes several test-only headers that are useful when validating retry logic and error handling. The original topics are kept here, with one important note: the current library only injects `X-Bz-Test-Mode` during `b2_authorize_account`, so upload-specific scenarios may require temporarily wiring the header into upload calls during local debugging.

### Testing Failures on Upload

Set `X-Bz-Test-Mode: fail_some_uploads` on upload-related API calls to trigger intermittent failures and exercise retry behavior.

### Testing Authorization Failures

Set `X-Bz-Test-Mode: expire_some_account_authorization_tokens` to force intermittent authorization expiry and validate token refresh handling.

### Testing Upload 403 Forbidden

Set `X-Bz-Test-Mode: force_cap_exceeded` before upload-related calls to simulate storage-cap enforcement.

### Retry Header

Backblaze can return `429 Too Many Requests` with a `Retry-After` header describing how long to wait before retrying. This library already retries several transient failures with exponential backoff, but it does not currently parse `Retry-After` directly.
