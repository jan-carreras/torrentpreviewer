package preview_test

import (
	"github.com/stretchr/testify/assert"
	"prevtorrent/internal/preview"
	"testing"
)

func TestMediaPart(t *testing.T) {
	torrentID := "ZOCmzqipffw7ollmic5hub6bpcsdeoqu"

	fi, err := preview.NewFileInfo(0, 1000, "movie.mp4")
	assert.NoError(t, err)
	files := []preview.FileInfo{fi}

	torrent, err := preview.NewInfo(torrentID, "test movie", 100, files, []byte("12345"))
	assert.NoError(t, err)

	pr := preview.NewPieceRange(torrent, fi, 0, 150, 100)
	data := []byte("1234")

	media := preview.NewMediaPart(torrentID, pr, data)

	assert.Equal(t, "ZOCmzqipffw7ollmic5hub6bpcsdeoqu.0.1-2.movie.mp4.jpg", media.Name())
	assert.Equal(t, data, media.Data())
	assert.Equal(t, pr, media.PieceRange())
}
