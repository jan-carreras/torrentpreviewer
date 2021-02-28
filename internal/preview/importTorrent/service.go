package importTorrent

import (
	"context"
	"errors"
	"prevtorrent/internal/platform/bus"
	"prevtorrent/internal/preview"

	"github.com/sirupsen/logrus"
)

type Service struct {
	logger            *logrus.Logger
	commandBus        bus.Command
	torrentDownloader preview.TorrentDownloader
	torrentRepository preview.TorrentRepository
}

func NewService(
	logger *logrus.Logger,
	commandBus bus.Command,
	torrentDownloader preview.TorrentDownloader,
	torrentRepository preview.TorrentRepository,
) Service {
	return Service{
		logger:            logger,
		commandBus:        commandBus,
		torrentDownloader: torrentDownloader,
		torrentRepository: torrentRepository,
	}
}

func (s Service) Import(ctx context.Context, cmd CMD) (preview.Torrent, error) {
	torrent, err := s.torrentDownloader.Import(ctx, cmd.TorrentRaw)
	if err != nil {
		return preview.Torrent{}, err
	}

	if _, err = s.torrentRepository.Get(ctx, torrent.ID()); err == nil {
		return torrent, nil // Exists already. Doing nothing
	}

	if errors.Is(err, preview.ErrNotFound) {
		err = nil
		if err := s.torrentRepository.Persist(ctx, torrent); err != nil {
			return preview.Torrent{}, err
		}
		if err := s.commandBus.Send(ctx, preview.NewTorrentCreatedEvent(torrent.ID())); err != nil {
			return preview.Torrent{}, err
		}
	}

	return torrent, err
}
