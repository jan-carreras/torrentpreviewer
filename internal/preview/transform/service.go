package transform

import (
	"context"
	"prevtorrent/internal/preview"
)

type Service struct {
	magnetResolver    preview.MagnetClient
	torrentRepository preview.TorrentRepository
}

func NewService(
	magnetResolver preview.MagnetClient,
	torrentRepository preview.TorrentRepository,
) *Service {
	return &Service{
		magnetResolver:    magnetResolver,
		torrentRepository: torrentRepository,
	}
}

type ServiceCMD struct {
	Magnet string
}

func (s *Service) ToTorrent(ctx context.Context, cmd ServiceCMD) error {
	m, err := preview.NewMagnet(cmd.Magnet)
	if err != nil {
		return err
	}

	torrent, err := s.magnetResolver.Resolve(ctx, m)
	if err != nil {
		return err
	}

	if err := s.torrentRepository.Persist(ctx, torrent); err != nil {
		return err
	}

	return nil
}
