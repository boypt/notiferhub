package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/boypt/notiferhub/aria2rpc"
	"github.com/boypt/notiferhub/ipocalen"
	"github.com/boypt/notiferhub/rss"
	"github.com/boypt/notiferhub/stock"
	"github.com/spf13/viper"
)

var (
	printonly bool
	debug     bool
	nosend    bool
	mode      string
	webPort   int
)

func notifyStock() {

	text, err := stock.GetSinaStockText(viper.GetString("stockids"))
	if err != nil {
		log.Println("notifyStock", err)
		return
	}

	if text == "" {
		log.Println("notifyStock", "text empty")
	}

	notify, err := stock.StockIndexText(text, !printonly, debug)
	if err != nil {
		log.Println(err)
		return
	}

	if printonly {
		fmt.Println(notify)
	}

	if nosend {
		return
	}

	if err := tgAPI(notify); err != nil {
		log.Println(err)
		return
	}
}

func notiIPOCalen() {

	s, err := ipocalen.FetchRootSelection()
	if err != nil {
		log.Println(err)
		return
	}
	texts := ipocalen.FindTodayCalendar(s)
	if len(texts) > 0 {

		notify := strings.Join(texts, "\n")
		if printonly {
			fmt.Println(notify)
		}

		if nosend {
			return
		}

		if err := tgAPI(notify); err != nil {
			log.Println(err)
			return
		}
	} else {
		log.Println("IPO text unclear:", texts)
	}
}

func startTaskWeb() {
	http.HandleFunc("/cld_save", saveTask)
	fmt.Println("Starting cld_save server at port", webPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", webPort), nil); err != nil {
		log.Fatal(err)
	}
}

func main() {

	flag.BoolVar(&printonly, "print", false, "printonly")
	flag.BoolVar(&debug, "debug", false, "debug")
	flag.BoolVar(&nosend, "nosend", false, "nosend")
	flag.StringVar(&mode, "mode", "dl", "mode: dl/stock")
	flag.StringVar(&mode, "m", "dl", "mode: dl/stock/ipo/noti")
	flag.IntVar(&webPort, "webport", 7267, "web server port")
	logst := flag.Bool("logts", false, "log time stamp")
	flag.Parse()

	if !*logst {
		log.SetFlags(0)
	}

	viper.SetConfigName("cmddl")
	viper.AddConfigPath("/srv")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("Error config file", err)
	}

	switch mode {
	case "noti":
		log.Println("using config:", viper.ConfigFileUsed())
		aria2Client = aria2rpc.NewAria2RPC(
			viper.GetString("aria2_token"),
			viper.GetString("aria2_url"),
		)
		go restoreFromRedis()
		go aria2KeepAlive()
		go startTaskWeb()
		setCronTask()
		notifyLoop()
	case "autorss":
		rss.FindFromRSS()
	case "stock":
		notifyStock()
	case "ipo":
		notiIPOCalen()
	default:
		log.Fatalln("unknow mode ", mode)
	}
}
