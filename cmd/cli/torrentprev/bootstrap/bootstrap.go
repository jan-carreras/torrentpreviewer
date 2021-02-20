package bootstrap

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"prevtorrent/internal/platform/bus/inmemory"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/platform/cli"
	"prevtorrent/internal/preview/transform"
	"runtime"
	"time"

	"github.com/spf13/viper"
)

const projectName = "prevtorrent"

func Run() error {
	c, err := newContainer()
	if err != nil {
		return err
	}
	bus := makeCommandBus(c)

	if conf, err := json.MarshalIndent(c.config, "", "  "); err != nil {
		return err
	} else {
		fmt.Println("Configuration:")
		fmt.Println(string(conf))
	}

	go goroutineLeak(c)
	return cli.Run(bus)
}

type config struct {
	ImageDir              string `yaml:"ImageDir"`
	SqlitePath            string `yaml:"SqlitePath"`
	EnableIPv6            bool   `yaml:"EnableIPv6"`
	EnableUTP             bool   `yaml:"EnableUTP"`
	EnableTorrentDebug    bool   `yaml:"EnableTorrentDebug"`
	LogLevel              string `yaml:"LogLevel"`
	ConnectionsPerTorrent int    `yaml:"ConnectionsPerTorrent"`
	TorrentListeningPort  int    `yaml:"TorrentListeningPort"`
	TorrentStorageDriver  string `yaml:"TorrentStorageDriver"`
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
	viper.SetDefault("EnableTorrentDebug", false)
	viper.SetDefault("LogLevel", "warning")
	viper.SetDefault("ConnectionsPerTorrent", "20")
	viper.SetDefault("TorrentListeningPort", "12345")
	viper.SetDefault("TorrentStorageDriver", "inmemory")

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
