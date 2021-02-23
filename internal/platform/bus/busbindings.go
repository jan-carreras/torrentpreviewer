package bus

import (
	"prevtorrent/internal/platform/container"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/unmagnetize"
	"prevtorrent/kit/command"
)

func MakeBindings(commandBus command.Bus, c container.Container) {
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
