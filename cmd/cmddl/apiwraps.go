package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	notifierhub "github.com/boypt/notiferhub"
	"github.com/boypt/notiferhub/tgbot"
	"github.com/spf13/viper"
)

func chanAPI(text ...string) error {
	purl := viper.GetString("chanapi")
	token := viper.GetString("chantoken")
	if purl == "" || token == "" {
		return nil
	}

	if len(text) < 2 {
		return nil
	}

	text[1] = strings.ReplaceAll(text[1], "*", "")
	bd := struct {
		Title   string `json:"title,omitempty"`
		Content string `json:"content,omitempty"`
	}{text[0], text[1]}
	b, _ := json.Marshal(bd)

	req, err := http.NewRequest("POST", purl, bytes.NewBuffer(b))
	req.Header.Set("mkt-token", token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()
	return nil
}

func tgAPI(text ...string) error {
	bottoken := viper.GetString("bottoken")
	if bottoken == "" {
		log.Fatal("token empty")
	}
	bot := tgbot.NewTGBot(bottoken)

	chid := viper.GetString("chatid")

	tgnotify := true
	if notifierhub.RedisClient != nil {
		rekey := "notiferhub_tg"
		if val, err := notifierhub.RedisClient.Get(rekey).Result(); err == nil && val == "sent" {
			log.Println("[tgAPI]: no notify")
			tgnotify = false
		}

		if _, err := notifierhub.RedisClient.Set(rekey, "sent", time.Minute*5).Result(); err != nil {
			fmt.Println(err)
		}
	}
	_, err := bot.SendMsg(chid, strings.Join(text, ""), tgnotify)
	return err
}
