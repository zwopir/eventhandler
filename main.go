package main

import (
	"eventhandler/config"
	"eventhandler/machine"
	"eventhandler/model"
	"flag"
	"fmt"
	"github.com/nats-io/go-nats"
	"os"
	"os/signal"
)

func main() {
	var (
		configFile  = flag.String("config", "config.yaml", "config file")
		natsAddress = flag.String("nats-url", nats.DefaultURL, "nats URL")
	)
	config, err := config.FromFile(*configFile)
	if err != nil {
		panic(err)
	}
	nc, err := nats.Connect(*natsAddress)
	if err != nil {
		panic(err)
	}
	natsEncConn, err := nats.NewEncodedConn(nc, "json")
	if err != nil {
		panic(err)
	}
	defer natsEncConn.Close()

	coordinator := machine.NewCoordinator(natsEncConn)
	err = coordinator.NatsListen(config.Global.Subject)
	if err != nil {
		panic(err)
	}

	for i := 0; i <= 10; i++ {
		err := natsEncConn.Publish(config.Global.Subject, model.Message{
			"count":      fmt.Sprintf("%d", i),
			"bla":        "blurks",
			"hostname":   "localhost",
			"check_name": "check_foo",
		})
		if err != nil {
			fmt.Println(err)
		}
	}

	filters, err := model.NewFilters(config.Command.Filters)
	if err != nil {
		panic(err)
	}
	coordinator.Dispatch(filters)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
	coordinator.Shutdown()
}
