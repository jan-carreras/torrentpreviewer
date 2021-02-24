package services

import "C"
import (
	"prevtorrent/internal/platform/container"
	"prevtorrent/internal/preview/downloadPartials"
	"prevtorrent/internal/preview/getTorrent"
	"prevtorrent/internal/preview/importTorrent"
	"prevtorrent/internal/preview/unmagnetize"
)

type Services struct {
	c                container.Container
	getTorrent       *getTorrent.Service
	unmagnetize      *unmagnetize.Service
	importTorrent    *importTorrent.Service
	downloadPartials *downloadPartials.Service
}

func NewServices(c container.Container) (Services, error) {
	return Services{c: c}, nil
}

func (s *Services) GetTorrent() getTorrent.Service {
	if s.getTorrent == nil {
		service := getTorrent.NewService(s.c.Logger, s.c.TorrentRepo, s.c.ImageRepository)
		s.getTorrent = &service
	}
	return *s.getTorrent
}

func (s *Services) Unmagnetize() unmagnetize.Service {
	if s.unmagnetize == nil {
		service := unmagnetize.NewService(s.c.Logger, s.c.MagnetClient(), s.c.TorrentRepo)
		s.unmagnetize = &service
	}

	return *s.unmagnetize
}

func (s *Services) ImportTorrent() importTorrent.Service {
	if s.importTorrent == nil {
		service := importTorrent.NewService(s.c.Logger, s.c.TorrentDownloader(), s.c.TorrentRepo)
		s.importTorrent = &service
	}
	return *s.importTorrent
}

func (s *Services) DownloadPartials() downloadPartials.Service {

	if s.downloadPartials == nil {
		service := downloadPartials.NewService(
			s.c.Logger,
			s.c.TorrentRepo,
			s.c.TorrentDownloader(),
			s.c.ImageExtractor(),
			s.c.ImagePersister,
			s.c.ImageRepository,
		)
		s.downloadPartials = &service
	}

	return *s.downloadPartials
}
