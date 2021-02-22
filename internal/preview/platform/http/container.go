package http

import (
	"database/sql"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/getTorrent"
	"prevtorrent/internal/preview/importTorrent"
	"prevtorrent/internal/preview/platform/client/bittorrentproto"
	"prevtorrent/internal/preview/platform/configuration"
	"prevtorrent/internal/preview/platform/storage/sqlite"
	"prevtorrent/internal/preview/unmagnetize"

	"github.com/anacrolix/torrent"
	"github.com/sirupsen/logrus"
)

const projectName = "prevtorrent"

type Services struct {
	GetTorrent    *getTorrent.Service // TODO: Remove pointer
	Unmagnetize   unmagnetize.Service
	ImportTorrent importTorrent.Service
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

	importTorrentService := importTorrent.NewService(logger, torrentIntegration, torrentRepo)

	return Container{
		Config: config,
		Logger: logger,
		repositories: Repositories{ // TODO: This can be deleted, it serves no purpose
			torrent: torrentRepo,
			image:   imageRepository,
		},
		services: Services{
			GetTorrent:    getTorrentService,
			Unmagnetize:   unmagnetizeService,
			ImportTorrent: importTorrentService,
		},
	}, nil
}
