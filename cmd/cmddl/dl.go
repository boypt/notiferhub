package main

import (
	"log"
	"os"
	"time"

	"github.com/boypt/notiferhub"
	"github.com/boypt/notiferhub/aria2rpc"
)

func notifyDL() {

	t, _ := notifierhub.NewDLfromCLD()

	switch t.Type {
	case "torrent":
		text := t.DLText()
		tryMax(3, tgAPI, text)
		tryMax(1, chanAPI, "torrent", text)
		time.Sleep(time.Second * 3)
		cldAPI(t.REST, t.Hash)
	case "file":
		// 5MB limit
		if size := t.SizeInt(); size < 0 || size < 5*1024*1024 {
			log.Println("file too small ", t.Path)
			break
		}
		if terr := tryMax(10, aria2rpc.JustAddURL, t.DLURL(), t.Path); terr != nil {
			f, err := os.OpenFile("/tmp/aria2_failing_uris.txt",
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			if _, err := f.WriteString(terr.Error() + "\n"); err != nil {
				log.Println(err)
			}
			tryMax(3, tgAPI, "Fail to call download file: "+t.Path)
		}
	default:
		log.Fatalln("unknow cldType ", t.Type)
	}
}
