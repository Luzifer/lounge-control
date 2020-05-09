package sioclient

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
)

type MessageType int

const (
	MessageTypeConnect MessageType = iota
	MessageTypeDisconnect
	MessageTypeEvent
	MessageTypeAck
	MessageTypeError
	MessageTypeBinaryEvent
	MessageTypeBinaryAck
)

type Config struct {
	MessageHandler func(*Message) error
	URL            string
}

type Client struct {
	EIO *EIOClient

	cfg Config
}

func New(c Config) (*Client, error) {
	var (
		client = new(Client)
		err    error
	)

	client.cfg = c

	if client.EIO, err = NewEIOClient(EIOClientConfig{
		MessageHandlerText: client.handleTextMessage,
		URL:                c.URL,
	}); err != nil {
		return nil, errors.Wrap(err, "Unable to create EIO client")
	}

	return client, nil
}

func (c Client) Close() error {
	return c.EIO.Close()
}

func (c Client) handleTextMessage(msg []byte) error {
	m, err := c.parseProto(msg)
	if err != nil {
		return errors.Wrap(err, "Unable to parse message")
	}

	return c.cfg.MessageHandler(m)
}

type Message struct {
	Type      MessageType
	Namespace string
	ID        int
	Payload   []json.RawMessage
}

func NewMessage(sType MessageType, id int, payloadType string, data interface{}) (*Message, error) {
	out := &Message{
		Type:      sType,
		Namespace: "/",
		ID:        id,
		Payload:   make([]json.RawMessage, 2),
	}

	var err error

	if out.Payload[0], err = json.Marshal(payloadType); err != nil {
		return nil, errors.Wrap(err, "Unable to marshal payloadType")
	}

	if out.Payload[1], err = json.Marshal(data); err != nil {
		return nil, errors.Wrap(err, "Unable to marshal data")
	}

	return out, nil
}

func (m Message) Encode() (string, error) {
	data, err := json.Marshal(m.Payload)
	if err != nil {
		return "", errors.Wrap(err, "Unable to marshal payload")
	}

	var msg = new(bytes.Buffer)
	msg.WriteString(strconv.Itoa(int(m.Type)))

	if m.Namespace != "" && m.Namespace != "/" {
		msg.WriteString(m.Namespace + ",")
	}

	if m.ID > 0 {
		msg.WriteString(strconv.Itoa(m.ID))
	}

	msg.Write(data)

	return msg.String(), nil
}

func (m Message) PayloadType() (string, error) {
	if len(m.Payload) == 0 {
		return "", errors.New("No payload type available")
	}

	var t string
	err := json.Unmarshal(m.Payload[0], &t)
	return t, errors.Wrap(err, "Unable to unmarshal payload type")
}

func (m Message) Send(c *Client) error {
	raw, err := m.Encode()
	if err != nil {
		return errors.Wrap(err, "Unable to encode message")
	}

	return c.EIO.SendTextMessage(EIOMessageTypeMessage, raw)
}

func (m Message) UnmarshalPayload(out interface{}) error { return json.Unmarshal(m.Payload[1], out) }

func (c Client) parseProto(msg []byte) (*Message, error) {
	var (
		err    error
		outMsg = new(Message)
		ptr    int
	)

	if len(msg) < 1 {
		return nil, errors.New("Message was empty")
	}

	// Get message type
	mType, err := strconv.Atoi(string(msg[ptr]))
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse message type")
	}
	outMsg.Type = MessageType(mType)
	ptr++

	// Message contains only message type
	if len(msg[ptr:]) == 0 {
		return outMsg, nil
	}

	// Binary
	if outMsg.Type == MessageTypeBinaryEvent || outMsg.Type == MessageTypeBinaryAck {
		return nil, errors.New("Binary is not supported")
	}

	// Check for namespace
	if msg[ptr] == '/' {
		return nil, errors.New("Namespaces is not supported")
	}

	// Read message ID if any
	if outMsg.ID, err = strconv.Atoi(string(msg[ptr])); err == nil {
		ptr++
	}

	// If there is no more data we have an empty message
	if len(msg[ptr:]) == 0 || outMsg.Type == MessageTypeConnect {
		return outMsg, nil
	}

	return outMsg, errors.Wrap(json.Unmarshal(msg[ptr:], &outMsg.Payload), "Unable to unmarshal message")
}
