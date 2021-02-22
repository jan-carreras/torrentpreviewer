package sqlite_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/platform/storage/sqlite"
	"testing"
)

func TestNewTorrentRepository_StoreTorrentAndImages(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"
	name := "Torrent Example"
	length := 250
	pieceLength := 10

	f1, err := preview.NewFileInfo(0, 100, "img.jpg")
	require.NoError(t, err)
	f2, err := preview.NewFileInfo(1, 150, "img2.jpg")
	require.NoError(t, err)
	files := []preview.FileInfo{f1, f2}

	raw := []byte("1234")

	db, sqlMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	sqlMock.ExpectExec(
		"INSERT INTO torrents (id, name, length, pieceLength, raw) VALUES (?, ?, ?, ?, ?)").
		WithArgs(torrentID, name, length, pieceLength, raw).
		WillReturnResult(driver.ResultNoRows)

	sqlMock.ExpectExec(
		"INSERT INTO files (torrent_id, id, name, length) VALUES (?, ?, ?, ?)").
		WithArgs(torrentID, f1.ID(), f1.Name(), f1.Length()).
		WillReturnResult(driver.RowsAffected(1))

	sqlMock.ExpectExec(
		"INSERT INTO files (torrent_id, id, name, length) VALUES (?, ?, ?, ?)").
		WithArgs(torrentID, f2.ID(), f2.Name(), f2.Length()).
		WillReturnResult(driver.RowsAffected(1))

	repository := sqlite.NewTorrentRepository(db)
	torrent, err := preview.NewInfo(torrentID, name, pieceLength, files, raw)
	require.NoError(t, err)

	err = repository.Persist(context.Background(), torrent)
	require.NoError(t, err)

	require.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestNewTorrentRepository_ErrorOnPersist(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"
	name := "Torrent Example"
	length := 250
	pieceLength := 10

	f1, err := preview.NewFileInfo(0, 100, "img.jpg")
	require.NoError(t, err)
	f2, err := preview.NewFileInfo(1, 150, "img2.jpg")
	require.NoError(t, err)
	files := []preview.FileInfo{f1, f2}

	raw := []byte("1234")

	db, sqlMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	sqlMock.ExpectExec(
		"INSERT INTO torrents (id, name, length, pieceLength, raw) VALUES (?, ?, ?, ?, ?)").
		WithArgs(torrentID, name, length, pieceLength, raw).
		WillReturnError(errors.New("fake error at insert"))

	repository := sqlite.NewTorrentRepository(db)
	torrent, err := preview.NewInfo(torrentID, name, pieceLength, files, raw)
	require.NoError(t, err)

	err = repository.Persist(context.Background(), torrent)
	require.Error(t, err)

	require.NoError(t, sqlMock.ExpectationsWereMet())
}
