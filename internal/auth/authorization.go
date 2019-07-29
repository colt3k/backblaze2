package auth

import (
	"encoding/json"
	"math/rand"
	"runtime"
	"strconv"
	"time"

	log "github.com/colt3k/nglog/ng"
	"github.com/colt3k/utils/encode"
	"github.com/colt3k/utils/encode/encodeenum"
	"github.com/colt3k/utils/osut"

	"github.com/colt3k/backblaze2/b2api"
	"github.com/colt3k/backblaze2/errs"
	"github.com/colt3k/backblaze2/internal/caller"
	"github.com/colt3k/backblaze2/internal/env"
	"github.com/colt3k/backblaze2/internal/uri"
)

var AuthCounter = 0
var MaxAuthTry  = 3
var AuthClear   = false

// AuthorizeAccount authorize account and retrieve token
func AuthorizeAccount(auth b2api.AuthConfig) (*b2api.AuthorizationResp, errs.Error) {
	// if it exists and is less than 24hrs old use it first, otherwise renew it
	token := env.Token(auth.AppName, auth.Clear)
	if token != nil {
		log.Logln(log.DEBUG, "returning existing token")
		return token, nil
	}

	header := make(map[string]string)
	acct := []byte(auth.AccountID + ":" + auth.ApplicationID)
	encAcct := "Basic " + encode.Encode(acct, encodeenum.B64STD)
	header["Authorization"] = encAcct

	header["X-Bz-Test-Mode"] = "expire_some_account_authorization_tokens"
	log.Logln(log.DEBUG, "obtaining new token")
	mapData, ers := caller.MakeCall("GET", uri.B2AuthAccount, nil, header)

	if ers != nil {
		if ers.Code() == "bad_auth_token" || ers.Code() == "expired_auth_token" || ers.Code() == "service_unavailable" {
			if ers.Code() == "bad_auth_token" || ers.Code() == "expired_auth_token" {
				log.Printf("%s: trying again", ers.Code())
			}
			// delete it and call again
			AuthCounter += 1
			if AuthCounter <= MaxAuthTry {
				if AuthCounter > 1 {
					sleep := 3*time.Second
					jitter := time.Duration(rand.Int63n(int64(sleep)))
					sleep = sleep + jitter/2
					time.Sleep(sleep)
				}
				if ers.Code() == "service_unavailable" {
					log.Println("service unavailable trying again, please stand by")
					sleep := 7*time.Second
					jitter := time.Duration(rand.Int63n(int64(sleep)))
					sleep = sleep + jitter/2
					time.Sleep(sleep)
				}
				AuthClear = true
				return AuthorizeAccount(auth)
			}
		}
		return nil, ers
	}
	b2Response := &b2api.AuthorizationResp{}
	errUn := json.Unmarshal(mapData["body"].([]byte), b2Response)
	if errUn != nil {
		return nil, errs.New(errUn, "")
	}
	if b2Response != nil && len(b2Response.AuthorizationToken) > 0 {
		//write out
		env.WriteToken(auth.AppName, mapData["body"].([]byte))
	}
	return b2Response, nil
}

// BuildAuthMap build authorization map
func BuildAuthMap(authToken string) map[string]string {
	header := make(map[string]string)
	header["Authorization"] = authToken
	header["charset"] = "utf-8"
	platform := osut.OS()

	header["User-Agent"] = "cloudstore/0.0.1+" + runtime.GOOS + "/" + strconv.Itoa(platform.VersionMajor) + "." + strconv.Itoa(platform.VersionMinor)
	return header
}

