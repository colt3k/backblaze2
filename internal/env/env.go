package env

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/colt3k/nglog/ers/bserr"
	log "github.com/colt3k/nglog/ng"
	"github.com/colt3k/utils/file"
	"github.com/colt3k/utils/file/filemeta"
	"github.com/colt3k/utils/file/filenative"
	"github.com/colt3k/utils/io"

	"github.com/colt3k/backblaze2/b2api"
)

const (
	// UploadFolder exported folder name
	UploadFolder = "upload"
	tokenName = "token.json"
)

func BaseDir(appName string) string {
	home := file.HomeFolder()
	appPath := filepath.Join(home, appName)
	return appPath
}
// BuildBaseDir build directory
func BuildBaseDir(appName string) string {
	path := BaseDir(appName)
	if _, err := os.Stat(path); err != nil {
		err = os.MkdirAll(path, os.ModePerm)
		bserr.StopErr(err)
	}
	return path
}
func tokenValid(appName string, clear bool) (*string,bool) {
	dir := BuildBaseDir(appName)
	path := filepath.Join(dir,tokenName)
	f := filenative.NewFile(path)
	if f.Available() {
		meta := filemeta.New(f)
		n := time.Now()
		n.Add(20+time.Hour)
		t := time.Unix(0, meta.LastMod() * int64(time.Millisecond))
		//log.Logln(log.DEBUG, "Now+20", n.String())
		//log.Logln(log.DEBUG, "Original", t.String())
		if t.After(n) || clear {
			err := os.Remove(path)
			if err != nil && !strings.Contains(err.Error(), "no such file or directory"){
				log.Logf(log.ERROR, "issue removing old token file %+v", err)
			}
			return nil,false
		}
		return &path,true
	}
	return nil,false
}

func Token(appName string, clear bool) *b2api.AuthorizationResp {
	if path,ok := tokenValid(appName, clear); ok {
		fo, err := os.Open(*path)
		defer fo.Close()
		if err != nil {
			return nil
		}
		dat,err := ioutil.ReadAll(fo)
		if err != nil {
			return nil
		}
		resp := &b2api.AuthorizationResp{}
		err = json.Unmarshal(dat, resp)
		if err != nil {
			log.Logln(log.ERROR, "issue unmarshalling token file")
		}
		return resp
	}
	return nil
}

func WriteToken(appName string, data []byte) error {
	dir := BuildBaseDir(appName)
	path := filepath.Join(dir,tokenName)
	_,err := io.WriteOut(data, path)
	if err != nil {
		return err
	}
	return nil
}

// BucketDir build bucket path
func BucketDir(appName, bucketName string) string {
	home := file.HomeFolder()
	appPath := filepath.Join(home, appName)
	bktPath := filepath.Join(appPath, bucketName)

	return bktPath
}

// BuildBucketDir build directory
func BuildBucketDir(appName, bucketName string) string {
	path := BucketDir(appName, bucketName)
	if _, err := os.Stat(path); err != nil {
		err = os.MkdirAll(path, os.ModePerm)
		bserr.StopErr(err)
	}
	return path
}

