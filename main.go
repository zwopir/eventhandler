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
	"github.com/prometheus/common/log"
	"github.com/nats-io/go-nats/encoders/protobuf"
	"errors"
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

	encConn, err := nats.NewEncodedConn(nc, protobuf.PROTOBUF_ENCODER)
	for i := 0; i <= 10; i++ {
		msg := &model.Envelope{
				[]byte(`nagios.example.com`),
				[]byte(`me.example.com`),
				[]byte(`{"check_name":"check_foo"}`),
				[]byte(`testSignature`),
			}
		log.Infof("sending message %s", msg.String())
		err = encConn.Publish(config.Global.Subject, msg)
		if err != nil {
			fmt.Println(err)
		}
	}

	filters, err := model.NewFiltererFromConfig(config.Command.Filters)
	if err != nil {
		panic(err)
	}
	coordinator.Dispatch(filters, func(v interface{}) error {
		msg, ok := v.(model.Envelope)
		if !ok {
			return errors.New("failed to type assert protobuf message to envelope")
		}
		fmt.Printf("got message %s\n", msg)
		return nil
	})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
	coordinator.Shutdown()
}
