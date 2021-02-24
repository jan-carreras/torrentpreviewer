package bus

import (
	"fmt"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/platform/bus/pubsub"
	"prevtorrent/internal/platform/container"
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
		cb = pubsub.NewPubSubCommandBus(c.Logger, nil) // TODO: We need to be able to pass the publisher here!
	} else {
		return cb, fmt.Errorf("unknown command bus type: %v", cbType)
	}

	makeBindings(cb, c)

	return cb, nil
}

func makeBindings(commandBus command.Bus, c container.Container) {
	commandBus.Register(
		unmagnetize.CommandType,
		unmagnetize.NewTransformCommandHandler(
			unmagnetize.NewService(
				c.Logger,
				c.MagnetClient,
				c.TorrentRepo,
			),
		),
	)

	commandBus.Register(
		downloadPartials.CommandType,
		downloadPartials.NewCommandHandler(
			downloadPartials.NewService(
				c.Logger,
				c.TorrentRepo,
				c.TorrentDownloader,
				c.ImageExtractor,
				c.ImagePersister,
				c.ImageRepository,
			),
		),
	)
}
