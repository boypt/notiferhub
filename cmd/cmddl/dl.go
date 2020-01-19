package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
	"unicode"

	"github.com/boypt/notiferhub"
	"github.com/boypt/notiferhub/aria2rpc"
	"github.com/boypt/notiferhub/common"
	"github.com/golang/protobuf/proto"
)

const (
	redisTaskKEY = "cld_task"
	redisGidKey  = "cld_aria_gid"
	redisTmpID   = "cld_tmpid_%d"
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
	gidMap, err := notifierhub.RedisClient.HGetAll(redisGidKey).Result()
	if err != nil {
		log.Panicln("RedisClient.HGetAll", err)
	}
	for gid := range gidMap {
		log.Println("[notifyDL] restore check for", gid)
		go checkGid(gid)
	}

	for {
		// log.Println("waiting for queue")
		procID := fmt.Sprintf(redisTmpID, rand.Intn(9999))
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
		if t.Size < 5*1024*1024 {
			log.Println("task file skiped:", t.SizeStr(), t.Path)
			notifierhub.RedisClient.LPop(listid)
			break
		}
		out := t.Path
		for _, rn := range []rune(t.Path) {
			if unicode.Is(unicode.Han, rn) {
				out = "剧集/" + t.Path
				break
			}
		}

		if gid, err := aria2Client.AddUri(t.DLURL(), out); err != nil {
			log.Println("aria2rpc.AddUri", err)
			if !t.IsSetFailed() {
				t.SetFailed()
				tgAPI("aria2rpc.AddUri Failed:", t.Path, err.Error())
			}
			time.Sleep(time.Second * 3)
			log.Println("task moved back to task list:", t)
			notifierhub.RedisClient.RPopLPush(listid, redisTaskKEY)
			// will retry
		} else {
			log.Println("aria2 created task gid:", gid)
			notifierhub.RedisClient.LPop(listid)
			nowtext, _ := time.Now().MarshalText()
			notifierhub.RedisClient.HSet(redisGidKey, gid, nowtext)
			go checkGid(gid)
		}
	default:
		log.Fatalln("unknow cldType ", t.Type, "leaving redis id:", listid)
	}
}

func checkGid(gid string) {
	startText, err := notifierhub.RedisClient.HGet(redisGidKey, gid).Result()
	if err != nil {
		log.Println(err)
		return
	}
	startTime := time.Time{}
	if err := startTime.UnmarshalText([]byte(startText)); err != nil {
		log.Println(err)
		return
	}

	defer notifierhub.RedisClient.HDel(redisGidKey, gid)

	for {
		s, err := aria2Client.TellStatus(gid)
		if err != nil {
			log.Println("task rpc.TellStatus error, retry in 10s", err)
			time.Sleep(time.Second * 10)
			continue
		}

		sleepDur := time.Second * 30
		switch s.Get("status") {
		case "complete":
			fn := s.Get("files")
			tl := s.Get("totalLength")

			if tlen, err := strconv.ParseInt(tl, 10, 64); err == nil {
				taskDur := time.Since(startTime)
				secs := taskDur.Seconds()
				speed := float64(tlen) / secs
				speedText := notifierhub.HumaneSize(int64(speed))
				log.Println("aria2 completed", gid, fn, speedText)
				go tgAPI(fmt.Sprintf(`Aria2: *%s*
Speed: *%s/s*
Dur: *%s*`, fn, speedText, notifierhub.KitchenDuration(taskDur)))
			} else {
				log.Fatalln("what?? parse err", err)
			}
			return
		case "removed":
			log.Println("aria2 task removed", gid)
			return
		case "waiting":
			sleepDur = time.Minute * 1
		case "active":
			if s.GetProgress() > 90 {
				sleepDur = time.Second * 10
			}
			log.Println("aria2 task", gid, s.String())
		default:
			log.Println("aria2 state default:", gid, s.String())
		}
		time.Sleep(sleepDur)
	}
}
