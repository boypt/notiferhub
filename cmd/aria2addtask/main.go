package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/boypt/notiferhub/aria2rpc"
)

var (
	rpc     string
	token   string
	testrpc bool
	uribase string
	dlname  string
)

func main() {
	flag.StringVar(&rpc, "rpc", "http://localhost:6800", "aria2 rpc")
	flag.StringVar(&token, "token", "", "aria2 token")
	flag.StringVar(&uribase, "baseuri", "", "uri base")
	flag.StringVar(&dlname, "dl", "", "dlname")
	flag.BoolVar(&testrpc, "testrpc", false, "test rpc")

	flag.Parse()

	c := aria2rpc.NewAria2RPC(token, rpc)

	if testrpc {
		ver, err := c.GetVersion()
		if err != nil {
			log.Panic(err)
		}
		println(ver)
		s, err := c.GetGlobalStat()
		if err != nil {
			log.Panic(err)
		}
		fmt.Println(s)
		os.Exit(0)
	}

	dlUrl := ""
	if c, ok := os.LookupEnv("CLD_PATH"); ok {
		dlUrl = fmt.Sprintf("%s%s", uribase, url.PathEscape(c))
	}

	if dlUrl == "" {
		os.Exit(1)
	}

	for {
		fmt.Println("Adding URL:", dlUrl)
		ret, err := c.AddUri([]string{dlUrl}, dlname)
		if err != nil {
			time.Sleep(time.Second * 3)
			continue
		}
		fmt.Printf("ret: gid %s\n", ret)
		break
	}
}
