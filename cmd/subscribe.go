package cmd

import (
	"errors"

	"bytes"
	"context"
	"eventhandler/machine"
	"eventhandler/model"
	"eventhandler/runner"
	"github.com/nats-io/go-nats"
	"github.com/prometheus/common/log"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"text/template"
	"time"
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
			log.Fatalf("can't connect to nats server at %s: %s", cfg.Global.NatsAddress, err)
		}
		defer nc.Close()

		coordinator, err := machine.NewCoordinator(nc, cfg.Command.Blackout, cfg.Command.MaxDispatches)
		if err != nil {
			log.Fatal(err)
		}

		stdinTemplate, err := template.New("stdinTemplate").Parse(cfg.Command.StdinTemplate)
		if err != nil {
			log.Fatalf("failed to parse stdin template:", err)
		}
		timeout, err := time.ParseDuration(cfg.Command.Timeout)
		if err != nil {
			log.Fatalf("failed to parse cmd timeout:", err)
		}

		runner := runner.NewPipeRunner(
			context.Background(),
			cfg.Command.Cmd,
			cfg.Command.CmdArgs,
			timeout,
			stdinTemplate,
		)

		err = coordinator.NatsListen(cfg.Global.Subject)
		if err != nil {
			log.Fatal(err)
		}

		filters, err := model.NewFiltererFromConfig(cfg.Command.Filters)
		if err != nil {
			log.Fatal(err)
		}
		coordinator.Dispatch(filters, func(v interface{}) error {
			cmdStdout := new(bytes.Buffer)
			msg, ok := v.(model.Envelope)
			if !ok {
				return errors.New("failed to type assert protobuf message to envelope")
			}
			log.Debugf("got message %s\n", msg)
			// TODO: add marshaling
			err := runner.Run(msg.Payload, cmdStdout)
			if err != nil {
				log.Errorf("failed to execute %s: %s", cfg.Command.Cmd, err)
				return err
			}
			log.Debugf("cmd stdout returned %s", cmdStdout.Bytes())
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
}
