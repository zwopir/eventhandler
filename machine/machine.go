package machine

import (
	"eventhandler/model"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats/encoders/protobuf"
	"github.com/prometheus/common/log"
)

type Coordinator struct {
	envelopeCh chan model.Envelope
	encConn    *nats.EncodedConn
	done       chan struct{}
}

func NewCoordinator(conn *nats.Conn) (Coordinator, error) {
	envelopeCh := make(chan model.Envelope)
	done := make(chan struct{})
	encConn, err := nats.NewEncodedConn(conn, protobuf.PROTOBUF_ENCODER)
	if err != nil {
		return Coordinator{}, err
	}
	return Coordinator{
		envelopeCh: envelopeCh,
		encConn:    encConn,
		done:       done,
	}, nil
}

func (c Coordinator) NatsListen(subject string) error {
	_, err := c.encConn.Subscribe(subject, func(m model.Envelope) {
		c.envelopeCh <- m
	})
	if err != nil {
		return err
	}
	return nil
}

func (c Coordinator) Dispatch(filters model.Filterer, actionFunc func(interface{}) error) {
	go func() {
		for message := range c.envelopeCh {
			matched, err := filters.Match(message)
			if err != nil {
				log.Errorf("failed to apply matcher on %v: %s", message, err)
				return
			}
			if matched {
				log.Debugf("dispatching message %v\n", message)
				err := actionFunc(message)
				if err != nil {
					log.Errorf("action func in dispatcher failed: %s", err)
				}
			} else {
				log.Debugf("discarding message %v\n", message)
			}
			select {
			case <-c.done:
				log.Info("shutting down dispatcher")
				return
			default:
			}
		}
	}()
}

func (c Coordinator) Shutdown() {
	defer close(c.done)
	defer close(c.envelopeCh)
	log.Info("shutting down coordinator...")
	c.encConn.Close()
}
