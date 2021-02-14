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
	torrent, err := s.torrentRepository.Get(ctx, cmd.ID)
	if err != nil {
		return err
	}

	plan := preview.NewDownloadPlan(torrent)
	if err := plan.AddAll(); err != nil {
		return err
	}

	registry, err := s.magnetClient.DownloadParts(ctx, *plan)
	if err != nil {
		return err
	}

	registry.ListenForPieces(ctx)

	for {
		select {
		case part, isOpen := <-registry.SubscribeAllPartsDownloaded():
			if !isOpen {
				return nil
			}
			err := s.extractAndStoreImage(ctx, registry, part, torrent.ID())
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return errors.New("context cancelled")
		case <-time.Tick(time.Second * 3):
			s.logger.WithFields(logrus.Fields{
				"torrent":   torrent.Name(),
				"torrentID": torrent.ID(),
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
	downloadedPart, err := bundle.Bundle(registry, part)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"torrentID":  torrentID,
			"name":       part.Name(),
			"pieceCount": part.PieceCount(),
			"error":      err,
		}).Error("error when bundling the image")
		return err
	}
	// TODO: If we don't need the files in bold.db those can be deleted
	img, err := s.imageExtractor.ExtractImage(ctx, downloadedPart.Data(), 5)
	if errors.Is(err, preview.ErrAtomNotFound) {
		s.logger.WithFields(logrus.Fields{
			"torrentID":  torrentID,
			"name":       part.Name(),
			"pieceCount": part.PieceCount(),
			"error":      err,
			"imgBytes":   len(img),
		}).Warn("atom not found error, ignoring video")
		return nil
	}

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"torrentID":  torrentID,
			"name":       part.Name(),
			"pieceCount": part.PieceCount(),
			"error":      err,
			"imgBytes":   len(img),
		}).Error("error when extracting image from video")
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
