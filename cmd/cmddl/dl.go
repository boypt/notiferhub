package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/boypt/notiferhub"
	"github.com/boypt/notiferhub/aria2rpc"
	"github.com/boypt/notiferhub/common"
	"github.com/golang/protobuf/proto"
)

const (
	redisTaskKEY = "cld_task"
	redisGidKey  = "cld_aria_gid"
)

var (
	aria2Client *aria2rpc.Aria2RPC
)

func saveTask() {
	t, _ := notifierhub.NewTorrentfromCLD()
	data, err := proto.Marshal(t)
	common.Must(err)

	_, err = notifierhub.RedisClient.LPush(redisTaskKEY, string(data)).Result()
	common.Must(err)
}

func aria2KeepAlive() {
	if ver, err := aria2Client.GetVersion(); err == nil {
		log.Println("Aria2 Connected", ver)
	} else {
		log.Println("Aria2 connect err", err)
	}

	var laststat string
	for {
		if stat, err := aria2Client.GetGlobalStat(); err == nil {
			curstat := fmt.Sprintf("%v", stat)
			if laststat != curstat {
				log.Println("Aria2 Stat", stat)
				laststat = curstat
			}
		} else {
			log.Println("Aria2 stat err", err)
		}

		time.Sleep(time.Minute)
	}
}

func notifyDL() {
	keys, err := notifierhub.RedisClient.SMembers(redisGidKey).Result()
	if err != nil {
		log.Panicln("RedisClient.SScan", err)
	}
	for _, gid := range keys {
		log.Println("[notifyDL] restore check for", gid)
		go checkGid(gid)
	}

	for {
		log.Println("waiting for queue")
		procID := fmt.Sprintf("cld_id-%d", rand.Intn(9999))
		r, err := notifierhub.RedisClient.BRPopLPush(redisTaskKEY, procID, 0).Result()
		common.Must(err)

		t := &notifierhub.TorrentTask{}
		common.Must(proto.Unmarshal([]byte(r), t))
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

		if resp, err := aria2Client.AddUri(t.DLURL(), t.Path); err != nil {
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
			gid := resp.Result.(string)
			log.Println("aria2 created task gid:", gid)
			notifierhub.RedisClient.LPop(listid)
			notifierhub.RedisClient.SAdd(redisGidKey, gid)
			go checkGid(gid)
		}
	default:
		log.Fatalln("unknow cldType ", t.Type, "leaving redis id:", listid)
	}
}

func checkGid(gid string) {
	defer notifierhub.RedisClient.SRem(redisGidKey, gid)
	for {
		stat, err := aria2Client.TellStatus(gid)
		if err != nil {
			log.Println("task rpc.TellStatus error, retry in 10s", err)
			time.Sleep(time.Second * 10)
			continue
		}

		sleepDur := time.Second * 30
		switch stat.Get("status") {
		case "complete":
			fn := stat.Get("files")
			log.Println("aria2 completed", gid, fn)
			go tgAPI(fmt.Sprintf("*%s*\n completed in aria2", fn))
			return
		case "removed":
			log.Println("aria2 task removed", gid)
			return
		case "waiting":
			sleepDur = time.Minute * 5
		case "active":
			if stat.GetProgress() > 90 {
				sleepDur = time.Second * 10
			}
			log.Println("aria2 task", gid, stat.String())
		default:
			log.Println("aria2 state default:", gid, stat.String())
		}
		time.Sleep(sleepDur)
	}
}
