package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/colt3k/nglog/ng"
	"github.com/colt3k/utils/encode"
	"github.com/colt3k/utils/encode/encodeenum"
	"github.com/colt3k/utils/file/filenative"
	"github.com/colt3k/utils/hash/hashenum"
	"github.com/colt3k/utils/hash/sha1"
	iout "github.com/colt3k/utils/io"

	"github.com/colt3k/backblaze2"
	"github.com/colt3k/backblaze2/b2api"
)

var (
	ACCT_ID = os.Getenv("ACCT_ID")
	APP_ID  = os.Getenv("APP_ID")
	bucketName = os.Getenv("BUCKET")
	bucketID = ""
	filePath   = ""
	fileID = ""
	fileName = ""
)

func main() {
	if len(ACCT_ID) <= 0 || len(APP_ID) <= 0 {
		fmt.Println("ACCT_ID and/or APP_ID not defined")
		os.Exit(-1)
	}
	//er := ListFileVersions()
	er := UploadFile()
	if er != nil {
		fmt.Printf("%v", er)
	}
}

func setupCloud() *backblaze2.Cloud {
	c := backblaze2.CloudStore(ACCT_ID, APP_ID)

	fmt.Printf("\nAuthToken: %s\n", c.AuthResponse.AuthorizationToken)

	return c
}
func ListKeys() error {
	c := setupCloud()

	r, err := c.ListKeys()
	if err != nil {
		return err
	}
	fmt.Printf("\nKeys: %v\n", r.Keys)

	return nil
}

func CreateKey() error {
	c := setupCloud()

	r, err := c.CreateKey("keyName", "keyBucket", []string{"capabilities"})
	if err != nil {
		return err
	}
	fmt.Printf("\nKeys: %v\n", r.KeyName)

	return nil
}

func DeleteKey() error {
	c := setupCloud()

	r, err := c.DeleteKey("keyid")
	if err != nil {
		return err
	}
	fmt.Printf("\nKeys: %v\n", r.Key)

	return nil
}

func ListBuckets() error {
	c := setupCloud()

	r, err := c.ListBuckets("", "", nil)
	if err != nil {
		return err
	}
	for _, k := range r.Buckets {
		fmt.Printf("\nBucket: %v - %s\n", k.BucketID, k.BucketName)
	}

	return nil
}

func CreateBucket() error {
	c := setupCloud()

	r, err := c.CreateBucket("testbucket", b2api.BucketType(b2api.All.Id("allPrivate")),
		nil, []b2api.CorsRules{}, []b2api.LifecycleRules{})
	if err != nil {
		return err
	}

	fmt.Println("Bucket Created: ", r.BucketName)

	return nil
}

func DeleteBucket() error {
	c := setupCloud()

	r, err := c.DeleteBucket(bucketID)
	if err != nil {
		return err
	}

	fmt.Println("Bucket Deleted: ", r.BucketName)

	return nil
}

func CreateVirtualFile() error {
	c := setupCloud()

	r, err := c.UploadVirtualFile(bucketID, fileName, []byte("test content"), 0)
	if err != nil {
		return err
	}

	fmt.Println("File created: ", r.FileName)

	return nil
}

func ListFileVersions() error {
	c := setupCloud()

	r, err := c.ListFileVersions(bucketID, fileName)
	if err != nil {
		return err
	}

	for _, k := range r.Files {
		fmt.Printf("File Version %v : %s - %s\n", k.UploadTimestamp, k.FileName, k.FileID)
	}

	return nil
}

func ListFiles() error {
	c := setupCloud()

	r, err := c.ListFiles(bucketID, "")
	if err != nil {
		return err
	}

	for _, k := range r.File {
		fmt.Printf("File %v : %s - %s\n", k.UploadTimestamp, k.FileName, k.FileID)
	}

	return nil
}

func UploadFile() error {
	c := setupCloud()

	up := backblaze2.NewUploader(bucketName, filePath, "")
	if up.Available() {
		fileId, err := up.Process(c)
		if err != nil {
			return err
		}
		fmt.Println("Uploaded FileID: ", fileId)
	} else {
		fmt.Println("file not found")
	}

	return nil
}

func DownloadFileByNameLatestOnly() error {
	c := setupCloud()

	localFilePath := "."
	rsp, err := c.DownloadByName(bucketName, fileName)
	if err != nil {
		return err
	}

	if rsp != nil {

		sha1Val := rsp["X-Bz-Info-Large_file_sha1"].([]string)[0]
		name := rsp["X-Bz-File-Name"].([]string)[0]
		// Read file data
		r := bytes.NewReader(rsp["body"].([]byte))
		b, er := ioutil.ReadAll(r)
		if er != nil {
			return er
		}

		// WRITE to local
		lclPath := filepath.Join(localFilePath, name)
		i, er := iout.WriteOut(b, lclPath)
		if er != nil {
			return er
		}
		log.Logf(log.DEBUG, "wrote out %d", i)

		// Validate file by sha1
		f := filenative.NewFile(lclPath)
		// create sha1 hash from saved file
		sha1hash := encode.Encode(f.Hash(sha1.NewHash(sha1.Format(hashenum.SHA1)), true), encodeenum.Hex)
		// Compare hash to original
		if sha1hash != sha1Val {
			fmt.Errorf("downloaded file doesn't match remote by sha1")
		}
	}

	return nil
}

func DownloadFileByID() error {
	localFilePath := "."
	c := setupCloud()
	path, err := c.MultipartDownloadById(fileID, localFilePath)
	if err != nil {
		return err
	}
	fmt.Println("Downloaded to ", path)

	return nil
}

func DeleteFile() error {
	c := setupCloud()

	r, err := c.DeleteFile(fileName, fileID)
	if err != nil {
		return err
	}

	fmt.Printf("deleted %s %s", r.FileName, r.FileID)

	return nil
}

func DeleteAllFileVersions() error {
	c := setupCloud()

	files, err := c.ListFileVersions(bucketID, fileName)
	if err != nil {
		return err
	}

	for _, f := range files.Files {
		log.Printf("Deleting: %s %s\n", f.FileName, f.FileID)
		r, err := c.DeleteFile(f.FileName, f.FileID)
		if err != nil {
			break
		}
		fmt.Printf("deleted %s %s", r.FileName, r.FileID)
	}

	return nil
}

func HideFile() error {
	c := setupCloud()

	r, err := c.HideFile(bucketID, fileName)
	if err != nil {
		return err
	}
	fmt.Printf("hidden %s %s", r.FileName, r.FileID)

	return nil
}

func GetFileInfo() error {
	c := setupCloud()

	r, err := c.GetFileInfo(fileID)
	if err != nil {
		return err
	}

	fmt.Printf("FileInfo: %s - %s SIZE: %d\n", r.FileName, r.FileID, r.ContentLength)

	return nil
}

func ListUnfinishedParts() error {
	c := setupCloud()

	r, err := c.ListUnfinishedLargeFiles(bucketID)
	if err != nil {
		return err
	}

	for i,k := range r.Files {
		fmt.Printf("%d Part: %s %d %d - %s", i, k.FileName, k.ContentLength, k.UploadTimestamp, k.FileID)
	}

	return nil
}

func CancelUnfinishedLargeUpload() error {
	c := setupCloud()

	r, err := c.CancelLargeFile(fileID)
	if err != nil {
		return err
	}

	fmt.Printf("Cancelled: %s - %s", r.FileName, r.FileId)

	return nil
}

func ListBucketSize() error {
	c := setupCloud()

	var bucketSize int64
	r, err := c.ListFileVersions(bucketID, "")
	if err != nil {
		return err
	}
	if r != nil {
		for _, d := range r.Files {
			bucketSize += d.ContentLength
		}
	}

	return nil
}