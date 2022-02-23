package main

import (
	"log"
	"sync"
	"time"

	"github.com/coreos/go-iptables/iptables"
)

type IPTCtrl struct {
	sync.Mutex
	dog       *time.Timer
	ipt       *iptables.IPTables
	iptRule   []string
	isMapping bool
}

func NewIPTCtrl(iptRule []string) (*IPTCtrl, error) {
	ipt, err := iptables.New()
	if err != nil {
		return nil, err
	}

	isMapping, err := ipt.Exists("nat", "PREROUTING", iptRule...)
	if err != nil {
		return nil, err
	}

	return &IPTCtrl{
		dog:       time.NewTimer(removeWait),
		ipt:       ipt,
		iptRule:   iptRule,
		isMapping: isMapping,
	}, nil
}

func (c *IPTCtrl) AddRule() error {

	c.Lock()
	defer c.Unlock()

	if !c.isMapping {
		if err := c.ipt.Insert("nat", "PREROUTING", 1, c.iptRule...); err != nil {
			return err
		}
		c.isMapping = true
		log.Println("rule added")
	}

	return nil
}

func (c *IPTCtrl) RemoveRule() error {

	c.Lock()
	defer c.Unlock()

	if c.isMapping {
		if err := c.ipt.Delete("nat", "PREROUTING", c.iptRule...); err != nil {
			return err
		}
		c.isMapping = false
		log.Println("rule removed")
	}

	return nil
}
