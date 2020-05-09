package main

import (
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/lounge-control/sioclient"
)

type commandFunc func(args []string) handlerFunc
type handlerFunc func(msg *sioclient.Message) error
type typedHandlerFunc func(mType string, msg *sioclient.Message) error

var (
	commands      = map[string]commandFunc{}
	commandsMutex = new(sync.RWMutex)
)

func availableCommands() (cmds []string) {
	commandsMutex.RLock()
	defer commandsMutex.RUnlock()

	for k := range commands {
		cmds = append(cmds, k)
	}

	sort.Strings(cmds)
	return cmds
}

func registerCommand(cmd string, cf commandFunc) {
	commandsMutex.Lock()
	defer commandsMutex.Unlock()

	if _, ok := commands[cmd]; ok {
		log.Fatalf("Duplicate registration of command %q", cmd)
	}

	commands[cmd] = cf
}
