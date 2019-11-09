package aria2rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Aria2Req struct {
	ID      string        `json:"id,omitempty"`
	Method  string        `json:"method,omitempty"`
	Params  []interface{} `json:"params,omitempty"`
	JSONRPC string        `json:"jsonrpc,omitempty"`
}

type Aria2Resp struct {
	ID      string      `json:"id,omitempty"`
	JSONRPC string      `json:"jsonrpc,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

type Aria2RPC struct {
	Token     string
	ServerURL string
	Timeout   time.Duration
}

func NewAria2RPC(token, url string) *Aria2RPC {
	c := &Aria2RPC{
		Token:     token,
		ServerURL: url,
		Timeout:   30 * time.Second,
	}

	return c
}

func (a *Aria2RPC) CallAria2Method(method string, args []string) (*Aria2Resp, error) {

	randID := strconv.Itoa(rand.Intn(9999))
	req := Aria2Req{
		ID:      randID,
		Method:  method,
		JSONRPC: "2.0",
		Params:  []interface{}{fmt.Sprintf("token:%s", a.Token), args},
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	hreq, err := http.NewRequest("POST", a.ServerURL, bytes.NewBuffer(payload))
	hreq.Header.Set("Content-Type", "application/json")
	c := http.Client{
		Timeout: a.Timeout,
	}
	hresp, err := c.Do(hreq)
	if err != nil {
		return nil, err
	}

	defer hresp.Body.Close()
	ret, _ := ioutil.ReadAll(hresp.Body)

	resp := &Aria2Resp{}
	if err := json.Unmarshal(ret, resp); err != nil {
		return nil, err
	}

	if resp.ID != randID {
		return nil, errors.New("what??? ID unmached")
	}
	return resp, nil
}

func (a *Aria2RPC) GetVersion() (string, error) {
	resp, err := a.CallAria2Method("aria2.getVersion", []string{})
	if err != nil {
		return "", err
	}

	r := resp.Result.(map[string]interface{})
	v := r["version"].(string)
	return v, nil
}

func (a *Aria2RPC) GetGlobalStat() (map[string]string, error) {
	resp, err := a.CallAria2Method("aria2.getGlobalStat", []string{})
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

func (a *Aria2RPC) AddUris(uris []string) error {
	resp, err := a.CallAria2Method("aria2.addUri", uris)
	if err != nil {
		return err
	}

	log.Printf("%#v\n", resp)
	return nil
}

func JustAddURL(url string) error {
	rpc := NewAria2RPC(os.Getenv("aria2_token"), os.Getenv("aria2_url"))
	err := rpc.AddUris([]string{url})
	if err != nil {
		return fmt.Errorf("%s, %w", url, err)
	}
	return nil
}
