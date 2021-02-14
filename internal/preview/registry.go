package preview

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrPriceRegistryWithNothingToWaitFor = errors.New("the plan has 0 pieces to wait for, thus using the registry to retrieve responses is useless")

//go:generate mockery --case=snake --outpkg=domainmocks --output=domainmocks --name=PieceStorage
type PieceStorage interface {
	Set(p *Piece)
	Get(id int) (*Piece, bool)
}

type PieceInMemoryStorage struct {
	pieceCount map[int]int
	storage    map[int]*Piece
	storageMux sync.RWMutex
}

func NewPieceInMemoryStorage(plan DownloadPlan) *PieceInMemoryStorage {
	count := make(map[int]int)
	for _, p := range plan.GetPlan() {
		for i := p.Start(); i <= p.End(); i++ {
			count[i]++
		}
	}

	return &PieceInMemoryStorage{
		pieceCount: count,
		storage:    make(map[int]*Piece),
	}
}

func (m *PieceInMemoryStorage) Set(p *Piece) {
	m.storageMux.Lock()
	defer m.storageMux.Unlock()
	m.storage[p.pieceID] = p
}

func (m *PieceInMemoryStorage) Get(id int) (*Piece, bool) {
	m.storageMux.RLock()
	defer m.storageMux.RUnlock()
	p, found := m.storage[id]
	if !found {
		return nil, found

	}
	m.pieceCount[id]--
	if m.pieceCount[id] == 0 {
		delete(m.pieceCount, id)
		delete(m.storage, id)
	}
	return p, found
}

type PieceRegistry struct {
	downloadPlan        *DownloadPlan
	storage             PieceStorage
	matcher             map[int][]*pieceRangeCounter
	pieceIncomingCh     chan *Piece
	pieceIncomingChMux  sync.Once
	plansCompletedCh    chan PieceRange // Well, because I would have to call "bundle" from here.
	plansCompletedChMux sync.Once
	notifiedPieces      int
}

func NewPieceRegistry(plan *DownloadPlan, storage PieceStorage) (*PieceRegistry, error) {
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
		downloadPlan:     plan,
		matcher:          matcher,
		storage:          storage,
		pieceIncomingCh:  make(chan *Piece, plan.CountPieces()),
		plansCompletedCh: make(chan PieceRange, plan.CountPieces()),
	}, nil
}

func (pr *PieceRegistry) GetPiece(idx int) (*Piece, bool) {
	return pr.storage.Get(idx)
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
	pr.storage.Set(p)
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
				//pr.logger.Debug("no more input. Seems that those were all the pieces")
				pr.closeIncomingChanel()
				return
			}

			/*log := pr.logger.WithFields(logrus.Fields{
				"torrentID": piece.TorrentID(),
				"piece":     piece.ID(),
			})*/

			if err := pr.addPiece(piece); err != nil {
				///log.Error(err)
				return
			}
			//log.Debug("part added to registry")
		case <-ctx.Done():
			pr.closeIncomingChanel()
			return
		}
	}
}
