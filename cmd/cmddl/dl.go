package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
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
			log.Fatal(fmt.Errorf("%w '%s'", err, r))
		}

		go processTask(t, procID)
	}
}

func processTask(t *notifierhub.TorrentTask, listid string) {

	switch t.Type {
	case "torrent":
		text := t.DLText()
		if err := tgAPI(text); err != nil {
			// retry
			log.Println("tgAPI failed, task moved back to task list:", t, err)
			notifierhub.RedisClient.RPopLPush(listid, redisTaskKEY)
			break
		}

		// no retry
		go chanAPI("torrent", text)
		time.Sleep(time.Second * 10)
		go cldAPI(t.Rest, t.Hash)
		return
	case "file":
		// 5MB limit
		if size := t.SizeInt(); size < 0 || size < 5*1024*1024 {
			log.Println("file too small, task skiped:", t.SizeStr(), t.Path)
			notifierhub.RedisClient.LPop(listid)
			break
		}

		rpc := aria2rpc.NewAria2RPC(os.Getenv("aria2_token"), os.Getenv("aria2_url"))
		if resp, err := rpc.AddUri(t.DLURL(), t.Path); err != nil {
			log.Println("aria2rpc.JustAddURL", err)
			if !t.IsSetFailed() {
				t.SetFailed()
				tgAPI("aria2rpc.JustAddURL Failed:", t.Path, err.Error())
			}
			time.Sleep(time.Second * 10)
			log.Println("task moved back to task list:", t)
			notifierhub.RedisClient.RPopLPush(listid, redisTaskKEY)
			// will retry
		} else {
			log.Println("aria2 created task id:", resp.Result.(string))
			notifierhub.RedisClient.LPop(listid)
			go func(gid string) {
				for {
					stat, err := rpc.TellStatus(gid)
					if err != nil {
						log.Println("task rpc.TellStatus error", err)
						return
					}
					if stat.GetStatus() != "complete" {
						log.Println("aria2 task", gid, stat.String())
						time.Sleep(time.Second * 30)
						continue
					}
					log.Println("aria2 task", gid, "completed")
					tgAPI("*Aria2 Download*\n", t.Path)
					return
				}
			}(resp.Result.(string))
		}
	default:
		log.Fatalln("unknow cldType ", t.Type, "leaving redis id:", listid)
	}
}
