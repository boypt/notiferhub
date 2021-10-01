package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/boypt/notiferhub/aria2rpc"
)

var (
	rpc     string
	token   string
	testrpc bool
	uri     string
	dlname  string
)

func main() {
	flag.StringVar(&rpc, "rpc", "http://localhost:6800", "aria2 rpc")
	flag.StringVar(&token, "token", "", "aria2 token")
	flag.StringVar(&uri, "uri", "", "uri to download")
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

	ret, err := c.AddUri([]string{uri}, dlname)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("ret: %v\n", ret)
}
