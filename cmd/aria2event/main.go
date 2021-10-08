package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/boypt/notiferhub/aria2rpc"
	"github.com/boypt/notiferhub/common"
	"github.com/boypt/notiferhub/tgbot"
	"github.com/spf13/viper"
)

var (
	a2wsclient *aria2rpc.Aria2WSRPC
	bot        *tgbot.TGBot
	botchid    int64
)

func main() {
	a2wsclient.WebsocketMsgBackgroundRoutine()
	log.Println("Listeing ...")

	log.Println(a2wsclient.GetVersion())
	log.Println(a2wsclient.GetSessionInfo())

	for ev := range a2wsclient.NotifyQueue {
		gid := ev.Params[0].(map[string]interface{})["gid"].(string)
		switch ev.Method {
		case "aria2.onDownloadComplete":
			if s, err := a2wsclient.TellStatus(gid); err == nil {
				msg := fmt.Sprintf("*%s*\nStatus: %s", s.Get("files"), s.Get("status"))
				go bot.SendMsg(botchid, msg, false)
			}
		case "aria2.onDownloadError":
			if s, err := a2wsclient.TellStatus(gid); err == nil {
				msg := fmt.Sprintf("*%s*\nStatus:Error (%s)", s.Get("files"), s.Get("errorMessage"))
				go bot.SendMsg(botchid, msg, false)
			}
		default:
			log.Println("unprocess event:", ev.Method)
		}
	}

	log.Panicln("main func exit")
}

func init() {
	log.SetFlags(0)

	viper.SetConfigName("cmddl")
	viper.AddConfigPath("/srv")
	viper.AddConfigPath(".")
	common.Must(viper.ReadInConfig())

	log.Println("using config: ", viper.ConfigFileUsed())

	bot = tgbot.NewTGBot(viper.GetString("bottoken"))
	botchid = viper.GetInt64("chatid")

	a2rpc := viper.GetString("aria2_url")
	a2url, err := url.Parse(a2rpc)
	common.Must(err)

	switch a2url.Scheme {
	case "ws", "http":
		a2url.Scheme = "ws"
	case "wss", "https":
		a2url.Scheme = "wss"
	default:
		log.Fatalln("aria2_url not found")
	}
	a2wsclient = aria2rpc.NewAria2WSRPC(viper.GetString("aria2_token"), a2url.String())
}
