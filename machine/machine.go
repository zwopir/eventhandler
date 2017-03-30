package machine

import (
	"ehclient/model"
	"github.com/nats-io/go-nats"
	"github.com/prometheus/common/log"
)

type Coordinator struct {
	messageCh chan *model.Message
	encConn   *nats.EncodedConn
	done      chan struct{}
}

func NewCoordinator(encConn *nats.EncodedConn) Coordinator {
	messageCh := make(chan *model.Message)
	done := make(chan struct{})
	return Coordinator{
		messageCh: messageCh,
		encConn:   encConn,
		done:      done,
	}
}

func (c Coordinator) NatsListen(subject string) error {
	_, err := c.encConn.Subscribe(subject, func(m *model.Message) {
		c.messageCh <- m
	})
	if err != nil {
		return err
	}
	return nil
}

func (c Coordinator) Dispatch(filters model.Filters) {
	go func() {
		for message := range c.messageCh {
			if filters.MatchAll(message) {
				log.Debugf("dispatching message %v\n", message)
			} else {
				log.Debugf("discarding message %v\n", message)
			}
			select {
			case <-c.done:
				return
			default:
			}
		}
	}()
}

func (c Coordinator) Shutdown() {
	defer close(c.done)
	defer close(c.messageCh)
	log.Info("shutting down coordinator...")
	c.encConn.Close()
}
