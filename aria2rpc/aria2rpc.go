package aria2rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boypt/notiferhub"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
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
	completed, _ := strconv.ParseInt(s["completedLength"], 10, 64)
	total, _ := strconv.ParseInt(s["totalLength"], 10, 64)
	speed, _ := strconv.ParseInt(s["downloadSpeed"], 10, 64)
	var progress float64
	if total > 0 {
		progress = float64(completed) / float64(total) * 100
	}
	return fmt.Sprintf("%s %.2f%% %s/s", s.GetStatus(), progress, notifierhub.HumaneSize(speed))
}

func (s Aria2Status) GetStatus() string {
	if s, ok := s["status"]; ok {
		return s
	}
	return "unknow"
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
	resp := &Aria2Resp{}
	if err := json.Unmarshal(ret, resp); err != nil {
		return nil, err
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
		Params:  []interface{}{fmt.Sprintf("token:%s", a.Token), gid, []string{"gid", "status", "totalLength", "completedLength", "downloadSpeed"}},
	}
	resp, err := a.CallAria2Req(req)
	if err != nil {
		return nil, err
	}

	st := Aria2Status{}
	for k, v := range resp.Result.(map[string]interface{}) {
		if sv, ok := v.(string); ok {
			st[k] = sv
		}
	}

	return st, nil
}

func (a *Aria2RPC) AddUri(uri, name string) (*Aria2Resp, error) {

	opt := struct {
		Out string `json:"out"`
	}{name}
	req := &Aria2Req{
		Method:  "aria2.addUri",
		JSONRPC: "2.0",
		Params:  []interface{}{fmt.Sprintf("token:%s", a.Token), []string{uri}, opt},
	}
	resp, err := a.CallAria2Req(req)
	if err != nil {
		return nil, fmt.Errorf("AddUri Fail to call CallAria2Req: %w", err)
	}

	return resp, nil
}
