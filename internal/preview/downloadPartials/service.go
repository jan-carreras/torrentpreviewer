package downloadPartials

import (
	"context"
	"prevtorrent/internal/preview"
)

const (
	MiB = 1 << (10 * 2)
)

type Service struct {
	torrentRepository preview.TorrentRepository
	magnetClient      preview.MagnetClient
}

func NewService(
	torrentRepository preview.TorrentRepository,
	magnetClient preview.MagnetClient,
) *Service {
	return &Service{torrentRepository: torrentRepository, magnetClient: magnetClient}
}

type CMD struct {
	ID string
}

func (s *Service) DownloadPartials(ctx context.Context, cmd CMD) error {
	info, err := s.torrentRepository.Get(ctx, cmd.ID)
	if err != nil {
		return err
	}

	downloadPlan := preview.NewDownloadPlan(info)
	for _, file := range info.SupportedFiles() {
		if err := downloadPlan.Download(file, 100*MiB, 0); err != nil {
			return err
		}
	}
	downloads, err := s.magnetClient.DownloadParts(ctx, *downloadPlan)
	_ = downloads // TODO: Store them in disk, ofc
	return err
}
