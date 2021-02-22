package preview_test

import (
	"prevtorrent/internal/preview"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"

	fi, err := preview.NewFileInfo(0, 1000, "movie.mp4")
	assert.NoError(t, err)
	f2, err := preview.NewFileInfo(0, 2000, "movie2.mp4")
	assert.NoError(t, err)
	f3, err := preview.NewFileInfo(0, 10, "subtitles.srt")
	assert.NoError(t, err)

	files := []preview.FileInfo{fi, f2, f3}
	supportedFile := []preview.FileInfo{fi, f2}

	torrent, err := preview.NewInfo(torrentID, "test movie", 100, files, []byte("12345"))
	assert.NoError(t, err)

	assert.Equal(t, "cb84ccc10f296df72d6c40ba7a07c178a4323a14", torrent.ID())
	assert.Equal(t, "test movie", torrent.Name())
	assert.Equal(t, []byte("12345"), torrent.Raw())
	assert.Equal(t, 100, torrent.PieceLength())
	assert.Equal(t, 3010, torrent.TotalLength())
	assert.Equal(t, supportedFile, torrent.SupportedFiles())
	assert.Equal(t, files, torrent.Files())
}

func TestInfo_InvalidTorrentID(t *testing.T) {
	torrentID := "invalid ID"

	fi, err := preview.NewFileInfo(0, 1000, "movie.mp4")
	assert.NoError(t, err)
	_, err = preview.NewInfo(torrentID, "", 100, []preview.FileInfo{fi}, []byte("12345"))
	assert.Equal(t, preview.ErrInvalidTorrentID, err)
}

func TestFileInfo(t *testing.T) {
	idx := 0
	length := 1000
	name := "movie.mp4"

	fi, err := preview.NewFileInfo(idx, length, name)
	assert.NoError(t, err)

	assert.Equal(t, idx, fi.ID())
	assert.Equal(t, length, fi.Length())
	assert.Equal(t, name, fi.Name())
	assert.Equal(t, length, fi.DownloadSize())
	assert.True(t, fi.IsSupportedExtension())
}

func TestFileInfo_BigFile(t *testing.T) {
	idx := 0
	length := preview.DownloadSize * 10
	name := "movie.mp4"

	fi, err := preview.NewFileInfo(idx, length, name)
	assert.NoError(t, err)

	assert.Equal(t, preview.DownloadSize, fi.DownloadSize())
}
