package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/boypt/notiferhub"
	"github.com/boypt/notiferhub/common"
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
	defer resp.Body.Close()
	return nil
}

func cldAPI(api, hash string) error {

	if api == "" {
		return nil
	}

	actions := []string{"stop:" + hash, "delete:" + hash}
	url := fmt.Sprintf("http://%s/api/torrent", api)

	for _, ac := range actions {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(ac)))
		if err != nil {
			log.Println(err)
			continue
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			continue
		}
		resp.Body.Close()
		time.Sleep(time.Second)
	}

	return nil
}

func tgAPI(text ...string) error {
	bottoken := viper.GetString("bottoken")
	if bottoken == "" {
		log.Fatal("token empty")
	}
	bot := tgbot.NewTGBot(bottoken)

	chid, err := strconv.ParseInt(viper.GetString("chatid"), 10, 64)
	common.Must(err)

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
	return bot.SendMsg(chid, strings.Join(text, ""), tgnotify)
}
