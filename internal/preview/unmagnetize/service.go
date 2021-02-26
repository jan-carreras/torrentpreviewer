package unmagnetize

import (
	"context"
	"errors"
	"prevtorrent/internal/platform/bus"
	"prevtorrent/internal/preview"

	"github.com/sirupsen/logrus"
)

type Service struct {
	log               *logrus.Logger
	eventBus          bus.Event
	magnetResolver    preview.MagnetClient
	torrentRepository preview.TorrentRepository
}

func NewService(
	log *logrus.Logger,
	eventBus bus.Event,
	magnetResolver preview.MagnetClient,
	torrentRepository preview.TorrentRepository,
) Service {
	return Service{
		log:               log,
		eventBus:          eventBus,
		magnetResolver:    magnetResolver,
		torrentRepository: torrentRepository,
	}
}

func (s Service) Handle(ctx context.Context, cmd CMD) (preview.Info, error) {
	m, err := preview.NewMagnet(cmd.Magnet)
	if err != nil {
		return preview.Info{}, err
	}

	t, err := s.torrentRepository.Get(ctx, m.ID())
	if err == nil {
		s.log.WithFields(logrus.Fields{
			"magnet":   m.Value(),
			"magnetID": m.ID(),
		}).Debug("already imported in DB. skipping.")
		return t, nil
	}

	if !errors.Is(err, preview.ErrNotFound) {
		s.log.WithFields(logrus.Fields{
			"magnet":   m.Value(),
			"magnetID": m.ID(),
			"error":    err,
		}).Debug("error when reading torrent")
		return preview.Info{}, err
	}

	s.log.WithFields(logrus.Fields{
		"magnet":   m.Value(),
		"magnetID": m.ID(),
	}).Debug("not found in db. about to resolve the magnet using network")

	torrent, err := s.magnetResolver.Resolve(ctx, m)
	if err != nil {
		return preview.Info{}, err
	}

	err = s.torrentRepository.Persist(ctx, torrent)
	if err != nil {
		return preview.Info{}, err
	}

	if err := s.eventBus.Publish(ctx, preview.NewTorrentCreatedEvent(torrent.ID())); err != nil {
		return preview.Info{}, err
	}

	return torrent, nil
}
