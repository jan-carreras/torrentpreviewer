package downloadPartials

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"prevtorrent/internal/preview"
)

type Service struct {
	logger            *logrus.Logger
	torrentRepository preview.TorrentRepository
	magnetClient      preview.MagnetClient
	imageExtractor    preview.ImageExtractor
	imageRepository   preview.ImageRepository
}

func NewService(
	logger *logrus.Logger,
	torrentRepository preview.TorrentRepository,
	magnetClient preview.MagnetClient,
	imageExtractor preview.ImageExtractor,
	imageRepository preview.ImageRepository,
) Service {
	return Service{
		logger:            logger,
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

	plan := preview.NewDownloadPlan(info)
	for _, file := range info.SupportedFiles() {
		if err := plan.AddDownloadToPlan(file, file.DownloadSize(), 0); err != nil {
			return err
		}
	}
	partsCh, err := s.magnetClient.DownloadParts(ctx, *plan)

	registry := preview.NewPieceRegistry(plan)

	go s.processParts(ctx, partsCh, registry)

	for part := range registry.SubscribeAllPartsDownloaded() {
		s.logger.WithFields(logrus.Fields{
			"name":       part.Name(),
			"pieceCount": part.PieceCount(),
		}).Debug("download completed. We have all the parts, we only need to merge")
		bundle := preview.NewBundlePlan()
		downloadedPart, err := bundle.Bundle(registry, info.ID(), part)
		if err != nil {
			return err
		}
		fmt.Println("downloaded:", downloadedPart.Name())
	}

	// TODO: If we don't need the files in bold.db those can be deleted
	return err
}

func (s Service) processParts(ctx context.Context, partsCh chan preview.Piece, registry *preview.PieceRegistry) {
	// TODO: Fucking "part" or fucking "piece"?!
	for {
		select {
		case piece, isOpen := <-partsCh:
			if !isOpen {
				return
			}

			log := s.logger.WithFields(logrus.Fields{
				"torrentID": piece.TorrentID(),
				"piece":     piece.ID(),
			})

			if err := registry.AddPiece(piece); err != nil {
				log.Error(err)
				return
			}
			log.Debug("part added to registry")
		case <-ctx.Done():
			return
		}
	}
}
