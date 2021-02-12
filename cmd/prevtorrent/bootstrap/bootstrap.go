package bootstrap

import (
	"context"
	"errors"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"github.com/sirupsen/logrus"
	"os"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/platform/client/bittorrentproto"
	"prevtorrent/internal/preview/platform/storage/file"
	"prevtorrent/internal/preview/transform"
	"prevtorrent/kit/command"
)

func Run() error {
	c, err := newContainer()
	if err != nil {
		return err
	}
	bus := makeCommandBus(c)
	return processCommand(bus)
}

func processCommand(bus command.Bus) error {
	action := "unknown"
	if len(os.Args) >= 2 {
		action = os.Args[1]
	}

	if action == "download" {
		return download(bus)
	} else if action == "transform" {
		return trans(bus)
	} else {
		return errors.New("unknown command")
	}
}

func trans(bus command.Bus) error {
	if len(os.Args) != 3 {
		return errors.New("invalid arguments. Second parameter must be a magnet")
	}
	magnet := os.Args[2]

	return bus.Dispatch(context.Background(), transform.CMD{
		Magnet: magnet,
	})
}

func download(bus command.Bus) error {
	return bus.Dispatch(context.Background(), downloadPartials.CMD{
		ID: "/tmp/Star Wars Episode V The Empire Strikes Back (1980) [1080p].torrent",
	})
}

func newContainer() (container, error) {
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}

	conf := torrent.NewDefaultClientConfig()
	conf.DefaultStorage = storage.NewBoltDB("/Users/jan/Documents/projects/langs/go/src/prevtorrent")

	torrentClient, err := torrent.NewClient(conf)
	if err != nil {
		return container{}, err
	}

	torrentIntegration := bittorrentproto.NewTorrentClient(torrentClient, logger)
	torrentRepo := file.NewTorrentRepository(logger)

	return container{
		torrentIntegration: torrentIntegration,
		torrentRepo:        torrentRepo,
		logger:             logger,
	}, nil
}

func makeCommandBus(c container) *inmemory.SyncCommandBus {
	commandBus := inmemory.NewSyncCommandBus(c.logger)

	commandBus.Register(
		transform.CommandType,
		transform.NewTransformCommandHandler(
			transform.NewService(
				c.torrentIntegration,
				c.torrentRepo,
			),
		),
	)

	commandBus.Register(
		downloadPartials.CommandType,
		downloadPartials.NewCommandHandler(
			downloadPartials.NewService(c.torrentRepo, c.torrentIntegration),
		),
	)

	return commandBus
}

type container struct {
	torrentIntegration preview.MagnetClient
	torrentRepo        preview.TorrentRepository
	logger             *logrus.Logger
}
