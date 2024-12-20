package main

import (
	"crypto/md5"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"notiferhub/aria2rpc"
)

var (
	rpc     string
	token   string
	testrpc bool
	dlset   string
	uribase string

	aria2Client *aria2rpc.Aria2RPC
)

func testAria2(c *aria2rpc.Aria2RPC) {
	v, err := c.GetVersion()
	fmt.Println(v, err)
	s, err := c.GetGlobalStat()
	fmt.Println(s, err)
}

func postUuidCache(escapedPath, validtime string) string {

	hash := md5.Sum([]byte(escapedPath + validtime + time.Now().GoString()))
	uid := base64.RawURLEncoding.EncodeToString(hash[:])

	hv := url.Values{
		"uuid":      []string{uid},
		"path":      []string{escapedPath},
		"validtime": []string{validtime},
	}

	resp, err := http.Post(dlset,
		"application/x-www-form-urlencoded",
		strings.NewReader(hv.Encode()))

	if err != nil {
		log.Fatal("hashdl post err", err)
	}

	log.Println("hashdl post resp", resp.Status)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		rb, _ := ioutil.ReadAll(resp.Body)
		log.Fatalln("hashdl post resp err", resp.Status, string(rb))
	}
	io.Copy(io.Discard, resp.Body) // nolint

	return strings.TrimSuffix(uribase, "/") + "/?" + uid
}

func addToAria2(contentPath, webdirPath string) {

	webPath := strings.TrimPrefix(contentPath, webdirPath)
	e := []string{}
	for _, p := range strings.Split(webPath, "/") {
		e = append(e, url.PathEscape(p))
	}
	escapedPath := strings.Join(e, "/")
	outPath := webPath
	multiUri := []string{}
	if (dlset != "") {
		for _, v := range []string{"3h", "24h", "72h"} {
			multiUri = append(multiUri, postUuidCache(escapedPath, v))
		}
	} else {
		multiUri = append(multiUri, strings.Join([]string{uribase, webPath}, "/"))
	}

	retries := 20
	for {
		retries--
		log.Printf("Adding URL:%v, (out: %s)", multiUri, outPath)
		ret, err := aria2Client.AddUri(multiUri, outPath)
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
	flag.StringVar(&dlset, "dlset", "", "hash dlset")
	flag.BoolVar(&testrpc, "testrpc", false, "test rpc")
	flag.Parse()

	aria2Client, _ = aria2rpc.NewAria2RPCTLS(token, rpc, true)
	if testrpc {
		testAria2(aria2Client)
		os.Exit(0)
	}

	uribase = strings.TrimSuffix(uribase, "/")

	w, wok := os.LookupEnv("_WEBDIR_PATH")
	c, cok := os.LookupEnv("_CONTENT_PATH")

	log.Printf("_CONTENT_PATH: %s, _WEBDIR_PATH: %s", c, w)

	if wok && cok {
		if fi, err := os.Stat(c); err == nil {
			if !fi.IsDir() {
				addToAria2(c, w)
			} else {

				//recurive walk
				if err := filepath.Walk(c, func(p string, f os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if f.IsDir() {
						return nil
					}

					const sizelim int64 = 1*1024*1024
					if f.Size() < sizelim {
						log.Println("skip smaller then ",sizelim, ", ", f.Name(), "size ", f.Size())
						return nil
					}

					log.Println("walk add,", f.Name())
					addToAria2(p, w)
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
