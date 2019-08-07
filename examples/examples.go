package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"

	log "github.com/colt3k/nglog/ng"
	"github.com/colt3k/utils/encode"
	"github.com/colt3k/utils/encode/encodeenum"
	"github.com/colt3k/utils/file/filenative"
	"github.com/colt3k/utils/hash/hashenum"
	"github.com/colt3k/utils/hash/sha1"
	iout "github.com/colt3k/utils/io"
	"github.com/colt3k/utils/mathut"

	"github.com/colt3k/backblaze2"
	"github.com/colt3k/backblaze2/b2api"
)

var (
	ACCT_ID = os.Getenv("ACCT_ID")
	APP_ID  = os.Getenv("APP_ID")
	bucketName = "ctcloudsync"
	bucketID = "f5e8ff6218490eb66093001f"
	filePath   = "/Users/gcollins/Downloads/MINE/gparted-live-0.33.0-1-i686.iso"
	fileID = "4_zf5e8ff6218490eb66093001f_f2051b8c8ba09564c_d20190807_m194535_c001_v0001116_t0025"
)

func main() {
	if len(ACCT_ID) <= 0 || len(APP_ID) <= 0 {
		fmt.Println("ACCT_ID and/or APP_ID not defined")
		os.Exit(-1)
	}
	er := DownloadFileByID2()
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

	r, err := c.UploadVirtualFile(bucketID, "myfile.txt", []byte("test content"), 0)
	if err != nil {
		return err
	}

	fmt.Println("File created: ", r.FileName)

	return nil
}

func ListFileVersions() error {
	c := setupCloud()

	r, err := c.ListFileVersions(bucketID, "myfile.txt")
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

func DownloadFile() error {
	c := setupCloud()

	localFilePath := "."
	rsp, err := c.DownloadByName(bucketName, "myfile.txt")
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

func DownloadFileByID2() error {
	var size int64 = 0 //301040503
	c := setupCloud()
	localFilePath := "/Users/gcollins/Desktop"
	r, err := c.GetFileInfo(fileID)
	if err != nil {
		return err
	}
	fmt.Println(r.ContentLength)
	size = r.ContentLength
	// If over 200MB multipart download

	totalPartsCount := uint64(math.Ceil(float64(size) / float64(backblaze2.DownloadFileChunk)))

	if size > 200000000 {
		sha1Val := ""
		lclPath := ""
		fmt.Println("file over 200MB, total parts to download", totalPartsCount)
		parts := make([]*backblaze2.UploaderPart, totalPartsCount)
		var lastsize, start, end int64

		for i := uint64(0); i < totalPartsCount; i++ {
			// file size - this chunk is the size of the part
			partSize := int64(math.Min(backblaze2.DownloadFileChunk, float64(size-int64(i*backblaze2.DownloadFileChunk))))

			//first time through set to partSize
			if lastsize == 0 {
				lastsize = partSize	//10485760
			}

			// start is equal to the loop id multiplied by lastsize
			start = int64(i) * lastsize
			if start == 0 {
				end = lastsize -1
			} else {
				end = start + lastsize -1
			}
			if i == totalPartsCount-1 {
				end = start + partSize
			}
			parts[i] = backblaze2.NewUploaderPart(i+1, start, end, partSize)
			fmt.Printf("part %d: full size: %d start %d end %d\n", i, size, start, end)
		}

		var file *os.File
		var err2 error
		var written int
		for i,j := range parts {

			rsp, err := c.DownloadByID(r.FileID, "bytes="+mathut.FmtInt(int(j.Start))+"-"+mathut.FmtInt(int(j.End)))
			if err != nil {
				return err
			}
			if rsp != nil {
				sha1Val = rsp["X-Bz-Info-Large_file_sha1"].([]string)[0]
				name := rsp["X-Bz-File-Name"].([]string)[0]
				fmt.Printf("Part %d of %s\n", i, name)

				r := bytes.NewReader(rsp["body"].([]byte))
				b, er := ioutil.ReadAll(r)
				if er != nil {
					return er
				}

				lclPath = filepath.Join(localFilePath, name)
				if i == 0 {
					file, err2 = os.OpenFile(lclPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
					if err2 != nil {
						fmt.Println(err2)
						os.Exit(1)
					}
				}
				n, err := file.Write(b)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				written += n
			}
		}

		file.Close()
		fmt.Printf("wrote out %d original size: %d\n", written, r.ContentLength)
		// Validate file by sha1
		f := filenative.NewFile(lclPath)
		// Create sha1 hash and encode it
		sha1hash := encode.Encode(f.Hash(sha1.NewHash(sha1.Format(hashenum.SHA1)), true), encodeenum.Hex)
		// Compare hash to original
		if sha1hash != sha1Val {
			return fmt.Errorf("downloaded file doesn't match remote by sha1 %s local %s", sha1Val, sha1hash)
		}
	}

	return nil
}

func DownloadFileById() error {

	c := setupCloud()
	localFilePath := "."
	rsp, err := c.DownloadByID("1234567890", "")
	if err != nil {
		return err
	}

	if rsp != nil {
		name := rsp["X-Bz-File-Name"].([]string)[0]
		sha1Val := rsp["X-Bz-Info-Large_file_sha1"].([]string)[0]

		// Read file data
		r := bytes.NewReader(rsp["body"].([]byte))
		b, er := ioutil.ReadAll(r)
		if er != nil {
			return er
		}

		lclPath := filepath.Join(localFilePath, name)
		i, er := iout.WriteOut(b, lclPath)
		if er != nil {
			return er
		}
		log.Logf(log.DEBUG, "wrote out %d", i)

		// Validate file by sha1
		f := filenative.NewFile(lclPath)
		// Create sha1 hash and encode it
		sha1hash := encode.Encode(f.Hash(sha1.NewHash(sha1.Format(hashenum.SHA1)), true), encodeenum.Hex)
		// Compare hash to original
		if sha1hash != sha1Val {
			return fmt.Errorf("downloaded file doesn't match remote by sha1 %s local %s", sha1Val, sha1hash)
		}
	}

	return nil
}

func DeleteFile() error {
	c := setupCloud()

	r, err := c.DeleteFile("filename", "fileID")
	if err != nil {
		return err
	}

	fmt.Printf("deleted %s %s", r.FileName, r.FileID)

	return nil
}

func DeleteAllFileVersions() error {
	c := setupCloud()

	files, err := c.ListFileVersions(bucketID, "myfile.txt")
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

	r, err := c.HideFile(bucketID, "myfile.txt")
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

	r, err := c.CancelLargeFile("fileID")
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