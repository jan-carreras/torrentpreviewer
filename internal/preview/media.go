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
	return fmt.Sprintf("%v.%v.%v-%v.%v",
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

type PieceRegistry struct {
	downloadPlan     *DownloadPlan
	matcher          map[int][]*pieceRangeCounter
	pieces           map[int]*Piece
	plansCompletedCh chan PieceRange
	notifiedPieces   int
}

func NewPieceRegistry(plan *DownloadPlan) *PieceRegistry {
	matcher := make(map[int][]*pieceRangeCounter)

	for _, priceRange := range plan.GetPlan() {
		counter := newPieceRangeCounter(priceRange)
		for i := priceRange.Start(); i <= priceRange.End(); i++ {
			matcher[i] = append(matcher[i], counter)
		}
	}

	return &PieceRegistry{
		downloadPlan:     plan,
		matcher:          matcher,
		pieces:           make(map[int]*Piece),
		plansCompletedCh: make(chan PieceRange, plan.CountPieces()),
	}
}

func (pr *PieceRegistry) AddPiece(p Piece) error {
	if _, found := pr.matcher[p.pieceID]; !found {
		return fmt.Errorf("part %v not previously registered in matcher", p.pieceID)
	}

	pr.registerPiece(p)

	for _, counter := range pr.matcher[p.pieceID] {
		counter.count()
		if counter.areAllPiecesDownloaded() {
			pr.notifyAllPartsDownloaded(counter.pieceRange)
		}
	}
	return nil
}

func (pr *PieceRegistry) registerPiece(p Piece) {
	pr.pieces[p.pieceID] = &p
}

func (pr *PieceRegistry) GetPiece(idx int) (Piece, bool) {
	p, found := pr.pieces[idx]
	return *p, found
}

func (pr *PieceRegistry) SubscribeAllPartsDownloaded() chan PieceRange {
	return pr.plansCompletedCh
}

func (pr *PieceRegistry) notifyAllPartsDownloaded(pieceRange PieceRange) {
	pr.plansCompletedCh <- pieceRange
	pr.notifiedPieces++
	if pr.notifiedPieces == len(pr.downloadPlan.GetPlan()) {
		close(pr.plansCompletedCh)
	}
}

type pieceRangeCounter struct {
	pieceRange       PieceRange
	piecesDownloaded int
}

func (c *pieceRangeCounter) count() {
	c.piecesDownloaded++
}

func (c *pieceRangeCounter) areAllPiecesDownloaded() bool {
	return c.piecesDownloaded >= c.pieceRange.PieceCount()
}

func newPieceRangeCounter(
	pieceRange PieceRange,
) *pieceRangeCounter {
	return &pieceRangeCounter{pieceRange: pieceRange}
}

type BundlePlan struct {
}

func NewBundlePlan() BundlePlan {
	return BundlePlan{}
}

func (b BundlePlan) Bundle(registry *PieceRegistry, torrentID string, plan PieceRange) (MediaPart, error) {
	piece := new(bytes.Buffer)

	for pieceIdx := plan.Start(); pieceIdx <= plan.End(); pieceIdx++ {
		p, found := registry.GetPiece(pieceIdx)
		if !found {
			return MediaPart{}, errors.New("piece not found in the registry. could be ignored but kept to further investigate")
		}
		start := plan.StartOffset(pieceIdx)
		end := plan.EndOffset(pieceIdx)
		if start > end {
			return MediaPart{}, fmt.Errorf("start offset %v bigger than end offset %v", start, end)
		}
		if start > len(p.data) {
			return MediaPart{}, fmt.Errorf("start offset %v is bigger than length of slice %v", start, len(p.data))
		}
		if end > len(p.data) {
			return MediaPart{}, fmt.Errorf("end offset %v is bigger than length of slice %v", start, len(p.data))
		}

		rawData := p.data[start:end]
		_, err := piece.Write(rawData)
		if err != nil {
			return MediaPart{}, err
		}
	}

	return NewMediaPart(torrentID, plan, piece.Bytes()), nil
}
