package bootstrap

import (
	"database/sql"
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
	magnetClient      preview.MagnetClient
	torrentDownloader preview.TorrentDownloader
	torrentRepo       preview.TorrentRepository
	logger            *logrus.Logger
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
	logger.Level = logrus.DebugLevel

	imageExtractor := ffmpeg.NewInMemoryFfmpeg(logger)

	imagePersister := file.NewImagePersister(logger, config.ImageDir)

	sqliteDatabase, err := sql.Open("sqlite3", config.SqlitePath)
	if err != nil {
		return container{}, err
	}

	torrentRepo := sqlite.NewTorrentRepository(sqliteDatabase)

	imageRepository := sqlite.NewImageRepository(sqliteDatabase)

	conf := torrent.NewDefaultClientConfig()
	conf.DefaultStorage = inmemory.NewTorrentStorage()
	conf.DisableIPv6 = !config.EnableIPv6
	conf.DisableUTP = !config.EnableUTP

	torrentClient, err := torrent.NewClient(conf)
	if err != nil {
		return container{}, err
	}

	torrentIntegration := bittorrentproto.NewTorrentClient(torrentClient, logger)

	return container{
		magnetClient:      torrentIntegration,
		torrentDownloader: torrentIntegration,
		torrentRepo:       torrentRepo,
		logger:            logger,
		imageExtractor:    imageExtractor,
		imagePersister:    imagePersister,
		imageRepository:   imageRepository,
	}, nil
}
