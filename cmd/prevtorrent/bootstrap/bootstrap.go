package bootstrap

import (
	"context"
	"errors"
	"github.com/spf13/viper"
	"os"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/transform"
	"prevtorrent/kit/command"
)

const projectName = "prevtorrent"

func Run() error {
	c, err := newContainer()
	if err != nil {
		return err
	}
	bus := makeCommandBus(c)
	return processCommand(bus)
}

// TODO: Move this method to wherever we should put the controllers or commands
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

// TODO: Move this method to wherever we should put the controllers or commands
func trans(bus command.Bus) error {
	if len(os.Args) != 3 {
		return errors.New("invalid arguments. Second parameter must be a magnet")
	}
	magnet := os.Args[2]

	return bus.Dispatch(context.Background(), transform.CMD{
		Magnet: magnet,
	})
}

// TODO: Move this method to wherever we should put the controllers or commands
func download(bus command.Bus) error {
	if len(os.Args) != 3 {
		return errors.New("second parameter must be the path to a valid torrent")
	}
	torr := os.Args[2]
	return bus.Dispatch(context.Background(), downloadPartials.CMD{ID: torr})
}

func getConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/" + projectName)
	viper.AddConfigPath("$HOME/." + projectName)
	viper.AddConfigPath(".")
	viper.SetDefault("TorrentDir", "./tmp/torrents")
	viper.SetDefault("BoltDBDir", "./")
	viper.SetDefault("DownloadsDir", "./tmp/downloads")
	viper.SetDefault("ImageDir", "./tmp/images")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
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
			downloadPartials.NewService(
				c.torrentRepo,
				c.torrentIntegration,
				c.imageExtractor,
				c.imageRepository,
			),
		),
	)

	return commandBus
}
