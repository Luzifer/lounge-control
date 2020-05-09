package main

import (
	"encoding/json"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/lounge-control/sioclient"
)

func addGenericHandler(f typedHandlerFunc) handlerFunc {
	return func(msg *sioclient.Message) error {
		if msg.Type != sioclient.MessageTypeEvent {
			// We don't care about anything but events
			return nil
		}

		pType, err := msg.PayloadType()
		if err != nil {
			log.Printf("Event message had no payload type: %#v - %s", msg, err)
			return nil
		}

		switch pType {

		case "auth:failed":
			log.Fatal("Login failed")

		case "auth:start":
			msg, err := sioclient.NewMessage(sioclient.MessageTypeEvent, 0, "auth:perform", map[string]string{"user": cfg.Username, "password": cfg.Password})
			if err != nil {
				return errors.Wrap(err, "Unable to create auth:peform")
			}

			if err := msg.Send(client); err != nil {
				return errors.Wrap(err, "Unable to create payload")
			}

		case "init":
			if err := json.Unmarshal(msg.Payload[1], &initData); err != nil {
				return errors.Wrap(err, "Unable to parse init payload")
			}

		}

		return f(pType, msg)
	}
}

// DEPRECATED: Only storing code for now
func handleMessage(msg *sioclient.Message) error {
	if msg.Type != sioclient.MessageTypeEvent {
		// We don't care about anything but events
		return nil
	}

	pType, err := msg.PayloadType()
	if err != nil {
		log.Printf("Event message had no payload type: %#v - %s", msg, err)
		return nil
	}

	switch pType {

	case "auth:failed":
		log.Fatal("Login failed")

	case "auth:start":
		msg, err := sioclient.NewMessage(sioclient.MessageTypeEvent, 0, "auth:perform", map[string]string{"user": cfg.Username, "password": cfg.Password})
		if err != nil {
			return errors.Wrap(err, "Unable to create auth:peform")
		}

		if err := msg.Send(client); err != nil {
			return errors.Wrap(err, "Unable to create payload")
		}

	case "auth:success":
		log.Info("Logged in successfully")

	case "init":
		if err := json.Unmarshal(msg.Payload[1], &initData); err != nil {
			return errors.Wrap(err, "Unable to parse init payload")
		}

	case "msg":
		// Message in channel
		var payload chatMessage
		if err := json.Unmarshal(msg.Payload[1], &payload); err != nil {
			return errors.Wrap(err, "Unable to parse init payload")
		}

		switch payload.Msg.Type {

		case "join", "part":
			// Don't care

		case "message", "notice":
			//log.Infof("Msg: %#v", payload)

		case "unhandled":
			log.Infof("CMD: %s", msg.Payload[1])

		default:
			log.Infof("Unhandled message %q: %#v", payload.Msg.Type, payload)

		}

	case "commands", "configuration", "names", "open", "push:issubscribed", "users":
		// Drop irrelevantt messages

	default:
		if len(msg.Payload) == 2 {
			log.Warnf("Recieved unhandled message %q: %s", pType, msg.Payload[1])
			return nil
		}
		log.Warnf("Recieved unhandled message %q: %#v", pType, msg)
	}

	return nil
}
