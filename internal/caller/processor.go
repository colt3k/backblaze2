package caller

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/colt3k/backblaze2/errs"

	"github.com/colt3k/nglog/ers/bserr"
	log "github.com/colt3k/nglog/ng"
	"github.com/colt3k/utils/netut/hc"
)

var client *hc.Client

// HttpCall data store
type HttpCall struct {
	Method     string
	URL        string
	Header     map[string]string
	ReturnHead bool
	ReturnKeys []string
}

// New HttpCall
func New(method, url string, header map[string]string, returnHead bool) *HttpCall {
	t := new(HttpCall)
	t.Method = method
	t.URL = url
	t.Header = header
	t.ReturnHead = returnHead
	return t
}

// HttpCall method on HttpCall object, pass in reader
func (h *HttpCall) HttpCall(data io.Reader) (map[string]interface{}, errs.Error) {
	log.Logf(log.DEBUG, "called httpCallData for Method %s URL: %s", h.Method, h.URL)
	tmp := make(map[string]interface{}, 0)

	if client == nil {
		// set to timeout after a day per request, accomodates file uploads
		client = hc.NewClient(hc.HttpClientRequestTimeout(3600), hc.DisableVerifyClientCert(false), hc.HttpClientResponseHeaderTimeout(300))
	}
	var resp, err = client.Fetch(h.Method, h.URL, nil, h.Header, data)

	if resp != nil {
		defer resp.Body.Close()
	}
	// 202 occurs when a http.DELETE is ran
	if err != nil && err.Error() != "202 Accepted" && strings.TrimSpace(err.Error()) != "206" {
		er := errs.New(err, "")
		if resp != nil {
			er.SetMessage("on " + h.URL)
			er.SetStatus(resp.StatusCode)
		}

		if resp != nil && resp.Body != nil {
			body, errRA := ioutil.ReadAll(resp.Body)
			if errRA != nil {
				return nil, er
			}
			log.Logln(log.DEBUG, "Body: ", string(body))
			_ = json.Unmarshal(body, &er)
			if len(strings.TrimSpace(er.Message())) <= 0 {
				er.SetMessage("on " + h.URL)
			}
			er.SetStatus(resp.StatusCode)
		}
		if resp != nil && resp.StatusCode == 401 {
			return nil, er
		}
		if strings.Index(err.Error(), "certificate signed by unknown authority") > -1 {
			er.SetMessage("invalid certificate, unknown authority '" + h.URL + "'")
			return nil, er
		} else if strings.Index(err.Error(), "Network call did not return SUCCESS!") > -1 {
			er.SetMessage("network did not return SUCCESS '" + h.URL + "'")
			return nil, er
		}
		if len(er.Message()) == 0 {
			er.SetMessage(fmt.Sprintf("site unreachable %s\n%+v", h.URL, err.Error()))
		}
		return nil, er
	}

	if h.ReturnHead {
		for name, value := range resp.Header {
			tmp[name] = value
		}
		//for _, d := range h.ReturnKeys {
		//	tmp[d] = resp.Header.Get(d)
		//}
	}

	// Read body to buffer
	body, err := ioutil.ReadAll(resp.Body)
	if bserr.Err(err, "Error reading body") {
		er := errs.New(err, "")
		er.SetStatus(500)
		er.SetMessage("unable to read response")
		return nil, er
	}

	tmp["body"] = body
	return tmp, nil
}
