package main

import (
	"log"
	"net"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/ti-mo/conntrack"
	"golang.zx2c4.com/wireguard/wgctrl"
)

const (
	IntvalSecs   = 180
	BytesPerSecs = 32
	HitThreshold = 5
	UnitName     = "wstun.service"
)

var (
	systemd   *dbus.Conn
	monIPAddr = net.IPv4(192, 168, 8, 58)
	stopWait  = time.Minute * 10
)

func wgMoinStop() {
	log.Println("wgmoinstop started")

	c, err := wgctrl.New()
	if err != nil {
		log.Fatalf("failed to open wgctrl: %v", err)
	}
	defer c.Close()

	threshold := int64(IntvalSecs * BytesPerSecs)
	hitcounter := 0
	ticker := time.NewTicker(time.Second * IntvalSecs)

	var lastRx, lastTx int64
	if devices, err := c.Devices(); err == nil && len(devices) > 0 {
		p := devices[0].Peers[0]
		lastTx = p.TransmitBytes
		lastRx = p.ReceiveBytes
	}
	log.Println("tx:", lastTx, "rx:", lastRx, "threshold diff:", threshold)

	for range ticker.C {

		if getUnitStatus() == "inactive" {
			log.Println("wstun inactive during stop moin, exiting stop moin")
			return
		}

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

		diffTx := p.TransmitBytes - lastTx
		diffRx := p.ReceiveBytes - lastRx

		log.Println("tx:", diffTx, "rx:", diffRx, "sum:", diffTx+diffRx, "threshold:", threshold)
		if diffTx+diffRx < threshold {
			hitcounter++
			if hitcounter > HitThreshold {
				stopUnit()
				return
			}
		} else {
			hitcounter = 0
		}

		lastTx = p.TransmitBytes
		lastRx = p.ReceiveBytes
	}
}

func wgMoinStart() {
	log.Println("wgmoinstart started")

	cc, err := conntrack.Dial(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cc.Close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {

		if getUnitStatus() == "active" {
			// restart ?
			log.Println("wstun active during start moin, exiting start moin")
			return
		}

		flows, err := cc.Dump()
		if err != nil {
			log.Fatalln(err)
		}

		flowcnt := 0
		for _, f := range flows {
			if f.TupleOrig.IP.SourceAddress.Equal(monIPAddr) {
				flowcnt++
			}
		}
		if flowcnt > 1 {
			log.Println("estabed flow > 1, start wstun")
			startUnit()
			return
		}
	}

}

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

func main() {
	if sc, err := dbus.NewSystemdConnection(); err != nil {
		log.Fatal("failed to connect to systemd:", err)
	} else {
		systemd = sc
	}

	stopTimer := time.NewTimer(stopWait)
	stopTimer.Stop()

	for {
		status := getUnitStatus()
		log.Println("wstun status:", status)
		switch status {
		case "active":
			wgMoinStop()
			stopTimer.Reset(stopWait)
			<-stopTimer.C
		case "inactive":
			wgMoinStart()
		default:
			log.Println("unknown unit status:", status)
		}

		time.Sleep(time.Second * 10)
	}
}
