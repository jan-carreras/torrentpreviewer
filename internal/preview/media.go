package preview

import (
	"bytes"
	"errors"
	"fmt"
)

type MediaPart struct {
	torrentID  string
	pieceRange PieceRange
	data       []byte
}

func NewMediaPart(torrentID string, pieceRange PieceRange, data []byte) MediaPart {
	return MediaPart{torrentID: torrentID, pieceRange: pieceRange, data: data}
}

func (p MediaPart) Name() string {
	return fmt.Sprintf("%v.%v.%v-%v.%v.jpg",
		p.torrentID,
		p.pieceRange.fi.idx,
		p.pieceRange.Start(),
		p.pieceRange.End(),
		p.pieceRange.Name(),
	)
}

func (p MediaPart) PieceRange() PieceRange {
	return p.pieceRange
}

func (p MediaPart) Data() []byte {
	return p.data
}

type pieceRangeCounter struct {
	pieceRange       PieceRange
	piecesDownloaded int
}

func newPieceRangeCounter(
	pieceRange PieceRange,
) *pieceRangeCounter {
	return &pieceRangeCounter{pieceRange: pieceRange}
}

func (c *pieceRangeCounter) addOne() {
	c.piecesDownloaded++
}

func (c *pieceRangeCounter) areAllPiecesDownloaded() bool {
	return c.piecesDownloaded >= c.pieceRange.PieceCount()
}

type BundlePlan struct{}

func NewBundlePlan() BundlePlan {
	return BundlePlan{}
}

func (b BundlePlan) Bundle(registry *PieceRegistry, pieceRange PieceRange) (MediaPart, error) {
	piece := new(bytes.Buffer)

	for pieceIdx := pieceRange.Start(); pieceIdx <= pieceRange.End(); pieceIdx++ {
		p, found := registry.GetPiece(pieceIdx)
		if !found {
			return MediaPart{}, errors.New("piece not found in the registry. could be ignored but kept to further investigate")
		}
		start := pieceRange.StartOffset(pieceIdx)
		end := pieceRange.EndOffset(pieceIdx)
		if start > end {
			return MediaPart{}, fmt.Errorf("start offset %v bigger than end offset %v", start, end)
		}
		if start > len(p.data) {
			return MediaPart{}, fmt.Errorf("start offset %v is bigger than length of slice %v", start, len(p.data))
		}
		if end > len(p.data) {
			return MediaPart{}, fmt.Errorf("end offset %v is bigger than length of slice %v", start, len(p.data))
		}

		rawData := p.data[start:end] // end of rang is exclusive
		_, err := piece.Write(rawData)
		if err != nil {
			return MediaPart{}, err
		}
	}

	return NewMediaPart(pieceRange.Torrent().ID(), pieceRange, piece.Bytes()), nil
}
