package sqlite_test

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
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

func Test_ImageRepositoryByTorrent(t *testing.T) {
	torrentID := "1234"

	db, sqlMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"torrent_id", "file_id", "name", "length"}).
		AddRow("torrent-1", 0, "img1.jpg", 100).
		AddRow("torrent-1", 1, "img2.jpg", 200)

	sqlMock.ExpectQuery(
		"SELECT media.torrent_id, media.file_id, media.name, media.length FROM media WHERE torrent_id = ? ORDER BY id ASC").
		WithArgs(torrentID).
		WillReturnRows(rows)

	imageRepository := sqlite.NewImageRepository(db)

	images, err := imageRepository.ByTorrent(context.Background(), torrentID)
	require.NoError(t, err)

	require.Len(t, images.Images(), 2)

	img1 := preview.NewImage("torrent-1", 0, "img1.jpg", 100)
	assert.Equal(t, img1, images.Images()[0])

	img2 := preview.NewImage("torrent-1", 1, "img2.jpg", 200)
	assert.Equal(t, img2, images.Images()[1])
}

func Test_ImageRepositoryByTorrent_QueryError(t *testing.T) {
	torrentID := "1234"

	db, sqlMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	sqlMock.ExpectQuery(
		"SELECT media.torrent_id, media.file_id, media.name, media.length FROM media WHERE torrent_id = ? ORDER BY id ASC").
		WithArgs(torrentID).
		WillReturnError(errors.New("fake query error"))

	imageRepository := sqlite.NewImageRepository(db)

	_, err = imageRepository.ByTorrent(context.Background(), torrentID)
	require.Error(t, err)
}
