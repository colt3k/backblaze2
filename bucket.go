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
func (c *Cloud) ListBuckets(bucketId, bucketName string, bucketType []b2api.BucketType) (*b2api.ListBucketsResp, errs.Error) {

	if perms.ListBuckets(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		bucketTypes := make([]string, 0)
		for _, d := range bucketType {
			bucketTypes = append(bucketTypes, d.String())
		}
		req := &b2api.ListBucketReq{
			AccountID:   c.AuthResponse.AccountID,
			BucketID:    bucketId,
			BucketName:  bucketName,
			BucketTypes: bucketTypes,
		}
		msg, errUn := caller.UnMarshalRequest(req)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		mapData, er := caller.MakeCall("POST", c.AuthResponse.APIURL+uri.B2ListBuckets, bytes.NewReader(msg), header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						shortSleep()
					}
					if testServiceUnavail(er){
						longSleep()
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.ListBuckets(bucketId, bucketName, bucketType)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		log.Logln(log.DEBUGX2, "Actual return: ", string(mapData["body"].([]byte)))
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
func (c *Cloud) CreateBucket(bucketName string, bucketType b2api.BucketType,
	bucketInfo map[string]interface{}, corsRules []b2api.CorsRules, lifeCycleRules []b2api.LifecycleRules) (*b2api.CreateBucketResp, errs.Error) {

	if perms.CreateBucket(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)
		req := b2api.CreateBucketReq{
			AccountID:      c.AuthResponse.AccountID,
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
		mapData, er := caller.MakeCall("POST", c.AuthResponse.APIURL+uri.B2CreateBucket, bytes.NewReader(msg), header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						shortSleep()
					}
					if testServiceUnavail(er){
						longSleep()
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.CreateBucket(bucketName, bucketType, bucketInfo, corsRules, lifeCycleRules)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		log.Logln(log.DEBUGX2, "Actual return: ", string(mapData["body"].([]byte)))
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
func (c *Cloud) UpdateBucket(bucketId string, bucketType b2api.BucketType,
	bucketInfo map[string]interface{}, corsRules []b2api.CorsRules, lifeCycleRules []b2api.LifecycleRules) (*b2api.CreateBucketResp, errs.Error) {

	if perms.UpdateBucket(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)
		req := b2api.UpdateBucketReq{
			AccountID:      c.AuthResponse.AccountID,
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
		mapData, er := caller.MakeCall("POST", c.AuthResponse.APIURL+uri.B2UpdateBucket, bytes.NewReader(msg), header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						shortSleep()
					}
					if testServiceUnavail(er){
						longSleep()
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.UpdateBucket(bucketId, bucketType, bucketInfo, corsRules, lifeCycleRules)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		log.Logln(log.DEBUGX2, "Actual return: ", string(mapData["body"].([]byte)))
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
func (c *Cloud) DeleteBucket(bucketId string) (*b2api.DeleteBucketResp, errs.Error) {

	if perms.DeleteBucket(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)
		req := b2api.DeleteBucketReq{
			AccountID: c.AuthResponse.AccountID,
			BucketID:  bucketId,
		}
		msg, errUn := caller.UnMarshalRequest(req)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		log.Logln(log.DEBUG, "Request:", string(msg))

		mapData, er := caller.MakeCall("POST", c.AuthResponse.APIURL+uri.B2DeleteBucket, bytes.NewReader(msg), header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						shortSleep()
					}
					if testServiceUnavail(er){
						longSleep()
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.DeleteBucket(bucketId)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		log.Logln(log.DEBUGX2, "Actual return: ", string(mapData["body"].([]byte)))
		var b2Response b2api.DeleteBucketResp
		errUn = json.Unmarshal(mapData["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}
