package bootstrap

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/platform/client/bittorrentproto"
	"prevtorrent/internal/preview/platform/storage/file"
	"prevtorrent/internal/preview/platform/storage/inmemory/ffmpeg"
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

	conf := torrent.NewDefaultClientConfig()
	conf.DefaultStorage = storage.NewBoltDB(viper.GetString("BoltDBDir"))
	conf.DisableIPv6 = true

	torrentClient, err := torrent.NewClient(conf)
	if err != nil {
		return container{}, err
	}

	torrentIntegration := bittorrentproto.NewTorrentClient(torrentClient, logger)
	torrentRepo := file.NewTorrentRepository(
		logger,
		viper.GetString("TorrentDir"),
		viper.GetString("DownloadsDir"),
	)

	imageExtractor := ffmpeg.NewInMemoryFfmpeg(logger)

	imageRepository := file.NewImageRepository(logger, viper.GetString("ImageDir"))

	return container{
		torrentIntegration: torrentIntegration,
		torrentRepo:        torrentRepo,
		logger:             logger,
		imageExtractor:     imageExtractor,
		imageRepository:    imageRepository,
	}, nil
}
