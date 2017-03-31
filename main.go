package main

import (
	"encoding/json"
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
	defer nc.Close()

	coordinator, err := machine.NewCoordinator(nc)
	if err != nil {
		panic(err)
	}
	err = coordinator.NatsListen(config.Global.Subject)
	if err != nil {
		panic(err)
	}

	for i := 0; i <= 10; i++ {
		msg, err := json.Marshal(model.Message{
			"count":      fmt.Sprintf("%d", i),
			"bla":        "blurks",
			"hostname":   "localhost",
			"check_name": "check_foo",
		})
		if err != nil {
			panic(err)
		}
		err = nc.Publish(config.Global.Subject, msg)
		if err != nil {
			fmt.Println(err)
		}
	}

	filters, err := model.NewFilters(config.Command.Filters)
	if err != nil {
		panic(err)
	}
	coordinator.Dispatch(filters, func(v interface{}) error {
		fmt.Printf("got message (%v)\n", v.(*model.Message))
		return nil
	})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
	coordinator.Shutdown()
}
