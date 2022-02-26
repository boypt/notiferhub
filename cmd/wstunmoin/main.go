package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/ti-mo/conntrack"
	"golang.zx2c4.com/wireguard/wgctrl"
)

const (
	FlowStopThreshold  = 3
	FlowStartThreshold = 8
	UnitName           = "wstun.service"
)

var (
	systemd   *dbus.Conn
	monIPAddr = net.IPv4(192, 168, 8, 58)
)

func getUnitStatus() string {
	if p, err := systemd.GetUnitProperty(UnitName, "ActiveState"); err == nil {
		return p.Value.Value().(string)
	} else {
		return err.Error()
	}
}

func stopUnit() {
	log.Println("stopping wstun")
	wait := make(chan string)
	systemd.StopUnit(UnitName, "replace", wait)
	<-wait
	log.Println("stopped wstun")
}

func startUnit() {
	log.Println("restarting wstun")
	wait := make(chan string)
	systemd.RestartUnit(UnitName, "replace", wait)
	<-wait
	log.Println("restarted wstun")
}

func ConnTrackWorking(c *conntrack.Conn) uint {
	flows, err := c.Dump()
	if err != nil {
		log.Println(err)
		return 0
	}

	var cnt uint
	for _, f := range flows {
		if f.TupleOrig.IP.SourceAddress.Equal(monIPAddr) {
			cnt++
		}
	}

	return cnt
}

func IsWgLocal(c *wgctrl.Client) bool {

	devs, err := c.Devices()
	if err != nil {
		log.Fatalf("failed to get devices: %v", err)
	}

	if len(devs) == 0 {
		return false
	}

	for _, dev := range devs {
		for _, peer := range dev.Peers {
			if !peer.Endpoint.IP.IsLoopback() {
				return false
			}
		}
	}

	return true
}

func main() {
	if sc, err := dbus.NewSystemdConnectionContext(context.Background()); err != nil {
		log.Fatal("failed to connect to systemd:", err)
	} else {
		systemd = sc
	}
	cc, err := conntrack.Dial(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cc.Close()

	wgc, err := wgctrl.New()
	if err != nil {
		log.Fatalf("failed to open wgctrl: %v", err)
	}
	defer wgc.Close()

	mainTicker := time.NewTicker(time.Second * 10)
	defer mainTicker.Stop()

	for range mainTicker.C {
		status := getUnitStatus()
		switch status {
		case "active":
			if f := ConnTrackWorking(cc); f < FlowStopThreshold {
				log.Println("wstun active, but not enough flows, stopping, flowcnt:", f, "/", FlowStopThreshold)
				stopUnit()
			}
		case "inactive":
			if IsWgLocal(wgc) {
				if f := ConnTrackWorking(cc); f > FlowStartThreshold {
					log.Println("wstun inactive, but enough flows, starting, flowcnt:", f, "/", FlowStartThreshold)
					startUnit()
				}
			}
		default:
			log.Println("unknown unit status:", status)
		}
	}
}
