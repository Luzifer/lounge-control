package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/lounge-control/sioclient"
)

func init() {
	registerCommand("sync-twitch-follows", commandSyncTwitchFollows)
}

func commandSyncTwitchFollows(args []string) handlerFunc {
	channelAct := func(lobbyID int, action, twitchName string) error {
		msg, err := sioclient.NewMessage(sioclient.MessageTypeEvent, 0, "input", map[string]interface{}{
			"text":   fmt.Sprintf("/%s #%s", action, twitchName),
			"target": lobbyID,
		})
		if err != nil {
			return errors.Wrap(err, "Unable to compose join message")
		}

		if err = msg.Send(client); err != nil {
			return errors.Wrap(err, "Unable to send join message")
		}

		// Twitch limits the number of actions, so we need an arbitrary delay
		time.Sleep(750 * time.Millisecond)
		return nil
	}

	return addGenericHandler(func(pType string, msg *sioclient.Message) error {
		if pType != "init" {
			return nil
		}

		network := initData.NetworkByNameOrUUID(cfg.Network)
		if network == nil {
			return errors.New("Network not found")
		}

		// Find lobby to send commands to
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

		// Get configured nickname (must match Twitch nick)
		var user = network.Nick
		log.WithField("username", user).Info("Synchronizing with twitch user")

		// Convert username into user ID
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.twitch.tv/kraken/users?login=%s", user), nil)
		req.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
		req.Header.Set("Client-ID", twitchClientID)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrap(err, "Unable to get user ID for Twitch user")
		}
		defer resp.Body.Close()

		var respObjUsers struct {
			Users []struct {
				ID string `json:"_id"`
			} `json:"users"`
		}
		if err = json.NewDecoder(resp.Body).Decode(&respObjUsers); err != nil {
			return errors.Wrap(err, "Unable to read Twitch response")
		}

		if l := len(respObjUsers.Users); l != 1 {
			return errors.Errorf("Received invalid number of user IDs: %d", l)
		}

		var userID = respObjUsers.Users[0].ID

		// Retrieve follows
		req, _ = http.NewRequest("GET", fmt.Sprintf("https://api.twitch.tv/kraken/users/%s/follows/channels?limit=100", userID), nil)
		req.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
		req.Header.Set("Client-ID", twitchClientID)

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrap(err, "Unable to get follows for Twitch user")
		}
		defer resp.Body.Close()

		var respObjFollows struct {
			Follows []struct {
				Channel struct {
					Name string `json:"name"`
				} `json:"channel"`
			} `json:"follows"`
		}
		if err = json.NewDecoder(resp.Body).Decode(&respObjFollows); err != nil {
			return errors.Wrap(err, "Unable to read Twitch response")
		}

		// Compare channel list and act on them
		var (
			expectedChannels = []string{user}
			presentChannels  []string
		)

		for _, c := range network.Channels {
			if c.Type != "channel" {
				continue
			}
			presentChannels = append(presentChannels, strings.TrimPrefix(c.Name, "#"))
		}

		for _, f := range respObjFollows.Follows {
			expectedChannels = append(expectedChannels, f.Channel.Name)
		}

		// Join new channels
		for _, cn := range expectedChannels {
			if str.StringInSlice(cn, presentChannels) {
				continue
			}
			log.WithField("channel", cn).Info("Joining new channel")
			if err = channelAct(lobby.ID, "join", cn); err != nil {
				return errors.Wrap(err, "Unable to execute channel action")
			}
		}

		// Leave unexpected channels
		for _, cn := range presentChannels {
			if str.StringInSlice(cn, expectedChannels) {
				log.WithField("channel", cn).Debug("Retaining channel")
				continue
			}
			log.WithField("channel", cn).Info("Leaving channel")
			if err = channelAct(lobby.ID, "part", cn); err != nil {
				return errors.Wrap(err, "Unable to execute channel action")
			}
		}

		interrupt <- os.Interrupt
		return nil
	})
}
