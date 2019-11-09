package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/boyppt/notifer/stock"
	"github.com/boyppt/notifer/tgbot"
	"github.com/joho/godotenv"
)

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
