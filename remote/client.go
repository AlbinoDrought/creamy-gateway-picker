package remote

// FirewallRule from remote interface
type FirewallRule interface {
	Source() string
	Destination() string
	Gateway() string
	Description() string
}

// Client connects to the remote Web UI
type Client interface {
	ListRules(iface string) ([]FirewallRule, error)
}
