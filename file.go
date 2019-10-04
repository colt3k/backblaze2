package backblaze2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/colt3k/nglog/ers/bserr"
	log "github.com/colt3k/nglog/ng"
	"github.com/colt3k/utils/concur"
	"github.com/colt3k/utils/encode"
	"github.com/colt3k/utils/encode/encodeenum"
	"github.com/colt3k/utils/file"
	"github.com/colt3k/utils/file/filemeta"
	"github.com/colt3k/utils/file/filenative"
	"github.com/colt3k/utils/file/filesize"
	"github.com/colt3k/utils/hash/hashenum"
	"github.com/colt3k/utils/hash/sha1"
	iout "github.com/colt3k/utils/io"
	"github.com/colt3k/utils/io/ioreader/passthrough"
	"github.com/colt3k/utils/io/iowriter"
	"github.com/colt3k/utils/mathut"
	"github.com/colt3k/utils/stringut"

	"github.com/colt3k/backblaze2/b2api"
	"github.com/colt3k/backblaze2/errs"
	"github.com/colt3k/backblaze2/internal/auth"
	"github.com/colt3k/backblaze2/internal/caller"
	"github.com/colt3k/backblaze2/internal/env"
	"github.com/colt3k/backblaze2/internal/uri"
	"github.com/colt3k/backblaze2/perms"
)

const fileChunk = 10 * (1 << 20)
const fileChunkLg = 100 * (1 << 20)
const DownloadFileChunk = 10 * (1 << 20)
const (
	// MinPartSize in MB
	MinPartSize = 6
	// MinParts count size
	MinParts = 2
	// MaxUploadTB size in TB
	MaxUploadTB     = 10
	maxParts        = 100
	DownSplitThresh = 200000000
)

var (
	//MaxPerSessionUploadPerPart sessions
	MaxPerSessionUploadPerPart = 3
	fcLg                       bool
)

func post(url string, req interface{}, header map[string]string) (map[string]interface{}, errs.Error) {
	msg, errUn := caller.UnMarshalRequest(req)
	if errUn != nil {
		return nil, errs.New(errUn, "")
	}
	log.Logln(log.DEBUG, "Request:", string(msg))

	mapData, er := caller.MakeCall("POST", url, bytes.NewReader(msg), header)
	if er != nil {
		return nil, er
	}
	//log.Logln(log.DEBUG, "Actual return: ", string(mapData["body"].([]byte)))
	return mapData, nil
}

// SendParts send parts to target
func (c *Cloud) SendParts(up *Upload) (bool, error) {

	// This controls the Upload of PARTS using UploadPart
	if perms.StartLargeFile(c.AuthResponse) {

		// Get parts we created earlier
		parts := up.RetrieveToUpload()

		//Create Worker Pool to upload ***************************************

		var tasks []*concur.Task
		fo, err := os.Open(up.File.Path())
		bserr.StopErr(err, "err opening file")

		defer fo.Close()
		for _, d := range parts {
			d := d
			if len(d.Etag) <= 0 {
				//Create Task, send to worker
				task := concur.NewTask(
					func() (error, []byte) {
						et := c.worker(up, d, fo)
						return nil, []byte(et)
					},
					NewRtnd(up))

				tasks = append(tasks, task)
			}
		}
		p := concur.NewPool(tasks, MaxPerSessionUploadPerPart)
		p.Run()

		// END WORKER POOL ****************************************************

		if up.Completed() {
			//Completed Finish off
			shas := make([]string, 0)
			for _, d := range parts {
				shas = append(shas, d.Etag)
			}
			_, err := c.FinishLargeFileUpload(up.FileID, shas)
			if err != nil {
				return false, err
			}
			return true, nil
		} else {
			// TRY TO RUN THROUGH INCOMPLETE ONES AGAIN after sleeping a bit
			AuthCounter += 1
			if AuthCounter <= MaxAuthTry {
				log.Logln(log.WARN,"[multipart] service unavailable trying again, please stand by")
				sleep := 7 * time.Second
				jitter := time.Duration(rand.Int63n(int64(sleep)))
				sleep = sleep + jitter/2
				time.Sleep(sleep)

				c.AuthConfig.Clear = true
				c.AuthAccount()
				return c.SendParts(up)
			}
		}
		return false, fmt.Errorf("issue upload all parts")
	}
	return false, fmt.Errorf("authorization issue uploading multipart file")
}

func (c *Cloud) worker(up *Upload, p *UploaderPart, fo *os.File) string {

	fupr, err := c.GetUploadPartURL(up.FileID)
	if err != nil {
		return err.Error()
	}
	meta := filemeta.New(up.File)

	//Not sent yet, send now
	if len(p.Etag) <= 0 {
		//Convert to load via START/STOP instead of actual file part
		resp, err := UploadPart(fupr, up, p, fo, meta)
		if err != nil {
			log.Logf(log.ERROR, "%+v", err)
			return ""
		}
		log.Logln(log.DEBUG, p.PartID, " Uploaded: ", resp)
		// return string of etag+^+partid
		etID := resp.ContentSha1 + "^" + strconv.FormatUint(p.PartID, 10)
		return etID
	}
	// return ^+partid
	return "^" + strconv.FormatUint(p.PartID, 10)
}

func UploadPart(fupr *b2api.GetFileUploadPartResponse, up *Upload, p *UploaderPart, fo *os.File, meta file.Meta) (*b2api.UploadPartResponse, error) {

	partID := int(p.PartID)
	size := int64(p.Size)
	header := auth.BuildAuthMap(fupr.AuthorizationToken)

	header["X-Bz-File-Name"] = up.File.Name()
	header["X-Bz-Part-Number"] = strconv.Itoa(partID)
	header["Content-Length"] = mathut.FmtInt(int(size))

	//Create a buffer of the partial size to load with data
	partBuffer := make([]byte, p.Size)
	//Read data into byte array
	fo.ReadAt(partBuffer, p.Start)

	sha1hash := encode.Encode(sha1.NewHash(sha1.Format(hashenum.SHA1)).String(string(partBuffer)), encodeenum.Hex)
	header["X-Bz-Content-Sha1"] = sha1hash
	header["X-Bz-Info-src_last_modified_millis"] = mathut.FmtInt(int(meta.LastMod()))

	r := ioutil.NopCloser(bytes.NewReader(partBuffer))
	log.Logln(log.DEBUG, "Uploading: ", partID, " Size: ", size)
	rc := passthrough.NewStream(r, size, up.File.Name(), int(partID), int(up.TotalPartsCount), 5)

	mapData, err := caller.MakeCall("POST", fupr.UploadURL, rc, header)
	if err != nil {
		if testRetryErr(err) {

		}
		return nil, err
	}

	//log.Logln(log.DEBUG, "Actual return: ", string(mapData["body"].([]byte)))

	var b2Response *b2api.UploadPartResponse
	errUn := json.Unmarshal(mapData["body"].([]byte), &b2Response)
	if errUn != nil {
		return nil, errs.New(errUn, "")
	}

	return b2Response, nil
}

// NewUploaderPart create uploader part
func NewUploaderPart(part uint64, start, end, size int64) *UploaderPart {
	t := new(UploaderPart)
	t.PartID = part
	t.Size = size
	t.Start = start
	t.End = end

	return t
}

// UploadURL retrieve upload url
func (c *Cloud) UploadURL(bucketId string) (*b2api.UploadURLResp, errs.Error) {

	if perms.GetUploadURL(c.AuthResponse) {

		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)
		req := b2api.UploadURLReq{
			BucketId: bucketId,
		}

		data, er := post(c.AuthResponse.APIURL+uri.B2GetUploadURL, req, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.UploadURL(bucketId)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.UploadURLResp
		errUn := json.Unmarshal(data["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// UploadFile file name and info must fit in 7k byte limit,
// 	the file to be uploaded is the message body and is not encoded in any way.
//		it's not URL encoded, it's not MIME encoded
func (c *Cloud) UploadFile(bucketId string, up *Upload) (*b2api.UploadResp, errs.Error) {

	if perms.UploadFile(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)
		upURL, er := c.UploadURL(bucketId)
		if er != nil {
			return nil, er
		}

		// Process File
		//	- Retrieve File Name
		//	- Retrieve file size
		//	- Retrieve SHA-1
		//	- Retrieve Last Modified in millis
		f := filenative.NewFile(up.Filepath)
		fullPathName := f.Name()
		if len(up.OverridePath) > 0 {
			fullPathName = up.OverridePath
		}
		name := url.URL{Path: fullPathName}

		meta := filemeta.New(f)
		// create sha1 hash and encode it
		sha1hash := encode.Encode(f.Hash(sha1.NewHash(sha1.Format(hashenum.SHA1)), true), encodeenum.Hex)

		header = auth.BuildAuthMap(upURL.AuthorizationToken)
		header["X-Bz-File-Name"] = name.String() // name of the file, in percent-encoded UTF-8
		header["Content-Type"] = "b2/x-auto"     // MIME type, b2/x-auto to automatically set the stored Content-Type post upload
		// https://www.backblaze.com/b2/docs/content-types.html
		header["Content-Length"] = mathut.FmtInt(int(f.Size())) // number of bytes in the file being uploaded +40 for SHA1
		// When sending the SHA1 checksum at the end, the Content-Length should be set to the size of the file plus the 40 bytes of hex checksum.
		header["X-Bz-Content-Sha1"] = sha1hash                                            // SHA1 checksum of the content of the file. B2 will check this when the file is uploaded, to make sure that the file arrived correctly
		header["X-Bz-Info-src_last_modified_millis"] = mathut.FmtInt(int(meta.LastMod())) // SHA1 checksum of the content of the file. B2 will check this when the file is uploaded, to make sure that the file arrived correctly
		//header["X-Bz-Info-b2-content-disposition"] = ""   // value must match the grammar specified in RFC 6266
		//header["X-Bz-Info-*"] = ""	// up to 10 of these replacing * with the value must be a percent encoded UTF8 string

		// File is passed as []byte or ReadCloser
		fle, errOp := os.Open(f.Path())
		if errOp != nil {
			return nil, errs.New(errOp, "")
		}
		data, errOp := ioutil.ReadAll(fle)
		if errOp != nil {
			return nil, errs.New(errOp, "")
		}
		fle.Close()

		rdr := bytes.NewReader(data)
		r := ioutil.NopCloser(rdr)
		log.Logln(log.DEBUG, "Uploading  Size: ", f.Size())
		rc := passthrough.NewStream(r, f.Size(), f.Name(), 1, 1, 1)

		mapData, er := caller.MakeCall("POST", upURL.UploadURL, rc, header)
		if er != nil {
			if rc != nil {
				rc.Close()
			}
			if r != nil {
				r.Close()
			}
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.UploadFile(bucketId, up)
				}
			}
		}
		AuthCounter = 0
		//log.Logln(log.DEBUG, "Actual return: ", string(mapData["body"].([]byte)))
		var b2Response b2api.UploadResp
		errUn := json.Unmarshal(mapData["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "unmarshall body")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

func (c *Cloud) UploadVirtualFile(bucketId, fname string, data []byte, lastMod int64) (*b2api.UploadResp, errs.Error) {

	if perms.UploadFile(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)
		upURL, er := c.UploadURL(bucketId)
		if er != nil {
			return nil, er
		}

		name := url.URL{Path: fname}
		sz := len(data)
		// create sha1 hash and encode it
		sha1hash := encode.Encode(sha1.NewHash(sha1.Format(hashenum.SHA1)).String(string(data)), encodeenum.Hex)

		header = auth.BuildAuthMap(upURL.AuthorizationToken)
		header["X-Bz-File-Name"] = name.String() // name of the file, in percent-encoded UTF-8
		header["Content-Type"] = "b2/x-auto"     // MIME type, b2/x-auto to automatically set the stored Content-Type post upload
		// https://www.backblaze.com/b2/docs/content-types.html
		header["Content-Length"] = mathut.FmtInt(sz) // number of bytes in the file being uploaded +40 for SHA1
		// When sending the SHA1 checksum at the end, the Content-Length should be set to the size of the file plus the 40 bytes of hex checksum.
		header["X-Bz-Content-Sha1"] = sha1hash                                     // SHA1 checksum of the content of the file. B2 will check this when the file is uploaded, to make sure that the file arrived correctly
		header["X-Bz-Info-src_last_modified_millis"] = mathut.FmtInt(int(lastMod)) // SHA1 checksum of the content of the file. B2 will check this when the file is uploaded, to make sure that the file arrived correctly
		//header["X-Bz-Info-b2-content-disposition"] = ""   // value must match the grammar specified in RFC 6266
		//header["X-Bz-Info-*"] = ""	// up to 10 of these replacing * with the value must be a percent encoded UTF8 string

		rdr := bytes.NewReader(data)
		r := ioutil.NopCloser(rdr)
		log.Logln(log.DEBUG, "Uploading  Size: ", sz)
		rc := passthrough.NewStream(r, int64(sz), fname, 1, 1, 1)
		mapData, er := caller.MakeCall("POST", upURL.UploadURL, rc, header)
		if er != nil {
			if rc != nil {
				rc.Close()
			}
			if r != nil {
				r.Close()
			}
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.UploadVirtualFile(bucketId, fname, data, lastMod)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		//log.Logln(log.DEBUG, "Actual return: ", string(mapData["body"].([]byte)))
		var b2Response b2api.UploadResp
		errUn := json.Unmarshal(mapData["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "unmarshall body")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// ListFiles list files by name or all in bucket
func (c *Cloud) ListFiles(bucketId, filename string, qty int) (*b2api.ListFilesResponse, errs.Error) {

	var maxFileCount b2api.MaxFileCount
	maxFileCount = 100
	if qty > 100 {
		maxFileCount = b2api.MaxFileCount(qty)
	}
	if perms.ListFileNames(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		req := &b2api.ListFileReq{
			BucketID:      bucketId,
			StartFileName: "",
			MaxFileCount:  maxFileCount,
			Prefix:        filename,
			Delimiter:     "",
		}

		data, er := post(c.AuthResponse.APIURL+uri.B2ListFileNames, req, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.ListFiles(bucketId, filename, qty)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.ListFilesResponse
		errUn := json.Unmarshal(data["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// ListFileVersions lists out all versions of file
func (c *Cloud) ListFileVersions(bucketId, fileName string) (*b2api.ListFileVersionsResponse, errs.Error) {

	if perms.ListFileVersions(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		req := &b2api.ListFileVersionsReq{
			BucketID:      bucketId,
			StartFileName: "",
			StartFileID:   "",
			MaxFileCount:  100,
			Prefix:        fileName,
			Delimiter:     "",
		}

		data, er := post(c.AuthResponse.APIURL+uri.B2ListFileVersions, req, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.ListFileVersions(bucketId, fileName)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.ListFileVersionsResponse
		errUn := json.Unmarshal(data["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// DeleteFile deletes file version by file id
func (c *Cloud) DeleteFile(fileName, fileID string) (*b2api.DeleteFileVersionResponse, errs.Error) {

	if perms.DeleteFiles(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		req := &b2api.DeleteFileVersionReq{
			FileName: fileName,
			FileID:   fileID,
		}

		data, er := post(c.AuthResponse.APIURL+uri.B2DeleteFileVersion, req, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.DeleteFile(fileName, fileID)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.DeleteFileVersionResponse
		errUn := json.Unmarshal(data["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// HideFile hides file by putting 0 content length in history
func (c *Cloud) HideFile(bucketId, fileName string) (*b2api.HideFileResponse, errs.Error) {

	if perms.HideFile(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		req := &b2api.HideFileReq{
			BucketId: bucketId,
			FileName: fileName,
		}

		data, er := post(c.AuthResponse.APIURL+uri.B2HideFile, req, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.HideFile(bucketId, fileName)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.HideFileResponse
		errUn := json.Unmarshal(data["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// GetFileInfo provides file information
func (c *Cloud) GetFileInfo(fileID string) (*b2api.GetFileInfoResponse, errs.Error) {

	if perms.GetFileInfo(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		req := &b2api.GetFileInfoReq{
			FileID: fileID,
		}

		data, er := post(c.AuthResponse.APIURL+uri.B2GetFileInfo, req, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.GetFileInfo(fileID)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.GetFileInfoResponse
		errUn := json.Unmarshal(data["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		sha1 := strings.TrimSpace(b2Response.ContentSha1)
		if sha1 == "none" || len(sha1) == 0 {
			//log.Logln(log.INFO, "data: ", b2Response.FileInfo.LargeFileSha1)
			sha1 = b2Response.FileInfo.LargeFileSha1
			b2Response.ContentSha1 = sha1
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// GetDownloadAuth required prior to download
func (c *Cloud) GetDownloadAuth(bucketID, filenamePrefix string, validDurationInSeconds int64) (*b2api.DownloadAuthResponse, errs.Error) {

	if perms.GetDownloadAuth(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		req := &b2api.DownloadAuthReq{
			BucketId:               bucketID,
			FilenamePrefix:         filenamePrefix,
			ValidDurationInSeconds: validDurationInSeconds,
			B2ContentDisposition:   "",
		}

		data, er := post(c.AuthResponse.APIURL+uri.B2GetDownloadAuth, req, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.GetDownloadAuth(bucketID, filenamePrefix, validDurationInSeconds)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.DownloadAuthResponse
		errUn := json.Unmarshal(data["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

/* GetDownloadAuth required prior to download
In map:
	body			- content of file
	Content-Length
	Content-Type
	X-Bz-File-Id
	X-Bz-File-Name
	X-Bz-Content-Sha1
	X-Bz-Info-author
	X-Bz-Upload-Timestamp
	Cache-Control : max-age (only); inherited from Bucket Info
PLUS:
	X-Bz-Info-* headers for any custom file info during upload
*/
func (c *Cloud) DownloadByName(bucketName, fileName string) (map[string]interface{}, errs.Error) {

	if perms.DownloadFile(c.AuthResponse) {

		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		url := c.AuthResponse.DownloadURL + "/file/" + bucketName + "/" + fileName + "?Authorization=" + c.AuthResponse.AuthorizationToken + "&b2-content-disposition=large_file_sha1"
		mapData, er := caller.MakeCall("GET", url, nil, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.DownloadByName(bucketName, fileName)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		return mapData, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

/* DownloadByID downloads file by id
In map:
	body			- content of file
	Content-Length
	Content-Type
	X-Bz-File-Id
	X-Bz-File-Name
	X-Bz-Content-Sha1
	X-Bz-Info-author
	X-Bz-Upload-Timestamp
	Cache-Control : max-age (only); inherited from Bucket Info
PLUS:
	X-Bz-Info-* headers for any custom file info during upload
*/
func (c *Cloud) DownloadByID(fileID, byteRange string) (map[string]interface{}, errs.Error) {

	if perms.DownloadFile(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		url := c.AuthResponse.DownloadURL + uri.B2DownloadFileById + "?fileId=" + fileID
		if len(byteRange) > 0 {
			header["Range"] = byteRange
		}
		mapData, er := caller.MakeCall("GET", url, nil, header)
		if er != nil {
			return nil, er
		}

		return mapData, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// StartLargeFile create initial start call
func (c *Cloud) StartLargeFile(bucketID, fileInfo string, up *Upload) (*b2api.StartLargeFileResponse, errs.Error) {

	if perms.StartLargeFile(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		// build sha1 and include as file info 'large_file_sha1'
		// create sha1 hash and encode it
		sha1hash := encode.Encode(up.File.Hash(sha1.NewHash(sha1.Format(hashenum.SHA1)), true), encodeenum.Hex)

		fm := filemeta.New(up.File)

		fullFilePath := up.File.Name()
		if len(up.OverridePath) > 0 {
			fullFilePath = up.OverridePath
		}
		req := &b2api.StartLargeFileReq{
			BucketId:    bucketID,
			FileName:    fullFilePath,
			ContentType: "b2/x-auto",
			FileInfo:    b2api.FileInfo{SrcLastModifiedMillis: mathut.FmtInt(int(fm.LastMod())), LargeFileSha1: sha1hash},
		}

		data, er := post(c.AuthResponse.APIURL+uri.B2StartLargeFile, req, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.StartLargeFile(bucketID, fileInfo, up)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.StartLargeFileResponse
		errUn := json.Unmarshal(data["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

func (c *Cloud) GetUploadPartURL(fileID string) (*b2api.GetFileUploadPartResponse, errs.Error) {

	if perms.GetUploadPartURL(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		req := &b2api.GetFileUploadPartReq{
			FileID: fileID,
		}

		data, er := post(c.AuthResponse.APIURL+uri.B2GetUploadPartURL, req, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.GetUploadPartURL(fileID)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.GetFileUploadPartResponse
		errUn := json.Unmarshal(data["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

func (c *Cloud) ListPartsURL(fileID string, startPartNo, maxPartCount int64) (*b2api.ListPartsResponse, errs.Error) {

	if perms.ListParts(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		req := &b2api.ListPartsReq{
			FileID:          fileID,
			StartPartNumber: startPartNo,
			MaxPartCount:    maxPartCount,
		}

		data, er := post(c.AuthResponse.APIURL+uri.B2ListParts, req, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.ListPartsURL(fileID, startPartNo, maxPartCount)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.ListPartsResponse
		errUn := json.Unmarshal(data["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// ListUnfinishLargeFiles show unfinished large file uploads
func (c *Cloud) ListUnfinishedLargeFiles(bucketID string) (*b2api.ListUnfinishedLargeFilesResponse, errs.Error) {

	if perms.ListUnfinishedLargeFiles(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		req := &b2api.ListUnfinishedLargeFilesReq{
			BucketId:     bucketID,
			NamePrefix:   "",
			StartFileId:  "",
			MaxFileCount: 100,
		}

		data, er := post(c.AuthResponse.APIURL+uri.B2ListUnfinishedLargeFiles, req, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.ListUnfinishedLargeFiles(bucketID)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.ListUnfinishedLargeFilesResponse
		errUn := json.Unmarshal(data["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// FinishLargeFileUpload call complete on large upload
func (c *Cloud) FinishLargeFileUpload(fileId string, sha1Array []string) (*b2api.FinishUpLgFileResp, errs.Error) {

	if perms.FinishLargeFile(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)
		req := &b2api.FinishUpLgFileReq{
			FileId: fileId,
			Sha1Ar: sha1Array,
		}
		data, er := post(c.AuthResponse.APIURL+uri.B2FinishLargeFile, req, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.FinishLargeFileUpload(fileId, sha1Array)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		//Convert Message
		var b2Response b2api.FinishUpLgFileResp
		errUn := json.Unmarshal(data["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// CancelLargeFile cancel large file upload process
func (c *Cloud) CancelLargeFile(fileId string) (*b2api.CanxUpLgFileResp, errs.Error) {

	if perms.CancelLargeFile(c.AuthResponse) {
		header := auth.BuildAuthMap(c.AuthResponse.AuthorizationToken)

		req := &b2api.CanxUpLgFileReq{
			FileId: fileId,
		}
		data, er := post(c.AuthResponse.APIURL+uri.B2CancelLargeFile, req, header)
		if er != nil {
			if testRetryErr(er) {
				// delete it and call again
				AuthCounter += 1
				if AuthCounter <= MaxAuthTry {
					if AuthCounter > 1 {
						sleep := 3 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					if er.Code() == "service_unavailable" {
						log.Logln(log.WARN,"service unavailable trying again, please stand by")
						sleep := 7 * time.Second
						jitter := time.Duration(rand.Int63n(int64(sleep)))
						sleep = sleep + jitter/2
						time.Sleep(sleep)
					}
					c.AuthConfig.Clear = true
					c.AuthAccount()
					return c.CancelLargeFile(fileId)
				}
			}
			return nil, er
		}
		AuthCounter = 0
		var b2Response b2api.CanxUpLgFileResp
		errUn := json.Unmarshal(data["body"].([]byte), &b2Response)
		if errUn != nil {
			return nil, errs.New(errUn, "")
		}
		return &b2Response, nil
	}
	return nil, errs.New(fmt.Errorf("not allowed"), "")
}

// *******************************************************
// UploaderPart struct
type UploaderPart struct {
	Size   int64  `json:"size"`
	PartID uint64 `json:"part"`
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
	Etag   string `json:"etag"`
}

// Upload struct
type Upload struct {
	AppName         string          `json:"-"`
	Bucket          string          `json:"bucket"`
	Filepath        string          `json:"filepath"`
	OverridePath    string          `json:"override_path"`
	UploadID        string          `json:"uploadid"`
	File            file.File       `json:"-"`
	FileID          string          `json:"fileId"`
	Parts           []*UploaderPart `json:"parts"`
	TotalPartsCount uint64          `json:"total_parts_count"`
	mu              sync.Mutex
}

// New create upload object
func NewUploader(bucketName, filepath, overridepath string) *Upload {
	t := new(Upload)
	t.Bucket = bucketName
	t.Filepath = filepath
	t.OverridePath = overridepath
	t.File = filenative.NewFile(t.Filepath)
	t.AppName = AppFolderName
	return t
}

// Available file exists
func (u *Upload) Available() bool {
	if len(strings.TrimSpace(u.Filepath)) <= 0 {
		return false
	}
	return u.File.Available()
}

// ValidateOverMinPartSize check size
func (u *Upload) ValidateOverMinPartSize() bool {
	log.Logln(log.DEBUG, "FileSize:", u.File.Size())
	log.Logln(log.DEBUG, "MinPartSize ", int64(filesize.SizeTypes(filesize.Bytes).Convert(MinPartSize, filesize.Mega, true)))
	u.TotalPartsCount = uint64(math.Ceil(float64(u.File.Size()) / float64(fileChunk)))
	// Ensure over 6MB
	if u.File.Size() > int64(filesize.SizeTypes(filesize.Bytes).Convert(MinPartSize, filesize.Mega, true)) && u.TotalPartsCount >= MinParts {
		return true
	}
	return false
}

// ValidateSize check size
func (u *Upload) ValidateSize() bool {
	if u.File.Size() > int64(filesize.SizeTypes(filesize.Bytes).Convert(MaxUploadTB, filesize.Tera, true)) {
		return false
	}
	return true
}

// ComputePartTotal compute total parts
func (u *Upload) ComputePartTotal() error {
	u.TotalPartsCount = uint64(math.Ceil(float64(u.File.Size()) / float64(fileChunk)))
	log.Logln(log.DEBUG, "How many pieces should we create?", u.TotalPartsCount)
	if u.TotalPartsCount > maxParts {
		//Set to 1GB if parts would be over 100
		fcLg = true
		u.TotalPartsCount = uint64(math.Ceil(float64(u.File.Size()) / float64(fileChunkLg)))
	}
	if u.TotalPartsCount > maxParts {
		return fmt.Errorf("part count over max allowed: %d, max: %d", u.TotalPartsCount, maxParts)
	}
	log.Logf(log.INFO, "Splitting into %d pieces.\n", u.TotalPartsCount)
	return nil
}

/*
SetupPartSizes determine part sizes
Pass in part count
Pass in data on each part size and status i.e. completed or not
*/
func (u *Upload) SetupPartSizes(fileID string) {
	u.UploadID = fileID
	u.FileID = fileID
	u.Parts = make([]*UploaderPart, u.TotalPartsCount)

	//Open file to read and determine start/end of each part
	//fo, err := os.Open(u.File.Path())
	//bserr.Err(err, "err opening file")
	//
	//defer fo.Close()

	var lastsize, start, end int64

	for i := uint64(0); i < u.TotalPartsCount; i++ {
		// file size - this chunk is the size of the part
		partSize := int64(math.Min(fileChunk, float64(u.File.Size()-int64(i*fileChunk))))
		if fcLg {
			partSize = int64(math.Min(fileChunkLg, float64(u.File.Size()-int64(i*fileChunkLg))))
		}

		//first time through set to partSize
		if lastsize == 0 {
			lastsize = partSize
		}

		// start is equal to the loop id multiplied by lastsize
		start = int64(i) * lastsize
		if start == 0 {
			end = lastsize
		} else {
			end = start + lastsize
		}
		if i == u.TotalPartsCount-1 {
			end = start + partSize
		}
		u.Parts[i] = NewUploaderPart(i+1, start, end, partSize)
		log.Logf(log.DEBUG, "part %d: full size: %d start %d end %d", i, u.File.Size(), start, end)
	}
}

// WriteOut write out upload info
func (u *Upload) WriteOutFileData2Upload(app string) {

	//Write out preparation of file for upload
	fmtd, err := json.MarshalIndent(u, "", "    ")

	if bserr.NotErr(err) {
		bucketDir := env.BuildBucketDir(u.AppName, u.Bucket)
		iowriter.NewFileWriter()
		fw := iowriter.NewFileWriter()
		fw.WriteOut(fmtd, path.Join(bucketDir, env.UploadFolder))
	}
}

func (u *Upload) Process(c *Cloud) (string, error) {
	// Find bucketId by name
	bkts, err := c.ListBuckets("", u.Bucket, nil)
	if err != nil {
		return "", err
	}
	if len(bkts.Buckets) <= 0 {
		return "", fmt.Errorf("no buckets found")
	}
	bucket := bkts.Buckets[0]

	if u.ValidateOverMinPartSize() && u.ValidateSize() {
		// determine amount of parts required
		err := u.ComputePartTotal()
		if err != nil {
			return "", err
		}

		// Trigger Multipart with a Start
		strtLgFileResp, err := c.StartLargeFile(bucket.BucketID, "", u)
		if err != nil {
			return "", err
		}
		// Setup for uploading of parts
		u.SetupPartSizes(strtLgFileResp.FileID)
		// Write file out to localhost in order to recover from failure
		u.WriteOutFileData2Upload(AppFolderName)

		// Start sending parts
		if ok, err := c.SendParts(u); ok {
			log.Logf(log.INFO, "Upload Completed successfully into %s", u.Bucket)
			uploadFile := UploaderDir(u.Bucket)
			log.Logf(log.INFO, "removing upload file %s", uploadFile)
			if !file.Delete(uploadFile) {
				log.Logf(log.WARN, "upload file not removed %s", uploadFile)
			}
			return u.FileID, nil
		} else {
			return "", fmt.Errorf("upload failed, try again later %v", err)
		}

	} else if !u.ValidateOverMinPartSize() {

		rsp, err := c.UploadFile(bucket.BucketID, u)
		if err != nil {
			return "", err
		}
		if rsp != nil {
			return rsp.FileID, nil
		}

	} else {
		return "", fmt.Errorf("file too large for upload, over %d TB", MaxUploadTB)
	}

	return "", nil
}

// UpdateEtag update each etag
func (u *Upload) UpdateEtag(partID int, appName, tag string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	for i, d := range u.Parts {
		if d.PartID == uint64(partID) {
			u.Parts[i].Etag = tag
		}
	}

	u.WriteOutFileData2Upload(appName)
}

// RetrieveToUpload retrieve etag
func (u *Upload) RetrieveToUpload() []*UploaderPart {
	toUpload := make([]*UploaderPart, 0)
	for _, d := range u.Parts {
		if len(strings.TrimSpace(d.Etag)) <= 0 {
			toUpload = append(toUpload, d)
		}
	}
	return toUpload
}

// Completed has this finished
func (u *Upload) Completed() bool {
	var counter = 0

	for _, d := range u.Parts {
		if len(d.Etag) > 0 {
			counter++
		}
	}
	if counter == len(u.Parts) {
		return true
	}
	return false
}

// Rtnd returned struct
type Rtnd struct {
	etag   string
	upload *Upload
	mu     sync.Mutex
}

// NewRtnd create new Rtnd struct
func NewRtnd(up *Upload) *Rtnd {
	tmp := &Rtnd{upload: up}
	return tmp
}

// Response set response
func (r *Rtnd) Response(b []byte) {
	if b == nil {
		return
	}
	//receives etag+^+partid
	ar := strings.Split(string(b), "^")
	var id int
	var err error
	if len(ar) == 2 {
		r.etag = ar[0]
		id, err = strconv.Atoi(ar[1])
		bserr.WarnErr(err, "id not an integer on response")
	} else {
		bserr.WarnErr(err, "id not passed on response")
		return
	}

	fmt.Println(id, " Returned: ", r.etag)

	r.upload.UpdateEtag(id, r.upload.AppName, r.etag)
}

func (c *Cloud) MultipartDownloadById(fileID, localFilePath, fileNameOverride string) (string, error) {
	r, err := c.GetFileInfo(fileID)
	if err != nil {
		return "", err
	}
	//fmt.Println(r.ContentLength)
	var size int64 = 0 //301040503
	size = r.ContentLength
	name := r.FileName
	sha1Val := r.ContentSha1

	// If over 200MB multipart download
	totalPartsCount := uint64(math.Ceil(float64(size) / float64(DownloadFileChunk)))

	if size > DownSplitThresh {

		lclPath := filepath.Join(localFilePath, name)
		if len(fileNameOverride) > 0 {
			lclPath = filepath.Join(localFilePath, fileNameOverride)
		}
		fmt.Printf("file over %s, total parts to download %d\n", stringut.HRByteCount(DownSplitThresh, true), totalPartsCount)
		parts := make([]*UploaderPart, totalPartsCount)
		var lastsize, start, end int64

		for i := uint64(0); i < totalPartsCount; i++ {
			// file size - this chunk is the size of the part
			partSize := int64(math.Min(DownloadFileChunk, float64(size-int64(i*DownloadFileChunk))))

			//first time through set to partSize
			if lastsize == 0 {
				lastsize = partSize //10485760
			}

			// start is equal to the loop id multiplied by lastsize
			start = int64(i) * lastsize
			if start == 0 {
				end = lastsize - 1
			} else {
				end = start + lastsize - 1
			}
			if i == totalPartsCount-1 {
				end = start + partSize
			}
			parts[i] = NewUploaderPart(i+1, start, end, partSize)
			//fmt.Printf("part %d: full size: %d start %d end %d\n", i, size, start, end)
		}
		fmt.Printf("Downloading in %d parts", totalPartsCount)

		file, err2 := os.OpenFile(lclPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err2 != nil {
			fmt.Println(err2)
			os.Exit(1)
		}
		file.Truncate(r.ContentLength)

		var tasks []*concur.Task
		for _, d := range parts {
			d := d
			//Create Task, send to worker
			task := concur.NewTask(
				func() (error, []byte) {
					workerDown(c, d, file, r.FileID)
					return nil, []byte(nil)
				},
				NewRtnd(nil))

			tasks = append(tasks, task)
		}
		p := concur.NewPool(tasks, 3)
		p.Run()

		file.Close()

		// Validate file by sha1
		f := filenative.NewFile(lclPath)
		// Create sha1 hash and encode it
		sha1hash := encode.Encode(f.Hash(sha1.NewHash(sha1.Format(hashenum.SHA1)), true), encodeenum.Hex)
		// Compare hash to original
		if sha1hash != sha1Val {
			return "", fmt.Errorf("downloaded file doesn't match remote by sha1 %s local %s", sha1Val, sha1hash)
		}
		return lclPath, nil
	} else {
		// download as a single file SEE DownloadFileById
		rsp, err := c.DownloadByID(fileID, "")
		if err != nil {
			return "", err
		}

		if rsp != nil {
			// Read file data
			r := bytes.NewReader(rsp["body"].([]byte))
			b, er := ioutil.ReadAll(r)
			if er != nil {
				return "", er
			}

			lclPath := filepath.Join(localFilePath, name)
			if len(fileNameOverride) > 0 {
				lclPath = filepath.Join(localFilePath, fileNameOverride)
			}
			i, er := iout.WriteOut(b, lclPath)
			if er != nil {
				return "", er
			}
			log.Logf(log.DEBUG, "wrote out %d", i)

			// Validate file by sha1
			f := filenative.NewFile(lclPath)
			// Create sha1 hash and encode it
			sha1hash := encode.Encode(f.Hash(sha1.NewHash(sha1.Format(hashenum.SHA1)), true), encodeenum.Hex)
			// Compare hash to original
			if sha1hash != sha1Val {
				return "", fmt.Errorf("downloaded file doesn't match remote by sha1 %s local %s", sha1Val, sha1hash)
			}
			return lclPath, nil
		}
	}

	return "", nil
}
func workerDown(c *Cloud, p *UploaderPart, file *os.File, fileID string) {
	rsp, err := c.DownloadByID(fileID, "bytes="+mathut.FmtInt(int(p.Start))+"-"+mathut.FmtInt(int(p.End)))
	bserr.StopErr(err, "err downloading by id")

	if rsp != nil {
		r := bytes.NewReader(rsp["body"].([]byte))
		b, er := ioutil.ReadAll(r)
		bserr.StopErr(er, "err reading body")

		n, err := file.WriteAt(b, p.Start)
		if err != nil {
			bserr.StopErr(err, "issue writing at")
		}
		fmt.Printf("Part %d wrote out %d\n", p.PartID, n)
	}
}
