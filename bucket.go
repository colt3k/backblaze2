package backblaze2

import (
	"bytes"
	"encoding/json"
	"fmt"

	log "github.com/colt3k/nglog/ng"

	"github.com/colt3k/backblaze2/b2api"
	"github.com/colt3k/backblaze2/errs"
	"github.com/colt3k/backblaze2/internal/auth"
	"github.com/colt3k/backblaze2/internal/caller"
	"github.com/colt3k/backblaze2/internal/uri"
	"github.com/colt3k/backblaze2/perms"
)

// ListBuckets list out buckets for account
func ListBuckets(authConfig b2api.AuthConfig, bucketId, bucketName string, bucketType []b2api.BucketType) (*b2api.ListBucketsResp, errs.Error) {

	authd, err := auth.AuthorizeAccount(authConfig)
	if err != nil {
		return nil, err
	}

	if perms.ListBuckets(authd) {
		header := auth.BuildAuthMap(authd.AuthorizationToken)

		bucketTypes := make([]string, 0)
		for _, d := range bucketType {
			bucketTypes = append(bucketTypes, d.String())
		}
		req := &b2api.ListBucketReq{
			AccountID:   authd.AccountID,
			BucketID:    bucketId,
			BucketName:  bucketName,
			BucketTypes: bucketTypes,
		}
		msg, errUn := caller.UnMarshalRequest(req)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		mapData, er := caller.MakeCall("POST", authd.APIURL+uri.B2ListBuckets, bytes.NewReader(msg), header)
		if er != nil {
			return nil, er
		}
		log.Logln(log.DEBUG, "Actual return: ", string(mapData["body"].([]byte)))
		var b2Response b2api.ListBucketsResp
		errUn = json.Unmarshal(mapData["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// CreateBucket create a bucket on account
func CreateBucket(authConfig b2api.AuthConfig, bucketName string, bucketType b2api.BucketType,
	bucketInfo map[string]interface{}, corsRules []b2api.CorsRules, lifeCycleRules []b2api.LifecycleRules) (*b2api.CreateBucketResp, errs.Error) {

	authd, err := auth.AuthorizeAccount(authConfig)

	if err != nil {
		return nil, err
	}

	if perms.CreateBucket(authd) {
		header := auth.BuildAuthMap(authd.AuthorizationToken)
		req := b2api.CreateBucketReq{
			AccountID:      authd.AccountID,
			BucketName:     bucketName,
			BucketType:     bucketType.String(),
			CorsRules:      corsRules,
			LifecycleRules: lifeCycleRules,
		}
		msg, errUn := caller.UnMarshalRequest(req)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		log.Logln(log.DEBUG, "Request:", string(msg))
		mapData, er := caller.MakeCall("POST", authd.APIURL+uri.B2CreateBucket, bytes.NewReader(msg), header)
		if er != nil {
			return nil, er
		}
		log.Logln(log.DEBUG, "Actual return: ", string(mapData["body"].([]byte)))
		var b2Response b2api.CreateBucketResp
		errUn = json.Unmarshal(mapData["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// UpdateBucket update bucket properties
func UpdateBucket(authConfig b2api.AuthConfig, bucketId string, bucketType b2api.BucketType,
	bucketInfo map[string]interface{}, corsRules []b2api.CorsRules, lifeCycleRules []b2api.LifecycleRules) (*b2api.CreateBucketResp, errs.Error) {

	authd, err := auth.AuthorizeAccount(authConfig)

	if err != nil {
		return nil, err
	}

	if perms.UpdateBucket(authd) {
		header := auth.BuildAuthMap(authd.AuthorizationToken)
		req := b2api.UpdateBucketReq{
			AccountID:      authd.AccountID,
			BucketID:       bucketId,
			BucketType:     bucketType.String(),
			CorsRules:      corsRules,
			LifecycleRules: lifeCycleRules,
		}
		msg, errUn := caller.UnMarshalRequest(req)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		log.Logln(log.DEBUG, "Request:", string(msg))
		mapData, er := caller.MakeCall("POST", authd.APIURL+uri.B2UpdateBucket, bytes.NewReader(msg), header)
		if er != nil {
			return nil, er
		}
		log.Logln(log.DEBUG, "Actual return: ", string(mapData["body"].([]byte)))
		var b2Response b2api.CreateBucketResp
		errUn = json.Unmarshal(mapData["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// DeleteBucket remove from account
func DeleteBucket(authConfig b2api.AuthConfig, bucketId string) (*b2api.DeleteBucketResp, errs.Error) {

	authd, err := auth.AuthorizeAccount(authConfig)
	if err != nil {
		return nil, err
	}

	if perms.DeleteBucket(authd) {
		header := auth.BuildAuthMap(authd.AuthorizationToken)
		req := b2api.DeleteBucketReq{
			AccountID: authd.AccountID,
			BucketID:  bucketId,
		}
		msg, errUn := caller.UnMarshalRequest(req)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		log.Logln(log.DEBUG, "Request:", string(msg))

		mapData, er := caller.MakeCall("POST", authd.APIURL+uri.B2DeleteBucket, bytes.NewReader(msg), header)
		if er != nil {
			return nil, er
		}
		log.Logln(log.DEBUG, "Actual return: ", string(mapData["body"].([]byte)))
		var b2Response b2api.DeleteBucketResp
		errUn = json.Unmarshal(mapData["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}
