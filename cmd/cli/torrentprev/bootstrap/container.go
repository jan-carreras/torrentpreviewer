package bootstrap

import (
	"database/sql"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/sirupsen/logrus"
	"prevtorrent/internal/platform/storage/inmemory"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/platform/client/bittorrentproto"
	"prevtorrent/internal/preview/platform/storage/file"
	"prevtorrent/internal/preview/platform/storage/inmemory/ffmpeg"
	"prevtorrent/internal/preview/platform/storage/sqlite"
)

type container struct {
	config            config
	logger            *logrus.Logger
	magnetClient      preview.MagnetClient
	torrentDownloader preview.TorrentDownloader
	torrentRepo       preview.TorrentRepository
	imageExtractor    preview.ImageExtractor
	imagePersister    preview.ImagePersister
	imageRepository   preview.ImageRepository
}

func newContainer() (container, error) {
	config, err := getConfig()
	if err != nil {
		return container{}, err
	}

	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{}

	logLevel, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return container{}, err
	}
	logger.Level = logLevel

	imageExtractor := ffmpeg.NewInMemoryFfmpeg(logger)

	imagePersister := file.NewImagePersister(logger, config.ImageDir)

	sqliteDatabase, err := sql.Open("sqlite3", config.SqlitePath)
	if err != nil {
		return container{}, err
	}

	torrentRepo := sqlite.NewTorrentRepository(sqliteDatabase)

	imageRepository := sqlite.NewImageRepository(sqliteDatabase)

	torrentClient, err := torrent.NewClient(getTorrentConf(config))
	if err != nil {
		return container{}, err
	}

	torrentIntegration := bittorrentproto.NewTorrentClient(torrentClient, logger)

	return container{
		config:            config,
		logger:            logger,
		magnetClient:      torrentIntegration,
		torrentDownloader: torrentIntegration,
		torrentRepo:       torrentRepo,
		imageExtractor:    imageExtractor,
		imagePersister:    imagePersister,
		imageRepository:   imageRepository,
	}, nil
}

func getTorrentConf(config config) *torrent.ClientConfig {
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
