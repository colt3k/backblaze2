package main

import (
	"fmt"

	"github.com/colt3k/backblaze2"
)

func main() {
	er := ListFiles()
	if er != nil {
		fmt.Printf("%v", er)
	}
}

func setupCloud() *backblaze2.Cloud {
	c := backblaze2.CloudStore("<your account id>", "<your app id>")

	fmt.Printf("\nAuthToken: %s\n", c.AuthResponse.AuthorizationToken)

	return c
}
func ListMyKeys() error {
	c := setupCloud()

	r, err := c.ListKeys()
	if err != nil {
		return err
	}
	fmt.Printf("\nKeys: %v\n",r.Keys)

	return nil
}

func CreateMyKey() error {
	c := setupCloud()

	r, err := c.CreateKey("keyName", "keyBucket", []string{"capabilities"})
	if err != nil {
		return err
	}
	fmt.Printf("\nKeys: %v\n",r.KeyName)

	return nil
}

func DeleteMyKey() error {
	c := setupCloud()

	r, err := c.DeleteKey("keyid")
	if err != nil {
		return err
	}
	fmt.Printf("\nKeys: %v\n",r.Key)

	return nil
}

func ListBuckets() error {
	c := setupCloud()

	r, err := c.ListBuckets("", "", nil)
	if err != nil {
		return err
	}
	for _,k := range r.Buckets {
		fmt.Printf("\nBucket: %v - %s\n",k.BucketID, k.BucketName)
	}


	return nil
}

func ListFiles() error {
	c := setupCloud()

	r, err := c.ListFiles("f5e8ff6218490eb66093001f", "")
	if err != nil {
		return err
	}

	for _,k := range r.File {
		fmt.Printf("\nFile: %v - %v\n",k.FileName, k.FileID)
	}

	return nil
}