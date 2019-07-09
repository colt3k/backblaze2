package perms

import "github.com/colt3k/backblaze2/b2api"

func authorized(authd *b2api.AuthorizationResp, perm string) bool {
	var allowed bool
	if authd == nil {
		return allowed
	}
	for _, k := range authd.Allowed.Capability {
		if perm == k {
			allowed = true
			break
		}
	}
	return allowed
}


func CancelLargeFile(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "writeFiles")
}
func FinishLargeFile(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "writeFiles")
}
func StartLargeFile(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "writeFiles")
}
func GetUploadPartURL(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "writeFiles")
}
func GetUploadURL(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "writeFiles")
}
func ListParts(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "writeFiles")
}
func ListUnfinishedLargeFiles(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "listFiles")
}
//func UploadPart(authd *b2api.AuthorizationResp) bool {
//	return authorized(authd, "writeFiles")
//}


func CreateBucket(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "writeBuckets")
}
func ListBuckets(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "listBuckets")
}
func UpdateBucket(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "writeBuckets")
}
func DeleteBucket(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "deleteBuckets")
}


func CreateKeys(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "writeKeys")
}
func DeleteKeys(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "deleteKeys")
}
func ListKeys(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "listKeys")
}


func DeleteFiles(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "deleteFiles")
}
func DownloadFile(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "readFiles")
}
func GetDownloadAuth(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "shareFiles")
}
func GetFileInfo(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "readFiles")
}
func HideFile(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "writeFiles")
}
func ListFileNames(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "listFiles")
}
func ListFileVersions(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "listFiles")
}
func UploadFile(authd *b2api.AuthorizationResp) bool {
	return authorized(authd, "writeFiles")
}