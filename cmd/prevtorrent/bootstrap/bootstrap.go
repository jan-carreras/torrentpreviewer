package bootstrap

import (
	"github.com/spf13/viper"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/platform/cli"
	"prevtorrent/internal/preview/transform"
)

const projectName = "prevtorrent"

func Run() error {
	c, err := newContainer()
	if err != nil {
		return err
	}
	bus := makeCommandBus(c)
	return cli.Run(bus)
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
	viper.SetDefault("SqlitePath", "./prevtorrent.sqlite")
	viper.SetDefault("TorrentDriver", "sqlite")

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
