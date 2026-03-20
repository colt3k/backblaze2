# Schema Reference

## Package Layout

All wire models live in `b2api/`. The root package uses those structs directly and does not maintain a second domain model.

## Core Auth Models

### `b2api.AuthConfig`

- `AccountID`: Backblaze account ID
- `ApplicationID`: application key or master key ID
- `Clear`: forces token refresh on next auth
- `AppName`: local cache directory name

### `b2api.AuthorizationResp`

- `APIURL`: base URL for most API calls
- `DownloadURL`: base URL for download calls
- `AuthorizationToken`: 24-hour auth token
- `Allowed`: capability and scope restrictions for the token
- `RecommendedPartSize`: server guidance for multipart uploads

### `b2api.Allowed`

- `BucketId`, `BucketName`: optional bucket scope
- `Capability`: permissions such as `listBuckets`, `writeFiles`, `readFiles`, `shareFiles`
- `NamePrefix`: optional key prefix restriction

## Bucket and Key Models

- `b2api.Bucket`: account, name, type, `bucketInfo`, CORS rules, lifecycle rules, and revision
- `b2api.Key`: application key metadata and scope restrictions
- `b2api.CorsRules`: allowed origins, operations, headers, and cache lifetime
- `b2api.LifecycleRules`: hide/delete timing by prefix

## File and Upload Models

### `b2api.UploadResp`

The canonical file record returned by upload, hide, large-file start, and file-info endpoints. Important fields are `FileID`, `FileName`, `ContentLength`, `ContentSha1`, `ContentType`, `FileInfo`, and `UploadTimestamp`.

### `b2api.FileInfo`

- `src_last_modified_millis`: source file modification time
- `large_file_sha1`: full-object SHA1 for multipart uploads

### `backblaze2.Upload`

Persisted multipart state written to `~/.cloudstore/<bucket>/upload`:

```json
{
  "bucket": "example-bucket",
  "filepath": "/tmp/archive.tar",
  "override_path": "snapshots/archive.tar",
  "uploadid": "4_z...",
  "fileId": "4_z...",
  "parts": [
    {"part": 1, "start": 0, "end": 10485760, "size": 10485760, "etag": "sha1"}
  ],
  "total_parts_count": 3
}
```

### `backblaze2.UploaderPart`

Stores the part number, byte range, part size, and the returned SHA1 for that uploaded part.

## Pagination and Limits

- `b2api.MaxFileCount` is used on list requests and defaults to `100`
- Small uploads use one request; multipart uploads split into 10 MiB or 100 MiB chunks depending on total part count
- Multipart download splits files above `DownSplitThresh` (`200000000` bytes)
