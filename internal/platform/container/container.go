package container

import (
	"database/sql"
	"prevtorrent/internal/platform/bus"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/makeDownloadPlan"
	"prevtorrent/internal/preview/platform/client/bittorrentproto"
	"prevtorrent/internal/preview/platform/configuration"
	"prevtorrent/internal/preview/platform/storage/file"
	"prevtorrent/internal/preview/platform/storage/inmemory/ffmpeg"
	"prevtorrent/internal/preview/platform/storage/sqlite"
	"prevtorrent/internal/preview/unmagnetize"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
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
	eventDriver       events
	cqrsFacade        *cqrs.Facade
	commandPublisher  message.Publisher
	commandSubscriber message.Subscriber
	eventPublisher    message.Publisher
	eventSubscriber   message.Subscriber
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

	eventDriver := makeEventDriver(config, loggerWatermill)

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
		eventSourcing: eventSourcing{
			eventDriver: eventDriver,
		},
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
				unmagnetize.NewCommandHandler(c.unmagnetizeService(eb)),
				downloadPartials.NewCommandHandler(c.downloadPartialsService()),
				makeDownloadPlan.NewCommandHandler(c.makeDownloadPlan(cb)),
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
				makeDownloadPlan.NewTorrentCreatedEventHandler(c.makeDownloadPlan(cb)),
			}
		},
		EventsPublisher:             c.eventPublisher(),
		EventsSubscriberConstructor: c.eventSourcing.eventDriver.eventSubscriber,
		Router:                      router,
		CommandEventMarshaler:       cqrs.JSONMarshaler{},
		Logger:                      c.loggerWatermill,
	})
	if err != nil {
		panic(err)
	}

	c.eventSourcing.cqrsFacade = cqrsFacade
	return c.eventSourcing.cqrsFacade
}

func (c *container) commandPublisher() message.Publisher {
	if c.eventSourcing.commandPublisher == nil {
		c.eventSourcing.commandPublisher = c.eventSourcing.eventDriver.commandPublisher()
	}

	return c.eventSourcing.commandPublisher
}

func (c *container) commandSubscriber() message.Subscriber {
	if c.eventSourcing.eventSubscriber != nil {
		return c.eventSourcing.eventSubscriber
	}
	c.eventSourcing.eventSubscriber = c.eventSourcing.eventDriver.commandSubscriber()
	return c.eventSourcing.eventSubscriber
}

func (c *container) eventPublisher() message.Publisher {
	if c.eventSourcing.eventPublisher == nil {
		c.eventSourcing.eventPublisher = c.eventSourcing.eventDriver.eventPublisher()
	}

	return c.eventSourcing.eventPublisher
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

func (c *container) makeDownloadPlan(cb bus.Command) makeDownloadPlan.Service {
	return makeDownloadPlan.NewService(
		c.logger,
		cb,
		c.repositories.torrent,
		c.repositories.image,
	)
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

func (c *container) unmagnetizeService(eb bus.Event) unmagnetize.Service {
	return unmagnetize.NewService(c.logger, eb, c.MagnetClient(), c.repositories.torrent)
}
