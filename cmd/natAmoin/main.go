package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/boypt/notiferhub/common"
	"github.com/mdlayher/netlink"
	"github.com/ti-mo/conntrack"
	"github.com/ti-mo/netfilter"
	"golang.org/x/time/rate"
)

const (
	FLOW_THRESHOLD = 3
)

var (
	monIPAddrStr  = flag.String("m", "192.168.8.58", "monitor ip address")
	removeWaitStr = flag.String("r", "10m", "time to wait before removing iptables rules")

	mapRule = []string{"-i", "", "-p", "udp", "-m", "udp", "--dport", "30000:65535", "-j", "DNAT", "--to-destination", "192.168.8.58"}

	removeWait time.Duration
	monIPAddr  net.IP
)

func listenConntrack(notify chan<- struct{}) {
	// Open a Conntrack connection.
	c, err := conntrack.Dial(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Make a buffered channel to receive event updates on.
	evCh := make(chan conntrack.Event, 1024)

	// Listen for all Conntrack and Conntrack-Expect events with 4 decoder goroutines.
	// All errors caught in the decoders are passed on channel errCh.
	errCh, err := c.Listen(evCh, 1, netfilter.GroupsCT)
	if err != nil {
		log.Fatal(err)
	}

	// Listen to Conntrack events from all network namespaces on the system.
	if err := c.SetOption(netlink.ListenAllNSID, true); err != nil {
		log.Fatal(err)
	}
	if err := c.SetOption(netlink.NoENOBUFS, true); err != nil {
		log.Fatal(err)
	}

	lim := rate.NewLimiter(rate.Every(time.Second*10), 1)
	for {
		select {
		case ev := <-evCh:
			if ev.Flow.TupleOrig.IP.SourceAddress.Equal(monIPAddr) && lim.Allow() {
				notify <- struct{}{}
			}
		case err := <-errCh:
			log.Fatalln(err)
		}
	}
}

func isConnTrackWorking(c *conntrack.Conn) bool {
	flows, err := c.Dump()
	if err != nil {
		log.Println(err)
		return false
	}

	var cnt uint
	for _, f := range flows {
		if f.TupleOrig.IP.SourceAddress.Equal(monIPAddr) {
			cnt++
		}
	}

	return cnt > FLOW_THRESHOLD
}

func main() {
	flag.Parse()

	monIPAddr = net.ParseIP(*monIPAddrStr)
	if w, err := time.ParseDuration(*removeWaitStr); err == nil {
		removeWait = w
	} else {
		log.Fatalf("failed to parse removewait: %v", err)
	}

	osexit := make(chan os.Signal, 1)
	signal.Notify(osexit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	if eth, err := common.ExecCommand("/bin/bash", "-c", "LAST='';for K in $(ip route list default); do if [[ $LAST == 'dev' ]]; then echo $K; break; fi; LAST=$K; done"); err == nil {
		mapRule[1] = eth
	} else {
		log.Fatalln("failed to get eth:", err)
	}

	ipt, err := NewIPTCtrl(mapRule)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("monIP:", monIPAddr, "removeWait:", removeWait, "mapRule:", strings.Join(mapRule, " "))

	cc, err := conntrack.Dial(nil)
	if err != nil {
		log.Fatal(err)
	}
	defer cc.Close()

	connEvent := make(chan struct{})
	circleTicker := time.NewTicker(time.Second * 60)
	defer circleTicker.Stop()
	for {
		select {
		case <-circleTicker.C:
			if isConnTrackWorking(cc) {
				ipt.dog.Reset(removeWait)
				if added, err := ipt.AddRule(); err == nil && added {
					log.Println("added iptables rule")
				}
			}
		case <-connEvent:
			if isConnTrackWorking(cc) {
				ipt.dog.Reset(removeWait)
				if added, err := ipt.AddRule(); err == nil && added {
					log.Println("added iptables rule")
				}
			}
		case <-ipt.dog.C:
			if !isConnTrackWorking(cc) {
				if rmed, err := ipt.RemoveRule(); err == nil && rmed {
					log.Println("removed iptables rule")
				}
			}
		case <-osexit:
			log.Println("exit signal fired")
			if rmed, err := ipt.RemoveRule(); err == nil && rmed {
				log.Println("removed iptables rule")
			}
			return
		}
	}
}
