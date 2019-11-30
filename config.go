package main

type gateway struct {
	Name  string
	Label string
}

type config struct {
	Debug bool `env:"CREAMY_GATEWAY_DEBUG"`

	RemoteHost      string `env:"CREAMY_GATEWAY_REMOTE_HOST"`
	RemoteUsername  string `env:"CREAMY_GATEWAY_REMOTE_USERNAME"`
	RemotePassword  string `env:"CREAMY_GATEWAY_REMOTE_PASSWORD"`
	RemoteInterface string `env:"CREAMY_GATEWAY_REMOTE_INTERFACE"`

	GatewayNames  []string `env:"CREAMY_GATEWAY_GATEWAYS" envSeparator:","`
	GatewayLabels []string `env:"CREAMY_GATEWAY_GATEWAY_LABELS" envSeparator:","`

	Gateways []gateway

	Port string `env:"CREAMY_GATEWAY_PORT" envDefault:"5000"`
}
