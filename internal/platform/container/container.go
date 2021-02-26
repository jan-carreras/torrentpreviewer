package container

import (
	"database/sql"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/platform/client/bittorrentproto"
	"prevtorrent/internal/preview/platform/configuration"
	"prevtorrent/internal/preview/platform/storage/file"
	"prevtorrent/internal/preview/platform/storage/inmemory/ffmpeg"
	"prevtorrent/internal/preview/platform/storage/sqlite"
	"prevtorrent/internal/preview/unmagnetize"

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
	loggerWatermill    watermill.LoggerAdapter
	cqrs               *cqrs.Facade
	bus                *cqrs.CommandBus
	publisher          message.Publisher
	messageSubscriber  message.Subscriber
	cqrsRouter         *message.Router
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

	loggerWatermill := watermill.NewStdLogger(false, false)

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
		loggerWatermill: loggerWatermill,
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

func (c *Container) CommandSubscriber() message.Subscriber {
	if c.messageSubscriber == nil {
		googleSubscriber, err := googlecloud.NewSubscriber(
			googlecloud.SubscriberConfig{
				GenerateSubscriptionName: googlecloud.TopicSubscriptionName,
				ProjectID:                c.Config.GooglePubSubProjectID,
			},
			c.loggerWatermill,
		)
		if err != nil {
			panic(err)
		}
		c.messageSubscriber = googleSubscriber
	}

	return c.messageSubscriber
}

func (c *Container) EventSubscriber() message.Subscriber {
	if c.messageSubscriber == nil {
		googleSubscriber, err := googlecloud.NewSubscriber(
			googlecloud.SubscriberConfig{
				GenerateSubscriptionName: googlecloud.TopicSubscriptionName,
				ProjectID:                c.Config.GooglePubSubProjectID,
			},
			c.loggerWatermill,
		)
		if err != nil {
			panic(err)
		}
		c.messageSubscriber = googleSubscriber
	}

	return c.messageSubscriber
}

func (c *Container) CommandPublisher() message.Publisher {
	if c.publisher == nil {
		publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{
			ProjectID: c.Config.GooglePubSubProjectID,
		}, c.loggerWatermill)
		if err != nil {
			panic(err)
		}
		c.publisher = publisher
	}

	return c.publisher
}

func (c *Container) EventPublisher() message.Publisher {
	if c.publisher == nil {
		publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{
			ProjectID: c.Config.GooglePubSubProjectID,
		}, c.loggerWatermill)
		if err != nil {
			panic(err)
		}
		c.publisher = publisher
	}

	return c.publisher
}

func (c *Container) CQRS() *cqrs.Facade {
	if c.cqrs != nil {
		return c.cqrs
	}

	logger := watermill.NewStdLogger(false, false)

	router, err := message.NewRouter(message.RouterConfig{}, c.loggerWatermill)
	if err != nil {
		panic(err)
	}
	router.AddMiddleware(middleware.Recoverer)

	c.cqrsRouter = router

	cqrsFacade, err := cqrs.NewFacade(cqrs.FacadeConfig{
		GenerateCommandsTopic: func(commandName string) string {
			return commandName
		},
		CommandHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.CommandHandler {
			return []cqrs.CommandHandler{
				unmagnetize.NewCommandHandler(eb, unmagnetize.NewService(c.Logger, eb, c.MagnetClient(), c.TorrentRepo)),
			}
		},
		CommandsPublisher: c.CommandPublisher(),
		CommandsSubscriberConstructor: func(handlerName string) (message.Subscriber, error) {
			return c.CommandSubscriber(), nil
		},
		GenerateEventsTopic: func(eventName string) string {
			return eventName
		},
		EventHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.EventHandler {
			return []cqrs.EventHandler{
				downloadPartials.NewTorrentCreatedEventHandler(eb, downloadPartials.NewService(
					c.Logger,
					c.TorrentRepo,
					c.TorrentDownloader(),
					c.ImageExtractor(),
					c.ImagePersister,
					c.ImageRepository),
				),
			}
		},
		EventsPublisher: c.EventPublisher(),
		EventsSubscriberConstructor: func(handlerName string) (message.Subscriber, error) {
			return c.EventSubscriber(), nil
		},
		Router:                router,
		CommandEventMarshaler: cqrs.JSONMarshaler{},
		Logger:                logger,
	})
	if err != nil {
		panic(err)
	}

	c.cqrs = cqrsFacade
	return c.cqrs
}

func (c *Container) CQRSRouter() *message.Router {
	if c.cqrsRouter != nil {
		return c.cqrsRouter
	}

	_ = c.CQRS() // It creates the router. A router without bindings is useless.

	return c.cqrsRouter
}
