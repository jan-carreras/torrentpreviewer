package preview_test

import (
	"github.com/stretchr/testify/assert"
	"prevtorrent/internal/preview"
	"testing"
)

func Test_Image(t *testing.T) {
	torrentID := "fake torrent"
	fileID := 0
	name := "test image.jpg"
	length := 100

	img := preview.NewImage(torrentID, fileID, name, length)

	assert.Equal(t, torrentID, img.TorrentID())
	assert.Equal(t, fileID, img.FileID())
	assert.Equal(t, name, img.Name())
	assert.Equal(t, length, img.Length())
}
