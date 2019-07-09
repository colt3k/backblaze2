package b2api

import (
	"encoding/json"
)

// MaxFileCount type for default
type MaxFileCount int

// UnmarshalJSON Overrides and sets default to 100 on custom type
func (m *MaxFileCount) UnmarshalJSON(b []byte) error {
	var i int
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}
	*m = 100

	return nil
}

// B2ListUnfinishUpLgFilesReq struct
type ListUnfinishUpLgFilesReq struct {
	BucketId     string       `json:"bucketId"`
	NamePrefix   string       `json:"namePrefix,omitempty"`   // optional, only files matching prefix will return
	StartFileId  string       `json:"startFileId,omitempty"`  // optional, first to return, if not found next will return
	MaxFileCount MaxFileCount `json:"maxFileCount,omitempty"` // optional, default 100, max 100
}

// B2ListUnfinishUpLgFilesResp struct
type ListUnfinishUpLgFilesResp struct {
	Files     []File `json:"files"`
	NextField string  `json:"nextField"`
}

// Files struct
type File struct {
	UploadResp
}

// B2FinishUpLgFileReq struct
type FinishUpLgFileReq struct {
	FileId string   `json:"fileId"`        // id returned by B2StartLargeFileResp
	Sha1Ar []string `json:"partSha1Array"` // file part sha1 arrays in correct order
}

// B2FinishUpLgFileResp struct
type FinishUpLgFileResp struct {
	UploadResp
}

// B2CanxUpLgFileReq struct
type CanxUpLgFileReq struct {
	FileId string `json:"fileId"` // id returned by B2StartLargeFileResp
}

// B2CanxUpLgFileResp struct
type CanxUpLgFileResp struct {
	FileId    string `json:"fileId"`
	AccountId string `json:"accountId"`
	BucketId  string `json:"bucketId"`
	FileName  string `json:"fileName"`
}

// B2StartUpLgFileReq struct
type StartUpLgFileReq struct {
	BucketId    string `json:"bucketId"`
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	FileInfo    string `json:"fileInfo"`
}

// B2StartUpLgFileResp struct
type StartUpLgFileResp struct {
	AccountId  string `json:"accountId"`
	UploadResp        // FileId stores unique id to reuse for part uploads
}

// B2UploadURLReq struct
type UploadURLReq struct {
	BucketId string `json:"bucketId"`
}

// B2UploadURLResp struct
type UploadURLResp struct {
	BucketId           string `json:"bucketId"`
	UploadURL          string `json:"uploadUrl"`
	AuthorizationToken string `json:"authorizationToken"`
}

// B2UploadResp struct
type UploadResp struct {
	AccountID       string   `json:"accountId"`
	Action          string   `json:"action"` // start,upload,hide,folder currently but others can be added
	BucketID        string   `json:"bucketId"`
	ContentLength   int64    `json:"contentLength"`
	ContentSha1     string   `json:"contentSha1"`
	ContentType     string   `json:"contentType"`
	FileID          string   `json:"fileId"`
	FileInfo        FileInfo `json:"fileInfo"`
	FileName        string   `json:"fileName"`
	UploadTimestamp int64    `json:"uploadTimestamp"`
}

type FileInfo struct {
	SrcLastModifiedMillis string `json:"src_last_modified_millis"`
	LargeFileSha1         string `json:"large_file_sha1,omitempty"`
}

type ListFileReq struct {
	BucketID      string       `json:"bucketId"`
	StartFileName string       `json:"startFileName,omitempty"`
	MaxFileCount  MaxFileCount `json:"maxFileCount,omitempty"`
	Prefix        string       `json:"prefix,omitempty"`
	Delimiter     string       `json:"delimiter,omitempty"`
}

type ListFilesResponse struct {
	File         []File `json:"files"`
	NextFileName string  `json:"nextFileName"`
}

type ListFileVersionsReq struct {
	BucketID      string       `json:"bucketId"`
	StartFileName string       `json:"startFileName,omitempty"`
	StartFileID   string       `json:"startFileId,omitempty"` // requires StartFileName if defined
	MaxFileCount  MaxFileCount `json:"maxFileCount,omitempty"`
	Prefix        string       `json:"prefix,omitempty"`
	Delimiter     string       `json:"delimiter,omitempty"`
}

type ListFileVersionsResponse struct {
	Files        []File `json:"files"`
	NextFileName string  `json:"nextFileName"`
	NextFileId   string  `json:"nextFileId"`
}

type DeleteFileVersionReq struct {
	FileName string `json:"fileName"`
	FileID   string `json:"fileId"`
}
type DeleteFileVersionResponse struct {
	FileID   string `json:"fileId"`
	FileName string `json:"fileName"`
}

type HideFileReq struct {
	BucketId string `json:"bucketId"`
	FileName string `json:"fileName"`
}
type HideFileResponse struct {
	UploadResp
}

type GetFileInfoReq struct {
	FileID string `json:"fileId"`
}
type GetFileInfoResponse struct {
	UploadResp
}

type DownloadAuthReq struct {
	BucketId               string `json:"bucketId"`                       //
	FilenamePrefix         string `json:"fileNamePrefix"`                 // Directory name below bucket
	ValidDurationInSeconds int64  `json:"validDurationInSeconds"`         // Min 1 second. The maximum value is 604800
	B2ContentDisposition   string `json:"b2ContentDisposition,omitempty"` // Optional RFC 6266
}
type DownloadAuthResponse struct {
	BucketId           string `json:"bucketId"`           //
	FilenamePrefix     string `json:"fileNamePrefix"`     // Directory name below bucket
	AuthorizationToken string `json:"authorizationToken"` // Token to use for download
}

type DownloadFileByIDReq struct {
	FileID string `json:"fileId"`
}

type StartLargeFileReq struct {
	BucketId    string   `json:"bucketId"` //
	FileName    string   `json:"fileName"`
	ContentType string   `json:"contentType"` // use as default b2/x-auto
	FileInfo    FileInfo `json:"fileInfo"`    // Optional
}
type StartLargeFileResponse struct {
	UploadResp
}

type GetFileUploadPartReq struct {
	FileID string `json:"fileId"` // Unique ID of file being uploaded
}
type GetFileUploadPartResponse struct {
	FileID             string `json:"fileId"`             // Unique ID of file being uploaded
	UploadURL          string `json:"uploadUrl"`          // URL that can be used to upload parts of this file
	AuthorizationToken string `json:"authorizationToken"` // Valid 24hrs or until endpoint rejects upload
}

type ListPartsReq struct {
	FileID          string `json:"fileId"` // Unique ID of file being uploaded
	StartPartNumber int64  `json:"startPartNumber"`
	MaxPartCount    int64  `json:"maxPartCount"`
}
type ListPartsResponse struct {
	FileID             string `json:"fileId"`             // Unique ID of file being uploaded
	UploadURL          string `json:"uploadUrl"`          // URL that can be used to upload parts of this file
	AuthorizationToken string `json:"authorizationToken"` // Valid 24hrs or until endpoint rejects upload
	Parts              []Part `json:"parts"`
	NextPartNumber     int64  `json:"nextPartNumber"` // What to pass into startPartNumber for next search to continue
}

type Part struct {
	FileID          string `json:"fileId"` // Unique ID of file being uploaded
	PartNumber      int64  `json:"partNumber"`
	ContentLength   int64  `json:"contentLength"`
	ContentSha1     string `json:"contentSha1"`
	UploadTimestamp int64  `json:"uploadTimestamp"`
}

type ListUnfinishedLargeFilesReq struct {
	BucketId     string       `json:"bucketId"`               //
	NamePrefix   string       `json:"namePrefix,omitempty"`   // Optional
	StartFileId  string       `json:"startFileId,omitempty"`  // Optional
	MaxFileCount MaxFileCount `json:"maxFileCount,omitempty"` // Optional
}

type ListUnfinishedLargeFilesResponse struct {
	Files      []UploadResp `json:"files"`
	NextFileId string       `json:"nextFileId"`
}

type UploadPartResponse struct {
	ContentLength   int64  `json:"contentLength"`
	ContentSha1     string `json:"contentSha1"`
	FileId          string `json:"fileId"`
	PartNumber      int64  `json:"partNumber"`
	UploadTimestamp int64  `json:"uploadTimestamp"`
}
