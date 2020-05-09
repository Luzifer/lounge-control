package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Luzifer/lounge-control/sioclient"
)

func init() {
	registerCommand("list-channels", commandListChannels)
}

func commandListChannels(args []string) handlerFunc {
	return addGenericHandler(func(pType string, msg *sioclient.Message) error {
		if pType != "init" {
			return nil
		}

		network := initData.NetworkByNameOrUUID(cfg.Network)
		if network == nil {
			return errors.New("Network not found")
		}

		var channels []string

		for _, c := range network.Channels {
			if c.Type == "lobby" {
				continue
			}

			channels = append(channels, c.Name)
		}

		sort.Strings(channels)

		fmt.Println(strings.Join(channels, "\n"))
		interrupt <- os.Interrupt
		return nil
	})
}
