package main

import (
	"flag"
	"io"
	"log"
	"log/syslog"
	"net/http"
	"net/url"
	"os"
	"time"

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
		log.Println("postform err", err)
	}

	defer response.Body.Close()
	io.Copy(io.Discard, response.Body)
}

func main() {
	log.Println("starting ...")
	for {
		log.Println("connecting to aria2 :", a2rpc)
		if c, err := NewAria2Conn(a2rpc, viper.GetString("aria2_token")); err == nil {
			c.InitInfo()
			c.EventLoop()
		} else {
			log.Println("connect err:", err)
		}
		time.Sleep(time.Second * 5)
		log.Println("restarting ...")
	}
}

func init() {
	test := flag.Bool("test", false, "fire test")
	lsyslog := flag.Bool("syslog", false, "log to syslog")
	flag.Parse()

	if *lsyslog {
		logwriter, e := syslog.New(syslog.LOG_NOTICE, "aria2event")
		if e == nil {
			log.SetOutput(logwriter)
			log.SetFlags(0)
		} else {
			log.Println("syslog err", e)
		}
	}

	viper.SetConfigName("cmddl")
	viper.AddConfigPath("/root")
	viper.AddConfigPath(".")
	common.Must(viper.ReadInConfig())

	log.Println("using config: ", viper.ConfigFileUsed())

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

	tgmidurl = viper.GetString("tgmidurl")
	log.Println("tgmidurl:", tgmidurl)

	if *test {
		postMessage("test message")
		os.Exit(0)
	}
}