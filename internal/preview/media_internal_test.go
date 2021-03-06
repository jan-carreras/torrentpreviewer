package preview

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPieceRangeCounter_TwoPieces(t *testing.T) {
	torrentID := "cb84ccc10f296df72d6c40ba7a07c178a4323a14"
	pieceLength := 100

	fi, err := NewFileInfo(0, 1000, "test/movie.mp4")
	assert.NoError(t, err)
	torrent, err := NewInfo(torrentID, "generic movie", pieceLength, []File{fi}, []byte(""))
	assert.NoError(t, err)

	pr, err := NewPieceRange(torrent, fi, 0, 0, 150)
	require.NoError(t, err)

	counter := newPieceRangeCounter(pr)
	assert.False(t, counter.areAllPiecesDownloaded())
	counter.addOne() // +1 . Total=1
	assert.False(t, counter.areAllPiecesDownloaded())
	counter.addOne() // +1 . Total=2
	assert.True(t, counter.areAllPiecesDownloaded())

	assert.Equal(t, pr.PieceCount(), counter.piecesDownloaded)
}
