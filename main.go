package main

import (
	"log"
	"strings"
	"sync"

	"github.com/AlbinoDrought/creamy-gateway-override/remote"
	"github.com/caarlos0/env"
	"github.com/imroc/req"
)

const dork = "[creamy-gateway]"
const deleteDork = "creamy-gateway-delete"

var statelock sync.Mutex
var client remote.Client

func setGateway(iface, source, gateway, label string) (remote.FirewallRule, error) {
	statelock.Lock()
	defer statelock.Unlock()

	rules, err := client.ListRules(iface)
	if err != nil {
		return nil, err
	}

	// check for old rules, remove them:
	for _, rule := range rules {
		if rule.Source() == source && strings.HasPrefix(rule.Description(), dork) {
			err = rule.Delete()
			if err != nil {
				return nil, err
			}

			break
		}
	}

	if gateway == deleteDork {
		return nil, nil
	}

	// create new rule:
	description := dork + " user chose \"" + label + "\" (" + gateway + ")"
	return client.AddRule(iface, source, "*", gateway, description)
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalln("error parsing config", err)
	}
	req.SetFlags(req.LreqHead | req.LreqBody)
	req.Debug = true

	client = remote.NewSensemillaClient(cfg.RemoteHost, cfg.RemoteUsername, cfg.RemotePassword)

	/*
		rules, err := client.ListRules(cfg.RemoteInterface)
		if err != nil {
			panic(err)
		}

		log.Println("all rules")
		for _, rule := range rules {
			log.Println(rule.Source(), rule.Destination(), rule.Gateway(), rule.Description())
		}

		rule, err := client.AddRule(cfg.RemoteInterface, "172.16.30.108", "*", "LOADBALANCE", "test")
		if err != nil {
			panic(err)
		}
		log.Println("created rule")
		log.Println(rule.Source(), rule.Destination(), rule.Gateway(), rule.Description())

		rule.Delete()
		log.Println("deleted rule")
	*/

	rule, err := setGateway(cfg.RemoteInterface, "172.16.30.108", "LOADBALANCE", "loadbalance")
	if err != nil {
		panic(err)
	}
	log.Println("created rule")
	log.Println(rule.Source(), rule.Destination(), rule.Gateway(), rule.Description())
}
