package remote

type sensemillaFirewallRule struct {
	id          string
	iface       string
	source      string
	destination string
	gateway     string
	description string

	client *sensemillaClient
}

func (rule *sensemillaFirewallRule) Source() string {
	return rule.source
}

func (rule *sensemillaFirewallRule) Destination() string {
	return rule.destination
}

func (rule *sensemillaFirewallRule) Gateway() string {
	return rule.gateway
}

func (rule *sensemillaFirewallRule) Description() string {
	return rule.description
}

func (rule *sensemillaFirewallRule) Delete() error {
	return rule.client.deleteRule(rule.iface, rule.id)
}
