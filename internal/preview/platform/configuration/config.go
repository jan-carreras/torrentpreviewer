package configuration

import (
	"encoding/json"
	"fmt"
	"io"
	"prevtorrent/internal/platform/storage/inmemory"

	"github.com/anacrolix/torrent"
	"github.com/spf13/viper"
)

const projectName = "prevtorrent"

type Config struct {
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

func (c Config) Print(w io.Writer) {
	if conf, err := json.MarshalIndent(c, "", "  "); err != nil {
		_, _ = fmt.Fprintf(w, "Error printing the configuration: %v", err)
	} else {
		_, _ = fmt.Fprint(w, "Configuration:")
		_, _ = fmt.Fprint(w, string(conf))
	}
}

func NewConfig() (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AddConfigPath("$HOME/.config/" + projectName)
	viper.AddConfigPath("$HOME/." + projectName)
	viper.AddConfigPath(".")
	viper.AddConfigPath("./testdata")

	viper.SetDefault("ImageDir", "./tmp/images")
	viper.SetDefault("SqlitePath", "./prevtorrent.sqlite")
	viper.SetDefault("EnableIPv6", false)
	viper.SetDefault("EnableUTP", true)
	viper.SetDefault("EnableTorrentDebug", false)
	viper.SetDefault("LogLevel", "debug")
	viper.SetDefault("ConnectionsPerTorrent", "20")
	viper.SetDefault("TorrentListeningPort", "12345")
	viper.SetDefault("TorrentStorageDriver", "inmemory")

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, err
	}

	conf := Config{}

	if err := viper.Unmarshal(&conf); err != nil {
		return Config{}, fmt.Errorf("unable to decode into config struct, %v", err)
	}

	return conf, nil
}

func GetTorrentConf(config Config) *torrent.ClientConfig {
	c := torrent.NewDefaultClientConfig()

	switch driver := config.TorrentStorageDriver; driver {
	case "inmemory":
		c.DefaultStorage = inmemory.NewTorrentStorage()
	case "file":
	default:
		panic(fmt.Errorf("unknown storage driver %v", driver))
	}

	c.DisableIPv6 = !config.EnableIPv6
	c.DisableUTP = !config.EnableUTP
	c.Debug = config.EnableTorrentDebug
	c.EstablishedConnsPerTorrent = config.ConnectionsPerTorrent
	c.ListenPort = config.TorrentListeningPort
	c.NoDefaultPortForwarding = false
	c.Seed = true
	return c
}
