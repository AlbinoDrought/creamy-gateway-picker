package remote

// FirewallRule from remote interface
type FirewallRule interface {
	Source() string
	Destination() string
	Gateway() string
	Description() string

	Delete() error
}

// Client connects to the remote Web UI
type Client interface {
	ListRules(iface string) ([]FirewallRule, error)
	AddRule(iface, source, destination, gateway, description string) (FirewallRule, error)
}
