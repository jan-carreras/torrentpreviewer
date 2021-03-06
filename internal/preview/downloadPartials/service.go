package downloadPartials

import (
	"context"
	"errors"
	"prevtorrent/internal/preview"

	"github.com/sirupsen/logrus"
)

const (
	frameTimeToExtract = 5 // in seconds
)

type Service struct {
	logger            *logrus.Logger
	torrentRepository preview.TorrentRepository
	torrentDownloader preview.TorrentDownloader
	imageExtractor    preview.ImageExtractor
	imagePersister    preview.ImagePersister
	imageRepository   preview.ImageRepository
}

func NewService(
	logger *logrus.Logger,
	torrentRepository preview.TorrentRepository,
	torrentDownloader preview.TorrentDownloader,
	imageExtractor preview.ImageExtractor,
	imagePersister preview.ImagePersister,
	imageRepository preview.ImageRepository,
) Service {
	return Service{
		logger:            logger,
		torrentRepository: torrentRepository,
		torrentDownloader: torrentDownloader,
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

	if len(cmd.Files) == 0 {
		s.logger.WithFields(logrus.Fields{
			"torrentID": torrent.ID(),
			"name":      torrent.Name(),
		}).Debug("the download plan have 0 files thus nothing to do")
		return nil
	}

	s.logger.WithFields(logrus.Fields{
		"torrentID": torrent.ID(),
		"name":      torrent.Name(),
	}).Debug("torrent to be processed")

	torrentImages, err := s.imageRepository.ByTorrent(ctx, cmd.ID)
	if err != nil {
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"torrentID":          torrent.ID(),
		"name":               torrent.Name(),
		"imagesTorrentCount": len(torrentImages.Images()),
	}).Debug("images that we already have for the torrent")

	plan := preview.NewDownloadPlan(torrent)
	for _, file := range cmd.Files {
		f := torrent.File(file.FileID)
		if err := plan.Add(torrentImages, f, file.Start, file.Length); err != nil {
			return err
		}
	}

	s.logger.WithFields(logrus.Fields{
		"torrentID":        torrent.ID(),
		"name":             torrent.Name(),
		"pieceLength":      torrent.PieceLength(),
		"pieceCount":       plan.CountPieces(),
		"downloadPlanSize": plan.DownloadSize(),
	}).Debug("pieces to download")

	registry, err := s.torrentDownloader.DownloadParts(ctx, *plan)
	if err != nil {
		return err
	}

	err = registry.RunOnPieceReady(ctx, func(part preview.PieceRange) error {
		downloaded, err := s.getBundle(registry, part)
		if err != nil {
			return err
		}
		imgBytes, err := s.extractImage(ctx, part, downloaded)
		if err != nil {
			return err
		}

		if err := s.storeBinaryImage(ctx, imgBytes, part.Name(), part); err != nil {
			return err
		}

		img := preview.NewImage(
			part.Torrent().ID(),
			part.FileID(),
			part.Name(),
			len(imgBytes),
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
	img, err := s.imageExtractor.ExtractImage(ctx, downloadedPart.Data(), frameTimeToExtract)
	if errors.Is(err, preview.ErrAtomNotFound) || errors.Is(err, preview.ErrNotAbleToGenerateImage) {
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
