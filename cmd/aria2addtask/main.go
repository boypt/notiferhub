package main

import (
	"flag"
	"fmt"
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
)

func testAria2(c *aria2rpc.Aria2RPC) {
	v, err := c.GetVersion()
	fmt.Println(v, err)
	s, err := c.GetGlobalStat()
	fmt.Println(s, err)
}

func main() {
	flag.StringVar(&rpc, "rpc", "http://localhost:6800", "aria2 rpc")
	flag.StringVar(&token, "token", "", "aria2 token")
	flag.StringVar(&uribase, "baseuri", "", "uri base")
	flag.BoolVar(&testrpc, "testrpc", false, "test rpc")
	flag.Parse()

	aria2Client, _ := aria2rpc.NewAria2RPCTLS(token, rpc, true)
	if testrpc {
		testAria2(aria2Client)
		os.Exit(0)
	}

	if c, ok := os.LookupEnv("CLD_PATH"); ok {
		dlUrl := fmt.Sprintf("%s%s", uribase, url.PathEscape(c))
		for {
			fmt.Println("Adding URL:", dlUrl)
			ret, err := aria2Client.AddUri([]string{dlUrl}, c)
			if err != nil {
				fmt.Println("error occur, wait 3", err)
				time.Sleep(time.Second * 3)
				continue
			}
			fmt.Printf("ret: gid %s\n", ret)
			break
		}
	} else {
		fmt.Println("CLD_PATH not found")
	}
}
