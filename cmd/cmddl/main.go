package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/boypt/notiferhub/aria2rpc"
	"github.com/boypt/notiferhub/common"
	"github.com/boypt/notiferhub/ipocalen"
	"github.com/boypt/notiferhub/stock"
	"github.com/spf13/viper"
)

var (
	printonly bool
	debug     bool
	nosend    bool
	mode      string
)

func notifyStock() {

	text, err := stock.GetSinaStockText(viper.GetString("stockids"))
	common.Must(err)

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

	common.Must(tgAPI(notify))
}

func notiIPOCalen() {

	s, err := ipocalen.FetchRootSelection()
	common.Must(err)
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

		common.Must(tgAPI(notify))
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

	viper.SetConfigName("cmddl")
	viper.AddConfigPath("/srv")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln("Error config file", err)
	}

	switch mode {
	case "dl":
		saveTask()
	case "noti":
		log.Println("using config:", viper.ConfigFileUsed())
		aria2Client = aria2rpc.NewAria2RPC(
			viper.GetString("aria2_token"),
			viper.GetString("aria2_url"),
		)
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
