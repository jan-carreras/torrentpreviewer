package importTorrent

import (
	"context"
	"errors"
	"prevtorrent/internal/preview"

	"github.com/sirupsen/logrus"
)

type Service struct {
	logger            *logrus.Logger
	torrentDownloader preview.TorrentDownloader
	torrentRepository preview.TorrentRepository
}

func NewService(
	logger *logrus.Logger,
	torrentDownloader preview.TorrentDownloader,
	torrentRepository preview.TorrentRepository,
) Service {
	return Service{
		logger:            logger,
		torrentDownloader: torrentDownloader,
		torrentRepository: torrentRepository,
	}
}

type CMD struct {
	TorrentRaw []byte
}

func (s Service) Import(ctx context.Context, cmd CMD) (preview.Info, error) {
	torrent, err := s.torrentDownloader.Import(ctx, cmd.TorrentRaw)
	if err != nil {
		return preview.Info{}, err
	}

	if _, err = s.torrentRepository.Get(ctx, torrent.ID()); err == nil {
		return torrent, nil // Exists already. Doing nothing
	}

	if errors.Is(err, preview.ErrNotFound) {
		err = nil
		if err := s.torrentRepository.Persist(ctx, torrent); err != nil {
			return preview.Info{}, err
		}
	}

	return torrent, err
}
