package main

type config struct {
	RemoteHost      string `env:"CREAMY_GATEWAY_REMOTE_HOST"`
	RemoteUsername  string `env:"CREAMY_GATEWAY_REMOTE_USERNAME"`
	RemotePassword  string `env:"CREAMY_GATEWAY_REMOTE_PASSWORD"`
	RemoteInterface string `env:"CREAMY_GATEWAY_REMOTE_INTERFACE"`
	RemoteReferrer  string `env:"CREAMY_GATEWAY_REMOTE_REFERRER"`
}
