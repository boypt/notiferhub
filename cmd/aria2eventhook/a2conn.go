package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"notiferhub/aria2rpc"
	"notiferhub/common"
)

type Aria2Conn struct {
	rpcurl string
	token  string
	rpc    *aria2rpc.Aria2WSRPC
	taskts map[string]time.Time
}

func NewAria2Conn(rpcurl, token string) *Aria2Conn {

	return &Aria2Conn{
		rpcurl: rpcurl,
		token:  token,
		taskts: make(map[string]time.Time),
	}
}

func (a *Aria2Conn) InitConn() error {

	client, err := aria2rpc.NewAria2WSRPC(a.token, a.rpcurl)
	if err != nil {
		return err
	}

	client.WebsocketMsgBackgroundRoutine()
	a.rpc = client

	return nil
}

func (a *Aria2Conn) InitInfo() {
	log.Println(a.rpc.GetVersion())
	log.Println(a.rpc.GetSessionInfo())
}

func (a *Aria2Conn) OnDownloadComplete(gid string) error {
	log.Println("OnDownloadComplete", gid)
	if s, err := a.rpc.TellStatus(gid); err == nil {
		msg := ""
		fn := s.Get("files")
		tl := s.Get("totalLength")
		dir := s.Get("dir")

		// log.Println("files:", fn, "dir:", dir)
		fn = strings.Split(fn, "|")[0]
		fn = strings.TrimPrefix(fn, dir+"/")

		ts, exists := a.taskts[gid]
		if exists {
			if tlen, err := strconv.ParseInt(tl, 10, 64); err == nil {
				taskDur := time.Since(ts)
				secs := taskDur.Seconds()
				speed := float64(tlen) / secs
				speedText := common.HumaneSize(int64(speed))
				log.Println("completed", gid, fn, speedText)
				msg = fmt.Sprintf("*%s*\nStatus: *complete*\nDur: *%s*\nAvg: *%s/s*", fn, common.KitchenDuration(taskDur), speedText)
			}
			delete(a.taskts, gid)

		} else {
			msg = fmt.Sprintf("*%s*\nStatus: *complete*", fn)
		}

		go postUntilSuccess(msg)
		log.Println("complete sent", msg)
	} else {
		return err
	}

	return nil
}

func (a *Aria2Conn) OnDownloadError(gid string) {
	log.Println("OnDownloadError", gid)
	if s, err := a.rpc.TellStatus(gid); err == nil {
		msg := fmt.Sprintf("Error (%s)\n*%s*", s.Get("errorMessage"), s.Get("files"))
		go postUntilSuccess(msg)
		log.Println("error sent", msg)
	}
}

func (a *Aria2Conn) EventLoop() {

	fireTimer := time.NewTicker(time.Hour * 5)
	defer fireTimer.Stop()

	for {
		method := ""
		gid := ""

		select {
		case ev := <-a.rpc.NotifyQueue:
			gid = ev.Params[0].(map[string]interface{})["gid"].(string)
			method = ev.Method
		case <-a.rpc.Close:
			log.Println("a2wsclient.Close closed")
			return
		case <-fireTimer.C:
			log.Println("fireTimer fired")
			if err := a.rpc.Terminate(); err != nil {
				log.Println("Terminate err", err)
			}
			return
		}

		switch method {
		case "aria2.onDownloadStart":
			a.taskts[gid] = time.Now()
			log.Println("onDownloadStart, ", gid)
		case "aria2.onDownloadComplete":
			log.Println("complete, ", gid)
			if err := a.OnDownloadComplete(gid); err != nil {
				log.Println("OnDownloadComplete err", err)
			}
		case "aria2.onDownloadError":
			log.Println("error, ", gid)
			a.OnDownloadError(gid)
		default:
			log.Println("unprocess event:", method)
		}
	}
}
