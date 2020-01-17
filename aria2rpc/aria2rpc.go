package aria2rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	// "log"
	"math/rand"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/boypt/notiferhub"
)

type Aria2Req struct {
	ID      string        `json:"id,omitempty"`
	Method  string        `json:"method,omitempty"`
	Params  []interface{} `json:"params,omitempty"`
	JSONRPC string        `json:"jsonrpc,omitempty"`
}

type Aria2Err struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Aria2Resp struct {
	ID      string      `json:"id,omitempty"`
	JSONRPC string      `json:"jsonrpc,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Aria2Err   `json:"error,omitempty"`
}

type Aria2RPC struct {
	Token     string
	ServerURL string
	Timeout   time.Duration
	client    http.Client
}

type Aria2Status map[string]string

func (s Aria2Status) String() string {
	speed, _ := strconv.ParseInt(s["downloadSpeed"], 10, 64)
	progress := s.GetProgress()
	return fmt.Sprintf("%s %.2f%% %s/s", s.Get("status"), progress, notifierhub.HumaneSize(speed))
}

func (s Aria2Status) Get(k string) string {
	if s, ok := s[k]; ok {
		return s
	}
	return "unknow"
}

func (s Aria2Status) GetProgress() float64 {
	var progress float64
	completed, _ := strconv.ParseInt(s["completedLength"], 10, 64)
	total, _ := strconv.ParseInt(s["totalLength"], 10, 64)
	if total > 0 {
		progress = float64(completed) / float64(total) * 100
	}
	return progress
}

func NewAria2RPC(token, url string) *Aria2RPC {
	c := &Aria2RPC{
		Token:     token,
		ServerURL: url,
		Timeout:   30 * time.Second,
	}

	c.client = http.Client{
		Timeout: c.Timeout,
	}

	return c
}

func (a *Aria2RPC) CallAria2Req(req *Aria2Req) (*Aria2Resp, error) {

	req.ID = strconv.Itoa(rand.Intn(9999))
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	hreq, err := http.NewRequest("POST", a.ServerURL, bytes.NewBuffer(payload))
	hreq.Header.Set("Content-Type", "application/json")
	hresp, err := a.client.Do(hreq)
	if err != nil {
		return nil, err
	}

	defer hresp.Body.Close()
	ret, _ := ioutil.ReadAll(hresp.Body)
	if hresp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CallAria2Req remote non ok: '%s'", string(ret))
	}
	resp := &Aria2Resp{}
	if err := json.Unmarshal(ret, resp); err != nil {
		return nil, fmt.Errorf("%w '%s'", err, string(ret))
	}

	if resp.ID != req.ID {
		return nil, errors.New("what??? req ID unmached")
	}

	if resp.Error != nil && resp.Error.Code != 0 {
		return nil, fmt.Errorf("aria2 error: code %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}
	return resp, nil
}

func (a *Aria2RPC) GetVersion() (string, error) {

	req := &Aria2Req{
		Method:  "aria2.getVersion",
		JSONRPC: "2.0",
		Params:  []interface{}{fmt.Sprintf("token:%s", a.Token)},
	}
	resp, err := a.CallAria2Req(req)
	if err != nil {
		return "", err
	}

	r := resp.Result.(map[string]interface{})
	v := r["version"].(string)
	return v, nil
}

func (a *Aria2RPC) GetGlobalStat() (map[string]string, error) {
	req := &Aria2Req{
		Method:  "aria2.getGlobalStat",
		JSONRPC: "2.0",
		Params:  []interface{}{fmt.Sprintf("token:%s", a.Token)},
	}
	resp, err := a.CallAria2Req(req)
	if err != nil {
		return nil, err
	}

	// log.Printf("%#v\n", resp)
	r := resp.Result.(map[string]interface{})
	resmap := map[string]string{}
	for k, v := range r {
		resmap[k] = v.(string)
	}

	return resmap, nil
}

func (a *Aria2RPC) TellStatus(gid string) (Aria2Status, error) {
	req := &Aria2Req{
		Method:  "aria2.tellStatus",
		JSONRPC: "2.0",
		Params:  []interface{}{fmt.Sprintf("token:%s", a.Token), gid, []string{"gid", "status", "totalLength", "completedLength", "downloadSpeed", "files"}},
	}
	resp, err := a.CallAria2Req(req)
	if err != nil {
		return nil, err
	}

	// log.Printf("%#v", resp.Result)
	st := Aria2Status{}
	for k, sv := range resp.Result.(map[string]interface{}) {
		switch v := sv.(type) {
		case string:
			st[k] = v
		case []interface{}:
			if k == "files" {
				name := ""
				for _, f := range v {
					if fm, ok := f.(map[string]interface{}); ok {
						name += path.Base(fmt.Sprintf("%v", fm["path"]))
					}
				}
				st[k] = name
			}
		}
	}

	return st, nil
}

func (a *Aria2RPC) AddUri(uris []string, name string) (*Aria2Resp, error) {

	opt := struct {
		Out string `json:"out"`
	}{name}
	req := &Aria2Req{
		Method:  "aria2.addUri",
		JSONRPC: "2.0",
		Params:  []interface{}{fmt.Sprintf("token:%s", a.Token), uris, opt},
	}
	resp, err := a.CallAria2Req(req)
	if err != nil {
		return nil, fmt.Errorf("AddUri Fail to call CallAria2Req: %w", err)
	}

	return resp, nil
}
