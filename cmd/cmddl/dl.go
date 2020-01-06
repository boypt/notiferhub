package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/boypt/notiferhub/aria2rpc"
)

func byteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func dlText(path, size string) string {
	sizecnt, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		sizecnt = 0
	}

	return fmt.Sprintf(`*%s*
Size: *%s*
Time: *%s*`, path, byteCountSI(sizecnt), time.Now().Format(time.Stamp))
}

func dlURL(path string) string {
	base := os.Getenv("sourceroot")
	return fmt.Sprintf("%s/%s", base, url.PathEscape(path))
}

func notifyDL() {
	cldPath := os.Getenv("CLD_PATH")
	cldType := os.Getenv("CLD_TYPE")
	cldSize := os.Getenv("CLD_SIZE")
	cldRest := os.Getenv("CLD_RESTAPI")
	cldHash := os.Getenv("CLD_HASH")

	switch cldType {
	case "torrent":
		text := dlText(cldPath, cldSize)
		tryMax(3, tgAPI, text)
		tryMax(1, chanAPI, "torrent", text)
		time.Sleep(time.Second * 3)
		cldAPI(cldRest, cldHash)
	case "file":
		sizecnt, err := strconv.ParseInt(cldSize, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		// 5MB limit
		if sizecnt < 5*1024*1024 {
			log.Println("file too small ", cldPath)
			break
		}
		if terr := tryMax(10, aria2rpc.JustAddURL, dlURL(cldPath)); terr != nil {
			f, err := os.OpenFile("/tmp/aria2_failing_uris.txt",
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			if _, err := f.WriteString(terr.Error() + "\n"); err != nil {
				log.Println(err)
			}
			tryMax(3, tgAPI, "Fail to call download file: "+cldPath)
		}
	default:
		log.Fatalln("unknow cldType ", cldType)
	}
}
