package backblaze2

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	log "github.com/colt3k/nglog/ng"
	"github.com/colt3k/utils/encode"
	"github.com/colt3k/utils/encode/encodeenum"
	"github.com/colt3k/utils/file"

	"github.com/colt3k/backblaze2/b2api"
	"github.com/colt3k/backblaze2/errs"
	"github.com/colt3k/backblaze2/internal/caller"
	"github.com/colt3k/backblaze2/internal/env"
	"github.com/colt3k/backblaze2/internal/uri"
)

const (
	AppFolderName = ".cloudstore"
	// UploadFolder exported folder name
	UploadFolder = "upload"
	MaxAuthTry  = 3
)
var AuthCounter = 0

type Cloud struct {
	AuthConfig   b2api.AuthConfig
	AuthResponse *b2api.AuthorizationResp
}

func CloudStore(accountId, appId string) *Cloud {
	t := new(Cloud)
	t.AuthConfig = b2api.AuthConfig{AccountID: accountId, ApplicationID: appId, Clear: false, AppName: "example"}
	t.AuthAccount()
	return t
}

func (c *Cloud) AuthAccount() {
	// if it exists and is less than 24hrs old use it first, otherwise renew it
	token := env.Token(c.AuthConfig.AppName, c.AuthConfig.Clear)
	if token != nil {
		//log.Logln(log.DEBUG, "returning existing token")
		c.AuthResponse = token
		return
	}

	header := make(map[string]string)
	acct := []byte(c.AuthConfig.AccountID + ":" + c.AuthConfig.ApplicationID)
	encAcct := "Basic " + encode.Encode(acct, encodeenum.B64STD)
	header["Authorization"] = encAcct

	if val, ok := os.LookupEnv("X-Bz-Test-Mode"); ok {
		//header["X-Bz-Test-Mode"] = "expire_some_account_authorization_tokens"
		//header["X-Bz-Test-Mode"] = "fail_some_uploads"
		//header["X-Bz-Test-Mode"] = "force_cap_exceeded"
		header["X-Bz-Test-Mode"] = val
		log.Logf(log.INFO, "X-Bz-Test-Mode %s enabled", val)
	}

	log.Logln(log.DEBUG, "obtaining new token")
	mapData, ers := caller.MakeCall("GET", uri.B2AuthAccount, nil, header)
	if ers != nil {
		if testRetryErr(ers) {
			// delete it and call again
			AuthCounter += 1
			if AuthCounter <= MaxAuthTry {
				if AuthCounter > 1 {
					sleep := 3 * time.Second
					jitter := time.Duration(rand.Int63n(int64(sleep)))
					sleep = sleep + jitter/2
					time.Sleep(sleep)
				}
				if ers.Code() == "service_unavailable" {
					log.Logln(log.WARN,"service unavailable trying again, please stand by")
					sleep := 7 * time.Second
					jitter := time.Duration(rand.Int63n(int64(sleep)))
					sleep = sleep + jitter/2
					time.Sleep(sleep)
				}
				c.AuthConfig.Clear = true
				c.AuthAccount()
				return
			}
		}
		panic(ers)
	}
	b2Response := &b2api.AuthorizationResp{}
	errUn := json.Unmarshal(mapData["body"].([]byte), b2Response)
	if errUn != nil {
		panic(errs.New(errUn, ""))
	}
	if b2Response != nil && len(b2Response.AuthorizationToken) > 0 {
		//write out
		env.WriteToken(c.AuthConfig.AppName, mapData["body"].([]byte))
	}
	c.AuthResponse = b2Response
}

// UploaderDir build uploader directory
func UploaderDir(bucketName string) string {
	home := file.HomeFolder()
	appPath := filepath.Join(home, AppFolderName)
	bktPath := filepath.Join(appPath, bucketName)
	uploadsFile := filepath.Join(bktPath, UploadFolder)

	return uploadsFile
}

func testRetryErr(er errs.Error) bool {
	if er.Code() == "bad_auth_token" || er.Code() == "expired_auth_token" || er.Code() == "service_unavailable" || er.Code() == "misc_error" {
		if er.Code() == "bad_auth_token" || er.Code() == "expired_auth_token" || er.Code() == "misc_error" {
			log.Logf(log.WARN,"%s: trying again", er.Code())
		}
		return true
	}
	return false
}