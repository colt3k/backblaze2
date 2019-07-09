package b2api

import "encoding/json"

// B2ListBucketReq b2 bucket list request
type ListBucketReq struct {
	AccountID   string   `json:"accountId"`             // required
	BucketID    string   `json:"bucketId,omitempty"`    // filter (optional)
	BucketName  string   `json:"bucketName,omitempty"`  // filter (optional)
	BucketTypes []string `json:"bucketTypes,omitempty"` // BucketType .Id or .String or Types
}
// B2CreateBucketReq create bucket request
type CreateBucketReq struct {
	AccountID      string           `json:"accountId"`                // required
	BucketName     string           `json:"bucketName"`               // required
	BucketType     string           `json:"bucketType"`               // BucketType .Id or .String or Types
	BucketInfo     string           `json:"bucketInfo,omitempty"`     // optional
	CorsRules      []CorsRules      `json:"corsRules,omitempty"`      // https://www.backblaze.com/b2/docs/cors_rules.html
	LifecycleRules []LifecycleRules `json:"lifecycleRules,omitempty"` // https://www.backblaze.com/b2/docs/lifecycle_rules.html
}
// B2UpdateBucketReq update request
type UpdateBucketReq struct {
	AccountID      string           `json:"accountId"`                // required
	BucketID       string           `json:"bucketId,omitempty"`       // BucketId used instead of name
	BucketType     string           `json:"bucketType,omitempty"`     // BucketType .Id or .String or Types
	BucketInfo     string           `json:"bucketInfo,omitempty"`     // optional
	CorsRules      []CorsRules      `json:"corsRules,omitempty"`      // https://www.backblaze.com/b2/docs/cors_rules.html
	LifecycleRules []LifecycleRules `json:"lifecycleRules,omitempty"` // https://www.backblaze.com/b2/docs/lifecycle_rules.html
	IfRevisionIs   bool             `json:"ifRevisionIs,omitempty"`   //
}
// B2DeleteBucketReq delete bucket request
type DeleteBucketReq struct {
	AccountID string `json:"accountId"` // required
	BucketID  string `json:"bucketId"`  // filter (optional)
}
// B2DeleteBucketResp delete bucket response
type DeleteBucketResp struct {
	Bucket
}
// B2ListBucketsResp list buckets response
type ListBucketsResp struct {
	Buckets []Bucket `json:"buckets"`
}
// B2CreateBucketResp create bucket response
type CreateBucketResp struct {
	Bucket
}
// Bucket bucket struct
type Bucket struct {
	AccountID      string           `json:"accountId"`
	BucketID       string           `json:"bucketId"`
	BucketName     string           `json:"bucketName"`
	BucketType     string           `json:"bucketType"` //"allPublic", "allPrivate", "snapshot", or other values in future
	BucketInfo     json.RawMessage  `json:"bucketInfo"` //user data stored with this bucket
	CorsRules      []CorsRules      `json:"corsRules"`  // CORS rules for this bucket.
	LifecycleRules []LifecycleRules `json:"lifecycleRules"`
	Revision       int              `json:"revision"`
}
// CorsRules struct
type CorsRules struct {
	CorsRuleName      string   `json:"corsRuleName"`
	ExposeHeaders     string   `json:"exposeHeaders"`
	MaxAgeSeconds     int      `json:"maxAgeSeconds"`
	AllowedHeaders    []string `json:"allowedHeaders"`
	AllowedOperations []string `json:"allowedOperations"`
	AllowedOrigins    []string `json:"allowedOrigins"`
}
// LifecycleRules struct
type LifecycleRules struct {
	DaysFromHidingToDeleting  int    `json:"daysFromHidingToDeleting"`
	DaysFromUploadingToHiding int    `json:"daysFromUploadingToHiding"`
	FileNamePrefix            string `json:"fileNamePrefix"`
}
