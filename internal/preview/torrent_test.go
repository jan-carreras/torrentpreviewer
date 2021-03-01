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
	f2, err := preview.NewFileInfo(1, 2000, "movie2.mp4")
	assert.NoError(t, err)
	f3, err := preview.NewFileInfo(2, 10, "subtitles.srt")
	assert.NoError(t, err)

	files := []preview.File{fi, f2, f3}
	supportedFile := []preview.File{fi, f2}

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

func TestInfo_From32BitIDTo40(t *testing.T) {
	torrentID := "ZOCmzqipffw7ollmic5hub6bpcsdeoqu"

	torrent, err := preview.NewInfo(torrentID, "test movie", 100, nil, []byte("12345"))
	assert.NoError(t, err)

	assert.Equal(t, "cb84ccc10f296df72d6c40ba7a07c178a4323a14", torrent.ID())
}

func TestInfo_InvalidBase32(t *testing.T) {
	torrentID := "11111111111111111111111111111111"
	_, err := preview.NewInfo(torrentID, "test movie", 100, nil, []byte("12345"))
	assert.Error(t, err)
}

func TestInfo_InvalidLength(t *testing.T) {
	torrentID := "ZOCmzqipffw7ollmic5hub6bpcsdeoqu00"

	_, err := preview.NewInfo(torrentID, "test movie", 100, nil, []byte("12345"))
	assert.Error(t, err)
}

func TestInfo_ValidateFilesHaveNonCorrelativeIndexes(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"

	fi, err := preview.NewFileInfo(0, 1000, "movie.mp4")
	assert.NoError(t, err)
	f2, err := preview.NewFileInfo(1, 2000, "movie2.mp4")
	assert.NoError(t, err)
	f3, err := preview.NewFileInfo(4, 10, "subtitles.srt")
	assert.NoError(t, err)

	files := []preview.File{fi, f2, f3}

	_, err = preview.NewInfo(torrentID, "test movie", 100, files, []byte("12345"))
	assert.Error(t, err)
}

func TestInfo_ValidateFilesHaveDuplicatedIDs(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"

	fi, err := preview.NewFileInfo(0, 1000, "movie.mp4")
	assert.NoError(t, err)
	f2, err := preview.NewFileInfo(0, 2000, "movie2.mp4")
	assert.NoError(t, err)
	f3, err := preview.NewFileInfo(0, 10, "subtitles.srt")
	assert.NoError(t, err)

	files := []preview.File{fi, f2, f3}

	_, err = preview.NewInfo(torrentID, "test movie", 100, files, []byte("12345"))
	assert.Error(t, err)
}

func TestInfo_GetFileByID(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"

	fi, err := preview.NewFileInfo(0, 1000, "movie.mp4")
	assert.NoError(t, err)
	f2, err := preview.NewFileInfo(1, 2000, "movie2.mp4")
	assert.NoError(t, err)
	f3, err := preview.NewFileInfo(2, 10, "subtitles.srt")
	assert.NoError(t, err)

	files := []preview.File{fi, f2, f3}

	torrent, err := preview.NewInfo(torrentID, "test movie", 100, files, []byte("12345"))
	assert.NoError(t, err)

	assert.Equal(t, &fi, torrent.File(0))
	assert.Equal(t, &f2, torrent.File(1))
	assert.Equal(t, &f3, torrent.File(2))
}

func TestInfo_GetUnknownFileID(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"

	torrent, err := preview.NewInfo(torrentID, "test movie", 100, nil, []byte("12345"))
	assert.NoError(t, err)

	assert.Nil(t, torrent.File(10))
}

func TestInfo_InvalidTorrentID(t *testing.T) {
	torrentID := "invalid ID"

	fi, err := preview.NewFileInfo(0, 1000, "movie.mp4")
	assert.NoError(t, err)
	_, err = preview.NewInfo(torrentID, "", 100, []preview.File{fi}, []byte("12345"))
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

func TestFileInfo_AddImages(t *testing.T) {
	idx := 33
	length := preview.DownloadSize * 10
	name := "movie.mp4"

	fi, err := preview.NewFileInfo(idx, length, name)
	assert.NoError(t, err)

	img := preview.NewImage("1234", idx, "test name", 10)
	assert.Len(t, fi.Images(), 0)

	err = fi.AddImage(img)
	assert.NoError(t, err)

	assert.Len(t, fi.Images(), 1)
	assert.Equal(t, img, fi.Images()[0])
}

func TestFileInfo_AddImagesWithNonMatchingID(t *testing.T) {
	idx := 33
	length := preview.DownloadSize * 10
	name := "movie.mp4"

	fi, err := preview.NewFileInfo(idx, length, name)
	assert.NoError(t, err)

	err = fi.AddImage(preview.NewImage("1234", idx+1, "test name", 10))
	assert.Error(t, err)
}
