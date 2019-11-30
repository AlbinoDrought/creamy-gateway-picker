package remote

type sensemillaGateway struct {
	name        string
	gateway     string // ip?
	monitor     string
	rtt         string
	rttsd       string
	loss        string
	status      string
	online      bool
	description string
}

func (gateway *sensemillaGateway) Name() string {
	return gateway.name
}

func (gateway *sensemillaGateway) Description() string {
	return gateway.description
}

func (gateway *sensemillaGateway) GatewayAddress() string {
	return gateway.gateway
}

func (gateway *sensemillaGateway) RoundtripTime() string {
	return gateway.rtt
}

func (gateway *sensemillaGateway) Online() bool {
	return gateway.online
}
