package cmd

import (
	"eventhandler/model"

	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/protobuf"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
)

var (
	sender    string
	recipient string
	payload   string
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		nc, err := nats.Connect(cfg.Global.NatsAddress)
		if err != nil {
			log.Fatalf("can't connect to nats server at %s: %s", cfg.Global.NatsAddress, err)
		}
		defer nc.Close()
		// TODO: add payload unmarshaling test
		encConn, err := nats.NewEncodedConn(nc, protobuf.PROTOBUF_ENCODER)
		msg := &model.Envelope{
			[]byte(sender),
			[]byte(recipient),
			[]byte(payload),
			[]byte(`testSignature`),
		}
		log.Debugf("sending message %s", msg.String())
		err = encConn.Publish(cfg.Global.Subject, msg)
		if err != nil {
			log.Error(err.Error())
		}
	},
}

func init() {
	RootCmd.AddCommand(publishCmd)

	publishCmd.Flags().StringVar(&sender, "sender", "localhost", "sender name")
	publishCmd.Flags().StringVar(&recipient, "recipient", "localhost", "recipient name")
	publishCmd.Flags().StringVar(&payload, "payload", "{}", "message payload")
}
