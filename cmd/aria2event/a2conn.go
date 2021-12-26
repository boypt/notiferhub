package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/boypt/notiferhub/aria2rpc"
	"github.com/boypt/notiferhub/common"
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

func (a *Aria2Conn) OnDownloadComplete(gid string) {
	log.Println("OnDownloadComplete", gid)
	if s, err := a.rpc.TellStatus(gid); err == nil {
		msg := ""
		fn := s.Get("files")
		tl := s.Get("totalLength")
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

		go postMessage(msg)
		log.Println("complete sent", msg)
	} else {
		log.Println("complete err", err)
	}
}

func (a *Aria2Conn) OnDownloadError(gid string) {
	log.Println("OnDownloadError", gid)
	if s, err := a.rpc.TellStatus(gid); err == nil {
		msg := fmt.Sprintf("Error (%s)\n*%s*", s.Get("errorMessage"), s.Get("files"))
		go postMessage(msg)
		log.Println("error sent", msg)
	}
}

func (a *Aria2Conn) EventLoop() {

	fireTimer := time.After(time.Hour * 3)

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
		case <-fireTimer:
			log.Println("fireTimer fired")
			err := a.rpc.Terminate()
			if err != nil {
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
			a.OnDownloadComplete(gid)
		case "aria2.onDownloadError":
			log.Println("error, ", gid)
			a.OnDownloadError(gid)
		default:
			log.Println("unprocess event:", method)
		}
	}
}
