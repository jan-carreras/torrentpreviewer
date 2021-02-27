package container

import (
	"database/sql"
	"prevtorrent/internal/platform/bus"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/platform/client/bittorrentproto"
	"prevtorrent/internal/preview/platform/configuration"
	"prevtorrent/internal/preview/platform/storage/file"
	"prevtorrent/internal/preview/platform/storage/inmemory/ffmpeg"
	"prevtorrent/internal/preview/platform/storage/sqlite"
	"prevtorrent/internal/preview/unmagnetize"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"

	"github.com/ThreeDotsLabs/watermill"

	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"github.com/anacrolix/torrent"
	"github.com/sirupsen/logrus"
)

//go:generate mockery --case=snake --outpkg=containermocks --output=containermocks --name=Container
type Container interface {
	Config() configuration.Config
	Logger() *logrus.Logger
	ImagePersister() preview.ImagePersister
	ImageExtractor() preview.ImageExtractor
	MagnetClient() preview.MagnetClient
	TorrentDownloader() preview.TorrentDownloader
	CommandBus() bus.Command
	EventBus() bus.Event
	TorrentRepository() preview.TorrentRepository
	ImageRepository() preview.ImageRepository
}

type repositories struct {
	torrent preview.TorrentRepository
	image   preview.ImageRepository
}

type eventSourcing struct {
	cqrsFacade        *cqrs.Facade
	publisher         message.Publisher
	messageSubscriber message.Subscriber
	cqrsRouter        *message.Router
}

type container struct {
	config             configuration.Config
	logger             *logrus.Logger
	torrentIntegration *bittorrentproto.TorrentClient
	imageExtractor     preview.ImageExtractor
	imagePersister     preview.ImagePersister
	repositories       repositories
	loggerWatermill    watermill.LoggerAdapter
	eventSourcing      eventSourcing

	db *sql.DB
}

type testingContainer struct {
	*container
}

func NewDefaultContainer() (*container, error) {
	config, err := configuration.NewConfig()
	if err != nil {
		return nil, err
	}

	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{}

	logLevel, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, err
	}
	logger.Level = logLevel

	loggerWatermill := watermill.NewStdLogger(false, false)

	imagePersister := file.NewImagePersister(logger, config.ImageDir)

	sqliteDatabase, err := sql.Open("sqlite3", config.SqlitePath)
	if err != nil {
		return nil, err
	}

	torrentRepo := sqlite.NewTorrentRepository(sqliteDatabase)
	imageRepository := sqlite.NewImageRepository(sqliteDatabase)

	return &container{
		config:          config,
		logger:          logger,
		loggerWatermill: loggerWatermill,
		repositories: repositories{
			torrent: torrentRepo,
			image:   imageRepository,
		},
		imagePersister: imagePersister,
		db:             sqliteDatabase,
	}, nil
}

func NewTestingContainer() (testingContainer, error) {
	c, err := NewDefaultContainer()
	return testingContainer{
		container: c,
	}, err
}

func (t *testingContainer) GetSQLDatabase() *sql.DB {
	return t.db
}

func (c *container) Config() configuration.Config {
	return c.config
}

func (c *container) Logger() *logrus.Logger {
	return c.logger
}

func (c *container) ImagePersister() preview.ImagePersister {
	return c.imagePersister
}

func (c *container) ImageExtractor() preview.ImageExtractor {
	if c.imageExtractor == nil {
		imageExtractor, err := ffmpeg.NewInMemoryFfmpeg(c.logger)
		if err != nil {
			logrus.Fatal(err)
		}
		c.imageExtractor = imageExtractor
	}
	return c.imageExtractor
}

func (c *container) MagnetClient() preview.MagnetClient {
	return c.getTorrentIntegration()
}

func (c *container) TorrentDownloader() preview.TorrentDownloader {
	return c.getTorrentIntegration()
}

func (c *container) CommandBus() bus.Command {
	return c.cqrs().CommandBus()
}

func (c *container) EventBus() bus.Event {
	return c.cqrs().EventBus()
}

func (c *container) CQRSRouter() *message.Router {
	if c.eventSourcing.cqrsRouter != nil {
		return c.eventSourcing.cqrsRouter
	}

	_ = c.cqrs() // It creates the router. A router without bindings is useless.

	return c.eventSourcing.cqrsRouter
}

func (c *container) TorrentRepository() preview.TorrentRepository {
	return c.repositories.torrent
}

func (c *container) ImageRepository() preview.ImageRepository {
	return c.repositories.image
}

func (c *container) cqrs() *cqrs.Facade {
	if c.eventSourcing.cqrsFacade != nil {
		return c.eventSourcing.cqrsFacade
	}

	router, err := message.NewRouter(message.RouterConfig{}, c.loggerWatermill)
	if err != nil {
		panic(err)
	}
	router.AddMiddleware(middleware.Recoverer)
	c.eventSourcing.cqrsRouter = router

	cqrsFacade, err := cqrs.NewFacade(cqrs.FacadeConfig{
		GenerateCommandsTopic: func(commandName string) string {
			return commandName
		},
		CommandHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.CommandHandler {
			return []cqrs.CommandHandler{
				unmagnetize.NewCommandHandler(eb, c.unmagnetizeService(eb)),
				downloadPartials.NewCommandHandler(eb, c.downloadPartialsService()),
			}
		},
		CommandsPublisher: c.commandPublisher(),
		CommandsSubscriberConstructor: func(handlerName string) (message.Subscriber, error) {
			return c.commandSubscriber(), nil
		},
		GenerateEventsTopic: func(eventName string) string {
			return eventName
		},
		EventHandlers: func(cb *cqrs.CommandBus, eb *cqrs.EventBus) []cqrs.EventHandler {
			return []cqrs.EventHandler{
				downloadPartials.NewTorrentCreatedEventHandler(eb, c.downloadPartialsService()),
			}
		},
		EventsPublisher: c.eventPublisher(),
		EventsSubscriberConstructor: func(handlerName string) (message.Subscriber, error) {
			return c.eventSubscriber(), nil
		},
		Router:                router,
		CommandEventMarshaler: cqrs.JSONMarshaler{},
		Logger:                c.loggerWatermill,
	})
	if err != nil {
		panic(err)
	}

	c.eventSourcing.cqrsFacade = cqrsFacade
	return c.eventSourcing.cqrsFacade
}

func (c *container) commandSubscriber() message.Subscriber {
	if c.eventSourcing.messageSubscriber == nil {
		googleSubscriber, err := googlecloud.NewSubscriber(
			googlecloud.SubscriberConfig{
				GenerateSubscriptionName: googlecloud.TopicSubscriptionName,
				ProjectID:                c.config.GooglePubSubProjectID,
			},
			c.loggerWatermill,
		)
		if err != nil {
			panic(err)
		}
		c.eventSourcing.messageSubscriber = googleSubscriber
	}

	return c.eventSourcing.messageSubscriber
}

func (c *container) eventSubscriber() message.Subscriber {
	if c.eventSourcing.messageSubscriber == nil {
		googleSubscriber, err := googlecloud.NewSubscriber(
			googlecloud.SubscriberConfig{
				GenerateSubscriptionName: googlecloud.TopicSubscriptionName,
				ProjectID:                c.config.GooglePubSubProjectID,
			},
			c.loggerWatermill,
		)
		if err != nil {
			panic(err)
		}
		c.eventSourcing.messageSubscriber = googleSubscriber
	}

	return c.eventSourcing.messageSubscriber
}

func (c *container) commandPublisher() message.Publisher {
	if c.eventSourcing.publisher == nil {
		publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{
			ProjectID: c.config.GooglePubSubProjectID,
		}, c.loggerWatermill)
		if err != nil {
			panic(err)
		}
		c.eventSourcing.publisher = publisher
	}

	return c.eventSourcing.publisher
}

func (c *container) eventPublisher() message.Publisher {
	if c.eventSourcing.publisher == nil {
		publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{
			ProjectID: c.config.GooglePubSubProjectID,
		}, c.loggerWatermill)
		if err != nil {
			panic(err)
		}
		c.eventSourcing.publisher = publisher
	}

	return c.eventSourcing.publisher
}

func (c *container) getTorrentIntegration() *bittorrentproto.TorrentClient {
	if c.torrentIntegration == nil {
		torrentClient, err := torrent.NewClient(configuration.GetTorrentConf(c.config))
		if err != nil {
			panic(err)
		}
		c.torrentIntegration = bittorrentproto.NewTorrentClient(torrentClient, c.logger)
	}
	return c.torrentIntegration
}

func (c *container) downloadPartialsService() downloadPartials.Service {
	return downloadPartials.NewService(
		c.logger,
		c.repositories.torrent,
		c.TorrentDownloader(),
		c.ImageExtractor(),
		c.imagePersister,
		c.repositories.image,
	)
}

func (c *container) unmagnetizeService(eb *cqrs.EventBus) unmagnetize.Service {
	return unmagnetize.NewService(c.logger, eb, c.MagnetClient(), c.repositories.torrent)
}
