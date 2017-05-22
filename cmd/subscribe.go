package cmd

import (
	"errors"

	"bytes"
	"context"
	"encoding/json"
	"eventhandler/filter"
	"eventhandler/machine"
	"eventhandler/model"
	"eventhandler/runner"
	"fmt"
	"github.com/nats-io/go-nats"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"text/template"
	"time"
)

// subscribeCmd represents the subscribe command
var subscribeCmd = &cobra.Command{
	Use:   "subscribe",
	Short: "Subscribe to the eventhandler queue",
	Long: `Subscribe to the eventhandler queue.

The process listens on the specfied nats topic and runs the specified command if it receives a matching
message. The message payload is rendered via the configured templated and passed to the commands stdin.`,
	Run: func(cmd *cobra.Command, args []string) {
		natsUrl := viper.GetString("nats_url")
		subject := viper.GetString("subject")
		natsOptions := nats.Options{
			Url:            natsUrl,
			AllowReconnect: true,
			MaxReconnect:   -1,
			ReconnectWait:  2 * time.Second,
			DisconnectedCB: func(conn *nats.Conn) {
				log.Warnf("disconnected from nats server(s) %s", conn.Servers())
			},
			ReconnectedCB: func(conn *nats.Conn) {
				log.Infof("successfully reconnected to %s", conn.ConnectedUrl())
			},
		}
		nc, err := natsOptions.Connect()
		if err != nil {
			log.Fatalf("can't connect to nats server at %s: %s", natsUrl, err)
		}
		defer nc.Close()

		// create a coordinator
		coordinator, err := machine.NewCoordinator(nc, cfg.Command.Blackout, cfg.Command.MaxDispatches)
		if err != nil {
			log.Fatal(err)
		}

		// parse the configured template
		stdinTemplate, err := template.New("stdinTemplate").Parse(cfg.Command.StdinTemplate)
		if err != nil {
			log.Fatalf("failed to parse stdin template:", err)
		}

		// a command only waits `timeout` for a command termination.
		// Commands running longer than the timeout are kill -9'ed
		// For further documentation see godoc os/exec CommandContext
		timeout, err := time.ParseDuration(cfg.Command.Timeout)
		if err != nil {
			log.Fatalf("failed to parse cmd timeout:", err)
		}

		// create the runner
		runner := runner.NewPipeRunner(
			context.Background(),
			cfg.Command.Cmd,
			cfg.Command.CmdArgs,
			timeout,
			stdinTemplate,
		)

		// buffer that receives the commands stdout
		cmdStdout := new(bytes.Buffer)

		// start listening on the configured nats topic
		err = coordinator.NatsListen(subject)
		if err != nil {
			log.Fatal(err)
		}

		// create filterer from config
		filters, err := filter.NewFiltererFromConfig(cfg.Command.Filters)
		if err != nil {
			log.Fatal(err)
		}

		// dispatch messaged received from the queue to the handling function, i.e. the runner
		coordinator.Dispatch(filters, func(v interface{}) error {
			var err error
			msg, ok := v.(model.Envelope)
			if !ok {
				return errors.New("failed to type assert protobuf message to envelope")
			}
			log.Debugf("got message %s\n", msg)

			// unmarshal the payload to map[string]string
			payloadData := make(map[string]string)
			err = json.Unmarshal(msg.Payload, &payloadData)
			if err != nil {
				return fmt.Errorf("failed to unmarshal payload: %s", err)
			}

			// run the command with the unmarshaled payload data
			err = runner.Run(payloadData, cmdStdout)
			if err != nil {
				log.Errorf("failed to execute %s: %s", cfg.Command.Cmd, err)
				cmdStdout.Reset()
				return err
			}
			log.Debugf("cmd stdout returned %s", cmdStdout.String())
			cmdStdout.Reset()
			return nil
		})

		// shutdown coordinator on SIGKILL
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)

		<-signalChan
		coordinator.Shutdown()
	},
}

func init() {
	RootCmd.AddCommand(subscribeCmd)

	subscribeCmd.Flags().String("subject", "eventhandler", "nats subject")
	subscribeCmd.Flags().String("nats_url", nats.DefaultURL, "nats url")

	viper.BindPFlag("subject", subscribeCmd.Flags().Lookup("subject"))
	viper.BindPFlag("nats_url", subscribeCmd.Flags().Lookup("nats_url"))
}
