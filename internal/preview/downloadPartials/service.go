package downloadPartials

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"prevtorrent/internal/preview"
)

type Service struct {
	logger            *logrus.Logger
	torrentRepository preview.TorrentRepository
	magnetClient      preview.MagnetClient
	imageExtractor    preview.ImageExtractor
	imagePersister    preview.ImagePersister
	imageRepository   preview.ImageRepository
}

func NewService(
	logger *logrus.Logger,
	torrentRepository preview.TorrentRepository,
	magnetClient preview.MagnetClient,
	imageExtractor preview.ImageExtractor,
	imagePersister preview.ImagePersister,
	imageRepository preview.ImageRepository,
) Service {
	return Service{
		logger:            logger,
		torrentRepository: torrentRepository,
		magnetClient:      magnetClient,
		imageExtractor:    imageExtractor,
		imagePersister:    imagePersister,
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

	err = registry.RunOnPieceReady(ctx, func(part preview.PieceRange) error {
		downloaded, err := s.getBundle(registry, part)
		if err != nil {
			return err
		}
		imgBytes, err := s.extractImage(ctx, part, downloaded)
		if err != nil {
			return err
		}

		if err := s.storeBinaryImage(ctx, imgBytes, downloaded.Name(), part); err != nil {
			return err
		}

		img := preview.NewImage(
			part.Torrent().ID(),
			part.FileID(),
			downloaded.Name(),
			len(imgBytes),
			downloaded.Name(),
		)
		if err := s.imageRepository.Persist(ctx, img); err != nil {
			return err
		}

		return nil
	})
	return err
}

func (s Service) getBundle(registry *preview.PieceRegistry, part preview.PieceRange) (preview.MediaPart, error) {
	s.logger.WithFields(logrus.Fields{
		"torrentID":  part.Torrent().ID(),
		"name":       part.Name(),
		"pieceCount": part.PieceCount(),
	}).Debug("download completed")
	bundle := preview.NewBundlePlan()
	downloadedPart, err := bundle.Bundle(registry, part)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"torrentID":  part.Torrent().ID(),
			"name":       part.Name(),
			"pieceCount": part.PieceCount(),
			"error":      err,
		}).Error("error when bundling the image")
		return preview.MediaPart{}, err
	}

	return downloadedPart, nil
}

func (s Service) extractImage(ctx context.Context, part preview.PieceRange, downloadedPart preview.MediaPart) ([]byte, error) {
	// TODO: If we don't need the files in bold.db those can be deleted
	// TODO: if the image is 0 bytes, means probably an MOOV ATOm problem and we don't need it to be saved
	img, err := s.imageExtractor.ExtractImage(ctx, downloadedPart.Data(), 5)
	if errors.Is(err, preview.ErrAtomNotFound) {
		s.logger.WithFields(logrus.Fields{
			"torrentID":  part.Torrent().ID(),
			"name":       part.Name(),
			"pieceCount": part.PieceCount(),
			"error":      err,
			"imgBytes":   len(img),
		}).Warn("atom not found error, ignoring video")
		return nil, nil
	}

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"torrentID":  part.Torrent().ID(),
			"name":       part.Name(),
			"pieceCount": part.PieceCount(),
			"error":      err,
			"imgBytes":   len(img),
		}).Error("error when extracting image from video")
		return nil, err
	}
	s.logger.WithFields(logrus.Fields{
		"torrentID": part.Torrent().ID(),
		"name":      part.Name(),
	}).Debug("image extracted successfully")

	return img, nil
}

func (s Service) storeBinaryImage(ctx context.Context, img []byte, name string, part preview.PieceRange) error {
	// TODO: Register persisted image in the DB for reference
	err := s.imagePersister.PersistFile(ctx, name, img)
	if err != nil {
		return err
	}
	s.logger.WithFields(logrus.Fields{
		"torrentID": part.Torrent().ID(),
		"name":      name,
	}).Debug("image persisted successfully")

	return nil
}
