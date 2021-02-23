package bootstrap

import (
	"os"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/platform/cli"
	"prevtorrent/internal/preview/unmagnetize"
	"prevtorrent/kit/command"
)

func Run() error {
	c, err := newContainer()
	if err != nil {
		return err
	}
	bus := makeCommandBus(c)

	c.config.Print(os.Stdout)

	return cli.Run(bus)
}

func makeCommandBus(c container) command.Bus {
	commandBus := inmemory.NewSyncCommandBus(c.logger)

	commandBus.Register(
		unmagnetize.CommandType,
		unmagnetize.NewTransformCommandHandler(
			unmagnetize.NewService(
				c.logger,
				c.magnetClient,
				c.torrentRepo,
			),
		),
	)

	commandBus.Register(
		downloadPartials.CommandType,
		downloadPartials.NewCommandHandler(
			downloadPartials.NewService(
				c.logger,
				c.torrentRepo,
				c.torrentDownloader,
				c.imageExtractor,
				c.imagePersister,
				c.imageRepository,
			),
		),
	)

	return commandBus
}
