package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/boypt/notiferhub/aria2rpc"
	"github.com/boypt/notiferhub/common"
	"github.com/spf13/viper"
)

var (
	a2rpc    string
	tgmidurl string
)

func postMessage(msg string) {
	//one-line post request/response...
	response, err := http.PostForm(tgmidurl, url.Values{"message": {msg}})

	//okay, moving on...
	if err != nil {
		//handle postform error
	}

	io.Copy(io.Discard, response.Body)
	defer response.Body.Close()
}

func main() {

	test := flag.Bool("test", false, "fire test")
	flag.Parse()
	if *test {
		postMessage("test message")
		return
	}

	a2wsclient := aria2rpc.NewAria2WSRPC(viper.GetString("aria2_token"), a2rpc)
	a2wsclient.WebsocketMsgBackgroundRoutine()
	log.Println("Listeing ...")

	log.Println(a2wsclient.GetVersion())
	log.Println(a2wsclient.GetSessionInfo())

	tsmap := make(map[string]time.Time)

	for {
		method := ""
		gid := ""

		select {
		case ev := <-a2wsclient.NotifyQueue:
			gid = ev.Params[0].(map[string]interface{})["gid"].(string)
			method = ev.Method
		case <-a2wsclient.Close:
			log.Println("a2wsclient.Close closed")
			return
		}

		switch method {
		case "aria2.onDownloadStart":
			tsmap[gid] = time.Now()
			log.Println("start, ", gid)
		case "aria2.onDownloadComplete":
			log.Println("complete, ", gid)

			if s, err := a2wsclient.TellStatus(gid); err == nil {
				msg := ""
				fn := s.Get("files")
				tl := s.Get("totalLength")
				ts, exists := tsmap[gid]
				if exists {
					if tlen, err := strconv.ParseInt(tl, 10, 64); err == nil {
						taskDur := time.Since(ts)
						secs := taskDur.Seconds()
						speed := float64(tlen) / secs
						speedText := common.HumaneSize(int64(speed))
						log.Println("completed", gid, fn, speedText)
						msg = fmt.Sprintf("*%s*\nStatus: *complete*\nDur: *%s*\nAvg: *%s/s*", fn, common.KitchenDuration(taskDur), speedText)
					}
					delete(tsmap, gid)

				} else {
					msg = fmt.Sprintf("*%s*\nStatus: *complete*", fn)
				}

				go postMessage(msg)
				log.Println("complete sent", msg)
			} else {
				log.Println("complete err", err)
			}
		case "aria2.onDownloadError":
			log.Println("error, ", gid)
			if s, err := a2wsclient.TellStatus(gid); err == nil {
				msg := fmt.Sprintf("*%s*\nStatus:Error (%s)", s.Get("files"), s.Get("errorMessage"))
				go postMessage(msg)
				log.Println("error sent", msg)
			}
		default:
			log.Println("unprocess event:", method)
		}
	}

	log.Panicln("main func exit")
}

func init() {
	// log.SetFlags(0)

	viper.SetConfigName("cmddl")
	viper.AddConfigPath("/root")
	viper.AddConfigPath(".")
	common.Must(viper.ReadInConfig())

	log.Println("using config: ", viper.ConfigFileUsed())

	tgmidurl = viper.GetString("tgmidurl")
	a2url, err := url.Parse(viper.GetString("aria2_url"))
	common.Must(err)

	switch a2url.Scheme {
	case "ws", "http":
		a2url.Scheme = "ws"
	case "wss", "https":
		a2url.Scheme = "wss"
	default:
		log.Fatalln("aria2_url not found")
	}

	a2rpc = a2url.String()
}
