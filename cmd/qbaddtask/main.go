package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/boypt/notiferhub/aria2rpc"
)

var (
	rpc     string
	token   string
	testrpc bool
	uribase string

	aria2Client *aria2rpc.Aria2RPC
)

func testAria2(c *aria2rpc.Aria2RPC) {
	v, err := c.GetVersion()
	fmt.Println(v, err)
	s, err := c.GetGlobalStat()
	fmt.Println(s, err)
}

func LocalPath2WebPath(localfile, localbase, webbase string) (string, string) {
	webPath := localfile[len(localbase):]
	paths := strings.Split(webPath, "/")
	escaped := []string{}
	for _, p := range paths {
		escaped = append(escaped, url.PathEscape(p))
	}

	return webPath, strings.TrimRight(webbase, "/") + "/" + strings.Join(escaped, "/")
}

func addToAria2(contentPath, savePath, catelog string) {

	webPath, dlUrl := LocalPath2WebPath(contentPath, savePath, uribase)
	outPath := path.Join(catelog, webPath)

	retries := 20
	for {
		retries--
		log.Printf("Adding  URL:%s, (out: %s)", dlUrl, outPath)
		ret, err := aria2Client.AddUri([]string{dlUrl}, outPath)
		if err != nil {
			fmt.Println("error occur, wait 3", err)
			if retries == 0 {
				fmt.Println("error occur 20 times, next")
			}
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
	flag.Parse()

	aria2Client, _ = aria2rpc.NewAria2RPCTLS(token, rpc, true)
	if testrpc {
		testAria2(aria2Client)
		os.Exit(0)
	}

	c, cok := os.LookupEnv("_CONTENT_PATH")
	s, sok := os.LookupEnv("_SAVE_PATH")
	l, _ := os.LookupEnv("_CATALOG")


	log.Printf("_CONTENT_PATH: %s, _SAVE_PATH: %s, _CATALOG: %s", c, s, l)

	if cok && sok {
		if fi, err := os.Stat(c); err == nil {
			if !fi.IsDir() {
				addToAria2(c, s, l)
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
					addToAria2(p, s, l)
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
