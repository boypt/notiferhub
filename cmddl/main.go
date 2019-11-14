package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/boypt/notiferhub/aria2rpc"
	"github.com/boypt/notiferhub/stock"
	"github.com/boypt/notiferhub/tgbot"
	"github.com/joho/godotenv"
)

var (
	debug bool
	mode  string
)

func notifyText(path, size string) string {
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

func tryMax(max int, fun func(string) error, arg string) error {

	var terr error

	for {
		max--
		if max < 0 {
			break
		}

		err := fun(arg)
		if err == nil {
			return nil
		}

		terr = fmt.Errorf("run error > %w", err)
		log.Printf("tryMax got err %v", err)
	}

	return terr
}

func notifyStock() {

	text, err := stock.GetSinaStockText(os.Getenv("STOCKIDS"))
	if err != nil {
		log.Fatal(err)
	}

	if text == "" {
		log.Fatal("text empty")
	}

	notify, err := stock.StockIndexText(text, !debug)
	if err != nil {
		log.Print(err)
		return
	}

	if debug {
		fmt.Println(notify)
	}

	if err := tgbot.JustNotify(notify); err != nil {
		log.Fatal(err)
	}
}

func chanAPI(text string) error {
	purl := os.Getenv("chanapi")
	token := os.Getenv("chantoken")
	if purl == "" {
		return nil
	}

	req, err := http.NewRequest("POST", purl, bytes.NewBuffer([]byte(text)))
	req.Header.Set("mkt-token", token)
	c := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func notifyAria() {
	cldPath := os.Getenv("CLD_PATH")
	cldType := os.Getenv("CLD_TYPE")
	cldSize := os.Getenv("CLD_SIZE")

	switch cldType {
	case "torrent":
		text := notifyText(cldPath, cldSize)
		tryMax(3, tgbot.JustNotify, text)
		tryMax(1, chanAPI, text)
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
		}
	default:
		log.Fatalln("unknow cldType ", cldType)
	}
}

func main() {

	flag.BoolVar(&debug, "debug", false, "debug")
	flag.StringVar(&mode, "mode", "dl", "mode: dl/stock")
	flag.StringVar(&mode, "m", "dl", "mode: dl/stock")
	flag.Parse()

	homedir, _ := os.UserHomeDir()
	conf := path.Join(homedir, ".ptutils.config")
	err := godotenv.Load(conf)
	if err != nil {
		log.Fatal("Error loading .env file ", conf)
	}

	switch mode {
	case "dl":
		notifyAria()
	case "stock":
		notifyStock()
	default:
		log.Fatalln("unknow mode ", mode)
	}
}
