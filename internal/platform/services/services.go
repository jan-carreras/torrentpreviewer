package services

import "C"
import (
	"prevtorrent/internal/platform/container"
	"prevtorrent/internal/preview/getTorrent"
	"prevtorrent/internal/preview/importTorrent"
	"prevtorrent/internal/preview/unmagnetize"
)

type Services struct {
	GetTorrent    getTorrent.Service
	Unmagnetize   unmagnetize.Service
	ImportTorrent importTorrent.Service
}

func NewServices(c container.Container) (Services, error) {
	getTorrentService := getTorrent.NewService(c.Logger, c.TorrentRepo, c.ImageRepository)
	unmagnetizeService := unmagnetize.NewService(c.Logger, c.MagnetClient, c.TorrentRepo)
	importTorrentService := importTorrent.NewService(c.Logger, c.TorrentDownloader, c.TorrentRepo)

	return Services{
		GetTorrent:    getTorrentService,
		Unmagnetize:   unmagnetizeService,
		ImportTorrent: importTorrentService,
	}, nil
}
