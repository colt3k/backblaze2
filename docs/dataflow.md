# Data Flow

## 1. Authorization

1. `CloudStore` fills `b2api.AuthConfig` and calls `AuthAccount`.
2. `AuthAccount` first checks `internal/env.Token(appName, clear)` for a cached token.
3. If no valid token exists, it calls `b2_authorize_account` through `internal/caller.MakeCall`.
4. The response is unmarshaled into `b2api.AuthorizationResp` and written back to `~/<appName>/token.json`.

## 2. Request Execution

1. Public methods validate the capability through `perms/auth.go`.
2. `internal/auth.BuildAuthMap` creates the `Authorization`, `charset`, and `User-Agent` headers.
3. Request structs are JSON-encoded by `caller.UnMarshalRequest`.
4. `caller.MakeCall` routes everything through `internal/caller.HttpCall`, which returns a map containing `body` plus captured response headers.

## 3. Small Upload

1. `Upload.Process` resolves the destination bucket ID with `ListBuckets`.
2. If the file stays under the multipart thresholds, `UploadFile` requests an upload URL, computes SHA1 and `src_last_modified_millis`, then streams the file body to B2.
3. The upload response is unmarshaled into `b2api.UploadResp`.

## 4. Multipart Upload

1. `Process` calls `StartLargeFile`.
2. `SetupPartSizes` computes `UploaderPart` ranges.
3. `WriteOutFileData2Upload` persists the upload manifest to `~/.cloudstore/<bucket>/upload`.
4. `SendParts` creates a worker pool and requests an upload-part URL per part.
5. Each part upload computes a SHA1, streams the byte range, and stores the returned hash in the manifest.
6. When all parts have hashes, `FinishLargeFileUpload` closes the large-file session.

## 5. Download

1. `DownloadByName` performs a direct download against `DownloadURL`.
2. `DownloadByID` optionally sends a `Range` header for partial reads.
3. `MultipartDownloadById` first calls `GetFileInfo`, then either downloads once or splits the file into `DownloadFileChunk` ranges and writes parts with `WriteAt`.
4. The final local file SHA1 is compared to the remote SHA1.

## 6. Retry and Failure Handling

- Auth, bucket, key, and most file methods retry on `bad_auth_token`, `expired_auth_token`, `misc_error`, `service_unavailable`, and generic `5xx` errors.
- Backoff is exponential with jitter via `shortSleep` and `longSleep`.
- Multipart upload progress is persisted after part completion so operators can inspect partial progress if a run stops mid-transfer.
