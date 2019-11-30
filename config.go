package main

type gateway struct {
	Name       string
	Label      string
	StatusName string
}

type config struct {
	Debug bool `env:"CREAMY_GATEWAY_DEBUG"`

	RemoteHost      string `env:"CREAMY_GATEWAY_REMOTE_HOST"`
	RemoteUsername  string `env:"CREAMY_GATEWAY_REMOTE_USERNAME"`
	RemotePassword  string `env:"CREAMY_GATEWAY_REMOTE_PASSWORD"`
	RemoteInterface string `env:"CREAMY_GATEWAY_REMOTE_INTERFACE"`

	GatewayNames       []string `env:"CREAMY_GATEWAY_GATEWAYS" envSeparator:","`
	GatewayLabels      []string `env:"CREAMY_GATEWAY_GATEWAY_LABELS" envSeparator:","`
	GatewayStatusNames []string `env:"CREAMY_GATEWAY_GATEWAY_STATUS_NAMES" envSeparator:","`

	Gateways []gateway

	TrustForwardedHeaders bool   `env:"CREAMY_GATEWAY_TRUST_FORWARDED_HEADERS"`
	Port                  string `env:"CREAMY_GATEWAY_PORT" envDefault:"5000"`
}
