package tgbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type TGBot struct {
	botBaseURL string
	Client     http.Client
}

func NewTGBot(token string) *TGBot {

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	baseURL := fmt.Sprintf("https://api.telegram.org/bot%s/", token)
	b := &TGBot{
		botBaseURL: baseURL,
		Client:     client,
	}

	return b
}

/*TGMessage ...
https://core.telegram.org/bots/api#sendmessage
*/
type TGMessage struct {
	ChatID                int64  `json:"chat_id,omitempty"`
	Text                  string `json:"text,omitempty"`
	ParseMode             string `json:"parse_mode,omitempty"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
	DisableNotification   bool   `json:"disable_notification,omitempty"`
}

type TGResp struct {
	OK     bool
	Result interface{}
}

func (b *TGBot) post(action string, payload []byte) ([]byte, error) {
	req, _ := http.NewRequest("POST", b.botBaseURL+action, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	ret, _ := ioutil.ReadAll(resp.Body)
	return ret, nil
}

func (b *TGBot) SendMsg(id int64, text string, notify bool) error {

	msg := TGMessage{
		ChatID:              id,
		Text:                text,
		ParseMode:           "Markdown",
		DisableNotification: !notify,
	}

	payload, _ := json.Marshal(msg)

	ret, err := b.post("sendMessage", payload)
	if err != nil {
		return err
	}

	r := &TGResp{}
	if err := json.Unmarshal(ret, r); err != nil {
		return err
	}

	if !r.OK {
		return fmt.Errorf("%#v", r.Result)
	}

	return nil
}
