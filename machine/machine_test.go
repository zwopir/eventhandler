package machine

import (
	"eventhandler/config"
	"eventhandler/model"
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
	conn := nats.Conn{}
	_, err := NewCoordinator(&conn)
	if err != nil {
		t.Errorf("failed to construct Coordinator: %s", err)
	}
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
	defer nc.Close()

	coordinator, err := NewCoordinator(nc)
	if err != nil {
		t.Errorf("failed to construct Coordinator: %s", err)
	}
	err = coordinator.NatsListen(subject)
	if err != nil {
		t.Errorf("NatsListen returned an error: %s", err)
	}
}

var (
	dispatchTestTable = []struct {
		configFilters      []config.Filter
		messagesToDispatch []model.Envelope
		receivedMessages   []model.Envelope
	}{
		{
			[]config.Filter{
				{
					Context: "payload",
					Type:    "regexp",
					Args: map[string]string{
						"field":  "check_name",
						"regexp": "check_.+",
					},
				},
				{
					Context: "envelope",
					Type:    "regexp",
					Args: map[string]string{
						"field":  "sender",
						"regexp": "testS.+",
					},
				},
			},
			[]model.Envelope{
				{
					[]byte(`testSender`),
					[]byte(`testRecipient`),
					[]byte(`{"check_name":"check_foo"}`),
					[]byte(`testSignature`),
				},
				{
					[]byte(`testSender`),
					[]byte(`testRecipient`),
					[]byte(`{"check_name":"not_gonna_match"}`),
					[]byte(`testSignature`),
				},
				{
					[]byte(`anotherSender`),
					[]byte(`testRecipient`),
					[]byte(`{"check_name":"check_foo"}`),
					[]byte(`testSignature`),
				},
				{
					[]byte(`testSammy`),
					[]byte(`testRecipient`),
					[]byte(`{"check_name":"check_foo"}`),
					[]byte(`testSignature`),
				},
			},
			[]model.Envelope{
				{
					[]byte(`testSender`),
					[]byte(`testRecipient`),
					[]byte(`{"check_name":"check_foo"}`),
					[]byte(`testSignature`),
				},
				{
					[]byte(`testSammy`),
					[]byte(`testRecipient`),
					[]byte(`{"check_name":"check_foo"}`),
					[]byte(`testSignature`),
				},
			},
		},
	}
)

func TestCoordinator_Dispatch(t *testing.T) {
	for _, tt := range dispatchTestTable {
		// create a coordinator
		conn := nats.Conn{}
		coordinator, err := NewCoordinator(&conn)
		if err != nil {
			t.Errorf("failed to construct Coordinator: %s", err)
			t.Fail()
		}

		// create test filters
		filters, err := model.NewFiltererFromConfig(tt.configFilters)
		if err != nil {
			t.Errorf("failed to create filter for %v (%s)",
				tt.configFilters,
				err,
			)
			t.Fail()
		}
		// create chan to receive messages from dispatcher
		recv := make(chan model.Envelope)

		// start dispatcher
		coordinator.Dispatch(filters, func(message interface{}) error {
			msg, ok := message.(model.Envelope)
			if !ok {
				t.Error("assertion failed")
			}
			recv <- msg
			return nil
		})

		// collect dispatched messages in a go routine
		dispatchedMessages := []model.Envelope{}
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

		// send test messages to coordinator message chan
		for _, messageToDispatch := range tt.messagesToDispatch {
			t.Logf("sending %v to message channel", messageToDispatch)
			coordinator.envelopeCh <- messageToDispatch
		}
		time.Sleep(time.Second)
		close(coordinator.done)

		if !reflect.DeepEqual(dispatchedMessages, tt.receivedMessages) {
			t.Errorf("expected the following messages %v, got %v", tt.receivedMessages, dispatchedMessages)
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

	conn, err := nats.Connect(fmt.Sprintf("nats://%s:%s", host, port))
	if err != nil {
		t.Errorf("internal test error: %s", err)
	}

	coordinator, err := NewCoordinator(conn)
	if err != nil {
		t.Errorf("failed to construct Coordinator: %s", err)
		t.Fail()
	}
	coordinator.Shutdown()
	time.Sleep(time.Second)
	testmessage := &model.Envelope{
		Sender:    []byte("testSender"),
		Recipient: []byte("testRecipient"),
		Payload:   []byte(`testPayload`),
		Signature: []byte(`testSignature`),
	}
	err = coordinator.encConn.Publish("failsubject", testmessage)
	if err == nil {
		t.Error("publishing on a closed connection should fail")
	} else {
		t.Logf("publishing on a closed connection correctly fails: %s", err)
	}
	_, doneNotClosed := <-coordinator.done
	if doneNotClosed {
		t.Error("done chan hasn't been closed")
	}
}
