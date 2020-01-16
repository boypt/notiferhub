package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/boypt/notiferhub/aria2rpc"
	"github.com/boypt/notiferhub/ipocalen"
	"github.com/boypt/notiferhub/stock"
	"github.com/joho/godotenv"
)

var (
	printonly bool
	debug     bool
	nosend    bool
	mode      string
)

func notifyStock() {

	text, err := stock.GetSinaStockText(os.Getenv("STOCKIDS"))
	if err != nil {
		log.Fatal(err)
	}

	if text == "" {
		log.Fatal("text empty")
	}

	notify, err := stock.StockIndexText(text, !printonly, debug)
	if err != nil {
		log.Print(err)
		return
	}

	if printonly {
		fmt.Println(notify)
	}

	if nosend {
		return
	}

	if err := tgAPI(notify); err != nil {
		log.Fatal(err)
	}
}

func notiIPOCalen() {

	s, err := ipocalen.FetchRootSelection()
	if err != nil {
		log.Fatalln(err)
	}
	texts := ipocalen.FindTodayCalendar(s)
	if len(texts) > 1 {
		texts[0] = fmt.Sprintf("*%s*", texts[0])
		notify := strings.Join(texts, "\n")
		if printonly {
			fmt.Println(notify)
		}

		if nosend {
			return
		}

		if err := tgAPI(notify); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {

	flag.BoolVar(&printonly, "print", false, "printonly")
	flag.BoolVar(&debug, "debug", false, "debug")
	flag.BoolVar(&nosend, "nosend", false, "nosend")
	flag.StringVar(&mode, "mode", "dl", "mode: dl/stock")
	flag.StringVar(&mode, "m", "dl", "mode: dl/stock")
	logst := flag.Bool("logts", false, "log time stamp")
	flag.Parse()

	if !*logst {
		log.SetFlags(0)
	}

	homedir, _ := os.UserHomeDir()
	conf := path.Join(homedir, ".ptutils.config")
	err := godotenv.Load(conf)
	if err != nil {
		log.Fatal("Error loading .env file ", conf)
	}

	switch mode {
	case "dl":
		saveTask()
	case "noti":
		aria2Client = aria2rpc.NewAria2RPC(os.Getenv("aria2_token"), os.Getenv("aria2_url"))
		go aria2KeepAlive()
		notifyDL()
	case "stock":
		notifyStock()
	case "ipo":
		notiIPOCalen()
	default:
		log.Fatalln("unknow mode ", mode)
	}
}
