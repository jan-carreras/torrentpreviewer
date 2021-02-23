package bootstrap

import (
	"database/sql"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/platform/client/bittorrentproto"
	"prevtorrent/internal/preview/platform/configuration"
	"prevtorrent/internal/preview/platform/storage/file"
	"prevtorrent/internal/preview/platform/storage/inmemory/ffmpeg"
	"prevtorrent/internal/preview/platform/storage/sqlite"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/sirupsen/logrus"
)

type container struct {
	config            configuration.Config
	logger            *logrus.Logger
	magnetClient      preview.MagnetClient
	torrentDownloader preview.TorrentDownloader
	torrentRepo       preview.TorrentRepository
	imageExtractor    preview.ImageExtractor
	imagePersister    preview.ImagePersister
	imageRepository   preview.ImageRepository
	subscriber        *googlecloud.Subscriber
}

func newContainer() (container, error) {
	config, err := configuration.NewConfig()
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

	imageExtractor, err := ffmpeg.NewInMemoryFfmpeg(logger)
	if err != nil {
		return container{}, err
	}

	imagePersister := file.NewImagePersister(logger, config.ImageDir)

	sqliteDatabase, err := sql.Open("sqlite3", config.SqlitePath)
	if err != nil {
		return container{}, err
	}

	torrentRepo := sqlite.NewTorrentRepository(sqliteDatabase)

	imageRepository := sqlite.NewImageRepository(sqliteDatabase)

	torrentClient, err := torrent.NewClient(configuration.GetTorrentConf(config))
	if err != nil {
		return container{}, err
	}

	torrentIntegration := bittorrentproto.NewTorrentClient(torrentClient, logger)

	loggerWatermill := watermill.NewStdLogger(false, false)
	subscriber, err := googlecloud.NewSubscriber(
		googlecloud.SubscriberConfig{
			GenerateSubscriptionName: googlecloud.TopicSubscriptionName,
			ProjectID:                "torrentpreview",
			InitializeTimeout:        time.Second * 10,
			ClientOptions:            nil,
		},
		loggerWatermill,
	)
	if err != nil {
		return container{}, err
	}

	return container{
		config:            config,
		logger:            logger,
		magnetClient:      torrentIntegration,
		torrentDownloader: torrentIntegration,
		torrentRepo:       torrentRepo,
		imageExtractor:    imageExtractor,
		imagePersister:    imagePersister,
		imageRepository:   imageRepository,
		subscriber:        subscriber, // TODO: I need my own abstraction, I don't like this
	}, nil
}
