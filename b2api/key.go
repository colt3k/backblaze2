package b2api

// B2CreateKeyResp create key response
type CreateKeyResp struct {
	KeyName          string   `json:"keyName"`
	ApplicationKeyId string   `json:"applicationKeyId"`
	ApplicationKey   string   `json:"applicationKey"`
	Capability       []string `json:"capabilities"` // listKeys, writeKeys, deleteKeys, listBuckets, writeBuckets,
	// deleteBuckets, listFiles, readFiles, shareFiles, writeFiles,
	// and deleteFiles
	AccountID           string `json:"accountId"`
	ExpirationTimestamp string `json:"expirationTimestamp"`
	BucketId            string `json:"bucketId"`
	NamePrefix          string `json:"namePrefix"` // when present, access restricted to names that start with this
}

// B2DeleteKeyResp delete key response
type DeleteKeyResp struct {
	Key Key
}

// B2KeysResp keys response
type KeysResp struct {
	NextApplicationKeyId string `json:"nextApplicationKeyId"`
	Keys                 []Key  `json:"keys"`
}

// Key struct
type Key struct {
	KeyName             string   `json:"keyName"`
	ApplicationKeyId    string   `json:"applicationKeyId"`
	Capabilities        []string `json:"capabilities"`
	AccountId           string   `json:"accountId"`
	ExpirationTimestamp string   `json:"expirationTimestamp"`
	BucketId            string   `json:"bucketId"`
	NamePrefix          string   `json:"namePrefix"`
}

// B2ListKeyReq list key request
type ListKeyReq struct {
	AccountId             string `json:"accountId"`
	MaxKeyCount           int    `json:"maxKeyCount,omitempty"`
	StartApplicationKeyId string `json:"startApplicationKeyId,omitempty"`
}

// B2DeleteKeyReq delete key req
type DeleteKeyReq struct {
	ApplicationKeyId string `json:"applicationKeyId"`
}

// B2CreateKeyReq create req
type CreateKeyReq struct {
	AccountId              string   `json:"accountId"`
	Capabilities           []string `json:"capabilities,omitempty"`
	KeyName                string   `json:"keyName,omitempty"`
	ValidDurationInSeconds int      `json:"validDurationInSeconds,omitempty"`
	BucketId               string   `json:"bucketId,omitempty"`
	NamePrefix             string   `json:"namePrefix,omitempty"`
}
