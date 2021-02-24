package container

import (
	"database/sql"
	"prevtorrent/internal/platform/bus/pubsub"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/platform/client/bittorrentproto"
	"prevtorrent/internal/preview/platform/configuration"
	"prevtorrent/internal/preview/platform/storage/file"
	"prevtorrent/internal/preview/platform/storage/inmemory/ffmpeg"
	"prevtorrent/internal/preview/platform/storage/sqlite"
	"prevtorrent/kit/command"

	"github.com/ThreeDotsLabs/watermill"

	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"github.com/anacrolix/torrent"
	"github.com/sirupsen/logrus"
)

type Container struct {
	Config            configuration.Config
	Logger            *logrus.Logger
	MagnetClient      preview.MagnetClient
	TorrentDownloader preview.TorrentDownloader
	TorrentRepo       preview.TorrentRepository
	ImageExtractor    preview.ImageExtractor
	ImagePersister    preview.ImagePersister
	ImageRepository   preview.ImageRepository
	Subscriber        command.Subscriber
}

func NewDefaultContainer() (Container, error) {
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

	imageExtractor, err := ffmpeg.NewInMemoryFfmpeg(logger)
	if err != nil {
		return Container{}, err
	}

	imagePersister := file.NewImagePersister(logger, config.ImageDir)

	sqliteDatabase, err := sql.Open("sqlite3", config.SqlitePath)
	if err != nil {
		return Container{}, err
	}

	torrentRepo := sqlite.NewTorrentRepository(sqliteDatabase)

	imageRepository := sqlite.NewImageRepository(sqliteDatabase)

	torrentClient, err := torrent.NewClient(configuration.GetTorrentConf(config))
	if err != nil {
		return Container{}, err
	}

	torrentIntegration := bittorrentproto.NewTorrentClient(torrentClient, logger)

	loggerWindMill := watermill.NewStdLogger(false, false)
	googleSubscriber, err := googlecloud.NewSubscriber(
		googlecloud.SubscriberConfig{
			GenerateSubscriptionName: googlecloud.TopicSubscriptionName,
			ProjectID:                config.GooglePubSubProjectID,
		},
		loggerWindMill,
	)
	if err != nil {
		return Container{}, err
	}

	subscriber, err := pubsub.NewSubscriber(googleSubscriber)
	if err != nil {
		return Container{}, err
	}

	return Container{
		Config:            config,
		Logger:            logger,
		MagnetClient:      torrentIntegration,
		TorrentDownloader: torrentIntegration,
		TorrentRepo:       torrentRepo,
		ImageExtractor:    imageExtractor,
		ImagePersister:    imagePersister,
		ImageRepository:   imageRepository,
		Subscriber:        subscriber,
	}, nil
}
