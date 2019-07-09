package b2api

type AuthConfig struct {
	AccountID     string
	ApplicationID string
	Clear         bool
	AppName       string
}

// B2Response contains the response from the B2Cloud
type AuthorizationResp struct {
	AbsoluteMinimumPartSize int64   `json:"absoluteMinimumPartSize"` // smallest possible size part of a large file except last one
	AccountID               string  `json:"accountId"`
	Allowed                 Allowed `json:"allowed"` // base url for all calls except up/download
	APIURL                  string  `json:"apiUrl"`
	AuthorizationToken      string  `json:"authorizationToken"`  // valid for 24hr
	DownloadURL             string  `json:"downloadUrl"`         // base url for downloading files
	RecommendedPartSize     int64   `json:"recommendedPartSize"` // recommended size for each part of a large file
	//MinimumPartSize			int64	`json:"minimumPartSize"`	// deprecated by recommendedPartSize
}

// Valid is authorization valid
func (a *AuthorizationResp) Valid() bool {
	if a != nil && len(a.APIURL) > 0 {
		return true
	}
	return false
}

// Allowed lists capabilities and any restrictions
type Allowed struct {
	BucketId   string   `json:"bucketId"`     // when present, access is restricted to one bucket
	BucketName string   `json:"bucketName"`   // if it exists and bucketId is present this will be the name
	Capability []string `json:"capabilities"` // listKeys, writeKeys, deleteKeys, listBuckets, writeBuckets,
	// deleteBuckets, listFiles, readFiles, shareFiles, writeFiles,
	// and deleteFiles
	NamePrefix string `json:"namePrefix"` // when present, access restricted to names that start with this
}
