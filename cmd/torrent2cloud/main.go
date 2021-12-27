package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/joho/godotenv"
)

var (
	del = flag.Bool("del", false, "delete trackers")
)

func postTorrent(buff *bytes.Buffer) error {
	apiHost := os.Getenv("CLDTORRENT")
	magapi := apiHost + "/api/torrentfile"
	req, err := http.NewRequest("POST", magapi, buff)
	if err != nil {
		return err
	}

	req.Header.Set("content-type", "application/json;charset=utf-8")
	req.Header.Set("referer", apiHost)
	req.Header.Set("Cookie", strings.TrimPrefix(os.Getenv("CLDCOOKIE"), "cookie: "))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if b := string(body); !strings.HasPrefix(b, "OK") {
		return fmt.Errorf("return not ok:%s", b)
	}

	return nil
}

func torrent2cloud(fn string) error {

	buff := &bytes.Buffer{}
	mi, err := metainfo.LoadFromFile(fn)
	if err != nil {
		return err
	}

	if ifo, err := mi.UnmarshalInfo(); err == nil {
		fmt.Println("--> [", ifo.Name, "]", filepath.Base(fn))
	}

	if *del {
		if strings.Contains(mi.Announce, "plab.site") {
			log.Println("remove Annouce: ", mi.Announce)
			mi.Announce = ""
			mi.AnnounceList = metainfo.AnnounceList{}
		}

		if err := mi.Write(buff); err != nil {
			return err
		}
	} else {
		fbyte, err := ioutil.ReadFile(fn)
		if err != nil {
			return err
		}
		buff = bytes.NewBuffer(fbyte)
	}

	return postTorrent(buff)
}

func main() {

	flag.Parse()
	cur, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	_ = godotenv.Load(filepath.Join(cur.HomeDir, ".ptutils.config"))
	_ = godotenv.Load() // for .env

	tors := []string{}
	for _, path := range strings.Split(os.Getenv("CLDTORRENTDIR"), " ") {
		if t, err := filepath.Glob(fmt.Sprintf("%s/*.torrent", path)); err == nil {
			log.Printf("Path %s found %d\n", path, len(t))
			tors = append(tors, t...)
		}
	}

	for _, torf := range tors {
		for {
			err := torrent2cloud(torf)
			if err == nil {
				os.Remove(torf)
				fmt.Println("===================================")
				break
			}
			fmt.Println("err", err)
		}
	}

	if runtime.GOOS == "windows" {
		fmt.Println("===================================")
		fmt.Println("\nPress 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}
