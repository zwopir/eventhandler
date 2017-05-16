package cmd

import (
	"eventhandler/model"

	"encoding/json"
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
	Short: "Publish a message to the eventhandler queue",
	Long: `Publish a messsage to the eventhandler queue.

The payload must be a hash of strings
formatted as json (for example {"check_name":"check_connection"})`,
	Run: func(cmd *cobra.Command, args []string) {
		nc, err := nats.Connect(cfg.Global.NatsAddress)
		if err != nil {
			log.Fatalf("can't connect to nats server at %s: %s", cfg.Global.NatsAddress, err)
		}
		defer nc.Close()

		// unmarshal payload to make sure it can be unmarshaled in subscriber
		// the unmarshaled data is discarded
		payloadData := make(map[string]string)
		err = json.Unmarshal([]byte(payload), &payloadData)
		if err != nil {
			log.Fatal("payload is not json unmarshalable")
		}
		// protobuf encode message in the nats queue
		encConn, err := nats.NewEncodedConn(nc, protobuf.PROTOBUF_ENCODER)
		if err != nil {
			log.Fatalf("failed to create encoded nats connection: %s", err)
		}
		msg := &model.Envelope{
			[]byte(sender),
			[]byte(recipient),
			[]byte(payload),
			[]byte(`testSignature`),
		}
		log.Debugf("sending message %s", msg.String())
		err = encConn.Publish(cfg.Global.Subject, msg)
		if err != nil {
			log.Fatalf("failed to publish message: %s", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(publishCmd)

	publishCmd.Flags().StringVar(&sender, "sender", "localhost", "sender name")
	publishCmd.Flags().StringVar(&recipient, "recipient", "localhost", "recipient name")
	publishCmd.Flags().StringVar(&payload, "payload", "{}", "message payload")
}
