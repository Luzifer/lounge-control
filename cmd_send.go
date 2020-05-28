package main

import (
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/lounge-control/sioclient"
)

func init() {
	registerCommand("send", commandSend)
}

func commandSend(args []string) handlerFunc {
	if len(args) != 2 {
		log.Fatal("Usage: send <target> <message>")
	}

	var (
		channelName = args[0]
		message     = args[1]
	)

	return addGenericHandler(func(pType string, msg *sioclient.Message) error {
		if pType != "init" {
			return nil
		}

		// After join command is finished we can execute the joins
		network := initData.NetworkByNameOrUUID(cfg.Network)
		if network == nil {
			return errors.New("Network not found")
		}

		var target *channel
		for _, c := range network.Channels {
			if channelName == "lobby" && c.Type == "lobby" {
				target = &c
				break
			} else if channelName == c.Name {
				target = &c
				break
			}
		}

		if target == nil {
			return errors.New("Unable to find channel in network")
		}

		msg, err := sioclient.NewMessage(sioclient.MessageTypeEvent, 0, "input", map[string]interface{}{
			"text":   message,
			"target": target.ID,
		})
		if err != nil {
			return errors.Wrap(err, "Unable to compose message")
		}

		if err = msg.Send(client); err != nil {
			return errors.Wrap(err, "Unable to send message")
		}

		interrupt <- os.Interrupt
		return nil
	})
}
