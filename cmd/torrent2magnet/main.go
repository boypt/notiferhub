package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/joho/godotenv"
)

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
		fmt.Println("===================================")
		if torr, err := metainfo.LoadFromFile(torf); err == nil {
			i, _ := torr.UnmarshalInfo()
			fmt.Println(i.Name)
			torr.Announce = ""
			torr.AnnounceList = [][]string{}
			m := torr.Magnet(nil, nil)
			fmt.Println(m.String())
		}
	}

	if runtime.GOOS == "windows" {
		fmt.Println("===================================")
		fmt.Println("\nPress 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}
