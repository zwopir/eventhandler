package cmd

import (
	"fmt"
	"errors"

	"eventhandler/machine"
	"eventhandler/model"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/protobuf"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
)

// subscribeCmd represents the subscribe command
var subscribeCmd = &cobra.Command{
	Use:   "subscribe",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		nc, err := nats.Connect(cfg.Global.NatsAddress)
		if err != nil {
			panic(err)
		}
		defer nc.Close()

		coordinator, err := machine.NewCoordinator(nc)
		if err != nil {
			panic(err)
		}
		err = coordinator.NatsListen(cfg.Global.Subject)
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
			err = encConn.Publish(cfg.Global.Subject, msg)
			if err != nil {
				fmt.Println(err)
			}
		}

		filters, err := model.NewFiltererFromConfig(cfg.Command.Filters)
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
	},
}

func init() {
	RootCmd.AddCommand(subscribeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// subscribeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// subscribeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
