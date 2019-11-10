package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/boyppt/notiferhub/stock"
	"github.com/boyppt/notiferhub/tgbot"
	"github.com/joho/godotenv"
)

func main() {

	var debug bool
	flag.BoolVar(&debug, "debug", false, "debug")
	flag.Parse()

	homedir, _ := os.UserHomeDir()
	conf := path.Join(homedir, ".ptutils.config")

	err := godotenv.Load(conf)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	text, err := stock.GetSinaStockText(os.Getenv("STOCKIDS"))
	if err != nil {
		log.Fatal(err)
	}

	if text == "" {
		log.Fatal("text empty")
	}

	notify, err := stock.StockIndexText(text, !debug)
	if err != nil {
		if errors.Is(err, stock.ErrMarketClosed) {
			log.Println(err)
			return
		}
		log.Fatal(err)
	}

	if debug {
		fmt.Println(notify)
		return
	}

	if err := tgbot.JustNotify(notify); err != nil {
		log.Fatal(err)
	}
}
