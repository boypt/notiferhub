package main

import (
	"flag"
	"io/ioutil"
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
	testTell bool
	testch   chan struct{}
	a2rpc    string
	tgmidurl string
)

func postMessage(msg string) error {
	if testTell {
		defer close(testch)
	}

	//one-line post request/response...
	resp, err := http.PostForm(tgmidurl, url.Values{"message": {msg}})

	//okay, moving on...
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	t, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println("resp:", resp.Status, string(t))
	return nil
}

func postUntilSuccess(msg string) {
	retries := 0
	for {
		retries++
		if retries > 10 {
			log.Println("retries > 10, exit")
			return
		}

		if err := postMessage(msg); err != nil {
			log.Println("post failed, retry in 1s")
			time.Sleep(time.Second)
			continue
		}

		return
	}
}

func main() {
	log.Println("starting ...")

	c := NewAria2Conn(a2rpc, viper.GetString("aria2_token"))

	if testTell {
		if flag.Arg(0) != "" {
			if err := c.InitConn(); err == nil {
				c.InitInfo()
				testch = make(chan struct{})
				if err := c.OnDownloadComplete(flag.Arg(0)); err == nil {
					<-testch
				}
			}
		}
		return
	}

	for {
		log.Println("connecting to aria2 :", a2rpc)
		if err := c.InitConn(); err == nil {
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
	flag.BoolVar(&testTell, "testtell", false, "test tell")
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
