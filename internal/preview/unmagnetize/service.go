package unmagnetize

import (
	"context"
	"errors"
	"prevtorrent/internal/preview"
	"strings"

	"github.com/sirupsen/logrus"
)

type Service struct {
	log               *logrus.Logger
	magnetResolver    preview.MagnetClient
	torrentRepository preview.TorrentRepository
}

func NewService(
	log *logrus.Logger,
	magnetResolver preview.MagnetClient,
	torrentRepository preview.TorrentRepository,
) Service {
	return Service{
		log:               log,
		magnetResolver:    magnetResolver,
		torrentRepository: torrentRepository,
	}
}

func (s Service) Handle(ctx context.Context, cmd CMD) (string, error) {
	m, err := preview.NewMagnet(strings.TrimSpace(cmd.Magnet)) // TODO: Add test
	if err != nil {
		return "", err
	}

	t, err := s.torrentRepository.Get(ctx, m.ID())
	if err == nil {
		s.log.WithFields(logrus.Fields{
			"magnet":   m.Value(),
			"magnetID": m.ID(),
		}).Debug("already imported in DB. skipping.")
		return t.ID(), nil
	}

	if !errors.Is(err, preview.ErrNotFound) {
		s.log.WithFields(logrus.Fields{
			"magnet":   m.Value(),
			"magnetID": m.ID(),
			"error":    err,
		}).Debug("error when reading torrent")
		return "", err
	}

	s.log.WithFields(logrus.Fields{
		"magnet":   m.Value(),
		"magnetID": m.ID(),
	}).Debug("not found in db. about to resolve the magnet using network")

	torrent, err := s.magnetResolver.Resolve(ctx, m)
	if err != nil {
		return "", err
	}

	return torrent.ID(), s.torrentRepository.Persist(ctx, torrent)
}
