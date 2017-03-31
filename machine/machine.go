package machine

import (
	"eventhandler/model"
	"github.com/nats-io/go-nats"
	"github.com/prometheus/common/log"
)

type Coordinator struct {
	messageCh chan *model.Message
	encConn   *nats.EncodedConn
	done      chan struct{}
}

func NewCoordinator(conn *nats.Conn) (Coordinator, error) {
	messageCh := make(chan *model.Message)
	done := make(chan struct{})
	encConn, err := nats.NewEncodedConn(conn, "json")
	if err != nil {
		return Coordinator{}, err
	}
	return Coordinator{
		messageCh: messageCh,
		encConn:   encConn,
		done:      done,
	}, nil
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

func (c Coordinator) Dispatch(filters model.Filters, callback func(interface{}) error) {
	go func() {
		for message := range c.messageCh {
			if filters.MatchAll(message) {
				log.Debugf("dispatching message %v\n", message)
				err := callback(message)
				if err != nil {
					log.Errorf("callback in dispatcher failed: %s", err)
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
	defer close(c.messageCh)
	log.Info("shutting down coordinator...")
	c.encConn.Close()
}
