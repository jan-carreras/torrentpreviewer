package bootstrap

import (
	"os"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/platform/cli"
	"prevtorrent/internal/preview/unmagnetize"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

func Run() error {
	c, err := newContainer()
	if err != nil {
		return err
	}
	bus := makeCommandBus(c)

	c.config.Print(os.Stdout)

	go goroutineLeak(c)
	return cli.Run(bus)
}

func makeCommandBus(c container) *inmemory.SyncCommandBus {
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

func goroutineLeak(c container) {
	// TODO/BUG: It seems that the torrent library, for some reason, leaks goroutines leading to
	//           out of memory error. I think it has to do with the logic that checks the hash of
	//           a chunk, but not sure. It's problematic and I need this here keep an eye on it.
	//           Sigh. FML.
	for {
		goroutines := runtime.NumGoroutine()
		c.logger.WithFields(logrus.Fields{
			"goroutines": goroutines,
		}).Warn("stats")
		time.Sleep(time.Second)
	}
}
