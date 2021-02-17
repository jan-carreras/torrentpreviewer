package preview

import (
	"fmt"
	"strings"
)

type DownloadPlan struct {
	torrent     Info
	pieceRanges []PieceRange
}

func NewDownloadPlan(torrent Info) *DownloadPlan {
	return &DownloadPlan{
		torrent:     torrent,
		pieceRanges: make([]PieceRange, 0),
	}
}

func (dp *DownloadPlan) GetTorrent() Info {
	return dp.torrent
}
func (dp *DownloadPlan) GetPlan() []PieceRange {
	return dp.pieceRanges
}

func (dp *DownloadPlan) AddAll() error {
	for _, file := range dp.torrent.SupportedFiles() {
		if err := dp.addDownloadToPlan(file, file.DownloadSize(), 0); err != nil {
			return err
		}
	}
	return nil
}

func (dp *DownloadPlan) CountPieces() int {
	count := map[int]interface{}{}
	for _, p := range dp.pieceRanges {
		for i := p.pieceStart; i <= p.pieceEnd; i++ {
			count[i] = struct{}{}
		}
	}
	return len(count)
}

func (dp *DownloadPlan) addDownloadToPlan(fi FileInfo, length, offset int) error {
	if !fi.IsSupportedExtension() {
		return fmt.Errorf("file %s has not a supported extension", fi.name)
	}

	if length+offset > fi.length {
		return fmt.Errorf("length+offset should be less than the total size of the file. having=%v, expecting <= %v", length+offset, fi.length)
	}
	if length <= 0 {
		return fmt.Errorf("only valid positive non-zero lengths. having=%v", length)
	}
	if length < 0 {
		return fmt.Errorf("only valid positive or zero offsets. having=%v", offset)
	}

	start := findStartingByteOfFile(dp.torrent, fi)

	dp.addToDownloadPlan(NewPieceRange(dp.torrent, fi, start, offset, length))
	return nil
}

func (dp *DownloadPlan) addToDownloadPlan(piece PieceRange) {
	dp.pieceRanges = append(dp.pieceRanges, piece)
}

type PieceRange struct {
	torrent          Info
	fi               FileInfo
	pieceStart       int // Piece pieceStart
	pieceEnd         int // Piece pieceEnd
	firstPieceOffset int // In Bytes. The file not necessarily starts at the byte 0 of the Piece. This offset indicates when it starts inside the piece
	lastPieceOffset  int // In Bytes. The file not necessarily ends at the pieceEnd of the last Piece. This offset indicates when it ends inside the piece
	pieceLength      int // In Bytes. The length of each piece of this torrent
}

func NewPieceRange(torrent Info, fi FileInfo, start, offset, length int) PieceRange {
	startPosition := start + offset
	length = length - 1
	return PieceRange{
		torrent:          torrent,
		fi:               fi,
		pieceStart:       startPosition / torrent.PieceLength(),
		pieceEnd:         (startPosition + length) / torrent.PieceLength(),
		firstPieceOffset: startPosition % torrent.PieceLength(),
		lastPieceOffset:  (startPosition + length) % torrent.PieceLength(),
		pieceLength:      torrent.PieceLength(),
	}
}

func (p PieceRange) Name() string {
	return strings.ReplaceAll(p.fi.name, "/", "--")
}

func (p PieceRange) FileID() int {
	return p.fi.ID()
}

func (p PieceRange) Start() int {
	return p.pieceStart
}

func (p PieceRange) End() int {
	return p.pieceEnd
}

func (p PieceRange) StartOffset(idx int) int {
	if idx == p.pieceStart {
		return p.firstPieceOffset
	}
	return 0
}

func (p PieceRange) EndOffset(idx int) int {
	if idx == p.pieceEnd {
		return p.lastPieceOffset + 1
	}
	return p.pieceLength
}

func (p PieceRange) PieceCount() int {
	return p.pieceEnd - p.pieceStart + 1 // pieceStart is zero index
}

func (p PieceRange) Torrent() Info {
	return p.torrent
}

func findStartingByteOfFile(t Info, fi FileInfo) int {
	start := 0
	for _, f := range t.files {
		if f == fi {
			break
		}
		start += f.length
	}
	return start
}

type Piece struct {
	torrentID string
	pieceID   int
	data      []byte
}

func NewPiece(torrentID string, pieceID int, data []byte) *Piece {
	return &Piece{torrentID: torrentID, pieceID: pieceID, data: data}
}

func (p Piece) TorrentID() string {
	return p.torrentID
}

func (p Piece) ID() int {
	return p.pieceID
}

func (p Piece) Data() []byte {
	return p.data
}
