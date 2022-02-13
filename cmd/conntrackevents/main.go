package main

import (
	"context"
	"flag"
	"log"
	"net"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/mdlayher/netlink"
	"github.com/ti-mo/conntrack"
	"github.com/ti-mo/netfilter"
)

var (
	monIP     = flag.String("monIP", "", "IP address of the monitor")
	monIPAddr net.IP

	timeout     = time.Duration(time.Minute * 5)
	systemdConn *dbus.Conn
)

func getUnitStatus() string {
	if p, err := systemdConn.GetUnitProperty("wstun.service", "ActiveState"); err == nil {
		return p.Value.Value().(string)
	} else {
		return err.Error()
	}
}

func main() {

	flag.Parse()
	if *monIP == "" {
		log.Fatal("-monIP is required")
	}

	monIPAddr = net.ParseIP(*monIP)
	log.Println("monitoring", monIPAddr)

	// Open a Conntrack connection.
	c, err := conntrack.Dial(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Make a buffered channel to receive event updates on.
	evCh := make(chan conntrack.Event, 1024)

	// Listen for all Conntrack and Conntrack-Expect events with 4 decoder goroutines.
	// All errors caught in the decoders are passed on channel errCh.
	// errCh, err := c.Listen(evCh, 2, append(netfilter.GroupsCT, netfilter.GroupsCTExp...))
	errCh, err := c.Listen(evCh, 2, netfilter.GroupsCT)
	if err != nil {
		log.Fatal(err)
	}

	// Listen to Conntrack events from all network namespaces on the system.
	err = c.SetOption(netlink.ListenAllNSID, true)
	if err != nil {
		log.Fatal(err)
	}

	if sc, err := dbus.NewSystemdConnection(); err != nil {
		log.Fatal("failed to connect to systemd:", err)
	} else {
		systemdConn = sc
	}

	log.Println("starting, unit status: ", getUnitStatus())
	timer := time.NewTimer(timeout)

	go func() {

		cc, err := conntrack.Dial(nil)
		if err != nil {
			log.Fatal(err)
		}

		for range time.Tick(time.Second * 1) {
			estabtcp := 0
			flows, err := cc.Dump()
			if err != nil {
				log.Fatalln(err)
			}
			for _, f := range flows {

				// log.Println(f.ProtoInfo)
				if f.TupleOrig.IP.SourceAddress.Equal(monIPAddr) &&
					f.ProtoInfo.TCP != nil &&
					f.ProtoInfo.TCP.State == 0x03 {
					estabtcp += 1
				}
			}
			if estabtcp > 1 {
				log.Println("estabed tcp > 1, reset timer")
				timer.Reset(timeout)
			}
		}
	}()

	// Start a goroutine to print all incoming messages on the event channel.
	go func() {
		// lim := rate.NewLimiter(rate.Every(time.Second), 1)

		unreplycnt := 0
		for {
			ev := <-evCh

			if (ev.Type == conntrack.EventNew || ev.Type == conntrack.EventUpdate) &&
				ev.Flow.TupleOrig.IP.SourceAddress.Equal(monIPAddr) {

				// normal flow
				// log.Println(ev)
				if getUnitStatus() != "active" {
					log.Println("starting wstun")
					systemdConn.StartUnit("wstun.service", "replace", nil)
				}
				if !ev.Flow.Status.SeenReply() {
					log.Println("unrepled:", unreplycnt)
					unreplycnt += 1
				} else {
					unreplycnt = 0
				}

				if unreplycnt > 30 {
					log.Println("unrely 100, restart wstun")
					systemdConn.RestartUnitContext(context.Background(), "wstun.service", "replace", nil)
					unreplycnt = 0
				}
			}
		}
	}()

	for {
		select {
		case <-timer.C:
			if getUnitStatus() == "active" {
				log.Println("stopping wstun")
				systemdConn.StopUnit("wstun.service", "replace", nil)
			}
		case err := <-errCh:
			log.Println(err)
			break
		}
	}
}
