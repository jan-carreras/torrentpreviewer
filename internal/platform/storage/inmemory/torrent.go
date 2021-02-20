package inmemory

import (
	"io"
	"sync"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

type TorrentStorage struct {
	torrents map[string]*torrentDownloaded
	mutex    sync.Mutex
}

// NewTorrentStorage return a TorrentStorage
func NewTorrentStorage() *TorrentStorage {
	return &TorrentStorage{
		torrents: make(map[string]*torrentDownloaded),
	}
}

// Close deletes all the torrents information
func (m *TorrentStorage) Close() error {
	m.mutex.Lock()
	m.torrents = make(map[string]*torrentDownloaded)
	m.mutex.Unlock()
	return nil
}

// OpenTorrent returns an a TorrentImpl that manages the pieces and handles reads/writes
func (m *TorrentStorage) OpenTorrent(info *metainfo.Info, infoHash metainfo.Hash) (storage.TorrentImpl, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if t, found := m.torrents[infoHash.String()]; found {
		return t.torrent, nil
	}

	t := &torrentDownloaded{
		info:     info,
		infoHash: infoHash,
		torrent:  newMemoryTorrentImpl(info),
	}

	m.torrents[infoHash.String()] = t

	return t.torrent, nil
}

type memoryTorrentImpl struct {
	info   *metainfo.Info
	pieces map[int]storage.PieceImpl
}

func newMemoryTorrentImpl(info *metainfo.Info) *memoryTorrentImpl {
	return &memoryTorrentImpl{
		info:   info,
		pieces: make(map[int]storage.PieceImpl),
	}
}

// Pieces returns the requested piece
func (m *memoryTorrentImpl) Piece(p metainfo.Piece) storage.PieceImpl {
	if piece, found := m.pieces[p.Index()]; found {
		return piece
	}

	m.pieces[p.Index()] = NewPiece(m.info.PieceLength)
	return m.pieces[p.Index()]
}

// Close frees all Pieces information from a torrent
func (m *memoryTorrentImpl) Close() error {
	m.pieces = make(map[int]storage.PieceImpl)
	return nil
}

type Piece struct {
	io.ReaderAt
	io.WriterAt
	data     []byte
	dataMux  sync.RWMutex
	complete *bool
}

func NewPiece(pieceLength int64) *Piece {
	return &Piece{
		data: make([]byte, pieceLength),
	}
}

func (p *Piece) WriteAt(b []byte, off int64) (int, error) {
	p.dataMux.Lock()
	defer p.dataMux.Unlock()
	return copy(p.data[off:], b), nil
}

func (p *Piece) ReadAt(b []byte, off int64) (n int, err error) {
	p.dataMux.RLock()
	defer p.dataMux.RUnlock()
	end := int(off) + len(b)
	return copy(b, p.data[off:end]), nil
}

func (p *Piece) Completion() storage.Completion {
	p.dataMux.RLock()
	defer p.dataMux.RUnlock()
	ok := true
	complete := false
	if p.complete == nil {
		ok = false
	} else {
		complete = *p.complete
	}

	return storage.Completion{
		Complete: complete,
		Ok:       ok,
	}
}

func (p *Piece) MarkComplete() error {
	p.dataMux.Lock()
	defer p.dataMux.Unlock()
	c := true
	p.complete = &c
	return nil
}

func (p *Piece) MarkNotComplete() error {
	p.dataMux.Lock()
	defer p.dataMux.Unlock()
	c := false
	p.complete = &c
	return nil
}

type torrentDownloaded struct {
	info     *metainfo.Info
	infoHash metainfo.Hash
	torrent  storage.TorrentImpl
}
