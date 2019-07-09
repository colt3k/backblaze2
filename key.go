package backblaze2

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/colt3k/backblaze2/b2api"
	"github.com/colt3k/backblaze2/errs"
	"github.com/colt3k/backblaze2/internal/auth"
	"github.com/colt3k/backblaze2/internal/caller"
	"github.com/colt3k/backblaze2/internal/uri"
	"github.com/colt3k/backblaze2/perms"
)

// ListKeys list account keys
func ListKeys(authConfig b2api.AuthConfig) (*b2api.KeysResp, errs.Error) {
	authd, err := auth.AuthorizeAccount(authConfig)
	if err != nil {
		return nil, err
	}

	if perms.ListKeys(authd) {
		header := auth.BuildAuthMap(authd.AuthorizationToken)

		req := b2api.ListKeyReq{
			AccountId: authd.AccountID,
		}

		msg, errUn := caller.UnMarshalRequest(req)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		mapData, er := caller.MakeCall("POST", authd.APIURL+uri.B2ListKeys, bytes.NewReader(msg), header)
		if er != nil {
			return nil, er
		}

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
func CreateKey(authConfig b2api.AuthConfig, keyName, keyBucket string, capabilities []string) (*b2api.CreateKeyResp, errs.Error) {
	authd, err := auth.AuthorizeAccount(authConfig)
	if err != nil {
		return nil, err
	}

	if perms.CreateKeys(authd) {
		header := auth.BuildAuthMap(authd.AuthorizationToken)

		var req *b2api.CreateKeyReq
		if len(keyBucket) > 0 {
			req = &b2api.CreateKeyReq{
				AccountId:    authd.AccountID,
				Capabilities: capabilities,
				KeyName:      keyName,
				BucketId:     keyBucket,
			}
		} else {
			req = &b2api.CreateKeyReq{
				AccountId:    authd.AccountID,
				Capabilities: capabilities,
				KeyName:      keyName,
			}
		}
		msg, errUn := caller.UnMarshalRequest(req)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		mapData, er := caller.MakeCall("POST", authd.APIURL+uri.B2CreateKey, bytes.NewReader(msg), header)
		if er != nil {
			return nil, er
		}

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
func DeleteKey(authConfig b2api.AuthConfig, keyId string) (*b2api.DeleteKeyResp, errs.Error) {
	authd, err := auth.AuthorizeAccount(authConfig)
	if err != nil {
		return nil, err
	}

	if perms.DeleteKeys(authd) {
		header := auth.BuildAuthMap(authd.AuthorizationToken)

		req := &b2api.DeleteKeyReq{
			ApplicationKeyId: keyId,
		}
		msg, errUn := caller.UnMarshalRequest(req)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		mapData, er := caller.MakeCall("POST", authd.APIURL+uri.B2DeleteKey, bytes.NewReader(msg), header)
		if er != nil {
			return nil, er
		}

		var b2Response b2api.DeleteKeyResp
		errUn = json.Unmarshal(mapData["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}
