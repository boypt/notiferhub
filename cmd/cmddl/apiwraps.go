package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/boypt/notiferhub/tgbot"
	"github.com/go-redis/redis"
)

func chanAPI(text ...string) error {
	purl := os.Getenv("chanapi")
	token := os.Getenv("chantoken")
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

func cldAPI(api, hash string) {

	if api == "" {
		return
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
}

func fetchRedis() {
	tgnotify = true

	rekey := "notiferhub_tg"
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	if _, err := client.Ping().Result(); err != nil {
		fmt.Println(err)
		return
	}

	if val, err := client.Get(rekey).Result(); err == nil && val == "sent" {
		fmt.Println("no notify")
		tgnotify = false
	}

	if _, err := client.Set(rekey, "sent", time.Minute*5).Result(); err != nil {
		fmt.Println(err)
		return
	}
}

func tgAPI(text ...string) error {
	bottoken := os.Getenv("BOTTOKEN")
	if bottoken == "" {
		log.Fatal("token empty")
	}
	bot := tgbot.NewTGBot(bottoken)

	chid, err := strconv.ParseInt(os.Getenv("CHATID"), 10, 64)
	if err != nil {
		log.Fatal("chatid parse fail")
	}

	fetchRedis()
	return bot.SendMsg(chid, strings.Join(text, ""), tgnotify)
}
