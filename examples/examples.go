package main

import (
	"fmt"

	"github.com/colt3k/backblaze2"
	"github.com/colt3k/backblaze2/b2api"
)

const (
	ACCT_ID = ""
	APP_ID  = ""
)

var (
	bucketName = "testbucket"
	filePath = ""
)

func main() {
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

	r, err := c.DeleteBucket("1234567890")
	if err != nil {
		return err
	}

	fmt.Println("Bucket Deleted: ", r.BucketName)

	return nil
}

func CreateVirtualFile() error {
	c := setupCloud()

	r, err := c.UploadVirtualFile("1234567890", "myfile.txt", []byte("test content"), 0)
	if err != nil {
		return err
	}

	fmt.Println("File created: ", r.FileName)

	return nil
}

func ListFileVersions() error {
	c := setupCloud()

	r, err := c.ListFileVersions("1234567890", "myfile.txt")
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

	r, err := c.ListFiles("1234567890", "")
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

	up := backblaze2.NewUploader(bucketName, filePath)
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
