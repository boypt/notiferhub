package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coreos/go-iptables/iptables"
	"github.com/mdlayher/netlink"
	"github.com/ti-mo/conntrack"
	"github.com/ti-mo/netfilter"
	"golang.org/x/time/rate"
)

var (
	monIPAddrStr  = flag.String("m", "192.168.8.58", "monitor ip address")
	removeWaitStr = flag.String("r", "10m", "time to wait before removing iptables rules")

	mainInt = flag.String("e", "eth0", "interface of main traffic")
	mapRule = []string{"-p", "udp", "-m", "udp", "--dport", "30000:65535", "-j", "DNAT", "--to-destination", "192.168.8.58"}

	removeWait time.Duration
	monIPAddr  net.IP
)

func main() {
	flag.Parse()

	osexit := make(chan os.Signal, 1)
	signal.Notify(osexit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	iptRule := append([]string{"-i", *mainInt}, mapRule...)
	monIPAddr = net.ParseIP(*monIPAddrStr)
	if w, err := time.ParseDuration(*removeWaitStr); err == nil {
		removeWait = w
	} else {
		log.Fatalf("failed to parse removewait: %v", err)
	}

	log.Println("monIP:", monIPAddr, "removeWait:", removeWait, "rule:", iptRule)
	// Open a Conntrack connection.
	c, err := conntrack.Dial(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Make a buffered channel to receive event updates on.
	evCh := make(chan conntrack.Event, 1024)

	// Listen for all Conntrack and Conntrack-Expect events with 4 decoder goroutines.
	// All errors caught in the decoders are passed on channel errCh.
	errCh, err := c.Listen(evCh, 4, append(netfilter.GroupsCT, netfilter.GroupsCTExp...))
	if err != nil {
		log.Fatal(err)
	}

	// Listen to Conntrack events from all network namespaces on the system.
	err = c.SetOption(netlink.ListenAllNSID, true)
	if err != nil {
		log.Fatal(err)
	}

	go func() {

		dog := time.NewTimer(removeWait)
		ipt, err := iptables.New()
		if err != nil {
			log.Fatalln(err)
		}

		isMapping, err := ipt.Exists("nat", "PREROUTING", iptRule...)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("nat-type A rules:", isMapping)
		lim := rate.NewLimiter(rate.Every(time.Second), 1)

		for {
			select {
			case ev := <-evCh:
				if ev.Flow.TupleOrig.IP.SourceAddress.Equal(monIPAddr) ||
					ev.Flow.TupleReply.IP.DestinationAddress.Equal(monIPAddr) {

					if lim.Allow() {
						dog.Reset(removeWait)
						if !isMapping {
							log.Println("adding nat-type A iptables rules")
							if err := ipt.Insert("nat", "PREROUTING", 1, iptRule...); err != nil {
								log.Fatalln(err)
							}
							isMapping = true
						}
					}
				}
			case <-dog.C:
				if isMapping {
					log.Println("removeing nat-type A iptables rules")
					if err := ipt.Delete("nat", "PREROUTING", iptRule...); err != nil {
						log.Fatalln(err)
					}
					isMapping = false
				}
			case <-osexit:
				if isMapping {
					log.Println("exit removing iptables rules")
					if err := ipt.Delete("nat", "PREROUTING", iptRule...); err != nil {
						log.Fatalln(err)
					}
				}
				os.Exit(0)
			}
		}
	}()

	// Stop the program as soon as an error is caught in a decoder goroutine.
	log.Print(<-errCh)
}