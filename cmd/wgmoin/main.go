package main

import (
	"context"
	"log"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"golang.zx2c4.com/wireguard/wgctrl"
)

type statusTimer struct {
	timer   *time.Timer
	stopped bool
}

func (t *statusTimer) Stop() bool {
	if t.stopped {
		return false
	}
	t.timer.Stop()
	t.stopped = true
	return true
}

func (t *statusTimer) Reset(d time.Duration) bool {
	t.timer.Reset(d)
	t.stopped = false
	return true
}

var (
	startTimeout = time.Second * 10
	stopTimeout  = time.Minute * 1
	startTimer   *statusTimer
	stopTimer    *statusTimer
	systemd      *dbus.Conn
	lastRx       int64
	lastTx       int64
)

func main() {

	c, err := wgctrl.NewSetMode(false)
	if err != nil {
		log.Fatalf("failed to open wgctrl: %v", err)
	}
	defer c.Close()

	var noTrafficCounter uint
	var restartTrafficCounter uint

	for range time.Tick(time.Second * 1) {
		devices, err := c.Devices()
		if err != nil {
			log.Fatalf("failed to get devices: %v", err)
		}
		if len(devices) > 0 {
			p := devices[0].Peers[0]
			if !p.Endpoint.IP.IsLoopback() {
				log.Println("non-loopback peer found")
				startTimer.Stop()
				stopTimer.Stop()
				continue
			}

			diffTx := p.TransmitBytes - lastTx
			diffRx := p.ReceiveBytes - lastRx

			// init
			if lastTx == 0 || lastRx == 0 {
				log.Println("init")
				goto ACTION
			}

			if diffTx > 0 && diffRx > 0 {
				// traffic fine
				log.Println("traffic fine")
				noTrafficCounter = 0
				restartTrafficCounter = 0
				stopTimer.Reset(stopTimeout)
				startTimer.Reset(startTimeout)
				goto ACTION
			}

			if diffRx == 0 && diffTx > 0 {
				log.Println("tunnel needs restart")
				restartTrafficCounter++
				noTrafficCounter = 0
				goto ACTION
			}

			if diffRx == 0 && diffTx == 0 {
				noTrafficCounter++
				log.Println("no traffic +1", p.TransmitBytes)
				goto ACTION
			}

		ACTION:
			lastTx = p.TransmitBytes
			lastRx = p.ReceiveBytes

			// start the stop counter, ready to stop
			if noTrafficCounter > 5 && stopTimer.stopped {
				log.Println("no traffic counter > 5, start stoptimer")
				stopTimer.Reset(stopTimeout)

				startTimer.Stop()
			}

			if restartTrafficCounter > 2 && startTimer.stopped {
				log.Println("has traffic counter > 2, start starttimer")
				startTimer.Reset(startTimeout)

				stopTimer.Stop()
			}
		}
	}

}

func getUnitStatus() string {
	if p, err := systemd.GetUnitProperty("wstun.service", "ActiveState"); err == nil {
		return p.Value.Value().(string)
	} else {
		return err.Error()
	}
}

func timerAction() {
	for {
		select {
		case <-startTimer.timer.C:
			log.Println("start timer fired")
			startTimer.stopped = true
			systemd.RestartUnitContext(context.Background(), "wstun.service", "replace", nil)
		case <-stopTimer.timer.C:
			log.Println("stop timer fired")
			stopTimer.stopped = true
			systemd.StopUnit("wstun.service", "replace", nil)
		}
	}
}

func init() {

	if sc, err := dbus.NewSystemdConnection(); err != nil {
		log.Fatal("failed to connect to systemd:", err)
	} else {
		systemd = sc
	}

	startTimer = &statusTimer{
		timer: time.NewTimer(startTimeout),
	}
	startTimer.Stop()
	stopTimer = &statusTimer{
		timer: time.NewTimer(stopTimeout),
	}
	stopTimer.Stop()

	go timerAction()
}
