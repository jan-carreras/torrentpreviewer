package http

import (
	"database/sql"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/platform/storage/sqlite"

	"github.com/sirupsen/logrus"
)

type Container struct {
	Logger          *logrus.Logger
	TorrentRepo     preview.TorrentRepository
	ImageRepository preview.ImageRepository
}

func NewContainer() (Container, error) {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{}
	logger.Level = logrus.DebugLevel

	sqliteDatabase, err := sql.Open("sqlite3", "prevtorrent.sqlite")
	if err != nil {
		return Container{}, err
	}

	return Container{
		Logger:          logger,
		TorrentRepo:     sqlite.NewTorrentRepository(sqliteDatabase),
		ImageRepository: sqlite.NewImageRepository(sqliteDatabase),
	}, nil
}
