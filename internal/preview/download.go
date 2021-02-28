package preview

import (
	"fmt"
	"strings"
)

// DownloadPlan helps to describe what we want to download from the torrent.
type DownloadPlan struct {
	torrent     Torrent
	pieceRanges []PieceRange
}

// NewDownloadPlan returns a DownloadPlan
func NewDownloadPlan(torrent Torrent) *DownloadPlan {
	return &DownloadPlan{
		torrent:     torrent,
		pieceRanges: make([]PieceRange, 0),
	}
}

// GetTorrent returns the Torrent to download
func (dp *DownloadPlan) GetTorrent() Torrent {
	return dp.torrent
}

// GetPlan returns the plan to download. Each PieceRange usually is a part of a file,
// but could describe various data ranges from the same file.
func (dp *DownloadPlan) GetPlan() []PieceRange {
	return dp.pieceRanges
}

// AddAll adds all the supported files of the torrent to download with a pre-set settings:
//       Start at the beginning of the file and download the recommended file.DownloadSize()
// Note that AddAll with check in TorrentImages for the files already downloaded and will skip those
func (dp *DownloadPlan) AddAll(torrentImages *TorrentImages, maxSizeDownloaded int) error {
	for _, file := range dp.torrent.SupportedFiles() {
		if maxSizeDownloaded != 0 && dp.DownloadSize() > maxSizeDownloaded {
			break
		}

		if err := dp.addDownloadToPlan(file, torrentImages); err != nil {
			return err
		}
	}
	return nil
}

// CountPieces returns the number of unique pieces to be downloaded. Remember the multiple PieceRange,
// could share some pieces.
func (dp *DownloadPlan) CountPieces() int {
	count := map[int]interface{}{}
	for _, p := range dp.pieceRanges {
		for i := p.pieceStart; i <= p.pieceEnd; i++ {
			count[i] = struct{}{}
		}
	}
	return len(count)
}

func (dp *DownloadPlan) DownloadSize() int {
	return dp.CountPieces() * dp.torrent.pieceLength
}

func (dp *DownloadPlan) addDownloadToPlan(fi File, torrentImages *TorrentImages) error {
	length := fi.DownloadSize()
	offset := 0

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
	pr := NewPieceRange(dp.torrent, fi, start, offset, length)

	if torrentImages.HaveImage(pr.Name()) {
		return nil
	}

	dp.addToDownloadPlan(pr, length)
	return nil
}

func (dp *DownloadPlan) addToDownloadPlan(piece PieceRange, downloadSize int) {
	dp.pieceRanges = append(dp.pieceRanges, piece)
}

// PieceRange describes a section of a file we want to download
// We need to store various indexes an offsets commented in the struct.
type PieceRange struct {
	torrent          Torrent
	fi               File
	pieceStart       int // Piece pieceStart
	pieceEnd         int // Piece pieceEnd
	firstPieceOffset int // In Bytes. The file not necessarily starts at the byte 0 of the Piece. This offset indicates when it starts inside the piece
	lastPieceOffset  int // In Bytes. The file not necessarily ends at the pieceEnd of the last Piece. This offset indicates when it ends inside the piece
	pieceLength      int // In Bytes. The length of each piece of this torrent
}

// NewPieceRange returns a PieceRange
func NewPieceRange(torrent Torrent, fi File, start, offset, length int) PieceRange {
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

// Name returns the name of the file. It's supposed to be HTTP friendly
func (p PieceRange) Name() string {
	name := strings.ReplaceAll(p.fi.name, "/", "--")
	name = strings.ReplaceAll(name, " ", "-")
	return fmt.Sprintf("%v.%v.%v-%v.%v.jpg",
		p.Torrent().ID(),
		p.fi.idx,
		p.Start(),
		p.End(),
		name,
	)
}

// FileID returns the obvious
func (p PieceRange) FileID() int {
	return p.fi.ID()
}

// Starts returns the piece we're going to start with
func (p PieceRange) Start() int {
	return p.pieceStart
}

// End returns the piece we're going to end with. It **must** be downloaded. It's [start,end] (both inclusive)
func (p PieceRange) End() int {
	return p.pieceEnd
}

// StartOffset returns the offset were we have to read given a PieceID
func (p PieceRange) StartOffset(idx int) int {
	if idx == p.pieceStart {
		return p.firstPieceOffset
	}
	return 0
}

// EndOffset returns the offset were we have to read given a PieceID
func (p PieceRange) EndOffset(idx int) int {
	if idx == p.pieceEnd {
		return p.lastPieceOffset + 1
	}
	return p.pieceLength
}

// PieceCount returns the number of pieces for this PieceRange
func (p PieceRange) PieceCount() int {
	return p.pieceEnd - p.pieceStart + 1 // pieceStart is zero index
}

// Torrent returns the obvious
func (p PieceRange) Torrent() Torrent {
	return p.torrent
}

func findStartingByteOfFile(t Torrent, fi File) int {
	start := 0
	for _, f := range t.files {
		if f.IsEqual(fi) {
			break
		}
		start += f.length
	}
	return start
}

// Piece describes a chunk/piece of raw data of the torrent.
// The atom (smallest piece) struct for our application
type Piece struct {
	torrentID string
	pieceID   int
	data      []byte
}

// NewPiece returns (no? really! good job Sherlock!) a Piece
func NewPiece(torrentID string, pieceID int, data []byte) *Piece {
	return &Piece{torrentID: torrentID, pieceID: pieceID, data: data}
}

// TorrentID returns the obivous
func (p Piece) TorrentID() string {
	return p.torrentID
}

// ID returns the PieceID
func (p Piece) ID() int {
	return p.pieceID
}

// Data returns the raw data. By itself means nothing. Might contains a portion of a file
// or multiples files. We don't know without a DownloadPlan and its PieceRange
func (p Piece) Data() []byte {
	return p.data
}
