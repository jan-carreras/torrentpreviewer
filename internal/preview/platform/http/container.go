package http

import (
	"database/sql"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/getTorrent"
	"prevtorrent/internal/preview/platform/client/bittorrentproto"
	"prevtorrent/internal/preview/platform/configuration"
	"prevtorrent/internal/preview/platform/storage/sqlite"
	"prevtorrent/internal/preview/unmagnetize"
)

const projectName = "prevtorrent"

type Services struct {
	GetTorrent  *getTorrent.Service // TODO: Remove pointer
	Unmagnetize unmagnetize.Service
}

type Repositories struct {
	torrent preview.TorrentRepository
	image   preview.ImageRepository
}

type Container struct {
	Config       configuration.Config
	Logger       *logrus.Logger
	repositories Repositories
	services     Services
}

type config struct {
	ImageDir   string `yaml:"ImageDir"`
	SqlitePath string `yaml:"SqlitePath"`
	LogLevel   string `yaml:"LogLevel"`
}

func getConfig() (config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AddConfigPath("$HOME/.config/" + projectName)
	viper.AddConfigPath("$HOME/." + projectName)
	viper.AddConfigPath(".")

	viper.SetDefault("ImageDir", "./tmp/images")
	viper.SetDefault("SqlitePath", "./prevtorrent.sqlite")
	viper.SetDefault("LogLevel", "warning")

	if err := viper.ReadInConfig(); err != nil {
		return config{}, err
	}

	conf := config{}
	if err := viper.Unmarshal(&conf); err != nil {
		return config{}, fmt.Errorf("unable to decode into config struct, %v", err)
	}

	return conf, nil
}

func NewContainer() (Container, error) {
	config, err := configuration.NewConfig()
	if err != nil {
		return Container{}, err
	}

	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{}

	logLevel, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return Container{}, err
	}
	logger.Level = logLevel

	sqliteDatabase, err := sql.Open("sqlite3", config.SqlitePath)
	if err != nil {
		return Container{}, err
	}

	torrentRepo := sqlite.NewTorrentRepository(sqliteDatabase)
	imageRepository := sqlite.NewImageRepository(sqliteDatabase)

	getTorrentService := getTorrent.NewService(logger, torrentRepo, imageRepository)

	torrentClient, err := torrent.NewClient(configuration.GetTorrentConf(config))
	if err != nil {
		return Container{}, err
	}

	torrentIntegration := bittorrentproto.NewTorrentClient(torrentClient, logger)

	unmagnetizeService := unmagnetize.NewService(logger, torrentIntegration, torrentRepo)

	return Container{
		Config: config,
		Logger: logger,
		repositories: Repositories{
			torrent: torrentRepo,
			image:   imageRepository,
		},
		services: Services{
			GetTorrent:  getTorrentService,
			Unmagnetize: unmagnetizeService,
		},
	}, nil
}
