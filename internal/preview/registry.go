package preview

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

var ErrPriceRegistryWithNothingToWaitFor = errors.New("the plan has 0 pieces to wait for, thus using the registry to retrieve responses is useless")

type PieceStorage interface {
	Set(p *Piece)
	Get(id int) (*Piece, bool)
}

// PieceInMemoryStorage is in charge of registering all the pieces/chunks received via a peer
// and store it until is read.
// Various pieces might create a file. Different files might share the same pieces. That's why
// we inspect the DownloadPlan to know how many complete files depend on a given piece,
// and when we read from the storage we subtract 1 from the pieceCount.
// If the pieceCount gets to 0, we free the space. So this storage is mean to be read just once
// per file. We do this to keep memory footprint low.
type PieceInMemoryStorage struct {
	pieceCount map[int]int
	storage    map[int]*Piece
	storageMux sync.RWMutex
}

// NewPieceInMemoryStorage creates a PieceInMemoryStorage
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

// Set saves a piece in the storage
func (m *PieceInMemoryStorage) Set(p *Piece) {
	m.storageMux.Lock()
	defer m.storageMux.Unlock()
	m.storage[p.pieceID] = p
}

// Get returns a piece from the storage. Each piece has an associates counter,
// and each time we read a piece the counter gets down by one. If it gets to
// 0, we delete the piece
func (m *PieceInMemoryStorage) Get(id int) (*Piece, bool) {
	m.storageMux.RLock()
	p, found := m.storage[id]
	m.storageMux.RUnlock()

	if !found {
		return nil, found

	}

	m.storageMux.Lock()
	defer m.storageMux.Unlock()
	m.pieceCount[id]--
	if m.pieceCount[id] == 0 {
		delete(m.pieceCount, id)
		delete(m.storage, id)
	}
	return p, found
}

// PieceRegistry keeps track of all the pieces downloaded for a DownloadPlan
// and knows when we have all the pieces to generate a file.
// It operates reading from pieceIncomingCh channel and outputs to the
// plansCompletedCh when a PieceRange has all its pieces. Then an outside actor
// can read the pieces from the storage.
// TODO: This struct has too many responsabilities, too many channel logic and probably
//       a lot of hidden bugs.
type PieceRegistry struct {
	logger              *logrus.Logger
	downloadPlan        *DownloadPlan
	storage             PieceStorage
	matcher             map[int][]*pieceRangeCounter
	pieceIncomingCh     chan *Piece
	pieceIncomingChMux  sync.Once
	plansCompletedCh    chan PieceRange
	plansCompletedChMux sync.Once
	notifiedPieces      int
}

// NewPieceRegistry creates a PieceRegistry
func NewPieceRegistry(ctx context.Context, logger *logrus.Logger, plan *DownloadPlan, storage PieceStorage) (*PieceRegistry, error) {
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

	pr := &PieceRegistry{
		logger:           logger,
		downloadPlan:     plan,
		matcher:          matcher,
		storage:          storage,
		pieceIncomingCh:  make(chan *Piece, plan.CountPieces()),
		plansCompletedCh: make(chan PieceRange, plan.CountPieces()),
	}

	pr.listenForPieces(ctx)

	return pr, nil
}

// GetPiece reads a piece from the storage and returns it
func (pr *PieceRegistry) GetPiece(idx int) (*Piece, bool) {
	return pr.storage.Get(idx)
}

// SubscribeAllPartsDownloaded returns the channel where all the PieceRange
// are going to be published when completed.
func (pr *PieceRegistry) SubscribeAllPartsDownloaded() chan PieceRange {
	return pr.plansCompletedCh
}

//  RegisterPiece sends a piece downloaded by the torrent without blocking
func (pr *PieceRegistry) RegisterPiece(piece *Piece) {
	pr.pieceIncomingCh <- piece
}

// NoMorePieces notifies that there is no more incoming data and we can stop
// listening for more pices. It usually means that the torrent has downloaded
// all what's in the DownloadPlan
func (pr *PieceRegistry) NoMorePieces() {
	pr.pieceIncomingChMux.Do(func() {
		close(pr.pieceIncomingCh)
	})
}

// RunOnPieceReady receives a callback and executes it every time a PieceRange
// from the DownloadPlan has been completed
func (pr *PieceRegistry) RunOnPieceReady(ctx context.Context, fnx func(part PieceRange) error) error {
	for {
		select {
		case part, isOpen := <-pr.SubscribeAllPartsDownloaded():
			if !isOpen {
				return nil
			}

			if err := fnx(part); err != nil {
				return err
			}

		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %v", ctx.Err())
		}
	}
}

func (pr *PieceRegistry) listenForPieces(ctx context.Context) {
	go pr.listen(ctx)
}

func (pr *PieceRegistry) closeIncomingChanel() {
	pr.plansCompletedChMux.Do(func() {
		close(pr.plansCompletedCh)
	})
}

func (pr *PieceRegistry) addPiece(p *Piece) error {
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
