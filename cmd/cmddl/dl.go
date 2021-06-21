package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"unicode"

	notifierhub "github.com/boypt/notiferhub"
	"github.com/boypt/notiferhub/aria2rpc"
	"github.com/boypt/notiferhub/common"
	"github.com/boypt/notiferhub/rss"
	"github.com/golang/protobuf/proto"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

const (
	redisTaskKEY   = "cld_task"
	redisGidKey    = "cld_aria_gid"
	redisTmpID     = "cld_tmpid_%d"
	redisDelayTask = "cld_delay_tasks"
)

var (
	aria2Client  *aria2rpc.Aria2RPC
	errFailAria2 = errors.New("fail to add aria2 task")
)

func saveTask(w http.ResponseWriter, r *http.Request) {
	defer fmt.Fprintf(w, "GOT")

	t, err := notifierhub.NewTorrentfromReq(r)
	if err != nil {
		log.Println("no notification:", err)
		fmt.Fprintf(w, "%#v", err)
		return
	}

	if r.PostFormValue("ONLYNOTIFY") != "" {
		if err := tgAPI(t.DLText()); err != nil {
			log.Println(err)
		}
		return
	}

	if data, err := proto.Marshal(t); err != nil {
		w.WriteHeader(500)
		log.Println(err)
	} else {
		common.Must2(notifierhub.RedisClient.LPush(redisTaskKEY, string(data)).Result())
	}
	log.Println("Added Task", t.Hash, t.Path)
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
			curSpeed := fmt.Sprintf("Global DL: %s/s", common.HumaneSize(stat["downloadSpeed"]))
			if laststat != curSpeed {
				log.Println("Aria2", curSpeed)
				laststat = curSpeed
			}
		} else {
			log.Println("Aria2 stat err", err)
		}

		time.Sleep(time.Minute)
	}
}

func restoreFromRedis() {

	gids, err := notifierhub.RedisClient.HKeys(redisGidKey).Result()
	common.Must(err)
	for _, gid := range gids {
		log.Println("[restore] GidCheck for", gid)
		go checkGid(gid)
	}

	hashs, err := notifierhub.RedisClient.HKeys(redisDelayTask).Result()
	common.Must(err)
	for _, hash := range hashs {
		log.Println("[restore] Delay task", hash)
		go delayCleanTask(hash)
	}
}

func notifyLoop() {

	for {
		// log.Println("waiting for queue")
		tmpID := fmt.Sprintf(redisTmpID, rand.Intn(9999))
		r, err := notifierhub.RedisClient.BRPopLPush(redisTaskKEY, tmpID, 0).Result()
		common.Must(err)

		t := &notifierhub.TorrentTask{}
		common.Must(proto.Unmarshal([]byte(r), t))
		go processTask(t, tmpID, r)
	}
}

func aria2AddTask(t *notifierhub.TorrentTask) error {

	out := t.Path
	for _, rn := range []rune(t.Path) {
		if unicode.Is(unicode.Han, rn) {
			out = "剧集/" + out
			break
		}
	}

	if gid, err := aria2Client.AddUri(t.DLURL(), out); err == nil {
		log.Println("aria2 created task gid:", gid)
		if t.Aria2TS == 0 {
			t.Aria2TS = time.Now().Unix()
		}
		tx, _ := proto.Marshal(t)
		notifierhub.RedisClient.HSet(redisGidKey, gid, string(tx))
		go checkGid(gid)
	} else {
		log.Println("aria2rpc.AddUri", err)
		if !t.IsSetFailed() {
			t.SetFailed()
			tgAPI("aria2rpc.AddUri Failed:", t.Path, err.Error())
		}
		time.Sleep(time.Second * 3)
		return errFailAria2
	}

	return nil
}

func processTask(t *notifierhub.TorrentTask, tmpID string, origTask string) {

	defer notifierhub.RedisClient.LPop(tmpID)
	switch t.Type {
	case "torrent":
		text := t.DLText()
		if err := tgAPI(text); err != nil {
			// retry
			log.Println("tgAPI failed, task moved back to task list:", t, err)
			notifierhub.RedisClient.RPopLPush(tmpID, redisTaskKEY)
			break
		}

		// no retry
		go chanAPI("torrent", text)
		notifierhub.RedisClient.HSet(redisDelayTask, t.Hash, origTask)
		go delayCleanTask(t.Hash)
	case "file":
		// 5MB limit
		if t.Size < 5*1024*1024 {
			log.Println("task file skiped:", common.HumaneSize(t.Size), t.Path)
			break
		}

		if err := aria2AddTask(t); err != nil {
			if errors.Is(err, errFailAria2) {
				log.Println("task moved back to task list:", t)
				notifierhub.RedisClient.RPopLPush(tmpID, redisTaskKEY)
			}
		}
	default:
		log.Fatalln("unknow cldType ", t.Type, "leaving redis id:", tmpID)
	}
}

func checkGid(gid string) {
	taskText, err := notifierhub.RedisClient.HGet(redisGidKey, gid).Result()
	if err != nil {
		log.Println(err)
		return
	}
	task := &notifierhub.TorrentTask{}
	common.Must(proto.Unmarshal([]byte(taskText), task))

	defer notifierhub.RedisClient.HDel(redisGidKey, gid)

	for {
		s, err := aria2Client.TellStatus(gid)
		if err != nil {
			log.Printf("task rpc.TellStatus error, %v, %s", err, err.Error())
			if merr, ok := err.(aria2rpc.Aria2Err); ok {
				log.Println("task removed:", merr.Code, merr.Message)
				return
			}
			time.Sleep(time.Second * 10)
			continue
		}

		sleepDur := time.Second * 30
		switch s.Get("status") {
		case "complete":
			fn := s.Get("files")
			tl := s.Get("totalLength")

			if tlen, err := strconv.ParseInt(tl, 10, 64); err == nil {
				taskDur := time.Since(time.Unix(task.Aria2TS, 0))
				secs := taskDur.Seconds()
				speed := float64(tlen) / secs
				speedText := common.HumaneSize(int64(speed))
				log.Println("aria2 completed", gid, fn, speedText)
				go tgAPI(fmt.Sprintf("Aria2: *%s*\nDur: *%s*\nAvg: *%s/s*",
					fn, common.KitchenDuration(taskDur), speedText))
			} else {
				log.Fatalln("what?? parse err", err)
			}
			return
		case "removed":
			log.Println("aria2 task removed", gid)
			return
		case "paused":
			log.Println("aria2 task paused, unpause", gid)
			aria2Client.UnPause(gid)
			sleepDur = time.Second * 30
		case "waiting":
			sleepDur = time.Minute
		case "active":
			if speed := s.Speed(); speed > 0 && speed < 102400 && s.GetProgress() < 90.0 {
				// slower than 100k/s, pause it
				aria2Client.Pause(gid)
				log.Println("aria2 task", gid, s.String(), "too slow, paused it")
				sleepDur = time.Second * 10
				break
			}
			if s.GetProgress() > 90 {
				sleepDur = time.Second * 10
			}
			log.Println("aria2 task", gid, s.String())
		case "error":
			task.ErrorCounter++
			if task.ErrorCounter < 5 {
				log.Println("aria2 task error, recreate aria2 task", gid, s.String())
				if err := aria2AddTask(task); err != nil {
					log.Println("aria2 recreate failed", err)
				}
			} else {
				log.Println("aria2 task error, dropping", s.String())
			}
			return
		default:
			log.Println("aria2 state default:", gid, s.String())
		}
		time.Sleep(sleepDur)
	}
}

func delayCleanTask(hash string) {
	defer notifierhub.RedisClient.HDel(redisDelayTask, hash)
	task, err := notifierhub.RedisClient.HGet(redisDelayTask, hash).Result()
	common.Must(err)
	t := &notifierhub.TorrentTask{}
	common.Must(proto.Unmarshal([]byte(task), t))

	durHold := time.Minute * time.Duration(viper.GetInt64("delay_remove"))
	finished := time.Unix(t.FinishTS, 0)
	if time.Since(finished) > durHold {
		log.Println("[delayCleanTask] remove task now", hash, t.Path)
		t.StopAndRemove()
		return
	}

	now := time.Now()
	dur := durHold - now.Sub(finished)
	log.Println("[delayCleanTask] remove task in", dur, hash, t.Path)
	<-time.After(dur)
	t.StopAndRemove()
}

func setCronTask() {
	tz, _ := time.LoadLocation("Asia/Shanghai")
	c := cron.New(cron.WithLocation(tz))
	// cron.WithLogger(cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", 0))))

	c.AddFunc("35 09 * * 1-5", notiIPOCalen)
	c.AddFunc("*/15 * * * *", rss.FindFromRSS)
	for _, job := range viper.GetStringSlice("StorkCron") {
		c.AddFunc(job, notifyStock)
	}
	c.Start()
}
