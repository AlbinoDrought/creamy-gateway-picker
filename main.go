package main

import (
	"log"

	"github.com/AlbinoDrought/creamy-gateway-override/remote"
	"github.com/caarlos0/env"
	"github.com/imroc/req"
)

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalln("error parsing config", err)
	}
	req.SetFlags(req.LreqHead | req.LreqBody)
	req.Debug = true

	client := remote.NewSensemillaClient(cfg.RemoteHost, cfg.RemoteUsername, cfg.RemotePassword)

	rules, err := client.ListRules(cfg.RemoteInterface)
	if err != nil {
		panic(err)
	}

	for _, rule := range rules {
		log.Println(rule.Source(), rule.Destination(), rule.Gateway(), rule.Description())
	}
}
