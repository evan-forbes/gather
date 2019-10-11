package ws

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/evan-forbes/gather/manage"
	"github.com/evan-forbes/sink"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1000000
)

// Socket is a small wrapper around the typical gorrila websocket
// to do some lite data and error handling.
type Socket struct {
	Conn      *websocket.Conn
	URL       string
	Errc      chan error
	StayAlive bool
}

// NewSocket issues a new websocket and attempts to connect to
// the provided URL
func NewSocket(url string, errc chan error) (*Socket, error) {
	s := Socket{URL: url, Errc: errc}
	err := s.Connect()
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create new sock to %s:", url)
	}
	return &s, nil
}

// Listen starts the reading and writing channels to the websocket
func (s *Socket) Listen(ctx context.Context, input chan interface{}) <-chan []byte {
	s.WritePump(input)
	go func() {
		<-ctx.Done()
		s.Close()
	}()
	return s.ReadPump()
}

// Connect will attempt to establish a connection to the URL in the Socket
// instance
func (s *Socket) Connect() error {
	conn, resp, err := websocket.DefaultDialer.Dial(s.URL, nil)
	if err != nil {
		errors.Wrapf(err, "Could not connect ws %s \n %v+: ", s.URL, s.Conn)
		return err
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return errors.New("Could not upgrade protocal with ws " + s.URL)
	}
	s.Conn = conn
	s.StayAlive = true
	manage.Logger <- sink.Wrap("Connections", fmt.Sprintf("Successfully connected to %s at %s", s.URL, time.Now()))
	return nil
}

// Close issues commands and messages to gracefully close the
// websocket
func (s *Socket) Close() error {
	manage.Logger <- sink.Wrap("Connections", fmt.Sprintf("closing connection to %s at %s\n", s.URL, time.Now()))
	s.StayAlive = false
	err := s.Conn.WriteMessage(websocket.CloseMessage, []byte{})
	if err != nil {
		return errors.Wrapf(err, "Issue closing ws %s:", s.URL)
	}
	return nil
}

// WritePump starts a thread to wrtie all messages as JSON to the
// to the websocket
func (s *Socket) WritePump(input chan interface{}) {
	go func() {
		for msg := range input {
			err := s.Conn.WriteJSON(msg)
			if err != nil {
				s.Errc <- errors.Wrapf(err, "Issue writing %s", s.URL)
			}
		}
		s.Close()
	}()
}

// ReadPump starts a thread to read all incoming messages into the returned
// channel
func (s *Socket) ReadPump() <-chan []byte {
	output := make(chan []byte, 20)
	go func() {
		defer close(output)
		s.Conn.SetReadLimit(maxMessageSize)
		for s.StayAlive {
			_, payload, err := s.Conn.ReadMessage()
			if err != nil && s.StayAlive {
				s.Errc <- manage.SignalWrap(
					errors.Wrapf(err, "ReadPump stopped for %s:", s.URL),
					manage.REBOOT,
				)
				break
			}
			output <- payload
		}
	}()
	return output
}
