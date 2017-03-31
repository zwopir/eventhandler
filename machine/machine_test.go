package machine

import (
	"ehclient/config"
	"ehclient/model"
	"fmt"
	"github.com/nats-io/gnatsd/server"
	testserver "github.com/nats-io/gnatsd/test"
	"github.com/nats-io/go-nats"
	"net"
	"reflect"
	"testing"
	"time"
)

var (
	subject string = "testsubject"
)

func TestNewCoordinator(t *testing.T) {
	encConn := nats.EncodedConn{}
	_ = NewCoordinator(&encConn)
}

func TestCoordinator_NatsListen(t *testing.T) {
	opts := &server.Options{Host: "127.0.0.1", Port: server.RANDOM_PORT}
	s := testserver.RunServer(opts)
	defer s.Shutdown()

	addr := s.Addr()
	host, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		t.Fatalf("Expected no error: Got %v\n", err)
	}
	t.Logf("started test server on nats://%s:%s", host, port)

	nc, err := nats.Connect(fmt.Sprintf("nats://%s:%s", host, port))
	if err != nil {
		t.Errorf("internal test error: %s", err)
	}
	natsEncConn, err := nats.NewEncodedConn(nc, "json")
	if err != nil {
		t.Errorf("internal test error: %s", err)
	}
	defer natsEncConn.Close()
	coordinator := NewCoordinator(natsEncConn)
	err = coordinator.NatsListen(subject)
	if err != nil {
		t.Errorf("NatsListen returned an error: %s", err)
	}
}

var (
	dispatchTestTable = []struct {
		configFilters      []config.Filter
		messagesToDispatch []*model.Message
		recvMessages       []*model.Message
	}{
		{
			[]config.Filter{
				{
					SourceField: "key_a",
					RegexpMatch: ".+123.+",
				},
			},
			[]*model.Message{
				{"key_a": "matched123matched"},
				{"key_a": "whatever"},
			},
			[]*model.Message{
				{"key_a": "matched123matched"},
			},
		},
	}
)

func TestCoordinator_Dispatch(t *testing.T) {
	for _, tt := range dispatchTestTable {
		// create a coordinator
		encConn := nats.EncodedConn{}
		coordinator := NewCoordinator(&encConn)

		// create test filters
		filters, err := model.NewFilters(tt.configFilters)
		if err != nil {
			t.Errorf("failed to create filter for %v (%s)",
				tt.configFilters,
				err,
			)
		}
		// create chan to receive messages from dispatcher
		recv := make(chan *model.Message)

		// start dispatcher
		coordinator.Dispatch(filters, func(message interface{}) error {
			recv <- message.(*model.Message)
			return nil
		})

		// collect dispatched messages in a go routine
		dispatchedMessages := []*model.Message{}
		go func() {
			for m := range recv {
				dispatchedMessages = append(dispatchedMessages, m)
				select {
				case <-coordinator.done:
					return
				default:
				}
			}
		}()

		// send test messages to
		for _, messageToDispatch := range tt.messagesToDispatch {
			coordinator.messageCh <- messageToDispatch
		}
		time.Sleep(time.Second)
		close(coordinator.done)

		if !reflect.DeepEqual(dispatchedMessages, tt.recvMessages) {
			t.Errorf("expected the following messages (%v), got (%v)", tt.recvMessages, dispatchedMessages)
		}
	}
}

func TestCoordinator_Shutdown(t *testing.T) {
	opts := &server.Options{Host: "127.0.0.1", Port: server.RANDOM_PORT}
	s := testserver.RunServer(opts)
	defer s.Shutdown()

	addr := s.Addr()
	host, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		t.Fatalf("Expected no error: Got %v\n", err)
	}
	t.Logf("started test server on nats://%s:%s", host, port)

	nc, err := nats.Connect(fmt.Sprintf("nats://%s:%s", host, port))
	if err != nil {
		t.Errorf("internal test error: %s", err)
	}
	natsEncConn, err := nats.NewEncodedConn(nc, "json")
	if err != nil {
		t.Errorf("internal test error: %s", err)
	}
	coordinator := NewCoordinator(natsEncConn)
	coordinator.Shutdown()
	time.Sleep(time.Second)
	err = coordinator.encConn.Publish("failsubject", "whatever")
	if err == nil {
		t.Error("publishing on a closed connection should fail")
	} else {
		t.Logf("publishing on a closed connection correctly fails: %s", err)
	}
}