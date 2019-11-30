package remote

// FirewallRule from remote interface
type FirewallRule interface {
	Source() string
	Destination() string
	Gateway() string
	Description() string

	Delete() error
}

// Gateway configured on remote interface
type Gateway interface {
	Name() string
	Description() string

	GatewayAddress() string
	RoundtripTime() string
	Online() bool
}

// Client connects to the remote Web UI
type Client interface {
	ListGateways() ([]Gateway, error)

	ListRules(iface string) ([]FirewallRule, error)
	AddRule(iface, source, destination, gateway, description string) (FirewallRule, error)
}
