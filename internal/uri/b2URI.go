package uri

const (
	B2BaseURL    = "https://api.backblazeb2.com"
	B2API         = "/b2api"
	B2Ver         = "/v2"
	B2APIVer      = B2API + B2Ver
	B2AuthAccount = B2BaseURL + B2APIVer + "/b2_authorize_account"


	B2ListKeys  = B2APIVer + "/b2_list_keys"
	B2CreateKey = B2APIVer + "/b2_create_key"
	B2DeleteKey = B2APIVer + "/b2_delete_key"

	B2CreateBucket = B2APIVer + "/b2_create_bucket"
	B2UpdateBucket = B2APIVer + "/b2_update_bucket"
	B2DeleteBucket = B2APIVer + "/b2_delete_bucket"
	B2ListBuckets  = B2APIVer + "/b2_list_buckets"

	B2GetFileInfo       = B2APIVer + "/b2_get_file_info"
	B2HideFile          = B2APIVer + "/b2_hide_file"
	B2ListFileNames     = B2APIVer + "/b2_list_file_names"
	B2ListFileVersions  = B2APIVer + "/b2_list_file_versions"
	B2DeleteFileVersion = B2APIVer + "/b2_delete_file_version"

	//B2UploadFile               = B2APIVer + "/b2_upload_file"	(GET UPLOAD URL used instead)
	B2GetUploadURL             = B2APIVer + "/b2_get_upload_url"
	B2GetUploadPartURL         = B2APIVer + "/b2_get_upload_part_url"
	//B2UploadPart               = B2APIVer + "/b2_upload_part"(GET UPLOAD PART URL used instead)
	B2ListParts                = B2APIVer + "/b2_list_parts"
	B2ListUnfinishedLargeFiles = B2APIVer + "/b2_list_unfinished_large_files"
	B2StartLargeFile           = B2APIVer + "/b2_start_large_file"

	B2FinishLargeFile = B2APIVer + "/b2_finish_large_file"
	B2CancelLargeFile = B2APIVer + "/b2_cancel_large_file"

	B2GetDownloadAuth = B2APIVer + "/b2_get_download_authorization"
	B2DownloadFileById = B2APIVer + "/b2_download_file_by_id"
	B2DownloadFileByName = B2APIVer + "/b2_download_file_by_name"
	
)
