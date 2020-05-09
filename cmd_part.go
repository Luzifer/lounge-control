package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Luzifer/lounge-control/sioclient"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func init() {
	registerCommand("part", commandPart)
}

func commandPart(args []string) handlerFunc {
	if len(args) == 0 {
		log.Fatal("No channels given to part")
	}

	return addGenericHandler(func(pType string, msg *sioclient.Message) error {
		if pType != "init" {
			return nil
		}

		// After join command is finished we can execute the joins
		network := initData.NetworkByNameOrUUID(cfg.Network)
		if network == nil {
			return errors.New("Network not found")
		}

		var lobby *channel
		for _, c := range network.Channels {
			if c.Type == "lobby" {
				lobby = &c
				break
			}
		}

		if lobby == nil {
			return errors.New("Unable to find lobby for network")
		}

		for _, ch := range args {
			if !strings.HasPrefix(ch, "#") {
				ch = "#" + ch
			}

			msg, err := sioclient.NewMessage(sioclient.MessageTypeEvent, 0, "input", map[string]interface{}{
				"text":   fmt.Sprintf("/part %s", ch),
				"target": lobby.ID,
			})
			if err != nil {
				return errors.Wrap(err, "Unable to compose part message")
			}

			if err = msg.Send(client); err != nil {
				return errors.Wrap(err, "Unable to send part message")
			}
		}

		interrupt <- os.Interrupt
		return nil
	})
}
