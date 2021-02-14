package downloadPartials

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"prevtorrent/internal/preview"
	"time"
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

	supportedFiles := info.SupportedFiles()
	if len(supportedFiles) == 0 {
		s.logger.WithFields(logrus.Fields{
			"name":       info.Name(),
			"pieceCount": len(info.SupportedFiles()),
		}).Info("this torrent has not any supported file, thus skipped")
		return nil
	}

	plan := preview.NewDownloadPlan(info)
	for _, file := range supportedFiles {
		if err := plan.AddDownloadToPlan(file, file.DownloadSize(), 0); err != nil {
			return err
		}
	}
	partsCh, err := s.magnetClient.DownloadParts(ctx, *plan)
	if err != nil {
		return err
	}

	registry := preview.NewPieceRegistry(plan)

	go s.processParts(ctx, partsCh, registry)

	for {
		select {
		case part, isOpen := <-registry.SubscribeAllPartsDownloaded():
			if !isOpen {
				return nil
			}
			err := s.extractAndStoreImage(ctx, registry, part, info.ID())
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return errors.New("context cancelled")
		case <-time.Tick(time.Second * 3):
			s.logger.WithFields(logrus.Fields{
				"torrent":   info.Name(),
				"torrentID": info.ID(),
			}).Debug("waiting for parts downloaded to arrive")
		}
	}
}

func (s Service) extractAndStoreImage(ctx context.Context, registry *preview.PieceRegistry, part preview.PieceRange, torrentID string) error {
	s.logger.WithFields(logrus.Fields{
		"torrentID":  torrentID,
		"name":       part.Name(),
		"pieceCount": part.PieceCount(),
	}).Debug("download completed")
	bundle := preview.NewBundlePlan()
	downloadedPart, err := bundle.Bundle(registry, torrentID, part)
	if err != nil {
		return err
	}
	// TODO: If we don't need the files in bold.db those can be deleted
	img, err := s.imageExtractor.ExtractImage(ctx, downloadedPart.Data(), 5)
	if err != nil {
		return err
	}
	s.logger.WithFields(logrus.Fields{
		"torrentID": torrentID,
		"name":      part.Name(),
	}).Debug("image extracted successfully")

	// TODO: Register persisted image in the DB for reference
	err = s.imageRepository.PersistFile(ctx, downloadedPart.Name(), img)
	if err != nil {
		return err
	}
	s.logger.WithFields(logrus.Fields{
		"torrentID": torrentID,
		"name":      downloadedPart.Name(),
	}).Debug("image persisted successfully")

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
