package sioclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type EIOMessageType int

const (
	EIOMessageTypeOpen EIOMessageType = iota
	EIOMessageTypeClose
	EIOMessageTypePing
	EIOMessageTypePong
	EIOMessageTypeMessage
	EIOMessageTypeUpgrade
	EIOMessageTypeNoop
)

var ErrNotConnected = errors.New("Websocket is not connected")

type EIOClientConfig struct {
	MessageHandlerBinary func([]byte) error
	MessageHandlerText   func([]byte) error
	URL                  string
}

type eioSessionStart struct {
	SID          string   `json:"sid"`
	Upgrades     []string `json:"upgrades"`
	PingTimeout  int64    `json:"pingTimeout"`
	PingInterval int64    `json:"pingInterval"`
}

type EIOClient struct {
	cfg         EIOClientConfig
	dialer      *websocket.Dialer
	errC        chan error
	isConnected bool
	writeMutex  *sync.Mutex
	ws          *websocket.Conn
}

func NewEIOClient(config EIOClientConfig) (*EIOClient, error) {
	var (
		client = new(EIOClient)
		err    error
	)

	socketURL, err := url.Parse(config.URL)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse URL")
	}

	qVars := socketURL.Query()
	qVars.Set("EIO", "3")
	qVars.Set("transport", "websocket")

	socketURL.RawQuery = qVars.Encode()

	client.cfg = config
	client.errC = make(chan error, 10)
	client.writeMutex = new(sync.Mutex)
	client.ws, _, err = client.dialer.Dial(socketURL.String(), http.Header{})
	if err != nil {
		return nil, errors.Wrap(err, "Unable to dial to given URL")
	}

	// Mark connection as established
	client.isConnected = true

	defaultCloseHandler := client.ws.CloseHandler()
	client.ws.SetCloseHandler(func(code int, text string) error {
		client.isConnected = false
		return defaultCloseHandler(code, text)
	})

	go func() {
		for client.isConnected {
			messageType, message, err := client.ws.ReadMessage()
			if err != nil {
				client.errC <- err
				continue
			}

			if err = client.handleMessage(messageType, message); err != nil {
				client.errC <- err
				continue
			}
		}
	}()

	return client, nil
}

func (e EIOClient) Close() error { return e.ws.Close() }

func (e EIOClient) Errors() <-chan error { return e.errC }

func (e EIOClient) SendTextMessage(t EIOMessageType, data string) error {
	if !e.isConnected {
		return ErrNotConnected
	}

	e.writeMutex.Lock()
	defer e.writeMutex.Unlock()

	return errors.Wrap(
		e.ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%d%s", t, data))),
		"Unable to transmit message",
	)
}

func (e EIOClient) handleMessage(messageType int, message []byte) error {
	if len(message) < 1 {
		return errors.New("Empty message received")
	}

	var mType EIOMessageType

	switch messageType {

	case websocket.TextMessage:
		v, err := strconv.Atoi(string(message[0]))
		if err != nil {
			return errors.Wrap(err, "Unable to parse message type")
		}
		mType = EIOMessageType(v)

	case websocket.BinaryMessage:
		mType = EIOMessageType(message[0])

	}

	switch mType {

	case EIOMessageTypeOpen:
		var handshake eioSessionStart

		if err := json.Unmarshal(message[1:], &handshake); err != nil {
			return errors.Wrap(err, "Unable to unmarshal handshake")
		}

		// Start pinger
		go func() {
			for t := time.NewTicker(time.Duration(handshake.PingInterval) * time.Millisecond); e.isConnected; <-t.C {
				e.SendTextMessage(EIOMessageTypePing, "")
			}
		}()

	case EIOMessageTypeClose:
		e.ws.Close()

	case EIOMessageTypePing:
		e.SendTextMessage(EIOMessageTypePong, "")

	case EIOMessageTypePong:
		// Ignore

	case EIOMessageTypeMessage:
		var hdl func([]byte) error

		switch messageType {
		case websocket.TextMessage:
			hdl = e.cfg.MessageHandlerText
		case websocket.BinaryMessage:
			hdl = e.cfg.MessageHandlerBinary
		}

		if err := hdl(message[1:]); err != nil {
			return errors.Wrap(err, "Failed to handle message")
		}

	case EIOMessageTypeUpgrade:
		// Ignore?

	case EIOMessageTypeNoop:
		// Noop!

	default:
		return errors.Errorf("Received unknown EIO message type %d", mType)

	}

	return nil
}
