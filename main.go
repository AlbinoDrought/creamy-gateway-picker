package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/AlbinoDrought/creamy-gateway-override/remote"
	"github.com/caarlos0/env"
	"github.com/imroc/req"
)

var client remote.Client
var cfg config

func main() {
	if err := env.Parse(&cfg); err != nil {
		log.Fatalln("error parsing config", err)
	}

	if len(cfg.GatewayLabels) != len(cfg.GatewayNames) {
		log.Println("gateway label and name mismatch, using names as labels")
		cfg.GatewayLabels = cfg.GatewayNames
	}

	if len(cfg.GatewayStatusNames) != len(cfg.GatewayNames) {
		log.Println("gateway status name and normal name mismatch, using names as status names")
		cfg.GatewayStatusNames = cfg.GatewayNames
	}

	gateways := make([]gateway, len(cfg.GatewayNames))
	for i, gatewayName := range cfg.GatewayNames {
		gateways[i].Name = gatewayName
		gateways[i].Label = cfg.GatewayLabels[i]
		gateways[i].StatusName = cfg.GatewayStatusNames[i]
	}
	cfg.Gateways = gateways

	if cfg.Debug {
		req.SetFlags(req.LreqHead | req.LreqBody)
		req.Debug = true
	}

	client = remote.NewSensemillaClient(cfg.RemoteHost, cfg.RemoteUsername, cfg.RemotePassword)

	ctx, cancel := context.WithCancel(context.Background())

	gracefulWaitGroup := sync.WaitGroup{}
	gracefulShutdownComplete := make(chan bool, 1)

	go func() {
		log.Println("performing client self-check")
		_, err := getGatewayStatus()
		if err != nil {
			log.Fatal("error during list gateways self-check", err)
		}
		log.Println("client self-check passed!")
	}()

	serverFinished := bootServer(ctx)
	gracefulWaitGroup.Add(1)
	go func() {
		err := <-serverFinished
		if err != nil {
			log.Println("server exited with error", err)
		}
		gracefulWaitGroup.Done()
	}()

	go func() {
		gracefulWaitGroup.Wait()
		gracefulShutdownComplete <- true
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	firstInterrupt := true

	for {
		select {
		case <-gracefulShutdownComplete:
			log.Println("Graceful shutdown finished, bye!")
			return
		case <-c:
			if firstInterrupt {
				firstInterrupt = false
				cancel()
				log.Println("Interrupt received, initiated graceful shutdown")
			} else {
				log.Println("Performing unclean shutdown")
				return
			}
		}
	}
}
