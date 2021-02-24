package bus

import (
	"fmt"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/platform/bus/pubsub"
	"prevtorrent/internal/platform/container"
	"prevtorrent/internal/platform/services"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/unmagnetize"
	"prevtorrent/kit/command"
)

var Sync = "sync"
var PubSub = "pubsub"

func MakeCommandBus(cbType string, c container.Container) (cb command.Bus, err error) {
	if cbType == Sync {
		cb = inmemory.NewSyncCommandBus(c.Logger)
	} else if cbType == PubSub {
		cb = pubsub.NewPubSubCommandBus(c.Logger, c.Publisher)
	} else {
		return cb, fmt.Errorf("unknown command bus type: %v", cbType)
	}

	srv, err := services.NewServices(c)
	if err != nil {
		return nil, err
	}

	makeBindings(cb, srv)

	return cb, nil
}

func makeBindings(bus command.Bus, s services.Services) {
	bus.Register(unmagnetize.CommandType, unmagnetize.NewTransformCommandHandler(s.Unmagnetize()))
	bus.Register(downloadPartials.CommandType, downloadPartials.NewCommandHandler(s.DownloadPartials()))
}
