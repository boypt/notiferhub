package aria2rpc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	// "log"

	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/boypt/notiferhub/common"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

func (a Aria2Err) Error() string {
	return a.Message
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

func (s Aria2Status) Speed() int64 {
	speed, _ := strconv.ParseInt(s["downloadSpeed"], 10, 64)
	return speed
}

func (s Aria2Status) String() string {
	progress := s.GetProgress()
	return fmt.Sprintf("%s %.2f%% %s/s", s.Get("status"), progress, common.HumaneSize(s.Speed()))
}

func (s Aria2Status) Status() string {
	return fmt.Sprintf("%s %s", s.Get("files"), s.Get("status"))
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

func NewAria2RPCTLS(token, url string, skipCert bool) (*Aria2RPC, error) {
	c := &Aria2RPC{
		Token:     token,
		ServerURL: url,
		Timeout:   30 * time.Second,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipCert},
	}

	c.client = http.Client{
		Transport: tr,
		Timeout:   c.Timeout,
	}

	return c, nil
}

func (a *Aria2RPC) CallAria2Req(req *Aria2Req) (*Aria2Resp, error) {

	req.ID = uuid.NewString()
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
	if hresp.StatusCode >= 500 {
		io.Copy(ioutil.Discard, hresp.Body)
		return nil, fmt.Errorf("jsonrpc returned %d %s", hresp.StatusCode, hresp.Status)
	}

	ret, _ := ioutil.ReadAll(hresp.Body)
	resp := &Aria2Resp{}
	if err := json.Unmarshal(ret, resp); err != nil {
		return nil, fmt.Errorf("%w '%s'", err, string(ret))
	}

	if resp.Error != nil {
		return nil, *resp.Error
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

func (a *Aria2RPC) GetGlobalStat() (map[string]int64, error) {
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
	resmap := map[string]int64{}
	for k, v := range r {
		val, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			continue
		}
		resmap[k] = val
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

func (a *Aria2RPC) Pause(gid string) error {
	req := &Aria2Req{
		Method:  "aria2.pause",
		JSONRPC: "2.0",
		Params:  []interface{}{fmt.Sprintf("token:%s", a.Token), gid},
	}
	resp, err := a.CallAria2Req(req)
	if err != nil {
		return err
	}

	if rgid, ok := resp.Result.(string); ok {
		log.Println("aria2 unpused", rgid)
	}
	return nil
}

func (a *Aria2RPC) UnPause(gid string) error {
	req := &Aria2Req{
		Method:  "aria2.unpause",
		JSONRPC: "2.0",
		Params:  []interface{}{fmt.Sprintf("token:%s", a.Token), gid},
	}
	resp, err := a.CallAria2Req(req)
	if err != nil {
		return err
	}

	if rgid, ok := resp.Result.(string); ok {
		log.Println("aria2 unpused", rgid)
	}
	return nil
}

func (a *Aria2RPC) AddUri(uris []string, name string) (string, error) {

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
		return "", fmt.Errorf("AddUri Fail to call CallAria2Req: %w", err)
	}

	if gid, ok := resp.Result.(string); ok {
		return gid, nil
	}

	return "", fmt.Errorf("gid can get from result, %#v", resp)
}

type Aria2WSRPC struct {
	Token       string
	ServerURL   string
	Timeout     time.Duration
	wsclient    *websocket.Conn
	Close       chan struct{}
	WriteQueue  chan *Aria2Req
	NotifyQueue chan *Aria2Req
	respMap     map[string]chan *Aria2Resp
}

func NewAria2WSRPC(token, rpcurl string) *Aria2WSRPC {
	c := &Aria2WSRPC{
		Token:     token,
		ServerURL: rpcurl,
		Timeout:   30 * time.Second,
	}

	d := websocket.Dialer{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		EnableCompression: true,
	}

	client, _, err := d.Dial(rpcurl, nil)
	if err != nil {
		log.Fatal("ws dial:", rpcurl, err)
	}
	c.wsclient = client
	c.Close = make(chan struct{})
	c.WriteQueue = make(chan *Aria2Req)
	c.NotifyQueue = make(chan *Aria2Req)
	c.respMap = make(map[string]chan *Aria2Resp)

	return c
}

func (a *Aria2WSRPC) WebsocketMsgBackgroundRoutine() {
	go func() {
		for {
			_, message, err := a.wsclient.ReadMessage()
			if err != nil {
				log.Println("WebsocketMsgBackgroundRoutine: ", err)
				a.wsclient.Close()
				defer close(a.Close)
				return
			}

			req := &Aria2Req{}
			if err := json.Unmarshal(message, req); err == nil && strings.HasPrefix(req.Method, "aria2.on") {
				a.NotifyQueue <- req
				continue
			}

			rsp := &Aria2Resp{}
			if err := json.Unmarshal(message, rsp); err == nil {

				if rsp.ID != "" {
					if ch, ok := a.respMap[rsp.ID]; ok {
						ch <- rsp
					} else {
						log.Println("reqid not found:", rsp.ID)
					}
				}

				if rsp.Error != nil {
					if rsp.Error.Code != -32600 { // likely a ping response
						log.Println("got error:", rsp.Error.Code, rsp.Error.Message)
					}
				}

				continue
			}

			log.Println("unknow message: ", string(message))
		}
	}()

	go func() {
		for {
			select {
			case req := <-a.WriteQueue:
				err := a.wsclient.WriteJSON(req)
				if err != nil {
					log.Println("WebsocketMsgBackgroundRoutine: ", err)
					close(a.Close)
					return
				}
			case <-a.Close:
				return
			}
		}
	}()

	go func() {
		tk := time.NewTicker(30 * time.Second)
		defer tk.Stop()

		for {
			select {
			case <-a.Close:
				return
			case <-tk.C:
			}

			if err := a.wsclient.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second*10)); err != nil {
				log.Println("ping error", err)
				close(a.Close)
			}
		}
	}()
}

func (a *Aria2WSRPC) CallAria2Req(req *Aria2Req) (*Aria2Resp, error) {
	recv := make(chan *Aria2Resp)
	req.ID = uuid.NewString()
	a.respMap[req.ID] = recv
	a.WriteQueue <- req

	tr := time.NewTimer(time.Second * 10)
	defer tr.Stop()

	select {
	case resp := <-recv:
		delete(a.respMap, req.ID)
		close(recv)
		return resp, nil
	case <-tr.C:
	}
	return nil, fmt.Errorf("call req timeout")
}

func (a *Aria2WSRPC) TellStatus(gid string) (Aria2Status, error) {
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

func (a *Aria2WSRPC) stringMethod(method string) (string, error) {

	req := &Aria2Req{
		Method:  method,
		JSONRPC: "2.0",
		Params:  []interface{}{fmt.Sprintf("token:%s", a.Token)},
	}
	resp, err := a.CallAria2Req(req)
	if err != nil {
		return "", err
	}

	ret := ""
	for _, sv := range resp.Result.(map[string]interface{}) {
		switch v := sv.(type) {
		case string:
			ret += fmt.Sprintf("%s,", v)
		case []interface{}:
			for _, val := range v {
				if s, ok := val.(string); ok {
					ret += fmt.Sprintf("%s,", s)
				}
			}
		}
	}

	return ret, nil
}

func (a *Aria2WSRPC) GetVersion() (string, error) {
	return a.stringMethod("aria2.getVersion")
}

func (a *Aria2WSRPC) GetSessionInfo() (string, error) {
	return a.stringMethod("aria2.getSessionInfo")
}
