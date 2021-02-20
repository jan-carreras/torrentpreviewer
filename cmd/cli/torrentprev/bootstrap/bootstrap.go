package bootstrap

import (
	"fmt"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/platform/cli"
	"prevtorrent/internal/preview/transform"

	"github.com/spf13/viper"
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

type config struct {
	ImageDir   string `yaml:"ImageDir"`
	SqlitePath string `yaml:"SqlitePath"`
	EnableIPv6 bool   `yaml:"EnableIPv6"`
	EnableUTP  bool   `yaml:"EnableUTP"`
}

func getConfig() (config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/" + projectName)
	viper.AddConfigPath("$HOME/." + projectName)
	viper.AddConfigPath(".")
	viper.SetDefault("ImageDir", "./tmp/images")
	viper.SetDefault("SqlitePath", "./prevtorrent.sqlite")
	viper.SetDefault("EnableIPv6", false)
	viper.SetDefault("EnableUTP", true)

	if err := viper.ReadInConfig(); err != nil {
		return config{}, err
	}

	conf := config{}

	if err := viper.Unmarshal(&conf); err != nil {
		return config{}, fmt.Errorf("unable to decode into config struct, %v", err)
	}

	return conf, nil
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
