package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"notiferhub/aria2rpc"
	"notiferhub/common"

	"github.com/spf13/viper"
)

var (
	hostname string
	a2rpc    string
	a2tok    string
	dir      string
	tgmidurl string
	chanurl  string
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

func chanMessage(msg string) error {

	b := strings.SplitN(msg, "\n", 2)
	title := b[0]
	desp := b[1]

	//one-line post request/response...
	resp, err := http.PostForm(chanurl, url.Values{"title": {title}, "desp": {desp}, "tags": {"NASAria2"}})

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
			break
		}

		if err := chanMessage(msg); err != nil {
			log.Println("post failed, retry in 1s", err)
			time.Sleep(time.Second)
			continue
		}
		break
	}
	for {
		retries++
		if retries > 10 {
			log.Println("retries > 10, exit")
			break
		}

		if err := postMessage(msg); err != nil {
			log.Println("post failed, retry in 1s", err)
			time.Sleep(time.Second)
			continue
		}
		break
	}
}

func main() {

	gid = os.Args[1]
	aria2h := aria2rpc.NewAria2RPC(a2tok, a2rpc)
	o, err := aria2h.GetOption(gid)
	common.Must(err)
	if d, ok := o["dir"]; ok {
		dir = d.(string)
	}
	s, err := aria2h.TellStatus(gid)
	common.Must(err)

	msg := ""
	fn := s.Get("files")
	fn = fn[len(dir)+1:]
	switch s.Get("status") {
	case "error":
		msg = fmt.Sprintf("*Aria2@%s Error*\n\n%s\n\n`%s`", hostname, s.Get("errorMessage"), fn)
		log.Println("error", msg)
	case "complete":
		msg = fmt.Sprintf("*Aria2@%s Complete*\n\n`%s\n%s`", hostname, fn, s.String())
		log.Println("complete", msg)
	}

	postUntilSuccess(msg)
}

func init() {
	log.SetFlags(0)

	if h, err := os.Hostname(); (err == nil) {
		hostname = h
	}

	viper.SetConfigName("cmddl")
	viper.AddConfigPath("/etc")
	viper.AddConfigPath(".")
	common.Must(viper.ReadInConfig())
	a2rpc = viper.GetString("aria2_url")
	a2tok = viper.GetString("aria2_token")
	tgmidurl = viper.GetString("tgmidurl")
	chanurl = viper.GetString("chanurl")
	if len(os.Args) == 1 {
		postUntilSuccess("Aria2Hook Test\nmessage at " + time.Now().Format("20060102 15:04:05"))
		os.Exit(0)
	}
}
