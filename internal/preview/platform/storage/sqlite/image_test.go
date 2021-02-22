package sqlite_test

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/platform/storage/sqlite"
	"testing"
)

func Test_ImageRepositoryPersists(t *testing.T) {
	torrentID := "1234"
	fileID := 0
	name := "image.jpg"
	length := 100

	db, sqlMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	sqlMock.ExpectExec(
		"INSERT INTO media (torrent_id, file_id, name, length) VALUES (?, ?, ?, ?)").
		WithArgs(torrentID, fileID, name, length).
		WillReturnResult(sqlmock.NewResult(0, 1))

	imageRepository := sqlite.NewImageRepository(db)

	err = imageRepository.Persist(context.Background(), preview.NewImage(torrentID, fileID, name, length))
	require.NoError(t, err)
}

func Test_ImageRepositoryErrorOnPersist(t *testing.T) {
	torrentID := "1234"
	fileID := 0
	name := "image.jpg"
	length := 100

	db, sqlMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	sqlMock.ExpectExec(
		"INSERT INTO media (torrent_id, file_id, name, length) VALUES (?, ?, ?, ?)").
		WithArgs(torrentID, fileID, name, length).
		WillReturnError(errors.New("fake UNIQUE CONSTRAINT FAIL"))

	imageRepository := sqlite.NewImageRepository(db)

	err = imageRepository.Persist(context.Background(), preview.NewImage(torrentID, fileID, name, length))
	require.Error(t, err)
}
