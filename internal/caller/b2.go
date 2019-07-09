package caller

import (
	"encoding/json"
	"io"

	"github.com/colt3k/backblaze2/errs"
)

// 10GB of free storage, unlimited free uploads, and 1GB of downloads each day

/*
Abilities
	1. Create Application Key
		Use master app key and b2_authorize_account to create an authorization token that is capable of creating application keys.
		Authorization tokens are only good for 24 hours.
		These API calls can take up to 5 minutes to process
		a. b2_create_key
		b. b2_list_keys`
		c. b2_delete_key
*/

// MakeCall to http client
func MakeCall(method, URI string, msg io.Reader, header map[string]string) (map[string]interface{}, errs.Error) {
	c := New(method, URI, header, true)
	mapData, err := c.HttpCall(msg)
	if err != nil {
		return nil, err
	}
	return mapData, nil
}

// UnMarshalRequest to byte array
func UnMarshalRequest(req interface{}) ([]byte, error) {
	msg, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
