package main

import (
	"strings"
	"sync"

	"github.com/AlbinoDrought/creamy-gateway-picker/remote"
)

const dork = "[creamy-gateway]"
const deleteDork = "creamy-gateway-delete"

// our connection to the remote client is stateful.
// we can only be performing one thing at a time.
var statelock sync.Mutex

func getGatewayStatus() ([]remote.Gateway, error) {
	statelock.Lock()
	defer statelock.Unlock()

	return client.ListGateways()
}

func getActiveRule(iface, source string) (remote.FirewallRule, error) {
	statelock.Lock()
	defer statelock.Unlock()

	rules, err := client.ListRules(iface)
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		if rule.Source() == source && strings.HasPrefix(rule.Description(), dork) {
			return rule, nil
		}
	}

	return nil, nil
}

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
