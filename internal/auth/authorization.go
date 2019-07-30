package auth

import (
	"runtime"
	"strconv"

	"github.com/colt3k/utils/osut"
)

// BuildAuthMap build authorization map
func BuildAuthMap(authToken string) map[string]string {
	header := make(map[string]string)
	header["Authorization"] = authToken
	header["charset"] = "utf-8"
	platform := osut.OS()

	header["User-Agent"] = "cloudstore/0.0.1+" + runtime.GOOS + "/" + strconv.Itoa(platform.VersionMajor) + "." + strconv.Itoa(platform.VersionMinor)
	return header
}

