package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/lounge-control/sioclient"
	"github.com/Luzifer/rconfig/v2"
)

var (
	cfg = struct {
		LogLevel       string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		Network        string `flag:"network,n" description:"Name or UUID of the network to act on"`
		Password       string `flag:"password,p" description:"Password for the given username" validate:"nonzero"`
		SocketURL      string `flag:"socket-url" description:"URL to TheLounge websocket (i.e. 'wss://example.com/socket.io/')" validate:"nonzero"`
		Username       string `flag:"username,u" description:"Username to log into the socket" validate:"nonzero"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	client    *sioclient.Client
	initData  initMessage
	interrupt = make(chan os.Signal, 1)

	version = "dev"
)

func init() {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("lounge-control %s\n", version)
		os.Exit(0)
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	} else {
		log.SetLevel(l)
	}
}

func main() {
	signal.Notify(interrupt, os.Interrupt)

	args := rconfig.Args()[1:]
	if len(args) == 0 {
		log.Fatalf("No command given. Available commands: %s", strings.Join(availableCommands(), ", "))
	}

	commandsMutex.RLock()
	cf, ok := commands[args[0]]
	commandsMutex.RUnlock()
	if !ok {
		log.Fatalf("Unknown command %q. Available commands: %s", args[0], strings.Join(availableCommands(), ", "))
	}

	var err error
	client, err = sioclient.New(sioclient.Config{
		MessageHandler: cf(args[1:]),
		URL:            cfg.SocketURL,
	})
	if err != nil {
		log.WithError(err).Fatal("Unable to connect to server")
	}
	defer client.Close()

	for {
		select {

		case <-interrupt:
			return

		case err := <-client.EIO.Errors():
			log.WithError(err).Error("Error in in command / socket")
			return

		}
	}
}
