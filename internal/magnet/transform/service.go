package transform

import (
	"context"
	"prevtorrent/internal/magnet"
)

type Service struct {
	magnetResolver    magnet.MagnetResolver
	torrentRepository magnet.TorrentRepository
}

func NewService(
	magnetResolver magnet.MagnetResolver,
	torrentRepository magnet.TorrentRepository,
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
	m, err := magnet.NewMagnet(cmd.Magnet)
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
