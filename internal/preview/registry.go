package preview

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
)

var ErrPriceRegistryWithNothingToWaitFor = errors.New("the plan has 0 pieces to wait for, thus using the registry to retrieve responses is useless")

type PieceRegistry struct {
	logger              *logrus.Logger
	downloadPlan        *DownloadPlan
	matcher             map[int][]*pieceRangeCounter
	pieces              map[int]*Piece
	pieceIncomingCh     chan *Piece
	pieceIncomingChMux  sync.Once
	plansCompletedCh    chan PieceRange
	plansCompletedChMux sync.Once
	notifiedPieces      int
}

func NewPieceRegistry(plan *DownloadPlan, logger ...*logrus.Logger) (*PieceRegistry, error) {
	if plan.CountPieces() == 0 {
		return nil, ErrPriceRegistryWithNothingToWaitFor
	}

	matcher := make(map[int][]*pieceRangeCounter)

	for _, priceRange := range plan.GetPlan() {
		counter := newPieceRangeCounter(priceRange)
		for i := priceRange.Start(); i <= priceRange.End(); i++ {
			matcher[i] = append(matcher[i], counter)
		}
	}

	return &PieceRegistry{
		logger:           logger[0],
		downloadPlan:     plan,
		matcher:          matcher,
		pieces:           make(map[int]*Piece),
		pieceIncomingCh:  make(chan *Piece, plan.CountPieces()),
		plansCompletedCh: make(chan PieceRange, plan.CountPieces()),
	}, nil
}

func (pr *PieceRegistry) GetPiece(idx int) (Piece, bool) {
	p, found := pr.pieces[idx]
	if found {
		return *p, true
	}
	return Piece{}, found
}

func (pr *PieceRegistry) ListenForPieces(ctx context.Context) {
	go pr.listen(ctx)
}

func (pr *PieceRegistry) SubscribeAllPartsDownloaded() chan PieceRange {
	// TODO: Error if not ListeningForPieces first
	return pr.plansCompletedCh
}

func (pr *PieceRegistry) RegisterPiece(piece *Piece) {
	pr.pieceIncomingCh <- piece
}

func (pr *PieceRegistry) NoMorePieces() {
	pr.pieceIncomingChMux.Do(func() {
		close(pr.pieceIncomingCh)
	})
}

func (pr *PieceRegistry) closeIncomingChanel() {
	pr.plansCompletedChMux.Do(func() {
		close(pr.plansCompletedCh)
	})
}

func (pr *PieceRegistry) addPiece(p *Piece) error {
	// TODO: Add error if writing on closed channel

	if _, found := pr.matcher[p.pieceID]; !found {
		return fmt.Errorf("part %v not previously registered in matcher", p.pieceID)
	}

	pr.registerPiece(p)

	for _, counter := range pr.matcher[p.pieceID] {
		counter.addOne()
		if counter.areAllPiecesDownloaded() {
			pr.notifyAllPartsDownloaded(counter.pieceRange)
		}
	}
	return nil
}

func (pr *PieceRegistry) registerPiece(p *Piece) {
	pr.pieces[p.pieceID] = p
}

func (pr *PieceRegistry) notifyAllPartsDownloaded(pieceRange PieceRange) {
	pr.plansCompletedCh <- pieceRange
	pr.notifiedPieces++
	if pr.notifiedPieces == len(pr.downloadPlan.GetPlan()) {
		pr.closeIncomingChanel()
	}
}

func (pr *PieceRegistry) listen(ctx context.Context) {
	for {
		select {
		case piece, isOpen := <-pr.pieceIncomingCh:
			if !isOpen {
				pr.logger.Debug("no more input. Seems that those were all the pieces")
				pr.closeIncomingChanel()
				return
			}

			log := pr.logger.WithFields(logrus.Fields{
				"torrentID": piece.TorrentID(),
				"piece":     piece.ID(),
			})

			if err := pr.addPiece(piece); err != nil {
				log.Error(err)
				return
			}
			log.Debug("part added to registry")
		case <-ctx.Done():
			pr.closeIncomingChanel()
			return
		}
	}
}
