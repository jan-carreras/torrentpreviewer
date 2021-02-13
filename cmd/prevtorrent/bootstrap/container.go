package bootstrap

import (
	"database/sql"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/platform/client/bittorrentproto"
	"prevtorrent/internal/preview/platform/storage/file"
	"prevtorrent/internal/preview/platform/storage/inmemory/ffmpeg"
	"prevtorrent/internal/preview/platform/storage/sqlite"
	"sort"
	"strings"
)

type container struct {
	torrentIntegration preview.MagnetClient
	torrentRepo        preview.TorrentRepository
	logger             *logrus.Logger
	imageExtractor     preview.ImageExtractor
	imageRepository    preview.ImageRepository
}

func newContainer() (container, error) {
	if err := getConfig(); err != nil {
		return container{}, err
	}

	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{}
	logger.Level = logrus.DebugLevel

	imageExtractor := ffmpeg.NewInMemoryFfmpeg(logger)

	imageRepository := file.NewImageRepository(logger, viper.GetString("ImageDir"))

	torrentRepo, err := getTorrentRepository(logger)
	if err != nil {
		return container{}, err
	}

	conf := torrent.NewDefaultClientConfig()
	conf.DefaultStorage = storage.NewBoltDB(viper.GetString("BoltDBDir"))
	conf.DisableIPv6 = true

	torrentClient, err := torrent.NewClient(conf)
	if err != nil {
		return container{}, err
	}

	torrentIntegration := bittorrentproto.NewTorrentClient(torrentClient, logger)

	return container{
		torrentIntegration: torrentIntegration,
		torrentRepo:        torrentRepo,
		logger:             logger,
		imageExtractor:     imageExtractor,
		imageRepository:    imageRepository,
	}, nil
}

func getTorrentRepository(logger *logrus.Logger) (preview.TorrentRepository, error) {
	fileDriver := func() (preview.TorrentRepository, error) {
		return file.NewTorrentRepository(
			logger,
			viper.GetString("TorrentDir"),
			viper.GetString("DownloadsDir"),
		), nil
	}
	sqliteDrier := func() (preview.TorrentRepository, error) {
		sqliteDatabase, err := sql.Open("sqlite3", viper.GetString("SqlitePath"))
		if err != nil {
			return nil, err
		}

		return sqlite.NewTorrentRepository(sqliteDatabase), nil
	}

	drivers := map[string]func() (preview.TorrentRepository, error){
		"file":   fileDriver,
		"sqlite": sqliteDrier,
	}

	driver := viper.GetString("TorrentDriver")
	if maker, found := drivers[driver]; !found {
		supported := make([]string, 0, len(drivers))
		for k := range drivers {
			supported = append(supported, k)
		}
		sort.Strings(supported)
		return nil, fmt.Errorf("unsopported driver %v. supported: %v", driver, strings.Join(supported, ", "))
	} else {
		return maker()
	}
}
