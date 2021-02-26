package container

import (
	"database/sql"
	"fmt"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"prevtorrent/internal/platform/bus/pubsub"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/platform/client/bittorrentproto"
	"prevtorrent/internal/preview/platform/configuration"
	"prevtorrent/internal/preview/platform/storage/file"
	"prevtorrent/internal/preview/platform/storage/inmemory/ffmpeg"
	"prevtorrent/internal/preview/platform/storage/sqlite"
	"prevtorrent/internal/preview/unmagnetize"
	"prevtorrent/kit/command"

	"github.com/ThreeDotsLabs/watermill/message"

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
	loggerWatermill    watermill.LoggerAdapter
	bus                *cqrs.CommandBus
	messageSubscriber  message.Subscriber
	cqrs               *cqrs.Facade
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

func (c *Container) Subscriber() command.Subscriber {
	if c.subscriber == nil {
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

		subscriber, err := pubsub.NewSubscriber(googleSubscriber)
		if err != nil {
			panic(err)
		}
		c.subscriber = subscriber
	}

	return c.subscriber
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

func (c *Container) Bus() *cqrs.CommandBus {
	if c.bus == nil {
		generateTopic := func(commandName string) string {
			return fmt.Sprintf("%v-topic", commandName)
		}
		bus, err := cqrs.NewCommandBus(c.CommandPublisher(), generateTopic, cqrs.JSONMarshaler{})
		if err != nil {
			panic(err)
		}
		c.bus = bus
	}
	return c.bus
}

func (c *Container) CQRS() *cqrs.Facade {
	if c.cqrs != nil {
		return c.cqrs
	}

	logger := watermill.NewStdLogger(false, false)
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}
	router.AddMiddleware(middleware.Recoverer)

	cqrsMarshaler := cqrs.JSONMarshaler{}

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
		/*EventHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.EventHandler {
			return []cqrs.EventHandler{
				createTorrent.NewUnmagnetizeEventHandler(createTorrent.NewService(c.Logger, c.TorrentRepo)), // TODO: Move to another place
			}
		},*/
		EventsPublisher: c.EventPublisher(),
		EventsSubscriberConstructor: func(handlerName string) (message.Subscriber, error) {
			return c.EventSubscriber(), nil
		},
		Router:                router,
		CommandEventMarshaler: cqrsMarshaler,
		Logger:                logger,
	})
	if err != nil {
		panic(err)
	}

	c.cqrs = cqrsFacade
	return c.cqrs
}
