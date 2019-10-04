package backblaze2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/colt3k/backblaze2/b2api"
	"github.com/colt3k/backblaze2/errs"
	"github.com/colt3k/backblaze2/internal/auth"
	"github.com/colt3k/backblaze2/internal/caller"
	"github.com/colt3k/backblaze2/internal/uri"
	"github.com/colt3k/backblaze2/perms"
)

// ListKeys list account keys
func (c *Cloud) ListKeys() (*b2api.KeysResp, errs.Error) {
	if perms.ListKeys(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		req := b2api.ListKeyReq{
			AccountId: c.AuthResponse.AccountID,
		}

		msg, errUn := caller.UnMarshalRequest(req)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		mapData, er := caller.MakeCall("POST", c.AuthResponse.APIURL+uri.B2ListKeys, bytes.NewReader(msg), header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if testServiceUnavail(er){
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.ListKeys()
				}
			}

			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.KeysResp
		errUn = json.Unmarshal(mapData["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// CreateKey create account key
func (c *Cloud) CreateKey(keyName, keyBucket string, capabilities []string) (*b2api.CreateKeyResp, errs.Error) {

	if perms.CreateKeys(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		var req *b2api.CreateKeyReq
		if len(keyBucket) > 0 {
			req = &b2api.CreateKeyReq{
				AccountId:    c.AuthResponse.AccountID,
				Capabilities: capabilities,
				KeyName:      keyName,
				BucketId:     keyBucket,
			}
		} else {
			req = &b2api.CreateKeyReq{
				AccountId:    c.AuthResponse.AccountID,
				Capabilities: capabilities,
				KeyName:      keyName,
			}
		}
		msg, errUn := caller.UnMarshalRequest(req)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		mapData, er := caller.MakeCall("POST", c.AuthResponse.APIURL+uri.B2CreateKey, bytes.NewReader(msg), header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if testServiceUnavail(er){
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.CreateKey(keyName, keyBucket, capabilities)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.CreateKeyResp
		errUn = json.Unmarshal(mapData["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// DeleteKey delete account key
func (c *Cloud) DeleteKey(keyId string) (*b2api.DeleteKeyResp, errs.Error) {

	if perms.DeleteKeys(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		req := &b2api.DeleteKeyReq{
			ApplicationKeyId: keyId,
		}
		msg, errUn := caller.UnMarshalRequest(req)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		mapData, er := caller.MakeCall("POST", c.AuthResponse.APIURL+uri.B2DeleteKey, bytes.NewReader(msg), header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if testServiceUnavail(er){
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.DeleteKey(keyId)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.DeleteKeyResp
		errUn = json.Unmarshal(mapData["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}
