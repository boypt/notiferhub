package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/boypt/notiferhub"
	"github.com/boypt/notiferhub/aria2rpc"
	"github.com/golang/protobuf/proto"
)

const (
	redisTaskKEY = "cld_task"
)

func saveTask() {
	t, _ := notifierhub.NewTorrentfromCLD()
	data, err := proto.Marshal(t)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	if _, err := notifierhub.RedisClient.LPush(redisTaskKEY, string(data)).Result(); err != nil {
		log.Fatal("redis error: ", err)
	}
}

func notifyDL() {
	for {
		log.Println("waiting for queue")
		procID := fmt.Sprintf("cld_id-%d", rand.Intn(9999))
		r, err := notifierhub.RedisClient.BRPopLPush(redisTaskKEY, procID, 0).Result()
		if err != nil {
			log.Fatal(err)
			return
		}

		t := &notifierhub.TorrentTask{}
		if err := proto.Unmarshal([]byte(r), t); err != nil {
			log.Fatal(err)
			return
		}

		go processTask(t, procID)
	}
}

func processTask(t *notifierhub.TorrentTask, listid string) {
	defer notifierhub.RedisClient.LPop(listid)
	for {
		switch t.Type {
		case "torrent":
			text := t.DLText()
			if err := tgAPI(text); err != nil {
				log.Println("tgAPI", err)
				// retry
				break
			}

			go chanAPI("torrent", text) // no retry
			time.Sleep(time.Second * 10)

			if err := cldAPI(t.Rest, t.Hash); err != nil {
				log.Println("cldAPI", err)
			}

			return
		case "file":
			// 5MB limit
			if size := t.SizeInt(); size < 0 || size < 5*1024*1024 {
				log.Println("file too small: ", t.SizeStr(), t.Path)
				return
			}

			if resp, err := aria2rpc.JustAddURL(t.DLURL(), t.Path); err != nil {
				log.Println("aria2rpc", err)
				tgAPI("Fail to call download file: ", t.Path, err.Error())
				time.Sleep(time.Second * 10)
				// will retry
				break
			} else {
				log.Println("aria2Resp", resp)
			}
			return
		default:
			log.Fatalln("unknow cldType ", t.Type)
			return
		}
	}
}
