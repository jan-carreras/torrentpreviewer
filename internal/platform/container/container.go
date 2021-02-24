package container

import (
	"database/sql"
	"github.com/ThreeDotsLabs/watermill/message"
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
	Config             configuration.Config
	Logger             *logrus.Logger
	torrentIntegration *bittorrentproto.TorrentClient
	TorrentRepo        preview.TorrentRepository
	imageExtractor     preview.ImageExtractor
	ImagePersister     preview.ImagePersister
	ImageRepository    preview.ImageRepository
	subscriber         command.Subscriber
	publisher          message.Publisher
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

	imagePersister := file.NewImagePersister(logger, config.ImageDir)

	sqliteDatabase, err := sql.Open("sqlite3", config.SqlitePath)
	if err != nil {
		return Container{}, err
	}

	torrentRepo := sqlite.NewTorrentRepository(sqliteDatabase)
	imageRepository := sqlite.NewImageRepository(sqliteDatabase)

	return Container{
		Config:          config,
		Logger:          logger,
		TorrentRepo:     torrentRepo,
		ImagePersister:  imagePersister,
		ImageRepository: imageRepository,
	}, nil
}

func (c *Container) ImageExtractor() preview.ImageExtractor {
	if c.imageExtractor == nil {
		imageExtractor, err := ffmpeg.NewInMemoryFfmpeg(c.Logger)
		if err != nil {
			logrus.Fatal(err)
		}
		c.imageExtractor = imageExtractor
	}
	return c.imageExtractor
}

func (c *Container) getTorrentIntegration() *bittorrentproto.TorrentClient {
	if c.torrentIntegration == nil {
		torrentClient, err := torrent.NewClient(configuration.GetTorrentConf(c.Config))
		if err != nil {
			panic(err)
		}
		c.torrentIntegration = bittorrentproto.NewTorrentClient(torrentClient, c.Logger)
	}
	return c.torrentIntegration
}

func (c *Container) MagnetClient() preview.MagnetClient {
	return c.getTorrentIntegration()
}

func (c *Container) TorrentDownloader() preview.TorrentDownloader {
	return c.getTorrentIntegration()
}

func (c *Container) Subscriber() command.Subscriber {
	if c.subscriber == nil {
		loggerWindMill := watermill.NewStdLogger(false, false)
		googleSubscriber, err := googlecloud.NewSubscriber(
			googlecloud.SubscriberConfig{
				GenerateSubscriptionName: googlecloud.TopicSubscriptionName,
				ProjectID:                c.Config.GooglePubSubProjectID,
			},
			loggerWindMill,
		)
		if err != nil {
			panic(err)
		}

		subscriber, err := pubsub.NewSubscriber(googleSubscriber)
		if err != nil {
			panic(err)
		}
		c.subscriber = subscriber
	}

	return c.subscriber
}

func (c *Container) Publisher() message.Publisher {
	if c.publisher == nil {
		loggerWindMill := watermill.NewStdLogger(false, false) // TODO: We are creating this twice...
		publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{
			ProjectID: c.Config.GooglePubSubProjectID,
		}, loggerWindMill)
		if err != nil {
			panic(err)
		}
		c.publisher = publisher
	}

	return c.publisher
}
