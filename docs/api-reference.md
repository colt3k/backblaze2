# API Reference

## Client Lifecycle

### `CloudStore(accountId, appId, appName string) *Cloud`

Creates a client, authorizes immediately, and caches the returned token under `~/<appName>/token.json`.

### `(*Cloud) AuthAccount()`

Refreshes authorization when the cache is missing, stale, or explicitly cleared.

## Bucket Operations

All bucket methods require capabilities from `AuthorizationResp.Allowed.Capability`.

- `ListBuckets(bucketId, bucketName string, bucketType []b2api.BucketType)`
- `CreateBucket(bucketName string, bucketType b2api.BucketType, bucketInfo map[string]interface{}, corsRules []b2api.CorsRules, lifeCycleRules []b2api.LifecycleRules)`
- `UpdateBucket(bucketId string, bucketType b2api.BucketType, bucketInfo map[string]interface{}, corsRules []b2api.CorsRules, lifeCycleRules []b2api.LifecycleRules)`
- `DeleteBucket(bucketId string)`

`BucketType` values are `all`, `allPublic`, `allPrivate`, and `snapshot`.

## Key Operations

- `ListKeys()`
- `CreateKey(keyName, keyBucket string, capabilities []string)`
- `DeleteKey(keyId string)`

Keys are modeled by `b2api.Key` and `b2api.CreateKeyResp`.

## File Operations

- `UploadURL(bucketId string)`
- `UploadFile(bucketId string, up *Upload)`
- `UploadVirtualFile(bucketId, fname string, data []byte, lastMod int64)`
- `ListFiles(bucketId, filename, startFileName string, qty int)`
- `ListFileVersions(bucketId, fileName, startFileName, startFileID string, qty int)`
- `DeleteFile(fileName, fileID string)`
- `HideFile(bucketId, fileName string)`
- `GetFileInfo(fileID string)`
- `GetDownloadAuth(bucketID, filenamePrefix string, validDurationInSeconds int64)`
- `DownloadByName(bucketName, fileName string)`
- `DownloadByID(fileID, byteRange string)`
- `MultipartDownloadById(fileID, localFilePath, fileNameOverride string)`

`DownloadByName` and `DownloadByID` return a header-and-body map from `internal/caller`, including `body` and any response headers captured by the HTTP client.

## Large File Operations

- `StartLargeFile(bucketID, fileInfo string, up *Upload)`
- `GetUploadPartURL(fileID string)`
- `ListPartsURL(fileID string, startPartNo, maxPartCount int64)`
- `ListUnfinishedLargeFiles(bucketID string)`
- `FinishLargeFileUpload(fileId string, sha1Array []string)`
- `CancelLargeFile(fileId string)`
- `SendParts(up *Upload)`

`Upload.Process` is the main entrypoint for callers. It looks up the bucket by name, decides whether multipart upload is needed, writes upload progress to disk, and finishes the large file when all parts have an SHA1.

## Upload State Helpers

- `NewUploader(bucketName, filepath, overridepath string) *Upload`
- `(*Upload) Available() bool`
- `Load(filepath string) *Upload`
- `(*Upload) Process(c *Cloud) (string, error)`
- `(*Upload) Completed() bool`

Use `overridepath` when the remote B2 object path should differ from the local file name.
