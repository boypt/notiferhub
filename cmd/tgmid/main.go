package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/boypt/notiferhub/common"
	"github.com/boypt/notiferhub/tgbot"
	"github.com/spf13/viper"
)

var (
	webListen string
	bot       *tgbot.TGBot
	botchid   int64
)

func startTaskWeb() {
	http.HandleFunc("/tgmid/messages", midMessage)
	fmt.Println("Starting tgmid server at port", webListen)
	if err := http.ListenAndServe(webListen, nil); err != nil {
		log.Fatal(err)
	}
}

func midMessage(w http.ResponseWriter, r *http.Request) {
	defer fmt.Fprintf(w, "GOT")
	if r.Method != "POST" {
		return
	}

	msg := r.FormValue("message")
	if len(msg) > 0 {
		go bot.SendMsg(botchid, msg, false)
	}
}

func main() {

	flag.StringVar(&webListen, "webport", "127.0.0.1:7267", "web server listen")
	viper.SetConfigName("cmddl")
	viper.AddConfigPath("/srv")
	viper.AddConfigPath(".")
	common.Must(viper.ReadInConfig())

	log.Println("using config: ", viper.ConfigFileUsed())

	bot = tgbot.NewTGBot(viper.GetString("bottoken"))
	botchid = viper.GetInt64("chatid")

	startTaskWeb()
}
