package main

type config struct {
	RemoteHost      string `env:"CREAMY_GATEWAY_REMOTE_HOST"`
	RemoteUsername  string `env:"CREAMY_GATEWAY_REMOTE_USERNAME"`
	RemotePassword  string `env:"CREAMY_GATEWAY_REMOTE_PASSWORD"`
	RemoteInterface string `env:"CREAMY_GATEWAY_REMOTE_INTERFACE"`

	Gateways      []string `env:"CREAMY_GATEWAY_GATEWAYS" envSeparator:","`
	GatewayLabels []string `env:"CREAMY_GATEWAY_GATEWAY_LABELS" envSeparator:","`
}
