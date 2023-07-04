package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"notiferhub/common"

	"notiferhub/tgbot"

	"github.com/spf13/viper"
)

var (
	webListen string
	bot       *tgbot.TGBot
	botchid   string
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
	flag.Parse()
	viper.SetConfigName("cmddl")
	viper.AddConfigPath("/srv")
	viper.AddConfigPath(".")
	common.Must(viper.ReadInConfig())

	log.Println("using config: ", viper.ConfigFileUsed())

	bot = tgbot.NewTGBot(viper.GetString("bottoken"))
	chanid := viper.GetString("chatid")

	ret, err := bot.SendMsg(chanid, "TGMid bot startted", false)
	if err != nil {
		log.Printf("ret: %#v", ret)
		log.Fatalln(err)
	}

	ch, _ := ret.Result.(map[string]interface{})["chat"].(map[string]interface{})
	botchid = fmt.Sprintf("%.0f", ch["id"].(float64))

	log.Println("got botchid:", botchid)
	startTaskWeb()
}
