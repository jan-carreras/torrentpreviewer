package sqlite_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"prevtorrent/internal/preview"
	"prevtorrent/internal/preview/platform/storage/sqlite"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
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

	sqlMock.ExpectBegin()
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
	sqlMock.ExpectCommit()

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

	sqlMock.ExpectBegin()
	sqlMock.ExpectExec(
		"INSERT INTO torrents (id, name, length, pieceLength, raw) VALUES (?, ?, ?, ?, ?)").
		WithArgs(torrentID, name, length, pieceLength, raw).
		WillReturnError(errors.New("fake error at insert"))
	sqlMock.ExpectRollback()

	repository := sqlite.NewTorrentRepository(db)
	torrent, err := preview.NewInfo(torrentID, name, pieceLength, files, raw)
	require.NoError(t, err)

	err = repository.Persist(context.Background(), torrent)
	require.Error(t, err)

	require.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestTorrentRepository_Get_ErrorOnRead(t *testing.T) {
	torrentID := "1234"

	db, sqlMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	sqlMock.ExpectQuery("SELECT torrents.id, torrents.name, torrents.length, torrents.pieceLength, torrents.raw FROM torrents WHERE id = ?").
		WithArgs(torrentID).
		WillReturnError(errors.New("fake error"))

	repository := sqlite.NewTorrentRepository(db)

	_, err = repository.Get(context.Background(), torrentID)
	require.Error(t, err)
	require.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestTorrentRepository_Get(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"
	torrentName := "torrent name"
	torrentLength := 100
	torrentPieceLength := 10
	torrentRaw := []byte("raw data")

	db, sqlMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id", "name", "length", "pieceLength", "raw"}).
		AddRow(torrentID, torrentName, torrentLength, torrentPieceLength, torrentRaw)

	sqlMock.ExpectQuery("SELECT torrents.id, torrents.name, torrents.length, torrents.pieceLength, torrents.raw FROM torrents WHERE id = ?").
		WithArgs(torrentID).
		WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"torrent_id", "id", "name", "length"}).
		AddRow(torrentID, 0, "img1.jpg", 40).
		AddRow(torrentID, 1, "img2.jpg", 60)

	sqlMock.ExpectQuery("SELECT files.torrent_id, files.id, files.name, files.length FROM files WHERE torrent_id = ? ORDER BY id ASC").
		WithArgs(torrentID).
		WillReturnRows(rows)

	repository := sqlite.NewTorrentRepository(db)

	torrent, err := repository.Get(context.Background(), torrentID)
	require.NoError(t, err)

	assert.Equal(t, torrentID, torrent.ID())
	assert.Equal(t, torrentName, torrent.Name())
	assert.Equal(t, torrentLength, torrent.TotalLength())
	assert.Equal(t, torrentPieceLength, torrent.PieceLength())
	assert.Equal(t, torrentRaw, torrent.Raw())

	require.Len(t, torrent.Files(), 2)

	img := torrent.File(0)
	require.Equal(t, 0, img.ID())
	require.Equal(t, "img1.jpg", img.Name())
	require.Equal(t, 40, img.Length())

	require.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestTorrentRepository_Get_ErrorReadingFiles(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"
	torrentName := "torrent name"
	torrentLength := 100
	torrentPieceLength := 10
	torrentRaw := []byte("raw data")

	db, sqlMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id", "name", "length", "pieceLength", "raw"}).
		AddRow(torrentID, torrentName, torrentLength, torrentPieceLength, torrentRaw)

	sqlMock.ExpectQuery("SELECT torrents.id, torrents.name, torrents.length, torrents.pieceLength, torrents.raw FROM torrents WHERE id = ?").
		WithArgs(torrentID).
		WillReturnRows(rows)

	sqlMock.ExpectQuery("SELECT files.torrent_id, files.id, files.name, files.length FROM files WHERE torrent_id = ? ORDER BY id ASC").
		WithArgs(torrentID).
		WillReturnError(errors.New("fake error reading files"))

	repository := sqlite.NewTorrentRepository(db)

	_, err = repository.Get(context.Background(), torrentID)
	require.Error(t, err)
}
