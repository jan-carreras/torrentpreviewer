package getTorrent

import (
	"context"
	"fmt"
	"prevtorrent/internal/preview"

	"github.com/sirupsen/logrus"
)

type Service struct {
	logger          *logrus.Logger
	torrentRepo     preview.TorrentRepository
	imageRepository preview.ImageRepository
}

func NewService(
	logger *logrus.Logger,
	torrentRepo preview.TorrentRepository,
	imageRepository preview.ImageRepository,
) *Service {
	return &Service{
		logger:          logger,
		torrentRepo:     torrentRepo,
		imageRepository: imageRepository,
	}
}

type CMD struct {
	TorrentID string
}

func (s *Service) Get(ctx context.Context, cmd CMD) (preview.Info, error) {
	torrent, err := s.torrentRepo.Get(ctx, cmd.TorrentID)
	if err != nil {
		return preview.Info{}, err
	}

	images, err := s.imageRepository.ByTorrent(ctx, torrent.ID())
	if err != nil {
		return preview.Info{}, err
	}

	fmt.Println(images.Images())
	for _, img := range images.Images() {
		file := torrent.File(img.FileID())
		if err := file.AddImage(img); err != nil {
			return preview.Info{}, err
		}
	}

	return torrent, nil
}
