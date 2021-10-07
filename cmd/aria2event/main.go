package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/boypt/notiferhub/aria2rpc"
	"github.com/boypt/notiferhub/common"
	"github.com/boypt/notiferhub/tgbot"
	"github.com/spf13/viper"
)

var (
	a2httpclient *aria2rpc.Aria2RPC
	a2wsclient   *aria2rpc.Aria2WSRPC
	bot          *tgbot.TGBot
	botchid      int64
)

func main() {
	go a2wsclient.WsListenMsg()
	go a2wsclient.KeepAlive(30 * time.Second)
	log.Println("Listeing ...")

	for msg := range a2wsclient.WsQueue {
		ev := &aria2rpc.Aria2Req{}
		if err := json.Unmarshal(msg, ev); err != nil {
			log.Println("error", err)
		}
		// log.Println(string(msg))
		if ev.Method == "aria2.onDownloadComplete" {
			pmap := ev.Params[0].(map[string]interface{})
			go tgNotify(pmap["gid"].(string))
		}
	}

	log.Panicln("main func exit")
}

func tgNotify(gid string) {
	s, err := a2httpclient.TellStatus(gid)
	if err != nil {
		log.Println(err)
		return
	}
	bot.SendMsg(botchid,
		fmt.Sprintf("*%s*\nStatus: %s", s.Get("files"), s.Get("status")),
		false)
}

func init() {
	log.SetFlags(0)

	viper.SetConfigName("cmddl")
	viper.AddConfigPath("/srv")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("Error config file", err)
	}
	log.Println("using config: ", viper.ConfigFileUsed())

	bottoken := viper.GetString("bottoken")
	if bottoken == "" {
		log.Fatal("bottoken empty")
	}
	bot = tgbot.NewTGBot(bottoken)
	botchid = common.Must2(strconv.ParseInt(viper.GetString("chatid"), 10, 64)).(int64)

	a2token := viper.GetString("aria2_token")
	a2rpc := viper.GetString("aria2_url")
	a2url, err := url.Parse(a2rpc)
	common.Must(err)

	switch a2url.Scheme {
	case "ws", "http":
		a2url.Scheme = "ws"
		a2wsclient = aria2rpc.NewAria2WSRPC(a2token, a2url.String())
		a2url.Scheme = "http"
		a2httpclient = aria2rpc.NewAria2RPC(a2token, a2url.String())
	case "wss", "https":
		a2url.Scheme = "wss"
		a2wsclient = aria2rpc.NewAria2WSRPC(a2token, a2url.String())
		a2url.Scheme = "https"
		a2httpclient = aria2rpc.NewAria2RPC(a2token, a2url.String())
	default:
		log.Fatalln("aria2_url not found")
	}
}
