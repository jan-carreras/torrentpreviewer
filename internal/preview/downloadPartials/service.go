package downloadPartials

import (
	"context"
	"prevtorrent/internal/preview"
)

type Service struct {
	torrentRepository preview.TorrentRepository
	magnetClient      preview.MagnetClient
	imageExtractor    preview.ImageExtractor
	imageRepository   preview.ImageRepository
}

func NewService(
	torrentRepository preview.TorrentRepository,
	magnetClient preview.MagnetClient,
	imageExtractor preview.ImageExtractor,
	imageRepository preview.ImageRepository,
) Service {
	return Service{
		torrentRepository: torrentRepository,
		magnetClient:      magnetClient,
		imageExtractor:    imageExtractor,
		imageRepository:   imageRepository,
	}
}

func (s Service) DownloadPartials(ctx context.Context, cmd CMD) error {
	info, err := s.torrentRepository.Get(ctx, cmd.ID)
	if err != nil {
		return err
	}

	downloadPlan := preview.NewDownloadPlan(info)
	for _, file := range info.SupportedFiles() {
		if err := downloadPlan.Download(file, file.DownloadSize(), 0); err != nil {
			return err
		}
	}
	downloads, err := s.magnetClient.DownloadParts(ctx, *downloadPlan)

	for _, download := range downloads {
		data, err := s.imageExtractor.ExtractImage(ctx, download.Data(), 0)
		if err != nil {
			continue // TODO: We are ignoring the error to try to see if other videos can be recovered
		}
		err = s.imageRepository.PersistFile(ctx, download.Name(), data)
		if err != nil {
			return err
		}
	}

	// TODO: If we don't need the files in bold.db those can be deleted
	return err
}
