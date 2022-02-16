package main

import (
	"log"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"golang.zx2c4.com/wireguard/wgctrl"
)

const (
	IntvalSecs   = 180
	BytesPerSecs = 32
	HitThreshold = 5
)

var (
	systemd *dbus.Conn
	lastRx  int64
	lastTx  int64
)

func main() {

	c, err := wgctrl.New()
	if err != nil {
		log.Fatalf("failed to open wgctrl: %v", err)
	}
	defer c.Close()

	threshold := int64(IntvalSecs * BytesPerSecs)
	hitcounter := 0

	log.Println("mon threshold:", threshold)
	for range time.Tick(time.Second * IntvalSecs) {
		devices, err := c.Devices()
		if err != nil {
			log.Fatalf("failed to get devices: %v", err)
		}
		if len(devices) == 0 {
			continue
		}

		p := devices[0].Peers[0]
		if !p.Endpoint.IP.IsLoopback() {
			continue
		}

		if getUnitStatus() == "inactive" {
			continue
		}

		diffTx := p.TransmitBytes - lastTx
		diffRx := p.ReceiveBytes - lastRx

		log.Println("tx:", diffTx, "rx:", diffRx, "sum:", diffTx+diffRx, "threshold:", threshold)
		if diffTx+diffRx < threshold {
			hitcounter++
			if hitcounter > HitThreshold {
				hitcounter = 0
				go stopUnit()
			}
		} else {
			hitcounter = 0
		}

		lastTx = p.TransmitBytes
		lastRx = p.ReceiveBytes
	}

}

func getUnitStatus() string {
	if p, err := systemd.GetUnitProperty("wstun.service", "ActiveState"); err == nil {
		return p.Value.Value().(string)
	} else {
		return err.Error()
	}
}

func stopUnit() {
	log.Println("stopping wstun")
	wait := make(chan string)
	systemd.StopUnit("wstun.service", "replace", wait)
	<-wait
	log.Println("stopped wstun")
}

func init() {
	if sc, err := dbus.NewSystemdConnection(); err != nil {
		log.Fatal("failed to connect to systemd:", err)
	} else {
		systemd = sc
	}
}
