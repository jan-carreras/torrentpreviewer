package preview

import (
	"bytes"
	"context"
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

type PieceRegistry struct {
	downloadPlan     *DownloadPlan
	matcher          map[int][]*pieceRangeCounter
	pieces           map[int]*Piece
	pieceIncomingCh  chan Piece
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
	// TODO: If the plan has 0 pieces, there is little to be done here. We should error in case of trying to append something?
	// Would block indefinitively, really

	return &PieceRegistry{
		downloadPlan:     plan,
		matcher:          matcher,
		pieces:           make(map[int]*Piece),
		pieceIncomingCh:  make(chan Piece, plan.CountPieces()),
		plansCompletedCh: make(chan PieceRange, plan.CountPieces()),
	}
}

func (pr *PieceRegistry) GetPiece(idx int) (Piece, bool) {
	p, found := pr.pieces[idx]
	return *p, found
}

func (pr *PieceRegistry) ListenForPieces(ctx context.Context) {
	go pr.listen(ctx)
}

func (pr *PieceRegistry) SubscribeAllPartsDownloaded() chan PieceRange {
	// TODO: Error if not ListeningForPieces first
	return pr.plansCompletedCh
}

func (pr *PieceRegistry) RegisterPiece(piece Piece) {
	pr.pieceIncomingCh <- piece

}

func (pr *PieceRegistry) NoMorePieces() {
	// TODO: Is this good?
	close(pr.pieceIncomingCh)
}

func (pr *PieceRegistry) doneAddingPieces() {
	close(pr.plansCompletedCh)
}

func (pr *PieceRegistry) addPiece(p Piece) error {
	// TODO: Add error if writing on closed channel

	if _, found := pr.matcher[p.pieceID]; !found {
		return fmt.Errorf("part %v not previously registered in matcher", p.pieceID)
	}

	pr.registerPiece(p)

	for _, counter := range pr.matcher[p.pieceID] {
		counter.addOne()
		fmt.Println(counter.piecesDownloaded)
		if counter.areAllPiecesDownloaded() {
			fmt.Println("pieces are downloaded")
			pr.notifyAllPartsDownloaded(counter.pieceRange)
		}
	}
	return nil
}

func (pr *PieceRegistry) registerPiece(p Piece) {
	pr.pieces[p.pieceID] = &p
}

func (pr *PieceRegistry) notifyAllPartsDownloaded(pieceRange PieceRange) {
	pr.plansCompletedCh <- pieceRange
	pr.notifiedPieces++
	if pr.notifiedPieces == len(pr.downloadPlan.GetPlan()) {
		close(pr.plansCompletedCh)
	}
}

func (pr *PieceRegistry) listen(ctx context.Context) {
	for {
		select {
		case piece, isOpen := <-pr.pieceIncomingCh:
			fmt.Println("hi there", piece)
			if !isOpen {
				pr.doneAddingPieces()
				return
			}

			// TODO: Add/Remove logger
			/*log := s.logger.WithFields(logrus.Fields{
				"torrentID": piece.TorrentID(),
				"piece":     piece.ID(),
			})*/

			if err := pr.addPiece(piece); err != nil {
				/*log.Error(err)*/
				return
			}
			/*log.Debug("part added to registry")*/
		case <-ctx.Done():
			pr.doneAddingPieces()
			return
		}
	}
}

type pieceRangeCounter struct {
	pieceRange       PieceRange
	piecesDownloaded int
}

func (c *pieceRangeCounter) addOne() {
	c.piecesDownloaded++
}

func (c *pieceRangeCounter) areAllPiecesDownloaded() bool {
	fmt.Println("piece count", c.pieceRange.PieceCount())
	return c.piecesDownloaded >= c.pieceRange.PieceCount()
}

func newPieceRangeCounter(
	pieceRange PieceRange,
) *pieceRangeCounter {
	return &pieceRangeCounter{pieceRange: pieceRange}
}

type BundlePlan struct{}

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

		rawData := p.data[start : end+1] // end of rang is exclusive
		_, err := piece.Write(rawData)
		if err != nil {
			return MediaPart{}, err
		}
	}

	return NewMediaPart(torrentID, plan, piece.Bytes()), nil
}
