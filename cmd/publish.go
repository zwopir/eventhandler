package cmd

import (
	"eventhandler/model"

	"bytes"
	"encoding/json"
	"eventhandler/verify"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/protobuf"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var payload string

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a message to the eventhandler queue",
	Long: `Publish a messsage to the eventhandler queue.

The payload must be a hash of strings
formatted as json (for example {"check_name":"check_connection"})`,
	Run: func(cmd *cobra.Command, args []string) {
		// get config values
		sender := viper.GetString("sender")
		recipient := viper.GetString("recipient")
		privateKeyPath := viper.GetString("signkey")

		// validate payload
		if payload == "" {
			log.Fatal("payload is a mandatory parameter")
		}

		// initialize signer if requested
		signMessage := false
		signer := &verify.Signer{}
		if privateKeyPath != "" {
			privkeyBuffer, err := os.Open(privateKeyPath)
			if err != nil {
				log.Fatal(err)
			}
			signer, err = verify.NewSigner(privkeyBuffer)
			if err != nil {
				log.Fatalf("failed to initialize signer: %s", err)
			}
			signMessage = true
		}
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
			log.Fatalf("payload %v is not json unmarshalable", payload)
		}
		// protobuf encode message in the nats queue
		encConn, err := nats.NewEncodedConn(nc, protobuf.PROTOBUF_ENCODER)
		if err != nil {
			log.Fatalf("failed to create encoded nats connection: %s", err)
		}

		// calculate signature if requested
		signature := []byte(``)
		if signMessage {
			signBuffer := new(bytes.Buffer)
			signBuffer.WriteString(sender)
			signBuffer.WriteString(recipient)
			signBuffer.WriteString(payload)
			signature, err = signer.Sign(signBuffer)
			if err != nil {
				log.Fatalf("failed to sign message: %s", err)
			}
		}
		msg := &model.Envelope{
			Sender:    []byte(sender),
			Recipient: []byte(recipient),
			Payload:   []byte(payload),
			Signature: signature,
		}
		log.Debugf("sending message %s", msg.String())
		err = encConn.Publish(cfg.Global.Subject, msg)
		if err != nil {
			log.Fatalf("failed to publish message: %s", err)
		}
		log.Info("sent message")
	},
}

func init() {
	RootCmd.AddCommand(publishCmd)

	// define flags
	publishCmd.Flags().String("sender", "localhost", "sender name")
	publishCmd.Flags().String("recipient", "localhost", "recipient name")
	publishCmd.Flags().String("signkey", "", "private key file for message signing")

	// bind cobra flags to viper
	viper.BindPFlag("sender", publishCmd.Flags().Lookup("sender"))
	viper.BindPFlag("recipient", publishCmd.Flags().Lookup("recipient"))
	viper.BindPFlag("signkey", publishCmd.Flags().Lookup("signkey"))

	// payload is not a viper config value
	publishCmd.Flags().StringVar(&payload, "payload", "", "message payload")

}
