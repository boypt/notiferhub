package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/boypt/notiferhub/aria2rpc"
)

var (
	rpc     string
	token   string
	testrpc bool
	isDir   bool
	uribase string

	aria2Client *aria2rpc.Aria2RPC
)

func testAria2(c *aria2rpc.Aria2RPC) {
	v, err := c.GetVersion()
	fmt.Println(v, err)
	s, err := c.GetGlobalStat()
	fmt.Println(s, err)
}

func addToAria2(content, save string) {

	webPath := content[len(save):]
	paths := strings.Split(webPath, "/")
	escaped := []string{}
	for _, p := range paths {
		escaped = append(escaped, url.PathEscape(p))
	}

	dlUrl := fmt.Sprintf("%s%s", uribase, strings.Join(escaped, "/"))
	for {
		log.Printf("Adding (out: %s) URL:%s", webPath, dlUrl)
		ret, err := aria2Client.AddUri([]string{dlUrl}, webPath)
		if err != nil {
			fmt.Println("error occur, wait 3", err)
			time.Sleep(time.Second * 3)
			continue
		}
		fmt.Printf("ret: gid %s\n", ret)
		break
	}
}

func main() {
	flag.StringVar(&rpc, "rpc", "http://localhost:6800", "aria2 rpc")
	flag.StringVar(&token, "token", "", "aria2 token")
	flag.StringVar(&uribase, "baseuri", "", "uri base")
	flag.BoolVar(&testrpc, "testrpc", false, "test rpc")
	flag.BoolVar(&isDir, "dir", false, "is dir")
	flag.Parse()

	aria2Client, _ = aria2rpc.NewAria2RPCTLS(token, rpc, true)
	if testrpc {
		testAria2(aria2Client)
		os.Exit(0)
	}

	c, cok := os.LookupEnv("_CONTENT_PATH")
	s, sok := os.LookupEnv("_SAVE_PATH")

	log.Println("content:", c)
	log.Println("savepath:", s)

	if cok && sok {
		if fi, err := os.Stat(c); err == nil {
			if !fi.IsDir() {
				addToAria2(c, s)
			} else {

				//recurive walk
				if err := filepath.Walk(c, func(p string, f os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if f.IsDir() {
						return nil
					}
					if f.Size() < 5*1024*1024 {
						log.Println("skip small file ", f.Name(), "size ", f.Size())
						return nil
					}

					log.Println("walk add,", f.Name())
					addToAria2(p, s)
					return nil

				}); err != nil {
					log.Fatal(err)
				}
			}
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Println("CLD_PATH not found")
	}
}
