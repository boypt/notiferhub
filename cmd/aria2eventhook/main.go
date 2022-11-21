package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"notiferhub/aria2rpc"
	"notiferhub/common"

	"github.com/spf13/viper"
)

var (
	a2rpc    string
	a2tok    string
	tgmidurl string
	gid      string
)

func postMessage(msg string) error {

	//one-line post request/response...
	resp, err := http.PostForm(tgmidurl, url.Values{"message": {msg}})

	//okay, moving on...
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	t, err := io.ReadAll(resp.Body)
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
			log.Println("post failed, retry in 1s", err)
			time.Sleep(time.Second)
			continue
		}

		return
	}
}

func main() {

	gid = os.Args[1]
	aria2h := aria2rpc.NewAria2RPC(a2tok, a2rpc)
	s, err := aria2h.TellStatus(gid)
	if err != nil {
		log.Panicln(err)
	}

	msg := ""
	fn := s.Get("files")
	switch s.Get("status") {
	case "error":
		msg = fmt.Sprintf("*Error:\n\n%s\n\n%s*", s.Get("errorMessage"), fn)
		log.Println("error", msg)
	case "complete":
		msg = fmt.Sprintf("*Complete*\n\n*%s*\n", fn)
		log.Println("complete", msg)
	}

	postUntilSuccess(msg)
}

func init() {
	test := flag.Bool("test", false, "fire test")

	viper.SetConfigName("cmddl")
	viper.AddConfigPath("/root")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath(".")
	common.Must(viper.ReadInConfig())
	log.Println("using config: ", viper.ConfigFileUsed())

	a2url, err := url.Parse(viper.GetString("aria2_url"))
	common.Must(err)
	a2rpc = a2url.String()
	a2tok = viper.GetString("aria2_token")
	tgmidurl = viper.GetString("tgmidurl")
	if *test || len(os.Args) == 1 {
		postMessage("test message")
		os.Exit(0)
	}
}
