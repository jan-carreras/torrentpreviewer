package inmemory

import (
	"testing"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/stretchr/testify/assert"
)

const (
	mb = 1 << (10 * 2) // MiB, really
)

func TestTorrentStorage(t *testing.T) {
	info := &metainfo.Info{}

	hash := metainfo.NewHashFromHex("0123456789012345678901234567890123456789")
	hash2 := metainfo.NewHashFromHex("0000000000000000000000000000000000000000")

	storage := NewTorrentStorage()
	assert.Len(t, storage.torrents, 0)

	torrent, err := storage.OpenTorrent(info, hash)
	assert.NoError(t, err)
	assert.Len(t, storage.torrents, 1)

	torrentSame, err := storage.OpenTorrent(info, hash)
	assert.NoError(t, err)
	assert.Len(t, storage.torrents, 1)

	assert.Equal(t, torrent, torrentSame)

	_, err = storage.OpenTorrent(info, hash2)
	assert.NoError(t, err)
	assert.Len(t, storage.torrents, 2)

	err = storage.Close()
	assert.NoError(t, err)

	assert.Len(t, storage.torrents, 0)
}

func TestMemoryTorrentImpl(t *testing.T) {
	info := &metainfo.Info{}

	torrentPieces := newMemoryTorrentImpl(info)
	assert.Len(t, torrentPieces.pieces, 0)

	metaPiece := metainfo.Piece{Info: info}
	piece := torrentPieces.Piece(metaPiece)
	assert.Len(t, torrentPieces.pieces, 1)

	pieceSame := torrentPieces.Piece(metaPiece)
	assert.Len(t, torrentPieces.pieces, 1)

	assert.Equal(t, piece, pieceSame)

	err := torrentPieces.Close()
	assert.NoError(t, err)
	assert.Len(t, torrentPieces.pieces, 0)
}

func TestPiece(t *testing.T) {
	pieceLength := int64(4 * mb)

	piece := NewPiece(pieceLength)
	// To save memory, a piece should take 0 bytes in memory once initialized
	assert.Len(t, piece.data, 0)

	buf := make([]byte, pieceLength)
	n, err := piece.ReadAt(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(buf), n)

	// We have read from the piece (without writing before) and we still take 0 bytes in memory
	assert.Len(t, piece.data, 0)

	buf = make([]byte, 1)
	for i := 0; i < len(buf); i++ {
		buf[i] = 1
	}
	n, err = piece.WriteAt(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(buf), n)
	// When we store something, even one byte, we then allocate all the space needed for the whole piece
	assert.Len(t, piece.data, int(pieceLength))

	buf = make([]byte, 2)
	n, err = piece.ReadAt(buf, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(buf), n)

	assert.Equal(t, []byte{1, 0}, buf)

}

func TestPieceCompletion(t *testing.T) {
	pieceLength := int64(4 * mb)

	piece := NewPiece(pieceLength)
	completion := piece.Completion()
	assert.False(t, completion.Ok)
	assert.False(t, completion.Complete)

	assert.NoError(t, piece.MarkNotComplete())
	completion = piece.Completion()
	assert.True(t, completion.Ok)
	assert.False(t, completion.Complete)

	assert.NoError(t, piece.MarkComplete())
	completion = piece.Completion()
	assert.True(t, completion.Ok)
	assert.True(t, completion.Complete)

	assert.NoError(t, piece.MarkNotComplete())
	completion = piece.Completion()
	assert.True(t, completion.Ok)
	assert.False(t, completion.Complete)
}
