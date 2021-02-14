package preview

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPieceRangeCounter_TwoPieces(t *testing.T) {
	torrentID := "zocmzqipffw7ollmic5hub6bpcsdeoqu"
	pieceLength := 100

	fi, err := NewFileInfo(0, 1000, "test/movie.mp4")
	assert.NoError(t, err)
	torrent, err := NewInfo(torrentID, "generic movie", pieceLength, []FileInfo{fi}, []byte(""))
	assert.NoError(t, err)

	pr := NewPieceRange(torrent, fi, 0, 0, 150)

	counter := newPieceRangeCounter(pr)
	assert.False(t, counter.areAllPiecesDownloaded())
	counter.addOne() // +1 . Total=1
	assert.False(t, counter.areAllPiecesDownloaded())
	counter.addOne() // +1 . Total=2
	assert.True(t, counter.areAllPiecesDownloaded())

	assert.Equal(t, pr.PieceCount(), counter.piecesDownloaded)
}
