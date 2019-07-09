package errs

import (
	"fmt"
	"runtime"

	"github.com/colt3k/utils/store"
)

var (
	B2CloudCodes = store.NewMVKeySet()
)

func init() {
	B2CloudCodes.Add(400, "bad_request")
	B2CloudCodes.Add(400, "bad_bucket_id")
	B2CloudCodes.Add(400, "invalid_bucket_id")     // invalid bucket it
	B2CloudCodes.Add(400, "out_of_range")          // maxfilecount out of range
	B2CloudCodes.Add(400, "file_not_present")      // file not present
	B2CloudCodes.Add(401, "unauthorized")          // application key is bad
	B2CloudCodes.Add(401, "unsupported")           // application key is only supported in a later version of API
	B2CloudCodes.Add(401, "bad_auth_token")        // bad auth token
	B2CloudCodes.Add(401, "expired_auth_token")    // expired auth token
	B2CloudCodes.Add(403, "forbidden")             // reached storage cap limit, or account access may be impacted
	B2CloudCodes.Add(403, "cap_exceeded")          // Usage cap exceeded
	B2CloudCodes.Add(405, "method_not_allowed")    // Only POST is supported
	B2CloudCodes.Add(408, "request_timeout")       // service timed out trying to read your request
	B2CloudCodes.Add(416, "range_not_satisfiable") // Range header in the request is outside the size of the file.
	B2CloudCodes.Add(429, "too_many_requests")     // b2 may limit API requests on per account basis
	B2CloudCodes.Add(500, "internal_error")        // unexpected error occurred
	B2CloudCodes.Add(500, "misc_error")            // unexpected error occurred
	B2CloudCodes.Add(503, "service_unavailable")   // temp unavailable try with exponential back off
	B2CloudCodes.Add(503, "bad_request")           // Timed out while iterating and skipping files
}

type Error interface {
	Code() string
	SetCode(string)
	Message() string
	SetMessage(string)
	Status() int
	SetStatus(int)
	Error() string
}

type B2Error struct {
	CodeStr    string `json:"code"`
	MessageStr string `json:"message"`
	StatusId   int    `json:"status"`
	File       string
	Line       int
}

func New(err error, msg string) Error {
	_, file, line, _ := runtime.Caller(1)
	t := new(B2Error)
	t.MessageStr = fmt.Sprintf("%s\n%+v", msg, err)
	t.StatusId = 500
	t.CodeStr = "misc_error"
	t.File = file
	t.Line = line
	return t
}
func (e *B2Error) Code() string {
	return e.CodeStr
}
func (e *B2Error) SetCode(m string) {
	e.CodeStr = m
}
func (e *B2Error) Message() string {
	return e.MessageStr
}
func (e *B2Error) SetMessage(m string) {
	e.MessageStr = m
}
func (e *B2Error) SetStatus(i int) {
	e.StatusId = i
}
func (e *B2Error) Status() int {
	return e.StatusId
}
func (e *B2Error) Error() string {
	return fmt.Sprintf("%s[%d]\n%d - %s %s", e.File, e.Line, e.StatusId, e.CodeStr, e.MessageStr)
}
